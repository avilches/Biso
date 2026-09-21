package cli

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the output of docs/spec/cmd/prime.md: the startup message,
// the size budget of docs/spec/presupuestos.md#el-presupuesto-de-tamaño
// and the cascade that trims the summary when it does not fit, plus the
// `prime` envelope.
//
// The fixed half of the message lives in prime_text.go. What is here is
// everything that depends on the board.

// The two halves of docs/spec/presupuestos.md#el-presupuesto-de-tamaño.
// They add up to the hard cap of 5,504 bytes, which is the one number of
// the whole specification that
// docs/spec/estabilidad.md#el-contrato-de-estabilidad freezes.
const (
	PrimeFixedBudget   = 3840
	primeSummaryBudget = 1664
	PrimeBudget        = PrimeFixedBudget + primeSummaryBudget
)

// The three count lines of the summary, each one with the filter of
// `biso ls` that answers its block whole.
const (
	inProgressFilter  = "biso ls --active"
	needsAnswerFilter = "biso ls --waiting"
	nextUpFilter      = "biso ls --not-active --not-waiting"
)

// The four headings, in the order they are printed.
const (
	inProgressHeading  = "IN PROGRESS"
	needsAnswerHeading = "NEEDS ANSWER"
	assignedHeading    = "ASSIGNED TO YOU"
	nextUpHeading      = "NEXT UP  (not assigned to you, by urgency)"
)

func runPrime(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	if asJSON && p.Has("full") {
		// The rejection is plain text on stderr and not the error
		// envelope, because --json is itself the invalid part of the call
		// (docs/spec/cmd/prime.md#parámetros).
		return fail(s, false, &model.Error{
			ExitCode: 2,
			Code:     "incompatible_flags",
			Message:  "--full does not combine with --json",
			Detail: []string{
				"       --full's output is help text, and --json has no room for it",
			},
		}, warningsOf(p))
	}
	params, err := primeParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.Prime(env, params)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)
	if asJSON {
		writeEnvelope(s, env, "prime", primeData(result))
		return 0
	}
	fmt.Fprint(s.Stdout, renderPrime(result, Version, p.Has("full")))
	return 0
}

func primeParams(p *Parsed) (ops.PrimeParams, error) {
	if len(p.Positionals) > 0 {
		return ops.PrimeParams{}, errUnexpectedArgument(p.Positionals[0])
	}
	var params ops.PrimeParams
	if v, ok := p.Value("limit"); ok {
		limit, err := parseLimit(v)
		if err != nil {
			return ops.PrimeParams{}, err
		}
		params.Limit, params.HasLimit = limit, true
	}
	return params, nil
}

// primeCut is how much of each part of the summary the message prints: how
// many rows of each of the four blocks of tasks, and how many elements of
// each of the three lists of the BOARD block. Those seven numbers are what
// the cascade of docs/spec/presupuestos.md#el-presupuesto-de-tamaño
// reduces, in that order, and nothing else about the message changes with
// them.
type primeCut struct {
	inProgress, needsAnswer, assigned, nextUp int
	priorities, types, statuses               int
}

// renderPrime is the whole message. The fixed part never changes, so the
// only thing that can overflow the cap is the summary, and the cascade
// below is what brings it back inside.
func renderPrime(r *ops.PrimeResult, version string, full bool) string {
	// The two pieces of the summary are not next to each other in the
	// message: the BOARD block goes above the fixed part and the four
	// blocks of tasks below it. They are built and measured together,
	// because both are what changes with the board.
	board, blocks := primeSummary(r, primeCutThatFits(r))
	// The floor of the cascade, which is what makes the cap hold for every
	// board there can be and not only for every board that has tasks.
	board = primeBoardWithinItsRoom(board, primeSummaryBudget-len(blocks))
	message := primeTitle(version) + "\n" + board +
		primeCommands + "\n" + primeFieldFlags + "\n" + primeRules + "\n" +
		blocks + primeClosing
	if full {
		message += "\n" + primeFullBlock()
	}
	return message
}

// primeSummary is the half of the message that depends on the board, in its
// two pieces: the BOARD block with the blank line that follows it, and then
// the blocks of tasks, each one with its own
// (docs/spec/presupuestos.md#el-presupuesto-de-tamaño).
func primeSummary(r *ops.PrimeResult, cut primeCut) (board, blocks string) {
	return primeBoardBlock(r, cut) + "\n", primeBlocks(r, cut)
}

