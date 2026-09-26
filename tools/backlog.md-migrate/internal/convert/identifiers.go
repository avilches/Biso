package convert

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

// This is phase 4b of the conversion engine (docs/especificacion.md,
// "Identificadores"). Unlike phase 4a (convert.go), it looks at the whole
// batch and at the destination board at once, because an identifier can
// only be decided knowing every source task and every task already on the
// destination: which numbers are free, which task is a repeat of an earlier
// import, and which id every mention, parent, and dependency in the batch
// resolves to.

// TaskInput pairs one source.Task with the Result phase 4a already computed
// for it (Task, in convert.go). Identifiers needs both together for every
// task of the batch: it reads Result.CreatedAt (already converted to UTC)
// to decide whether a task is already on the destination, and it rewrites
// mentions inside Result.AcceptanceCriteria and Result.Comments (already
// merged with Definition of Done, and already date-converted) alongside the
// source task's own free text fields, which Result does not carry at all.
type TaskInput struct {
	Task   source.Task
	Result Result
}

// Identified is phase 4b's output for one source task that was NOT skipped
// as already on the destination: every piece that needed the whole batch
// and the destination board to compute. A later phase (4c/5) joins this
// with the matching TaskInput.Result by SourceID to build one line of
// NDJSON. Every field Result already owns and this phase never touches
// (Status, Type, Priority, Assignees, Due, CreatedAt, UpdatedAt) is
// deliberately absent here, so the two are combined rather than duplicated.
type Identified struct {
	// SourceID is the source.Task.ID this Identified was computed from,
	// exactly as read (its original prefix and zero-padding, if any): the
	// join key a later phase uses to find the matching Result.
	SourceID string

	// ID is the final id: the destination's task_prefix plus either the
	// source task's own number (docs task point 2) or a reassigned one
	// (docs task point 3).
	ID string
	// Parent is Task.ParentTaskID rewritten through the equivalence table
	// (docs task point 6), or empty when the task had no parent, or its
	// parent named an id outside the whole source batch (dropped, with a
	// Finding).
	Parent string
	// Dependencies is Task.Dependencies rewritten through the equivalence
	// table, in the same order, with any element naming an id outside the
	// batch removed (with a Finding) rather than the whole line failing.
	Dependencies []string

	// Title, Description, Plan, Notes, and Summary are the source task's
	// own text fields (never Result's, which does not carry them), each
	// with every mention of a source id rewritten to its final id (docs
	// task point 5).
	Title       string
	Description string
	Plan        string
	Notes       string
	Summary     string

	// AcceptanceCriteria mirrors Result.AcceptanceCriteria (already merged
	// with Definition of Done by phase 4a) item for item, in the same
	// order: Number and Checked untouched, Text with every mention
	// rewritten.
	AcceptanceCriteria []source.Checkbox

	// Comments mirrors Result.Comments (already date-converted by phase
	// 4a) item for item, in the same order: Author and CreatedAt
	// untouched (docs task point 5 explicitly excludes a comment's
	// author from mention rewriting), Body with every mention rewritten.
	Comments []Comment

	// Labels is Result.Labels (already including milestone:: and
	// project:: from phase 4a) with backlog.id::<source-id> appended when
	// the source id has a dot (docs task point 9). A source label
	// colliding with the backlog.id key is dropped first, with a Finding,
	// exactly the way phase 4a already drops one colliding with milestone
	// or project (RemoveCollidingScopedLabel, reused here unchanged). A
	// task whose source id has no dot never gets this label at all.
	Labels []string
}

// sourceIDPattern matches the two shapes docs/especificacion.md,
// "Identificadores", point 1 allows for a source id: "<PREFIX>-<n>" or, for
// a subtask, "<PREFIX>-<n>.<m>", with or without leading zeros on n or m.
// The prefix is one or more Unicode letters: point 1 never restricts it to
// ASCII.
var sourceIDPattern = regexp.MustCompile(`^([\p{L}]+)-([0-9]+)(?:\.([0-9]+))?$`)

// parsedSourceID is one source id split into the pieces this file compares
// and sorts by: the prefix as written (any case, only used to check it
// against the batch's shared prefix), the main number, and the subtask
// number when the id has a dot. Numbers are ints, not strings, precisely so
// "with or without leading zeros" (docs task point 1) never matters again
// after parsing: 001 and 1 parse to the same int and are treated as the
// same number everywhere in this file except the literal id strings used in
// Finding messages and as equivalents map keys, which keep whatever
// padding the source actually used.
type parsedSourceID struct {
	prefix string
	number int
	hasSub bool
	sub    int
}

