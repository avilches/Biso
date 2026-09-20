package ops

import (
	"fmt"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is docs/spec/cmd/verbos-del-ciclo.md: `start`, `note`,
// `comment`, `finish`, `ask` and `answer`.
//
// The six are sugar over `biso set` and they are written on top of it, never
// beside it. Each one takes the same field flags, hands them to the same
// engine of write.go through the same loop of set.go, and adds three things
// of its own: a default or two, the refusals of its table, and the effects
// that have no flag (the lease `start` claims, the question `ask` fills and
// `answer` empties). Nothing here opens a transaction, resolves a reference
// or writes a field: a second implementation of any of that is exactly what
// doing this step after `biso set` was meant to avoid.

// Text is one positional text of a verb of the cycle: what was typed on the
// command line, and the value it resolved to through the three forms of
// docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo.
//
// The two halves both travel because each answers a different question. The
// value is what gets written; the typed form is what the rule of the
// positional that looks like an identifier judges, which is what keeps
// `biso note MYP-11 @question.md` working when the file happens to hold
// nothing but "MYP-2".
type Text struct {
	Typed string
	Value string
}

// StartParams is one `biso start` call (docs/spec/cmd/verbos-del-ciclo.md#biso-start).
type StartParams struct {
	Refs []string
	Mode RefMode
	// Reopen allows starting a task that is already in the terminal
	// status, which is otherwise exit code 6.
	Reopen  bool
	Changes []Change
	DryRun  bool
	Print   bool
}

// NoteParams is one `biso note` call, and the same shape serves `comment`,
// `ask` and `answer`: the four take exactly one reference and one or more
// positional texts (docs/spec/cmd/verbos-del-ciclo.md).
type NoteParams struct {
	Ref     string
	Mode    RefMode
	Texts   []Text
	Changes []Change
	DryRun  bool
	Print   bool
}

// CommentParams is one `biso comment` call. --comment-author travels among
// the changes, like everywhere else: it is the same flag of the same table,
// and this command differs only in taking it without --comment
// (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
type CommentParams = NoteParams

// AskParams is one `biso ask` call, and AnswerParams one `biso answer`.
type (
	AskParams    = NoteParams
	AnswerParams = NoteParams
)

// FinishParams is one `biso finish` call
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
type FinishParams struct {
	Refs []string
	Mode RefMode
	// Strict refuses to close a task that is missing something, and its
	// default is the board's finish_strict. NoChecks skips every check and
	// every warning, and the two cannot be written together.
	Strict   bool
	NoChecks bool
	Changes  []Change
	DryRun   bool
	Print    bool
}

// Start takes one or more tasks (docs/spec/cmd/verbos-del-ciclo.md#biso-start).
func Start(env Env, p StartParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return StartOn(b, env, p)
}

// StartOn is Start over a board that is already open.
func StartOn(b *board.Board, env Env, p StartParams) (*WriteResult, error) {
	changes := p.Changes
	if !writesFlag(changes, "status") {
		// -s names another status, and then there is no lease to claim;
		// with no -s the status is the board's active one, which is the
		// whole point of the verb.
		changes = append(changes, Change{
			Flag: "status", Step: StepScalar, Value: b.Config.ActiveStatus,
		})
	}

	return writeOn(b, env, SetParams{
		Refs: p.Refs, Mode: p.Mode, Changes: changes,
		DryRun: p.DryRun, Print: p.Print,
	}, verb{
		name: "start",
		before: func(w *writer, t *model.Task) error {
			// Archived first: a task can be archived and finished at
			// once, and being archived is the harder refusal, the only
			// one of the six verbs, because only `start` claims a lease
			// (docs/spec/lease.md#el-vaciado).
			if t.Archived {
				return &model.Error{
					ExitCode: 6,
					Code:     "precondition_failed",
					Message:  t.ID + " is archived",
					Hints: []string{fmt.Sprintf(
						"unarchive it first with `biso archive %s --unarchive`", t.ID)},
				}
			}
			if t.Status == w.b.Config.TerminalStatus && !p.Reopen {
				return &model.Error{
					ExitCode: 6,
					Code:     "already_finished",
					Message:  fmt.Sprintf("%s is already %s", t.ID, t.Status),
					Hints: []string{fmt.Sprintf(
						"start it again with `biso start %s --reopen`", t.ID)},
				}
			}
			return nil
		},
		after: func(w *writer, t, before *model.Task, byID map[string]*model.Task) error {
			w.startNotes(t, before)
			w.assignToCaller(t, before)
			w.warnAboutUnresolvedDependencies(t, byID)
			w.warnAboutOpenQuestionOnStart(t)
			return nil
		},
		settled: func(w *writer, t, before *model.Task) {
			if t.Status != w.b.Config.ActiveStatus {
				// Outside the active status there is no lease to claim,
				// and the one the task had was emptied a moment ago by
				// docs/spec/lease.md#el-vaciado.
				return
			}
			expired := before.LeaseHolder != "" &&
				!before.LeaseExpiresAt.IsZero() &&
				!before.LeaseExpiresAt.After(w.now)
			w.claimLease(t)
			if expired && t.LeaseHolder == w.env.Me {
				// Reclaiming an expired lease is checked again inside
				// the transaction that writes it, so that of two
				// simultaneous claims only one wins
				// (docs/spec/cmd/verbos-del-ciclo.md#biso-start).
				w.claims = append(w.claims, board.LeaseClaim{
					TaskID: t.ID, Holder: w.env.Me, Now: w.now,
				})
			}
		},
	})
}

// startNotes is the three notes of the table of
// docs/spec/cmd/verbos-del-ciclo.md#biso-start that speak of the status.
func (w *writer) startNotes(t, before *model.Task) {
	active := w.b.Config.ActiveStatus
	switch {
	case t.Status != active:
		w.note(fmt.Sprintf("%s was moved to %s, no lease was claimed", t.ID, t.Status))
	case before.Status == active:
		w.note(fmt.Sprintf("%s was already %s", t.ID, active))
	}
}

// assignToCaller is the second of the four things `biso start` does: the
// task goes to whoever is calling, but only when nobody had it. A task
// somebody else already has is left alone, with a note, because taking it
// away is not what starting means.
func (w *writer) assignToCaller(t, before *model.Task) {
	switch {
	case len(before.Assignees) > 0:
		if !containsString(before.Assignees, w.env.Me) {
			w.note(fmt.Sprintf("%s is assigned to %s, left as is",
				t.ID, before.Assignees[0]))
		}
	case len(t.Assignees) > 0:
		// -a named somebody on a task that had nobody, so that is who
		// has it now and `me` is not added on top.
	case w.env.Me == "":
		// No identity to attribute it to, so the task is left
		// unassigned and, by the invariant of docs/spec/lease.md,
		// without a lease either.
		w.note("no identity configured, task left unassigned")
	default:
		t.Assignees = append(t.Assignees, w.env.Me)
	}
}

// Note appends one or more paragraphs to the implementation notes of one
// task (docs/spec/cmd/verbos-del-ciclo.md#biso-note).
func Note(env Env, p NoteParams) (*WriteResult, error) {
	return onOpenBoard(env, func(b *board.Board) (*WriteResult, error) {
		return NoteOn(b, env, p)
	})
}

// NoteOn is Note over a board that is already open.
func NoteOn(b *board.Board, env Env, p NoteParams) (*WriteResult, error) {
	if len(p.Texts) == 0 && !writesAnyField(p.Changes) {
		return nil, missingText("note", "a text to append",
			`biso note `+p.Ref+` "The parser already normalized LF, CRLF was missing"`)
	}
	changes, warnings, err := textChanges(b, "note", "append-note", StepAdd, p, noteIDLikeHint)
	if err != nil {
		return nil, err
	}
	return writeOn(b, env, setParamsOf(p, changes, warnings), verb{name: "note"})
}

// Comment appends a discussion comment to one task
// (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
func Comment(env Env, p CommentParams) (*WriteResult, error) {
	return onOpenBoard(env, func(b *board.Board) (*WriteResult, error) {
		return CommentOn(b, env, p)
	})
}

// CommentOn is Comment over a board that is already open.
func CommentOn(b *board.Board, env Env, p CommentParams) (*WriteResult, error) {
	if len(p.Texts) == 0 && !writesAnyField(p.Changes) {
		return nil, missingText("comment", "a text to append",
			`biso comment `+p.Ref+` "A user with a Windows clone reported this"`)
	}
	changes, warnings, err := textChanges(b, "comment", "comment", StepComment, p, commentIDLikeHint)
	if err != nil {
		return nil, err
	}
	return writeOn(b, env, setParamsOf(p, changes, warnings), verb{
		name: "comment",
		after: func(w *writer, t, before *model.Task, _ map[string]*model.Task) error {
			// The key the comment just took is the one --rm-comment and
			// --set-comment-date accept afterwards, so it is worth a
			// note (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
			//
			// Which ones are new is asked of the keys and not of the
			// length of the list, because the same call may have removed
			// some with --rm-comment. The counter only grows, so a key
			// the task had not handed out yet is a comment of this call
			// (docs/spec/modelo-de-datos/comentarios.md).
			first := before.NextCommentKey
			for _, c := range t.Comments {
				if c.Key >= first {
					w.note(fmt.Sprintf("comment #%d by %s", c.Key, c.Author))
				}
			}
			return nil
		},
	})
}

// Ask parks one task on a question for a person
// (docs/spec/cmd/verbos-del-ciclo.md#biso-ask).
func Ask(env Env, p AskParams) (*WriteResult, error) {
	return onOpenBoard(env, func(b *board.Board) (*WriteResult, error) {
		return AskOn(b, env, p)
	})
}

// AskOn is Ask over a board that is already open.
func AskOn(b *board.Board, env Env, p AskParams) (*WriteResult, error) {
	body, err := questionBody(b, env, "ask", "a question", "question", p)
	if err != nil {
		return nil, err
	}
	return writeOn(b, env, setParamsOf(p, p.Changes, nil), verb{
		name: "ask",
		before: func(w *writer, t *model.Task) error {
			if t.Question != nil {
				return &model.Error{
					ExitCode: 6,
					Code:     "open_question_exists",
					Message:  t.ID + " already has an open question",
					Hints: []string{fmt.Sprintf(
						"answer it first with `biso answer %s <text>`", t.ID)},
				}
			}
			if t.Status == w.b.Config.TerminalStatus {
				return &model.Error{
					ExitCode: 6,
					Code:     "already_finished",
					Message:  fmt.Sprintf("%s is already %s", t.ID, t.Status),
					Hints: []string{fmt.Sprintf(
						"reopen it first with `biso start %s --reopen`", t.ID)},
				}
			}
			return nil
		},
		after: func(w *writer, t, _ *model.Task, _ map[string]*model.Task) error {
			// The author and the instant are the program's, and only the
			// body is the caller's
			// (docs/spec/modelo-de-datos/pregunta-abierta.md). The status
			// is not touched.
			t.Question = &model.Question{Author: w.env.Me, AskedAt: w.now, Body: body}
			return nil
		},
	})
}

// Answer answers the open question of one task and unparks it
// (docs/spec/cmd/verbos-del-ciclo.md#biso-answer).
func Answer(env Env, p AnswerParams) (*WriteResult, error) {
	return onOpenBoard(env, func(b *board.Board) (*WriteResult, error) {
		return AnswerOn(b, env, p)
	})
}

// AnswerOn is Answer over a board that is already open.
func AnswerOn(b *board.Board, env Env, p AnswerParams) (*WriteResult, error) {
	body, err := questionBody(b, env, "answer", "an answer", "answer", p)
	if err != nil {
		return nil, err
	}
	// The question that is about to become a comment, read before anything
	// is written and kept for the three effects below.
	var asked *model.Question
	return writeOn(b, env, setParamsOf(p, p.Changes, nil), verb{
		name: "answer",
		before: func(w *writer, t *model.Task) error {
			if t.Question == nil {
				return &model.Error{
					ExitCode: 6,
					Code:     "no_open_question",
					Message:  t.ID + " has no open question",
					Hints:    []string{"use `biso comment` to add a comment"},
				}
			}
			asked = t.Question
			// A task in the terminal status answers like any other: it
			// is the one way of recovering a question that stayed open
			// when the task was closed, and the asymmetry with `biso
			// ask` is deliberate.
			w.leading = func(t *model.Task) error {
				// 1. The question becomes a comment, with the author and
				// the instant it was asked with, not the ones of
				// whoever answers.
				t.AddComment(asked.Author, asked.AskedAt, asked.Body)
				// 2. The answer goes behind it, signed by the caller.
				t.AddComment(w.env.Me, w.now, body)
				return nil
			}
			return nil
		},
		after: func(w *writer, t, _ *model.Task, _ map[string]*model.Task) error {
			// 3. Emptying the question is a step of this verb's own,
			// after all nine of
			// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura,
			// and it is always the last effect of the write.
			t.Question = nil
			return nil
		},
	})
}

// Finish closes one or more tasks
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
func Finish(env Env, p FinishParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return FinishOn(b, env, p)
}

// FinishOn is Finish over a board that is already open.
func FinishOn(b *board.Board, env Env, p FinishParams) (*WriteResult, error) {
	changes := p.Changes
	if !writesFlag(changes, "status") {
		changes = append(changes, Change{
			Flag: "status", Step: StepScalar, Value: b.Config.TerminalStatus,
		})
	}
	// --strict defaults to the board's finish_strict, and --no-checks wins
	// over both: it skips every check, so there is nothing left to be
	// strict about (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
	strict := (p.Strict || b.Config.FinishStrict) && !p.NoChecks

	result, err := writeOn(b, env, SetParams{
		Refs: p.Refs, Mode: p.Mode, Changes: changes,
		DryRun: p.DryRun, Print: p.Print,
	}, verb{
		name:                "finish",
		ownTerminalWarnings: true,
		after: func(w *writer, t, before *model.Task, byID map[string]*model.Task) error {
			if before.Status == t.Status && t.Status == w.b.Config.TerminalStatus {
				w.note(fmt.Sprintf("%s was already %s", t.ID, t.Status))
			}
			if p.NoChecks {
				return nil
			}
			return w.finishChecks(t, byID, strict)
		},
	})
	return result, err
}

// finishChecks is the table of docs/spec/cmd/verbos-del-ciclo.md#biso-finish.
//
// **Every one of them is read off the task as this write leaves it, never
// off which flags this call happened to write.** A task that already had all
// its criteria checked does not warn for not passing --check-ac now, and one
// that kept the summary of an earlier close does not warn for not passing
// --append-summary.
//
// The open question is the one that warns and never refuses, not even with
// --strict: refusing would only push the caller into `biso set`.
func (w *writer) finishChecks(t *model.Task, byID map[string]*model.Task, strict bool) error {
	if t.Status != w.b.Config.TerminalStatus {
		// These are the checks of arriving at a terminal status, and -s
		// named another one: nothing is being closed here, so there is
		// nothing to warn about and nothing for --strict to refuse
		// (docs/spec/salida-y-terminal.md#notas-y-avisos). The rest of
		// the write, the lease included, happens all the same.
		return nil
	}
	blocking := []*Warning{
		acUncheckedWarning(t, t.Status),
		noSummaryWarning(t),
		w.unfinishedSubtasks(t, byID),
	}
	if strict {
		var missing []string
		for _, warning := range blocking {
			if warning != nil {
				missing = append(missing, "  "+warning.Message)
			}
		}
		if len(missing) > 0 {
			return &model.Error{
				ExitCode: 6,
				Code:     "precondition_failed",
				Message:  t.ID + " is not ready to finish",
				Detail:   missing,
				Hints:    []string{"finish it without --strict, or write what is missing"},
			}
		}
	}
	for _, warning := range blocking {
		if warning != nil {
			w.warn(*warning)
		}
	}
	if warning := openQuestionOnTerminalWarning(t, t.Status); warning != nil {
		w.warn(*warning)
	}
	return nil
}

// setParamsOf is the one-reference shape of the four verbs that take one,
// written as the parameters of the shared loop.
func setParamsOf(p NoteParams, changes []Change, warnings []Warning) SetParams {
	return SetParams{
		Refs: []string{p.Ref}, Mode: p.Mode, Changes: changes,
		DryRun: p.DryRun, Print: p.Print, Warnings: warnings,
		// The four of them are a write even when they change no field:
		// `biso note MYP-11 ""` adds nothing and says so, and `biso ask`
		// writes a question that no field flag can write, so neither one
		// is the "nothing to change" of `biso set`.
		AllowNoChanges: true,
	}
}

// onOpenBoard is the entry point every verb that takes one reference shares:
// it opens the board, runs the logic and closes it.
func onOpenBoard(env Env, run func(*board.Board) (*WriteResult, error)) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return run(b)
}

