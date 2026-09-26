package cli

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"biso/internal/model"
)

// testCommands is a table shaped like the real one will be, with one entry per
// family of docs/spec/familias-de-flags.md, so that every cross-cutting rule
// can be exercised without any command existing yet.
func testCommands() []CommandSpec {
	fieldFlags := []FlagSpec{
		{Name: "clear-labels", Category: Clear},
		{Name: "clear-desc", Category: Clear},
		{Name: "clear-type", Category: Clear},
		{
			Name: "replace-labels", Value: PlainValue, Repeatable: true,
			Comma: true, Category: Replace, Alphabet: TokenAlphabet, Noun: "label",
			Field: "labels",
		},
		{
			Name: "rm-labels", Value: PlainValue, Repeatable: true,
			Comma: true, Category: Remove, Alphabet: TokenAlphabet, Noun: "label",
			Field: "labels",
		},
		{
			Name: "add-labels", Short: "l", Value: PlainValue, Repeatable: true,
			Comma: true, Category: Add, Alphabet: TokenAlphabet, Noun: "label",
			Field: "labels",
		},
		{
			Name: "add-refs", Value: PlainValue, Repeatable: true,
			Comma: true, Category: Add, Field: "references",
		},
		{Name: "append-desc", Short: "d", Value: TextValue, Repeatable: true, Category: Add},
		{Name: "append-plan", Value: TextValue, Repeatable: true, Category: Add},
		{Name: "append-note", Value: TextValue, Repeatable: true, Category: Add},
		{Name: "add-ac", Value: TextValue, Repeatable: true, Category: Add, Field: "acceptanceCriteria"},
		{Name: "title", Short: "t", Value: TextValue, Category: Scalar, SingleLine: true},
		{Name: "status", Short: "s", Value: PlainValue, Category: Scalar, ClosedVocabulary: true},
		{Name: "type", Value: PlainValue, Category: Scalar, ClosedVocabulary: true, ClearFlag: "clear-type"},
		{Name: "author", Value: PlainValue, Category: Scalar, SingleLine: true, ClearFlag: "clear-author"},
		{Name: "due", Value: PlainValue, Category: Scalar, ClearFlag: "clear-due"},
		{Name: "check-ac", Value: PlainValue, Repeatable: true, Category: CheckAC},
		{
			Name: "set-comment-date", Value: PlainValue, Repeatable: true,
			Category: CommentDate, Pair: PairAtLastEquals, PairSyntax: "<sel>=<instant>",
		},
		{Name: "comment", Value: TextValue, Repeatable: true, Category: AddComment},
		{Name: "comment-author", Value: PlainValue, Category: NotAChange, Requires: []string{"comment"}},
	}
	return []CommandSpec{
		{Name: "set", Flags: fieldFlags},
		{Name: "ls", ReadOnly: true, Flags: []FlagSpec{
			{Name: "limit", Value: PlainValue, Conflicts: []string{"all"}},
			{Name: "mine"},
			{Name: "all", Conflicts: []string{"limit"}},
		}},
		{Name: "init", AffectsNoTask: true, Flags: []FlagSpec{
			{Name: "at", Value: PlainValue},
		}},
		// doctor is the command of docs/spec/cmd/flags-globales.md that is
		// read-only until --fix turns it into a writing one.
		{Name: "doctor", ReadOnly: true, AffectsNoTask: true,
			WriteFlags: []string{"fix"},
			Flags:      []FlagSpec{{Name: "fix"}}},
	}
}

func parse(t *testing.T, argv ...string) (*Parsed, error) {
	t.Helper()
	return Parse(argv, testCommands(), Env{Stdin: strings.NewReader("")})
}

func mustParse(t *testing.T, argv ...string) *Parsed {
	t.Helper()
	p, err := parse(t, argv...)
	if err != nil {
		t.Fatalf("parse(%q) failed: %v", argv, err)
	}
	return p
}

func bisoError(t *testing.T, err error) *model.Error {
	t.Helper()
	var e *model.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is %T, want *model.Error: %v", err, err)
	}
	return e
}

// wantError asserts the exit code and the code of a failed parse, which is
// what docs/spec/codigos-de-salida.md promises a caller can branch on without
// reading the message.
func wantError(t *testing.T, err error, exitCode int, code string) *model.Error {
	t.Helper()
	if err == nil {
		t.Fatalf("parse succeeded, want error %d/%s", exitCode, code)
	}
	e := bisoError(t, err)
	if e.ExitCode != exitCode || e.Code != code {
		t.Fatalf("got %d/%s (%q), want %d/%s", e.ExitCode, e.Code, e.Message, exitCode, code)
	}
	return e
}

func TestCommandAndPositionals(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "MYP-2", "--add-labels", "urgent")
	if p.Command != "set" {
		t.Errorf("command is %q, want set", p.Command)
	}
	if got := strings.Join(p.Positionals, ","); got != "MYP-1,MYP-2" {
		t.Errorf("positionals are %q, want MYP-1,MYP-2", got)
	}
}

func TestNoArgumentsIsNoCommand(t *testing.T) {
	p := mustParse(t)
	if p.Command != "" || p.Action != ActionNone {
		t.Errorf("got command %q action %v, want the empty invocation", p.Command, p.Action)
	}
}

