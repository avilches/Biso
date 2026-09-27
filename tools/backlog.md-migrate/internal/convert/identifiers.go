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

// This file implements identifier resolution (docs/especificacion.md,
// "Identificadores"). Unlike convert.go's per-task conversion, it looks at
// the whole batch and at the destination board at once, because an
// identifier can only be decided knowing every source task and every task
// already on the destination: which numbers are free, which task is a
// repeat of an earlier import, and which id every mention, parent, and
// dependency in the batch resolves to.

// TaskInput pairs one source.Task with the Result convert.Task already
// computed for it. Identifiers needs both together for every
// task of the batch: it reads Result.CreatedAt (already converted to UTC)
// to decide whether a task is already on the destination, and it rewrites
// mentions inside Result.AcceptanceCriteria and Result.Comments (already
// merged with Definition of Done, and already date-converted) alongside the
// source task's own free text fields, which Result does not carry at all.
type TaskInput struct {
	Task   source.Task
	Result Result
}

// Identified is identifier resolution's output for one source task that was
// NOT skipped as already on the destination: every piece that needed the
// whole batch and the destination board to compute. Assemble (batch.go)
// joins this with the matching TaskInput.Result by SourceID to build one
// line of NDJSON. Every field Result already owns and identifier resolution
// never touches (Status, Type, Priority, Assignees, Due, CreatedAt,
// UpdatedAt) is deliberately absent here, so the two are combined rather
// than duplicated.
type Identified struct {
	// SourceID is the source.Task.ID this Identified was computed from,
	// exactly as read (its original prefix and zero-padding, if any). It is
	// NOT a unique key on its own: docs/especificacion.md, "Identificadores",
	// point 1's id-reuse-after-archiving case means two different
	// Identified values can carry the exact same SourceID (Backlog.md hands
	// the reused task the very same id string as the one it archived). A
	// caller matching an Identified back to the TaskInput it came from must
	// use BatchIndex instead.
	SourceID string

	// BatchIndex is the position, in the batch slice passed to Identifiers,
	// of the TaskInput this Identified was computed from. Unlike SourceID,
	// it is always unique: it is the real position of one particular
	// source.Task in the batch, so a caller (Assemble, in batch.go) uses it
	// to find that task's own Result and source.Task (batch[BatchIndex].Result,
	// batch[BatchIndex].Task) instead of indexing a map by SourceID, which
	// would silently collide for two tasks sharing a reused id.
	BatchIndex int

	// ID is the final id: the destination's task_prefix plus either the
	// source task's own number (docs/especificacion.md, "Identificadores",
	// point 2) or a reassigned one (point 4).
	ID string
	// Parent is Task.ParentTaskID resolved by its full canonical form, not
	// the case-sensitive mention pattern (docs/especificacion.md,
	// "Identificadores", point 7: prefix folded to uppercase, main number
	// as an int, and subtask number as an int when it has one), or empty
	// when the task had no parent, or its parent named an id outside the
	// whole source batch (dropped, with a Finding).
	Parent string
	// Dependencies is Task.Dependencies resolved the same way as Parent, in
	// the same order, with any element naming an id outside the batch
	// removed (with a Finding) rather than the whole line failing, and any
	// element resolving to the same final id as an earlier one in this same
	// list dropped silently (point 7's deduplication).
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
	// with Definition of Done by convert.Task) item for item, in the same
	// order: Number and Checked untouched, Text with every mention
	// rewritten.
	AcceptanceCriteria []source.Checkbox

	// Comments mirrors Result.Comments (already date-converted by
	// convert.Task) item for item, in the same order: Author and CreatedAt
	// untouched (point 5 explicitly excludes a comment's author from mention
	// rewriting), Body with every mention rewritten.
	Comments []Comment

	// Labels is Result.Labels (already including milestone:: and
	// project:: from convert.Task) with backlog.id::<source-id> appended
	// when the source id has a dot (docs/especificacion.md,
	// "Identificadores", point 9). A source label colliding with the
	// backlog.id key is dropped first, with a Finding, exactly the way
	// convert.Task already drops one colliding with milestone or project
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

// naturalSourceIDLess reports whether a sorts strictly before b in the
// source id's own natural order: main number ascending; when two ids share
// a main number, a SIMPLE id always sorts before any subtask of that same
// number ("TASK-1" reads before "TASK-1.0"), and two subtasks of the same
// parent sort by their own subtask number ascending. Identifiers uses this
// to break ties among ids reassigned by point 4 (see below), and
// ordinal.go's AssignOrdinals reuses it unchanged to break a tie between two
// source tasks that share the same ordinal (docs/especificacion.md, "Orden
// manual"), so that both files agree on what "the source id's own natural
// order" means and a batch converts the same way every time it runs.
func naturalSourceIDLess(a, b parsedSourceID) bool {
	if a.number != b.number {
		return a.number < b.number
	}
	if a.hasSub != b.hasSub {
		return !a.hasSub
	}
	return a.sub < b.sub
}

// canonicalIDKey returns the canonical form docs/especificacion.md,
// "Identificadores", point 1's last paragraph defines for deciding whether
// two source ids are the exact same id, explained further in
// docs/decisiones.md, section "Los identificadores conservan su número y
// cambian de prefijo", in its closing paragraph about same-number
// collisions between source ids: the prefix folded to uppercase, the main
// number as an int (ignoring leading zeros), and, only when the id has a
// dot, the subtask number as an int too (also ignoring leading zeros).
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

// groupTieBreakKey stands in for docs/especificacion.md, "Identificadores",
// point 4's "ruta relativa completa desde la raíz del origen" tie-break, in
// the one situation two members of the same shared-id group (docs/decisiones.md,
// "Los identificadores conservan su número y cambian de prefijo", the
// paragraph on id reuse) can actually need it: a SUBTASK id shared by one
// non-archived copy and one or more archived copies, all reassigned together
// because a subtask never conserves its number regardless of sharing. Within
// one shared-id group, two files can only have the exact same
// source.Task.File when exactly one of them is archived and the other is
// not: they then live in different directories (archive/tasks/ versus
// tasks/ or completed/). Two ARCHIVED copies of the same group always live
// in the very same archive/tasks/ directory, so an equal File there would
// mean they are literally the same file, not two distinct ones; and this
// tie-break is never reached by two non-archived copies of the same group at
// all, since more than one non-archived copy is already fatal
// (validateSourceShape). Prefixing File with whether the task is archived is
// therefore already enough to make this synthetic key unique in every case
// this tie-break is ever reached, without adding a real relative-path field
// to source.Task just for this.
//
// The prefix must sort an archived copy BEFORE a non-archived one, because
// that is what the real relative path always does: "archive/" sorts before
// both "tasks/" and "completed/" alphabetically ('a' < 'c' and 'a' < 't'),
// so archive/tasks/<file> is always less than tasks/<file> or
// completed/<file>, whatever <file> itself is named. "0-" for an archived
// copy and "1-" for a non-archived one reproduce exactly that ordering
// without needing a real directory name; within either bucket, the tie still
// falls through to File itself, ascending, unchanged from before.
func groupTieBreakKey(t source.Task) string {
	if t.Archived {
		return "0-" + t.File
	}
	return "1-" + t.File
}

// hasValidCreatedDate reports whether b's already-converted created_date
// (b.Result.CreatedAt, docs/especificacion.md, "Fechas") is present:
// convert.go's ConvertDate already turned an absent or invalid
// source.Task.CreatedDate into the empty string, the same criterion the rest
// of this module uses for "does this task have a valid created_date at all".
func hasValidCreatedDate(b TaskInput) bool {
	return b.Result.CreatedAt != ""
}

// sharedGroupAssignmentLess breaks a tie between two toReassign entries that
// share the exact same parsedSourceID: this can only happen between members
// of the same shared-id group (docs/especificacion.md, "Identificadores",
// point 1's last paragraph guarantees two DIFFERENT ids never collide in
// canonical form), so this implements point 4's own two-level tie-break, on
// top of naturalSourceIDLess: every entry with a valid created_date sorts
// before every entry without one; two entries that both have one sort by
// that date ascending; and a final tie, either between two equal dates or
// between two entries that both lack a date, is broken by groupTieBreakKey
// ascending.
func sharedGroupAssignmentLess(a, b reassignEntry) bool {
	if a.hasDate != b.hasDate {
		return a.hasDate
	}
	if a.hasDate && a.createdAt != b.createdAt {
		return a.createdAt < b.createdAt
	}
	return groupTieBreakKey(a.task) < groupTieBreakKey(b.task)
}

// nonArchivedMember returns the index (into batch) of the one non-archived
// task among members, and true, when there is one. validateSourceShape
// already guarantees a shared-id group never has more than one, so the first
// one found is the only one there is.
func nonArchivedMember(members []int, batch []TaskInput) (int, bool) {
	for _, i := range members {
		if !batch[i].Task.Archived {
			return i, true
		}
	}
	return 0, false
}

// pickGroupWinner implements the tie-break docs/especificacion.md,
// "Identificadores", point 5 defines for a shared-id group with NO
// non-archived member at all: among the members with a valid created_date,
// if there are any, the one with the most recent date wins, ties broken by
// the largest groupTieBreakKey; when no member has a valid date at all, the
// one with the largest groupTieBreakKey wins outright. This deliberately
// does NOT reuse point 4's own bucketed order (dated members entirely before
// undated ones) as a single sort-and-take-the-last, because that would let
// an undated member outrank a dated one whenever the group mixes both,
// exactly backwards from point 5's "at least one has a valid date" wording:
// point 5 only ever falls back to undated members when NONE of them has a
// date, so the two cases below are evaluated as an if/else on that
// condition, each with its own ascending sort ending on its own winner,
// rather than as one merged order.
func pickGroupWinner(members []int, batch []TaskInput) int {
	var dated []int
	for _, i := range members {
		if hasValidCreatedDate(batch[i]) {
			dated = append(dated, i)
		}
	}

	pool := dated
	if len(pool) == 0 {
		pool = append([]int(nil), members...)
	}

	sort.SliceStable(pool, func(x, y int) bool {
		ix, iy := pool[x], pool[y]
		cx, cy := batch[ix].Result.CreatedAt, batch[iy].Result.CreatedAt
		if cx != cy {
			return cx < cy
		}
		return groupTieBreakKey(batch[ix].Task) < groupTieBreakKey(batch[iy].Task)
	})
	return pool[len(pool)-1]
}

// ambiguousGroupInfo records, for one shared-id group's canonical key, which
// of docs/especificacion.md, "Identificadores", point 5's two ambiguous
// cases produced its equivalentsByCanonical entry: SelectedFromArchived is
// false when exactly one member is non-archived (the mention/parent/
// dependency resolves to that one), and true when none of the group's
// members is active nor terminated (it resolves to pickGroupWinner's choice
// among the archived copies instead). A canonical key with no entry in this
// map at all is not shared by more than one task, so resolving against it is
// never ambiguous and never raises the Finding this map exists to drive.
type ambiguousGroupInfo struct {
	SelectedFromArchived bool
}

// ambiguousMessage builds the Finding text docs/especificacion.md,
// "Identificadores", points 5 and 7 define for a mention, parent, or
// dependency that resolved against a shared id (a number Backlog.md reused
// after archiving, point 1) rather than a single unambiguous source task.
// kind is "mention", "parent", or "dependency": point 7 reuses point 5's
// wording verbatim for parent and dependency, only that one word differs.
// count is 1 for the exact literal wording points 5 and 7 both give; any
// other count means this file, field, and shared id resolved this way more
// than once, grouped into one line with a count instead of one Finding per
// occurrence, the same treatment point 6 already gives a repeated unresolved
// mention. Neither point fixes a literal wording for that grouped case, the
// same way point 6 does not either.
func ambiguousMessage(kind string, count int, canonical, finalID string, info ambiguousGroupInfo) string {
	role := "the non-archived task sharing this id"
	if info.SelectedFromArchived {
		role = "the archived task selected among those sharing this id"
	}
	if count == 1 {
		return fmt.Sprintf("%s: %s resolved to %s, %s", canonical, kind, finalID, role)
	}
	return fmt.Sprintf("%d %s(s) of %s resolve to %s, %s", count, kind, canonical, finalID, role)
}

// seenSourceID records, for validateSourceShape's duplicate check, which
// file and which literal id string first produced a given canonical key
// among the batch's NON-ARCHIVED tasks, so the error message can name both
// tasks by file and by their own original id, not just by the canonical key
// they collide on.
type seenSourceID struct {
	file string
	id   string
}

// validateSourceShape implements docs/especificacion.md, "Identificadores",
// point 1, over ALL of the batch's source tasks at once (active, completed,
// and archived tasks together): every id must have the shape parseSourceID
// accepts, and every id must share the same prefix once folded to uppercase
// (Backlog.md always writes the prefix in uppercase in the id itself, even
// when its own configuration stores it differently, so no configuration
// needs to be read here).
//
// The one case point 1 still makes fatal is MORE THAN ONE non-archived task
// (from tasks/ or completed/, never archive/tasks/) canonicalizing
// (canonicalIDKey) to the same id: same prefix folded, same main number, the
// same shape (both simple or both subtask, since a simple id and a subtask
// id with the same main number are never the same id), and, only when both
// are subtasks, the same subtask number too. Backlog.md reuses the number of
// an archived task in the next one it creates (docs/decisiones.md, "Los
// identificadores conservan su número y cambian de prefijo"), so any number
// of ARCHIVED tasks may canonicalize to the same id, alone or alongside a
// single non-archived one, without this ever being fatal: it is Identifiers
// itself (point 4's shared-id handling, further down this file) that decides
// which of them conserves the number and which are reassigned with "id
// reused after archiving".
//
// It returns a parsedSourceID per task, keyed by the task's literal id
// (harmlessly overwritten with an identical value when two tasks happen to
// share the exact same literal id string, which parseSourceID always parses
// the same way), and the shared prefix already folded to uppercase: the form
// every later step in this file writes into a final id and matches text
// mentions against.
//
// This is the one failure "Identificadores" turns into a Go error rather
// than a Finding: docs/especificacion.md, section "Códigos de salida", the
// row for exit code 3, which aborts the whole batch instead of skipping one
// task. Translating this error into that exit code is a later phase's job;
// this function only needs to say clearly, in its error's message, which of
// the three things failed and with what value.
func validateSourceShape(tasks []source.Task) (parsed map[string]parsedSourceID, prefixUpper string, err error) {
	parsed = make(map[string]parsedSourceID, len(tasks))
	firstNonArchivedIn := make(map[string]seenSourceID, len(tasks))

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

		if !t.Archived {
			canonical := canonicalIDKey(folded, p)
			if prev, duplicate := firstNonArchivedIn[canonical]; duplicate {
				return nil, "", fmt.Errorf(
					"identifiers: %s (id %q) and %s (id %q) are the same id",
					prev.file, prev.id, t.File, t.ID,
				)
			}
			firstNonArchivedIn[canonical] = seenSourceID{file: t.File, id: t.ID}
		}

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

// reassignReason records, for one entry of toReassign, which of point 4's
// (or point 8's) three reasons put it there, so the assignment loop further
// down can print the exact wording point 8 fixes for each one.
type reassignReason int

const (
	// reassignCollision is a simple id whose own number is already taken on
	// the destination (point 2's collision, point 8's "is taken on the
	// destination, reassigned to" wording).
	reassignCollision reassignReason = iota
	// reassignSubtask is any id with a dot: a subtask never conserves its
	// number, shared id or not (point 8's "subtask ids have no equivalent").
	reassignSubtask
	// reassignReused is a simple id that is one of the ARCHIVED copies of a
	// shared id (point 1, point 4): reassigned regardless of whether its own
	// number would also have collided (point 8's "id reused after
	// archiving", the only Finding printed for it even then).
	reassignReused
)

// reassignEntry is one task that could not keep its own number under
// docs/especificacion.md, "Identificadores", point 4, waiting to be
// assigned a new one. createdAt and hasDate are b.Result's
// already-converted created_date (docs/especificacion.md, "Fechas") and
// whether it is present at all, carried alongside task and parsed purely so
// sharedGroupAssignmentLess can break a tie without needing the whole
// TaskInput.
type reassignEntry struct {
	index     int
	task      source.Task
	parsed    parsedSourceID
	createdAt string
	hasDate   bool
	reason    reassignReason
}

// Identifiers resolves identifiers for the whole batch
// (docs/especificacion.md, "Identificadores"). It receives the whole batch
// at once, because an
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
//     (point 4 and point 8's exact wording, "id reused after archiving"
//     included), an unresolved or wrong-case mention grouped by file and
//     field (point 6), an ambiguous resolution against a shared id grouped
//     by file, field, and shared id (points 5 and 7), a dropped parent or
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

	// naiveTitle only ever substitutes a mention whose number names a
	// SIMPLE source id (docs/especificacion.md, "Identificadores", point 3;
	// see naiveTitle's own comment for the full rule), so this is every
	// main number that belongs to a non-subtask task in the batch, built
	// once from parsedByID rather than inside naiveTitle's own closure per
	// call.
	simpleSourceNumbers := make(map[int]bool, len(parsedByID))
	for _, p := range parsedByID {
		if !p.hasSub {
			simpleSourceNumbers[p.number] = true
		}
	}

	// groups collects, for every canonical key (canonicalIDKey), the indices
	// into batch of every task that canonicalizes to it, in batch order.
	// Almost every key maps to exactly one index; a key mapping to more than
	// one is docs/especificacion.md, "Identificadores", point 1's shared-id
	// case (Backlog.md reusing a number after archiving), which
	// validateSourceShape already allowed through because at most one member
	// of the group is non-archived. canonicalOf[i] is the same key for
	// batch[i], kept alongside so later steps do not need to reparse
	// batch[i].Task.ID to find its own group again.
	groups := make(map[string][]int, len(batch))
	canonicalOf := make([]string, len(batch))
	for i, b := range batch {
		key := canonicalIDKey(prefixUpper, parsedByID[b.Task.ID])
		canonicalOf[i] = key
		groups[key] = append(groups[key], i)
	}

	var findings []source.Finding

	// docs/especificacion.md, "Identificadores", point 3: a task is a
	// repeat of an earlier import when an existing destination task has the
	// same title and the same already-converted createdAt. The title
	// compared here is NOT the source task's raw title: point 3 requires
	// the naive mention rewrite naiveTitle computes (see its own comment),
	// so that a title mentioning another source id still matches the title
	// biso actually wrote on a previous run. The equivalent of a skipped
	// task is the id that destination task ALREADY has, and it is skipped
	// rather than emitted. This is evaluated per task, by INDEX rather than
	// by literal source id: two tasks that share a literal id (point 1's
	// reuse case, where Backlog.md hands the exact same id string to the
	// task that reused a number) must still be checked against the
	// destination independently of one another, point 3's own closing
	// sentence.
	perTaskFinalID := make([]string, len(batch))
	skipped := make([]bool, len(batch))
	for i, b := range batch {
		candidateTitle := naiveTitle(b.Task.Title, pattern, prefixUpper, board.Config.TaskPrefix, simpleSourceNumbers)
		for _, dt := range board.Tasks {
			if candidateTitle == dt.Title && b.Result.CreatedAt == dt.CreatedAt {
				perTaskFinalID[i] = dt.ID
				skipped[i] = true
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

	var toReassign []reassignEntry

	for i, b := range batch {
		if skipped[i] {
			continue
		}
		p := parsedByID[b.Task.ID]

		if p.hasSub {
			// A subtask never conserves its number, shared id or not (point
			// 4's own closing sentence for the subtask case).
			toReassign = append(toReassign, reassignEntry{
				index: i, task: b.Task, parsed: p,
				createdAt: b.Result.CreatedAt, hasDate: hasValidCreatedDate(b),
				reason: reassignSubtask,
			})
			continue
		}

		// designated is whether THIS task is the one point 4 lets compete
		// for its own number under the normal collision rule: either it is
		// the only member of its canonical group at all (the common case,
		// no id reuse involved), or it is the group's one non-archived
		// member. validateSourceShape already guarantees a shared group
		// never has more than one non-archived member, so a non-archived
		// task in a shared group is always that one member.
		group := groups[canonicalOf[i]]
		designated := len(group) == 1 || !b.Task.Archived

		if designated {
			if !destNumbers[p.number] {
				perTaskFinalID[i] = fmt.Sprintf("%s-%d", board.Config.TaskPrefix, p.number)
				continue
			}
			toReassign = append(toReassign, reassignEntry{
				index: i, task: b.Task, parsed: p,
				createdAt: b.Result.CreatedAt, hasDate: hasValidCreatedDate(b),
				reason: reassignCollision,
			})
			continue
		}

		// A simple id, in a shared group of more than one member, and this
		// copy is NOT the designated (non-archived) one: it is one of the
		// archived copies of a number Backlog.md later reused, reassigned
		// regardless of whether its own number would also have collided
		// (point 8's note that "id reused after archiving" is the only
		// Finding printed for it in that case).
		toReassign = append(toReassign, reassignEntry{
			index: i, task: b.Task, parsed: p,
			createdAt: b.Result.CreatedAt, hasDate: hasValidCreatedDate(b),
			reason: reassignReused,
		})
	}

	// Point 4 assigns reassigned numbers in the source id's own natural
	// order: main number ascending; when two entries share a main number, a
	// SIMPLE id always sorts before any subtask of that same number ("TASK-1"
	// reads before "TASK-1.0"), and two subtasks of the same parent sort by
	// their own subtask number ascending. Two entries can share a main
	// number more than one way: two subtasks of the same parent, a simple id
	// that collides with the destination reassigned alongside one of its own
	// subtasks, or, since this arrangement, two or more members of the same
	// shared-id group (point 1's reuse case), which naturalSourceIDLess
	// cannot tell apart at all since they carry the identical
	// parsedSourceID. sharedGroupAssignmentLess breaks exactly that last
	// tie, point 4's own two-level created_date/path order; it is never
	// consulted for the other two cases, where naturalSourceIDLess already
	// decides one way or the other.
	sort.SliceStable(toReassign, func(i, j int) bool {
		a, b := toReassign[i], toReassign[j]
		if naturalSourceIDLess(a.parsed, b.parsed) {
			return true
		}
		if naturalSourceIDLess(b.parsed, a.parsed) {
			return false
		}
		return sharedGroupAssignmentLess(a, b)
	})

	nextNumber := max(maxDestNumber, maxSourceNumber) + 1
	for _, r := range toReassign {
		newID := fmt.Sprintf("%s-%d", board.Config.TaskPrefix, nextNumber)
		perTaskFinalID[r.index] = newID

		var message string
		switch r.reason {
		case reassignSubtask:
			message = fmt.Sprintf("%s: id %s assigned (subtask ids have no equivalent)", r.task.ID, newID)
		case reassignReused:
			message = fmt.Sprintf("%s: id %s assigned (id reused after archiving)", r.task.ID, newID)
		default: // reassignCollision
			wouldBeID := fmt.Sprintf("%s-%d", board.Config.TaskPrefix, r.parsed.number)
			message = fmt.Sprintf("%s: id %s is taken on the destination, reassigned to %s", r.task.ID, wouldBeID, newID)
		}
		findings = append(findings, source.Finding{File: r.task.File, Field: "id", Message: message})

		nextNumber++
	}

	// docs/especificacion.md, "Identificadores", points 5 and 7: for every
	// canonical group, decide the one final id a mention, parent, or
	// dependency naming it resolves to. A group of one is the ordinary case,
	// point 5's "TASK-1" finding its own equivalent; a shared group (point
	// 1's reuse case) resolves to its one non-archived member's own final id
	// when it has one, or, when every member is archived, to
	// pickGroupWinner's choice among them (point 5's "at least one has a
	// created_date... resolves to the archived task selected among those
	// sharing this id" wording, reused verbatim by point 7). ambiguousGroups
	// records which of the two shared-group cases produced each entry,
	// purely so identifyTask can raise the right Finding when a
	// mention/parent/dependency actually resolves against one; a canonical
	// key with no entry there was never shared, so resolving against it is
	// never ambiguous.
	equivalentsByCanonical := make(map[string]string, len(groups))
	ambiguousGroups := make(map[string]ambiguousGroupInfo, len(groups))
	for key, members := range groups {
		if len(members) == 1 {
			equivalentsByCanonical[key] = perTaskFinalID[members[0]]
			continue
		}
		if nonArchived, ok := nonArchivedMember(members, batch); ok {
			equivalentsByCanonical[key] = perTaskFinalID[nonArchived]
			ambiguousGroups[key] = ambiguousGroupInfo{SelectedFromArchived: false}
			continue
		}
		winner := pickGroupWinner(members, batch)
		equivalentsByCanonical[key] = perTaskFinalID[winner]
		ambiguousGroups[key] = ambiguousGroupInfo{SelectedFromArchived: true}
	}

	// docs/especificacion.md, "Identificadores", points 5, 6, 7, and 9:
	// build the actual output for every task that was not skipped, in
	// batch order, now that perTaskFinalID, equivalentsByCanonical, and
	// ambiguousGroups (all built by points 2, 3, and 4 together) are
	// complete.
	var out []Identified
	for i, b := range batch {
		if skipped[i] {
			continue
		}
		identified, taskFindings := identifyTask(b, parsedByID[b.Task.ID], perTaskFinalID[i], equivalentsByCanonical, ambiguousGroups, pattern, prefixUpper)
		identified.BatchIndex = i
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

// resolveByCanonicalForm implements docs/especificacion.md,
// "Identificadores", point 7's full canonical resolution for parent and
// dependencies: value is parsed the same way any source id is
// (parseSourceID), and its prefix is folded to uppercase and compared
// against prefixUpper WITHOUT the case sensitivity point 5's free-text
// mention pattern requires. docs/decisiones.md's paragraph on parent and
// dependencies explains why the two differ: a parent or dependency field is
// never free text where an id could appear by accident (unlike a branch
// name such as "task-10-modelo" sitting in prose), so there is nothing to
// protect from a false match by keeping the prefix case-sensitive here.
//
// A value that does not even parse as an id (parseSourceID's ok is false),
// or whose prefix does not fold to prefixUpper, is never found: ok is
// false, matching the "value that does not have the shape of an id at all"
// case point 7 still drops with a Finding, and canonical is the empty string
// in that case too. When ok is true, the caller uses canonical to check
// ambiguousGroups, since resolveByCanonicalForm itself does not know whether
// the group it resolved against was shared at all.
func resolveByCanonicalForm(value, prefixUpper string, equivalentsByCanonical map[string]string) (resolved, canonical string, ok bool) {
	parsed, parseOK := parseSourceID(value)
	if !parseOK || strings.ToUpper(parsed.prefix) != prefixUpper {
		return "", "", false
	}
	canonical = canonicalIDKey(prefixUpper, parsed)
	resolved, ok = equivalentsByCanonical[canonical]
	if !ok {
		return "", "", false
	}
	return resolved, canonical, true
}

// identifyTask builds the Identified for one non-skipped task, plus every
// Finding raised while rewriting its parent, dependencies, mentions, and
// backlog.id:: label. p is parsedByID[b.Task.ID], and finalID is
// perTaskFinalID for this same task, both passed in rather than recomputed
// since the caller already has them (finalID by INDEX rather than by literal
// source id, so that two tasks sharing a literal id, docs/especificacion.md,
// "Identificadores", point 1's reuse case, still each get their own value
// here).
//
// equivalentsByCanonical (keyed by canonical form) and ambiguousGroups
// (keyed the same way, but present only for a canonical key more than one
// task shares) serve two lookups here:
//
//   - Parent and Dependencies resolve docs/especificacion.md,
//     "Identificadores", point 7's full canonical form, prefix folded to
//     uppercase included: resolveByCanonicalForm, built on
//     equivalentsByCanonical, the same index rewriteMentions uses.
//   - A mention found inside free text (point 5) also resolves by
//     canonical form, through equivalentsByCanonical, but with the
//     prefix's case kept significant, enforced by mentionMatch.ExactCase
//     before rewriteMentions even looks anything up: docs/decisiones.md's
//     paragraph on parent/dependencies explains why the two differ (a
//     parent or dependency field is never free text where an id could
//     appear by accident, so there is no false match to protect against by
//     keeping its prefix case-sensitive, unlike a mention in prose).
//
// Both consult ambiguousGroups, by the canonical key they just resolved
// against, to raise points 5 and 7's "mention/parent/dependency resolved
// to..." Finding whenever that resolution went through a shared id.
func identifyTask(
	b TaskInput,
	p parsedSourceID,
	finalID string,
	equivalentsByCanonical map[string]string,
	ambiguousGroups map[string]ambiguousGroupInfo,
	pattern *regexp.Regexp,
	prefixUpper string,
) (Identified, []source.Finding) {
	var findings []source.Finding
	unresolvedCounts := make(map[string]int, len(mentionFieldOrder))
	// ambiguousByField accumulates, per field, how many times each shared
	// canonical id was actually resolved against while rewriting that
	// field's mentions (points 5 and 6's shared grouping: file, field, AND
	// shared id together, docs/especificacion.md, "Identificadores", point
	// 5's closing paragraph), across however many rewrite() calls that field
	// takes (one for title/description/plan/notes/summary, one per
	// acceptance criterion, one per comment).
	ambiguousByField := make(map[string]map[string]int, len(mentionFieldOrder))

	rewrite := func(field, text string) string {
		rewritten, unresolved, hits := rewriteMentions(text, pattern, prefixUpper, equivalentsByCanonical, ambiguousGroups)
		unresolvedCounts[field] += unresolved
		for canonical, n := range hits {
			agg := ambiguousByField[field]
			if agg == nil {
				agg = make(map[string]int)
				ambiguousByField[field] = agg
			}
			agg[canonical] += n
		}
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
		if agg := ambiguousByField[field]; len(agg) > 0 {
			// Deterministic regardless of Go's map iteration order: sorted
			// alphabetically by the shared id's own canonical text, since
			// neither point 5 nor point 7 fixes an order for two DIFFERENT
			// shared ids resolved in the same file and field.
			canonicals := make([]string, 0, len(agg))
			for canonical := range agg {
				canonicals = append(canonicals, canonical)
			}
			sort.Strings(canonicals)
			for _, canonical := range canonicals {
				findings = append(findings, source.Finding{
					File:    b.Task.File,
					Field:   field,
					Message: ambiguousMessage("mention", agg[canonical], canonical, equivalentsByCanonical[canonical], ambiguousGroups[canonical]),
				})
			}
		}
	}

	parent := ""
	if b.Task.ParentTaskID != "" {
		if resolved, canonical, ok := resolveByCanonicalForm(b.Task.ParentTaskID, prefixUpper, equivalentsByCanonical); ok {
			parent = resolved
			if info, shared := ambiguousGroups[canonical]; shared {
				findings = append(findings, source.Finding{
					File:    b.Task.File,
					Field:   "parent",
					Message: ambiguousMessage("parent", 1, canonical, resolved, info),
				})
			}
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

	// docs/especificacion.md, "Identificadores", point 7's deduplication:
	// two elements of the same task's Dependencies that resolve to the same
	// final id by canonical form keep only the first, in original order,
	// the same rule CleanTokenList (tokens.go) already applies when two
	// labels or assignees of one task collapse to the same value.
	//
	// depAmbiguous and depAmbiguousOrder group point 7's ambiguous-resolution
	// Finding by shared id, in the order each shared id was first
	// encountered in Dependencies: a single field can have at most one
	// grouped line per shared id here, same as a free-text field's mentions,
	// even though Dependencies is a list rather than one block of prose.
	var dependencies []string
	seenDependencies := make(map[string]bool, len(b.Task.Dependencies))
	depAmbiguous := make(map[string]int)
	var depAmbiguousOrder []string
	for _, dep := range b.Task.Dependencies {
		resolved, canonical, ok := resolveByCanonicalForm(dep, prefixUpper, equivalentsByCanonical)
		if !ok {
			findings = append(findings, source.Finding{
				File:  b.Task.File,
				Field: "dependencies",
				Message: fmt.Sprintf(
					"dependency %q does not name any task in the source batch, dropped",
					dep,
				),
			})
			continue
		}
		if _, shared := ambiguousGroups[canonical]; shared {
			if depAmbiguous[canonical] == 0 {
				depAmbiguousOrder = append(depAmbiguousOrder, canonical)
			}
			depAmbiguous[canonical]++
		}
		if seenDependencies[resolved] {
			continue
		}
		seenDependencies[resolved] = true
		dependencies = append(dependencies, resolved)
	}
	for _, canonical := range depAmbiguousOrder {
		findings = append(findings, source.Finding{
			File:    b.Task.File,
			Field:   "dependencies",
			Message: ambiguousMessage("dependency", depAmbiguous[canonical], canonical, equivalentsByCanonical[canonical], ambiguousGroups[canonical]),
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
		ID:                 finalID,
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
// through pattern. That is the single-pass substitution point 5 requires,
// the one that keeps a reassigned number from ever being substituted a
// second time, and it is what makes rewriteMentions and naiveTitle share
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
// finds whose case exactly matches prefixUpper is looked up in
// equivalentsByCanonical by its CANONICAL form, not by its exact string
// (docs/decisiones.md's paragraph on mentions and canonical form: "TASK-1"
// in text must find a source task recorded as "TASK-001" just as it would
// find one recorded as "TASK-1"), and substituted with its final id when
// found (point 5). Every other candidate, whether its case matches but its
// canonical form names no task in the batch, or its case does not match at
// all (point 6's "Xyz-002"/"task-12"), is left untouched and counted in
// unresolved, since both share the same grouped-by-file-and-field Finding
// identifyTask raises from that count.
//
// The pattern that decides whether something counts as a mention at all is
// unchanged by this: canonical form only enters AFTER a candidate already
// passed scanMentions's exact-case, word-boundary check, to decide which
// batch task it names.
//
// ambiguousGroups is nil-safe (a nil map is only ever read from, never
// written to) so a caller that never deals with a shared id, such as
// TestRewriteMentionsSubstitutesFromTheOriginalTextOnly, can pass nil.
// ambiguousHits counts, by canonical key, how many candidates in this one
// text actually resolved against a shared id (a key present in
// ambiguousGroups): identifyTask merges this into its own per-field
// grouping for points 5 and 7's "resolved to..., the ... task sharing this
// id" Finding. It is nil when this text raised no such case at all.
func rewriteMentions(text string, pattern *regexp.Regexp, prefixUpper string, equivalentsByCanonical map[string]string, ambiguousGroups map[string]ambiguousGroupInfo) (rewritten string, unresolved int, ambiguousHits map[string]int) {
	rewritten = scanMentions(text, pattern, prefixUpper, func(m mentionMatch) string {
		if m.ExactCase {
			if parsed, ok := parseSourceID(m.Text); ok {
				canonical := canonicalIDKey(prefixUpper, parsed)
				if final, ok := equivalentsByCanonical[canonical]; ok {
					if _, shared := ambiguousGroups[canonical]; shared {
						if ambiguousHits == nil {
							ambiguousHits = make(map[string]int)
						}
						ambiguousHits[canonical]++
					}
					return final
				}
			}
		}
		unresolved++
		return m.Text
	})
	return rewritten, unresolved, ambiguousHits
}

// naiveTitle implements the "naive" mention rewrite docs/especificacion.md,
// "Identificadores", point 3, defines for exactly one purpose: deciding
// whether a source task is already on the destination, BEFORE this phase
// knows the equivalence table. That table needs to know which tasks are
// skipped before it can be completed, so this comparison cannot depend on
// it without becoming circular; naiveTitle breaks that circularity by
// computing something simpler on its own. The full reasoning, and the
// limitation it knowingly accepts, are in docs/decisiones.md, section "Los
// identificadores conservan su número y cambian de prefijo", in its
// paragraph about the naive title comparison and the one right after it
// about the limitation this rewrite accepts.
//
// point 3's rule is narrower than "substitute every candidate": a candidate
// is substituted, with the destination's own prefix and the SAME number it
// names, ONLY
// when its case exactly matches the source prefix AND its number, with no
// subtask suffix at all, matches a SIMPLE (non-subtask) source id present
// in simpleSourceNumbers. Every other exact-case candidate is left
// untouched:
//
//   - one shaped like a subtask (it has a dot) is always left untouched,
//     whether or not that exact subtask exists in the batch, because a
//     subtask always gets a fresh number this phase cannot predict yet
//     (point 4); this is the accepted limitation docs/decisiones.md
//     describes for a title that mentions a subtask;
//   - one shaped like a simple id but whose number names no simple source
//     task at all is left untouched too, the same way point 6 leaves an
//     unresolved mention untouched in the definitive rewrite.
//
// A candidate whose case does not exactly match the source prefix is left
// untouched as well, the same as rewriteMentions leaves it.
//
// The result is used ONLY for the "already on the destination" comparison
// and is discarded right after: the title actually written to the
// destination, for a task that turns out not to be a duplicate, always
// comes from rewriteMentions once the equivalence table is complete.
// Reusing this naive result as a task's final title would be wrong even
// for a task that is not skipped, since it never accounts for a collision
// or a reassignment at all.
func naiveTitle(title string, pattern *regexp.Regexp, prefixUpper, destinationPrefix string, simpleSourceNumbers map[int]bool) string {
	return scanMentions(title, pattern, prefixUpper, func(m mentionMatch) string {
		if !m.ExactCase {
			return m.Text
		}

		numeric := m.Text[len(prefixUpper)+1:]
		if strings.IndexByte(numeric, '.') >= 0 {
			// Shaped like a subtask: never substituted here, regardless of
			// whether that exact subtask exists in the batch (see the
			// function comment's first bullet).
			return m.Text
		}

		n, err := strconv.Atoi(numeric)
		if err != nil || !simpleSourceNumbers[n] {
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
