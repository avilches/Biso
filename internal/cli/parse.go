package cli

import (
	"fmt"
	"sort"
	"strings"
)

// Parse turns an argv into the invocation it describes: the command, its
// positional arguments, and every value of every flag, already read through
// the three forms of docs/spec/valores-de-entrada.md and already classified
// into the fixed steps of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura.
//
// argv comes without the program name, the way cmd/biso/main.go hands over
// os.Args[1:]. A single generic function walks every command's table, which is
// what keeps the cross-cutting rules written once instead of once per command,
// as section 6 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md asks.
//
// Every error it returns is a *model.Error, complete from birth, and Parse
// stops at the first one: the order in which the specification's cases are
// checked is the order of the command line, so a caller that writes two wrong
// things hears about the first.
//
// An error comes back together with the analysis as far as it got, which is
// never a usable invocation and does carry the warnings already found. A
// warning is never suppressed
// (docs/spec/salida-y-terminal.md#notas-y-avisos), and one that the call
// had already earned before writing something wrong is no exception: it
// travels with the error instead of disappearing with it.
func Parse(argv []string, commands []CommandSpec, env Env) (*Parsed, error) {
	st := &parser{
		p:        &Parsed{emptied: map[string]bool{}},
		commands: commands,
		env:      env.withDefaults(),
		dupes:    map[string]int{},
		times:    map[string]int{},
	}
	if err := st.run(argv); err != nil {
		return st.p, err
	}
	return st.p, nil
}

// Action is what docs/spec/cmd/flags-globales.md separates from a mode:
// --version and --help are not something that stays on while it is written,
// they are an action that prints something and ends the program with code 0 as
// soon as it is read.
type Action int

const (
	ActionNone Action = iota
	ActionHelp
	ActionVersion
)

// Parsed is one invocation, already analyzed.
type Parsed struct {
	// Command is the name of the command, empty when the call named none.
	Command string
	// Positionals are the arguments that are not flags, in the order they
	// were written, exactly as they were typed.
	Positionals []string
	// Texts are the positional arguments the command reads as long text
	// values, from CommandSpec.TextPositionalsFrom on: each one as it was
	// typed and as it resolved through the three forms of
	// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo.
	Texts []PositionalText
	// Action is set when the call ended early because it asked for the
	// help or for the version.
	Action Action
	// Warnings are the ones the analysis itself produced, in the order they
	// were found. They never change the exit code, which stays 0
	// (docs/spec/salida-y-terminal.md#notas-y-avisos).
	Warnings []Warning

	occs    []*occurrence
	emptied map[string]bool
	cmd     *CommandSpec
}

// PositionalText is one positional argument of kind TextValue: what was
// typed, and the value it resolved to. The two travel together because the
// rule of the positional that looks like an identifier judges the first and
// the write stores the second
// (docs/spec/cmd/verbos-del-ciclo.md#el-posicional-que-parece-un-identificador).
type PositionalText struct {
	Typed string
	Value string
}

// Warning is one line of docs/spec/salida-y-terminal.md#notas-y-avisos: its
// stable code, the text that follows "warning: " on stderr, the hint lines
// that follow it when the case has one, and the fields the JSON object of
// data.warnings carries.
type Warning struct {
	Code    string
	Message string
	Hints   []string
	Fields  map[string]any
	// Detail are the display lines printed under the message, each one
	// carrying its own indentation, such as the unchecked criteria of
	// docs/spec/cmd/verbos-del-ciclo.md#biso-finish. They never reach the
	// JSON object, which carries Fields instead.
	Detail []string
}

// Change is one value of one flag that writes a field, with the step of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura it
// belongs to. Key is filled only for a flag whose value is a pair, which is
// --set-comment-date.
type Change struct {
	Flag     *FlagSpec
	Category Category
	Key      string
	Value    string
}

// occurrence is one value of one flag as it was found. A dropped occurrence is
// one the rules of the specification take back out: a repeated value of a list
// kept once, or an empty value that adds nothing.
type occurrence struct {
	flag    *FlagSpec
	key     string
	value   string
	order   int
	dropped bool
}

