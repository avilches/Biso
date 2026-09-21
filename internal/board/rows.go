package board

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"biso/internal/model"
)

// This file is the whole translation between a task and its rows, in one
// place so that the two directions cannot drift apart.
//
// Two conventions come from the model and are repeated here only because
// this is where they touch SQL: the empty string is how a string field
// with no value is stored, since no caller can ever store a real empty
// string (docs/spec/valores-de-entrada.md#el-valor-vacío), and the zero
// time is how a date field with no value is stored, for the same reason.
// Ordinal is one more string of that first rule: an ordinal key can never
// be the empty string, so the empty string is a task with no key.

// taskColumns is the task row, in the order both directions use.
const taskColumns = `id, num, title, status, type, priority, parent, author, due, ordinal,
	description, plan, notes, summary, created_at, updated_at, archived,
	lease_expires_at, lease_holder, question_author, question_asked_at, question_body,
	next_criterion_key, next_comment_key`

// atLeastOne is the one normalization this translation does: a counter of
// criterion or of comment keys is stored as 1 and never as 0, because the
// first key of either list is 1 (docs/spec/modelo-de-datos/criterios.md#los-criterios-y-sus-claves-estables)
// and a task that has never had one is at the same point whichever of the
// two numbers it carries.
//
// Writing them apart would be a difference between two boards that behave
// identically, and the round trip of docs/spec/cmd/export.md would not
// reproduce it: the counters are deduced on import, above the highest key
// the line brought, so an imported task with no criteria comes back with a
// 1. The dump of the two databases that the symmetry test compares is what
// showed it.
//
// The task itself is settled and not only its row, so that what a caller
// holds after writing is what the next read gives back.
func atLeastOne(counter int) int {
	if counter < 1 {
		return 1
	}
	return counter
}

// writeTask inserts a task and all of its children. The caller has already
// removed whatever was there under the same identifier, so this is the
// only place that builds a row out of a task.
func writeTask(tx *sql.Tx, task *model.Task, num int) error {
	task.NextCriterionKey = atLeastOne(task.NextCriterionKey)
	task.NextCommentKey = atLeastOne(task.NextCommentKey)

	var questionAuthor, questionAskedAt, questionBody string
	if task.Question != nil {
		questionAuthor = task.Question.Author
		questionAskedAt = formatInstant(task.Question.AskedAt)
		questionBody = task.Question.Body
	}

	_, err := tx.Exec(`INSERT INTO task (`+taskColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, num, task.Title, task.Status, task.Type, task.Priority, task.Parent,
		task.Author, formatDate(task.Due), task.Ordinal,
		task.Description, task.Plan, task.Notes, task.Summary,
		formatInstant(task.CreatedAt), formatInstant(task.UpdatedAt), task.Archived,
		formatInstant(task.LeaseExpiresAt), task.LeaseHolder,
		questionAuthor, questionAskedAt, questionBody,
		task.NextCriterionKey, task.NextCommentKey,
	)
	if err != nil {
		return err
	}

	for _, field := range model.ListFields() {
		values, err := task.ListField(field)
		if err != nil {
			return err
		}
		for position, value := range values {
			if _, err := tx.Exec(
				"INSERT INTO task_list_item (task_id, field, position, value) VALUES (?, ?, ?, ?)",
				task.ID, string(field), position, value,
			); err != nil {
				return err
			}
		}
	}

	for position, c := range task.AcceptanceCriteria {
		if _, err := tx.Exec(
			"INSERT INTO task_criterion (task_id, key, position, text, checked) VALUES (?, ?, ?, ?, ?)",
			task.ID, c.Key, position, c.Text, c.Checked,
		); err != nil {
			return err
		}
	}

	for position, c := range task.Comments {
		if _, err := tx.Exec(
			"INSERT INTO task_comment (task_id, key, position, author, created_at, body) VALUES (?, ?, ?, ?, ?, ?)",
			task.ID, c.Key, position, c.Author, formatInstant(c.CreatedAt), c.Body,
		); err != nil {
			return err
		}
	}
	return nil
}