// textChanges turns the positional texts of `biso note` and `biso comment`
// into the changes of the field flag each one is sugar for, in front of the
// values of that same flag written by hand, which is the order the command
// line had.
//
// It is also where the rule of the positional that looks like an identifier
// lives: these two verbs take one single task, unlike `set`, `start`,
// `finish` and `archive`, and a second identifier among the texts is far
// more likely to be that mistake than a note that really says "MYP-2"
// (docs/spec/cmd/verbos-del-ciclo.md#el-posicional-que-parece-un-identificador).
func textChanges(b *board.Board, command, flag string, step Step,
	p NoteParams, hint func(ref, text string) []string) ([]Change, []Warning, error) {
	var (
		out      []Change
		warnings []Warning
	)
	for _, text := range p.Texts {
		if err := rejectIDLike(b, command, p.Ref, text, hint); err != nil {
			return nil, nil, err
		}
		if strings.TrimSpace(text.Value) == "" {
			// The empty value adds nothing and says so, exactly as it
			// does behind the flag this verb stands for
			// (docs/spec/valores-de-entrada.md#el-valor-vacío).
			warnings = append(warnings, Warning{
				Code:    "empty_append",
				Message: "--" + flag + ": empty value, nothing was added",
				Fields:  map[string]any{"flag": "--" + flag},
			})
			continue
		}
		out = append(out, Change{Flag: flag, Step: step, Value: text.Value})
	}
	return append(out, p.Changes...), warnings, nil
}

