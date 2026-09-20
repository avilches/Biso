package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/term"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the whole of the translation the architecture puts in this
// layer: argv in, the typed parameters of a command, the call into
// internal/ops, and the result or the error out as text or as the JSON
// envelope of docs/spec/contrato-json.md. cmd/biso/main.go builds the
// Streams of the real world, calls Run and exits with what it answers, so
// there is one place that turns an error into text, into JSON and into the
// exit code of the process, and a test can drive all of it without a
// terminal and without touching the machine it runs on.

// Streams is the world outside the program.
type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	// Getenv reads an environment variable and says whether it is defined,
	// which NO_COLOR needs: having it defined at all is what counts
	// (docs/spec/invocacion.md#variables-de-entorno).
	Getenv func(string) (string, bool)

	// Dir is the process's working directory and Home the caller's home
	// directory, which is the cap of the upward search
	// (docs/spec/resolucion-del-tablero.md#el-tope-de-la-búsqueda-hacia-arriba).
	Dir  string
	Home string

	// StdoutIsTerminal and StderrIsTerminal are the only two things biso
	// ever asks about the terminal, and they are asked per stream because
	// each one has its own destination. Nothing reads them yet: no output
	// of `init` or of `where` carries color, and the commands whose output
	// does arrive later, with the answer of UseColor.
	StdoutIsTerminal bool
	StderrIsTerminal bool

	// Now and NewID are the clock and the source of board identifiers, so
	// that a test can make both of them say the same thing twice.
	Now   func() time.Time
	NewID func() (string, error)
}

func (s Streams) getenv(name string) string {
	if s.Getenv == nil {
		return ""
	}
	v, _ := s.Getenv(name)
	return v
}

// Run is the program. It answers the exit code of
// docs/spec/codigos-de-salida.md and writes everything it has to write to
// the streams it was given.
func Run(argv []string, s Streams) int {
	// Whether the error envelope is JSON is read off argv and not off the
	// analysis, because a call that fails to parse has no analysis and a
	// caller that asked for JSON is owed JSON all the same
	// (docs/spec/contrato-json.md#los-errores-en-json).
	asJSON := wantsJSON(argv)

	// An analysis that fails still answers the warnings it had already
	// found, because a warning is never suppressed
	// (docs/spec/salida-y-terminal.md#notas-y-avisos): they go with the
	// error instead of being dropped with it.
	p, err := Parse(argv, Commands(), Env{Stdin: s.Stdin})
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}

	switch p.Action {
	case ActionVersion:
		fmt.Fprintf(s.Stdout, "biso %s\n", Version)
		return 0
	case ActionHelp:
		fmt.Fprint(s.Stdout, helpOf(p.Command))
		return 0
	}
	if p.Command == "" {
		// A call that names no command has nothing to do and nothing to
		// complain about, so it prints what `biso --help` prints
		// (docs/spec/cmd/help.md#la-ayuda-de-primer-nivel).
		fmt.Fprint(s.Stdout, topLevelHelp)
		return 0
	}

	env, err := environment(s, p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}

	switch p.Command {
	case "init":
		return runInit(s, p, env)
	case "where":
		return runWhere(s, p, env)
	case "new":
		return runNew(s, p, env)
	case "ls":
		return runList(s, p, env)
	case "get":
		return runGet(s, p, env)
	case "set":
		return runSet(s, p, env)
	case "export":
		return runExport(s, p, env)
	case "snapshot":
		return runSnapshot(s, p, env)
	}
	// Parse only ever answers a command of the table, so this is
	// unreachable; answering the internal error keeps it honest.
	return fail(s, asJSON, &model.Error{
		ExitCode: 1, Code: "internal",
		Message: fmt.Sprintf("no implementation for command %q", p.Command),
	}, warningsOf(p))
}

// wantsJSON answers whether the call wrote --json, reading argv the way the
// parser does: the flag takes no value, and everything behind "--" is a
// positional argument and not a flag.
func wantsJSON(argv []string) bool {
	for _, arg := range argv {
		if arg == "--" {
			return false
		}
		if arg == "--json" {
			return true
		}
	}
	return false
}

// environment gathers what every command needs from outside itself: which
// directory the board is resolved from, this machine's configuration, and
// who is calling.
func environment(s Streams, p *Parsed) (ops.Env, error) {
	dir := s.Dir
	if v := s.getenv("BISO_CWD"); v != "" {
		dir = v
	}
	if v, ok := p.Value("cwd"); ok {
		// The flag of the command line wins over the variable, because it
		// is the most specific thing about the call
		// (docs/spec/invocacion.md#variables-de-entorno).
		dir = v
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(s.Dir, dir)
	}
	dir = filepath.Clean(dir)

	// Reading ~/.biso/config.json is internal/ops's job and not this
	// layer's: this package depends on ops and on model, and on nothing
	// below them (section 3 of
	// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md).
	return ops.NewEnv(ops.Call{
		Dir: dir, Home: s.Home, Me: s.getenv("BISO_ME"), Now: s.Now, NewID: s.NewID,
	})
}