type parser struct {
	p        *Parsed
	commands []CommandSpec
	env      Env

	cmd          *CommandSpec
	stdinTakenBy *FlagSpec
	order        int

	// dupes remembers which warning already reports a repeated value or a
	// repeated key, and times how many appearances that warning is counting,
	// so that the third appearance corrects that one warning instead of
	// adding a second one that says "twice" again.
	dupes map[string]int
	times map[string]int

	// posClosed holds the rule that the positional arguments of a call all
	// come in front of the flags of their command: the first flag written
	// after the command name closes them, and what is left over after that
	// is precisely how a forgotten value is found
	// (docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion).
	posClosed bool
}

func (st *parser) run(argv []string) error {
	// The encoding is judged before anything else is read, over the whole
	// argv, so that an invalid byte can never reach a message or the given
	// key of an error envelope
	// (docs/spec/salida-y-terminal.md#codificación-y-texto).
	for i, arg := range argv {
		if at := invalidUTF8At(arg); at >= 0 {
			return errInvalidEncodingArgument(i+1, at, arg)
		}
	}
	// And the rule of one standard input per invocation is answered next,
	// over the whole command line and before a single value is read. It
	// cannot wait until the second "-" is resolved, because by then the
	// first one has already drained the stream and earned whatever
	// warnings its value deserved, on a call that was never going to run
	// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).
	if err := st.refuseTwoStdin(argv); err != nil {
		return err
	}
	afterDashDash := false
	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		switch {
		case afterDashDash:
			// Everything behind "--" is a positional argument, whatever
			// it looks like, which is the second mechanism of
			// docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion.
			st.p.Positionals = append(st.p.Positionals, tok)
		case tok == "--":
			afterDashDash = true
		case strings.HasPrefix(tok, "--") && len(tok) > 2:
			name, inline, hasInline := strings.Cut(tok[2:], "=")
			f := lookupLong(st.cmd, name)
			if f == nil {
				return errUnknownFlag("--" + name)
			}
			used, err := st.flag(f, inline, hasInline, argv, i)
			if err != nil {
				return err
			}
			if st.p.Action != ActionNone {
				return nil
			}
			i += used
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			name, inline, hasInline := strings.Cut(tok[1:], "=")
			f := lookupShort(st.cmd, name)
			if f == nil {
				return errUnknownFlag("-" + name)
			}
			used, err := st.flag(f, inline, hasInline, argv, i)
			if err != nil {
				return err
			}
			if st.p.Action != ActionNone {
				return nil
			}
			i += used
		case st.cmd == nil:
			// The first argument that is not a flag names the command.
			// Only a global flag may come before it: a command's own flag
			// written there is unknown, because there is no table yet to
			// look it up in.
			if err := st.command(tok); err != nil {
				return err
			}
		default:
			if err := st.positional(tok); err != nil {
				return err
			}
		}
	}
	return st.finish()
}

// refuseTwoStdin walks the command line looking only for the flags and the
// positional arguments that ask for standard input with "-", and answers the
// error of two of them before anything has been read. It resolves no value,
// records no occurrence and emits no warning: everything it finds, the main
// walk finds again.
//
// Anything it cannot resolve, such as a flag or a command that does not
// exist, ends the scan without an answer, because that is an error the main
// walk gives at the same token and it has to be the one that comes out.
func (st *parser) refuseTwoStdin(argv []string) error {
	var cmd *CommandSpec
	var first *FlagSpec
	positionals, posClosed, afterDashDash := 0, false, false

	asks := func(f *FlagSpec) error {
		if first == nil {
			first = f
			return nil
		}
		if first == f {
			// One flag repeated is not two flags fighting over the
			// stream, exactly as it is not when the value is read.
			return errStdinTwiceInOneFlag(f)
		}
		return errTwoStdin(first, f)
	}
	positional := func(tok string) error {
		defer func() { positionals++ }()
		if cmd == nil || cmd.TextPositionalsFrom <= 0 ||
			positionals < cmd.TextPositionalsFrom || tok != "-" {
			return nil
		}
		return asks(&textPositional)
	}

	for i := 0; i < len(argv); i++ {
		tok := argv[i]
		switch {
		case afterDashDash:
			if err := positional(tok); err != nil {
				return err
			}
		case tok == "--":
			afterDashDash = true
		case strings.HasPrefix(tok, "-") && len(tok) > 1:
			var f *FlagSpec
			if strings.HasPrefix(tok, "--") && len(tok) > 2 {
				name, _, _ := strings.Cut(tok[2:], "=")
				f = lookupLong(cmd, name)
			} else {
				name, _, _ := strings.Cut(tok[1:], "=")
				f = lookupShort(cmd, name)
			}
			if f == nil {
				return nil
			}
			if cmd != nil {
				posClosed = true
			}
			if f.Name == "help" || f.Name == "version" {
				// The call ends as soon as one of the two is read, so
				// nothing behind it is part of it.
				return nil
			}
			if f.Value == NoValue {
				continue
			}
			raw, hasInline := "", false
			if _, inline, ok := strings.Cut(tok, "="); ok {
				raw, hasInline = inline, true
			}
			if !hasInline {
				if i+1 >= len(argv) {
					return nil
				}
				raw = argv[i+1]
				i++
			}
			if f.Value == TextValue && raw == "-" {
				if err := asks(f); err != nil {
					return err
				}
			}
		case cmd == nil:
			cmd = lookupCommand(st.commands, tok)
			if cmd == nil {
				return nil
			}
		case posClosed:
			return nil
		default:
			if err := positional(tok); err != nil {
				return err
			}
		}
	}
	return nil
}