// questionBody is the text of `biso ask` and of `biso answer`: one paragraph
// per positional, the identity they both need, and the three ways the call
// can be wrong before any task is read.
//
// what is the noun with the article the page wrote in front of it, "a
// question" and "an answer", because the two messages are literal text of
// docs/spec/cmd/verbos-del-ciclo.md and neither one is built out of the
// other.
func questionBody(b *board.Board, env Env, command, what, noun string, p NoteParams) (string, error) {
	var paragraphs []string
	for _, text := range p.Texts {
		if err := rejectIDLike(b, command, p.Ref, text, literalTextHint); err != nil {
			return "", err
		}
		paragraphs = append(paragraphs, text.Value)
	}
	if len(paragraphs) == 0 {
		return "", missingText(command, what,
			fmt.Sprintf("biso %s %s \"...\"", command, p.Ref))
	}
	if env.Me == "" {
		// The author of a question and the author of an answer are
		// always the configured identity, and neither command takes
		// --comment-author (docs/spec/invocacion.md#variables-de-entorno).
		return "", &model.Error{
			ExitCode: 2,
			Code:     "missing_identity",
			Message: fmt.Sprintf(
				"biso %s needs an identity; set BISO_ME, or add \"me\" to ~/.biso/config.json",
				command),
			Field: "me",
		}
	}
	body := strings.Join(paragraphs, "\n\n")
	if strings.TrimSpace(body) == "" {
		return "", &model.Error{
			ExitCode: 3,
			Code:     "empty_scalar_value",
			Message:  "the " + noun + " cannot be empty",
			Field:    "question",
			Given:    "",
		}
	}
	return body, nil
}