// parseSourceID parses id against sourceIDPattern. ok is false for any
// other shape, which validateSourceShape turns into the code-3 error docs
// task point 1 describes.
func parseSourceID(id string) (parsed parsedSourceID, ok bool) {
	m := sourceIDPattern.FindStringSubmatch(id)
	if m == nil {
		return parsedSourceID{}, false
	}
	number, err := strconv.Atoi(m[2])
	if err != nil {
		return parsedSourceID{}, false
	}
	parsed = parsedSourceID{prefix: m[1], number: number}
	if m[3] != "" {
		sub, err := strconv.Atoi(m[3])
		if err != nil {
			return parsedSourceID{}, false
		}
		parsed.hasSub = true
		parsed.sub = sub
	}
	return parsed, true
}

// validateSourceShape implements docs/especificacion.md, "Identificadores",
// point 1, over ALL of the batch's source tasks at once (active, completed,
// and archived together, exactly as the encargo requires): every id must
// have the shape parseSourceID accepts, every id must share the same
// prefix once folded to uppercase (Backlog.md always writes the prefix in
// uppercase in the id itself, even when its own configuration stores it
// differently, so no configuration needs to be read here), and no two
// tasks may share the exact same id string.
//
// It returns a parsedSourceID per task, keyed by the task's literal id, and
// the shared prefix already folded to uppercase: the form every later step
// in this file writes into a final id and matches text mentions against.
//
// This is the one failure "Identificadores" turns into a Go error rather
// than a Finding: docs/especificacion.md, "Códigos de salida", code 3 ("El
// origen no se puede leer"), which aborts the whole batch instead of
// skipping one task. Translating this error into that exit code is a later
// phase's job; this function only needs to say clearly, in its error's
// message, which of the three things failed and with what value.
func validateSourceShape(tasks []source.Task) (parsed map[string]parsedSourceID, prefixUpper string, err error) {
	parsed = make(map[string]parsedSourceID, len(tasks))
	seenIn := make(map[string]string, len(tasks))

	for _, t := range tasks {
		p, ok := parseSourceID(t.ID)
		if !ok {
			return nil, "", fmt.Errorf(
				"identifiers: %s has id %q, which is not <PREFIX>-<n> or <PREFIX>-<n>.<m>",
				t.File, t.ID,
			)
		}

		folded := strings.ToUpper(p.prefix)
		if prefixUpper == "" {
			prefixUpper = folded
		} else if folded != prefixUpper {
			return nil, "", fmt.Errorf(
				"identifiers: source ids do not share one prefix: %s has id %q (prefix %q), but the batch already has prefix %q",
				t.File, t.ID, folded, prefixUpper,
			)
		}

		if otherFile, duplicate := seenIn[t.ID]; duplicate {
			return nil, "", fmt.Errorf(
				"identifiers: %s and %s share the exact same id %q",
				otherFile, t.File, t.ID,
			)
		}
		seenIn[t.ID] = t.File

		parsed[t.ID] = p
	}

	return parsed, prefixUpper, nil
}

// destinationNumber extracts the number of an existing destination id
// ("<task_prefix>-<n>"): the part after the last '-'. The destination's
// task_prefix is already known from board.Config, so this never needs to
// re-derive it, only to strip it off.
func destinationNumber(id string) (int, bool) {
	idx := strings.LastIndexByte(id, '-')
	if idx < 0 {
		return 0, false
	}
	n, err := strconv.Atoi(id[idx+1:])
	if err != nil {
		return 0, false
	}
	return n, true
}

