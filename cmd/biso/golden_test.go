package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"biso/internal/board"
)

// This file holds the golden tests: every block the specification prints
// character for character lives under testdata, transcribed from its page,
// and the output is compared with it byte by byte. No fixture is ever
// generated from this code, which would only prove that the code does what
// it does; the last test of the file reads the same blocks straight out of
// docs/spec/ and checks that the transcription is still faithful.
//
// They live here, next to cmd/biso, because a golden test of a command runs
// the command: the program is compiled and executed, and what is compared
// is what the process wrote on each of its two streams. Checking the return
// value of a rendering function instead would leave everything between that
// function and the process untested, which is precisely where the literal
// output of a command is decided.
//
// Two things a real process cannot be told are the clock and the source of
// identifiers, so the two places a fixture carries one say so: an envelope's
// generatedAt is normalized after checking its shape, and a board whose id
// the specification fixes is built here with that id instead of being
// created by the program.

func TestInitPrintsTheOutputOfTheSpecification(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "init", "My project", "--prefix", "MYP",
		"--at", "my-project-board", "--extensions", "trello.card").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "init-output.txt"), "the output of biso init")
	assertEqual(t, got.stderr, fixture(t, "init-notes.txt"), "the notes of biso init")
}

func TestInitWritesThePointerOfTheSpecification(t *testing.T) {
	m := newMachine(t)

	m.run(t, "init", "My project", "--prefix", "MYP",
		"--at", "my-project-board", "--extensions", "trello.card").assertCode(t, 0)

	// docs/spec/cmd/init.md prints this pointer for this very call. The id
	// is the one this run minted, because nothing can hand a real process
	// the one of the example.
	id := board.MarkerID(filepath.Join(m.dir, "my-project-board"))
	assertEqual(t,
		m.read(t, filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \""+id+"\", \"path\": \"my-project-board\" }\n",
		"the project pointer")
}

func TestInitEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "init", "My project", "--prefix", "MYP", "--at", "my-project-board",
		"--extensions", "trello.card", "--json").assertCode(t, 0)

	assertSameJSON(t, got.stdout, fixture(t, "init-json.txt"))
	// The notes keep travelling as text on stderr with --json, because no
	// envelope carries one (docs/spec/salida-y-terminal.md#notas-y-avisos).
	assertEqual(t, got.stderr, fixture(t, "init-notes.txt"), "the notes of biso init --json")
}

func TestWherePrintsTheOutputOfTheSpecification(t *testing.T) {
	m := newMachine(t)
	m.buildTheBoardOfTheSpecification(t)
	m.env["BISO_ME"] = "@claude"

	got := m.at(m.projectOfTheSpecification()).run(t, "where").assertCode(t, 0)

	assertEqual(t, got.stdout, m.substituted(fixture(t, "where-output.txt")),
		"the output of biso where")
}

func TestWherePrintsTheDiscardedCandidateOfTheSpecification(t *testing.T) {
	m := newMachine(t)
	dir := m.buildTheBoardOfTheSpecification(t)
	m.env["BISO_ME"] = "@claude"

	// The second board of the example, and a call made from inside the
	// first one while the pointer that is there names the second: the
	// working directory wins and says what it turned down
	// (docs/spec/cmd/where.md#cuando-hay-más-de-un-candidato).
	other := filepath.Join(m.home, "Volumes", "work", "boards", "other-project-7a1b2c3d")
	m.buildBoard(t, other, "7a1b2c3d", "Other project", "OTHER", 0, 0, 0)
	m.write(t, filepath.Join(dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"7a1b2c3d\", \"path\": \""+other+"\" }\n")

	got := m.at(dir).run(t, "where").assertCode(t, 0)

	assertEqual(t, got.stdout, m.substituted(fixture(t, "where-candidates.txt")),
		"the output of biso where with a discarded candidate")
}

func TestWhereEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := newMachine(t)
	m.buildTheBoardOfTheSpecification(t)
	m.env["BISO_ME"] = "@claude"

	got := m.at(m.projectOfTheSpecification()).run(t, "where", "--json").assertCode(t, 0)

	assertSameJSON(t, got.stdout, m.substituted(fixture(t, "where-json.txt")))
}

func TestWhereWithoutABoardPrintsWhatItSearched(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "where").assertCode(t, 20)

	assertEqual(t, got.stderr, m.substituted(fixture(t, "where-no-board.txt")),
		"the no-board message of biso where")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso where")
}

