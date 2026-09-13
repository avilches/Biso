// gen crea un tablero de prueba con la forma que describen
// docs/spec/modelo-de-datos.md (el modelo de datos) y "La decision de
// persistencia" de docs/DECISIONES.md (una base de datos SQLite por tablero).
//
// Se compila con modernc.org/sqlite a proposito, para que generar el tablero no
// necesite ninguna herramienta de C. El fichero que sale es un SQLite corriente
// y los cuatro binarios de medida leen exactamente el mismo.
//
// El argumento -tasks cuenta las tareas que `biso ls` va a listar, es decir las
// vivas y no terminales. Encima de esas, el generador anade un 20 por ciento de
// tareas en el estado terminal y un 10 por ciento de archivadas, porque un
// tablero real las tiene y porque asi la tabla es mayor que lo que la consulta
// devuelve, que es el caso honesto.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"bisobench/internal/board"
)

var (
	statuses    = []string{"To Do", "In Progress", "In Review", "Blocked"}
	types       = []string{"bug", "task", "docs", "chore", "feature"}
	priorities  = []string{"high", "medium", "low", ""}
	projects    = []string{"core", "cli", "docs", "storage", ""}
	milestones  = []string{"v1.0", "v1.1", "backlog", ""}
	people      = []string{"@avilches", "@claude", "@sara", "@marc", "@lucia", "@opencode"}
	labelNames  = []string{"parser", "storage", "cli", "urgent", "regression", "ux", "perf", "flaky", "spec", "release"}
	verbs       = []string{"Fix", "Rewrite", "Normalize", "Retry", "Cache", "Split", "Document", "Measure", "Reject", "Recover", "Validate", "Deduplicate"}
	objects     = []string{"the pointer resolution", "the lease expiry", "the column widths", "the snapshot writer", "the vocabulary matcher", "the exit codes", "the urgency formula", "the archived filter", "the WAL fallback", "the identifier counter", "the criteria keys", "the prime message"}
	qualifiers  = []string{"on an empty board", "when the board is locked", "for non ASCII titles", "under a read only filesystem", "across two roots", "with a stale pointer", "in a worktree", "on the first run", "", "", ""}
	fancyTitles = []string{
		"Recortar títulos con acentos combinantes sin partir el grafema",
		"東アジアの全角文字で列がずれないようにする",
		"Emoji in a title must count as two cells 🚀",
		"Rechazar una clave de frontmatter desconocida al importar",
	}
)

func main() {
	out := flag.String("out", "board.db", "ruta del fichero de base de datos que se crea")
	tasks := flag.Int("tasks", 300, "tareas vivas y no terminales, las que `biso ls` lista")
	seed := flag.Int64("seed", 20260909, "semilla, para que el tablero sea reproducible")
	flag.Parse()

	if err := run(*out, *tasks, *seed); err != nil {
		fmt.Fprintf(os.Stderr, "gen: %v\n", err)
		os.Exit(1)
	}
}