func runInit(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	params, err := initParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.Init(env, params)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)
	switch {
	case asJSON:
		writeEnvelope(s, env, "init", initData(result))
	case p.Has("quiet"), result.DryRun:
		// With --quiet stdout carries the identifiers a write affected, and
		// this command affects no task, so it carries nothing
		// (docs/spec/cmd/flags-globales.md). A preview has nothing to show
		// either: its whole answer is the note below
		// (docs/spec/cmd/init.md).
	default:
		fmt.Fprint(s.Stdout, renderInit(result))
	}
	for _, note := range notesOf(result) {
		printNote(s, p, note)
	}
	if result.DryRun && result.Restored {
		fmt.Fprintf(s.Stderr, "%s would be created, nothing was written (--dry-run)\n",
			plural(result.RestoredTasks, "task"))
	}
	return 0
}

func runWhere(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	if len(p.Positionals) > 0 {
		return fail(s, asJSON, errUnexpectedArgument(p.Positionals[0]), warningsOf(p))
	}
	result, err := ops.Where(env)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)
	if asJSON {
		writeEnvelope(s, env, "where", whereData(result))
		return 0
	}
	fmt.Fprint(s.Stdout, renderWhere(result))
	return 0
}

// initParams turns the analyzed call into the typed parameters of the
// command, which is the last thing this layer does before the logic of
// internal/ops takes over.
func initParams(p *Parsed) (ops.InitParams, error) {
	if len(p.Positionals) > 1 {
		return ops.InitParams{}, errUnexpectedArgument(p.Positionals[1])
	}
	params := ops.InitParams{
		Statuses:        p.Values("statuses"),
		HasStatuses:     p.Has("statuses"),
		Types:           p.Values("types"),
		HasTypes:        p.Has("types"),
		Priorities:      p.Values("priorities"),
		HasPriorities:   p.Has("priorities"),
		Extensions:      p.Values("extensions"),
		HasExtensions:   p.Has("extensions"),
		OverwriteConfig: p.Has("overwrite-config"),
		DryRun:          p.Has("dry-run"),
	}
	if len(p.Positionals) == 1 {
		params.Name, params.HasName = p.Positionals[0], true
	}
	scalars := []struct {
		flag  string
		value *string
		has   *bool
	}{
		{"at", &params.At, &params.HasAt},
		{"initial-status", &params.InitialStatus, &params.HasInitialStatus},
		{"active-status", &params.ActiveStatus, &params.HasActiveStatus},
		{"terminal-status", &params.TerminalStatus, &params.HasTerminalStatus},
		{"prefix", &params.Prefix, &params.HasPrefix},
		{"from", &params.From, &params.HasFrom},
	}
	for _, sc := range scalars {
		if v, ok := p.Value(sc.flag); ok {
			*sc.value, *sc.has = v, true
		}
	}
	if params.HasFrom && params.HasName {
		// The table of docs/spec/cmd/init.md declares --from incompatible
		// with the board name, which is a positional argument and therefore
		// not a row of the flag table the parser walks.
		return ops.InitParams{}, &model.Error{
			ExitCode: 2,
			Code:     "incompatible_flags",
			Message:  "--from and a board name cannot be used together",
			Hints:    []string{"the name comes from the snapshot's board.json"},
		}
	}
	return params, nil
}

// helpOf answers the help of a command, or the top-level one when the call
// named none.
func helpOf(command string) string {
	switch command {
	case "init":
		return initHelp
	case "where":
		return whereHelp
	case "new":
		return newHelp
	case "ls":
		return lsHelp
	case "get":
		return getHelp
	case "set":
		return setHelp
	case "export":
		return exportHelp
	case "snapshot":
		return snapshotHelp
	}
	return topLevelHelp
}

// notesOf is the notes a result carries. `biso init` is the only command of
// this step that has any: the two of --at, and the one of a preview.
//
// The note of a preview says what that preview would have done, and the two
// endings are not the same thing: docs/spec/cmd/init.md fixes "board would
// be created at" for a board that does not exist yet, and a preview of
// --overwrite-config over one that does exist would create nothing.
func notesOf(result *ops.InitResult) []string {
	notes := append([]string(nil), result.Notes...)
	switch {
	case !result.DryRun:
	case result.Restored:
		// A preview of --from counts the tasks of the snapshot instead of
		// naming the board, and it does not repeat the note of the board
		// besides, so that nothing says twice that nothing was written
		// (docs/spec/cmd/init.md). The sentence itself is not a note, so
		// runInit prints it apart.
	case result.Action == ops.Rewrote:
		notes = append(notes, fmt.Sprintf(
			"the configuration of board %s at %s\n"+
				"would be rewritten, and no task would change (--dry-run)",
			result.ID, result.DryRunPath))
	default:
		notes = append(notes, fmt.Sprintf(
			"board would be created at %s (--dry-run)", result.DryRunPath))
	}
	return notes
}

