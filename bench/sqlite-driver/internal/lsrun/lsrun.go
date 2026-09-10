// Package lsrun es el cuerpo de `biso ls` compartido por los tres
// controladores que hablan por database/sql. Cada binario aporta solo su
// funcion de apertura.
package lsrun

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"bisobench/internal/board"
	"bisobench/internal/features"
	"bisobench/internal/phase"
)

// DefaultLimit es el limite por defecto de `biso ls` (seccion 10.4 de SPEC.md).
const DefaultLimit = 30

// Main hace lo que hara `biso ls`: abrir la base de datos, leer las tareas
// vivas y no terminales, calcular la urgencia, ordenar, recortar al limite e
// imprimir las ocho columnas.
//
// Uso: <binario> <ruta a board.db> [limite]
func Main(name string, open func(path string) (*sql.DB, error)) {
	// Los mismos binarios sirven para comprobar que soporta cada controlador,
	// porque son los unicos que tienen su import. Dos controladores distintos no
	// caben en un solo binario: mattn y ncruces registran los dos el nombre
	// "sqlite3" en database/sql y el segundo en registrarse hace panic.
	if os.Getenv("BENCH_FEATURES") == "1" {
		os.Exit(features.RunSQL(name, open))
	}

	t := phase.New(name)

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <board.db> [limit]\n", name)
		os.Exit(2)
	}
	path := os.Args[1]
	limit := DefaultLimit
	if len(os.Args) > 2 {
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: bad limit: %v\n", name, err)
			os.Exit(2)
		}
		limit = n
	}
	t.Mark("args")

	db, err := open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: open: %v\n", name, err)
		os.Exit(1)
	}
	// La apertura de database/sql es perezosa, asi que sin una consulta de
	// verdad no se habria abierto nada y la fase mediria cero. Esta es la
	// primera lectura que hace tambien el programa real.
	var journal string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil {
		fmt.Fprintf(os.Stderr, "%s: journal_mode: %v\n", name, err)
		os.Exit(1)
	}
	t.Mark("open")

	c, err := board.LoadSQL(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: query: %v\n", name, err)
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
	db.Close()
}
