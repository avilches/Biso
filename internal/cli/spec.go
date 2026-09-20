// Package cli turns an argv into the typed parameters of a command, and a
// result back into text or into the JSON envelope of
// docs/spec/contrato-json.md.
//
// This file holds the specification table that section 6 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md asks
// for: one entry per flag, saying whether it takes a value, whether that value
// admits the three forms of docs/spec/valores-de-entrada.md, and whether it is
// a list or a scalar. A single generic function (Parse, in parse.go) walks the
// table, so every cross-cutting rule of the specification is written once
// instead of once per command.
package cli

import (
	"math"
	"sort"
	"strings"
)

// ValueKind says whether a flag takes a value, and how that value is read.
type ValueKind int

const (
	// NoValue is a switch: writing it turns something on, and it takes no
	// value at all. docs/spec/cmd/flags-globales.md calls these "the flags
	// that take no value".
	NoValue ValueKind = iota
	// PlainValue is a value taken exactly as it is typed. It is what the
	// person fields use, which never interpret a leading "@"
	// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).
	PlainValue
	// TextValue is a value in the three forms of
	// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo:
	// the literal text, "@path" for the contents of a file, and "-" for
	// standard input, with "@@" as the only escape there is.
	TextValue
)

// Category is the step of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura that a
// flag belongs to. The parser does not apply changes in the order they arrive
// in argv: it classifies them into these fixed steps, and within a step keeps
// the order of the command line. The numeric order of these constants is the
// order of that section, so sorting by it is the specified order.
type Category int

const (
	// NotAChange is every flag that writes no field: the global flags, and
	// each command's own switches. They never appear in Changes.
	NotAChange  Category = iota
	Clear                // step 1: every --clear-*, --clear-ext included
	Replace              // step 2: every --replace-*
	Remove               // step 3: every --rm-*, --rm-comment and --rm-ext included
	Add                  // step 4: the additions, --add-* and --append-*
	ExtKey               // step 5: the map fields, --ext
	Scalar               // step 6: the scalar fields
	CheckAC              // step 7: --check-ac and --uncheck-ac
	CommentDate          // step 8: --set-comment-date
	AddComment           // step 9: --comment, the one that adds
)

// Alphabet is the closed character set of a token field, from
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token. Fields
// that are free text (references, documentation, touched files) use AnyText,
// because a URL or a path cannot have its alphabet closed without leaving
// legitimate values out.
type Alphabet int

const (
	AnyText        Alphabet = iota
	TokenAlphabet           // labels and assignees: letters, digits, and - _ . : @
	ExtKeyAlphabet          // an ext key: letters, digits, and - _ .
)

// symbols is the list of symbols this alphabet admits, written as the hint of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token prints it.
func (a Alphabet) symbols() string {
	switch a {
	case TokenAlphabet:
		return "- _ . : @"
	case ExtKeyAlphabet:
		return "- _ ."
	}
	return ""
}

// PairKind says whether a value is a pair of the form "<left>=<right>", and
// which "=" separates the two halves. The two rules of the specification are
// different on purpose and
// docs/spec/familias-de-flags.md#comentarios explains why: --ext can cut at
// the first "=" because an ext key cannot contain one, and
// --set-comment-date has to cut at the last one because its left half is free
// text that may well carry another.
type PairKind int

const (
	NotAPair PairKind = iota
	PairAtFirstEquals
	PairAtLastEquals
)

