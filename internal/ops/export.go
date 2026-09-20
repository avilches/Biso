package ops

import (
	"bytes"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is docs/spec/cmd/export.md: the whole board as NDJSON, one task
// per line, in exactly the shape `biso new --from` reads back. The shape
// itself is not here, it is in interchange.go, which is the one definition
// the two directions share.

// ExportParams is one `biso export` call, already read off the command
// line. The filters are the ones of `biso ls`, because that is what the
// specification gives this command; what it does not give it is the shaping
// flags, which a dump has no shape to choose.
type ExportParams struct {
	Filters

	// HasStatus says the call wrote -s at all. Without it `export` takes
	// every status, the terminal one included, which is the one difference
	// between its base and the base of `biso ls`.
	HasStatus bool
	// Mine is resolved into Assignee, like in a listing.
	Mine bool
	// NoArchived leaves the archived tasks out. Archived tasks come in by
	// default here, so this is the only flag of `export` about the archive.
	NoArchived bool
}

// ExportResult is what `biso export` answers: the bytes to write, whatever
// they are written to, and what was left behind.
type ExportResult struct {
	// NDJSON is the whole dump, one task per line, each line ended by a
	// newline.
	NDJSON []byte
	// Tasks is how many lines it carries.
	Tasks int
	// Skipped are the identifiers of the tasks that could not be read,
	// which is what makes the exit code 6 instead of 0
	// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
	Skipped  []string
	Warnings []Warning
	Notes    []string
}

// Export writes the board as NDJSON (docs/spec/cmd/export.md).
func Export(env Env, p ExportParams) (*ExportResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return ExportOn(b, env, p)
}

// ExportOn is Export over a board that is already open.
func ExportOn(b *board.Board, env Env, p ExportParams) (*ExportResult, error) {
	r := newReader(b, env)
	if err := r.load(); err != nil {
		return nil, err
	}

	filters, err := r.resolveFilters(exportAsListing(p))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	result := &ExportResult{}
	for _, t := range r.all {
		if !r.matches(t, filters) {
			continue
		}
		line, err := encodeTask(t)
		if err != nil {
			// A task that cannot be written as a line of the format is a
			// task this command cannot read for its purpose, so it takes
			// the same way out as an undecodable one: it is left out,
			// counted and named, and the exit code is 6.
			r.undecodable(t, &model.Error{
				ExitCode: 3,
				Code:     "undecodable_task",
				Message:  t.ID + " cannot be written as a line of the export format: " + err.Error(),
			})
			continue
		}
		buf.Write(line)
		buf.WriteByte('\n')
		result.Tasks++
	}

	r.warnAboutSkipped()
	result.NDJSON = buf.Bytes()
	result.Skipped = r.skippedIDs()
	result.Warnings = r.warnings
	result.Notes = r.notes
	return result, nil
}

// exportAsListing turns the parameters of `export` into the ones of a
// listing, which is what resolves the filters: the same filters, matched
// against the same vocabulary, with the two defaults of this command
// spelled out.
//
// Archived tasks come in unless --no-archived says otherwise, and without
// an explicit -s every status counts, the terminal one included. The second
// is written as --any-status, because that flag is exactly "every status
// the board configures" and `export` has no other way of saying it.
func exportAsListing(p ExportParams) ListParams {
	filters := p.Filters
	filters.Archived = !p.NoArchived
	filters.OnlyArchived = false
	if !p.HasStatus {
		filters.AnyStatus = true
	}
	return ListParams{Filters: filters, HasStatus: p.HasStatus, Mine: p.Mine}
}