// primeCutThatFits is the cascade of
// docs/spec/presupuestos.md#el-presupuesto-de-tamaño: the summary is tried
// whole, and when it does not fit the rows of NEXT UP go first, then those
// of ASSIGNED TO YOU, then NEEDS ANSWER, then IN PROGRESS. A block that
// ends with no row at all is left as its single count line, which is the
// fifth step of that list.
//
// The sixth step is the BOARD block, and it comes last because it is the
// only thing in the summary that describes the board instead of listing its
// tasks: its three lists lose elements, priorities first, then types, then
// the counts line. It is not an ornament of the cascade but the only step
// that reaches a board with no task at all, whose whole summary is that
// block.
//
// Each of the seven numbers is searched and not walked one step at a time:
// the summary only grows when a block gains a row or a list gains an
// element, so the largest number that fits can be found by halving. It
// matters, because this runs inside the startup budget of
// docs/spec/presupuestos.md#el-presupuesto-de-arranque and a board can have
// hundreds of tasks in flight.
func primeCutThatFits(r *ops.PrimeResult) primeCut {
	cut := primeCut{
		inProgress:  len(r.InProgress),
		needsAnswer: len(r.NeedsAnswer),
		assigned:    len(r.AssignedToYou),
		nextUp:      len(r.NextUp),
		priorities:  len(r.Board.Priorities),
		types:       len(r.Board.Types),
		statuses:    len(r.Board.Statuses),
	}
	fits := func() bool {
		board, blocks := primeSummary(r, cut)
		return len(board)+len(blocks) <= primeSummaryBudget
	}
	for _, shown := range []*int{
		&cut.nextUp, &cut.assigned, &cut.needsAnswer, &cut.inProgress,
		&cut.priorities, &cut.types, &cut.statuses,
	} {
		if fits() {
			return cut
		}
		low, high, best := 0, *shown, 0
		for low <= high {
			middle := (low + high) / 2
			*shown = middle
			if fits() {
				best, low = middle, middle+1
			} else {
				high = middle - 1
			}
		}
		*shown = best
	}
	return cut
}

// primeBoardWithinItsRoom is the seventh and last step of that cascade: a
// BOARD block that does not fit even with its three lists empty is cut
// where it is and closed with the three dots a title gets.
//
// It is the one place in the message where text is cut in the middle of a
// sentence, and it exists so that the cap has no exception at all: a board
// name is free text of any length, so no amount of dropping elements can
// bound this block on its own.
func primeBoardWithinItsRoom(board string, room int) string {
	if len(board) <= room {
		return board
	}
	// The tail carries the blank line that separates the block from what
	// follows it, which belongs to the summary like the block itself.
	const tail = ellipsis + "\n\n"
	keep := room - len(tail)
	if keep < 0 {
		keep = 0
	}
	return cutBytes(board, keep) + tail
}

// cutBytes is s cut to at most n bytes, never in the middle of a rune: the
// message is UTF-8 and half a rune is not text.
func cutBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// primeBoardBlock is the BOARD block: what the board is, what its
// vocabulary is, who is calling, and the one line that says a task could
// not be read.
//
// Its three lists are printed as far as cut says, which is what the sixth
// step of the cascade reduces (docs/spec/cmd/prime.md#el-recorte-en-cascada).
func primeBoardBlock(r *ops.PrimeResult, cut primeCut) string {
	var b strings.Builder
	fmt.Fprintf(&b, "BOARD  %s\n", r.Board.Name)

	counts := make([]string, 0, len(r.Board.Statuses))
	for _, status := range r.Board.Statuses {
		counts = append(counts, fmt.Sprintf("%s %d", status, r.Board.CountByStatus[status]))
	}
	fmt.Fprintf(&b, "  %s\n", shortenedList(counts, cut.statuses, " | "))
	fmt.Fprintf(&b, "  new tasks start in %s; `biso start` moves to %s; `biso finish` to %s\n",
		r.Board.InitialStatus, r.Board.ActiveStatus, r.Board.TerminalStatus)

	row := func(label, value string) {
		fmt.Fprintf(&b, "  %-12s%s\n", label, value)
	}
	row("types", vocabulary(r.Board.Types, cut.types))
	row("priorities", vocabulary(r.Board.Priorities, cut.priorities))
	// identityOrNotice is the same sentence `biso where` prints for a
	// caller with no identity, and it travels inside the message because
	// biso prime writes nothing on stderr at all.
	row("you are", identityOrNotice(r.Me))
	if n := len(r.Skipped); n > 0 {
		row("unreadable", skippedSentence(n))
	}
	return b.String()
}