// read answers the tasks of the board, or the single one with that id when
// id is not empty, in ascending identifier order, together with the ones it
// could not decode.
//
// It runs four queries and never one per task: the whole point of the
// startup budget of docs/spec/presupuestos.md#el-presupuesto-de-arranque
// is that reading a board of 300 tasks is a handful of scans, not fifteen
// hundred round trips.
//
// A row it cannot turn into a task does not end the read: the task is left
// out of the first list and named in the second, which is the first case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar.
// The error it returns is for the other case, the one where there is no
// readable board at all, and that one does end the read.
func (r *Tasks) read(id string) ([]*model.Task, []Skipped, error) {
	tasks, byID, bad, err := r.readTasks(id)
	if err != nil {
		return nil, nil, err
	}
	if len(tasks) == 0 {
		return nil, r.skippedList(bad), nil
	}
	for _, readChildren := range []func(string, map[string]*model.Task, map[string]*model.Error) error{
		r.readListItems,
		r.readCriteria,
		r.readComments,
	} {
		if err := readChildren(id, byID, bad); err != nil {
			return nil, nil, err
		}
	}
	return withoutTheSkipped(tasks, bad), r.skippedList(bad), nil
}

// withoutTheSkipped answers the tasks that decoded whole, keeping the
// order. A task is dropped when any of its parts failed, and never
// half-filled: half a task read as a whole one is the silence that
// docs/spec/garantias.md forbids as much as aborting.
func withoutTheSkipped(tasks []*model.Task, bad map[string]*model.Error) []*model.Task {
	if len(bad) == 0 {
		return tasks
	}
	kept := make([]*model.Task, 0, len(tasks)-len(bad))
	for _, task := range tasks {
		if bad[task.ID] == nil {
			kept = append(kept, task)
		}
	}
	return kept
}

// skippedList turns the map the five reads fill into the list the caller
// sees, in ascending identifier order like everything else, which is the
// order the warning of docs/spec/garantias.md names them in.
func (r *Tasks) skippedList(bad map[string]*model.Error) []Skipped {
	if len(bad) == 0 {
		return nil
	}
	ids := make([]string, 0, len(bad))
	for id := range bad {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		// By number and not by text, so MYP-9 comes before MYP-10. An
		// identifier this board cannot parse falls back to the text, which
		// keeps the order total whatever ends up stored.
		a, aerr := r.number(ids[i])
		b, berr := r.number(ids[j])
		if aerr == nil && berr == nil {
			return a < b
		}
		return ids[i] < ids[j]
	})

	skipped := make([]Skipped, 0, len(ids))
	for _, id := range ids {
		skipped = append(skipped, Skipped{ID: id, Reason: bad[id]})
	}
	return skipped
}

// undecodable builds the error of a task that cannot be turned into a
// task: exit code 3 and the code undecodable_task of
// docs/spec/contrato-json.md#los-identificadores-de-error.
//
// The specification does not fix its text, only that a targeted read gives
// "el motivo exacto", so the message names the field that could not be
// read and carries the reason underneath.
func undecodable(id, field string, reason error) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "undecodable_task",
		Message:  fmt.Sprintf("%s cannot be read: %s: %v", id, field, reason),
		Field:    field,
	}
}

// query runs one of the five reads, adding the single task filter when
// there is one. column is what that table calls the task's identifier.
func (r *Tasks) query(base, column, id, order string) (*sql.Rows, error) {
	if id == "" {
		return r.store.Query(base + " ORDER BY " + order)
	}
	return r.store.Query(base+" WHERE "+column+" = ? ORDER BY "+order, id)
}

func (r *Tasks) readTasks(id string) ([]*model.Task, map[string]*model.Task, map[string]*model.Error, error) {
	rows, err := r.query("SELECT "+taskColumns+" FROM task", "id", id, "num")
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	byID := map[string]*model.Task{}
	bad := map[string]*model.Error{}
	for rows.Next() {
		var (
			task                                      model.Task
			num                                       int
			due, createdAt, updatedAt, leaseExpiresAt string
			questionAuthor, questionAskedAt           string
			questionBody                              string
		)
		if err := rows.Scan(
			&task.ID, &num, &task.Title, &task.Status, &task.Type, &task.Priority,
			&task.Parent, &task.Author, &due, &task.Ordinal,
			&task.Description, &task.Plan, &task.Notes, &task.Summary,
			&createdAt, &updatedAt, &task.Archived,
			&leaseExpiresAt, &task.LeaseHolder,
			&questionAuthor, &questionAskedAt, &questionBody,
			&task.NextCriterionKey, &task.NextCommentKey,
		); err != nil {
			return nil, nil, nil, r.store.Classify(err)
		}

		// Every date of the row is decoded before the task counts as read.
		// A single one that is not an instant this program wrote makes the
		// task undecodable, and that is a fact about that task alone: the
		// board around it is perfectly readable, so the read goes on.
		for _, date := range []struct {
			field string
			text  string
			into  *time.Time
		}{
			{"due", due, &task.Due},
			{"createdAt", createdAt, &task.CreatedAt},
			{"updatedAt", updatedAt, &task.UpdatedAt},
			{"leaseExpiresAt", leaseExpiresAt, &task.LeaseExpiresAt},
		} {
			parsed, err := parseAnyDate(date.field, date.text)
			if err != nil {
				bad[task.ID] = undecodable(task.ID, date.field, err)
				break
			}
			*date.into = parsed
		}
		// A stored key that does not keep its form is a datum the program
		// cannot interpret, and the rule for one of those is the general
		// one: the task is left out of the read and named in the second
		// list, never a reason to fail
		// (docs/spec/modelo-de-datos/orden-manual.md#una-clave-guardada-que-no-cumple-la-regla).
		if task.Ordinal != "" && !model.ValidOrdinal(task.Ordinal) {
			bad[task.ID] = undecodable(task.ID, "ordinal",
				fmt.Errorf("%q is not an ordinal key: %s", task.Ordinal, model.OrdinalHint))
		}
		if questionAskedAt != "" || questionAuthor != "" || questionBody != "" {
			askedAt, err := parseInstant(questionAskedAt)
			if err != nil {
				bad[task.ID] = undecodable(task.ID, "question.askedAt", err)
			}
			task.Question = &model.Question{
				Author:  questionAuthor,
				AskedAt: askedAt,
				Body:    questionBody,
			}
		}

		tasks = append(tasks, &task)
		byID[task.ID] = &task
	}
	return tasks, byID, bad, r.store.Classify(rows.Err())
}

