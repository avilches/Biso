package board

import (
	"database/sql"
)

// LoadSQL ejecuta las cinco consultas de `biso ls` contra un `*sql.DB`, que es
// la via que ofrecen mattn/go-sqlite3, modernc.org/sqlite y
// ncruces/go-sqlite3. Los tres comparten este cargador entero: entre ellos solo
// cambia el `import` del controlador y la cadena de conexion.
func LoadSQL(db *sql.DB) (*Collected, error) {
	c := NewCollected()

	rows, err := db.Query(QueryConfig)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			rows.Close()
			return nil, err
		}
		c.AddConfig(k, v)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = db.Query(QueryTasks, TerminalStatus)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var (
			id                            int64
			title, status, typ, prio, due string
			ordinal                       sql.NullInt64
			createdAt                     string
			openQuestion                  bool
		)
		if err := rows.Scan(&id, &title, &status, &typ, &prio, &due, &ordinal, &createdAt, &openQuestion); err != nil {
			rows.Close()
			return nil, err
		}
		c.AddTask(id, title, status, typ, prio, due, ordinal.Int64, ordinal.Valid, createdAt, openQuestion)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = db.Query(QueryAssignees, TerminalStatus)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var who string
		if err := rows.Scan(&id, &who); err != nil {
			rows.Close()
			return nil, err
		}
		c.AddAssignee(id, who)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = db.Query(QueryCriteria, TerminalStatus)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int64
		var total, done int
		if err := rows.Scan(&id, &total, &done); err != nil {
			rows.Close()
			return nil, err
		}
		c.AddCriteria(id, total, done)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	rows, err = db.Query(QueryEdges)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var from, to int64
		var fromStatus, toStatus string
		if err := rows.Scan(&from, &fromStatus, &to, &toStatus); err != nil {
			rows.Close()
			return nil, err
		}
		c.AddEdge(from, fromStatus, to, toStatus)
	}
	if err := closeRows(rows); err != nil {
		return nil, err
	}

	return c, nil
}

func closeRows(rows *sql.Rows) error {
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	return rows.Close()
}