func TestWhereWithAPointerThatNamesAnAbsentBoard(t *testing.T) {
	m := newMachine(t)
	m.write(t, filepath.Join(m.dir, ".biso.json"), "{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")

	got := m.run(t, "where").assertCode(t, 20)

	assertEqual(t, got.stderr, fixture(t, "where-unresolved.txt"),
		"the pointer_unresolved message")
}

func TestHelpTextsAreTheOnesOfTheSpecification(t *testing.T) {
	m := newMachine(t)

	for _, c := range []struct {
		argv    []string
		want    string
		subject string
	}{
		{[]string{"--help"}, fixture(t, "top-help.txt"), "biso --help"},
		{[]string{"-h"}, fixture(t, "top-help.txt"), "biso -h"},
		{[]string{}, fixture(t, "top-help.txt"), "biso with no command"},
		{[]string{"init", "--help"}, fixture(t, "init-help.txt"), "biso init --help"},
		{[]string{"where", "--help"}, fixture(t, "where-help.txt"), "biso where --help"},
		{[]string{"where", "-h"}, fixture(t, "where-help.txt"), "biso where -h"},
		{[]string{"new", "--help"}, fixture(t, "new-help.txt"), "biso new --help"},
		{[]string{"ls", "--help"}, fixture(t, "ls-help.txt"), "biso ls --help"},
		{[]string{"get", "--help"}, fixture(t, "get-help.txt"), "biso get --help"},
		{[]string{"prime", "--help"}, fixture(t, "prime-help.txt"), "biso prime --help"},
		{[]string{"set", "--help"}, fixture(t, "set-help.txt"), "biso set --help"},
		{[]string{"start", "--help"}, fixture(t, "start-help.txt"), "biso start --help"},
		{[]string{"note", "--help"}, fixture(t, "note-help.txt"), "biso note --help"},
		{[]string{"comment", "--help"}, fixture(t, "comment-help.txt"), "biso comment --help"},
		{[]string{"finish", "--help"}, fixture(t, "finish-help.txt"), "biso finish --help"},
		{[]string{"ask", "--help"}, fixture(t, "ask-help.txt"), "biso ask --help"},
		{[]string{"answer", "--help"}, fixture(t, "answer-help.txt"), "biso answer --help"},
	} {
		got := m.run(t, c.argv...).assertCode(t, 0)
		assertEqual(t, got.stdout, c.want, c.subject)
		if got.stderr != "" {
			t.Errorf("%s wrote to stderr: %s", c.subject, got.stderr)
		}
	}
}

func TestVersion(t *testing.T) {
	m := newMachine(t)
	for _, argv := range [][]string{{"--version"}, {"-V"}, {"where", "--version"}} {
		got := m.run(t, argv...).assertCode(t, 0)
		assertEqual(t, got.stdout, "biso 1.0.0\n", "biso "+strings.Join(argv, " "))
	}
}