func run(out string, live int, seed int64) error {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if err := os.Remove(out + suffix); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	db, err := sql.Open("sqlite", "file:"+out)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	for _, stmt := range strings.Split(board.Schema, ";") {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema %.40q: %w", strings.TrimSpace(stmt), err)
		}
	}

	rnd := rand.New(rand.NewSource(seed))
	done := live * 20 / 100
	archived := live * 10 / 100
	total := live + done + archived

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	cfg := map[string]string{
		"project_name":     "Biso Bench",
		"task_prefix":      "TASK",
		"lease_minutes":    "90",
		"me":               "@avilches",
		"urgency.priority": "6.0",
		"urgency.active":   "4.0",
		"urgency.blocking": "8.0",
		"urgency.blocked":  "-5.0",
		"urgency.due":      "12.0",
		"urgency.criteria": "1.0",
		"urgency.age":      "0.5",
		"statuses":         "To Do,In Progress,In Review,Blocked,Done",
		"status_active":    board.ActiveStatus,
		"status_terminal":  board.TerminalStatus,
		"types":            strings.Join(types, ","),
		"priorities":       "high,medium,low",
		"projects":         "core,cli,docs,storage",
		"vcs":              "git",
		"schema_version":   "1",
	}
	for k, v := range cfg {
		if _, err := tx.Exec(`INSERT INTO board(key, value) VALUES(?, ?)`, k, v); err != nil {
			return err
		}
	}

	insTask, err := tx.Prepare(`
INSERT INTO task(id, title, status, type, priority, project, milestone, parent, reporter,
                 due, ordinal, created_at, updated_at, archived,
                 lease_expires_at, lease_holder,
                 description, plan, notes, summary,
                 question_text, question_asked_at, question_answer)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	insAssignee, err := tx.Prepare(`INSERT INTO assignee(task_id, pos, who) VALUES(?,?,?)`)
	if err != nil {
		return err
	}
	insLabel, err := tx.Prepare(`INSERT INTO label(task_id, pos, name) VALUES(?,?,?)`)
	if err != nil {
		return err
	}
	insCriterion, err := tx.Prepare(`INSERT INTO criterion(task_id, kind, key, text, checked) VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}
	insComment, err := tx.Prepare(`INSERT INTO comment(task_id, seq, author, created_at, body) VALUES(?,?,?,?,?)`)
	if err != nil {
		return err
	}
	insText, err := tx.Prepare(`INSERT INTO textlist(task_id, kind, pos, value) VALUES(?,?,?,?)`)
	if err != nil {
		return err
	}
	insExt, err := tx.Prepare(`INSERT INTO ext(task_id, key, value) VALUES(?,?,?)`)
	if err != nil {
		return err
	}
	insDep, err := tx.Prepare(`INSERT INTO dependency(task_id, pos, depends_on) VALUES(?,?,?)`)
	if err != nil {
		return err
	}

	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	nulls := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}

	for i := 1; i <= total; i++ {
		var status string
		isArchived := 0
		switch {
		case i <= live:
			status = statuses[rnd.Intn(len(statuses))]
		case i <= live+done:
			status = board.TerminalStatus
		default:
			status = statuses[rnd.Intn(len(statuses))]
			isArchived = 1
		}
		title := makeTitle(rnd)
		typ := types[rnd.Intn(len(types))]
		prio := priorities[rnd.Intn(len(priorities))]
		proj := projects[rnd.Intn(len(projects))]
		mile := milestones[rnd.Intn(len(milestones))]

		created := now.Add(-time.Duration(rnd.Intn(400*24)) * time.Hour)
		updated := created.Add(time.Duration(rnd.Intn(72)) * time.Hour)

		var due any
		if rnd.Intn(100) < 18 {
			due = created.Add(time.Duration(rnd.Intn(120)-30) * 24 * time.Hour).Format("2006-01-02")
		}
		var ordinal any
		if rnd.Intn(100) < 30 {
			ordinal = rnd.Intn(500)
		}
		var parent any
		if i > 20 && rnd.Intn(100) < 22 {
			parent = 1 + rnd.Intn(i-1)
		}

		var leaseAt, leaseWho any
		nAssignees := 0
		switch n := rnd.Intn(100); {
		case n < 42:
			nAssignees = 0
		case n < 85:
			nAssignees = 1
		case n < 96:
			nAssignees = 2
		default:
			nAssignees = 3
		}
		if status == board.ActiveStatus && nAssignees > 0 && isArchived == 0 && rnd.Intn(100) < 70 {
			leaseWho = people[rnd.Intn(len(people))]
			leaseAt = now.Add(time.Duration(rnd.Intn(240)-120) * time.Minute).Format(time.RFC3339)
		}

		var qText, qAsked, qAnswer any
		if rnd.Intn(100) < 12 {
			qText = "Should the pointer be relative or absolute for this project?"
			qAsked = created.Add(48 * time.Hour).Format(time.RFC3339)
			if rnd.Intn(100) < 40 {
				qAnswer = "Relative, because the worktrees live inside the project."
			}
		}

		_, err := insTask.Exec(
			i, title, status, nulls(typ), nulls(prio), nulls(proj), nulls(mile), parent,
			people[rnd.Intn(len(people))],
			due, ordinal, created.Format(time.RFC3339), updated.Format(time.RFC3339), isArchived,
			leaseAt, leaseWho,
			prose(rnd, 3), prose(rnd, 2), prose(rnd, 2), prose(rnd, 1),
			qText, qAsked, qAnswer,
		)
		if err != nil {
			return fmt.Errorf("task %d: %w", i, err)
		}

		used := map[string]bool{}
		for pos := 0; pos < nAssignees; pos++ {
			who := people[rnd.Intn(len(people))]
			if used[who] {
				continue
			}
			used[who] = true
			if _, err := insAssignee.Exec(i, pos, who); err != nil {
				return err
			}
		}

		for pos, n := 0, rnd.Intn(4); pos < n; pos++ {
			if _, err := insLabel.Exec(i, pos, labelNames[rnd.Intn(len(labelNames))]); err != nil {
				return err
			}
		}

		if rnd.Intn(100) < 62 {
			for key, n := 1, 1+rnd.Intn(5); key <= n; key++ {
				checked := 0
				if rnd.Intn(100) < 45 {
					checked = 1
				}
				if _, err := insCriterion.Exec(i, "ac", key, criterionText(rnd), checked); err != nil {
					return err
				}
			}
		}
		if rnd.Intn(100) < 38 {
			for key, n := 1, 1+rnd.Intn(3); key <= n; key++ {
				checked := 0
				if rnd.Intn(100) < 55 {
					checked = 1
				}
				if _, err := insCriterion.Exec(i, "dod", key, criterionText(rnd), checked); err != nil {
					return err
				}
			}
		}

		if rnd.Intn(100) < 55 {
			for seq, n := 1, 1+rnd.Intn(4); seq <= n; seq++ {
				body := proseAlways(rnd, 2)
				at := created.Add(time.Duration(seq*17) * time.Hour).Format(time.RFC3339)
				if _, err := insComment.Exec(i, seq, people[rnd.Intn(len(people))], at, body); err != nil {
					return err
				}
			}
		}

		for _, kind := range []string{"references", "documentation", "modifiedFiles"} {
			for pos, n := 0, rnd.Intn(3); pos < n; pos++ {
				if _, err := insText.Exec(i, kind, pos, fmt.Sprintf("docs/%s/%s-%02d.md", kind, types[rnd.Intn(len(types))], pos)); err != nil {
					return err
				}
			}
		}

		if rnd.Intn(100) < 25 {
			if _, err := insExt.Exec(i, "jira", fmt.Sprintf("BISO-%d", 1000+rnd.Intn(9000))); err != nil {
				return err
			}
		}
	}

	// Las dependencias van en una segunda pasada para poder apuntar a
	// cualquier tarea, no solo a las anteriores.
	depCount := 0
	for i := 1; i <= total; i++ {
		if rnd.Intn(100) >= 28 {
			continue
		}
		n := 1 + rnd.Intn(2)
		seen := map[int]bool{i: true}
		for pos := 0; pos < n; pos++ {
			target := 1 + rnd.Intn(total)
			if seen[target] {
				continue
			}
			seen[target] = true
			if _, err := insDep.Exec(i, pos, target); err != nil {
				return err
			}
			depCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	if _, err := db.Exec("ANALYZE"); err != nil {
		return err
	}
	// Sin este punto de control los datos se quedarian en el fichero -wal y el
	// board.db mediria 4 KB, que no es el estado en el que un tablero vive
	// entre invocaciones. Con TRUNCATE queda todo en el fichero principal y la
	// medida es reproducible.
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return err
	}

	var rows int
	if err := db.QueryRow(`SELECT COUNT(*) FROM task WHERE archived = 0 AND status <> ?`, board.TerminalStatus).Scan(&rows); err != nil {
		return err
	}
	var journal string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil {
		return err
	}
	fi, err := os.Stat(out)
	if err != nil {
		return err
	}
	fmt.Printf("board=%s tasks_total=%d listed_by_ls=%d dependencies=%d journal_mode=%s size=%d bytes\n",
		out, total, rows, depCount, journal, fi.Size())
	return nil
}

