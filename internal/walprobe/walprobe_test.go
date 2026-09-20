package walprobe

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAnOrdinaryLocalDirectoryIsSafe is the guard against the one way this
// probe can be wrong that nobody would notice: answering false everywhere.
// A false negative is invisible from the outside, because the warning it
// raises is worded as a property of the filesystem and reads perfectly
// plausible on any machine, so it has to be a test and not a reading of the
// code.
//
// The temporary directory of a test is an ordinary local filesystem, which
// is exactly where WAL mode is safe.
func TestAnOrdinaryLocalDirectoryIsSafe(t *testing.T) {
	if !Safe(t.TempDir()) {
		t.Error("the probe calls an ordinary local directory unsafe for WAL")
	}
}

// TestTheProbeLeavesNothingBehind checks step 5 of the protocol: the test
// file is deleted whatever the answer was.
func TestTheProbeLeavesNothingBehind(t *testing.T) {
	dir := t.TempDir()
	Safe(dir)

	left, err := filepath.Glob(filepath.Join(dir, ".biso-wal-probe-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(left) > 0 {
		t.Errorf("the probe left %v behind", left)
	}
}

// TestADirectoryThatCannotBeWrittenIsNotSafe is step 1: a filesystem where
// biso cannot even create a file is one where WAL cannot work, and the
// probe stops there instead of going on to the other two steps.
func TestADirectoryThatCannotBeWrittenIsNotSafe(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes into a directory with no write permission")
	}
	dir := filepath.Join(t.TempDir(), "read-only")
	if err := os.Mkdir(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	if Safe(dir) {
		t.Error("the probe calls a directory it cannot write into safe for WAL")
	}
}

// TestADirectoryThatIsNotThereIsNotSafe is the same step for a path that
// does not exist at all, which is what a board directory looks like after
// somebody moved it.
func TestADirectoryThatIsNotThereIsNotSafe(t *testing.T) {
	if Safe(filepath.Join(t.TempDir(), "absent")) {
		t.Error("the probe calls a directory that is not there safe for WAL")
	}
}
