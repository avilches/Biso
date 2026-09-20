package main

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the conditional claim of an expired lease
// (docs/spec/cmd/verbos-del-ciclo.md#biso-start), exercised from the
// compiled program and not from a function: what makes the claim a claim is
// that it is checked inside the transaction that writes, so a test that
// never runs two writers over one file proves nothing about it.

// raceWindow is how long the test holds the board's write lock while the
// program runs. The program spends that time waiting at the BEGIN of its
// own transaction, having already read the task, which is the window the
// claim exists for. It is far shorter than the five seconds the program
// waits for the lock (guarantee 5 of docs/spec/garantias.md) and far longer
// than starting a process and reading a board of thirteen tasks.
const raceWindow = time.Second

// The ordinary outcome: nobody else was claiming, so the caller takes the
// expired lease and the write goes through.
func TestStartClaimsAnExpiredLeaseFromTheCommandLine(t *testing.T) {
	m := exampleBoard(t)
	writeLease(t, m, "MYP-11", "@sara", time.Now().Add(-time.Hour))

	m.run(t, "start", "MYP-11").assertCode(t, 0)

	if holder := readLease(t, m, "MYP-11"); holder != "@claude" {
		t.Errorf("leaseHolder = %q, want the caller's", holder)
	}
}

// And the outcome the claim exists for: another claim of the same expired
// lease got there first, so this one writes nothing at all, not even the
// note it also carried, and says so with exit code 8 and the code
// `lease_lost` (docs/spec/contrato-json.md#los-identificadores-de-error).
func TestStartThatLosesTheRaceForAnExpiredLeaseWritesNothing(t *testing.T) {
	m := exampleBoard(t)
	writeLease(t, m, "MYP-11", "@sara", time.Now().Add(-time.Hour))

	b := openTheBoard(t, m)
	defer b.Close()

	var running *process
	// Holding one write transaction open is what puts the program's read
	// and the program's write on either side of somebody else's write,
	// which is the order a real race has and the only order this check
	// can be seen in.
	err := b.Store.WithTx(func(tx *sql.Tx) error {
		running = m.start(t, "start", "MYP-11", "--append-note", "Taken over", "--json")
		time.Sleep(raceWindow)
		_, err := tx.Exec(
			"UPDATE task SET lease_holder = ?, lease_expires_at = ? WHERE id = ?",
			"@sara", time.Now().Add(time.Hour).UTC().Format(model.InstantLayout), "MYP-11")
		return err
	})
	if err != nil {
		t.Fatal(err)
	}

	got := running.wait(t).assertCode(t, 8)

	envelope := envelopeOf(t, got.stderr)
	failure, ok := envelope["error"].(map[string]any)
	if !ok {
		t.Fatalf("the envelope carries no error: %s", got.stderr)
	}
	if failure["code"] != "lease_lost" {
		t.Errorf("code = %v, want lease_lost", failure["code"])
	}
	if holder := readLease(t, m, "MYP-11"); holder != "@sara" {
		t.Errorf("leaseHolder = %q, want the winner's", holder)
	}
	// Nothing of the losing call survived, and the note it carried is the
	// field that says so: the whole transaction was rolled back.
	if notes := m.run(t, "get", "MYP-11", "--section", "notes").assertCode(t, 0).stdout; notes != "" {
		t.Errorf("notes = %q, and the losing call wrote nothing", notes)
	}
}

// writeLease puts a lease straight into the database, which is the one
// state no command can produce on its own: an expired one.
func writeLease(t *testing.T, m *machine, id, holder string, expiresAt time.Time) {
	t.Helper()
	b := openTheBoard(t, m)
	defer b.Close()
	if _, err := b.Store.Exec(
		"UPDATE task SET lease_holder = ?, lease_expires_at = ? WHERE id = ?",
		holder, expiresAt.UTC().Format(model.InstantLayout), id); err != nil {
		t.Fatal(err)
	}
}

func readLease(t *testing.T, m *machine, id string) string {
	t.Helper()
	b := openTheBoard(t, m)
	defer b.Close()
	task, err := b.Tasks.Load(id)
	if err != nil {
		t.Fatal(err)
	}
	return task.LeaseHolder
}

// openTheBoard opens the one board of a machine from inside the test, which
// is how a fixture is written and how a second writer is played.
func openTheBoard(t *testing.T, m *machine) *board.Board {
	t.Helper()
	dirs, err := filepath.Glob(filepath.Join(m.home, ".biso", "boards", "*"))
	if err != nil || len(dirs) != 1 {
		t.Fatalf("the machine has %d boards, and it should have one: %v", len(dirs), err)
	}
	b, err := board.Open(&board.Location{Dir: dirs[0], ID: board.MarkerID(dirs[0])}, board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// process is one run of the program that is still going, which is what a
// race needs and m.run cannot give: m.run waits for the program to finish.
type process struct {
	cmd    *exec.Cmd
	stdout *os.File
	stderr *os.File
}

// start launches the program in this machine and answers before it has
// finished. The two streams go to files, because a pipe nobody is reading
// fills up and stops the process.
func (m *machine) start(t *testing.T, argv ...string) *process {
	t.Helper()
	cmd := exec.Command(binary(t), argv...)
	cmd.Dir = m.dir
	cmd.Env = append(os.Environ(), "HOME="+m.home)
	for name, value := range m.env {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	p := &process{cmd: cmd, stdout: streamFile(t, "stdout"), stderr: streamFile(t, "stderr")}
	cmd.Stdout, cmd.Stderr = p.stdout, p.stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("biso could not be started: %v", err)
	}
	return p
}

// wait is the rest of m.run: what the process wrote and what it exited
// with, once it is over.
func (p *process) wait(t *testing.T) call {
	t.Helper()
	code := 0
	if err := p.cmd.Wait(); err != nil {
		exit, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("biso could not be run: %v", err)
		}
		code = exit.ExitCode()
	}
	return call{code: code, stdout: readStream(t, p.stdout), stderr: readStream(t, p.stderr)}
}

func streamFile(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Create(filepath.Join(t.TempDir(), name))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func readStream(t *testing.T, f *os.File) string {
	t.Helper()
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
