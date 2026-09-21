package main

import (
	"strings"
	"testing"
	"time"
)

// These tests are about the verdict of the startup budget and not about the
// machine: they feed the functions durations made up here, so they give the
// same answer on an idle laptop and on one with every core busy.

const verdictBudget = budgetMillis * time.Millisecond

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

// sampled answers a sample function that replays the given durations with
// the given exit codes and counts how many times it was called.
func sampled(times []time.Duration, codes []int) (func() (time.Duration, int), *int) {
	calls := 0
	return func() (time.Duration, int) {
		i := calls
		calls++
		code := 0
		if codes != nil {
			code = codes[i]
		}
		return times[i], code
	}, &calls
}

func TestOneFastRunAmongManySlowOnesPasses(t *testing.T) {
	times := make([]time.Duration, budgetRuns)
	for i := range times {
		times[i] = ms(80)
	}
	times[17] = ms(14)
	sample, calls := sampled(times, nil)

	got, err := takeSamples(budgetRuns, verdictBudget, sample)
	if err != nil {
		t.Fatal(err)
	}
	if !summarize(got).fastestWithin(verdictBudget) {
		t.Errorf("a fastest run of 14ms was refused: %s", summarize(got))
	}
	if *calls != 18 || len(got) != 18 {
		t.Errorf("calls = %d, samples = %d, want both 18: it stops at the first run that fits", *calls, len(got))
	}
}

func TestAFirstRunThatFitsIsTheOnlyOneTaken(t *testing.T) {
	sample, calls := sampled([]time.Duration{ms(12), ms(90)}, nil)

	got, err := takeSamples(budgetRuns, verdictBudget, sample)
	if err != nil || len(got) != 1 || *calls != 1 {
		t.Errorf("samples = %v, calls = %d, err = %v, want one run and no error", got, *calls, err)
	}
}

func TestAllTheRunsBeingSlowFails(t *testing.T) {
	times := make([]time.Duration, budgetRuns)
	for i := range times {
		times[i] = ms(26 + i)
	}
	sample, calls := sampled(times, nil)

	got, err := takeSamples(budgetRuns, verdictBudget, sample)
	if err != nil {
		t.Fatal(err)
	}
	if *calls != budgetRuns {
		t.Errorf("calls = %d, want all %d before giving up", *calls, budgetRuns)
	}
	s := summarize(got)
	if s.fastestWithin(verdictBudget) {
		t.Errorf("a fastest run of %s was accepted", s.min)
	}
	want := "min 26ms, median 41ms, max 55ms"
	if s.String() != want {
		t.Errorf("summary = %q, want %q", s.String(), want)
	}
}

func TestTheBudgetItselfIsWithinAndOneMoreMillisecondIsNot(t *testing.T) {
	if !summarize([]time.Duration{verdictBudget}).fastestWithin(verdictBudget) {
		t.Error("exactly 25ms was refused")
	}
	if summarize([]time.Duration{verdictBudget + time.Millisecond}).fastestWithin(verdictBudget) {
		t.Error("26ms was accepted")
	}
	if summarize(nil).fastestWithin(verdictBudget) {
		t.Error("no runs at all was accepted")
	}
}

func TestANonZeroExitCodeFailsEvenWhenTheRunWasFast(t *testing.T) {
	sample, _ := sampled([]time.Duration{ms(90), ms(3)}, []int{0, 2})

	_, err := takeSamples(budgetRuns, verdictBudget, sample)
	if err == nil || !strings.Contains(err.Error(), "run 2 exited with code 2") {
		t.Errorf("err = %v, want it to name run 2 and code 2", err)
	}
}

func TestSummarizeDoesNotReorderItsInput(t *testing.T) {
	in := []time.Duration{ms(30), ms(10), ms(20)}
	s := summarize(in)
	if in[0] != ms(30) || in[1] != ms(10) || in[2] != ms(20) {
		t.Errorf("input reordered: %v", in)
	}
	if s.min != ms(10) || s.median != ms(20) || s.max != ms(30) || s.runs != 3 {
		t.Errorf("summary = %+v", s)
	}
}
