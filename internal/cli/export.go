package cli

import (
	"fmt"
	"os"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the output of docs/spec/cmd/export.md: NDJSON on stdout or
// into the file --out names, and the one command of the program whose answer
// carries no envelope, because its format is already one JSON object per
// line.

func runExport(s Streams, p *Parsed, env ops.Env) int {
	if p.Has("json") {
		// --json is itself the invalid part of the call, so the refusal is
		// plain text on stderr and not the error envelope: there is no
		// output mode to wrap it in
		// (docs/spec/contrato-json.md#los-errores-en-json).
		fmt.Fprintln(s.Stderr,
			"error: --json does not apply to export, whose output is already NDJSON")
		return 2
	}
	params, err := exportParams(p)
	if err != nil {
		return fail(s, false, err, warningsOf(p))
	}
	result, err := ops.Export(env, params)
	if err != nil {
		return fail(s, false, err, warningsOf(p))
	}
	printWarnings(s, p)
	printOpsWarnings(s, result.Warnings)

	if err := writeExport(s, p, result.NDJSON); err != nil {
		return fail(s, false, err, nil)
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	if len(result.Skipped) > 0 {
		// A dump that lost a task cannot look like a clean one: this
		// command exists to lose nothing
		// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
		return 6
	}
	return 0
}

// writeExport puts the dump where --out says, which is standard output unless
// it names a file.
func writeExport(s Streams, p *Parsed, ndjson []byte) error {
	out, ok := p.Value("out")
	if !ok || out == "-" {
		_, err := s.Stdout.Write(ndjson)
		return err
	}
	if err := os.WriteFile(out, ndjson, 0o644); err != nil {
		return &model.Error{
			ExitCode: 8,
			Code:     "io_error",
			Message:  fmt.Sprintf("%s cannot be written: %s", out, err),
			Field:    "out",
			Given:    out,
		}
	}
	return nil
}

// exportParams turns the analyzed call into the typed parameters of the
// command. The filters are read exactly as `biso ls` reads them, because
// they are the same flags; what changes is the base they apply to, and
// that is internal/ops's business and not this layer's.
func exportParams(p *Parsed) (ops.ExportParams, error) {
	if len(p.Positionals) > 0 {
		return ops.ExportParams{}, errUnexpectedArgument(p.Positionals[0])
	}
	params := ops.ExportParams{
		Filters:    filtersOf(p),
		HasStatus:  p.Has("status"),
		Mine:       p.Has("mine"),
		NoArchived: p.Has("no-archived"),
	}
	if v, ok := p.Value("due-before"); ok {
		if _, err := ops.ParseCalendarDay("due-before", v); err != nil {
			return ops.ExportParams{}, err
		}
		params.DueBefore = &v
	}
	return params, nil
}