// Identifiers is phase 4b's entry point (docs/especificacion.md,
// "Identificadores"). It receives the whole batch at once, because an
// identifier can only be decided knowing every source task and every task
// already on the destination:
//
//   - batch holds one TaskInput per source.Task read from tasks/,
//     completed/, and archive/tasks/, all of them together.
//   - board is the destination.Board a prior destination.Read already
//     fetched: its Config.TaskPrefix and its existing Tasks.
//
// It returns, in this order:
//
//  1. One Identified per TaskInput that was NOT skipped as already on the
//     destination (docs task point 3), in the same relative order as batch.
//  2. Every Finding this phase raises: a skip (point 3), a reassignment
//     (point 3 and point 8's exact wording), an unresolved or wrong-case
//     mention grouped by file and field (point 6), a dropped parent or
//     dependency (point 7), and a dropped label colliding with backlog.id
//     (point 9).
//  3. A non-nil error, instead of any Identified or Finding, when
//     validateSourceShape's point-1 validation fails. Callers must check
//     this error before looking at anything else this function returns.
func Identifiers(batch []TaskInput, board destination.Board) ([]Identified, []source.Finding, error) {
	tasks := make([]source.Task, len(batch))
	for i, b := range batch {
		tasks[i] = b.Task
	}

	parsedByID, prefixUpper, err := validateSourceShape(tasks)
	if err != nil {
		return nil, nil, err
	}

	var findings []source.Finding

	// Docs task point 2: a task whose title and already-converted
	// createdAt exactly match an existing destination task is a repeat of
	// an earlier import. Its equivalent is the id that destination task
	// ALREADY has, and it is skipped rather than emitted.
	equivalents := make(map[string]string, len(batch))
	skipped := make(map[string]bool, len(batch))
	for _, b := range batch {
		for _, dt := range board.Tasks {
			if b.Task.Title == dt.Title && b.Result.CreatedAt == dt.CreatedAt {
				equivalents[b.Task.ID] = dt.ID
				skipped[b.Task.ID] = true
				findings = append(findings, source.Finding{
					File:    b.Task.File,
					Field:   "id",
					Message: "already on the destination, skipped",
				})
				break
			}
		}
	}

	// Docs task point 3: a simple id whose number is already taken on the
	// destination, and every subtask id, are reassigned. Everything else
	// (a simple id whose number is free) keeps its number.
	destNumbers := make(map[int]bool, len(board.Tasks))
	maxDestNumber := 0
	for _, dt := range board.Tasks {
		if n, ok := destinationNumber(dt.ID); ok {
			destNumbers[n] = true
			if n > maxDestNumber {
				maxDestNumber = n
			}
		}
	}
	// "mayor número principal entre TODAS las tareas de origen del lote,
	// saltadas incluidas": parsedByID has one entry per task validated
	// above, skipped tasks included, so ranging over it already covers
	// that requirement without a second pass over batch.
	maxSourceNumber := 0
	for _, p := range parsedByID {
		if p.number > maxSourceNumber {
			maxSourceNumber = p.number
		}
	}

	type reassignEntry struct {
		task   source.Task
		parsed parsedSourceID
	}
	var toReassign []reassignEntry

	for _, b := range batch {
		if skipped[b.Task.ID] {
			continue
		}
		p := parsedByID[b.Task.ID]
		if !p.hasSub && !destNumbers[p.number] {
			equivalents[b.Task.ID] = fmt.Sprintf("%s-%d", board.Config.TaskPrefix, p.number)
			continue
		}
		toReassign = append(toReassign, reassignEntry{task: b.Task, parsed: p})
	}

	// "en el orden natural de su id de origen (el número principal
	// ascendente, y si dos comparten numero principal ... el numero de
	// subtarea ascendente)": ties in the main number only happen between
	// subtasks of the same parent, per the encargo, so sorting by (number,
	// sub) covers both simple ids (which never tie on number here, a tie
	// would have meant one of them kept its number above) and subtasks.
	sort.SliceStable(toReassign, func(i, j int) bool {
		a, b := toReassign[i].parsed, toReassign[j].parsed
		if a.number != b.number {
			return a.number < b.number
		}
		return a.sub < b.sub
	})

	nextNumber := max(maxDestNumber, maxSourceNumber) + 1
	for _, r := range toReassign {
		newID := fmt.Sprintf("%s-%d", board.Config.TaskPrefix, nextNumber)
		equivalents[r.task.ID] = newID

		if r.parsed.hasSub {
			findings = append(findings, source.Finding{
				File:  r.task.File,
				Field: "id",
				Message: fmt.Sprintf(
					"%s: id %s assigned (subtask ids have no equivalent)",
					r.task.ID, newID,
				),
			})
		} else {
			wouldBeID := fmt.Sprintf("%s-%d", board.Config.TaskPrefix, r.parsed.number)
			findings = append(findings, source.Finding{
				File:  r.task.File,
				Field: "id",
				Message: fmt.Sprintf(
					"%s: id %s is taken on the destination, reassigned to %s",
					r.task.ID, wouldBeID, newID,
				),
			})
		}

		nextNumber++
	}

	// Docs task points 5, 6, 7, and 9: build the actual output for every
	// task that was not skipped, in batch order, now that equivalents (the
	// point-4 table) is complete.
	pattern := mentionPattern(prefixUpper)
	var out []Identified
	for _, b := range batch {
		if skipped[b.Task.ID] {
			continue
		}
		identified, taskFindings := identifyTask(b, parsedByID[b.Task.ID], equivalents, pattern, prefixUpper)
		findings = append(findings, taskFindings...)
		out = append(out, identified)
	}

	return out, findings, nil
}

