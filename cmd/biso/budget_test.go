package main

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

// This is the measurement of
// docs/spec/presupuestos.md#el-presupuesto-de-arranque: `biso ls` and
// `biso prime` over a board of three hundred real tasks, each one the
// compiled program run as a process, from the fork to the exit code.
//
// It lives here and not in internal/board because the budget is about
// those two commands and not about one function of one package: whoever
// pays the 25 milliseconds pays for the process starting, the board being
// resolved, its database being opened, its three hundred tasks being read
// and assembled, the urgency of every one of them being computed and the
// output being written. A test over a function would leave most of that
// out. internal/board keeps its own measurement of opening and reading,
// which is the floor of this one and not this one.

// budgetMillis is the figure of that section: 25 milliseconds of wall
// clock on the reference machine, which is the one that runs the project's
// continuous integration suite.
const budgetMillis = 25

// budgetRuns is the most times each command is measured, not a fixed
// number of them. The verdict is the fastest run (see the test below), so
// the measurement stops at the first run that fits the budget and only a
// command that never fits pays for all of them.
const budgetRuns = 30

// TestTheStartupBudgetOfLsAndPrime is the suite test that section asks
// for, and not an aspiration.
//
// What is measured is wall clock, of the whole process, from before it is
// started to after it exits, and the verdict is the FASTEST of up to
// budgetRuns runs, not one run and not the median. A busy machine (another
// process holding the core, the scheduler, a security agent reading every
// file that gets opened) can only add time to a run and never take any
// away, so the fastest run is the best bound on what the program costs by
// itself, and the only one that load cannot worsen while a single run
// escapes it. A real regression raises the fastest run like any other,
// because every run pays for the new code. The median of five runs, which
// this used to be, measured the machine: under load it went over 25
// milliseconds with no change to the program. The reasoning, the figures
// and the discarded alternatives are in
// docs/decisiones/lenguaje-y-rendimiento.md, section "El presupuesto de
// arranque se mide con la muestra mas rapida". It is wall clock and not CPU
// time on purpose: a wait (a sleep, a disk sync) spends clock and no CPU, so
// a test over CPU time would let a regression made of waiting go through.
//
// Every run has to exit with code 0, and the failure message prints the
// minimum, the median and the maximum of the runs taken, to tell a
// regression (all of them over) from a machine overwhelmed (only some).
//
// How it was checked, which is acceptance criterion 2 of the task that
// introduced this (the verdict functions have their own deterministic tests
// in budget_verdict_test.go, and this is the check of the real thing), on the
// development machine (16 cores, with its own security agent already busy):
//
//   - Under artificial load it passes. With 32 and with 64 processes
//     spinning on `while :; do :; done`, the compiled test binary ran with
//     -test.count=5 and then -test.count=20 for each of the two loads. Of 25
//     runs under 32 loops one failed, on `biso prime`, whose 30 runs had a
//     minimum of 25.04ms and a median of 36.0ms; the other 24 passed, and
//     the 25 under 64 loops all passed. When it passes under load the fastest
//     run is typically between 14ms and 25ms, so the margin there is
//     thin, and a run that fails on a saturated machine says so in its
//     message with the minimum, the median and the maximum.
//
//   - With an injected wait it fails. A time.Sleep(30 * time.Millisecond) as
//     the first statement of main(), reverted afterwards, made the test fail
//     on the idle machine, with `biso ls` at min 43.4ms, median 50.1ms, max
//     70.9ms and `biso prime` at min 44.8ms, median 49.2ms, max 87.4ms, in
//     4.5 seconds; and again under 32 loops, with minimums of 45.4ms and
//     50.3ms. Without the wait it passed again, on the first run of each
//     command (12.9ms and 14.3ms). A wait of 5 milliseconds did not make it
//     fail (fastest run 17.8ms to 23.4ms over five runs, all passing), which
//     is expected: the budget is watched whole, not that it does not grow.
//
// A build under the race detector skips it and says so. The process being
// measured is not instrumented, because `go build` compiles it without the
// detector whatever the suite was built with, but everything around it is:
// the suite that starts it, and the other packages running beside it. That
// is a machine other than the idle one the figure describes, and the
// answer would be a measurement of the detector. It does not get a looser
// limit either, because a limit nobody checks is not a limit: the build
// without the detector asserts the figure of the specification, which is
// the only one that can.
//
// The composition of the board is not fixed by the specification, which
// says so and why: no command branches on what a task holds. This one has
// tasks in every block of `biso prime` so that neither command is measured
// over a shape it would never see.
func TestTheStartupBudgetOfLsAndPrime(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 20, waiting: 10, assigned: 30})
	budget := budgetMillis * time.Millisecond

	for _, command := range []string{"ls", "prime"} {
		// One call before the clock starts, so that what is measured is a
		// command on a warm machine and not the first read of a file
		// nobody had opened yet.
		m.run(t, command).assertCode(t, 0)

		var last call
		samples, err := takeSamples(budgetRuns, budget, func() (time.Duration, int) {
			start := time.Now()
			last = m.run(t, command)
			return time.Since(start), last.code
		})
		if err != nil {
			t.Fatalf("biso %s: %v\nstdout:\n%s\nstderr:\n%s", command, err, last.stdout, last.stderr)
		}

		s := summarize(samples)
		t.Logf("biso %s over %d tasks: fastest of %d runs took %s (budget %dms)",
			command, blockBoardTasks, s.runs, s.min, budgetMillis)
		if raceDetector {
			continue
		}
		if !s.fastestWithin(budget) {
			t.Errorf(
				"biso %s over %d tasks: none of %d runs took %s or less, %s "+
					"(docs/spec/presupuestos.md#el-presupuesto-de-arranque)",
				command, blockBoardTasks, s.runs, budget, s)
		}
	}
	if raceDetector {
		t.Skipf(
			"measured under the race detector, which loads the machine around the process being timed; "+
				"the %dms of docs/spec/presupuestos.md#el-presupuesto-de-arranque are checked by the build without it",
			budgetMillis)
	}
}

// takeSamples calls sample up to limit times and answers how long each call
// took, stopping as soon as one fits the budget: that run is the fastest
// that matters and the verdict is already known. sample answers the elapsed
// time and the exit code of the process, and a code other than 0 stops the
// measurement with an error, because the time of a command that failed
// says nothing about the one that works.
func takeSamples(limit int, budget time.Duration, sample func() (time.Duration, int)) ([]time.Duration, error) {
	samples := make([]time.Duration, 0, limit)
	for i := 0; i < limit; i++ {
		elapsed, code := sample()
		if code != 0 {
			return samples, fmt.Errorf("run %d exited with code %d, want 0", i+1, code)
		}
		samples = append(samples, elapsed)
		if elapsed <= budget {
			break
		}
	}
	return samples, nil
}

// startupSummary is what the failure message shows of the runs taken.
type startupSummary struct {
	min, median, max time.Duration
	runs             int
}

// summarize does not reorder the slice it is given.
func summarize(samples []time.Duration) startupSummary {
	if len(samples) == 0 {
		return startupSummary{}
	}
	sorted := append([]time.Duration(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return startupSummary{
		min:    sorted[0],
		median: sorted[len(sorted)/2],
		max:    sorted[len(sorted)-1],
		runs:   len(sorted),
	}
}

// fastestWithin is the verdict: the fastest run is not over the budget.
// Without a single run there is no verdict in favour.
func (s startupSummary) fastestWithin(budget time.Duration) bool {
	return s.runs > 0 && s.min <= budget
}

func (s startupSummary) String() string {
	return fmt.Sprintf("min %s, median %s, max %s", s.min, s.median, s.max)
}
