package board

import (
	"database/sql"
	"fmt"
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

// writeTask inserts a task and all of its children. The caller has already
// removed whatever was there under the same identifier, so this is the
// only place that builds a row out of a task.
func writeTask(tx *sql.Tx, task *model.Task, num int) error {
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

	for _, key := range sortedKeys(task.Ext) {
		if _, err := tx.Exec(
			"INSERT INTO task_ext (task_id, key, value) VALUES (?, ?, ?)",
			task.ID, key, task.Ext[key],
		); err != nil {
			return err
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
// id is not empty, in ascending identifier order.
//
// It runs five queries and never one per task: the whole point of the
// startup budget of docs/spec/presupuestos.md#el-presupuesto-de-arranque
// is that reading a board of 300 tasks is a handful of scans, not fifteen
// hundred round trips.
func (r *Tasks) read(id string) ([]*model.Task, error) {
	tasks, byID, err := r.readTasks(id)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	if err := r.readListItems(id, byID); err != nil {
		return nil, err
	}
	if err := r.readExt(id, byID); err != nil {
		return nil, err
	}
	if err := r.readCriteria(id, byID); err != nil {
		return nil, err
	}
	if err := r.readComments(id, byID); err != nil {
		return nil, err
	}
	return tasks, nil
}

// query runs one of the five reads, adding the single task filter when
// there is one. column is what that table calls the task's identifier.
func (r *Tasks) query(base, column, id, order string) (*sql.Rows, error) {
	if id == "" {
		return r.store.Query(base + " ORDER BY " + order)
	}
	return r.store.Query(base+" WHERE "+column+" = ? ORDER BY "+order, id)
}

func (r *Tasks) readTasks(id string) ([]*model.Task, map[string]*model.Task, error) {
	rows, err := r.query("SELECT "+taskColumns+" FROM task", "id", id, "num")
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var tasks []*model.Task
	byID := map[string]*model.Task{}
	for rows.Next() {
		var (
			task                                      model.Task
			num                                       int
			due, createdAt, updatedAt, leaseExpiresAt string
			ordinal                                   sql.NullInt64
			questionAuthor, questionAskedAt           string
			questionBody                              string
		)
		if err := rows.Scan(
			&task.ID, &num, &task.Title, &task.Status, &task.Type, &task.Priority,
			&task.Parent, &task.Author, &due, &ordinal,
			&task.Description, &task.Plan, &task.Notes, &task.Summary,
			&createdAt, &updatedAt, &task.Archived,
			&leaseExpiresAt, &task.LeaseHolder,
			&questionAuthor, &questionAskedAt, &questionBody,
			&task.NextCriterionKey, &task.NextCommentKey,
		); err != nil {
			return nil, nil, err
		}

		if task.Due, err = parseDate(due); err != nil {
			return nil, nil, fmt.Errorf("%s: due: %w", task.ID, err)
		}
		if task.CreatedAt, err = parseInstant(createdAt); err != nil {
			return nil, nil, fmt.Errorf("%s: createdAt: %w", task.ID, err)
		}
		if task.UpdatedAt, err = parseInstant(updatedAt); err != nil {
			return nil, nil, fmt.Errorf("%s: updatedAt: %w", task.ID, err)
		}
		if task.LeaseExpiresAt, err = parseInstant(leaseExpiresAt); err != nil {
			return nil, nil, fmt.Errorf("%s: leaseExpiresAt: %w", task.ID, err)
		}
		if ordinal.Valid {
			value := int(ordinal.Int64)
			task.Ordinal = &value
		}
		if questionAskedAt != "" || questionAuthor != "" || questionBody != "" {
			askedAt, err := parseInstant(questionAskedAt)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: question.askedAt: %w", task.ID, err)
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
	return tasks, byID, rows.Err()
}

func (r *Tasks) readListItems(id string, byID map[string]*model.Task) error {
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
			return err
		}
		task := byID[taskID]
		if task == nil {
			continue
		}
		// A field name the model does not know is an error and never a
		// silently dropped value, which is the same rule on the way in and
		// on the way out.
		current, err := task.ListField(model.ListField(field))
		if err != nil {
			return fmt.Errorf("%s: %w", taskID, err)
		}
		if err := task.SetListField(model.ListField(field), append(current, value)); err != nil {
			return fmt.Errorf("%s: %w", taskID, err)
		}
	}
	return rows.Err()
}

func (r *Tasks) readExt(id string, byID map[string]*model.Task) error {
	rows, err := r.query("SELECT task_id, key, value FROM task_ext", "task_id", id, "task_id, key")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID, key, value string
		if err := rows.Scan(&taskID, &key, &value); err != nil {
			return err
		}
		task := byID[taskID]
		if task == nil {
			continue
		}
		if task.Ext == nil {
			task.Ext = map[string]string{}
		}
		task.Ext[key] = value
	}
	return rows.Err()
}

func (r *Tasks) readCriteria(id string, byID map[string]*model.Task) error {
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
			return err
		}
		if task := byID[taskID]; task != nil {
			task.AcceptanceCriteria = append(task.AcceptanceCriteria, c)
		}
	}
	return rows.Err()
}

func (r *Tasks) readComments(id string, byID map[string]*model.Task) error {
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
			return err
		}
		at, err := parseInstant(createdAt)
		if err != nil {
			return fmt.Errorf("%s: comment #%d: %w", taskID, c.Key, err)
		}
		c.CreatedAt = at
		if task := byID[taskID]; task != nil {
			task.Comments = append(task.Comments, c)
		}
	}
	return rows.Err()
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
