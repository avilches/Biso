// Binario de prueba de `biso ls` con github.com/ncruces/go-sqlite3, que es Go
// puro por otra via: lleva SQLite compilado a WebAssembly y lo ejecuta con
// wazero, en vez de traducir su codigo a Go.
//
// No importa el paquete `embed` a mano: desde la version 0.35 el controlador ya
// lleva el modulo dentro, y si se importa ademas avisa por la salida de error de
// que sobra, con lo que ensuciaria la medida.
package main

import (
	"database/sql"

	_ "github.com/ncruces/go-sqlite3/driver"

	"bisobench/internal/lsrun"
)

func main() {
	lsrun.Main("ls-ncruces", func(path string) (*sql.DB, error) {
		db, err := sql.Open("sqlite3", "file:"+path)
		if err != nil {
			return nil, err
		}
		db.SetMaxOpenConns(1)
		return db, nil
	})
}
