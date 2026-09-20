package ops

import (
	"os"
	"path/filepath"
	"testing"

	"biso/internal/board"
)

// TestSnapshotLeavesThePreviousTwoFilesWhenOneCannotBeWritten is the row of
// the case table of docs/spec/cmd/snapshot.md about not being able to write
// one of the two files: exit code 8, the two previous ones intact, and no
// revision attempted.
//
// It drives the writing step on its own and not the whole command, because
// what makes the promise true is the order inside that step: both temporary
// files are complete before either is renamed, so a failure while writing
// them leaves the pair that was there untouched. The second half, that no
// revision is attempted, is checked from outside in
// TestSnapshotDoesNotCommitWhenItCannotWriteAFile.
func TestSnapshotLeavesThePreviousTwoFilesWhenOneCannotBeWritten(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("a directory that cannot be written to does not stop the superuser")
	}
	dir := t.TempDir()
	before := map[string]string{
		board.SnapshotTasksFile:  "{\"id\":\"MYP-1\"}\n",
		board.SnapshotConfigFile: "{\n  \"task_prefix\": \"MYP\"\n}\n",
	}
	for name, content := range before {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	err := writeSnapshotFiles(dir, []byte("{\"id\":\"MYP-2\"}\n"), []byte("{}\n"))

	if err == nil {
		t.Fatal("writing into a directory that cannot be written to worked")
	}
	if err.ExitCode != 8 {
		t.Errorf("error = %d/%s (%s), want exit code 8", err.ExitCode, err.Code, err.Message)
	}
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range before {
		got, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil {
			t.Fatalf("%s is gone: %v", name, readErr)
		}
		if string(got) != content {
			t.Errorf("%s changed:\n got %q\nwant %q", name, got, content)
		}
	}
	// And nothing else was left behind either: the temporary files of a
	// failed write are removed.
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != len(before) {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("the directory holds %v and it held two files", names)
	}
}
