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
	// source task's own number (docs/especificacion.md, "Identificadores",
	// point 2) or a reassigned one (point 4).
	ID string
	// Parent is Task.ParentTaskID rewritten through the equivalence table
	// by an exact match, not the mention pattern (docs/especificacion.md,
	// "Identificadores", point 7), or empty when the task had no parent, or
	// its parent named an id outside the whole source batch (dropped, with
	// a Finding).
	Parent string
	// Dependencies is Task.Dependencies rewritten through the equivalence
	// table the same way as Parent, in the same order, with any element
	// naming an id outside the batch removed (with a Finding) rather than
	// the whole line failing.
	Dependencies []string

	// Title, Description, Plan, Notes, and Summary are the source task's
	// own text fields (never Result's, which does not carry them), each
	// with every mention of a source id rewritten to its final id
	// (docs/especificacion.md, "Identificadores", point 5). This is the
	// DEFINITIVE mention rewrite, built from the equivalence table once it
	// is complete; it is unrelated to naiveTitle, the separate, simpler
	// rewrite point 3 uses only to decide whether a task is already on the
	// destination.
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
	// 4a) item for item, in the same order: Author and CreatedAt untouched
	// (point 5 explicitly excludes a comment's author from mention
	// rewriting), Body with every mention rewritten.
	Comments []Comment

	// Labels is Result.Labels (already including milestone:: and
	// project:: from phase 4a) with backlog.id::<source-id> appended when
	// the source id has a dot (docs/especificacion.md, "Identificadores",
	// point 9). A source label colliding with the backlog.id key is
	// dropped first, with a Finding, exactly the way phase 4a already
	// drops one colliding with milestone or project
	// (RemoveCollidingScopedLabel, reused here unchanged). A task whose
	// source id has no dot never gets this label at all.
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
// "with or without leading zeros" (docs/especificacion.md, "Identificadores",
// point 1) never matters again after parsing: 001 and 1 parse to the same
// int and are treated as the same number everywhere in this file except the
// literal id strings used in Finding messages and as equivalents map keys,
// which keep whatever padding the source actually used.
type parsedSourceID struct {
	prefix string
	number int
	hasSub bool
	sub    int
}

// parseSourceID parses id against sourceIDPattern. ok is false for any
// other shape, which validateSourceShape turns into the code-3 error
// docs/especificacion.md, "Identificadores", point 1, describes.
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

// canonicalIDKey returns the canonical form docs/especificacion.md,
// "Identificadores", point 1's last paragraph, and docs/decisiones.md, "El
// mismo número, no la misma cadena, también decide cuándo dos ids de origen
// chocan entre sí", define for deciding whether two source ids are the
// exact same id: the prefix folded to uppercase, the main number as an int
// (ignoring leading zeros), and, only when the id has a dot, the subtask
// number as an int too (also ignoring leading zeros).
//
// A simple id and a subtask id are never the same canonical key even when
// their main number matches, because only a subtask's key has a dot at
// all: "TASK-1" canonicalizes to "TASK-1" and "TASK-1.2" canonicalizes to
// "TASK-1.2", two different strings, while "TASK-1" and "TASK-001" both
// canonicalize to "TASK-1", and "TASK-1.2" and "TASK-1.02" both
// canonicalize to "TASK-1.2".
func canonicalIDKey(prefixUpper string, p parsedSourceID) string {
	if p.hasSub {
		return fmt.Sprintf("%s-%d.%d", prefixUpper, p.number, p.sub)
	}
	return fmt.Sprintf("%s-%d", prefixUpper, p.number)
}

// seenSourceID records, for validateSourceShape's duplicate check, which
// file and which literal id string first produced a given canonical key, so
// the error message can name both tasks by file and by their own original
// id, not just by the canonical key they collide on.
type seenSourceID struct {
	file string
	id   string
}