// rejectIDLike is the error 2 of a positional text that the grammar of
// docs/spec/referencias.md#la-gramática would read as an identifier.
func rejectIDLike(b *board.Board, command, ref string, text Text,
	hint func(ref, text string) []string) error {
	if _, isID := parseTaskRef(b.Config.TaskPrefix, text.Typed); !isID {
		return nil
	}
	return &model.Error{
		ExitCode: 2,
		Code:     "id_like_positional",
		Message: fmt.Sprintf("%q looks like a task id, and `biso %s` takes only one task",
			text.Typed, command),
		Hints: hint(ref, text.Typed),
		Field: "text",
		Given: text.Typed,
	}
}

// The hints of that error, one per verb. `note` and `comment` each point at
// the field flag that writes the same thing without going through this
// check; `ask` and `answer` have no such flag, so the only way out is a file
// or standard input.
func noteIDLikeHint(ref, text string) []string {
	return []string{alignedHint(
		"to note the same thing on several tasks: ",
		fmt.Sprintf("biso set %s %s --append-note \"...\"", ref, text),
		"to write that text literally:",
		fmt.Sprintf("biso note %s --append-note %q", ref, text))}
}

func commentIDLikeHint(ref, text string) []string {
	return []string{alignedHint(
		"to comment the same thing on several tasks: ",
		fmt.Sprintf("biso set %s %s --comment \"...\"", ref, text),
		"to write that text literally:",
		fmt.Sprintf("biso comment %s --comment %q", ref, text))}
}

