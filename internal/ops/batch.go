package ops

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"biso/internal/board"
	"biso/internal/match"
	"biso/internal/model"
)

// This file is the batch of docs/spec/cmd/new.md#el-modo-lote: NDJSON in,
// one task per line, the whole file judged before a single row is written.
//
// It is one engine and not two, because `biso init --from` restores a
// snapshot by doing exactly this over a board it has just created
// (docs/spec/cmd/init.md). The difference between the two is which board
// the lines are judged against, and that is an argument.

// BatchParams is one `biso new --from` call: the NDJSON, already read from
// the file or from standard input by the layer that owns them.
type BatchParams struct {
	Content string
	DryRun  bool
}

// batchLine is one line of the file that passed its own validation: the
// task it describes, the number of the line it came from, and the keys the
// conversion of definitionOfDone created on it.
type batchLine struct {
	number  int
	task    *model.Task
	dodKeys []int
}

// NewBatch creates every task of an NDJSON file, all of them or none
// (docs/spec/cmd/new.md#el-modo-lote).
func NewBatch(env Env, p BatchParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return NewBatchOn(b, env, p)
}

// NewBatchOn is NewBatch over a board that is already open.
func NewBatchOn(b *board.Board, env Env, p BatchParams) (*WriteResult, error) {
	env = env.WithDefaults()
	lines, err := validateBatch(b, env, p.Content)
	if err != nil {
		return nil, err
	}

	result := &WriteResult{
		Created: true, Batch: true, DryRun: p.DryRun, Previewed: len(lines),
	}
	for _, line := range lines {
		if len(line.dodKeys) > 0 {
			result.Warnings = append(result.Warnings, dodWarning(line))
		}
	}
	if p.DryRun {
		// Nothing is written and no identifier is spent, so there is no
		// task to name: the whole answer is the count and the warnings the
		// real call would have produced (docs/spec/cmd/new.md#el-modo-lote).
		return result, nil
	}

	tasks := make([]*model.Task, 0, len(lines))
	for _, line := range lines {
		tasks = append(tasks, line.task)
	}
	if err := b.Tasks.CreateAll(tasks); err != nil {
		return result, err
	}

	// Every task of the batch is summarized the same way `biso new` and
	// `biso set` summarize theirs, because the schema of
	// docs/spec/cmd/set.md is one schema: `urgency` is a key that is always
	// there and `changed` is the list of the fields that really changed,
	// which for a task that has just been created is every field it
	// carries. The board is read once, after the whole batch is written, so
	// that the urgency of a line that blocks another one is the urgency of
	// the board this very call leaves behind.
	w := newWriter(b, env, nil)
	byID, boardErr := w.boardAfter(nil)
	if boardErr != nil {
		return result, boardErr
	}
	for _, line := range lines {
		summary, summaryErr := w.summarize(line.task, changedFields(&model.Task{}, line.task), byID)
		if summaryErr != nil {
			return result, summaryErr
		}
		// The keys the conversion of definitionOfDone created are the one
		// thing `biso new` announces about a key it assigned, because they
		// depend on what acceptanceCriteria brought on that same line
		// (docs/spec/cmd/new.md#el-modo-lote).
		if line.dodKeys != nil {
			summary.AcAdded = line.dodKeys
		}
		result.Tasks = append(result.Tasks, summary)
	}
	return result, nil
}

// dodWarning is the one of
// docs/spec/salida-y-terminal.md#notas-y-avisos for a line that converted
// at least one element of definitionOfDone. It names the line and not the
// task, so that it reads the same with --dry-run as without it.
func dodWarning(line *batchLine) Warning {
	message := fmt.Sprintf("line %d: %d definition-of-done items imported as acceptance criteria",
		line.number, len(line.dodKeys))
	if len(line.dodKeys) == 1 {
		message = fmt.Sprintf("line %d: 1 definition-of-done item imported as an acceptance criterion",
			line.number)
	}
	return Warning{
		Code:    "imported_dod_merged",
		Message: message,
		Fields:  map[string]any{"line": line.number, "count": len(line.dodKeys)},
	}
}

