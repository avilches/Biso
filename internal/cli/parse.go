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
func Parse(argv []string, commands []CommandSpec, env Env) (*Parsed, error) {
	st := &parser{
		p:        &Parsed{emptied: map[string]bool{}},
		commands: commands,
		env:      env.withDefaults(),
	}
	if err := st.run(argv); err != nil {
		return nil, err
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
	// were written.
	Positionals []string
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

// Warning is one line of docs/spec/salida-y-terminal.md#notas-y-avisos: its
// stable code, the text that follows "warning: " on stderr, the hint lines
// that follow it when the case has one, and the fields the JSON object of
// data.warnings carries.
type Warning struct {
	Code    string
	Message string
	Hints   []string
	Fields  map[string]any
}

// Change is one value of one flag that writes a field, with the step of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura it
// belongs to. Key is filled only for a flag whose value is a pair, such as
// --ext and --set-comment-date.
type Change struct {
	Flag     *FlagSpec
	Category Category
	Key      string
	Value    string
}

// occurrence is one value of one flag as it was found. A dropped occurrence is
// one the rules of the specification take back out: a repeated value of a list
// kept once, an ext key given twice and kept at its last value, or an empty
// value that adds nothing.
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

	// posStarted and posClosed hold the rule that the positional arguments
	// of a call are one single block: they may come before the flags or
	// after them, but not on both sides, because what is left over after a
	// flag that ate its value is precisely how a forgotten value is found
	// (docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion).
	posStarted bool
	posClosed  bool
}

func (st *parser) run(argv []string) error {
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

func (st *parser) command(name string) error {
	for i := range st.commands {
		if st.commands[i].Name == name {
			st.cmd = &st.commands[i]
			st.p.Command = name
			st.p.cmd = st.cmd
			return nil
		}
	}
	return errUnknownCommand(name, commandNames(st.commands))
}

func (st *parser) positional(tok string) error {
	if st.posClosed {
		return errUnexpectedArgument(tok)
	}
	if at := invalidUTF8At(tok); at >= 0 {
		return errInvalidEncoding(nil, at)
	}
	st.posStarted = true
	st.p.Positionals = append(st.p.Positionals, tok)
	return nil
}

// flag records one appearance of a flag and answers how many further arguments
// of argv it consumed, which is one when the value came as the next argument
// and none when it came attached with "=" or when the flag takes no value.
func (st *parser) flag(f *FlagSpec, inline string, hasInline bool, argv []string, i int) (int, error) {
	if st.posStarted {
		// A flag after the block of positional arguments closes it.
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
		return errInvalidEncoding(f, at)
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
// a string field, a domain of fixed values, and the rules of repetition.
func (st *parser) element(f *FlagSpec, v string) error {
	if f.Pair != NotAPair {
		return st.pair(f, v)
	}
	if isEmpty(v) {
		return st.empty(f, v)
	}
	if !f.Alphabet.allowed(v) {
		return errMalformedToken(f, v)
	}
	if f.SingleLine && strings.ContainsAny(v, "\n\r") {
		return errMalformedString(f, v)
	}
	if len(f.Domain) > 0 && !contains(f.Domain, v) {
		return errOutsideDomain(f, v)
	}
	if err := st.repetition(f, v); err != nil {
		return err
	}
	return nil
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
		// down untouched and internal/match answers with the code of an
		// unknown value.
		return st.repetition(f, v)
	}
	return errEmptyScalar(f)
}

// pair splits a value of the form "<left>=<right>". Where it cuts is the
// difference docs/spec/familias-de-flags.md#comentarios spells out, and the
// table says which of the two rules each flag follows.
func (st *parser) pair(f *FlagSpec, v string) error {
	var key, value string
	var ok bool
	if f.Pair == PairAtLastEquals {
		if idx := strings.LastIndex(v, "="); idx >= 0 {
			key, value, ok = v[:idx], v[idx+1:], true
		}
	} else {
		key, value, ok = strings.Cut(v, "=")
	}
	if !ok || key == "" {
		return errMalformedPair(f, v)
	}
	if !f.Alphabet.allowed(key) {
		return errMalformedToken(f, key)
	}
	// The same key twice in one call is not an error: the last value of the
	// command line wins, with a warning, because a key of a map behaves like
	// one more token and not like a scalar of the whole task.
	for _, o := range st.p.occs {
		if o.flag == f && !o.dropped && o.key == key {
			o.dropped = true
			st.warn(Warning{
				Code:    "duplicate_ext_key",
				Message: fmt.Sprintf("%s: key %q given twice, kept last value", f.long(), key),
				Fields:  map[string]any{"flag": f.long(), "key": key},
			})
		}
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
				st.warn(Warning{
					Code:    "duplicate_flag_value",
					Message: fmt.Sprintf("%s: %q given twice, kept once", f.long(), v),
					Fields:  map[string]any{"flag": f.long(), "value": v},
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
		Message: f.long() + ` contains a literal \n and no real newline; it will be stored as text`,
		Hints: []string{fmt.Sprintf(
			"use a real newline, or %s @file.md, or %s - to read from stdin",
			f.shortest(), f.shortest())},
		Fields: map[string]any{"flag": f.long()},
	})
}

// conflicts rejects a pair of flags that cannot share a call, such as any two
// of --json, --quiet and --print, which ask for three different shapes of the
// same output.
func (st *parser) conflicts(f *FlagSpec) error {
	for _, name := range f.Conflicts {
		if other := st.given(name); other != nil {
			return errIncompatible(other, f)
		}
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
	// The two rules of docs/spec/cmd/flags-globales.md: neither flag is ever
	// ignored in silence where it has nothing to do, and each one is
	// defined over a different thing, so each has its own list.
	if st.given("dry-run") != nil && st.cmd.ReadOnly {
		return errReadOnlyFlag("--dry-run does not apply to a read-only command")
	}
	if st.given("print") != nil && (st.cmd.ReadOnly || st.cmd.AffectsNoTask) {
		return errReadOnlyFlag("--print does not apply to a command that affects no task")
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