// mentionFieldOrder fixes the order docs task point 6's grouped findings
// come out in, one per field named in docs task point 5, so two runs over
// the same batch produce the same Finding order.
var mentionFieldOrder = []string{
	"title", "description", "plan", "notes", "summary",
	"acceptanceCriteria", "comments",
}

// identifyTask builds the Identified for one non-skipped task, plus every
// Finding raised while rewriting its parent, dependencies, mentions, and
// backlog.id:: label. p is parsedByID[b.Task.ID], passed in rather than
// re-parsed since the caller already has it.
func identifyTask(
	b TaskInput,
	p parsedSourceID,
	equivalents map[string]string,
	pattern *regexp.Regexp,
	prefixUpper string,
) (Identified, []source.Finding) {
	var findings []source.Finding
	unresolvedCounts := make(map[string]int, len(mentionFieldOrder))

	rewrite := func(field, text string) string {
		rewritten, unresolved := rewriteMentions(text, pattern, prefixUpper, equivalents)
		unresolvedCounts[field] += unresolved
		return rewritten
	}

	title := rewrite("title", b.Task.Title)
	description := rewrite("description", b.Task.Description)
	plan := rewrite("plan", b.Task.Plan)
	notes := rewrite("notes", b.Task.Notes)
	summary := rewrite("summary", b.Task.Summary)

	criteria := make([]source.Checkbox, len(b.Result.AcceptanceCriteria))
	for i, c := range b.Result.AcceptanceCriteria {
		c.Text = rewrite("acceptanceCriteria", c.Text)
		criteria[i] = c
	}

	comments := make([]Comment, len(b.Result.Comments))
	for i, c := range b.Result.Comments {
		c.Body = rewrite("comments", c.Body)
		comments[i] = c
	}

	for _, field := range mentionFieldOrder {
		if n := unresolvedCounts[field]; n > 0 {
			findings = append(findings, source.Finding{
				File:  b.Task.File,
				Field: field,
				Message: fmt.Sprintf(
					"%d mention(s) with the shape of a source id do not resolve to any source task "+
						"(unknown id, or the right shape with the wrong case), left unchanged",
					n,
				),
			})
		}
	}

	parent := ""
	if b.Task.ParentTaskID != "" {
		if resolved, ok := equivalents[b.Task.ParentTaskID]; ok {
			parent = resolved
		} else {
			findings = append(findings, source.Finding{
				File:  b.Task.File,
				Field: "parent",
				Message: fmt.Sprintf(
					"parent %q does not name any task in the source batch, dropped",
					b.Task.ParentTaskID,
				),
			})
		}
	}

	var dependencies []string
	for _, dep := range b.Task.Dependencies {
		if resolved, ok := equivalents[dep]; ok {
			dependencies = append(dependencies, resolved)
			continue
		}
		findings = append(findings, source.Finding{
			File:  b.Task.File,
			Field: "dependencies",
			Message: fmt.Sprintf(
				"dependency %q does not name any task in the source batch, dropped",
				dep,
			),
		})
	}

	labels := append([]string(nil), b.Result.Labels...)
	if p.hasSub {
		kept, labelFindings := RemoveCollidingScopedLabel(b.Task.File, labels, "backlog.id")
		labels = kept
		findings = append(findings, labelFindings...)
		labels = append(labels, "backlog.id::"+b.Task.ID)
	}

	return Identified{
		SourceID:           b.Task.ID,
		ID:                 equivalents[b.Task.ID],
		Parent:             parent,
		Dependencies:       dependencies,
		Title:              title,
		Description:        description,
		Plan:               plan,
		Notes:              notes,
		Summary:            summary,
		AcceptanceCriteria: criteria,
		Comments:           comments,
		Labels:             labels,
	}, findings
}