// printNote writes one note: "note: " and, under it, every further line of
// the same note aligned with the first
// (docs/spec/salida-y-terminal.md#notas-y-avisos). --quiet is the one thing
// that suppresses it; --json does not, and the note keeps travelling as
// text on stderr, because it is context for whoever is reading and never
// data, so no envelope carries it.
func printNote(s Streams, p *Parsed, note string) {
	if p.Has("quiet") {
		return
	}
	fmt.Fprint(s.Stderr, prefixed("note: ", note))
}

// printWarnings prints the warnings the analysis of the command line
// produced. A warning is never suppressed, not even by --quiet: silencing
// one is the caller's business, with 2>/dev/null.
func printWarnings(s Streams, p *Parsed) {
	for _, w := range warningsOf(p) {
		fmt.Fprint(s.Stderr, prefixed("warning: ", w.Message))
		for _, hint := range w.Hints {
			fmt.Fprint(s.Stderr, prefixed("hint: ", hint))
		}
	}
}

// warningsOf is the warnings of an analysis that may not have finished: a
// call that fails to parse answers a partial Parsed, and one that fails
// before it is even attempted answers none.
func warningsOf(p *Parsed) []Warning {
	if p == nil {
		return nil
	}
	return p.Warnings
}

// fail is the one place an error becomes text or JSON and an exit code. The
// error is printed exactly as it was born, several layers down, with nothing
// reinterpreted on the way up.
//
// warnings are the ones the call had already produced, which a failure
// never swallows (docs/spec/salida-y-terminal.md#notas-y-avisos). Without
// --json they are printed as text, before the error, in the order they were
// found; with --json they fold into the same envelope as the error, because
// stderr must not mix the text of a warning with the JSON object of an
// error (docs/spec/contrato-json.md#los-errores-en-json).
func fail(s Streams, asJSON bool, err error, warnings []Warning) int {
	// errors.As and not a type assertion, because an error of the
	// specification can travel inside a richer one: the ambiguous reference
	// of docs/spec/referencias.md carries its candidates alongside it.
	var e *model.Error
	if !errors.As(err, &e) {
		// Anything that is not a case of the specification is the program
		// failing, which is exit code 1 of docs/spec/codigos-de-salida.md.
		e = &model.Error{ExitCode: 1, Code: "internal", Message: err.Error()}
	}
	if asJSON {
		writeErrorEnvelope(s, e, warnings)
		return e.ExitCode
	}
	for _, w := range warnings {
		fmt.Fprint(s.Stderr, prefixed("warning: ", w.Message))
		for _, hint := range w.Hints {
			fmt.Fprint(s.Stderr, prefixed("hint: ", hint))
		}
	}
	fmt.Fprint(s.Stderr, prefixed("error: ", e.Message))
	if line := validValuesLine(e); line != "" {
		fmt.Fprintln(s.Stderr, line)
	}
	for _, line := range e.Detail {
		fmt.Fprintln(s.Stderr, line)
	}
	for _, note := range e.Notes {
		fmt.Fprint(s.Stderr, prefixed("note: ", note))
	}
	for _, hint := range e.Hints {
		fmt.Fprint(s.Stderr, prefixed("hint: ", hint))
	}
	return e.ExitCode
}

// prefixed writes one message under its prefix: the first line behind the
// prefix and every further line of the same message aligned under it, which
// is how docs/spec/ prints a note or a hint that spans more than one line.
func prefixed(prefix, message string) string {
	indent := make([]byte, len(prefix))
	for i := range indent {
		indent[i] = ' '
	}
	out := prefix
	for i, line := range splitLines(message) {
		if i > 0 {
			out += string(indent)
		}
		out += line + "\n"
	}
	return out
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// Terminal answers whether f is a terminal, which cmd/biso/main.go asks once
// for each of the two output streams.
func Terminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// validValuesLine is the second line that two families of error print under
// their message, aligned with it: the vocabulary a closed field configures
// (docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos)
// and the extension keys a board declares
// (docs/spec/modelo-de-datos/campos-externos.md).
//
// It is built here and not carried inside the error because the list it
// names already travels in `valid`, and the JSON envelope does not repeat
// the sentence that wraps it: whoever reads JSON reads the list.
func validValuesLine(e *model.Error) string {
	if len(e.Valid) == 0 {
		return ""
	}
	var what string
	switch e.Code {
	case "unknown_status":
		what = "valid statuses on this board"
	case "unknown_type":
		what = "valid types on this board"
	case "unknown_priority":
		what = "valid priorities on this board"
	case "unknown_extension_key":
		what = "declared keys on this board"
	case "unknown_section":
		// The two closed domains of the reading commands are not the
		// board's vocabulary: they are the same everywhere, so the line
		// does not say "on this board".
		what = "valid sections"
	case "unknown_sort_field":
		what = "valid sort fields"
	default:
		return ""
	}
	// Seven spaces, which is the width of "error: ", so the line sits under
	// the message exactly as docs/spec/vocabularios.md prints it.
	return "       " + what + ": " + strings.Join(e.Valid, ", ")
}
