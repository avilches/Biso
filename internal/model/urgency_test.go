package model

import (
	"math"
	"testing"
	"time"
)

// defaultContext is the board of the specification's examples: three
// priorities, "To Do" as the first status, "In Progress" as the active one
// and "Done" as the terminal one.
func defaultContext(today time.Time) UrgencyContext {
	return UrgencyContext{
		Coefficients:   DefaultUrgencyCoefficients,
		Priorities:     []string{"high", "medium", "low"},
		ActiveStatus:   "In Progress",
		TerminalStatus: "Done",
		Today:          today,
	}
}

func TestUrgencyOfTheWorkedExampleOfTheSpecification(t *testing.T) {
	today := time.Date(2026, 9, 6, 13, 31, 9, 0, time.UTC)
	task := &Task{
		ID:        "MYP-11",
		Status:    "In Progress",
		Priority:  "high",
		CreatedAt: today,
	}
	task.AddCriterion("The diff ignores CRLF")
	task.AddCriterion("There is a test that covers it")

	ctx := defaultContext(today)
	ctx.Blocking = true

	// docs/spec/modelo-de-datos/urgencia.md: 6.0 + 4.0 + 8.0 + 0.0 + 0.0 +
	// 1.0 + 0.0 = 19.0.
	if got := task.Urgency(ctx); got != 19.0 {
		t.Fatalf("Urgency() = %v, want 19.0", got)
	}
}

func TestUrgencyOfATerminalTaskIsZeroAndNothingElseIsComputed(t *testing.T) {
	today := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	task := &Task{
		ID:        "MYP-11",
		Status:    "Done",
		Priority:  "high",
		Due:       today.AddDate(0, 0, -100),
		CreatedAt: today.AddDate(-1, 0, 0),
	}
	task.AddCriterion("one")

	ctx := defaultContext(today)
	ctx.Blocking = true
	if got := task.Urgency(ctx); got != 0.0 {
		t.Fatalf("Urgency() of a terminal task = %v, want 0.0", got)
	}
}

func TestUrgencyPriorityWeightComesFromThePositionInTheVocabulary(t *testing.T) {
	cases := []struct {
		priorities []string
		priority   string
		want       float64
	}{
		{[]string{"high", "medium", "low"}, "high", 1.0},
		{[]string{"high", "medium", "low"}, "medium", 0.5},
		{[]string{"high", "medium", "low"}, "low", 0.0},
		{[]string{"p0", "p1", "p2", "p3"}, "p0", 1.0},
		{[]string{"p0", "p1", "p2", "p3"}, "p1", 2.0 / 3.0},
		{[]string{"p0", "p1", "p2", "p3"}, "p2", 1.0 / 3.0},
		{[]string{"p0", "p1", "p2", "p3"}, "p3", 0.0},
		{[]string{"only"}, "only", 1.0},
		{[]string{"high", "medium", "low"}, "", 0.3},
		{[]string{"only"}, "", 0.3},
	}
	for _, c := range cases {
		got := priorityWeight(c.priority, c.priorities)
		if math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("priorityWeight(%q, %v) = %v, want %v", c.priority, c.priorities, got, c.want)
		}
	}
}

func TestUrgencyProximityIsClampedAndAnOverdueTaskDoesNotScoreMore(t *testing.T) {
	today := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		due  time.Time
		want float64
	}{
		{time.Time{}, 0.0},
		{today.AddDate(0, 0, 60), 0.0},
		{today.AddDate(0, 0, 30), 0.0},
		{today.AddDate(0, 0, 15), 0.5},
		{today, 1.0},
		{today.AddDate(0, 0, -40), 1.0},
	}
	for _, c := range cases {
		got := proximity((&Task{Due: c.due}).Due, today)
		if math.Abs(got-c.want) > 1e-9 {
			t.Fatalf("proximity(%v) = %v, want %v", c.due, got, c.want)
		}
	}
}

func TestUrgencyAgeIsAWholeNumberOfCalendarDaysCappedAtFour(t *testing.T) {
	// docs/spec/modelo-de-datos/urgencia.md: a task created at
	// 2026-08-07T23:50:00Z and read twenty minutes later has an age of one
	// day, not a fraction of one.
	created := time.Date(2026, 8, 7, 23, 50, 0, 0, time.UTC)
	today := time.Date(2026, 8, 8, 0, 10, 0, 0, time.UTC)
	if got := ageInDays(created, today); got != 1 {
		t.Fatalf("ageInDays = %d, want 1", got)
	}

	task := &Task{Status: "To Do", CreatedAt: today.AddDate(-2, 0, 0)}
	ctx := defaultContext(today)
	// Only the age term and the default priority contribute:
	// 6.0*0.3 + 0.5*4.0 = 1.8 + 2.0 = 3.8.
	if got := task.Urgency(ctx); got != 3.8 {
		t.Fatalf("Urgency() = %v, want 3.8", got)
	}
}

func TestUrgencyActiveNeedsTheActiveStatusAndNoOpenQuestion(t *testing.T) {
	today := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	ctx := defaultContext(today)

	active := &Task{Status: "In Progress", CreatedAt: today}
	// 6.0*0.3 + 4.0 = 5.8.
	if got := active.Urgency(ctx); got != 5.8 {
		t.Fatalf("Urgency() of an active task = %v, want 5.8", got)
	}

	active.Question = &Question{Author: "@sara", AskedAt: today, Body: "Which encoding?"}
	// 6.0*0.3 = 1.8.
	if got := active.Urgency(ctx); got != 1.8 {
		t.Fatalf("Urgency() of an active task with an open question = %v, want 1.8", got)
	}
}

func TestUrgencyBlockedSubtracts(t *testing.T) {
	today := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	task := &Task{Status: "To Do", Priority: "low", CreatedAt: today}
	ctx := defaultContext(today)
	ctx.Blocked = true
	// 6.0*0.0 - 5.0 = -5.0.
	if got := task.Urgency(ctx); got != -5.0 {
		t.Fatalf("Urgency() of a blocked task = %v, want -5.0", got)
	}
}

func TestUrgencyRoundsHalfAwayFromZero(t *testing.T) {
	// docs/spec/modelo-de-datos/urgencia.md gives both of these examples,
	// and both of them are the cases the binary representation gets wrong
	// when the rounding is done by multiplying by ten.
	cases := []struct {
		in, want float64
	}{
		{1.45, 1.5},
		{15.35, 15.4},
		{1.44, 1.4},
		{-1.45, -1.5},
		{-15.35, -15.4},
		{19.0, 19.0},
		{0.0, 0.0},
		{2.0 / 3.0, 0.7},
	}
	for _, c := range cases {
		if got := roundToOneDecimal(c.in); got != c.want {
			t.Fatalf("roundToOneDecimal(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestUrgencyCoefficientsAreTheSevenOfTheSpecification(t *testing.T) {
	c := DefaultUrgencyCoefficients
	if c.Priority != 6.0 || c.Active != 4.0 || c.Blocking != 8.0 || c.Blocked != -5.0 ||
		c.Due != 12.0 || c.Criteria != 1.0 || c.Age != 0.5 {
		t.Fatalf("DefaultUrgencyCoefficients = %+v", c)
	}
}
