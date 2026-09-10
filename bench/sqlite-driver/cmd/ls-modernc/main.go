// Binario de prueba de `biso ls` con modernc.org/sqlite, la traduccion de
// SQLite a Go puro. No necesita cgo.
package main

import (
	"database/sql"

	_ "modernc.org/sqlite"

	"bisobench/internal/lsrun"
)

func main() {
	lsrun.Main("ls-modernc", func(path string) (*sql.DB, error) {
		db, err := sql.Open("sqlite", "file:"+path)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(1)
		return db, nil
	})
}