func TestUnknownCommand(t *testing.T) {
	_, err := parse(t, "sett", "MYP-1")
	e := wantError(t, err, 2, "unknown_command")
	if e.Message != `unknown command: "sett"` {
		t.Errorf("message is %q", e.Message)
	}
}

func TestUnknownFlag(t *testing.T) {
	for _, argv := range [][]string{
		{"set", "--nope"},
		{"set", "--nope=1"},
		{"set", "-Z"},
		{"--type", "bug", "set"}, // a command's flag cannot precede its command
	} {
		_, err := Parse(argv, testCommands(), Env{})
		wantError(t, err, 2, "unknown_flag")
	}
}

func TestGlobalFlagsOnEitherSideOfTheCommand(t *testing.T) {
	for _, argv := range [][]string{
		{"--json", "ls"},
		{"ls", "--json"},
	} {
		p, err := Parse(argv, testCommands(), Env{})
		if err != nil {
			t.Fatalf("parse(%q): %v", argv, err)
		}
		if !p.Has("json") || p.Command != "ls" {
			t.Errorf("parse(%q) lost the global flag or the command", argv)
		}
	}
}

// TestNoCommandRedefinesAGlobalFlag holds the rule of
// docs/spec/cmd/flags-globales.md that no command may redefine a global flag
// or change what it means: a table that names one anyway is not what the
// parser reads.
func TestNoCommandRedefinesAGlobalFlag(t *testing.T) {
	commands := []CommandSpec{{Name: "ls", Flags: []FlagSpec{
		{Name: "json", Value: PlainValue, Category: Scalar},
	}}}
	_, err := Parse([]string{"ls", "--json=yes"}, commands, Env{})
	wantError(t, err, 2, "unexpected_argument")
}

// TestHelpAndVersionStopAtOnce covers the rule of
// docs/spec/cmd/flags-globales.md that --version and --help are not a mode:
// they are an action that ends the program as soon as it is read.
func TestHelpAndVersionStopAtOnce(t *testing.T) {
	p := mustParse(t, "set", "--help", "--nope")
	if p.Action != ActionHelp || p.Command != "set" {
		t.Errorf("got action %v command %q, want help for set", p.Action, p.Command)
	}
	p = mustParse(t, "-V")
	if p.Action != ActionVersion {
		t.Errorf("got action %v, want version", p.Action)
	}
	if _, err := parse(t, "set", "--nope", "--help"); err == nil {
		t.Error("an unknown flag before --help must still fail")
	}
}

func TestMissingValue(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--status")
	e := wantError(t, err, 2, "missing_value")
	if e.Message != "--status requires a value" {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "status" {
		t.Errorf("field is %q, want status", e.Field)
	}
}

func TestSwitchTakesNoValue(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--json=true")
	e := wantError(t, err, 2, "unexpected_argument")
	if e.Message != "--json takes no value" {
		t.Errorf("message is %q", e.Message)
	}
}

// TestValueStartingWithHyphen is the third mechanism of
// docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion: a value that
// starts with a hyphen behind a flag that demands one is taken as it is, with
// no heuristics.
func TestValueStartingWithHyphen(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--append-note", "-x")
	if got := p.Values("append-note"); len(got) != 1 || got[0] != "-x" {
		t.Errorf("note is %q, want [-x]", got)
	}
	p = mustParse(t, "set", "MYP-1", "--append-desc=-5 degrees")
	if got := p.Values("append-desc"); len(got) != 1 || got[0] != "-5 degrees" {
		t.Errorf("desc is %q", got)
	}
}

// TestForgottenValueIsFoundByWhatIsLeftOver is the consequence the same
// section spells out: biso set MYP-1 --append-note --priority high stores the
// note "--priority" and then fails on what is left over.
func TestForgottenValueIsFoundByWhatIsLeftOver(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--append-note", "--status", "high")
	e := wantError(t, err, 2, "unexpected_argument")
	if e.Message != "unexpected argument: high" {
		t.Errorf("message is %q", e.Message)
	}
}

func TestDoubleHyphenEndsOptionParsing(t *testing.T) {
	p := mustParse(t, "set", "--", "-n is not a flag", "--json")
	if got := strings.Join(p.Positionals, "|"); got != "-n is not a flag|--json" {
		t.Errorf("positionals are %q", got)
	}
	if p.Has("json") {
		t.Error("--json after -- must not be read as a flag")
	}
}

// TestPositionalsComeBeforeTheFlags holds the rule of
// docs/spec/valores-de-entrada.md#cómo-se-lee-la-línea-de-comandos: the
// positional arguments of a call all come between the command name and the
// first flag of that command. Every signature and every example of
// docs/spec/cmd/ is written that way, and the rule is what turns a forgotten
// value into an error instead of into a stored flag name.
func TestPositionalsComeBeforeTheFlags(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "MYP-2", "--add-labels", "a")
	if got := strings.Join(p.Positionals, ","); got != "MYP-1,MYP-2" {
		t.Errorf("positionals are %q", got)
	}

	// A global flag in front of the command name does not close anything,
	// because nothing can be positional there: biso -C dir set MYP-1 works.
	p = mustParse(t, "-C", "/tmp", "set", "MYP-1")
	if got := strings.Join(p.Positionals, ","); got != "MYP-1" {
		t.Errorf("positionals are %q, want MYP-1", got)
	}

	// Behind the flags, a loose argument is an error, whether it is the only
	// one or a second block.
	for _, argv := range [][]string{
		{"set", "--add-labels", "a", "MYP-1"},
		{"set", "MYP-1", "--add-labels", "a", "MYP-2"},
	} {
		_, err := parse(t, argv...)
		e := wantError(t, err, 2, "unexpected_argument")
		if !strings.HasPrefix(e.Message, "unexpected argument: MYP-") {
			t.Errorf("parse(%q) said %q", argv, e.Message)
		}
	}

	// And "--" is still the escape that always works.
	p = mustParse(t, "set", "--add-labels", "a", "--", "MYP-1")
	if got := strings.Join(p.Positionals, ","); got != "MYP-1" {
		t.Errorf("positionals are %q, want MYP-1", got)
	}
}

