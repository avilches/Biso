// Package model defines biso's domain types that depend on no storage
// mechanism and no I/O: pure data and validation.
package model

// Error is biso's cross-cutting error type. It is born in the deepest
// layer that detects a case from the specification and travels unchanged
// up to the layer that translates it to text or to JSON: no intermediate
// package reinterprets or wraps it, it only propagates it.
//
// ExitCode, Code and Message are always present. Hints carries one entry
// per "hint: " line that follows the message on stderr, in the order the
// specification prints them, and is empty when the case has none: two cases
// of docs/spec/ print two of those lines, so this is a list and not a
// single string. Notes is the same thing for the "note: " lines, which the
// two identifier errors of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro print
// between the message and the hints. None of the three reaches the JSON
// envelope of docs/spec/contrato-json.md#los-errores-en-json, which has
// neither a hint key, nor a note one, nor a detail one.
//
// Each entry of Hints and of Notes is one message, and it may carry its own
// line breaks: several cases of docs/spec/ print one over more than one
// display line, and where they break is fixed text and not a width, because
// the same note breaks between the same words whatever the value
// interpolated into it (compare the two spellings of the relative-path note
// in docs/spec/cmd/init.md). Whoever writes to stderr prints the prefix on
// the first line and aligns every further line under it, which is six
// spaces for both "note: " and "hint: ".
//
// Detail is a third kind of line, and it is verbatim: the display lines that
// two cases of the specification print between the message and the hints,
// carrying their own indentation and their own inner columns. They are the
// two directories of a duplicated board id
// (docs/spec/resolucion-del-tablero.md#el-mismo-id-en-dos-sitios) and the
// "searched" block of `biso where` (docs/spec/cmd/where.md). Neither one
// fits in Notes, which would prefix them, nor in the message, which is one
// line and also travels to the JSON envelope.
//
// The five detail fields accompany only the code they correspond to, per
// the table in docs/spec/contrato-json.md#los-errores-en-json, and stay at
// their zero value when they do not apply.
type Error struct {
	ExitCode int      // the exit code, see docs/spec/codigos-de-salida.md
	Code     string   // the identifier, see docs/spec/contrato-json.md#los-identificadores-de-error
	Message  string   // the text that follows "error: " on stderr
	Detail   []string // verbatim display lines printed right after the message, before the notes, empty when the case has none
	Notes    []string // the "note: " lines that follow those, in order, empty when the case has none
	Hints    []string // the "hint: " lines that follow those, in order, empty when the case has none

	Field     string   // on exit code 3, and exit code 2 errors that name a flag
	Given     string   // always accompanies Field
	Valid     []string // on errors that reject a value against a known set
	Details   []*Error // only on batch_invalid and dry_run_failed, same shape, one per failure
	VCSOutput []string // only on vcs_commit_failed and vcs_push_failed
}

// Error implements the standard library's error interface.
func (e *Error) Error() string {
	return e.Message
}