// lookupCommand answers the specification of a command by its name, or nil.
func lookupCommand(commands []CommandSpec, name string) *CommandSpec {
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	return nil
}

func (st *parser) command(name string) error {
	if cmd := lookupCommand(st.commands, name); cmd != nil {
		st.cmd = cmd
		st.p.Command = name
		st.p.cmd = cmd
		return nil
	}
	if err := errNoDeleteCommand(name); err != nil {
		// The absence of a delete command is specified, so the three
		// names somebody would reach for answer what to do instead of
		// dumping the list of commands
		// (docs/spec/cmd/archive.md#biso-delete-no-existe-y-su-ausencia-está-especificada).
		return err
	}
	return errUnknownCommand(name, commandNames(st.commands))
}

func (st *parser) positional(tok string) error {
	if st.posClosed {
		return errUnexpectedArgument(tok)
	}
	st.p.Positionals = append(st.p.Positionals, tok)
	return nil
}

// flag records one appearance of a flag and answers how many further arguments
// of argv it consumed, which is one when the value came as the next argument
// and none when it came attached with "=" or when the flag takes no value.
func (st *parser) flag(f *FlagSpec, inline string, hasInline bool, argv []string, i int) (int, error) {
	if st.cmd != nil {
		// A flag of the command closes the block of positional arguments.
		// A global flag in front of the command name does not, because
		// nothing can be positional there: the first argument that is not
		// a flag is the command.
		st.posClosed = true
	}
	if err := st.conflicts(f); err != nil {
		return 0, err
	}

	if f.Value == NoValue {
		if hasInline {
			return 0, errTakesNoValue(f, inline)
		}
		switch f.Name {
		case "help":
			st.p.Action = ActionHelp
			return 0, nil
		case "version":
			st.p.Action = ActionVersion
			return 0, nil
		}
		st.record(f, "", "")
		return 0, nil
	}

	raw := inline
	used := 0
	if !hasInline {
		if i+1 >= len(argv) {
			return 0, errMissingValue(f)
		}
		// A value that starts with a hyphen behind a flag that demands one
		// is taken as it is, with no heuristics.
		raw = argv[i+1]
		used = 1
	}
	if err := st.value(f, raw); err != nil {
		return 0, err
	}
	return used, nil
}

// value reads one value of a flag and records everything it yields: one entry
// for a scalar, and as many as the commas of a list flag produce.
func (st *parser) value(f *FlagSpec, raw string) error {
	resolved, err := st.readValue(f, raw)
	if err != nil {
		return err
	}
	if at := invalidUTF8At(resolved); at >= 0 {
		// argv was already judged whole, so what can still arrive
		// undecodable here is a file or the standard input.
		return errInvalidEncoding(f, raw, at)
	}
	resolved = normalizeNewlines(resolved)

	if f.Value == TextValue {
		st.warnLiteralNewline(f, resolved)
	}

	parts := []string{resolved}
	if f.Comma {
		parts = splitList(resolved)
	}
	for _, part := range parts {
		if err := st.element(f, part); err != nil {
			return err
		}
	}
	return nil
}