// FlagSpec is one row of the specification table: everything the generic
// parser needs to know about a flag. A command's table lists only its own
// flags; the global ones of docs/spec/cmd/flags-globales.md live in
// GlobalFlags and no command may redefine them.
type FlagSpec struct {
	// Name is the long name without the leading dashes ("add-labels"), and
	// Short the one-letter name without its dash ("l"), empty when the flag
	// has none. Every message names a flag by its long form, whichever of
	// the two was typed.
	Name  string
	Short string

	// Label is how a message names this entry when it is not a flag at
	// all: the positional text of `biso note` goes through the same value
	// rules as a flag of kind TextValue, and a message about it has to
	// read "text: file not found" and never "--text: file not found".
	Label string

	Value ValueKind

	// Repeatable says the flag accumulates a value on each appearance
	// instead of keeping the last one
	// (docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas).
	// A flag that is not repeatable, given twice with different values, is
	// a usage error.
	Repeatable bool

	// Comma says the value is also a list separated by commas, where "\," is
	// a literal comma. The prose fields and the acceptance criteria never
	// split on commas, because their text may contain one.
	Comma bool

	Category Category

	// Pair, with PairSyntax, describes a value of the form "<left>=<right>".
	// PairSyntax is the form the error message shows, such as
	// "<key>=<value>".
	Pair       PairKind
	PairSyntax string

	// LastKeyWins says what the same key twice in one call means, and the
	// two flags that take a pair answer differently on purpose. --ext keeps
	// the last value with a warning, because a key of a map behaves like one
	// more token
	// (docs/spec/familias-de-flags.md#campos-externos).
	// --set-comment-date does not: the same key with two different instants
	// is the error of a repeated scalar, and with the same instant it
	// applies once and says nothing
	// (docs/spec/familias-de-flags.md#comentarios).
	LastKeyWins bool

	// Alphabet and Noun close the character set of a token field. Noun is
	// the singular word its message uses ("label", "assignee",
	// "extension key"); the error code is "malformed_" plus that noun with
	// its spaces turned into underscores.
	Alphabet Alphabet
	Noun     string

	// SingleLine marks a field of type string, one line of text, which
	// admits no carriage return and no newline
	// (docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string).
	SingleLine bool

	// Field is the "field" key of the JSON error envelope of
	// docs/spec/contrato-json.md#los-errores-en-json. It defaults to Name.
	Field string

	// ClosedVocabulary marks a scalar whose value the board's vocabulary
	// judges, so an empty value is not rejected here: it travels down and
	// internal/match answers with the code of an unknown value, as
	// docs/spec/valores-de-entrada.md#el-valor-vacío requires.
	ClosedVocabulary bool

	// ClearFlag is the long name of the flag that empties this field, quoted
	// by the hint of an empty scalar. Empty when the field cannot be
	// emptied, such as the title or the status.
	ClearFlag string

	// Domain closes the set of values a flag accepts, such as the three of
	// --color, and DomainCode is the error code that rejecting one produces.
	Domain     []string
	DomainCode string

	// Conflicts and Requires are the last two columns of a command's table
	// of parameters: the long names this flag cannot share a call with, and
	// the ones it cannot appear without.
	Conflicts []string
	Requires  []string
}

// field is the name this flag answers to in the JSON error envelope.
func (f *FlagSpec) field() string {
	if f.Field != "" {
		return f.Field
	}
	return f.Name
}

// named is how a message names this entry: its long form, or its label when
// it is a positional and has no long form.
func (f *FlagSpec) named() string {
	if f.Label != "" {
		return f.Label
	}
	return f.long()
}

// long is how every message names this flag, whichever spelling was typed.
// Naming it always by its long form keeps a message from depending on how the
// call happened to be written.
func (f *FlagSpec) long() string { return "--" + f.Name }

// shortest is the spelling a hint uses when it shows the flag inside an
// example command line, so that the hint of --append-desc reads "-d @file.md"
// exactly as docs/spec/salida-y-terminal.md#codificación-y-texto prints it.
func (f *FlagSpec) shortest() string {
	if f.Short != "" {
		return "-" + f.Short
	}
	return f.named()
}

// CommandSpec is a command's own table: its name, its own flags, and the three
// properties that decide whether a global flag applies to it, from
// docs/spec/cmd/flags-globales.md.
type CommandSpec struct {
	Name  string
	Flags []FlagSpec

	// ReadOnly means no task of the board changes, which is what makes
	// --dry-run a usage error there. Writing files of its own does not take
	// a command off this list: export, snapshot and doctor without --fix
	// are read-only all the same.
	ReadOnly bool

	// WriteFlags are the long names that turn this read-only command into a
	// writing one, so that the table can say what
	// docs/spec/cmd/flags-globales.md declares: biso doctor is read-only,
	// and biso doctor --fix is not, so --dry-run is a usage error in the
	// first and valid in the second. A command that is never read-only
	// leaves it empty.
	WriteFlags []string

	// AffectsNoTask means the command affects no task that existed before,
	// which is what makes --print a usage error. init and config set are
	// the two that write without touching any existing task.
	AffectsNoTask bool

	// TextPositionalsFrom is the index from which this command's
	// positional arguments are long text values, and therefore go through
	// the three forms of
	// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo
	// exactly as a flag of kind TextValue does. It is zero for every
	// command whose positionals are not texts, and 1 for the four verbs of
	// the cycle whose first positional is the reference and whose rest are
	// the text (docs/spec/cmd/verbos-del-ciclo.md).
	TextPositionalsFrom int
}

