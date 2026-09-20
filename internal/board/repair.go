package board

import "database/sql"

// This file is the one write `biso doctor --fix` makes to a board's data,
// and it is here and not in internal/ops for the same reason every other
// write is: this package is the only one that knows the tables.

// Repairs are the data repairs of one `biso doctor --fix` call: the tasks
// whose lease fields are to be emptied, and the highest identifier the
// board should remember having handed out.
//
// They travel together because they go in one transaction, all or nothing
// (docs/spec/cmd/doctor.md#atomicidad-de---fix-con-varias-reparaciones).
// The <id>.id marker is not here: it is a file, and a file cannot be
// written inside a transaction of SQLite.
type Repairs struct {
	// ClearLease are the identifiers whose leaseExpiresAt and leaseHolder
	// are both to be emptied.
	ClearLease []string
	// HighestID is the value the board's counter should hold, and zero
	// when the counter needs no repair.
	HighestID int
}

// Empty answers whether there is anything to write.
func (r Repairs) Empty() bool { return len(r.ClearLease) == 0 && r.HighestID == 0 }

// Repair applies them, in one transaction.
func (b *Board) Repair(r Repairs) error {
	if r.Empty() {
		return nil
	}
	return b.Store.WithTx(func(tx *sql.Tx) error {
		for _, id := range r.ClearLease {
			if _, err := tx.Exec(
				`UPDATE task SET lease_expires_at = '', lease_holder = '' WHERE id = ?`,
				id); err != nil {
				return err
			}
		}
		if r.HighestID > 0 {
			if _, err := tx.Exec(
				`UPDATE board_counter SET last_task_num = ? WHERE id = 1`,
				r.HighestID); err != nil {
				return err
			}
		}
		return nil
	})
}

// DuplicateIDs answers the identifiers more than one row of the task table
// carries, and how many rows each one has.
//
// Today's schema makes the task's identifier its primary key, so this can
// only answer nothing; `biso doctor` asks anyway, because the question it
// is answering is whether the data is sound and not whether this program
// wrote it (docs/spec/cmd/doctor.md#qué-comprueba).
func (b *Board) DuplicateIDs() (map[string]int, error) {
	return b.groupCount(`SELECT id, count(*) FROM task GROUP BY id HAVING count(*) > 1`)
}

// DuplicateCriterionKeys answers, per task, a criterion key that more than
// one criterion of that task carries. It is unreachable for the same reason
// DuplicateIDs is, and asked for the same reason.
func (b *Board) DuplicateCriterionKeys() (map[string]int, error) {
	return b.groupCount(
		`SELECT task_id || ' ' || key, count(*) FROM task_criterion
		 GROUP BY task_id, key HAVING count(*) > 1`)
}

func (b *Board) groupCount(query string) (map[string]int, error) {
	rows, err := b.Store.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var key string
		var n int
		if err := rows.Scan(&key, &n); err != nil {
			return nil, err
		}
		out[key] = n
	}
	return out, rows.Err()
}
