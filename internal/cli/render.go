package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file turns a result into the two shapes the specification fixes for
// it: the literal text of the "Salida" block of each command's page, and the
// envelope of docs/spec/contrato-json.md.

// SchemaVersion is the version of the JSON envelope.
const SchemaVersion = 1

// renderInit is the output block of docs/spec/cmd/init.md.
//
// The first line says which of the three endings happened, because "Created"
// would be a lie over a board that was already there: --overwrite-config
// rewrites the configuration of one that exists, and --at on a whole board
// adopts it without touching anything of it.
func renderInit(r *ops.InitResult) string {
	var b strings.Builder
	switch r.Action {
	case ops.Adopted:
		fmt.Fprintf(&b, "Adopted board %q\n", r.Board.Name)
	case ops.Rewrote:
		fmt.Fprintf(&b, "Rewrote the configuration of board %q\n", r.Board.Name)
	default:
		fmt.Fprintf(&b, "Created board %q\n", r.Board.Name)
	}
	row := func(label, value string) {
		fmt.Fprintf(&b, "  %-12s%s\n", label, value)
	}
	row("statuses", statusesWithRoles(r.Board))
	row("types", strings.Join(r.Board.Types, ", "))
	row("priorities", strings.Join(r.Board.Priorities, ", "))
	row("prefix", r.Board.TaskPrefix)

	b.WriteString("This project now points at that board.\n")
	if r.Action != ops.Rewrote {
		b.WriteString("Run `biso prime` to see how to use it.\n")
	}
	return b.String()
}

// statusesWithRoles writes the statuses separated by a pipe, each one
// followed by the role it holds, if it holds one. A board may carry statuses
// that are neither the initial, the active nor the terminal one, and that
// breaks nothing (docs/spec/cmd/init.md).
func statusesWithRoles(cfg ops.BoardSummary) string {
	parts := make([]string, 0, len(cfg.Statuses))
	for _, status := range cfg.Statuses {
		switch status {
		case cfg.InitialStatus:
			status += " (initial)"
		case cfg.ActiveStatus:
			status += " (active)"
		case cfg.TerminalStatus:
			status += " (terminal)"
		}
		parts = append(parts, status)
	}
	return strings.Join(parts, " | ")
}

// renderWhere is the output block of docs/spec/cmd/where.md: the four data
// that identify the board, who is calling, the counts, and the block of the
// discarded candidate when there was one.
func renderWhere(r *ops.WhereResult) string {
	var b strings.Builder
	row := func(label, value string) {
		fmt.Fprintf(&b, "%-9s%s\n", label, value)
	}
	row("id", r.ID)
	row("board", r.Board)
	row("path", r.Path)
	row("source", r.Source)
	row("me", identityOrNotice(r.Me))
	row("tasks", countsLine(r))

	if len(r.Discarded) == 0 {
		return b.String()
	}
	b.WriteString("\n")
	candidate := func(label, id, dir, reason string) {
		fmt.Fprintf(&b, "%-11s%s at %s\n", label, id, dir)
		fmt.Fprintf(&b, "%-11s%s\n", "", reason)
	}
	candidate("chosen", r.ID, r.Path, "because "+chosenReason(r.Source))
	for _, d := range r.Discarded {
		candidate("discarded", d.ID, d.Path, d.Reason)
	}
	return b.String()
}

// identityOrNotice is the `me` row, which says how to set an identity when
// there is none instead of failing or emitting a note of its own
// (docs/spec/invocacion.md#variables-de-entorno).
func identityOrNotice(me string) string {
	if me == "" {
		return "(not set: run biso as BISO_ME=@you biso ...)"
	}
	return me
}

// countsLine says "not archived" and not "active" because a task that is not
// archived can be in any status, the terminal one included, and "active" is
// the name of a status role (docs/spec/cmd/where.md).
func countsLine(r *ops.WhereResult) string {
	highest := "no id assigned yet"
	if r.Counts.HighestEverAssigned > 0 {
		highest = fmt.Sprintf("highest id ever assigned %s",
			taskID(r.Prefix, r.Counts.HighestEverAssigned))
	}
	return fmt.Sprintf("%d not archived, %d archived, %s",
		r.Counts.NotArchived, r.Counts.Archived, highest)
}

func taskID(prefix string, n int) string { return fmt.Sprintf("%s-%d", prefix, n) }

// chosenReason turns the sentence of the `source` row into the one the
// candidates block prints after "because".
func chosenReason(source string) string {
	if source == "the working directory is this board" {
		return "the working directory is itself a board"
	}
	return "this project's pointer names it"
}

// The JSON envelope of docs/spec/contrato-json.md#el-sobre. A successful
// call carries data and never error; a failed one carries error and never
// data.
type envelope struct {
	SchemaVersion int          `json:"schemaVersion"`
	Kind          string       `json:"kind"`
	GeneratedAt   string       `json:"generatedAt"`
	Data          any          `json:"data,omitempty"`
	Error         *errorObject `json:"error,omitempty"`
}

