// Package features comprueba que un controlador soporta lo que el esquema y la
// especificacion de `biso` dan por hecho.
//
// No todos los SQLite son el mismo SQLite: cada controlador trae su propia copia
// compilada con sus propias opciones, asi que una extension como FTS5 o el modulo
// JSON puede estar en uno y faltar en otro. Estas comprobaciones existen para no
// descubrirlo a mitad de la implementacion.
//
// Lo que se comprueba, y de donde sale la exigencia:
//
//   - Modo WAL: "Concurrencia, atomicidad y garantias observables"
//     (docs/spec/garantias.md) promete que una lectura nunca falla por una
//     escritura en curso, y "El controlador de SQLite es modernc.org/sqlite,
//     sin cgo" (docs/DECISIONES.md) nombra el modo WAL como lo que da esa
//     garantia.
//   - BEGIN IMMEDIATE: es el acceso exclusivo de escritura del que hablan
//     "Concurrencia, atomicidad y garantias observables" y el codigo de
//     salida 8.
//   - busy_timeout: sin el, dos escrituras a la vez fallan en vez de esperar.
//   - Puntos de retorno: "Orden de aplicacion dentro de una escritura"
//     (docs/spec/garantias.md) aplica varios flags dentro de una escritura,
//     y el lote de `biso new --from` valida entero antes de escribir.
//   - integrity_check: es una de las dos comprobaciones que "La decision de
//     persistencia" anade a `biso doctor` (docs/spec/cmd/doctor.md).
//   - Claves ajenas: el esquema usa REFERENCES para las dependencias y el padre.
//   - user_version: es donde vive la version del esquema para migrar.
//   - Consultas recursivas: el grafo de dependencias y el `parent` son arboles.
//   - Modulo JSON: hace falta si `ext` o los criterios se guardan como JSON.
//   - FTS5: es la via barata para "La busqueda por texto"
//     (docs/spec/referencias.md), que usa `--search`.
package features

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// Check es una comprobacion con nombre.
type Check struct {
	Name  string
	Query string
	// Setup se ejecuta antes, y su fallo no invalida la comprobacion.
	Setup []string
}

// SQLChecks son las comprobaciones que se pueden hacer con una sola consulta
// que devuelve un valor.
var SQLChecks = []Check{
	{Name: "sqlite_version", Query: `SELECT sqlite_version()`},
	{Name: "journal_mode_wal", Query: `PRAGMA journal_mode`},
	{Name: "busy_timeout", Query: `PRAGMA busy_timeout = 5000`},
	{Name: "foreign_keys", Setup: []string{`PRAGMA foreign_keys = ON`}, Query: `PRAGMA foreign_keys`},
	{Name: "user_version", Setup: []string{`PRAGMA user_version = 7`}, Query: `PRAGMA user_version`},
	{Name: "integrity_check", Query: `PRAGMA integrity_check`},
	{Name: "recursive_cte", Query: `WITH RECURSIVE c(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM c WHERE n < 5) SELECT SUM(n) FROM c`},
	{Name: "json1", Query: `SELECT json_extract('{"a":{"b":42}}', '$.a.b')`},
	{Name: "window_functions", Query: `SELECT COUNT(*) FROM (SELECT ROW_NUMBER() OVER (ORDER BY 1) AS r)`},
	{Name: "collate_nocase_like", Query: `SELECT 'Ábaco' LIKE 'ábaco'`},
	{Name: "compile_options", Query: `SELECT COUNT(*) FROM pragma_compile_options`},
}