func literalTextHint(string, string) []string {
	return []string{"to write that text literally, use @file or - for stdin"}
}

// alignedHint writes the two lines of that hint with their command lines in
// the same column, which is how the block of the specification prints them.
func alignedHint(firstLabel, firstCommand, secondLabel, secondCommand string) string {
	padded := secondLabel
	for len(padded) < len(firstLabel) {
		padded += " "
	}
	return firstLabel + firstCommand + "\n" + padded + secondCommand
}

// missingText is the error 2 of a verb called with no text at all.
//
// It carries neither `field` nor `given`: those two belong to the errors
// that name a flag, a configuration key or a concrete value
// (docs/spec/contrato-json.md#los-errores-en-json), and an argument that was
// never written is none of the three, exactly like the reference that does
// not exist.
func missingText(command, what, example string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "missing_text",
		Message:  fmt.Sprintf("biso %s needs %s", command, what),
		Hints:    []string{example},
	}
}

// writesAnyField answers whether the call carries a field flag of
// docs/spec/familias-de-flags.md, which is the other half of the refusal of
// a verb called with no text at all.
//
// --comment-author is the one flag those verbs take that writes no field of
// its own: it only says who signs a comment that another flag writes. A
// `biso comment` that carries nothing but the author can produce no comment
// at all, so it is the same refusal as a call with nothing
// (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
func writesAnyField(changes []Change) bool {
	for _, c := range changes {
		if c.Flag != "comment-author" {
			return true
		}
	}
	return false
}

// writesFlag answers whether the call already wrote that field flag, which
// is how `start` and `finish` tell their own default apart from an explicit
// -s.
func writesFlag(changes []Change, flag string) bool {
	for _, c := range changes {
		if c.Flag == flag {
			return true
		}
	}
	return false
}