// TestShortFlagsAreNeitherGroupedNorGlued covers the rule of the same section:
// a line that can be read in more than one way stops being predictable, so -qV
// and -lparser are one unknown flag and not two flags or a flag with its value.
func TestShortFlagsAreNeitherGroupedNorGlued(t *testing.T) {
	for _, argv := range [][]string{
		{"ls", "-qV"},
		{"set", "MYP-1", "-lparser"},
	} {
		_, err := parse(t, argv...)
		wantError(t, err, 2, "unknown_flag")
	}
}

// TestAMessageNamesTheLongForm covers the rule that the text of an error does
// not depend on how the call was typed.
func TestAMessageNamesTheLongForm(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "-l", "urgent!")
	e := wantError(t, err, 2, "malformed_label")
	if e.Message != `malformed label: "urgent!"` || e.Field != "labels" {
		t.Errorf("message/field are %q/%q", e.Message, e.Field)
	}
}

func TestListFlagAccumulates(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "a", "--add-labels", "b")
	if got := strings.Join(p.Values("add-labels"), ","); got != "a,b" {
		t.Errorf("labels are %q, want a,b", got)
	}
	p = mustParse(t, "set", "MYP-1", "--add-labels", "a,b")
	if got := strings.Join(p.Values("add-labels"), ","); got != "a,b" {
		t.Errorf("labels are %q, want a,b", got)
	}
	p = mustParse(t, "set", "MYP-1", "-l", "a,b", "--add-labels", "c")
	if got := strings.Join(p.Values("add-labels"), ","); got != "a,b,c" {
		t.Errorf("labels are %q, want a,b,c", got)
	}
}

func TestEscapedComma(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-refs", `notes/a\,b.md`)
	if got := p.Values("add-refs"); len(got) != 1 || got[0] != "notes/a,b.md" {
		t.Errorf("refs are %q, want one notes/a,b.md", got)
	}
}

