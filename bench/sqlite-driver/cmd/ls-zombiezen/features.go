package main

import (
	"fmt"
	"os"
	"path/filepath"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"

	"bisobench/internal/features"
)

// runFeatures repite las comprobaciones de internal/features con la API de
// zombiezen. Hay que escribirlas otra vez porque este controlador no pasa por
// database/sql, y eso ya es un dato sobre lo que cuesta usarlo.
func runFeatures() int {
	const name = "zombiezen"

	dir, err := os.MkdirTemp("", "biso-features-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "features: %v\n", err)
		return 1
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "probe.db")

	open := func() (*sqlite.Conn, error) {
		return sqlite.OpenConn(path, sqlite.OpenReadWrite|sqlite.OpenCreate|sqlite.OpenWAL|sqlite.OpenURI|sqlite.OpenNoMutex)
	}

	conn, err := open()
	if err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "open", "FALLA: "+err.Error())
		return 1
	}
	defer conn.Close()

	run := func(q string, args ...any) error {
		return sqlitex.Execute(conn, q, &sqlitex.ExecOptions{Args: args})
	}
	scalar := func(q string) (string, error) {
		var v string
		err := sqlitex.Execute(conn, q, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error {
				v = stmt.ColumnText(0)
				return nil
			},
		})
		return v, err
	}

	if err := run(`PRAGMA journal_mode = WAL`); err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "set journal_mode=WAL", "FALLA: "+err.Error())
	}
	if err := run(`CREATE TABLE t(id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "create table", "FALLA: "+err.Error())
	}

	for _, c := range features.SQLChecks {
		for _, s := range c.Setup {
			run(s)
		}
		v, err := scalar(c.Query)
		if err != nil {
			fmt.Printf("%-12s %-22s %s\n", name, c.Name, "FALLA: "+err.Error())
			continue
		}
		fmt.Printf("%-12s %-22s %s\n", name, c.Name, v)
	}

	features.Report(name, "tx_rollback", func() error {
		if err := run(`BEGIN`); err != nil {
			return err
		}
		if err := run(`INSERT INTO t(v) VALUES('x')`); err != nil {
			run(`ROLLBACK`)
			return err
		}
		if err := run(`ROLLBACK`); err != nil {
			return err
		}
		n, err := scalar(`SELECT COUNT(*) FROM t`)
		if err != nil {
			return err
		}
		if n != "0" {
			return fmt.Errorf("la vuelta atras dejo %s filas", n)
		}
		return nil
	})

	features.Report(name, "begin_immediate", func() error {
		if err := run(`BEGIN IMMEDIATE`); err != nil {
			return err
		}
		if err := run(`INSERT INTO t(v) VALUES('y')`); err != nil {
			run(`ROLLBACK`)
			return err
		}
		return run(`COMMIT`)
	})

	features.Report(name, "savepoint", func() error {
		if err := run(`SAVEPOINT s1`); err != nil {
			return err
		}
		if err := run(`INSERT INTO t(v) VALUES('z')`); err != nil {
			return err
		}
		if err := run(`ROLLBACK TO s1`); err != nil {
			return err
		}
		return run(`RELEASE s1`)
	})

	features.Report(name, "wal_checkpoint", func() error {
		_, err := scalar(`PRAGMA wal_checkpoint(TRUNCATE)`)
		return err
	})

	features.Report(name, "fts5", func() error {
		if err := run(`CREATE VIRTUAL TABLE ft USING fts5(title)`); err != nil {
			return err
		}
		if err := run(`INSERT INTO ft(title) VALUES('normalize the pointer resolution')`); err != nil {
			return err
		}
		n, err := scalar(`SELECT COUNT(*) FROM ft WHERE ft MATCH 'pointer'`)
		if err != nil {
			return err
		}
		if n != "1" {
			return fmt.Errorf("encontro %s filas en vez de 1", n)
		}
		return nil
	})

	features.Report(name, "analyze", func() error { return run(`ANALYZE`) })

	features.Report(name, "read_during_write", func() error {
		writer, err := open()
		if err != nil {
			return err
		}
		defer writer.Close()
		if err := sqlitex.Execute(writer, `BEGIN IMMEDIATE`, nil); err != nil {
			return err
		}
		if err := sqlitex.Execute(writer, `INSERT INTO t(v) VALUES('during')`, nil); err != nil {
			sqlitex.Execute(writer, `ROLLBACK`, nil)
			return err
		}
		reader, err := open()
		if err != nil {
			sqlitex.Execute(writer, `ROLLBACK`, nil)
			return err
		}
		defer reader.Close()
		readErr := sqlitex.Execute(reader, `SELECT COUNT(*) FROM t`, &sqlitex.ExecOptions{
			ResultFunc: func(stmt *sqlite.Stmt) error { return nil },
		})
		sqlitex.Execute(writer, `ROLLBACK`, nil)
		return readErr
	})

	return 0
}