// lookupLong finds a flag by its long name, the command's own first and the
// global ones after, so that a command can never shadow a global one: its own
// entry with that name would be found first, which is exactly what
// docs/spec/cmd/flags-globales.md forbids, so the search order is reversed on
// purpose and the global table wins.
func lookupLong(cmd *CommandSpec, name string) *FlagSpec {
	for i := range globalFlags {
		if globalFlags[i].Name == name {
			return &globalFlags[i]
		}
	}
	if cmd == nil {
		return nil
	}
	for i := range cmd.Flags {
		if cmd.Flags[i].Name == name {
			return &cmd.Flags[i]
		}
	}
	return nil
}

// tableOrder is the position of a flag in the specification table: the global
// ones first, in the order of docs/spec/cmd/flags-globales.md, and then the
// command's own, in the order of its table of parameters. A message that names
// two flags names them in this order, so that its text does not depend on
// which of the two the call happened to write first.
func tableOrder(cmd *CommandSpec, f *FlagSpec) int {
	for i := range globalFlags {
		if &globalFlags[i] == f {
			return i
		}
	}
	if cmd != nil {
		for i := range cmd.Flags {
			if &cmd.Flags[i] == f {
				return len(globalFlags) + i
			}
		}
	}
	// A flag that belongs to neither table cannot be named by a message the
	// parser builds, but ordering it last keeps the comparison total.
	return math.MaxInt
}

// lookupShort is lookupLong for a one-letter name.
func lookupShort(cmd *CommandSpec, name string) *FlagSpec {
	for i := range globalFlags {
		if globalFlags[i].Short != "" && globalFlags[i].Short == name {
			return &globalFlags[i]
		}
	}
	if cmd == nil {
		return nil
	}
	for i := range cmd.Flags {
		if cmd.Flags[i].Short != "" && cmd.Flags[i].Short == name {
			return &cmd.Flags[i]
		}
	}
	return nil
}

// globalFlags is the table of docs/spec/cmd/flags-globales.md. They are valid
// in every command, they can be written before or after the command name, and
// no command may redefine one or change what it means.
var globalFlags = []FlagSpec{
	{Name: "cwd", Short: "C", Value: PlainValue},
	{
		Name:       "color",
		Value:      PlainValue,
		Domain:     []string{"auto", "always", "never"},
		DomainCode: "invalid_color_mode",
	},
	{Name: "json", Conflicts: []string{"quiet", "print"}},
	{Name: "quiet", Short: "q", Conflicts: []string{"json", "print"}},
	{Name: "print", Conflicts: []string{"json", "quiet"}},
	{Name: "dry-run"},
	{Name: "version", Short: "V"},
	{Name: "help", Short: "h"},
}

// GlobalFlags returns a copy of the global table, for a test or a help text
// that needs to walk it without being able to change it.
func GlobalFlags() []FlagSpec {
	out := make([]FlagSpec, len(globalFlags))
	copy(out, globalFlags)
	return out
}

// commandNames returns the names of a table, sorted, for the hint of an
// unknown command.
func commandNames(commands []CommandSpec) []string {
	names := make([]string, 0, len(commands))
	for i := range commands {
		names = append(names, commands[i].Name)
	}
	sort.Strings(names)
	return names
}

// article answers "a" or "an" for a noun, which is all the hint of a malformed
// token needs: "a label", "an assignee", "an extension key".
func article(noun string) string {
	if noun == "" {
		return "a"
	}
	if strings.ContainsRune("aeiou", rune(noun[0])) {
		return "an"
	}
	return "a"
}