// element applies to one value everything the specification says about a
// value: the empty one, the closed alphabet of a token field, the one line of
// a string field, a domain of fixed values, and the rules of repetition. A
// value of the form "<left>=<right>" is cut first and then the same rules
// apply to each half, the alphabet to the key and the rest to the value, so
// that none of them is unreachable for the flags that take a pair.
func (st *parser) element(f *FlagSpec, v string) error {
	key, value := "", v
	if f.Pair != NotAPair {
		var err error
		if key, value, err = splitPair(f, v); err != nil {
			return err
		}
		if !f.Alphabet.allowed(key) {
			return errMalformedToken(f, key)
		}
	}
	if isEmpty(value) {
		if key != "" {
			// A pair with nothing on its right is the same kind of
			// mistake as any other value the specification does not
			// document as empty.
			return errEmptyValue(f, key, v)
		}
		return st.empty(f, value)
	}
	if key == "" && !f.Alphabet.allowed(value) {
		return errMalformedToken(f, value)
	}
	if key == "" {
		if err := checkLabelSyntax(f, value); err != nil {
			return err
		}
	}
	if f.SingleLine && strings.ContainsAny(value, "\n\r") {
		return errMalformedString(f, value)
	}
	if len(f.Domain) > 0 && !contains(f.Domain, value) {
		return errOutsideDomain(f, value)
	}
	if key != "" {
		return st.pair(f, key, value)
	}
	return st.repetition(f, value)
}

// empty resolves the table of docs/spec/valores-de-entrada.md#el-valor-vacío.
// A value of nothing but spaces means something different in each family, and
// only one of the three is an error.
func (st *parser) empty(f *FlagSpec, v string) error {
	switch {
	case f.Category == Add:
		st.warn(Warning{
			Code:    "empty_append",
			Message: f.long() + ": empty value, nothing was added",
			Fields:  map[string]any{"flag": f.long()},
		})
		return nil
	case f.Category == Replace:
		// Replacing with nothing is emptying, and that is explicit.
		st.p.emptied[f.Name] = true
		return nil
	case f.ClosedVocabulary:
		// The board's vocabulary is what judges this value, so it travels
		// down and internal/match answers with the code of an unknown
		// value. What travels is the empty string and not the spaces that
		// were typed: turning a value of nothing but spaces into the empty
		// value is this layer's job, and a board that ever printed
		// `unknown status: "   "` would be reporting a caller that skipped
		// that conversion (docs/spec/vocabularios.md#el-algoritmo-de-coincidencia).
		return st.repetition(f, "")
	case f.Category == Scalar:
		return errEmptyScalar(f)
	}
	// Anywhere else, the empty value is nothing the specification
	// documents, and the board's vocabulary has nothing to say about it, so
	// it is a malformed command line and not an unknown value.
	return errEmptyValue(f, "", v)
}

// splitPair cuts a value of the form "<left>=<right>" at the last "=", which
// is what docs/spec/familias-de-flags.md#comentarios says about the selector
// of --set-comment-date: its left half is free text that may carry another.
func splitPair(f *FlagSpec, v string) (key, value string, err error) {
	idx := strings.LastIndex(v, "=")
	if idx <= 0 {
		return "", "", errMalformedPair(f, v)
	}
	return v[:idx], v[idx+1:], nil
}

// pair records one key of a flag whose value is a pair, and answers what the
// same key twice in one call means: with two different values it is the
// error of a repeated scalar, and with the same value it applies once and
// says nothing (docs/spec/familias-de-flags.md#comentarios).
func (st *parser) pair(f *FlagSpec, key, value string) error {
	for _, o := range st.p.occs {
		if o.flag != f || o.dropped || o.key != key {
			continue
		}
		if o.value != value {
			return errDuplicateKey(f, key, o.value, value)
		}
		return nil
	}
	st.record(f, key, value)
	return nil
}

