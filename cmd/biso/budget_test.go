package main

import (
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

// budgetRuns is how many times each command is measured. The answer is the
// median, so one run that landed on a scheduling hiccup does not decide
// the verdict and a program that really got slower still does.
const budgetRuns = 5

// TestTheStartupBudgetOfLsAndPrime is the prueba de la suite that section
// asks for, and not an aspiration.
//
// The composition of the board is not fixed by the specification, which
// says so and why: no command branches on what a task holds. This one has
// tasks in every block of `biso prime` so that neither command is measured
// over a shape it would never see.
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
func TestTheStartupBudgetOfLsAndPrime(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 20, waiting: 10, assigned: 30})

	for _, command := range []string{"ls", "prime"} {
		// One call before the clock starts, so that what is measured is a
		// command on a warm machine and not the first read of a file
		// nobody had opened yet.
		m.run(t, command).assertCode(t, 0)

		elapsed := medianOf(measure(t, m, command))
		t.Logf("biso %s over %d tasks took %s (budget %dms)",
			command, blockBoardTasks, elapsed, budgetMillis)
		if raceDetector {
			continue
		}
		if elapsed > budgetMillis*time.Millisecond {
			t.Errorf(
				"biso %s over %d tasks took %s, want less than %s (docs/spec/presupuestos.md#el-presupuesto-de-arranque)",
				command, blockBoardTasks, elapsed, budgetMillis*time.Millisecond)
		}
	}
	if raceDetector {
		t.Skipf(
			"measured under the race detector, which loads the machine around the process being timed; "+
				"the %dms of docs/spec/presupuestos.md#el-presupuesto-de-arranque are checked by the build without it",
			budgetMillis)
	}
}

// measure runs one command a few times and answers how long each run took,
// from before the process was started to after it exited.
func measure(t *testing.T, m *machine, command string) []time.Duration {
	t.Helper()
	times := make([]time.Duration, 0, budgetRuns)
	for i := 0; i < budgetRuns; i++ {
		start := time.Now()
		got := m.run(t, command)
		times = append(times, time.Since(start))
		got.assertCode(t, 0)
	}
	return times
}

func medianOf(times []time.Duration) time.Duration {
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	return times[len(times)/2]
}
