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
// Ordinal is the one field that needs a real NULL, because 0 is a value.

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

	var ordinal any
	if task.Ordinal != nil {
		ordinal = *task.Ordinal
	}
	var questionAuthor, questionAskedAt, questionBody string
	if task.Question != nil {
		questionAuthor = task.Question.Author
		questionAskedAt = formatInstant(task.Question.AskedAt)
		questionBody = task.Question.Body
	}

	_, err := tx.Exec(`INSERT INTO task (`+taskColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID, num, task.Title, task.Status, task.Type, task.Priority, task.Parent,
		task.Author, formatDate(task.Due), ordinal,
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

// undecodable builds the error of a task, or of one of its parts, that
// could not be turned into what the model expects: exit code 3 and the
// code undecodable_task of
// docs/spec/contrato-json.md#los-identificadores-de-error.
//
// given is what was stored, the empty string included, and reason is the
// whole sentence that follows "cannot be read: " on stderr: the exact
// wording docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible fixes
// for a date, and free text for everything else, but never the standard
// library's own message, which is not part of the contract.
func undecodable(id, field, given, reason string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "undecodable_task",
		Message:  fmt.Sprintf("%s cannot be read: %s", id, reason),
		Field:    field,
		Given:    given,
	}
}

// dateReason is the two sentences of
// docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible for a date
// field: `due` is a calendar day and the three instants are instants, and
// an empty value where the field is required reads the same as any other
// value that is not one.
func dateReason(field, given string, day bool) string {
	kind := "an instant (YYYY-MM-DDTHH:MM:SSZ)"
	if day {
		kind = "a calendar day (YYYY-MM-DD)"
	}
	return fmt.Sprintf("%s is not %s: %q", field, kind, given)
}

// notAWholeNumber and notZeroOrOne are the two shapes of the last row of
// docs/spec/garantias.md#qué-se-comprueba: a column the model reads as an
// integer or a boolean that the database does not hold as one. They never
// abort the read: the row is turned into `bad` for that one task, the same
// way a bad date or a bad vocabulary value is, and the rest of the board
// stays reachable.
func notAWholeNumber(field, given string) string {
	return fmt.Sprintf("%s is not a whole number: %q", field, given)
}

func notZeroOrOne(field, given string) string {
	return fmt.Sprintf("%s is not 0 or 1: %q", field, given)
}

// rawText is what undecodable's given carries for a column that was
// scanned as `any` because its declared type does not rule out a stored
// value of another kind: the same text `%v` would print, and the empty
// string for a column that came back NULL.
func rawText(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// scannedInt turns a column scanned into `any` into the whole number the
// model needs: an `int64` from the driver, and nothing else. A value
// SQLite's flexible typing let through as text or as a real number is not
// one, which is exactly what makes the column undecodable and not merely
// unusual.
func scannedInt(v any) (int, bool) {
	n, ok := v.(int64)
	return int(n), ok
}

// scannedBool is scannedInt for the model's booleans, which the schema
// only ever writes as the `int64` 0 or 1: anything else, including a
// number that is not one of the two, is not a boolean this program wrote.
func scannedBool(v any) (bool, bool) {
	n, ok := v.(int64)
	if !ok || (n != 0 && n != 1) {
		return false, false
	}
	return n == 1, true
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
			ordinalRaw, archivedRaw                   any
			nextCriterionRaw, nextCommentRaw          any
			questionAuthor, questionAskedAt           string
			questionBody                              string
		)
		if err := rows.Scan(
			&task.ID, &num, &task.Title, &task.Status, &task.Type, &task.Priority,
			&task.Parent, &task.Author, &due, &ordinalRaw,
			&task.Description, &task.Plan, &task.Notes, &task.Summary,
			&createdAt, &updatedAt, &archivedRaw,
			&leaseExpiresAt, &task.LeaseHolder,
			&questionAuthor, &questionAskedAt, &questionBody,
			&nextCriterionRaw, &nextCommentRaw,
		); err != nil {
			return nil, nil, nil, r.store.Classify(err)
		}

		// mark is the first reason found that this row does not decode.
		// Only the first is kept: docs/spec/garantias.md asks for the
		// motivo exacto of one fault, not an exhaustive list, and every
		// caller of it below only runs when the field it is about has not
		// already been read into the task.
		mark := func(field, given, reason string) {
			if bad[task.ID] == nil {
				bad[task.ID] = undecodable(task.ID, field, given, reason)
			}
		}

		// Every date of the row is decoded before the task counts as read.
		// `due` and `leaseExpiresAt` may be empty, `createdAt` and
		// `updatedAt` may not
		// (docs/spec/modelo-de-datos/fechas.md#una-fecha-guardada-que-no-es-una-fecha).
		// A field that fails this does not end the read: the task is left
		// for `bad` to carry, and the board around it stays reachable.
		for _, date := range []struct {
			field    string
			text     string
			into     *time.Time
			day      bool
			required bool
		}{
			{"due", due, &task.Due, true, false},
			{"createdAt", createdAt, &task.CreatedAt, false, true},
			{"updatedAt", updatedAt, &task.UpdatedAt, false, true},
			{"leaseExpiresAt", leaseExpiresAt, &task.LeaseExpiresAt, false, false},
		} {
			if date.text == "" {
				if !date.required {
					continue
				}
				mark(date.field, "", dateReason(date.field, "", date.day))
				break
			}
			parsed, err := parseRequiredDate(date.text, date.day)
			if err != nil {
				mark(date.field, date.text, dateReason(date.field, date.text, date.day))
				break
			}
			*date.into = parsed
		}

		if ordinalRaw != nil {
			if value, ok := scannedInt(ordinalRaw); ok {
				task.Ordinal = &value
			} else {
				mark("ordinal", rawText(ordinalRaw), notAWholeNumber("ordinal", rawText(ordinalRaw)))
			}
		}
		if archived, ok := scannedBool(archivedRaw); ok {
			task.Archived = archived
		} else {
			mark("archived", rawText(archivedRaw), notZeroOrOne("archived", rawText(archivedRaw)))
		}
		if key, ok := scannedInt(nextCriterionRaw); ok {
			task.NextCriterionKey = key
		} else {
			mark("next_criterion_key", rawText(nextCriterionRaw),
				notAWholeNumber("next_criterion_key", rawText(nextCriterionRaw)))
		}
		if key, ok := scannedInt(nextCommentRaw); ok {
			task.NextCommentKey = key
		} else {
			mark("next_comment_key", rawText(nextCommentRaw),
				notAWholeNumber("next_comment_key", rawText(nextCommentRaw)))
		}

		if questionAskedAt != "" || questionAuthor != "" || questionBody != "" {
			task.Question = &model.Question{Author: questionAuthor, Body: questionBody}
			if questionAskedAt == "" {
				mark("question.askedAt", "", dateReason("question.askedAt", "", false))
			} else if askedAt, err := parseInstant(questionAskedAt); err != nil {
				mark("question.askedAt", questionAskedAt, dateReason("question.askedAt", questionAskedAt, false))
			} else {
				task.Question.AskedAt = askedAt
			}
		}

		tasks = append(tasks, &task)
		byID[task.ID] = &task
	}
	return tasks, byID, bad, r.store.Classify(rows.Err())
}

