// Binario de prueba de `biso ls` con github.com/mattn/go-sqlite3, el enlace con
// la biblioteca de SQLite en C. Obliga a compilar con cgo.
package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"bisobench/internal/lsrun"
)

func main() {
	lsrun.Main("ls-mattn", func(path string) (*sql.DB, error) {
		db, err := sql.Open("sqlite3", "file:"+path)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(1)
		return db, nil
	})
}