// validateBatch reads and judges the whole file against the board, and
// answers either every line or every failure.
//
// It never stops at the first bad line: the code 7 of
// docs/spec/codigos-de-salida.md arrives with all of them, because a caller
// fixing an import wants the list and not one entry of it at a time.
func validateBatch(b *board.Board, env Env, content string) ([]*batchLine, *model.Error) {
	existing, _, err := b.Tasks.All()
	if err != nil {
		if e, ok := err.(*model.Error); ok {
			return nil, e
		}
		return nil, &model.Error{ExitCode: 8, Code: "io_error", Message: err.Error()}
	}
	return validateBatchAgainst(b.Config, existing, env, content)
}

// validateBatchAgainst is validateBatch over a configuration and a set of
// tasks instead of over an open board.
//
// `biso init --from` is why it is written this way: it restores a board
// that does not exist yet, so it has to judge the whole snapshot against
// the vocabulary the snapshot itself brings, before creating anything
// (docs/spec/cmd/init.md). Judging it after creating the board would leave
// a board behind on a failure that promises nothing was written.
func validateBatchAgainst(cfg board.Config, existing []*model.Task, env Env, content string) ([]*batchLine, *model.Error) {
	now := env.Now().UTC().Truncate(time.Second)

	taken := make(map[string]bool, len(existing))
	byID := make(map[string]*model.Task, len(existing))
	for _, t := range existing {
		taken[t.ID] = true
		byID[t.ID] = t
	}

	var lines []*batchLine
	// failures carry the line they came from beside the error, so that the
	// report can be put back in the order of the file however late a
	// failure is found (docs/spec/cmd/new.md#el-modo-lote).
	var failures []*batchFailure
	// claimedBy remembers which line of this same file took an identifier,
	// which is what tells the two collisions of `id_taken` apart: one is
	// about the board and the other one is not
	// (docs/spec/contrato-json.md#los-identificadores-de-error).
	claimedBy := map[string]int{}
	counted := 0
	for number, text := range strings.Split(content, "\n") {
		text = strings.TrimSpace(text)
		// A blank line and one that starts with # are ignored, which is
		// what lets a file carry a header of its own
		// (docs/spec/cmd/new.md#el-modo-lote).
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		counted++
		line := &batchLine{number: number + 1}
		if e := readBatchLine(cfg, line, text, now); e != nil {
			failures = append(failures, &batchFailure{line.number, e})
			continue
		}
		if line.task.ID != "" {
			if taken[line.task.ID] {
				failures = append(failures, &batchFailure{
					line.number, idTaken(line.task.ID, claimedBy),
				})
				continue
			}
			taken[line.task.ID] = true
			claimedBy[line.task.ID] = line.number
			byID[line.task.ID] = line.task
		}
		lines = append(lines, line)
	}

	// The graph is judged once every line has been read, because a
	// dependency may name a task that a later line of the same file
	// creates.
	for _, line := range lines {
		if e := checkBatchGraph(line.task, byID); e != nil {
			failures = append(failures, &batchFailure{line.number, e})
		}
	}

	if len(failures) > 0 {
		// The report goes in the order of the file, which the block of
		// docs/spec/cmd/new.md#el-modo-lote shows and this second pass
		// would otherwise break: a failure of the graph is found after
		// every line has been read, not while reading its own.
		sort.SliceStable(failures, func(i, j int) bool {
			return failures[i].line < failures[j].line
		})
		return nil, batchInvalid(failures, counted)
	}
	return lines, nil
}

// idTaken is the failure of an identifier that is not free, in its two
// shapes: the board already has it, or an earlier line of this same file
// took it. The second one does not mention the board, because the board has
// nothing to do with it (docs/spec/cmd/new.md#el-modo-lote).
func idTaken(id string, claimedBy map[string]int) *model.Error {
	message := fmt.Sprintf("id %q is already taken on this board", id)
	if line, ok := claimedBy[id]; ok {
		message = fmt.Sprintf("id %q is already taken by line %d of this file", id, line)
	}
	return &model.Error{
		ExitCode: 2,
		Code:     "id_taken",
		Message:  message,
		Field:    "id",
		Given:    id,
	}
}