// vocabulary writes a configured list, and says so when it is empty: a
// board with no types configured takes no --type at all, and whoever is
// starting has to learn that here (docs/spec/cmd/prime.md#la-salida-literal).
// That state is reachable: `biso config set types ""` empties the list when
// no task uses one of its values (docs/spec/cmd/config.md#comportamiento-caso-a-caso).
func vocabulary(values []string, shown int) string {
	if len(values) == 0 {
		return "(none)"
	}
	return shortenedList(values, shown, ", ")
}

// shortenedList joins the first `shown` values and says how many it left
// out, which is what the sixth step of the cascade does to the three lists
// of the BOARD block. The `+N more` carries no command, unlike the count
// line of a block of tasks, because no call of `biso ls` answers with the
// rest of a vocabulary (docs/spec/cmd/prime.md#el-recorte-en-cascada).
func shortenedList(values []string, shown int, separator string) string {
	if shown >= len(values) {
		return strings.Join(values, separator)
	}
	rest := fmt.Sprintf("+%d more", len(values)-shown)
	if shown <= 0 {
		return rest
	}
	return strings.Join(values[:shown], separator) + separator + rest
}

func skippedSentence(n int) string {
	if n == 1 {
		return "1 task could not be read and was skipped"
	}
	return fmt.Sprintf("%d tasks could not be read and were skipped", n)
}

// primeBlocks is the four blocks of tasks, each one followed by the blank
// line that belongs to it, or the empty-board notice that replaces all
// four.
//
// The widths of columns 1 to 7 are computed over the rows of the four
// blocks together, so that they read as one table
// (docs/spec/cmd/prime.md#la-salida-literal). The question of a parked
// task and the expired lease of an active one are not rows and do not
// count.
func primeBlocks(r *ops.PrimeResult, cut primeCut) string {
	if r.Empty {
		return primeEmptyBoard + "\n"
	}
	inProgress := r.InProgress[:cut.inProgress]
	needsAnswer := r.NeedsAnswer[:cut.needsAnswer]
	assigned := r.AssignedToYou[:cut.assigned]
	nextUp := r.NextUp[:cut.nextUp]

	var rows [][]string
	for _, block := range [][]ops.TaskView{inProgress, needsAnswer, assigned, nextUp} {
		for _, v := range block {
			rows = append(rows, listRow(v))
		}
	}
	widths := columnWidths(rows)

	// The hidden tasks of ASSIGNED TO YOU and NEXT UP are counted
	// together, because the two share one count line: what --limit left
	// out plus whatever the cascade took afterwards.
	shared := r.HiddenCount +
		(len(r.AssignedToYou) - cut.assigned) + (len(r.NextUp) - cut.nextUp)

	var b strings.Builder
	b.WriteString(primeBlock(inProgressHeading, primeRows(inProgress, widths, leaseLine),
		len(r.InProgress)-cut.inProgress, inProgressFilter))
	b.WriteString(primeBlock(needsAnswerHeading, primeRows(needsAnswer, widths, questionLine),
		len(r.NeedsAnswer)-cut.needsAnswer, needsAnswerFilter))

	assignedRows := primeRows(assigned, widths, nil)
	nextUpRows := primeRows(nextUp, widths, nil)
	switch {
	case len(nextUpRows) > 0:
		b.WriteString(primeBlock(assignedHeading, assignedRows, 0, ""))
		b.WriteString(primeBlock(nextUpHeading, nextUpRows, shared, nextUpFilter))
	case len(assignedRows) > 0:
		b.WriteString(primeBlock(assignedHeading, assignedRows, shared, nextUpFilter))
	default:
		// Neither block printed a row, so the count line they share
		// stands alone where the two would have gone, which is what
		// --limit 0 always does (docs/spec/cmd/prime.md#el-recorte-en-cascada).
		b.WriteString(primeBlock("", nil, shared, nextUpFilter))
	}
	return b.String()
}

