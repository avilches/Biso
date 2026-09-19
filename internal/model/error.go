// Package model defines biso's domain types that depend on no storage
// mechanism and no I/O: pure data and validation.
package model

// Error is biso's cross-cutting error type. It is born in the deepest
// layer that detects a case from the specification and travels unchanged
// up to the layer that translates it to text or to JSON: no intermediate
// package reinterprets or wraps it, it only propagates it.
//
// The first three fields are always present. The five detail fields
// accompany only the code they correspond to, per the table in
// docs/spec/contrato-json.md#los-errores-en-json, and stay at their zero
// value when they do not apply.
type Error struct {
	ExitCode int    // the exit code, see docs/spec/codigos-de-salida.md
	Code     string // the identifier, see docs/spec/contrato-json.md#los-identificadores-de-error
	Message  string // the text that follows "error: " on stderr

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