// TestEscapedBackslash covers the second escape of
// docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas: a
// double backslash is one literal backslash, which is the only way a value can
// end in a backslash right before a separating comma. A backslash in front of
// anything else stays in the value, so a path or a URL keeps what it carries.
func TestEscapedBackslash(t *testing.T) {
	for _, c := range []struct {
		given string
		want  []string
	}{
		{`C:\\dir\\,notes.md`, []string{`C:\dir\`, "notes.md"}},
		{`a\\\,b`, []string{`a\,b`}},
		{`notes\x.md`, []string{`notes\x.md`}},
		{`ends-in\\`, []string{`ends-in\`}},
		{`lonely\`, []string{`lonely\`}},
	} {
		p := mustParse(t, "set", "MYP-1", "--add-refs", c.given)
		if got := p.Values("add-refs"); strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("--add-refs %s gave %q, want %q", c.given, got, c.want)
		}
	}
}

// TestTheSameValueThreeTimesWarnsOnce covers
// docs/spec/salida-y-terminal.md#notas-y-avisos: one warning per repeated
// value, counting the appearances, instead of one warning per repetition all
// saying "twice".
func TestTheSameValueThreeTimesWarnsOnce(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "urgent,urgent,urgent")
	if got := strings.Join(p.Values("add-labels"), ","); got != "urgent" {
		t.Errorf("labels are %q, want urgent", got)
	}
	w := onlyWarning(t, p)
	if w.Message != `--add-labels: "urgent" given 3 times, kept once` {
		t.Errorf("warning is %q", w.Message)
	}

	// Two values that each repeat get one warning each, in the order they
	// were found.
	p = mustParse(t, "set", "MYP-1", "--add-labels", "a,b,a,b")
	if len(p.Warnings) != 2 {
		t.Fatalf("got %d warnings, want two: %v", len(p.Warnings), p.Warnings)
	}
	for i, want := range []string{
		`--add-labels: "a" given twice, kept once`,
		`--add-labels: "b" given twice, kept once`,
	} {
		if p.Warnings[i].Message != want {
			t.Errorf("warning %d is %q, want %q", i, p.Warnings[i].Message, want)
		}
	}
}

func TestRepeatedValueInAListWarnsAndIsKeptOnce(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "urgent,urgent")
	if got := strings.Join(p.Values("add-labels"), ","); got != "urgent" {
		t.Errorf("labels are %q, want urgent", got)
	}
	w := onlyWarning(t, p)
	if w.Code != "duplicate_flag_value" {
		t.Fatalf("warning code is %q", w.Code)
	}
	if w.Message != `--add-labels: "urgent" given twice, kept once` {
		t.Errorf("warning is %q", w.Message)
	}
}

// TestScalarGivenTwice covers the two halves of the rule in
// docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas.
func TestScalarGivenTwice(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--status", "In Progress", "--status", "Done")
	e := wantError(t, err, 2, "duplicate_scalar_flag")
	if e.Message != `--status given twice with different values: "In Progress" and "Done"` {
		t.Errorf("message is %q", e.Message)
	}
	p := mustParse(t, "set", "MYP-1", "--status", "Done", "--status", "Done")
	if got := p.Values("status"); len(got) != 1 || got[0] != "Done" {
		t.Errorf("status is %q, want one Done", got)
	}
}

func onlyWarning(t *testing.T, p *Parsed) Warning {
	t.Helper()
	if len(p.Warnings) != 1 {
		t.Fatalf("got %d warnings, want exactly one: %v", len(p.Warnings), p.Warnings)
	}
	return p.Warnings[0]
}

// TestEmptyValue walks docs/spec/valores-de-entrada.md#el-valor-vacío row by
// row: a value of nothing but spaces is empty wherever it came from.
func TestEmptyValue(t *testing.T) {
	// A flag that adds: nothing is added, it warns, and the exit code stays 0.
	p := mustParse(t, "set", "MYP-1", "--append-note", "   ")
	if got := p.Values("append-note"); len(got) != 0 {
		t.Errorf("note is %q, want nothing added", got)
	}
	w := onlyWarning(t, p)
	if w.Code != "empty_append" || w.Message != "--append-note: empty value, nothing was added" {
		t.Errorf("warning is %q/%q", w.Code, w.Message)
	}

	// A flag that replaces: it leaves the field empty, like --clear-labels.
	p = mustParse(t, "set", "MYP-1", "--replace-labels", "")
	if len(p.Values("replace-labels")) != 0 || !p.Emptied("replace-labels") {
		t.Errorf("--replace-labels %q did not empty the field", "")
	}
	if len(p.Warnings) != 0 {
		t.Errorf("emptying a list is explicit, it warns about nothing: %v", p.Warnings)
	}

	// A scalar with no vocabulary: error 3, and the hint names the flag that
	// does empty it.
	_, err := parse(t, "set", "MYP-1", "--author", "")
	e := wantError(t, err, 3, "empty_scalar_value")
	if e.Message != "--author cannot be empty" {
		t.Errorf("message is %q", e.Message)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "to clear it, use --clear-author" {
		t.Errorf("hints are %q", e.Hints)
	}

	// A scalar with a closed vocabulary: the parser lets it through, because
	// internal/match is what answers with unknown_status.
	p = mustParse(t, "set", "MYP-1", "--status", "")
	if got := p.Values("status"); len(got) != 1 || got[0] != "" {
		t.Errorf("status is %q, want one empty value left for the vocabulary", got)
	}
}

// TestEmptyValueOutsideATaskScalar is the rule of
// docs/spec/valores-de-entrada.md#cómo-se-lee-la-línea-de-comandos: an empty
// value in a flag that neither adds nor replaces nor is a scalar of a task is
// a malformed command line, exit code 2, and not the exit code 3 of a value
// the board's vocabulary does not recognize. --rm-labels removes nothing from
// a name that is not there, and --cwd names no directory.
func TestEmptyValueOutsideATaskScalar(t *testing.T) {
	for _, argv := range [][]string{
		{"set", "MYP-1", "--rm-labels", ""},
		{"set", "MYP-1", "--rm-labels", "   "},
		{"ls", "--cwd", ""},
		{"ls", "--limit", ""},
	} {
		_, err := parse(t, argv...)
		e := wantError(t, err, 2, "unexpected_argument")
		if !strings.HasSuffix(e.Message, " cannot be empty") {
			t.Errorf("parse(%q) said %q", argv, e.Message)
		}
	}
}

// TestPairWithAnEmptyValue is the same rule for a flag whose value is a pair:
// --set-comment-date 3= says nothing the specification documents: the date of
// a comment is corrected, never emptied.
func TestPairWithAnEmptyValue(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--set-comment-date", "3=")
	e := wantError(t, err, 2, "unexpected_argument")
	if e.Message != `--set-comment-date: the value of key "3" cannot be empty` {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "set-comment-date" || e.Given != "3=" {
		t.Errorf("field/given are %q/%q", e.Field, e.Given)
	}
}

func TestEmptyValueAfterASplit(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "a,")
	if got := strings.Join(p.Values("add-labels"), ","); got != "a" {
		t.Errorf("labels are %q, want a", got)
	}
	if onlyWarning(t, p).Code != "empty_append" {
		t.Errorf("warnings are %v", p.Warnings)
	}
}

// TestTheThreeForms covers
// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo.
func TestTheThreeForms(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plan.md")
	if err := os.WriteFile(path, []byte("from the file\r\nsecond line\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := mustParse(t, "set", "MYP-1", "--append-plan", "literal")
	if got := p.Values("append-plan"); got[0] != "literal" {
		t.Errorf("plan is %q", got)
	}

	p = mustParse(t, "set", "MYP-1", "--append-plan", "@"+path)
	if got := p.Values("append-plan"); got[0] != "from the file\nsecond line\n" {
		t.Errorf("plan is %q, want the file with its line endings normalized", got)
	}

	p, err := Parse(
		[]string{"set", "MYP-1", "--append-plan", "-"},
		testCommands(),
		Env{Stdin: strings.NewReader("from stdin")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Values("append-plan"); got[0] != "from stdin" {
		t.Errorf("plan is %q", got)
	}
}

// TestStdinIsTheValueAndNotTheSpelling holds the decision that "-" means
// standard input wherever it arrives from: the three forms describe the value,
// not how it was attached to its flag, so --append-plan=- reads stdin exactly
// like --append-plan - does.
func TestStdinIsTheValueAndNotTheSpelling(t *testing.T) {
	p, err := Parse(
		[]string{"set", "MYP-1", "--append-plan=-"},
		testCommands(),
		Env{Stdin: strings.NewReader("from stdin")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := p.Values("append-plan"); got[0] != "from stdin" {
		t.Errorf("plan is %q", got)
	}
}

// TestDashDashAfterFlagsStillOpensPositionals is why "--" stays the escape
// that always works: the block of positional arguments has to be one, but what
// comes behind "--" is a positional whatever the rest of the line did.
func TestDashDashAfterFlagsStillOpensPositionals(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "a", "--", "-x")
	if got := strings.Join(p.Positionals, "|"); got != "MYP-1|-x" {
		t.Errorf("positionals are %q", got)
	}
}

func TestDoubleAtIsTheOnlyEscape(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--append-plan", "@@literal")
	if got := p.Values("append-plan"); got[0] != "@literal" {
		t.Errorf("plan is %q, want @literal", got)
	}
}

// TestPersonFieldsNeverReadTheAt is the second rule of the same section:
// --comment-author @trello:juan stores that text and reads no file.
func TestPersonFieldsNeverReadTheAt(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--comment", "x", "--comment-author", "@trello:juan")
	if got := p.Values("comment-author"); got[0] != "@trello:juan" {
		t.Errorf("author is %q", got)
	}
}

func TestFileThatDoesNotExist(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--append-desc", "@docs/x.md")
	e := wantError(t, err, 4, "file_not_found")
	if e.Message != "--append-desc: file not found: docs/x.md" {
		t.Errorf("message is %q", e.Message)
	}
}

func TestFileThatCannotBeRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "locked.md")
	if err := os.WriteFile(path, []byte("x"), 0o000); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(path); err == nil {
		t.Skip("this user can read a file with no permissions")
	}
	_, err := parse(t, "set", "MYP-1", "--append-desc", "@"+path)
	wantError(t, err, 8, "file_unreadable")
}

// TestOnlyOneStdinPerInvocation covers the rule that "-" may appear once per
// invocation and not once per flag, because the second one would read an
// exhausted stream and store the empty value without anyone noticing.
func TestOnlyOneStdinPerInvocation(t *testing.T) {
	_, err := Parse(
		[]string{"set", "MYP-1", "--append-desc", "-", "--append-plan", "-"},
		testCommands(),
		Env{Stdin: strings.NewReader("only once")},
	)
	e := wantError(t, err, 2, "two_stdin")
	if e.Message != "- can be given only once per invocation; --append-desc and --append-plan both read stdin" {
		t.Errorf("message is %q", e.Message)
	}

	// One repeatable flag asking twice has its own text: naming the same
	// flag on both sides of "and" would read like a bug in the message.
	_, err = Parse(
		[]string{"set", "MYP-1", "--append-desc", "-", "--append-desc", "-"},
		testCommands(),
		Env{Stdin: strings.NewReader("only once")},
	)
	e = wantError(t, err, 2, "two_stdin")
	if e.Message != "- can be given only once per invocation; --append-desc reads stdin twice" {
		t.Errorf("message is %q", e.Message)
	}
}

// failingReader is standard input that breaks while it is being read, which is
// the environment failing and not the request.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("broken pipe") }

func TestStdinThatCannotBeRead(t *testing.T) {
	_, err := Parse(
		[]string{"set", "MYP-1", "--append-desc", "-"},
		testCommands(),
		Env{Stdin: failingReader{}},
	)
	e := wantError(t, err, 8, "io_error")
	if e.Message != "--append-desc: stdin cannot be read: broken pipe" {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "append-desc" || e.Given != "-" {
		t.Errorf("field/given are %q/%q", e.Field, e.Given)
	}
}

// TestInvalidUTF8InAnArgument covers the first half of
// docs/spec/salida-y-terminal.md#codificación-y-texto: the encoding of every
// argument is judged before the line is read at all, so the undecodable byte
// never reaches a message, a field or the JSON envelope. It travels escaped
// instead, and it is the argument that is named, because at that point nobody
// knows yet which flag it belonged to.
func TestInvalidUTF8InAnArgument(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--append-note", "ok\xffbad")
	e := wantError(t, err, 3, "invalid_encoding")
	if e.Message != `invalid UTF-8 in argument 4 at byte 2: "ok\xffbad"` {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "argument" || e.Given != `ok\xffbad` {
		t.Errorf("field/given are %q/%q", e.Field, e.Given)
	}
	if strings.ContainsRune(e.Message+e.Given, utf8.RuneError) ||
		!utf8.ValidString(e.Message) || !utf8.ValidString(e.Given) {
		t.Errorf("the invalid byte reached the error: %q / %q", e.Message, e.Given)
	}
	// The command name and the spelling of a flag are judged too, and not
	// only the values.
	for _, argv := range [][]string{
		{"se\xfft", "MYP-1"},
		{"set", "MYP-1", "--append-\xffnote", "x"},
	} {
		_, err := parse(t, argv...)
		wantError(t, err, 3, "invalid_encoding")
	}
}

// TestInvalidUTF8InAFile is the other half: what arrives from a file or from
// standard input is judged once it is read, and there the error does name the
// flag, with what was typed behind it in given.
func TestInvalidUTF8InAFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "desc.md")
	if err := os.WriteFile(path, []byte("ok\xffbad"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := parse(t, "set", "MYP-1", "--append-desc", "@"+path)
	e := wantError(t, err, 3, "invalid_encoding")
	if e.Message != "--append-desc: invalid UTF-8 at byte 2" {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "append-desc" || e.Given != "@"+path {
		t.Errorf("field/given are %q/%q", e.Field, e.Given)
	}
}

func TestMalformedToken(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--add-labels", "urgent!")
	e := wantError(t, err, 2, "malformed_label")
	if e.Message != `malformed label: "urgent!"` {
		t.Errorf("message is %q", e.Message)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "a label may contain letters, digits, and - _ . : @" {
		t.Errorf("hints are %q", e.Hints)
	}
}

func TestMalformedStringValue(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--title", "first line\nsecond line")
	e := wantError(t, err, 2, "malformed_string_value")
	if e.Message != `malformed title: "first line\nsecond line"` {
		t.Errorf("message is %q", e.Message)
	}
	if e.Field != "title" {
		t.Errorf("field is %q", e.Field)
	}

	// The rule belongs to the field and not to one flag: --author is the
	// other flag whose value is one line.
	_, err = parse(t, "set", "MYP-1", "--author", "first\rsecond")
	wantError(t, err, 2, "malformed_string_value")
}

func TestPairs(t *testing.T) {
	// --set-comment-date cuts at the last "=", because its selector is free
	// text that may carry another.
	p := mustParse(t, "set", "MYP-1", "--set-comment-date", "a=b=2026-08-14T10:22:00Z")
	c := onlyChange(t, p, "set-comment-date")
	if c.Key != "a=b" || c.Value != "2026-08-14T10:22:00Z" {
		t.Errorf("got %q=%q", c.Key, c.Value)
	}

	_, err := parse(t, "set", "MYP-1", "--set-comment-date", "trello")
	e := wantError(t, err, 2, "unexpected_argument")
	if e.Message != `--set-comment-date: expected <sel>=<instant>, got "trello"` {
		t.Errorf("message is %q", e.Message)
	}
}

// TestSameCommentDateKeyTwice is the rule of
// docs/spec/familias-de-flags.md#comentarios: the same comment with two
// different instants is the error of a repeated scalar, and with the same
// instant it applies once and says nothing.
func TestSameCommentDateKeyTwice(t *testing.T) {
	_, err := parse(t, "set", "MYP-1",
		"--set-comment-date", "3=2026-08-14T10:22:00Z",
		"--set-comment-date", "3=2026-08-15T10:22:00Z")
	e := wantError(t, err, 2, "duplicate_scalar_flag")
	want := `--set-comment-date: key "3" given twice with different values: ` +
		`"2026-08-14T10:22:00Z" and "2026-08-15T10:22:00Z"`
	if e.Message != want {
		t.Errorf("message is %q", e.Message)
	}

	p := mustParse(t, "set", "MYP-1",
		"--set-comment-date", "3=2026-08-14T10:22:00Z",
		"--set-comment-date", "3=2026-08-14T10:22:00Z")
	c := onlyChange(t, p, "set-comment-date")
	if c.Key != "3" || c.Value != "2026-08-14T10:22:00Z" {
		t.Errorf("got %q=%q", c.Key, c.Value)
	}
	if len(p.Warnings) != 0 {
		t.Errorf("the same instant twice says nothing: %v", p.Warnings)
	}

	// Two different comments in the same call are two changes, not a
	// repetition.
	p = mustParse(t, "set", "MYP-1",
		"--set-comment-date", "3=2026-08-14T10:22:00Z",
		"--set-comment-date", "4=2026-08-15T10:22:00Z")
	if len(p.Changes()) != 2 {
		t.Errorf("got %d changes, want two", len(p.Changes()))
	}
}

func onlyChange(t *testing.T, p *Parsed, name string) Change {
	t.Helper()
	var out []Change
	for _, c := range p.Changes() {
		if c.Flag.Name == name {
			out = append(out, c)
		}
	}
	if len(out) != 1 {
		t.Fatalf("got %d changes for %s, want one", len(out), name)
	}
	return out[0]
}

// TestChangesAreOrderedByCategory covers
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura: the
// parser classifies the changes into the fixed steps instead of applying them
// in the order of argv, and within a step it does keep that order.
func TestChangesAreOrderedByCategory(t *testing.T) {
	p := mustParse(t,
		"set", "MYP-1",
		"--comment", "last",
		"--add-labels", "b",
		"--status", "Done",
		"--clear-labels",
		"--add-labels", "a",
		"--rm-labels", "old",
		"--check-ac", "all",
		"--replace-labels", "r",
		"--set-comment-date", "3=2026-08-14T10:22:00Z",
	)
	var got []string
	for _, c := range p.Changes() {
		got = append(got, c.Flag.Name)
	}
	want := []string{
		"clear-labels",
		"replace-labels",
		"rm-labels",
		"add-labels", "add-labels",
		"status",
		"check-ac",
		"set-comment-date",
		"comment",
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("order is\n %v\nwant\n %v", got, want)
	}
	// Within one step, the order of the command line stands.
	var labels []string
	for _, c := range p.Changes() {
		if c.Flag.Name == "add-labels" {
			labels = append(labels, c.Value)
		}
	}
	if strings.Join(labels, ",") != "b,a" {
		t.Errorf("labels are %q, want b,a", labels)
	}
}

func TestIncompatibleGlobalFlags(t *testing.T) {
	for _, argv := range [][]string{
		{"ls", "--json", "--quiet"},
		{"ls", "--json", "--print"},
		{"ls", "--quiet", "--print"},
	} {
		_, err := Parse(argv, testCommands(), Env{})
		wantError(t, err, 2, "incompatible_flags")
	}
}

// TestIncompatibleFlagsAreNamedInTableOrder holds the rule that the text of an
// error does not depend on how the call was written: the pair is named in the
// order of the specification table, whichever of the two was typed first. The
// order is the global table first and then the command's own, each in the
// order of its own table of parameters.
func TestIncompatibleFlagsAreNamedInTableOrder(t *testing.T) {
	for _, c := range []struct {
		argv []string
		want string
	}{
		{[]string{"ls", "--json", "--quiet"}, "--json and --quiet cannot be used together"},
		{[]string{"ls", "--quiet", "--json"}, "--json and --quiet cannot be used together"},
		{[]string{"ls", "--limit", "10", "--all"}, "--limit and --all cannot be used together"},
		{[]string{"ls", "--all", "--limit", "10"}, "--limit and --all cannot be used together"},
	} {
		_, err := Parse(c.argv, testCommands(), Env{})
		e := wantError(t, err, 2, "incompatible_flags")
		if e.Message != c.want {
			t.Errorf("parse(%q) said %q, want %q", c.argv, e.Message, c.want)
		}
	}
}

// TestIncompatibleFlagsCarryNoField is the exception of
// docs/spec/contrato-json.md#los-errores-en-json: an error that names a pair of
// flags, neither of them more to blame than the other, has no field to name.
func TestIncompatibleFlagsCarryNoField(t *testing.T) {
	_, err := Parse([]string{"ls", "--json", "--quiet"}, testCommands(), Env{})
	e := wantError(t, err, 2, "incompatible_flags")
	if e.Field != "" || e.Given != "" {
		t.Errorf("field/given are %q/%q, want neither", e.Field, e.Given)
	}
}

// TestGlobalFlagsIsACopy covers the promise of GlobalFlags: a caller that
// walks the global table cannot change what the parser reads.
func TestGlobalFlagsIsACopy(t *testing.T) {
	table := GlobalFlags()
	if len(table) == 0 {
		t.Fatal("the global table is empty")
	}
	// Every conflict declared in the table is declared on both sides, which
	// is what makes the message independent of the order of the call.
	byName := map[string]FlagSpec{}
	for _, f := range table {
		byName[f.Name] = f
	}
	for _, f := range table {
		for _, name := range f.Conflicts {
			other, ok := byName[name]
			if !ok {
				t.Fatalf("--%s conflicts with an unknown --%s", f.Name, name)
			}
			if !contains(other.Conflicts, f.Name) {
				t.Errorf("--%s conflicts with --%s, but not the other way around", f.Name, name)
			}
		}
	}
	table[0].Name = "changed"
	if GlobalFlags()[0].Name == "changed" {
		t.Error("GlobalFlags handed out the table itself")
	}
}

func TestRequires(t *testing.T) {
	_, err := parse(t, "set", "MYP-1", "--comment-author", "@sara")
	e := wantError(t, err, 2, "incompatible_flags")
	if e.Message != "--comment-author requires --comment" {
		t.Errorf("message is %q", e.Message)
	}
}

// TestDryRunAndPrintWhereTheyDoNotApply covers the two rules of
// docs/spec/cmd/flags-globales.md, each with its own literal message.
func TestDryRunAndPrintWhereTheyDoNotApply(t *testing.T) {
	_, err := parse(t, "ls", "--dry-run")
	e := wantError(t, err, 2, "read_only_flag")
	if e.Message != "--dry-run does not apply to a read-only command" {
		t.Errorf("message is %q", e.Message)
	}

	_, err = parse(t, "ls", "--print")
	e = wantError(t, err, 2, "read_only_flag")
	if e.Message != "--print does not apply to a command that affects no task" {
		t.Errorf("message is %q", e.Message)
	}

	// The error names the flag, so it carries its field like every other
	// error of exit code 2 that names one.
	if e.Field != "print" {
		t.Errorf("field is %q, want print", e.Field)
	}

	// init writes, so --dry-run is valid there, and --print is not.
	mustParse(t, "init", "--dry-run")
	_, err = parse(t, "init", "--print")
	wantError(t, err, 2, "read_only_flag")

	// A write flag in a read-only command is the same family of error.
	mustParse(t, "set", "MYP-1", "--dry-run")
}

// TestAFlagCanTurnAReadOnlyCommandIntoAWritingOne is what
// docs/spec/cmd/flags-globales.md declares about biso doctor: without --fix it
// is read-only and --dry-run is a usage error there, and with --fix it writes
// and the flag behaves as it does anywhere else.
func TestAFlagCanTurnAReadOnlyCommandIntoAWritingOne(t *testing.T) {
	_, err := parse(t, "doctor", "--dry-run")
	e := wantError(t, err, 2, "read_only_flag")
	if e.Message != "--dry-run does not apply to a read-only command" || e.Field != "dry-run" {
		t.Errorf("message/field are %q/%q", e.Message, e.Field)
	}
	mustParse(t, "doctor", "--fix", "--dry-run")
	// The order of the two does not matter: the answer comes once the whole
	// line has been read.
	mustParse(t, "doctor", "--dry-run", "--fix")
	// --fix does not make it affect a task, so --print is still an error.
	_, err = parse(t, "doctor", "--fix", "--print")
	wantError(t, err, 2, "read_only_flag")
}

func TestClosedDomain(t *testing.T) {
	mustParse(t, "ls", "--color", "never")
	_, err := parse(t, "ls", "--color", "sometimes")
	e := wantError(t, err, 2, "invalid_color_mode")
	if strings.Join(e.Valid, ",") != "auto,always,never" {
		t.Errorf("valid is %q", e.Valid)
	}
	if e.Given != "sometimes" || e.Field != "color" {
		t.Errorf("field/given are %q/%q", e.Field, e.Given)
	}
}

func TestLiteralNewlineWarning(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--append-desc", `line one\nline two`)
	w := onlyWarning(t, p)
	if w.Code != "literal_newline" {
		t.Fatalf("warning is %v", w)
	}
	if w.Message != `--append-desc contains a literal \n and no real newline; it will be stored as text` {
		t.Errorf("message is %q", w.Message)
	}
	if len(w.Hints) != 1 || w.Hints[0] != "use a real newline, or -d @file.md, or -d - to read from stdin" {
		t.Errorf("hints are %q", w.Hints)
	}
	// With a real newline in the same value there is nothing to warn about.
	p = mustParse(t, "set", "MYP-1", "--append-desc", "line one\nliteral \\n")
	if len(p.Warnings) != 0 {
		t.Errorf("warnings are %v", p.Warnings)
	}
}

func TestValueAccessors(t *testing.T) {
	p := mustParse(t, "set", "MYP-1", "--add-labels", "a,b", "--status", "Done")
	if !p.Has("add-labels") || p.Has("add-refs") {
		t.Error("Has answers the wrong thing")
	}
	if v, ok := p.Value("status"); !ok || v != "Done" {
		t.Errorf("Value(status) is %q/%v", v, ok)
	}
	if _, ok := p.Value("type"); ok {
		t.Error("Value of a flag that was not given must say so")
	}
	if got := strings.Join(p.Values("add-labels"), ","); got != "a,b" {
		t.Errorf("Values is %q", got)
	}
}

// countingReader is standard input that remembers whether anybody read it,
// which is the whole question of the test below.
type countingReader struct {
	reads int
}

func (r *countingReader) Read(p []byte) (int, error) {
	r.reads++
	return 0, io.EOF
}

// TestTwoStdinIsAnsweredBeforeAnythingIsRead is the other half of the rule
// above: the refusal comes from reading the command line and not from
// finding the stream exhausted, so the call that is going to be refused
// never reads it at all, and therefore never earns a warning about what it
// read (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).
func TestTwoStdinIsAnsweredBeforeAnythingIsRead(t *testing.T) {
	for _, argv := range [][]string{
		{"set", "MYP-1", "--append-desc", "-", "--append-plan", "-"},
		{"set", "MYP-1", "--append-desc", "-", "--append-desc", "-"},
	} {
		stdin := &countingReader{}
		p, err := Parse(argv, testCommands(), Env{Stdin: stdin})
		wantError(t, err, 2, "two_stdin")
		if stdin.reads != 0 {
			t.Errorf("%v read standard input %d time(s) before refusing the call",
				argv, stdin.reads)
		}
		if len(p.Warnings) != 0 {
			t.Errorf("%v produced the warnings %v before refusing the call",
				argv, p.Warnings)
		}
	}
}