// primeBlock writes one block: its heading when it has any row, its rows,
// its count line when the cut left something out, and the blank line that
// separates it from what comes next. A block with neither rows nor hidden
// tasks writes nothing at all, not even its heading.
func primeBlock(heading string, rows []string, hidden int, filter string) string {
	if len(rows) == 0 && hidden <= 0 {
		return ""
	}
	var b strings.Builder
	if len(rows) > 0 {
		b.WriteString(heading + "\n")
	}
	for _, row := range rows {
		b.WriteString(row)
	}
	if hidden > 0 {
		fmt.Fprintf(&b, "  %d more not shown: `%s`\n", hidden, filter)
	}
	b.WriteString("\n")
	return b.String()
}

// primeRows writes the rows of one block, each one indented two spaces,
// with the extra indented line its block adds under some of them.
func primeRows(views []ops.TaskView, widths []int, extra func(ops.TaskView) string) []string {
	rows := make([]string, 0, len(views))
	for _, v := range views {
		row := "  " + renderRow(listRow(v), widths) + "\n"
		if extra != nil {
			row += extra(v)
		}
		rows = append(rows, row)
	}
	return rows
}

// leaseLine is the second line IN PROGRESS prints under a task whose lease
// expired: that block is the one whose heading claims somebody is working
// right now, and an expired lease contradicts exactly that.
func leaseLine(v ops.TaskView) string {
	if !v.LeaseExpired {
		return ""
	}
	return fmt.Sprintf("    lease expired %s, was held by %s\n",
		v.Task.LeaseExpiresAt.UTC().Format(model.InstantLayout), v.Task.LeaseHolder)
}

// questionLine is the second line of every task of NEEDS ANSWER: the body
// of the question on one line, cut to a hundred cells by the same rule the
// column algorithm applies to a title.
//
// Only the line breaks become a space, and no other run of whitespace is
// touched: the line is the question somebody wrote, not a normalized
// version of it (docs/spec/cmd/prime.md#la-salida-literal).
func questionLine(v ops.TaskView) string {
	if v.Task.Question == nil {
		return ""
	}
	return "    " + cutTitle(oneLine(v.Task.Question.Body)) + "\n"
}

// oneLine turns every line break of a body into a single space. A `\r\n` is
// one break and not two, and a lone `\r` is a break as well, which is the
// pair the rest of the specification treats alike
// (docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string).
func oneLine(body string) string {
	body = strings.ReplaceAll(body, "\r\n", " ")
	body = strings.ReplaceAll(body, "\r", " ")
	return strings.ReplaceAll(body, "\n", " ")
}

// primeData is the `prime` envelope of docs/spec/cmd/prime.md#el-esquema-json.
//
// It carries no rule and no flag name: whoever asks for JSON is a program,
// and a program does not need prose telling it how the writing flags are
// spelled.
func primeData(r *ops.PrimeResult) map[string]any {
	return map[string]any{
		"tool": map[string]any{"name": "biso", "version": Version},
		"board": map[string]any{
			"name":           r.Board.Name,
			"me":             orNull(r.Me),
			"statuses":       list(r.Board.Statuses),
			"initialStatus":  r.Board.InitialStatus,
			"activeStatus":   r.Board.ActiveStatus,
			"terminalStatus": r.Board.TerminalStatus,
			"types":          list(r.Board.Types),
			"priorities":     list(r.Board.Priorities),
			"countByStatus":  r.Board.CountByStatus,
		},
		"inProgress":    primeTaskObjects(r.InProgress),
		"needsAnswer":   primeTaskObjects(r.NeedsAnswer),
		"assignedToYou": primeTaskObjects(r.AssignedToYou),
		"nextUp":        primeTaskObjects(r.NextUp),
		"hiddenCount":   r.HiddenCount,
		"skipped":       list(r.Skipped),
	}
}

// primeTaskObjects is the reduced shape of a task in this envelope: what a
// row of the message shows and nothing else. The body of a question is not
// here, for the same reason `task.list` does not carry the body of a task:
// it is long text, and `biso get --section question` is what reads it.
func primeTaskObjects(views []ops.TaskView) []map[string]any {
	tasks := make([]map[string]any, 0, len(views))
	for _, v := range views {
		t := v.Task
		tasks = append(tasks, map[string]any{
			"id":           t.ID,
			"title":        t.Title,
			"status":       t.Status,
			"type":         orNull(t.Type),
			"priority":     orNull(t.Priority),
			"assignees":    list(t.Assignees),
			"due":          day(t.Due),
			"acDone":       t.AcDone(),
			"acTotal":      t.AcTotal(),
			"urgency":      urgency(v.Urgency),
			"leaseExpired": v.LeaseExpired,
		})
	}
	return tasks
}