// errorObject is the error of
// docs/spec/contrato-json.md#los-errores-en-json. The three keys that are in
// every error are always written, and each of the detail keys only
// accompanies the codes its row documents, which is what lets a caller that
// branches on code know what it will find. field and given travel together,
// so they are pointers: an empty given is still a given.
type errorObject struct {
	ExitCode  int            `json:"exitCode"`
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	Field     *string        `json:"field,omitempty"`
	Given     *string        `json:"given,omitempty"`
	Valid     []string       `json:"valid,omitempty"`
	Details   []*errorObject `json:"details,omitempty"`
	VCSOutput []string       `json:"vcsOutput,omitempty"`
}

type initEnvelopeData struct {
	Board          initEnvelopeBoard `json:"board"`
	PointerCreated bool              `json:"pointerCreated"`
}

type initEnvelopeBoard struct {
	Name           string   `json:"name"`
	Statuses       []string `json:"statuses"`
	InitialStatus  string   `json:"initialStatus"`
	ActiveStatus   string   `json:"activeStatus"`
	TerminalStatus string   `json:"terminalStatus"`
	Types          []string `json:"types"`
	Priorities     []string `json:"priorities"`
	TaskPrefix     string   `json:"taskPrefix"`
}

type whereEnvelopeData struct {
	ID        string                   `json:"id"`
	Board     string                   `json:"board"`
	Path      string                   `json:"path"`
	Source    string                   `json:"source"`
	Me        *string                  `json:"me"`
	Counts    whereEnvelopeCounts      `json:"counts"`
	Discarded []whereEnvelopeDiscarded `json:"discarded"`
}

type whereEnvelopeCounts struct {
	NotArchived           int     `json:"notArchived"`
	Archived              int     `json:"archived"`
	HighestIDEverAssigned *string `json:"highestIdEverAssigned"`
}

type whereEnvelopeDiscarded struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

func initData(r *ops.InitResult) initEnvelopeData {
	return initEnvelopeData{
		Board: initEnvelopeBoard{
			Name:           r.Board.Name,
			Statuses:       r.Board.Statuses,
			InitialStatus:  r.Board.InitialStatus,
			ActiveStatus:   r.Board.ActiveStatus,
			TerminalStatus: r.Board.TerminalStatus,
			Types:          r.Board.Types,
			Priorities:     r.Board.Priorities,
			TaskPrefix:     r.Board.TaskPrefix,
		},
		PointerCreated: r.PointerWritten,
	}
}

func whereData(r *ops.WhereResult) whereEnvelopeData {
	data := whereEnvelopeData{
		ID:     r.ID,
		Board:  r.Board,
		Path:   r.Path,
		Source: r.Source,
		Counts: whereEnvelopeCounts{
			NotArchived: r.Counts.NotArchived,
			Archived:    r.Counts.Archived,
		},
		// A list that had nothing in it is [] and never null, and it is
		// always present (docs/spec/contrato-json.md#números-fechas-y-ausencias).
		Discarded: []whereEnvelopeDiscarded{},
	}
	if r.Me != "" {
		me := r.Me
		data.Me = &me
	}
	if r.Counts.HighestEverAssigned > 0 {
		highest := taskID(r.Prefix, r.Counts.HighestEverAssigned)
		data.Counts.HighestIDEverAssigned = &highest
	}
	for _, d := range r.Discarded {
		data.Discarded = append(data.Discarded, whereEnvelopeDiscarded{
			ID: d.ID, Path: d.Path, Reason: d.Reason,
		})
	}
	return data
}

// writeEnvelope writes a successful call's envelope to stdout.
func writeEnvelope(s Streams, env ops.Env, kind string, data any) {
	write(s.Stdout, envelope{
		SchemaVersion: SchemaVersion,
		Kind:          kind,
		GeneratedAt:   generatedAt(env),
		Data:          data,
	})
}

// writeErrorEnvelope writes a failed call's envelope to stderr, which is
// where an error goes with --json as well
// (docs/spec/contrato-json.md#los-errores-en-json).
//
// It does not go through the environment, because a call can fail before
// there is one: the timestamp of an error envelope comes from the clock
// directly.
func writeErrorEnvelope(s Streams, e *model.Error) {
	obj := &errorObject{
		ExitCode: e.ExitCode,
		Code:     e.Code,
		Message:  e.Message,
		Valid:    e.Valid,
	}
	if e.Field != "" {
		field, given := e.Field, e.Given
		obj.Field, obj.Given = &field, &given
	}
	for _, d := range e.Details {
		detail := &errorObject{ExitCode: d.ExitCode, Code: d.Code, Message: d.Message, Valid: d.Valid}
		if d.Field != "" {
			field, given := d.Field, d.Given
			detail.Field, detail.Given = &field, &given
		}
		obj.Details = append(obj.Details, detail)
	}
	obj.VCSOutput = e.VCSOutput

	write(s.Stderr, envelope{
		SchemaVersion: SchemaVersion,
		Kind:          "error",
		GeneratedAt:   generatedAt(ops.Env{Now: s.Now}.WithDefaults()),
		Error:         obj,
	})
}

// generatedAt is the instant of docs/spec/contrato-json.md#números-fechas-y-ausencias:
// ISO 8601 in UTC, to the second, ending in Z.
func generatedAt(env ops.Env) string {
	return env.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func write(w io.Writer, v envelope) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		// Every envelope this package builds is made of strings, numbers
		// and lists of them, so there is nothing here that can fail to
		// marshal.
		return
	}
	w.Write(append(b, '\n'))
}
