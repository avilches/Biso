package board

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"biso/internal/model"
	"biso/internal/store"
)

// Board is one board, open: where it is, its configuration already loaded,
// and the store already opened over it. It is the first argument of every
// function of internal/ops, and it holds no logic of any command.

// Board joins the three things the architecture asks of this type.
type Board struct {
	Location *Location
	Config   Config
	Machine  Machine
	Store    *store.Store
	Tasks    *Tasks
}

// Open opens the board a resolution found: the store over its database, its
// identity and its configuration.
//
// The identity is read from the database when the directory carried no
// marker to read it from. The two hold the same id on purpose, and comparing
// them is `biso doctor`'s job, not this one's: the first way of choosing a
// board reads no marker at all, precisely so that a board missing one can
// still be opened and repaired (docs/spec/resolucion-del-tablero.md#el-orden-de-búsqueda).
func Open(loc *Location, m Machine) (*Board, error) {
	s, err := store.Open(loc.ID, filepath.Join(loc.Dir, DatabaseFile))
	if err != nil {
		return nil, err
	}
	if loc.ID == "" {
		id, err := ReadIdentity(s)
		if err != nil {
			s.Close()
			return nil, err
		}
		loc.ID = id
	}
	cfg, err := ReadConfig(s)
	if err != nil {
		s.Close()
		return nil, err
	}
	return &Board{
		Location: loc,
		Config:   cfg,
		Machine:  m,
		Store:    s,
		Tasks:    NewTasks(s, cfg.TaskPrefix, cfg.Extensions),
	}, nil
}

// Create makes the board directory if it is not there, creates its database
// and writes its identity and its configuration, both inside the same
// transaction. It does not write the marker or the exclusion file: those are
// files of the directory and not of the store, and `biso init` writes them
// itself (docs/spec/cmd/init.md).
func Create(dir, id string, cfg Config, m Machine) (*Board, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, cannotWrite(dir, err)
	}
	// A database that was not there and does not open is the directory
	// refusing to be written to, not a damaged board: the file the store
	// would be complaining about did not exist a moment ago. So this one
	// path answers exit code 8 and not the 21 of
	// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
	// which is what the case table of docs/spec/cmd/init.md promises for an
	// --at that cannot be written.
	fresh := !HasDatabase(dir)
	s, err := store.Open(id, filepath.Join(dir, DatabaseFile))
	if err != nil {
		if e, ok := err.(*model.Error); ok && fresh && e.ExitCode == 21 {
			return nil, cannotWrite(filepath.Join(dir, DatabaseFile), errors.New("it could not be created"))
		}
		return nil, err
	}
	if err := s.WithTx(func(tx *sql.Tx) error {
		if err := WriteIdentity(tx, id); err != nil {
			return err
		}
		return WriteConfig(tx, cfg)
	}); err != nil {
		s.Close()
		return nil, err
	}
	return &Board{
		Location: &Location{ID: id, Dir: dir, Way: WayWorkingDirectory},
		Config:   cfg,
		Machine:  m,
		Store:    s,
		Tasks:    NewTasks(s, cfg.TaskPrefix, cfg.Extensions),
	}, nil
}

// Close closes the store.
func (b *Board) Close() error { return b.Store.Close() }

// Rewrite replaces the board's configuration, which is what
// `biso init --overwrite-config` does, and never touches a task.
func (b *Board) Rewrite(cfg Config) error {
	if err := b.Store.WithTx(func(tx *sql.Tx) error {
		return WriteConfig(tx, cfg)
	}); err != nil {
		return err
	}
	b.Config = cfg
	b.Tasks = NewTasks(b.Store, cfg.TaskPrefix, cfg.Extensions)
	return nil
}

// WriteIdentity records the board's id inside its database, next to the
// <id>.id marker that repeats it in the directory.
func WriteIdentity(tx *sql.Tx, id string) error {
	_, err := tx.Exec(
		`INSERT INTO board (id, board_id) VALUES (1, ?)
		 ON CONFLICT (id) DO UPDATE SET board_id = excluded.board_id`, id)
	return err
}

// ReadIdentity answers the id the database holds, and the empty string when
// the board has none recorded yet.
func ReadIdentity(q rowQueryer) (string, error) {
	rows, err := q.Query(`SELECT board_id FROM board WHERE id = 1`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if !rows.Next() {
		return "", rows.Err()
	}
	var id string
	if err := rows.Scan(&id); err != nil {
		return "", err
	}
	return id, rows.Err()
}

type rowQueryer = queryer

// Counts are the two numbers of the `tasks` row of `biso where`: how many
// tasks are not archived and how many are. The row says "not archived" and
// not "active" because a task that is not archived can be in any status,
// the terminal one included (docs/spec/cmd/where.md).
type Counts struct {
	NotArchived int
	Archived    int
	// HighestEverAssigned is the highest number the board has ever handed
	// out, and zero when it has never handed out any.
	HighestEverAssigned int
}

// Counts reads them.
func (b *Board) Counts() (Counts, error) {
	var c Counts
	rows, err := b.Store.Query(
		`SELECT archived, count(*) FROM task GROUP BY archived`)
	if err != nil {
		return Counts{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var archived, n int
		if err := rows.Scan(&archived, &n); err != nil {
			return Counts{}, err
		}
		if archived == 0 {
			c.NotArchived = n
		} else {
			c.Archived = n
		}
	}
	if err := rows.Err(); err != nil {
		return Counts{}, err
	}
	last, err := b.Tasks.LastAllocated()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Counts{}, err
	}
	c.HighestEverAssigned = last
	return c, nil
}