// RunSQL ejecuta las comprobaciones contra un controlador de database/sql y
// escribe una linea por comprobacion, con el nombre del controlador delante.
func RunSQL(name string, open func(path string) (*sql.DB, error)) int {
	dir, err := os.MkdirTemp("", "biso-features-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "features: %v\n", err)
		return 1
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "probe.db")

	db, err := open(path)
	if err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "open", "FALLA: "+err.Error())
		return 1
	}
	defer db.Close()

	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "set journal_mode=WAL", "FALLA: "+err.Error())
	}
	if _, err := db.Exec(`CREATE TABLE t(id INTEGER PRIMARY KEY, v TEXT)`); err != nil {
		fmt.Printf("%-12s %-22s %s\n", name, "create table", "FALLA: "+err.Error())
	}

	for _, c := range SQLChecks {
		for _, s := range c.Setup {
			db.Exec(s)
		}
		var v any
		if err := db.QueryRow(c.Query).Scan(&v); err != nil {
			fmt.Printf("%-12s %-22s %s\n", name, c.Name, "FALLA: "+err.Error())
			continue
		}
		fmt.Printf("%-12s %-22s %s\n", name, c.Name, render(v))
	}

	// Transaccion con vuelta atras.
	report(name, "tx_rollback", func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO t(v) VALUES('x')`); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Rollback(); err != nil {
			return err
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return fmt.Errorf("la vuelta atras dejo %d filas", n)
		}
		return nil
	})

	// El acceso exclusivo de escritura de "Concurrencia, atomicidad y
	// garantias observables".
	report(name, "begin_immediate", func() error {
		if _, err := db.Exec(`BEGIN IMMEDIATE`); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO t(v) VALUES('y')`); err != nil {
			db.Exec(`ROLLBACK`)
			return err
		}
		return exec(db, `COMMIT`)
	})

	report(name, "savepoint", func() error {
		if err := exec(db, `SAVEPOINT s1`); err != nil {
			return err
		}
		if err := exec(db, `INSERT INTO t(v) VALUES('z')`); err != nil {
			return err
		}
		return exec(db, `ROLLBACK TO s1`, `RELEASE s1`)
	})

	report(name, "wal_checkpoint", func() error {
		var a, b, c any
		return db.QueryRow(`PRAGMA wal_checkpoint(TRUNCATE)`).Scan(&a, &b, &c)
	})

	report(name, "fts5", func() error {
		if err := exec(db,
			`CREATE VIRTUAL TABLE ft USING fts5(title)`,
			`INSERT INTO ft(title) VALUES('normalize the pointer resolution')`,
		); err != nil {
			return err
		}
		var n int
		if err := db.QueryRow(`SELECT COUNT(*) FROM ft WHERE ft MATCH 'pointer'`).Scan(&n); err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("encontro %d filas en vez de 1", n)
		}
		return nil
	})

	report(name, "analyze", func() error { return exec(db, `ANALYZE`) })

	// Dos conexiones a la vez sobre el mismo fichero: una lee mientras la otra
	// tiene una escritura abierta, que es lo que promete "Concurrencia,
	// atomicidad y garantias observables".
	report(name, "read_during_write", func() error {
		writer, err := open(path)
		if err != nil {
			return err
		}
		defer writer.Close()
		if _, err := writer.Exec(`BEGIN IMMEDIATE`); err != nil {
			return err
		}
		if _, err := writer.Exec(`INSERT INTO t(v) VALUES('during')`); err != nil {
			writer.Exec(`ROLLBACK`)
			return err
		}
		reader, err := open(path)
		if err != nil {
			writer.Exec(`ROLLBACK`)
			return err
		}
		defer reader.Close()
		var n int
		readErr := reader.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&n)
		writer.Exec(`ROLLBACK`)
		return readErr
	})

	return 0
}

func exec(db *sql.DB, stmts ...string) error {
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("%s: %w", s, err)
		}
	}
	return nil
}

func report(driver, name string, fn func() error) {
	if err := fn(); err != nil {
		fmt.Printf("%-12s %-22s %s\n", driver, name, "FALLA: "+err.Error())
		return
	}
	fmt.Printf("%-12s %-22s %s\n", driver, name, "ok")
}

// Report es la version publica de report, para el binario de zombiezen, que no
// pasa por database/sql.
func Report(driver, name string, fn func() error) { report(driver, name, fn) }

// Render formatea el valor que devuelve una comprobacion.
func Render(v any) string { return render(v) }

func render(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return string(x)
	case string:
		return x
	case int64:
		return fmt.Sprintf("%d", x)
	default:
		return fmt.Sprintf("%v", x)
	}
}
