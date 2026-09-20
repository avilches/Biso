package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the rules of docs/spec/cmd/verbos-del-ciclo.md that belong to
// this layer: which flags each of the six takes, and how the positional
// texts of the four that take any are read.

// parseReal runs the real command table, which is what these tests are
// about: parse_test.go uses a table of its own precisely so that it does not
// depend on which commands exist today.
func parseReal(t *testing.T, stdin string, argv ...string) (*Parsed, error) {
	t.Helper()
	return Parse(argv, Commands(), Env{Stdin: strings.NewReader(stdin)})
}

// The first positional is the reference and every one after it is a text,
// which is what separates `biso note MYP-11 MYP-2` from `biso set A B`.
func TestTheVerbsOfTheCycleReadTheirPositionalsAsAReferenceAndTexts(t *testing.T) {
	for _, command := range []string{"note", "comment", "ask", "answer"} {
		p, err := parseReal(t, "", command, "MYP-11", "First finding", "Second finding")
		if err != nil {
			t.Fatalf("biso %s: %v", command, err)
		}
		if len(p.Positionals) != 1 || p.Positionals[0] != "MYP-11" {
			t.Errorf("biso %s: positionals = %v, want the reference alone", command, p.Positionals)
		}
		if len(p.Texts) != 2 {
			t.Fatalf("biso %s: texts = %v, want two", command, p.Texts)
		}
		if p.Texts[0].Value != "First finding" || p.Texts[1].Value != "Second finding" {
			t.Errorf("biso %s: texts = %v", command, p.Texts)
		}
	}
}

// `biso start` and `biso finish` take several references, and none of their
// positionals is a text.
func TestStartAndFinishReadEveryPositionalAsAReference(t *testing.T) {
	for _, command := range []string{"start", "finish"} {
		p, err := parseReal(t, "", command, "MYP-11", "MYP-12")
		if err != nil {
			t.Fatalf("biso %s: %v", command, err)
		}
		if len(p.Positionals) != 2 || len(p.Texts) != 0 {
			t.Errorf("biso %s: positionals = %v, texts = %v", command, p.Positionals, p.Texts)
		}
	}
}

// A positional text is a long value like any other, so @file and - work
// there exactly as they do behind --append-note.
func TestAPositionalTextTakesAFileAndStandardInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "findings.md")
	if err := os.WriteFile(path, []byte("what the benchmark printed"), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := parseReal(t, "", "note", "MYP-11", "@"+path)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Texts) != 1 || p.Texts[0].Value != "what the benchmark printed" {
		t.Errorf("texts = %v, want the contents of the file", p.Texts)
	}
	if p.Texts[0].Typed != "@"+path {
		t.Errorf("typed = %q, want what was written on the command line", p.Texts[0].Typed)
	}

	p, err = parseReal(t, "from stdin", "note", "MYP-11", "-")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Texts) != 1 || p.Texts[0].Value != "from stdin" {
		t.Errorf("texts = %v, want what standard input carried", p.Texts)
	}
}

// A file that is not there is exit code 4, and the message names the
// positional and never a flag that was not written.
func TestAPositionalTextThatNamesAnAbsentFileIsExitCodeFour(t *testing.T) {
	_, err := parseReal(t, "", "note", "MYP-11", "@/nowhere/at/all.md")

	e := bisoError(t, err)
	if e.ExitCode != 4 || e.Code != "file_not_found" {
		t.Fatalf("error = %d/%s, want 4/file_not_found", e.ExitCode, e.Code)
	}
	if !strings.HasPrefix(e.Message, "text: file not found: ") {
		t.Errorf("message = %q, want the positional named as text", e.Message)
	}
}

// The escape of docs/spec/valores-de-entrada.md works here too: "@@" writes
// a literal "@".
func TestAPositionalTextEscapesTheLeadingAt(t *testing.T) {
	p, err := parseReal(t, "", "comment", "MYP-11", "@@sara asked for this")
	if err != nil {
		t.Fatal(err)
	}
	if p.Texts[0].Value != "@sara asked for this" {
		t.Errorf("value = %q", p.Texts[0].Value)
	}
}

// Every field flag of `biso set` works in every one of the six, which is the
// promise docs/spec/cmd/verbos-del-ciclo.md opens with.
func TestEveryVerbOfTheCycleTakesEveryFieldFlag(t *testing.T) {
	commands := Commands()
	byName := map[string]*CommandSpec{}
	for i := range commands {
		byName[commands[i].Name] = &commands[i]
	}
	for _, verb := range []string{"start", "note", "comment", "finish", "ask", "answer"} {
		cmd := byName[verb]
		if cmd == nil {
			t.Fatalf("biso %s is not in the table of commands", verb)
		}
		for _, f := range fieldFlags() {
			if f.Name == "comment-author" && (verb == "ask" || verb == "answer") {
				// The two that never take it, and it is the one
				// exception the page declares.
				if lookupLong(cmd, f.Name) != nil {
					t.Errorf("biso %s takes --comment-author, and it must not", verb)
				}
				continue
			}
			if lookupLong(cmd, f.Name) == nil {
				t.Errorf("biso %s does not take --%s", verb, f.Name)
			}
		}
	}
}