// parseRequiredDate parses a date field that is known not to be empty:
// `due` is a calendar day and every other field is an instant.
func parseRequiredDate(text string, day bool) (time.Time, error) {
	if day {
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
		// in and on the way out. It has no fixed message, so given is the
		// name itself, the same thing the field names.
		current, err := task.ListField(model.ListField(field))
		if err != nil {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, field, field, err.Error())
			}
			continue
		}
		if err := task.SetListField(model.ListField(field), append(current, value)); err != nil {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, field, field, err.Error())
			}
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
		var taskID, text string
		var keyRaw, checkedRaw any
		if err := rows.Scan(&taskID, &keyRaw, &text, &checkedRaw); err != nil {
			return r.store.Classify(err)
		}
		key, keyOK := scannedInt(keyRaw)
		if !keyOK {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, "criterion key", rawText(keyRaw),
					notAWholeNumber("criterion key", rawText(keyRaw)))
			}
			continue
		}
		checked, checkedOK := scannedBool(checkedRaw)
		if !checkedOK {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, "criterion checked", rawText(checkedRaw),
					notZeroOrOne("criterion checked", rawText(checkedRaw)))
			}
			continue
		}
		if task := byID[taskID]; task != nil {
			task.AcceptanceCriteria = append(task.AcceptanceCriteria,
				model.Criterion{Key: key, Text: text, Checked: checked})
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
		var taskID, author, createdAt, body string
		var keyRaw any
		if err := rows.Scan(&taskID, &keyRaw, &author, &createdAt, &body); err != nil {
			return r.store.Classify(err)
		}
		key, keyOK := scannedInt(keyRaw)
		if !keyOK {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, "comment key", rawText(keyRaw),
					notAWholeNumber("comment key", rawText(keyRaw)))
			}
			continue
		}
		field := fmt.Sprintf("comment #%d createdAt", key)
		if createdAt == "" {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, field, "", dateReason(field, "", false))
			}
			continue
		}
		at, err := parseInstant(createdAt)
		if err != nil {
			if bad[taskID] == nil {
				bad[taskID] = undecodable(taskID, field, createdAt, dateReason(field, createdAt, false))
			}
			continue
		}
		if task := byID[taskID]; task != nil {
			task.Comments = append(task.Comments, model.Comment{
				Key: key, Author: author, CreatedAt: at, Body: body,
			})
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