// validateSourceShape implements docs/especificacion.md, "Identificadores",
// point 1, over ALL of the batch's source tasks at once (active, completed,
// and archived tasks together): every id must have the shape parseSourceID
// accepts, every id must share the same prefix once folded to uppercase
// (Backlog.md always writes the prefix in uppercase in the id itself, even
// when its own configuration stores it differently, so no configuration
// needs to be read here), and no two tasks may canonicalize to the same id
// (canonicalIDKey): same prefix folded, same main number, the same shape
// (both simple or both subtask, since a simple id and a subtask id with the
// same main number are never the same id), and, only when both are
// subtasks, the same subtask number too.
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
	seenIn := make(map[string]seenSourceID, len(tasks))

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

		canonical := canonicalIDKey(folded, p)
		if prev, duplicate := seenIn[canonical]; duplicate {
			return nil, "", fmt.Errorf(
				"identifiers: %s (id %q) and %s (id %q) are the same id",
				prev.file, prev.id, t.File, t.ID,
			)
		}
		seenIn[canonical] = seenSourceID{file: t.File, id: t.ID}

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
//     destination (point 3), in the same relative order as batch.
//  2. Every Finding this phase raises: a skip (point 3), a reassignment
//     (point 4 and point 8's exact wording), an unresolved or wrong-case
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

	// Built once and reused for both naiveTitle (the "already on the
	// destination" comparison below) and the definitive rewriteMentions
	// pass later in this function: the two need the exact same candidate
	// pattern, only what they do with a candidate differs.
	pattern := mentionPattern(prefixUpper)

	var findings []source.Finding

	// docs/especificacion.md, "Identificadores", point 3: a task is a
	// repeat of an earlier import when an existing destination task has the
	// same title and the same already-converted createdAt. The title
	// compared here is NOT the source task's raw title: point 3 requires
	// the naive mention rewrite naiveTitle computes (see its own comment),
	// so that a title mentioning another source id still matches the title
	// biso actually wrote on a previous run. The equivalent of a skipped
	// task is the id that destination task ALREADY has, and it is skipped
	// rather than emitted.
	equivalents := make(map[string]string, len(batch))
	skipped := make(map[string]bool, len(batch))
	for _, b := range batch {
		candidateTitle := naiveTitle(b.Task.Title, pattern, prefixUpper, board.Config.TaskPrefix)
		for _, dt := range board.Tasks {
			if candidateTitle == dt.Title && b.Result.CreatedAt == dt.CreatedAt {
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

	// docs/especificacion.md, "Identificadores", point 4: a simple id whose
	// number is already taken on the destination, and every subtask id, are
	// reassigned. Everything else, a simple id whose number is free, keeps
	// that number (point 2).
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
	// Point 4's ceiling is the greater of the destination's own maximum
	// number and the maximum MAIN number among every source task in the
	// batch, skipped tasks included, not just the ones this phase ends up
	// reassigning: parsedByID already has one entry per task validated
	// above, skipped tasks included, so ranging over it covers that without
	// a second pass over batch.
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

	// Point 4 assigns reassigned numbers in the source id's own natural
	// order: main number ascending, and, when two entries share a main
	// number, subtask number ascending. Two entries only ever share a main
	// number when both are subtasks of the same parent: a simple id that
	// tied on number would already have kept it above, so it never reaches
	// this list. Sorting by (number, sub) therefore covers both a batch of
	// only simple ids and one mixing in subtasks, regardless of the order
	// they arrived in batch.
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

	// docs/especificacion.md, "Identificadores", points 5, 6, 7, and 9:
	// build the actual output for every task that was not skipped, in
	// batch order, now that equivalents (built by points 2, 3, and 4
	// together) is complete.
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

// mentionFieldOrder fixes the order docs/especificacion.md,
// "Identificadores", point 6's grouped findings come out in, one per field
// named in point 5, so two runs over the same batch produce the same
// Finding order.
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

// mentionPattern builds the candidate regex docs/especificacion.md,
// "Identificadores", points 5 and 6, both need: prefixUpper, followed by
// '-', one or more digits, and optionally '.' and one or more digits. The
// prefix half is matched case-insensitively on purpose: that is what lets
// scanMentions's callers tell apart, at each match, a valid mention (exact
// uppercase prefix, point 5) from one that has the right shape but the
// wrong case (point 6's "Xyz-002", "task-12" example), which must still be
// found even though it is never substituted. Point 5's own case-sensitive
// matching is enforced afterward, by comparing the matched text's prefix
// against prefixUpper byte for byte (mentionMatch.ExactCase).
func mentionPattern(prefixUpper string) *regexp.Regexp {
	return regexp.MustCompile(`(?i:` + regexp.QuoteMeta(prefixUpper) + `)-[0-9]+(?:\.[0-9]+)?`)
}

// mentionMatch is one candidate substring scanMentions found: a run shaped
// like a source id mention that also satisfies docs/especificacion.md,
// "Identificadores", point 5's word-boundary rules. ExactCase is true when
// its prefix matches prefixUpper byte for byte, the only shape point 5
// accepts as a real mention; a candidate whose prefix differs only in case
// (point 6's "Xyz-002"/"task-12") still reaches a scanMentions caller's
// handle function, with ExactCase false, so a caller that needs to notice
// it (rewriteMentions) can, and one that does not (naiveTitle) can just
// return it unchanged.
type mentionMatch struct {
	Text      string
	ExactCase bool
}

// scanMentions is the single left-to-right scan every mention-rewriting
// function in this file is built on. It finds every substring of text
// shaped like a source id mention (pattern), drops the ones that fail
// docs/especificacion.md, "Identificadores", point 5's word-boundary rules
// (mentionBoundaryBefore/mentionBoundaryAfter) without ever calling handle
// for them (left exactly as written, the protection point 5 describes for
// "SUBTASK-12" and "TASK-10-modelo"), and lets handle decide the
// replacement text for every candidate that passes.
//
// The result is rebuilt from pieces of the ORIGINAL text plus handle's
// replacements, in a single pass: handle's own output is never fed back
// through pattern. That is what point 5 means by "todas las menciones se
// sustituyen en una sola pasada, de modo que un número reasignado no vuelve
// a sustituirse", and it is what makes rewriteMentions and naiveTitle share
// this one scan instead of each reimplementing it.
func scanMentions(text string, pattern *regexp.Regexp, prefixUpper string, handle func(mentionMatch) string) string {
	if text == "" {
		return text
	}

	matches := pattern.FindAllStringIndex(text, -1)
	if matches == nil {
		return text
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
		b.WriteString(handle(mentionMatch{
			Text:      mention,
			ExactCase: strings.HasPrefix(mention, prefixUpper),
		}))
		last = end
	}
	b.WriteString(text[last:])

	return b.String()
}

// rewriteMentions implements docs/especificacion.md, "Identificadores",
// points 5 and 6 together, for one text field: every candidate scanMentions
// finds whose case exactly matches prefixUpper and is a key of equivalents
// is substituted with its final id (point 5); every other candidate,
// whether its case matches but it names no task in the batch, or its case
// does not match at all (point 6's "Xyz-002"/"task-12"), is left untouched
// and counted in unresolved, since both share the same
// grouped-by-file-and-field Finding identifyTask raises from that count.
func rewriteMentions(text string, pattern *regexp.Regexp, prefixUpper string, equivalents map[string]string) (rewritten string, unresolved int) {
	rewritten = scanMentions(text, pattern, prefixUpper, func(m mentionMatch) string {
		if m.ExactCase {
			if final, ok := equivalents[m.Text]; ok {
				return final
			}
		}
		unresolved++
		return m.Text
	})
	return rewritten, unresolved
}

// naiveTitle implements the "naive" mention rewrite docs/especificacion.md,
// "Identificadores", point 3, and docs/decisiones.md, "Los identificadores
// conservan su número y cambian de prefijo" (its paragraphs on "La
// comparación de 'ya está en el destino' usa el título ya reescrito, no el
// crudo"), define for exactly one purpose: deciding whether a source task
// is already on the destination, BEFORE this phase knows the equivalence
// table. That table needs to know which tasks are skipped before it can be
// completed, so this comparison cannot depend on it without becoming
// circular; naiveTitle breaks that circularity by computing something
// simpler on its own.
//
// It substitutes every candidate scanMentions finds whose case exactly
// matches the source prefix with the destination's own prefix and the SAME
// number the mention names, dropping any subtask suffix (a biso id is
// always simple, so a mention of "TASK-1.2" naively becomes "BISO-1"),
// regardless of whether that number collides with anything on the
// destination or the mentioned task ends up reassigned in this same batch.
// A candidate whose case does not exactly match the source prefix is left
// untouched, the same as rewriteMentions leaves it.
//
// The result is used ONLY for the "already on the destination" comparison
// and is discarded right after: the title actually written to the
// destination, for a task that turns out not to be a duplicate, always
// comes from rewriteMentions once the equivalence table is complete.
// Reusing this naive result as a task's final title would be wrong even
// for a task that is not skipped, since it never accounts for a collision
// or a reassignment at all.
func naiveTitle(title string, pattern *regexp.Regexp, prefixUpper, destinationPrefix string) string {
	return scanMentions(title, pattern, prefixUpper, func(m mentionMatch) string {
		if !m.ExactCase {
			return m.Text
		}

		numeric := m.Text[len(prefixUpper)+1:]
		if dot := strings.IndexByte(numeric, '.'); dot >= 0 {
			numeric = numeric[:dot]
		}
		n, err := strconv.Atoi(numeric)
		if err != nil {
			return m.Text
		}

		return fmt.Sprintf("%s-%d", destinationPrefix, n)
	})
}

// mentionBoundaryBefore reports whether text is not preceded, right at
// start, by a letter, a digit, or '_' or '-' (docs/especificacion.md,
// "Identificadores", point 5's first exclusion). The very start of the text
// always satisfies it.
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
// by a letter, a digit, '_', or by '-' then a letter or a digit
// (docs/especificacion.md, "Identificadores", point 5's second exclusion,
// the one protecting "TASK-10-modelo": after "TASK-10" comes "-m", and 'm'
// is a letter). The very end of the text always satisfies it, and a
// trailing '-' not itself followed by a letter or digit (end of text, or
// another symbol) also satisfies it.
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