// TestTheFixturesStillMatchTheSpecification is what keeps the transcription
// honest: it reads the same fenced blocks out of docs/spec/ and compares them
// with the files under testdata. A fixture that drifts from its page, in
// either direction, fails here and nowhere else.
func TestTheFixturesStillMatchTheSpecification(t *testing.T) {
	for _, c := range []struct {
		fixture string
		page    string
		// section is the level-two heading the block hangs from, needed
		// on a page that documents six commands and therefore repeats
		// every heading six times. It is empty where the heading alone
		// names one place.
		section string
		heading string
		index   int
	}{
		{"init-help.txt", "cmd/init.md", "", "`biso init --help`", 0},
		{"where-help.txt", "cmd/where.md", "", "`biso where --help`", 0},
		{"top-help.txt", "cmd/help.md", "", "La ayuda de primer nivel", 0},
		{"init-output.txt", "cmd/init.md", "", "Salida", 0},
		{"init-notes.txt", "cmd/init.md", "", "Salida", 1},
		{"init-json.txt", "cmd/init.md", "", "El esquema JSON", 0},
		{"where-output.txt", "cmd/where.md", "", "Salida", 0},
		{"where-candidates.txt", "cmd/where.md", "", "Cuando hay más de un candidato", 0},
		{"where-no-board.txt", "cmd/where.md", "", "Cuando hay más de un candidato", 1},
		{"where-unresolved.txt", "cmd/where.md", "", "Cuando hay más de un candidato", 2},
		{"where-json.txt", "cmd/where.md", "", "El esquema JSON", 0},
		{"new-help.txt", "cmd/new.md", "", "`biso new --help`", 0},
		{"set-help.txt", "cmd/set.md", "", "`biso set --help`", 0},
		{"set-status-line.txt", "cmd/set.md", "", "Salida", 0},
		{"set-added-ac.txt", "cmd/set.md", "", "Salida", 1},
		{"set-dry-run.txt", "cmd/set.md", "", "Salida", 2},
		{"set-overwrite.txt", "cmd/set.md", "", "Salida", 3},
		{"ls-help.txt", "cmd/ls.md", "", "`biso ls --help`", 0},
		{"ls-output.txt", "cmd/ls.md", "", "Salida", 0},
		{"ls-truncated.txt", "cmd/ls.md", "", "Salida", 1},
		{"ls-ids.txt", "cmd/ls.md", "", "Salida", 2},
		{"ls-count.txt", "cmd/ls.md", "", "Salida", 3},
		{"ls-json.txt", "cmd/ls.md", "", "El esquema JSON", 0},
		{"get-help.txt", "cmd/get.md", "", "`biso get --help`", 0},
		{"get-output.txt", "cmd/get.md", "", "Salida", 0},
		{"get-section-ac.txt", "cmd/get.md", "", "Salida", 1},
		{"get-question.txt", "cmd/get.md", "", "Salida", 2},
		{"get-explain.txt", "cmd/get.md", "", "Salida", 3},
		{"get-terminal-urgency.txt", "cmd/get.md", "", "Salida", 4},
		{"get-json.txt", "cmd/get.md", "", "El esquema JSON", 0},
		{"prime-help.txt", "cmd/prime.md", "", "`biso prime --help`", 0},
		{"prime-full-json.txt", "cmd/prime.md", "", "Parámetros", 0},
		{"prime-limit-negative.txt", "cmd/prime.md", "", "Qué hace, caso a caso", 0},
		{"prime-output.txt", "cmd/prime.md", "", "La salida literal", 0},
		{"prime-identity.txt", "cmd/prime.md", "", "La salida literal", 1},
		{"prime-unreadable.txt", "cmd/prime.md", "", "La salida literal", 2},
		{"prime-empty.txt", "cmd/prime.md", "", "Tablero vacío", 0},
		{"prime-full.txt", "cmd/prime.md", "", "`--full`", 0},
		{"prime-json.txt", "cmd/prime.md", "", "El esquema JSON", 0},
		{"start-help.txt", "cmd/verbos-del-ciclo.md", "`biso start`", "`biso start --help`", 0},
		{"note-help.txt", "cmd/verbos-del-ciclo.md", "`biso note`", "`biso note --help`", 0},
		{"comment-help.txt", "cmd/verbos-del-ciclo.md", "`biso comment`", "`biso comment --help`", 0},
		{"finish-help.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "`biso finish --help`", 0},
		{"ask-help.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "`biso ask --help`", 0},
		{"answer-help.txt", "cmd/verbos-del-ciclo.md", "`biso answer`", "`biso answer --help`", 0},
		{"start-archived.txt", "cmd/verbos-del-ciclo.md", "`biso start`", "Qué hace", 0},
		{"start-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso start`", "Salida", 0},
		{"start-dry-run.txt", "cmd/verbos-del-ciclo.md", "`biso start`", "Salida", 1},
		{"note-id-like.txt", "cmd/verbos-del-ciclo.md", "`biso note`", "El posicional que parece un identificador", 0},
		{"note-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso note`", "Salida", 0},
		{"comment-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso comment`", "Salida", 0},
		{"finish-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "Salida", 0},
		{"finish-ac-warning.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "Salida", 1},
		{"finish-subtasks-warning.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "Salida", 2},
		{"finish-dry-run.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "Salida", 3},
		{"ask-id-like.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "Parámetros propios", 0},
		{"ask-open-question.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "Qué hace", 0},
		{"ask-finished.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "Qué hace", 1},
		{"ask-missing-text.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "Qué hace", 2},
		{"ask-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso ask`", "Salida", 0},
		{"answer-id-like.txt", "cmd/verbos-del-ciclo.md", "`biso answer`", "Parámetros propios", 0},
		{"answer-no-question.txt", "cmd/verbos-del-ciclo.md", "`biso answer`", "Qué hace", 0},
		{"answer-missing-text.txt", "cmd/verbos-del-ciclo.md", "`biso answer`", "Qué hace", 1},
		{"answer-status-line.txt", "cmd/verbos-del-ciclo.md", "`biso answer`", "Salida", 0},
		{"start-already-finished.txt", "cmd/verbos-del-ciclo.md", "`biso start`", "Qué hace", 1},
		{"note-missing-text.txt", "cmd/verbos-del-ciclo.md", "`biso note`", "Qué hace", 0},
		{"comment-id-like.txt", "cmd/verbos-del-ciclo.md", "`biso comment`", "Firma", 1},
		{"finish-strict.txt", "cmd/verbos-del-ciclo.md", "`biso finish`", "Qué hace", 0},
	} {
		block := specBlockIn(t, c.page, c.section, c.heading, c.index)
		assertEqual(t, fixture(t, c.fixture), block,
			c.fixture+", against "+c.page)
	}
}

// specBlock reads the index-th fenced block under a heading of a page of the
// specification.
func specBlock(t *testing.T, page, heading string, index int) string {
	t.Helper()
	return specBlockIn(t, page, "", heading, index)
}

// specBlockIn is specBlock narrowed to one section of the page, because a
// page that documents six commands repeats "Salida" six times and the
// heading alone no longer names one block. within is the level-two heading
// the section hangs from, and the empty string means any.
func specBlockIn(t *testing.T, page, within, heading string, index int) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "spec", page))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	section, current := "", ""
	found := false
	var blocks []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#") {
			if current == heading && (within == "" || section == within) {
				found = true
				break
			}
			title := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if strings.HasPrefix(line, "## ") {
				section = title
			}
			current = title
			blocks = nil
			continue
		}
		if !strings.HasPrefix(strings.TrimLeft(line, " "), "```") {
			continue
		}
		// A fence inside a numbered list is indented, and so is its
		// content: the block is what the page shows, without the
		// indentation the list put in front of it.
		indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
		var body []string
		for i++; i < len(lines) && !strings.HasPrefix(strings.TrimLeft(lines[i], " "), "```"); i++ {
			body = append(body, strings.TrimPrefix(lines[i], indent))
		}
		blocks = append(blocks, strings.Join(body, "\n")+"\n")
	}
	if !found && !(current == heading && (within == "" || section == within)) {
		t.Fatalf("docs/spec/%s has no heading %q under %q", page, heading, within)
	}
	if index >= len(blocks) {
		t.Fatalf("docs/spec/%s, under %q, has %d blocks and not %d",
			page, heading, len(blocks), index+1)
	}
	return blocks[index]
}

// fixture reads one of the blocks transcribed from docs/spec/.
func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// assertEqual compares two texts character for character, and shows them
// both when they differ.
func assertEqual(t *testing.T, got, want, what string) {
	t.Helper()
	if got == want {
		return
	}
	t.Errorf("%s does not match, character for character.\n--- got ---\n%s\n--- want ---\n%s",
		what, got, want)
}

// generatedAt is the one key of an envelope that a compiled program cannot
// be made to repeat: it is the clock. The shape is checked here, against
// docs/spec/contrato-json.md#números-fechas-y-ausencias, and the value is
// then replaced by the one of the example so that everything else in the
// envelope is compared as it is.
var generatedAt = regexp.MustCompile(`"generatedAt": ?"[^"]*"`)

// assertSameJSON compares two envelopes by what they mean and not by how
// they are laid out: docs/spec/contrato-json.md says an envelope is one JSON
// object with or without indentation, indistinctly, so the schema in the
// specification is written for a person to read and the program's is written
// by a marshaller.
func assertSameJSON(t *testing.T, got, want string) {
	t.Helper()
	assertInstant(t, got)
	got = generatedAt.ReplaceAllString(got, generatedAt.FindString(want))

	var a, b any
	if err := json.Unmarshal([]byte(got), &a); err != nil {
		t.Fatalf("the program's envelope is not JSON: %v\n%s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &b); err != nil {
		t.Fatalf("the specification's schema is not JSON: %v\n%s", err, want)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("the envelope does not match the schema.\n--- got ---\n%s\n--- want ---\n%s",
			got, want)
	}
}

// instant is ISO 8601 in UTC, to the second, ending in Z
// (docs/spec/contrato-json.md#números-fechas-y-ausencias).
var instant = regexp.MustCompile(`^"generatedAt": "\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z"$`)

func assertInstant(t *testing.T, envelope string) {
	t.Helper()
	found := generatedAt.FindString(envelope)
	if !instant.MatchString(found) {
		t.Errorf("generatedAt is %q, and it is an instant in UTC to the second", found)
	}
}