// --comment-author is the author of a --comment everywhere, except in `biso
// comment`, where the positional text is the comment.
func TestCommentAuthorStandsOnItsOwnOnlyInTheCommentCommand(t *testing.T) {
	if _, err := parseReal(t, "", "comment", "MYP-11", "Something",
		"--comment-author", "@trello:juan"); err != nil {
		t.Errorf("biso comment --comment-author: %v", err)
	}

	_, err := parseReal(t, "", "set", "MYP-11", "--comment-author", "@trello:juan")
	e := bisoError(t, err)
	if e.ExitCode != 2 || e.Code != "incompatible_flags" {
		t.Errorf("biso set --comment-author alone = %d/%s, want 2/incompatible_flags",
			e.ExitCode, e.Code)
	}

	_, err = parseReal(t, "", "ask", "MYP-11", "Why?", "--comment-author", "@sara")
	e = bisoError(t, err)
	if e.ExitCode != 2 || e.Code != "unknown_flag" {
		t.Errorf("biso ask --comment-author = %d/%s, want 2/unknown_flag", e.ExitCode, e.Code)
	}
}

func TestFinishRefusesStrictAndNoChecksTogether(t *testing.T) {
	for _, argv := range [][]string{
		{"finish", "MYP-11", "--strict", "--no-checks"},
		{"finish", "MYP-11", "--no-checks", "--strict"},
	} {
		_, err := parseReal(t, "", argv...)
		e := bisoError(t, err)
		if e.ExitCode != 2 || e.Code != "incompatible_flags" {
			t.Errorf("%v = %d/%s, want 2/incompatible_flags", argv, e.ExitCode, e.Code)
		}
		if e.Message != "--strict and --no-checks cannot be used together" {
			t.Errorf("message = %q, want the two named in the order of the table", e.Message)
		}
	}
}

// The whole cycle through the program, which is what makes the six worth
// having: what `biso set` would need several calls for.
func TestTheCycleOfATaskThroughTheProgram(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run("init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run("new", "Normalize CRLF", "--add-ac", "The parser accepts CRLF").assertCode(t, 0)

	started := m.run("start", "MYP-1", "--append-plan", "1. Read the parser.").assertCode(t, 0)
	assertEqual(t, started.stdout, "MYP-1  In Progress  ac 0/1  urgency 6.8\n",
		"the status line of biso start")

	m.run("note", "MYP-1", "The parser already normalized LF").assertCode(t, 0)
	commented := m.run("comment", "MYP-1", "A user reported this").assertCode(t, 0)
	if !strings.Contains(commented.stderr, "note: comment #1 by @claude") {
		t.Errorf("stderr = %q, want the key the comment took", commented.stderr)
	}

	m.run("ask", "MYP-1", "Binary files too?").assertCode(t, 0)
	m.run("answer", "MYP-1", "Only text files.").assertCode(t, 0)

	finished := m.run("finish", "MYP-1", "--check-ac", "all",
		"--append-summary", "Normalizes CRLF").assertCode(t, 0)
	assertEqual(t, finished.stdout, "MYP-1  Done  ac 1/1  urgency 0.0\n",
		"the status line of biso finish")
	if finished.stderr != "" {
		t.Errorf("stderr = %q, and nothing was missing", finished.stderr)
	}
}

// Every one of the six answers the same envelope, so that whoever consumes
// the output does not have to tell which verb produced it
// (docs/spec/cmd/set.md#el-esquema-json).
func TestEveryVerbAnswersTheSameJSONKind(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run("init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run("new", "Normalize CRLF").assertCode(t, 0)

	for _, argv := range [][]string{
		{"start", "MYP-1"},
		{"note", "MYP-1", "A finding"},
		{"comment", "MYP-1", "Something"},
		{"ask", "MYP-1", "Why?"},
		{"answer", "MYP-1", "Because"},
		{"finish", "MYP-1"},
	} {
		got := m.run(append(argv, "--json")...).assertCode(t, 0)
		if !strings.Contains(got.stdout, "\"kind\": \"task.write\"") {
			t.Errorf("%v: stdout = %q, want the task.write envelope", argv, got.stdout)
		}
	}
}