// readBatchLine reads one line and judges everything that can be judged
// from the line and the board's configuration alone. It answers the first
// failure it finds: one line cannot be wrong in two ways at once as far as
// the report is concerned, and listing two failures of the same line would
// make the count of the message say something else.
func readBatchLine(cfg board.Config, line *batchLine, text string, now time.Time) *model.Error {
	d, err := decodeTask([]byte(text), now)
	if err != nil {
		if e, ok := err.(*model.Error); ok {
			return e
		}
		return &model.Error{ExitCode: 3, Code: "invalid_line", Message: err.Error()}
	}
	t := d.task
	line.task, line.dodKeys = t, d.dodKeys

	if t.ID != "" {
		if e := checkImportedID(cfg.TaskPrefix, t.ID); e != nil {
			return e
		}
	}
	// The status is the initial one when the line does not say, exactly as
	// a task created with no -s (docs/spec/cmd/new.md).
	t.Status = cfg.InitialStatus
	if d.hasStatus {
		status, matchErr := match.Match(match.Status, d.rawStatus, cfg.Statuses)
		if matchErr != nil {
			return matchErr.(*model.Error)
		}
		t.Status = status
	}
	if d.rawType != "" {
		value, matchErr := match.Match(match.Type, d.rawType, cfg.Types)
		if matchErr != nil {
			return matchErr.(*model.Error)
		}
		t.Type = value
	}
	if d.rawPriority != "" {
		value, matchErr := match.Match(match.Priority, d.rawPriority, cfg.Priorities)
		if matchErr != nil {
			return matchErr.(*model.Error)
		}
		t.Priority = value
	}
	if e := checkImportedLease(d, cfg.ActiveStatus); e != nil {
		return e
	}
	// The dates the line did not bring are the instant of the import, the
	// same ones a task created by hand gets
	// (docs/spec/modelo-de-datos/fechas.md).
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}
	if validateErr := t.Validate(cfg.Extensions); validateErr != nil {
		if e, ok := validateErr.(*model.Error); ok {
			return e
		}
		return &model.Error{ExitCode: 3, Code: "invalid_line", Message: validateErr.Error()}
	}
	return nil
}

// checkImportedID is the rule that an explicit id has to carry the task
// prefix of the board it is being imported into
// (docs/spec/cmd/new.md#el-modo-lote). It is the same protection that makes
// task_prefix immutable: a board of mixed identifiers has no way back.
func checkImportedID(prefix, id string) *model.Error {
	rest, ok := strings.CutPrefix(id, prefix+"-")
	if !ok {
		return &model.Error{
			ExitCode: 2,
			Code:     "malformed_id",
			Message: fmt.Sprintf("id %q does not match this board's task prefix %q",
				id, prefix),
			Field: "id",
			Given: id,
		}
	}
	if n, err := strconv.Atoi(rest); err != nil || n < 1 {
		return &model.Error{
			ExitCode: 2,
			Code:     "malformed_id",
			Message:  fmt.Sprintf("malformed task id: %q", id),
			Hints:    []string{"ids look like " + prefix + "-11: the prefix, a dash and a positive number"},
			Field:    "id",
			Given:    id,
		}
	}
	return nil
}

// checkImportedLease is the invariant of docs/spec/lease.md#el-vaciado
// checked at validation time, in its two halves, which is what makes the
// symmetry of `biso export` true for these two fields
// (docs/spec/cmd/new.md#el-modo-lote).
func checkImportedLease(d *decoded, activeStatus string) *model.Error {
	if !d.hasLeaseExpiresAt && !d.hasLeaseHolder {
		return nil
	}
	t := d.task
	field := "leaseExpiresAt"
	if d.hasLeaseHolder {
		field = "leaseHolder"
	}
	if t.Status != activeStatus || len(t.Assignees) == 0 {
		return &model.Error{
			ExitCode: 2,
			Code:     "invalid_lease",
			Message:  field + " on a task that is not both active and assigned",
			Field:    field,
		}
	}
	if d.hasLeaseExpiresAt != d.hasLeaseHolder {
		missing := "leaseExpiresAt"
		if d.hasLeaseExpiresAt {
			missing = "leaseHolder"
		}
		return &model.Error{
			ExitCode: 2,
			Code:     "invalid_lease",
			Message:  field + " given without " + missing + "; the two go together",
			Field:    field,
		}
	}
	return nil
}

