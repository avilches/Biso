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
const budgetRuns = 100

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
// because every run pays for the new code. The reasoning, the figures and
// the discarded alternatives (the median, CPU time, spacing the runs out,
// a looser limit) are in
// docs/decisiones/lenguaje-y-rendimiento.md, section "El presupuesto de
// arranque se mide con la muestra mas rapida". It is wall clock and not CPU
// time on purpose: a wait (a sleep, a disk sync) spends clock and no CPU, so
// a test over CPU time would let a regression made of waiting go through.
//
// Every run has to exit with code 0, and the failure message prints the
// minimum, the median and the maximum of the runs taken, to tell a
// regression (all of them over) from a machine overwhelmed (only some).
//
// How to check it again, which is what acceptance criterion 2 asks of this
// test. The verdict functions have their own deterministic tests in
// budget_verdict_test.go; the wiring of this test to them has no automatic
// test of its own, and is checked by hand as follows. Compile the test
// binary first (`go test -c -o /tmp/budget.test ./cmd/biso`) so that the load
// does not slow the compiler down, and run it from cmd/biso.
//
//   - It must pass under artificial load. Start 32 and then 64 processes
//     running `while :; do :; done` (twice and four times the cores of a
//     16 core machine), wait a second, run the binary with
//     -test.run 'TestTheStartupBudgetOfLsAndPrime$' -test.count=20, and kill
//     every loop afterwards (with a trap, and check with ps). Nearly every
//     count passes: a machine so saturated that none of budgetRuns runs fits
//     fails the test, which is correct, and its message says so with the
//     minimum, the median and the maximum.
//
//   - It must fail with an injected wait. The test builds biso from the
//     source when it runs, so the wait goes in the source and not in the
//     test binary: put `time.Sleep(30 * time.Millisecond)` as the first
//     statement of main() in cmd/biso/main.go, run the test, and revert
//     main.go. Every run is then over the budget and the test fails, on an
//     idle machine and under load, with a minimum over 25ms in the message.
//     A wait of 12 milliseconds must fail as well, on an idle machine and
//     under load. A wait of 5 milliseconds does not fail, which is expected:
//     the budget is watched whole, and not that it does not grow.
//
// The figures of those runs are in the decision above, not here.
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
		t.Logf("biso %s over %d tasks: fastest run %s, after %d taken (budget under %dms)",
			command, blockBoardTasks, s.min, s.runs, budgetMillis)
		if raceDetector {
			continue
		}
		if !s.fastestWithin(budget) {
			t.Errorf(
				"biso %s over %d tasks: none of %d runs took under %s, %s "+
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
// took, stopping as soon as one is under the budget: that run is the fastest
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
		if elapsed < budget {
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

// fastestWithin is the verdict: the fastest run is under the budget, as the
// specification says ("less than 25 milliseconds"), so exactly 25 fails.
// Without a single run there is no verdict in favour.
func (s startupSummary) fastestWithin(budget time.Duration) bool {
	return s.runs > 0 && s.min < budget
}

func (s startupSummary) String() string {
	return fmt.Sprintf("min %s, median %s, max %s", s.min, s.median, s.max)
}