func makeTitle(rnd *rand.Rand) string {
	if rnd.Intn(100) < 4 {
		return fancyTitles[rnd.Intn(len(fancyTitles))]
	}
	t := verbs[rnd.Intn(len(verbs))] + " " + objects[rnd.Intn(len(objects))]
	if q := qualifiers[rnd.Intn(len(qualifiers))]; q != "" {
		t += " " + q
	}
	if rnd.Intn(100) < 6 {
		t += " and also keep the previously documented behaviour for every other command in the suite, including the ones that only read"
	}
	return t
}

// prose devuelve prosa de relleno, o nil una de cada cuatro veces, porque los
// cuatro campos largos del modelo son opcionales.
func prose(rnd *rand.Rand, paragraphs int) any {
	if rnd.Intn(100) < 25 {
		return nil
	}
	return proseAlways(rnd, paragraphs)
}

// proseAlways devuelve siempre texto, para los campos que no admiten nulo.
func proseAlways(rnd *rand.Rand, paragraphs int) string {
	var sb strings.Builder
	for p := 0; p < paragraphs; p++ {
		for s, n := 0, 2+rnd.Intn(3); s < n; s++ {
			sb.WriteString(verbs[rnd.Intn(len(verbs))])
			sb.WriteString(" ")
			sb.WriteString(objects[rnd.Intn(len(objects))])
			sb.WriteString(", because the board must stay readable after every write. ")
		}
		sb.WriteString("\n\n")
	}
	return sb.String()
}

func criterionText(rnd *rand.Rand) string {
	return "The command exits with code 0 and prints " + objects[rnd.Intn(len(objects))]
}
