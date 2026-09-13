package board

// Schema es el esquema del tablero, con la forma que "La decision de
// persistencia" (docs/DECISIONES.md) da por hecha: una base de datos SQLite por
// tablero, con la configuracion dentro de ella y una tabla por cada campo de
// lista del modelo de docs/spec/modelo-de-datos.md.
//
// No pretende ser el esquema definitivo de `biso`. Pretende costar lo mismo de
// leer: los mismos indices, el mismo numero de tablas que `biso ls` tiene que
// cruzar y el mismo tamano de fila, para que la medicion valga.
const Schema = `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

CREATE TABLE board (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
) WITHOUT ROWID;

CREATE TABLE task (
  id                INTEGER PRIMARY KEY,
  title             TEXT    NOT NULL,
  status            TEXT    NOT NULL,
  type              TEXT,
  priority          TEXT,
  project           TEXT,
  milestone         TEXT,
  parent            INTEGER REFERENCES task(id),
  reporter          TEXT,
  due               TEXT,
  ordinal           INTEGER,
  created_at        TEXT    NOT NULL,
  updated_at        TEXT    NOT NULL,
  archived          INTEGER NOT NULL DEFAULT 0,
  lease_expires_at  TEXT,
  lease_holder      TEXT,
  description       TEXT,
  plan              TEXT,
  notes             TEXT,
  summary           TEXT,
  question_text     TEXT,
  question_asked_at TEXT,
  question_answer   TEXT
);

CREATE INDEX task_live ON task(archived, status);

CREATE TABLE assignee (
  task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  pos     INTEGER NOT NULL,
  who     TEXT    NOT NULL,
  PRIMARY KEY (task_id, pos)
) WITHOUT ROWID;

CREATE TABLE label (
  task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  pos     INTEGER NOT NULL,
  name    TEXT    NOT NULL,
  PRIMARY KEY (task_id, pos)
) WITHOUT ROWID;

CREATE TABLE dependency (
  task_id    INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  pos        INTEGER NOT NULL,
  depends_on INTEGER NOT NULL REFERENCES task(id),
  PRIMARY KEY (task_id, pos)
) WITHOUT ROWID;

CREATE INDEX dependency_target ON dependency(depends_on);

CREATE TABLE criterion (
  task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  kind    TEXT    NOT NULL,
  key     INTEGER NOT NULL,
  text    TEXT    NOT NULL,
  checked INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (task_id, kind, key)
) WITHOUT ROWID;

CREATE TABLE comment (
  task_id    INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  seq        INTEGER NOT NULL,
  author     TEXT    NOT NULL,
  created_at TEXT    NOT NULL,
  body       TEXT    NOT NULL,
  PRIMARY KEY (task_id, seq)
) WITHOUT ROWID;

CREATE TABLE textlist (
  task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  kind    TEXT    NOT NULL,
  pos     INTEGER NOT NULL,
  value   TEXT    NOT NULL,
  PRIMARY KEY (task_id, kind, pos)
) WITHOUT ROWID;

CREATE TABLE ext (
  task_id INTEGER NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  key     TEXT    NOT NULL,
  value   TEXT    NOT NULL,
  PRIMARY KEY (task_id, key)
) WITHOUT ROWID;
`

// Las cinco consultas que `biso ls` necesita, y ni una mas. La primera regla de
// "El presupuesto de arranque" (docs/spec/presupuestos.md) prohibe leer lo que
// la invocacion no va a imprimir, asi que aqui no se leen ni comentarios, ni
// etiquetas, ni prosa, ni campos externos: `biso ls` no imprime ninguno de ellos.
const (
	// QueryConfig trae la configuracion del tablero, de donde salen los siete
	// coeficientes de urgencia y el nombre del estado terminal.
	QueryConfig = `SELECT key, value FROM board`

	// QueryTasks trae las tareas vivas y no terminales, que es el filtro por
	// defecto de `biso ls` de docs/spec/cmd/ls.md.
	QueryTasks = `
SELECT id, title, status,
       COALESCE(type, ''), COALESCE(priority, ''), COALESCE(due, ''),
       ordinal, created_at,
       (question_text IS NOT NULL AND question_answer IS NULL)
  FROM task
 WHERE archived = 0 AND status <> ?`

	// QueryAssignees trae las personas asignadas de esas mismas tareas, en
	// orden, para la septima columna.
	QueryAssignees = `
SELECT a.task_id, a.who
  FROM assignee a JOIN task t ON t.id = a.task_id
 WHERE t.archived = 0 AND t.status <> ?
 ORDER BY a.task_id, a.pos`

	// QueryCriteria trae los recuentos de criterios de aceptacion, para la
	// sexta columna y para el termino `tiene_criterios` de la urgencia.
	QueryCriteria = `
SELECT c.task_id, COUNT(*), SUM(c.checked)
  FROM criterion c JOIN task t ON t.id = c.task_id
 WHERE c.kind = 'ac' AND t.archived = 0 AND t.status <> ?
 GROUP BY c.task_id`

	// QueryEdges trae las aristas de dependencia con el estado de los dos
	// extremos, de donde salen los derivados `blocked` y `blocks`.
	QueryEdges = `
SELECT d.task_id, a.status, d.depends_on, b.status
  FROM dependency d
  JOIN task a ON a.id = d.task_id
  JOIN task b ON b.id = d.depends_on`
)
