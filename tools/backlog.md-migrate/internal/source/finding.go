package source

import "fmt"

// Finding is one thing the reader could not carry over as-is while reading a
// Backlog.md board: an unparseable frontmatter, an unrecognized frontmatter
// key or body section, a date whose shape does not match either of
// Backlog.md's two forms, a multi-line acceptance criterion or Definition of
// Done item that had to be joined, a comment block without a "created"
// line, or a non-empty drafts/docs/decisions folder.
//
// A Finding never causes the read to stop: the rest of the board keeps being
// read normally. Printing findings and deciding an exit code is the job of a
// later phase (see docs/especificacion.md, "Los hallazgos"), not of this
// package, so Finding only carries the pieces that phase needs.
type Finding struct {
	// File is the source file name the finding is about (the base name,
	// not a full path), or "-" when the finding is about the board as a
	// whole rather than a single file (for example, a non-empty drafts
	// folder).
	File string
	// Field is the frontmatter key, the body section name, or the biso
	// field the finding is about.
	Field string
	// Message describes the finding in English.
	Message string
}

// String renders the finding the way docs/especificacion.md, "Los
// hallazgos", specifies it should eventually be printed:
//
//	warning: <file>: <field>: <message>
//
// This package never writes it anywhere itself; String only exists so a
// later phase does not have to re-derive the exact format.
func (f Finding) String() string {
	return fmt.Sprintf("warning: %s: %s: %s", f.File, f.Field, f.Message)
}
