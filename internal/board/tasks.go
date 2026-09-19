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
	"fmt"
	"sort"
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

// Save writes an existing task over itself. It does not touch updatedAt:
// whether a call changed anything is what decides that
// (docs/spec/modelo-de-datos/index.md), and only the caller knows.
func (r *Tasks) Save(task *model.Task) error {
	if err := r.validate(task); err != nil {
		return err
	}
	num, err := r.number(task.ID)
	if err != nil {
		return err
	}

	return r.store.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec("DELETE FROM task WHERE id = ?", task.ID); err != nil {
			return err
		}
		return writeTask(tx, task, num)
	})
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
func (r *Tasks) Load(id string) (*model.Task, error) {
	num, err := r.number(id)
	if err != nil {
		return nil, err
	}

	tasks, err := r.read(id)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 1 {
		return tasks[0], nil
	}

	last, err := r.LastAllocated()
	if err != nil {
		return nil, err
	}
	if num > last {
		return nil, &model.Error{
			ExitCode: 4,
			Code:     "never_allocated",
			Message:  fmt.Sprintf("%s has never existed on this board", id),
			Notes: []string{fmt.Sprintf(
				"the highest id ever assigned here is %s-%d", r.prefix, last,
			)},
		}
	}
	return nil, &model.Error{
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

// All answers every task of the board, archived ones included, in
// ascending identifier order: the same tie-break every listing of
// docs/spec/cmd/ls.md uses, and the order of the number and not of the
// text, so MYP-9 comes before MYP-10.
func (r *Tasks) All() ([]*model.Task, error) {
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

// validate rejects what the board cannot store, before anything is
// written and therefore before an identifier is spent.
//
// Today that is the `ext` keys, the one closed vocabulary the model owns
// on its own: a key outside the token alphabet is a problem of form (exit
// code 2) and a well formed key the board does not declare is an unknown
// value (exit code 3). The vocabularies that a board configures, statuses,
// types and priorities, are matched by internal/match and checked by the
// layer above this one.
func (r *Tasks) validate(task *model.Task) error {
	for _, key := range sortedKeys(task.Ext) {
		if err := model.ValidateExtensionKeySyntax(key); err != nil {
			return err
		}
		if err := model.ValidateExtensionKey(key, r.extensions); err != nil {
			return err
		}
	}
	return nil
}

// sortedKeys answers a map's keys in a fixed order, so that a task with
// two bad extension keys always fails on the same one.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