// checkBatchGraph is what `biso new` checks about a reference as it writes
// it, applied to a line: a parent and a dependency name a task, so a name
// that no task of the board and no line of this file carries is a failure,
// and so is a cycle.
//
// A reference of a batch is always an identifier and never a text to search
// for: the file was written by a program and not typed, and resolving text
// there would let two lines of the same file mean different tasks depending
// on what the board happens to hold.
func checkBatchGraph(t *model.Task, byID map[string]*model.Task) *model.Error {
	for _, dep := range t.Dependencies {
		if dep == t.ID && t.ID != "" {
			return &model.Error{
				ExitCode: 2,
				Code:     "self_dependency",
				Message:  fmt.Sprintf("%s cannot depend on itself", t.ID),
				Field:    "dependencies",
				Given:    dep,
			}
		}
		if _, ok := byID[dep]; !ok {
			return notOnThisBoard("dependencies", dep)
		}
	}
	if t.Parent != "" {
		if t.Parent == t.ID && t.ID != "" {
			return &model.Error{
				ExitCode: 2,
				Code:     "parent_cycle",
				Message:  fmt.Sprintf("%s cannot be its own parent", t.ID),
				Field:    "parent",
				Given:    t.Parent,
			}
		}
		if _, ok := byID[t.Parent]; !ok {
			return notOnThisBoard("parent", t.Parent)
		}
	}
	if t.ID == "" {
		// A line with no id of its own cannot be in a cycle: nothing can
		// name it, because its identifier does not exist yet.
		return nil
	}
	for _, dep := range t.Dependencies {
		if path := reaches(byID, dep, t.ID, func(x *model.Task) []string {
			return x.Dependencies
		}); path != nil {
			return &model.Error{
				ExitCode: 2,
				Code:     "dependency_cycle",
				Message: "dependencies close a cycle: " +
					strings.Join(append([]string{t.ID}, path...), " -> "),
				Field: "dependencies",
				Given: dep,
			}
		}
	}
	if t.Parent == "" {
		return nil
	}
	if path := reaches(byID, t.Parent, t.ID, func(x *model.Task) []string {
		if x.Parent == "" {
			return nil
		}
		return []string{x.Parent}
	}); path != nil {
		return &model.Error{
			ExitCode: 2,
			Code:     "parent_cycle",
			Message: "parent closes a cycle: " +
				strings.Join(append([]string{t.ID}, path...), " -> "),
			Field: "parent",
			Given: t.Parent,
		}
	}
	return nil
}

func notOnThisBoard(field, id string) *model.Error {
	return &model.Error{
		ExitCode: 4,
		Code:     "not_found",
		Message:  fmt.Sprintf("%s names %q, which is not on this board and not in this file", field, id),
		Field:    field,
		Given:    id,
	}
}

// batchFailure is one failure while the file is still being judged: the
// error and the line it belongs to, kept apart so that the report can be
// ordered by line before the number is written into any message.
type batchFailure struct {
	line int
	err  *model.Error
}

// lineError stamps a failure with the line it came from. The number lives
// in the message and not in a key of its own, because the objects of
// error.details have the shape of any other error
// (docs/spec/contrato-json.md#los-errores-en-json) and nothing would carry
// it otherwise.
func lineError(number int, e *model.Error) *model.Error {
	stamped := *e
	stamped.Message = fmt.Sprintf("line %d: %s", number, e.Message)
	return &stamped
}

// batchInvalid is the code 7 of docs/spec/cmd/new.md#el-modo-lote, with
// every failure under it and never only the first.
func batchInvalid(failures []*batchFailure, counted int) *model.Error {
	detail := make([]string, 0, len(failures))
	details := make([]*model.Error, 0, len(failures))
	for _, f := range failures {
		stamped := lineError(f.line, f.err)
		details = append(details, stamped)
		line := "  " + stamped.Message
		if len(stamped.Valid) > 0 {
			line += " (valid: " + strings.Join(stamped.Valid, ", ") + ")"
		}
		detail = append(detail, line)
	}
	return &model.Error{
		ExitCode: 7,
		Code:     "batch_invalid",
		Message: fmt.Sprintf("%d of %s are invalid, nothing was written",
			len(failures), plural(counted, "line")),
		Detail:  detail,
		Details: details,
	}
}
