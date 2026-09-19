package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"biso/internal/ops"
)

// This file holds the golden tests: every block the specification prints
// character for character lives under testdata, transcribed from its page,
// and the program's output is compared with it byte by byte. No fixture is
// ever generated from this code, which would only prove that the code does
// what it does; the last test of the file reads the same blocks straight out
// of docs/spec/ and checks that the transcription is still faithful.

func TestInitPrintsTheOutputOfTheSpecification(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	got := m.run("init", "My project", "--prefix", "MYP",
		"--at", "my-project-board", "--extensions", "trello.card").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "init-output.txt"), "the output of biso init")
	assertEqual(t, got.stderr, fixture(t, "init-notes.txt"), "the notes of biso init")
}

func TestInitWritesThePointerOfTheSpecification(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	m.run("init", "My project", "--prefix", "MYP",
		"--at", "my-project-board", "--extensions", "trello.card").assertCode(t, 0)

	// docs/spec/cmd/init.md prints this pointer for this very call.
	assertEqual(t,
		m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \"my-project-board\" }\n",
		"the project pointer")
}

func TestInitEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	got := m.run("init", "My project", "--prefix", "MYP", "--at", "my-project-board",
		"--extensions", "trello.card", "--json").assertCode(t, 0)

	assertSameJSON(t, got.stdout, fixture(t, "init-json.txt"))
	if got.stderr != "" {
		t.Errorf("with --json the notes do not travel as text on stderr, got:\n%s", got.stderr)
	}
}

// TestWherePrintsTheOutputOfTheSpecification drives the renderer with the
// very data the specification's example carries, because that board has two
// hundred and forty eight tasks and this suite creates none: what is being
// checked is the text, which is what that page fixes.
func TestWherePrintsTheOutputOfTheSpecification(t *testing.T) {
	assertEqual(t, renderWhere(specWhereResult()), fixture(t, "where-output.txt"),
		"the output of biso where")
}

func TestWherePrintsTheDiscardedCandidateOfTheSpecification(t *testing.T) {
	result := specWhereResult()
	result.Source = "the working directory is this board"
	result.Discarded = []ops.Discarded{{
		ID:     "7a1b2c3d",
		Path:   "/Volumes/work/boards/other-project-7a1b2c3d",
		Reason: "the project pointer would have named this board instead",
	}}

	assertEqual(t, renderWhere(result), fixture(t, "where-candidates.txt"),
		"the output of biso where with a discarded candidate")
}

func TestWhereEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	var stdout strings.Builder
	writeEnvelope(Streams{Stdout: &stdout}, ops.Env{Now: func() time.Time { return fixedNow }},
		"where", whereData(specWhereResult()))

	assertSameJSON(t, stdout.String(), fixture(t, "where-json.txt"))
}

// specWhereResult is the example board of docs/spec/cmd/where.md.
func specWhereResult() *ops.WhereResult {
	return &ops.WhereResult{
		ID:     "3f9a2b1c",
		Board:  "My project",
		Path:   "/Users/avilches/.biso/boards/my-project-3f9a2b1c",
		Source: "project pointer at /Users/avilches/Hub/Projects/My project",
		Me:     "@claude",
		Prefix: "MYP",
		Counts: ops.Counts{NotArchived: 248, Archived: 31, HighestEverAssigned: 290},
	}
}

func TestWhereWithoutABoardPrintsWhatItSearched(t *testing.T) {
	m := newMachine(t)

	got := m.run("where").assertCode(t, 20)

	// The fixture names the home directory of the example, and this
	// machine's is a temporary one: what the two have to agree on is
	// everything else, the cap of the search included.
	want := strings.ReplaceAll(fixture(t, "where-no-board.txt"), "/Users/avilches", m.home)
	assertEqual(t, got.stderr, want, "the no-board message of biso where")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso where")
}

func TestWhereWithAPointerThatNamesAnAbsentBoard(t *testing.T) {
	m := newMachine(t)
	m.writeFile(filepath.Join(m.dir, ".biso.json"), "{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")

	got := m.run("where").assertCode(t, 20)

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
	} {
		got := m.run(c.argv...).assertCode(t, 0)
		assertEqual(t, got.stdout, c.want, c.subject)
		if got.stderr != "" {
			t.Errorf("%s wrote to stderr: %s", c.subject, got.stderr)
		}
	}
}

func TestVersion(t *testing.T) {
	m := newMachine(t)
	for _, argv := range [][]string{{"--version"}, {"-V"}, {"where", "--version"}} {
		got := m.run(argv...).assertCode(t, 0)
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
		heading string
		index   int
	}{
		{"init-help.txt", "cmd/init.md", "`biso init --help`", 0},
		{"where-help.txt", "cmd/where.md", "`biso where --help`", 0},
		{"top-help.txt", "cmd/help.md", "La ayuda de primer nivel", 0},
		{"init-output.txt", "cmd/init.md", "Salida", 0},
		{"init-notes.txt", "cmd/init.md", "Salida", 1},
		{"init-json.txt", "cmd/init.md", "El esquema JSON", 0},
		{"where-output.txt", "cmd/where.md", "Salida", 0},
		{"where-candidates.txt", "cmd/where.md", "Cuando hay más de un candidato", 0},
		{"where-no-board.txt", "cmd/where.md", "Cuando hay más de un candidato", 1},
		{"where-unresolved.txt", "cmd/where.md", "Cuando hay más de un candidato", 2},
		{"where-json.txt", "cmd/where.md", "El esquema JSON", 0},
	} {
		block := specBlock(t, c.page, c.heading, c.index)
		assertEqual(t, fixture(t, c.fixture), block,
			c.fixture+", against "+c.page)
	}
}

// specBlock reads the index-th fenced block under a heading of a page of the
// specification.
func specBlock(t *testing.T, page, heading string, index int) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "docs", "spec", page))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	current := ""
	var blocks []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#") {
			if current == heading {
				break
			}
			current = strings.TrimSpace(strings.TrimLeft(line, "#"))
			blocks = nil
			continue
		}
		if !strings.HasPrefix(line, "```") {
			continue
		}
		var body []string
		for i++; i < len(lines) && !strings.HasPrefix(lines[i], "```"); i++ {
			body = append(body, lines[i])
		}
		blocks = append(blocks, strings.Join(body, "\n")+"\n")
	}
	if current != heading {
		t.Fatalf("docs/spec/%s has no heading %q", page, heading)
	}
	if index >= len(blocks) {
		t.Fatalf("docs/spec/%s, under %q, has %d blocks and not %d",
			page, heading, len(blocks), index+1)
	}
	return blocks[index]
}

// assertSameJSON compares two envelopes by what they mean and not by how
// they are laid out: docs/spec/contrato-json.md says an envelope is one JSON
// object with or without indentation, indistinctly, so the schema in the
// specification is written for a person to read and the program's is written
// by a marshaller.
func assertSameJSON(t *testing.T, got, want string) {
	t.Helper()
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
