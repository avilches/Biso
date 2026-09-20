// Package board is where a board's rows become tasks and a task becomes
// rows. It is the only layer that knows both the logical model of
// docs/spec/modelo-de-datos/index.md and the tables internal/store keeps
// them in, which is the split the architecture asks for: the store knows
// rows and columns and nothing about a Task, and the model knows tasks and
// nothing about SQLite.
//
// Resolving where a board lives and loading its config.json belong here
// too, in the Board type that `biso init` and `biso where` bring with
// them. What exists today is the half TASK-10 needs: the task repository.
package board

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"biso/internal/model"
	"biso/internal/store"
)

// Tasks reads and writes the tasks of one board.
//
// prefix is the board's task_prefix
// (docs/spec/modelo-de-datos/identificadores.md), already decided: it is
// derived from the board's name when the board is created, and by the time
// anything reads or writes a task it is just configuration.
//
// extensions is the `extensions` list the board declares, the closed
// vocabulary an `ext` key is checked against
// (docs/spec/modelo-de-datos/campos-externos.md).
type Tasks struct {
	store      *store.Store
	prefix     string
	extensions []string
}

// NewTasks builds the repository over an open store.
func NewTasks(s *store.Store, prefix string, extensions []string) *Tasks {
	return &Tasks{store: s, prefix: prefix, extensions: extensions}
}

// Create allocates the task's identifier and writes it, both inside the
// same write transaction.
//
// That is what makes guarantee 3 of docs/spec/garantias.md true: WithTx
// takes the write lock at its BEGIN, so the counter cannot be read by two
// processes before either of them increments it, however many copies of
// the project they work from. The counter only ever grows, which is also
// what keeps an identifier from being reused after a task is archived or
// after its row disappears.
//
// The task's dates follow docs/spec/modelo-de-datos/fechas.md: the ones it
// already carries are kept, which is what an import needs, and the ones it
// does not are taken from the clock, in UTC and to the second.
func (r *Tasks) Create(task *model.Task) error {
	if err := r.validate(task); err != nil {
		return err
	}

	now := time.Now().UTC().Truncate(time.Second)
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = now
	}

	return r.store.WithTx(func(tx *sql.Tx) error {
		num, err := allocate(tx)
		if err != nil {
			return err
		}
		task.ID = fmt.Sprintf("%s-%d", r.prefix, num)
		return writeTask(tx, task, num)
	})
}

