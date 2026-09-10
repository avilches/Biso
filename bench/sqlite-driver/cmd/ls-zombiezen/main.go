// Binario de prueba de `biso ls` con zombiezen.com/go/sqlite, que usa el mismo
// motor traducido que modernc.org/sqlite pero no pasa por database/sql: expone
// sentencias preparadas y un paso de fila explicito.
//
// Por eso este binario no puede reutilizar internal/lsrun ni
// board.LoadSQL: tiene que recorrer las cinco consultas con su propia API. Ese
// coste de escritura es parte de lo que se esta midiendo.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"bisobench/internal/board"
	"bisobench/internal/phase"
)

const defaultLimit = 30

func main() {
	if os.Getenv("BENCH_FEATURES") == "1" {
		os.Exit(runFeatures())
	}

	t := phase.New("ls-zombiezen")

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: ls-zombiezen <board.db> [limit]")
		os.Exit(2)
	}
	path := os.Args[1]
	limit := defaultLimit
	if len(os.Args) > 2 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "ls-zombiezen: bad limit: %v\n", err)
			os.Exit(2)
		}
		limit = n
	}
	t.Mark("args")

	conn, err := sqlite.OpenConn(path, sqlite.OpenReadWrite|sqlite.OpenWAL|sqlite.OpenURI|sqlite.OpenNoMutex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls-zombiezen: open: %v\n", err)
		os.Exit(1)
	}
	var journal string
	err = sqlitex.Execute(conn, "PRAGMA journal_mode", &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			journal = stmt.ColumnText(0)
			return nil
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls-zombiezen: journal_mode: %v\n", err)
		os.Exit(1)
	}
	t.Mark("open")

	c, err := load(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls-zombiezen: query: %v\n", err)
		os.Exit(1)
	}
	t.Mark("query")

	out, shown, total := c.Finish(time.Now().UTC(), limit)
	t.Mark("compute")

	w := bufio.NewWriterSize(os.Stdout, 1<<16)
	w.WriteString(out)
	w.Flush()
	if warn := board.TruncationWarning(shown, total); warn != "" {
		os.Stderr.WriteString(warn)
	}
	t.Mark("write")

	if os.Getenv("BENCH_JOURNAL") == "1" {
		fmt.Fprintf(os.Stderr, "journal_mode=%s\n", journal)
	}
	t.Report()
	conn.Close()
}

func load(conn *sqlite.Conn) (*board.Collected, error) {
	c := board.NewCollected()

	err := sqlitex.Execute(conn, board.QueryConfig, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			c.AddConfig(stmt.ColumnText(0), stmt.ColumnText(1))
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn, board.QueryTasks, &sqlitex.ExecOptions{
		Args: []any{board.TerminalStatus},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			hasOrdinal := stmt.ColumnType(6) != sqlite.TypeNull
			c.AddTask(
				stmt.ColumnInt64(0),
				stmt.ColumnText(1),
				stmt.ColumnText(2),
				stmt.ColumnText(3),
				stmt.ColumnText(4),
				stmt.ColumnText(5),
				stmt.ColumnInt64(6),
				hasOrdinal,
				stmt.ColumnText(7),
				stmt.ColumnBool(8),
			)
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn, board.QueryAssignees, &sqlitex.ExecOptions{
		Args: []any{board.TerminalStatus},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			c.AddAssignee(stmt.ColumnInt64(0), stmt.ColumnText(1))
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn, board.QueryCriteria, &sqlitex.ExecOptions{
		Args: []any{board.TerminalStatus},
		ResultFunc: func(stmt *sqlite.Stmt) error {
			c.AddCriteria(stmt.ColumnInt64(0), stmt.ColumnInt(1), stmt.ColumnInt(2))
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	err = sqlitex.Execute(conn, board.QueryEdges, &sqlitex.ExecOptions{
		ResultFunc: func(stmt *sqlite.Stmt) error {
			c.AddEdge(stmt.ColumnInt64(0), stmt.ColumnText(1), stmt.ColumnInt64(2), stmt.ColumnText(3))
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	return c, nil
}