// repetition holds the two halves of
// docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas: a
// repeatable flag accumulates and keeps a repeated value once, and one that is
// not repeatable, given twice with different values, is a usage error.
func (st *parser) repetition(f *FlagSpec, v string) error {
	for _, o := range st.p.occs {
		if o.flag != f || o.dropped {
			continue
		}
		if f.Repeatable {
			if o.value == v {
				st.warnAgain("value\x00"+f.Name+"\x00"+v, func(times int) Warning {
					return Warning{
						Code:    "duplicate_flag_value",
						Message: fmt.Sprintf("%s: %q given %s, kept once", f.long(), v, timesWritten(times)),
						Fields:  map[string]any{"flag": f.long(), "value": v},
					}
				})
				return nil
			}
			continue
		}
		if o.value != v {
			return errDuplicateScalar(f, o.value, v)
		}
		// The same value twice says the same thing twice: it is kept once
		// and there is nothing to report.
		return nil
	}
	st.record(f, "", v)
	return nil
}

// warnLiteralNewline is the rule of
// docs/spec/salida-y-terminal.md#codificación-y-texto: a text value carrying
// the two characters "\" and "n" and no real newline is almost always an
// accident, so it warns and is stored all the same.
func (st *parser) warnLiteralNewline(f *FlagSpec, v string) {
	if !strings.Contains(v, `\n`) || strings.Contains(v, "\n") {
		return
	}
	st.warn(Warning{
		Code:    "literal_newline",
		Message: f.named() + ` contains a literal \n and no real newline; it will be stored as text`,
		Hints: []string{fmt.Sprintf(
			"use a real newline, or %s @file.md, or %s - to read from stdin",
			f.shortest(), f.shortest())},
		Fields: map[string]any{"flag": f.long()},
	})
}

// conflicts rejects a pair of flags that cannot share a call, such as any two
// of --json, --quiet and --print, which ask for three different shapes of the
// same output. The message names them in the order of the specification table
// and not in the order they were typed, so that --json --quiet and --quiet
// --json fail with the same text.
func (st *parser) conflicts(f *FlagSpec) error {
	for _, name := range f.Conflicts {
		other := st.given(name)
		if other == nil {
			continue
		}
		first, second := other, f
		if tableOrder(st.cmd, second) < tableOrder(st.cmd, first) {
			first, second = second, first
		}
		return errIncompatible(first, second)
	}
	return nil
}

func (st *parser) record(f *FlagSpec, key, value string) {
	st.order++
	st.p.occs = append(st.p.occs, &occurrence{flag: f, key: key, value: value, order: st.order})
}

func (st *parser) warn(w Warning) {
	st.p.Warnings = append(st.p.Warnings, w)
}

// warnAgain reports something that repeats. The second appearance emits the
// warning; the third and every one after it correct that same warning in place
// instead of adding another one that would say "twice" again. id is what two
// appearances of the same thing share, and build writes the text for a given
// number of appearances.
func (st *parser) warnAgain(id string, build func(times int) Warning) {
	if at, ok := st.dupes[id]; ok {
		st.times[id]++
		st.p.Warnings[at] = build(st.times[id])
		return
	}
	st.dupes[id] = len(st.p.Warnings)
	st.times[id] = 2
	st.warn(build(2))
}

// given writes how many times something was written, the way the warnings of
// docs/spec/salida-y-terminal.md#notas-y-avisos say it.
func timesWritten(times int) string {
	if times == 2 {
		return "twice"
	}
	return fmt.Sprintf("%d times", times)
}

// given returns the specification of a flag the call has already written, or
// nil when it has not.
func (st *parser) given(name string) *FlagSpec {
	for _, o := range st.p.occs {
		if o.flag.Name == name {
			return o.flag
		}
	}
	return nil
}