// parseAnyDate reads either shape of date a task row carries: `due` is a
// calendar day and the three instants are instants.
func parseAnyDate(field, text string) (time.Time, error) {
	if field == "due" {
		return parseDate(text)
	}
	return parseInstant(text)
}

func (r *Tasks) readListItems(id string, byID map[string]*model.Task, bad map[string]*model.Error) error {
	rows, err := r.query(
		"SELECT task_id, field, value FROM task_list_item", "task_id", id,
		"task_id, field, position",
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID, field, value string
		if err := rows.Scan(&taskID, &field, &value); err != nil {
			return r.store.Classify(err)
		}
		task := byID[taskID]
		if task == nil {
			continue
		}
		// A field name the model does not know is never a silently dropped
		// value: it makes that task undecodable, the same rule on the way
		// in and on the way out.
		current, err := task.ListField(model.ListField(field))
		if err != nil {
			bad[taskID] = undecodable(taskID, field, err)
			continue
		}
		if err := task.SetListField(model.ListField(field), append(current, value)); err != nil {
			bad[taskID] = undecodable(taskID, field, err)
		}
	}
	return r.store.Classify(rows.Err())
}

func (r *Tasks) readCriteria(id string, byID map[string]*model.Task, bad map[string]*model.Error) error {
	rows, err := r.query(
		"SELECT task_id, key, text, checked FROM task_criterion", "task_id", id,
		"task_id, position",
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID string
		var c model.Criterion
		if err := rows.Scan(&taskID, &c.Key, &c.Text, &c.Checked); err != nil {
			return r.store.Classify(err)
		}
		if task := byID[taskID]; task != nil {
			task.AcceptanceCriteria = append(task.AcceptanceCriteria, c)
		}
	}
	return r.store.Classify(rows.Err())
}

func (r *Tasks) readComments(id string, byID map[string]*model.Task, bad map[string]*model.Error) error {
	rows, err := r.query(
		"SELECT task_id, key, author, created_at, body FROM task_comment", "task_id", id,
		"task_id, position",
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID, createdAt string
		var c model.Comment
		if err := rows.Scan(&taskID, &c.Key, &c.Author, &createdAt, &c.Body); err != nil {
			return r.store.Classify(err)
		}
		at, err := parseInstant(createdAt)
		if err != nil {
			bad[taskID] = undecodable(taskID, fmt.Sprintf("comment #%d createdAt", c.Key), err)
			continue
		}
		c.CreatedAt = at
		if task := byID[taskID]; task != nil {
			task.Comments = append(task.Comments, c)
		}
	}
	return r.store.Classify(rows.Err())
}

// formatInstant writes an instant the way
// docs/spec/contrato-json.md#números-fechas-y-ausencias fixes it: UTC,
// second precision, ending in Z. The zero time is no instant at all and is
// written as the empty string.
func formatInstant(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Truncate(time.Second).Format(model.InstantLayout)
}

// parseInstant is formatInstant backwards.
func parseInstant(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(model.InstantLayout, s)
}

// formatDate writes a calendar day as YYYY-MM-DD, the shape `due` travels
// in, with no time of day and no zone.
func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(model.DateLayout)
}

// parseDate is formatDate backwards, and gives back a day at UTC midnight.
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation(model.DateLayout, s, time.UTC)
}
