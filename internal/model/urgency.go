package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// UrgencyCoefficients holds the seven configurable coefficients of
// docs/spec/modelo-de-datos/urgencia.md, one per term of the formula. The
// weight of each priority is deliberately not one of them: it comes from
// the position in the configured `priorities` list, never from the name.
type UrgencyCoefficients struct {
	Priority float64 // urgency.priority
	Active   float64 // urgency.active
	Blocking float64 // urgency.blocking
	Blocked  float64 // urgency.blocked
	Due      float64 // urgency.due
	Criteria float64 // urgency.criteria
	Age      float64 // urgency.age
}

// DefaultUrgencyCoefficients are the values the specification gives for a
// board that has not configured them. The stability contract lets them
// change between minor versions, which is why they are data and not
// constants sprinkled through the formula.
var DefaultUrgencyCoefficients = UrgencyCoefficients{
	Priority: 6.0,
	Active:   4.0,
	Blocking: 8.0,
	Blocked:  -5.0,
	Due:      12.0,
	Criteria: 1.0,
	Age:      0.5,
}

// UrgencyContext is everything the formula needs that the task does not
// carry itself. Three of the five come from the board's configuration and
// two from the rest of the board:
//
//   - Priorities is the configured vocabulary, in order, most urgent first.
//   - ActiveStatus and TerminalStatus are the two status roles the formula
//     asks about. They are compared for equality: the task's stored status
//     is already a canonical value of the vocabulary, and turning what the
//     caller typed into one of those is internal/match's job.
//   - Blocking and Blocked are the two terms that depend on other tasks:
//     whether some unfinished task depends on this one, and whether this
//     one depends on some unfinished task. They are arguments and not
//     fields of Task because neither is stored, and computing them needs
//     the whole board (docs/spec/modelo-de-datos/index.md#los-campos-derivados).
//   - Today is the instant the reading happens. Only its UTC calendar date
//     matters, so the same board read at the same real instant from two
//     time zones gives the same urgency.
type UrgencyContext struct {
	Coefficients   UrgencyCoefficients
	Priorities     []string
	ActiveStatus   string
	TerminalStatus string
	Blocking       bool
	Blocked        bool
	Today          time.Time
}

// Urgency computes the derived field of the same name, exactly as
// docs/spec/modelo-de-datos/urgencia.md writes it. It is never stored: the
// result is recomputed on every read.
//
// It answers an error, and not a number alone, for a task whose priority
// is not one of the ones the board configures. `priority` is a closed
// vocabulary, and the rule of the project is that a value that is not in
// one is an error whether it is being written or read, so there is no
// number to give: such a task is the undecodable task of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// exactly like one carrying an extension key the configuration has
// dropped. Which of the two things happens next, failing with exit code 3
// on a targeted read or skipping the task with a warning on a set read, is
// the caller's decision and not this function's.
//
// That check comes before the shortcut for a terminal task, whose urgency
// is 0.0 with no term computed: a task the board cannot interpret is
// unreadable whatever its status, and hiding it behind a zero would be the
// silence the rule is there to prevent.
func (t *Task) Urgency(ctx UrgencyContext) (float64, *Error) {
	weight, ok := priorityWeight(t.Priority, ctx.Priorities)
	if !ok {
		return 0.0, t.unknownPriority(ctx.Priorities)
	}
	if t.Status == ctx.TerminalStatus {
		return 0.0, nil
	}

	c := ctx.Coefficients
	today := ctx.Today.UTC()

	active := 0.0
	if t.Status == ctx.ActiveStatus && t.Question == nil {
		active = 1.0
	}
	blocking := 0.0
	if ctx.Blocking {
		blocking = 1.0
	}
	blocked := 0.0
	if ctx.Blocked {
		blocked = 1.0
	}
	criteria := 0.0
	if len(t.AcceptanceCriteria) > 0 {
		criteria = 1.0
	}
	age := float64(ageInDays(t.CreatedAt, today)) / 30.0
	if age > 4.0 {
		age = 4.0
	}

	sum := c.Priority*weight +
		c.Active*active +
		c.Blocking*blocking +
		c.Blocked*blocked +
		c.Due*proximity(t.Due, today) +
		c.Criteria*criteria +
		c.Age*age

	return roundToOneDecimal(sum), nil
}