// finish runs the checks that can only be answered once the whole command line
// has been read.
func (st *parser) finish() error {
	for _, o := range st.p.occs {
		for _, name := range o.flag.Requires {
			if st.given(name) == nil {
				return errRequires(o.flag, name)
			}
		}
	}
	if st.cmd == nil {
		return nil
	}
	if err := st.textPositionals(); err != nil {
		return err
	}
	// The two rules of docs/spec/cmd/flags-globales.md: neither flag is ever
	// ignored in silence where it has nothing to do, and each one is
	// defined over a different thing, so each has its own list.
	readOnly := st.cmd.ReadOnly
	for _, name := range st.cmd.WriteFlags {
		if st.given(name) != nil {
			// biso doctor --fix is the case this exists for: the same
			// command is read-only or not depending on one flag.
			readOnly = false
			break
		}
	}
	if f := st.given("dry-run"); f != nil && readOnly {
		return errReadOnlyFlag(f, "--dry-run does not apply to a read-only command")
	}
	if f := st.given("print"); f != nil && (readOnly || st.cmd.AffectsNoTask) {
		return errReadOnlyFlag(f, "--print does not apply to a command that affects no task")
	}
	return nil
}

// textPositional is the entry the value rules are applied through for a
// positional of kind TextValue. It is one single value for the whole
// program, so that two of them asking for standard input in the same call
// collide exactly as two flags would.
var textPositional = FlagSpec{Name: "text", Label: "text", Value: TextValue, Field: "text"}

// textPositionals reads the positional arguments a command declares as long
// texts through the three forms of
// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo, which
// is what makes `biso note MYP-11 @findings.md` mean the same as
// `biso set MYP-11 --append-note @findings.md`.
func (st *parser) textPositionals() error {
	from := st.cmd.TextPositionalsFrom
	if from <= 0 {
		return nil
	}
	for i := from; i < len(st.p.Positionals); i++ {
		typed := st.p.Positionals[i]
		value, err := st.readValue(&textPositional, typed)
		if err != nil {
			return err
		}
		if at := invalidUTF8At(value); at >= 0 {
			return errInvalidEncoding(&textPositional, typed, at)
		}
		value = normalizeNewlines(value)
		st.warnLiteralNewline(&textPositional, value)
		st.p.Texts = append(st.p.Texts, PositionalText{Typed: typed, Value: value})
	}
	if len(st.p.Positionals) > from {
		st.p.Positionals = st.p.Positionals[:from]
	}
	return nil
}

// Has answers whether the call wrote this flag at all, which is the question a
// switch answers with.
func (p *Parsed) Has(name string) bool {
	for _, o := range p.occs {
		if o.flag.Name == name && !o.dropped {
			return true
		}
	}
	return p.emptied[name]
}

// Value is the value of a flag that keeps one, and false when the call did not
// write it.
func (p *Parsed) Value(name string) (string, bool) {
	for i := len(p.occs) - 1; i >= 0; i-- {
		if p.occs[i].flag.Name == name && !p.occs[i].dropped {
			return p.occs[i].value, true
		}
	}
	return "", false
}

// Values are every value of a flag, in the order of the command line, whether
// they arrived by repeating the flag, by separating them with commas, or by
// mixing the two.
func (p *Parsed) Values(name string) []string {
	var out []string
	for _, o := range p.occs {
		if o.flag.Name == name && !o.dropped {
			out = append(out, o.value)
		}
	}
	return out
}

// Emptied says a flag that replaces a whole list was given an empty value,
// which leaves the field empty just like its --clear-* sibling.
func (p *Parsed) Emptied(name string) bool { return p.emptied[name] }

// EmptiedFlags are all of those, sorted, so that a caller that walks them
// never depends on the order a map happens to have.
func (p *Parsed) EmptiedFlags() []string {
	names := make([]string, 0, len(p.emptied))
	for name := range p.emptied {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Changes are the values that write a field, sorted into the fixed steps of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura and, inside
// a step, in the order of the command line. The caller applies them in the
// order it gets them, and that order does not depend on how the call was
// written.
func (p *Parsed) Changes() []Change {
	out := make([]Change, 0, len(p.occs))
	kept := make([]*occurrence, 0, len(p.occs))
	for _, o := range p.occs {
		if o.dropped || o.flag.Category == NotAChange {
			continue
		}
		kept = append(kept, o)
	}
	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].flag.Category != kept[j].flag.Category {
			return kept[i].flag.Category < kept[j].flag.Category
		}
		return kept[i].order < kept[j].order
	})
	for _, o := range kept {
		out = append(out, Change{
			Flag:     o.flag,
			Category: o.flag.Category,
			Key:      o.key,
			Value:    o.value,
		})
	}
	return out
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}