// mentionPattern builds the candidate regex docs task points 5 and 6 both
// need: prefixUpper, followed by '-', one or more digits, and optionally
// '.' and one or more digits. The prefix half is matched case-insensitively
// on purpose: that is what lets rewriteMentions tell apart, at each match,
// a valid mention (exact uppercase prefix, docs task point 5) from one that
// has the right shape but the wrong case (docs task point 6's "Xyz-002",
// "task-12" example), which must still be found and counted even though it
// is never substituted. Point 5's own case-sensitive matching is enforced
// afterward, in rewriteMentions, by comparing the matched text's prefix
// against prefixUpper byte for byte.
func mentionPattern(prefixUpper string) *regexp.Regexp {
	return regexp.MustCompile(`(?i:` + regexp.QuoteMeta(prefixUpper) + `)-[0-9]+(?:\.[0-9]+)?`)
}

// rewriteMentions scans text once for every substring shaped like a source
// id and, for each one that also satisfies docs task point 5's
// word-boundary rules, either substitutes it or leaves it and counts it:
//
//   - Not a candidate at all: pattern found nothing there, or the match is
//     preceded by a letter, a digit, '_', or '-' (mentionBoundaryBefore),
//     or followed by a letter, a digit, '_', or by '-' then a letter or a
//     digit (mentionBoundaryAfter). Left untouched, no Finding, exactly the
//     protection docs task point 5 describes for "SUBTASK-12" and
//     "TASK-10-modelo".
//   - A candidate whose case exactly matches prefixUpper (docs task point
//     5's only valid mention shape) and is a key of equivalents: replaced
//     with its final id.
//   - Any other candidate reaching this far, whether its case matches
//     prefixUpper but it is not in equivalents (an id that names no task in
//     the batch), or its case does not exactly match prefixUpper at all
//     (docs task point 6's "Xyz-002"/"task-12"): left untouched, but
//     counted in the returned unresolved count, since both cases share the
//     same grouped-by-file-and-field Finding.
//
// Every replacement is written from the ORIGINAL text: matches come from a
// single FindAllStringIndex pass taken before any substitution happens, so
// an id reassigned to a number that collides by coincidence with another
// source id in the batch is never substituted a second time.
func rewriteMentions(text string, pattern *regexp.Regexp, prefixUpper string, equivalents map[string]string) (rewritten string, unresolved int) {
	if text == "" {
		return text, 0
	}

	matches := pattern.FindAllStringIndex(text, -1)
	if matches == nil {
		return text, 0
	}

	var b strings.Builder
	last := 0
	for _, m := range matches {
		start, end := m[0], m[1]
		if !mentionBoundaryBefore(text, start) || !mentionBoundaryAfter(text, end) {
			continue
		}

		mention := text[start:end]
		b.WriteString(text[last:start])

		if strings.HasPrefix(mention, prefixUpper) {
			if final, ok := equivalents[mention]; ok {
				b.WriteString(final)
				last = end
				continue
			}
		}

		b.WriteString(mention)
		unresolved++
		last = end
	}
	b.WriteString(text[last:])

	return b.String(), unresolved
}

// mentionBoundaryBefore reports whether text is not preceded, right at
// start, by a letter, a digit, or '_' or '-' (docs task point 5's first
// exclusion). The very start of the text always satisfies it.
func mentionBoundaryBefore(text string, start int) bool {
	if start == 0 {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(text[:start])
	if r == '_' || r == '-' {
		return false
	}
	return !isLetterOrDigit(r)
}

// mentionBoundaryAfter reports whether text is not followed, right at end,
// by a letter, a digit, '_', or by '-' then a letter or a digit (docs task
// point 5's second exclusion, the one protecting "TASK-10-modelo": after
// "TASK-10" comes "-m", and 'm' is a letter). The very end of the text
// always satisfies it, and a trailing '-' not itself followed by a letter
// or digit (end of text, or another symbol) also satisfies it.
func mentionBoundaryAfter(text string, end int) bool {
	if end >= len(text) {
		return true
	}
	r, size := utf8.DecodeRuneInString(text[end:])
	if r == '_' || isLetterOrDigit(r) {
		return false
	}
	if r == '-' {
		next := end + size
		if next < len(text) {
			r2, _ := utf8.DecodeRuneInString(text[next:])
			if isLetterOrDigit(r2) {
				return false
			}
		}
	}
	return true
}