// priorityWeight is the priority rule of
// docs/spec/modelo-de-datos/urgencia.md: the weight falls linearly from 1
// at the most urgent level to 0 at the least urgent one, whatever the
// names are and however many there are. A task with no priority weighs
// 0.3, whatever the vocabulary and its size.
//
// It answers false for a priority the vocabulary does not contain, which
// is not the same case as having none: the specification gives that 0.3
// to the task that was never given a priority, and giving it as well to a
// priority the board no longer has would turn a task nobody can interpret
// into an ordinary one without a word. The one way to reach it is a
// vocabulary that changed under stored data, which is precisely the
// undecodable task of docs/spec/garantias.md.
func priorityWeight(priority string, priorities []string) (float64, bool) {
	if priority == "" {
		return 0.3, true
	}
	index := -1
	for i, p := range priorities {
		if p == priority {
			index = i
			break
		}
	}
	if index < 0 {
		return 0.0, false
	}
	if len(priorities) == 1 {
		return 1.0, true
	}
	return 1.0 - float64(index)/float64(len(priorities)-1), true
}

// unknownPriority builds the error of the first case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// exit code 3 with the code undecodable_task of
// docs/spec/contrato-json.md#los-identificadores-de-error. The
// specification fixes no literal text for it, so the message names the
// reason and the value, which is what that section asks for ("con el
// motivo exacto"); the layer that prints it is free to word it otherwise
// the day the specification says how.
func (t *Task) unknownPriority(priorities []string) *Error {
	valid := make([]string, len(priorities))
	copy(valid, priorities)
	return &Error{
		ExitCode: 3,
		Code:     "undecodable_task",
		Message: fmt.Sprintf(
			"%s cannot be read: its priority is %q, which this board does not configure",
			t.ID, t.Priority,
		),
		Field: "priority",
		Given: t.Priority,
		Valid: valid,
	}
}

// proximity is the due-date rule of
// docs/spec/modelo-de-datos/urgencia.md, clamped to [0, 1]. An overdue
// task saturates at 1.0 and therefore scores no more than one due today:
// telling them apart is the job of the --overdue filter, not of a term
// with no ceiling.
func proximity(due, today time.Time) float64 {
	if due.IsZero() {
		return 0.0
	}
	days := calendarDaysBetween(today, due)
	p := (30.0 - float64(days)) / 30.0
	if p < 0.0 {
		return 0.0
	}
	if p > 1.0 {
		return 1.0
	}
	return p
}

// ageInDays is the age term's numerator: the whole number of calendar days
// between the day the task was created and today, both in UTC, looking at
// neither one's time of day.
func ageInDays(createdAt, today time.Time) int {
	if createdAt.IsZero() {
		return 0
	}
	days := calendarDaysBetween(createdAt, today)
	if days < 0 {
		return 0
	}
	return days
}

// calendarDaysBetween answers how many whole days separate the UTC
// calendar date of from and the one of to, positive when to is later.
func calendarDaysBetween(from, to time.Time) int {
	a := time.Date(from.UTC().Year(), from.UTC().Month(), from.UTC().Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(to.UTC().Year(), to.UTC().Month(), to.UTC().Day(), 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}

// roundToOneDecimal rounds half away from zero, as
// docs/spec/modelo-de-datos/urgencia.md fixes it: 1.45 gives 1.5 and 15.35
// gives 15.4.
//
// It rounds the decimal expansion and not the binary one, because that is
// exactly the operation the specification describes and it needs no
// argument about which floats land where: formatting with 'f' and -1 gives
// the shortest decimal that reads back as the same float, "15.35", and the
// digit after the tenths decides, with no multiplication in between that
// could move the value across the half.
//
// The obvious math.Round(x*10)/10 is not known to be wrong here. It gets
// both of the specification's examples right (15.35*10 is exactly 153.5 in
// float64, and 1.45*10 is exactly 14.5), and comparing the two over every
// hundredth from -1000 to 4000 and over every multiple of 1/3000 up to
// 1333, which covers the values the formula can reach, gives no
// discrepancy at all. So this is a choice of the safer construction and
// not the fix of a measured bug.
func roundToOneDecimal(x float64) float64 {
	s := strconv.FormatFloat(x, 'f', -1, 64)
	negative := strings.HasPrefix(s, "-")
	if negative {
		s = s[1:]
	}

	dot := strings.IndexByte(s, '.')
	if dot < 0 || len(s)-dot-1 <= 1 {
		return x // already at most one decimal, so there is nothing to round
	}

	tenths, err := strconv.ParseInt(s[:dot]+s[dot+1:dot+2], 10, 64)
	if err != nil {
		return x // an infinity or a NaN, which the formula cannot produce
	}
	// The remainder is 0.<the digits after the tenths>, so it reaches one
	// half exactly when its first digit does: away from zero on a tie.
	if s[dot+2] >= '5' {
		tenths++
	}

	rounded := float64(tenths) / 10.0
	if negative {
		return -rounded
	}
	return rounded
}