// CreateAll writes a whole batch of new tasks inside one transaction,
// which is what makes the all-or-nothing promise of
// docs/spec/cmd/new.md#el-modo-lote real: a file of two hundred lines
// either leaves two hundred tasks or leaves none.
//
// A task that already carries an identifier keeps it, and the board's
// counter is moved above the highest one imported, so the board never
// hands out an identifier the batch has just reserved. A task without one
// is allocated the next number, after that move, which is why an explicit
// identifier and an implicit one can never collide inside the same call.
//
// Whether an identifier is free is the caller's question and not this
// one's: the batch judges the whole file before anything is written, and
// answering it again here would answer it a second time with a worse
// message.
func (r *Tasks) CreateAll(tasks []*model.Task) error {
	nums := make([]int, len(tasks))
	highest := 0
	for i, task := range tasks {
		if err := r.validate(task); err != nil {
			return err
		}
		if task.ID == "" {
			continue
		}
		num, err := r.number(task.ID)
		if err != nil {
			return err
		}
		nums[i] = num
		if num > highest {
			highest = num
		}
	}

	now := time.Now().UTC().Truncate(time.Second)
	for _, task := range tasks {
		if task.CreatedAt.IsZero() {
			task.CreatedAt = now
		}
		if task.UpdatedAt.IsZero() {
			task.UpdatedAt = now
		}
	}

	return r.store.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(
			"UPDATE board_counter SET last_task_num = ? WHERE id = 1 AND last_task_num < ?",
			highest, highest,
		); err != nil {
			return err
		}
		for i, task := range tasks {
			if task.ID == "" {
				num, err := allocate(tx)
				if err != nil {
					return err
				}
				task.ID = fmt.Sprintf("%s-%d", r.prefix, num)
				nums[i] = num
			}
			if err := writeTask(tx, task, nums[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// Save writes an existing task over itself.
//
// It enforces the two immutabilities of
// docs/spec/modelo-de-datos/index.md that it is in a position to enforce.
// The task has to be there, so saving never brings back an identifier
// whose row is gone; and createdAt keeps the value the board already has,
// whatever the task in memory carries, because no flag of the program
// changes it and only an import can set it, which goes through Create.
//
// It does not touch updatedAt, in contrast: whether a call changed
// anything is what decides that, and only the caller knows.
func (r *Tasks) Save(task *model.Task) error {
	return r.SaveAll([]*model.Task{task})
}

// SaveAll writes several existing tasks over themselves, all of them inside
// one single transaction.
//
// That one transaction is what makes guarantee 2 of
// docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables
// observable: `biso set A B C` where the third task turns out not to be
// writable leaves A and B exactly as they were, and a concurrent reader
// never sees the board halfway through. Everything that can be judged
// without touching the database is judged first, over every task, so that a
// call that is going to fail does not even open one.
func (r *Tasks) SaveAll(tasks []*model.Task) error {
	nums := make([]int, len(tasks))
	for i, task := range tasks {
		if err := r.validate(task); err != nil {
			return err
		}
		num, err := r.number(task.ID)
		if err != nil {
			return err
		}
		nums[i] = num
	}

	return r.store.WithTx(func(tx *sql.Tx) error {
		for i, task := range tasks {
			if err := r.saveWithin(tx, task, nums[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

// saveWithin writes one task inside a transaction the caller owns.
func (r *Tasks) saveWithin(tx *sql.Tx, task *model.Task, num int) error {
	var createdAt string
	err := tx.QueryRow("SELECT created_at FROM task WHERE id = ?", task.ID).Scan(&createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		var last int
		if err := tx.QueryRow("SELECT last_task_num FROM board_counter WHERE id = 1").Scan(&last); err != nil {
			return err
		}
		return r.missing(task.ID, num, last)
	}
	if err != nil {
		return err
	}
	if task.CreatedAt, err = parseInstant(createdAt); err != nil {
		return fmt.Errorf("%s: createdAt: %w", task.ID, err)
	}

	// The children go with it: every one of their tables declares the
	// task as a foreign key with ON DELETE CASCADE, and the connection
	// runs with foreign keys on.
	if _, err := tx.Exec("DELETE FROM task WHERE id = ?", task.ID); err != nil {
		return err
	}
	return writeTask(tx, task, num)
}

// LastAllocated answers the highest task number the board has ever
// assigned, which is the number the two identifier messages of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro quote and
// the one `biso doctor` reports. It is zero on a board with no task yet.
func (r *Tasks) LastAllocated() (int, error) {
	rows, err := r.store.Query("SELECT last_task_num FROM board_counter WHERE id = 1")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("the board counter row is missing")
	}
	var last int
	if err := rows.Scan(&last); err != nil {
		return 0, err
	}
	return last, rows.Err()
}

// Load answers one task by its identifier.
//
// When there is no such task it tells the two cases of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro apart,
// which is exactly what the counter is for: an identifier above the
// highest one ever assigned never existed, and one below it was assigned
// at some point and is gone, which only happens when something outside
// biso touched the data.
//
// It is a targeted read, so the other half of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// applies: a task that is there and cannot be decoded is exit code 3 with
// the reason, and not the skip of a set read. There is nothing else to
// answer with when the caller asked for that one task.
func (r *Tasks) Load(id string) (*model.Task, error) {
	num, err := r.number(id)
	if err != nil {
		return nil, err
	}

	tasks, skipped, err := r.read(id)
	if err != nil {
		return nil, err
	}
	if len(skipped) == 1 {
		return nil, skipped[0].Reason
	}
	if len(tasks) == 1 {
		return tasks[0], nil
	}

	last, err := r.LastAllocated()
	if err != nil {
		return nil, err
	}
	return nil, r.missing(id, num, last)
}

// missing builds the error for an identifier the board does not have,
// choosing between the two of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro by
// comparing it with the highest one ever assigned.
func (r *Tasks) missing(id string, num, last int) *model.Error {
	if num > last {
		return &model.Error{
			ExitCode: 4,
			Code:     "never_allocated",
			Message:  fmt.Sprintf("%s has never existed on this board", id),
			Notes: []string{fmt.Sprintf(
				"the highest id ever assigned here is %s-%d", r.prefix, last,
			)},
		}
	}
	return &model.Error{
		ExitCode: 4,
		Code:     "not_found",
		Message:  fmt.Sprintf("%s is not on this board", id),
		Notes: []string{fmt.Sprintf(
			"%s was assigned at some point, but this board's current data does not have it", id,
		)},
		Hints: []string{
			"this only happens when something outside biso touched the data, such as a " +
				"snapshot restored over a newer one or a database edited by hand; " +
				"`biso doctor` diagnoses damage to the board",
		},
	}
}

// Skipped is one task a set read could not decode, with the reason.
//
// It exists because docs/spec/garantias.md asks a set read for two things
// at once: never to abort because of one bad task, and never to hide it.
// Answering only the tasks that could be read would meet the first and
// break the second, and a listing missing a task would be read as a fact
// about the board.
type Skipped struct {
	ID     string
	Reason *model.Error
}

// All answers every task of the board, archived ones included, in
// ascending identifier order: the same tie-break every listing of
// docs/spec/cmd/ls.md uses, and the order of the number and not of the
// text, so MYP-9 comes before MYP-10.
//
// It is a set read in the sense of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// so a task it cannot decode is left out of the first list and named in
// the second, never a reason to fail: the warning that names them, and the
// exit code 6 that `biso export` and `biso snapshot` answer when the
// second list is not empty, belong to the commands that call this.
func (r *Tasks) All() ([]*model.Task, []Skipped, error) {
	return r.read("")
}

// number answers the <n> of a <prefix>-<n> identifier.
//
// It is not the reference grammar of docs/spec/referencias.md, which takes
// what a caller typed, in several shapes, and is the job of the layer that
// resolves references. This one takes an identifier the program itself
// built or read, so anything else is a bug and not a case of the
// specification.
func (r *Tasks) number(id string) (int, error) {
	rest, ok := strings.CutPrefix(id, r.prefix+"-")
	if !ok {
		return 0, fmt.Errorf("task id %q does not belong to board prefix %q", id, r.prefix)
	}
	num, err := strconv.Atoi(rest)
	if err != nil || num < 1 {
		return 0, fmt.Errorf("task id %q does not end in a positive number", id)
	}
	return num, nil
}

// validate rejects what the board must not store, before anything is
// written and therefore before an identifier is spent.
//
// The rules themselves are the model's, in Task.Validate: they are rules
// about a task and not about rows, so they hold whether the task came from
// a flag or from an import. The board only supplies the one thing the
// model cannot know, the `extensions` list this board declares. The
// vocabularies that a board configures, statuses, types and priorities,
// are matched by internal/match and checked by the layer above this one.
func (r *Tasks) validate(task *model.Task) error {
	return task.Validate(r.extensions)
}

// allocate moves the board's counter one forward and answers the number it
// reaches. Both statements run inside the caller's write transaction, so
// no two processes can read the same value.
func allocate(tx *sql.Tx) (int, error) {
	if _, err := tx.Exec("UPDATE board_counter SET last_task_num = last_task_num + 1 WHERE id = 1"); err != nil {
		return 0, err
	}
	var num int
	if err := tx.QueryRow("SELECT last_task_num FROM board_counter WHERE id = 1").Scan(&num); err != nil {
		return 0, err
	}
	return num, nil
}
