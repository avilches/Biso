package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These are the golden tests of the batch of docs/spec/cmd/new.md#el-modo-lote.
// Its two blocks speak of a file of two hundred and forty two lines, five of
// them invalid for five different reasons and at five given line numbers, so
// the file the tests build is exactly that file: the numbers of the message
// are not adjusted to whatever came out.
//
// One of the five is a failure of the graph, which is found only once the
// whole file has been read, and it sits between the other two that surround
// it: that is the order of the file the block promises.

// batchBoard is a board with the prefix the examples of the batch use.
func batchBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	return m
}

// twoHundredAndFortyTwoLines writes the file of the examples: two hundred
// and forty two tasks, with the ones the failing example names replaced by
// the bad lines it shows, each one at its own line number.
func twoHundredAndFortyTwoLines(bad map[int]string) string {
	var b strings.Builder
	for line := 1; line <= 242; line++ {
		if text, ok := bad[line]; ok {
			b.WriteString(text + "\n")
			continue
		}
		fmt.Fprintf(&b, "{\"title\":\"Task %d\"}\n", line)
	}
	return b.String()
}

func TestBatchDryRunPrintsTheCountOfTheSpecification(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, twoHundredAndFortyTwoLines(nil))

	got := m.run(t, "new", "--from", path, "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "new-batch-dry-run.txt"),
		"the preview line of biso new --from")
	assertEqual(t, got.stdout, "", "the standard output of a preview, which creates nothing")
	// A preview writes nothing, and the board is the proof.
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a preview left behind")
}

func TestBatchListsEveryInvalidLineOfTheSpecification(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, twoHundredAndFortyTwoLines(map[int]string{
		12:  `{"id":"OTHER-5","title":"From another board"}`,
		47:  `{"title":"Wrong status","status":"Pendiente"}`,
		88:  `{"title":"Unknown key","trelloCard":"5f2a8c1e"}`,
		130: `{"title":"Nowhere to hang from","parent":"MYP-900"}`,
		201: `{"title":"   "}`,
	}))

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	assertEqual(t, got.stderr, fixture(t, "new-batch-invalid.txt"),
		"the failures of an invalid batch")
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks an invalid batch left behind")
}

func TestBatchPrintsOneIdentifierPerTask(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, strings.Join([]string{
		`# a comment line, which is ignored`,
		``,
		`{"id":"MYP-101","title":"One"}`,
		`{"id":"MYP-102","title":"Two"}`,
		`{"id":"MYP-103","title":"Three"}`,
		``,
	}, "\n"))

	got := m.run(t, "new", "--from", path).assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "new-batch-ids.txt"),
		"the identifiers of a batch")
	// The counter ends above the highest identifier the batch reserved, so
	// the next task created by hand never collides with one of them.
	assertEqual(t, m.run(t, "new", "Next").assertCode(t, 0).stdout, "MYP-104\n",
		"the identifier after a batch that reserved up to MYP-103")
}

func TestBatchRefusesTheTwoHalvesOfALeaseSeparately(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	var lines []string
	for i := 1; i <= 31; i++ {
		switch i {
		case 14:
			lines = append(lines, `{"title":"Not active","leaseHolder":"@claude",`+
				`"leaseExpiresAt":"2026-09-08T14:00:00Z"}`)
		case 31:
			lines = append(lines, `{"title":"Half a lease","status":"In Progress",`+
				`"assignees":["@claude"],"leaseHolder":"@claude"}`)
		default:
			lines = append(lines, fmt.Sprintf(`{"title":"Task %d"}`, i))
		}
	}
	m.write(t, path, strings.Join(lines, "\n")+"\n")

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	// The two sentences of the block of the specification, each one on the
	// line of the file it belongs to.
	for _, want := range strings.Split(strings.TrimRight(fixture(t, "new-batch-lease.txt"), "\n"), "\n") {
		if !strings.Contains(got.stderr, want) {
			t.Errorf("the failures do not carry %q:\n%s", want, got.stderr)
		}
	}
}

// runWithStdin runs the compiled program with something on its standard
// input, which is what `--from -` reads.
func runWithStdin(t *testing.T, m *machine, stdin string, argv ...string) call {
	t.Helper()
	cmd := exec.Command(binary(t), argv...)
	cmd.Dir = m.dir
	cmd.Env = append(os.Environ(), "HOME="+m.home)
	for name, value := range m.env {
		cmd.Env = append(cmd.Env, name+"="+value)
	}
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("biso %s: %v", strings.Join(argv, " "), err)
	}
	return call{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func TestBatchReadsStandardInput(t *testing.T) {
	m := batchBoard(t)

	got := runWithStdin(t, m, "{\"title\":\"From a pipe\"}\n", "new", "--from", "-")

	got.assertCode(t, 0)
	assertEqual(t, got.stdout, "MYP-1\n", "the identifier of a task read from stdin")
}

func TestBatchRefusesPrintAndAFileThatIsNotUTF8(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, `{"title":"One"}`+"\n")
	m.run(t, "new", "--from", path, "--print").assertCode(t, 2)

	broken := filepath.Join(m.dir, "broken.ndjson")
	m.write(t, broken, "{\"title\":\"\xff\"}\n")
	m.run(t, "new", "--from", broken).assertCode(t, 3)
}

// TestBatchEnvelopeCarriesTheUrgencyAndWhatChanged is the schema of
// docs/spec/cmd/set.md#el-esquema-json applied to the batch: `urgency` is a
// key that is always there and has to be the urgency of the task that was
// just written, and `changed` is what really changed, which on a task that
// did not exist is every field its line carried.
//
// The urgency is compared against the one `biso ls` prints for those same
// tasks, because two numbers that disagree about the same task would be
// worse than one missing key.
func TestBatchEnvelopeCarriesTheUrgencyAndWhatChanged(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, strings.Join([]string{
		`{"id":"MYP-1","title":"Blocking one","priority":"high","status":"In Progress","assignees":["@sara"]}`,
		`{"id":"MYP-2","title":"Blocked one","priority":"low","dependencies":["MYP-1"]}`,
		`{"id":"MYP-3","title":"A plain one"}`,
	}, "\n")+"\n")

	got := m.run(t, "new", "--from", path, "--json").assertCode(t, 0)

	var envelope struct {
		Data struct {
			Tasks []struct {
				ID      string   `json:"id"`
				Urgency float64  `json:"urgency"`
				Changed []string `json:"changed"`
			} `json:"tasks"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the envelope of a batch is not JSON: %v\n%s", err, got.stdout)
	}
	if len(envelope.Data.Tasks) != 3 {
		t.Fatalf("the envelope carries %d tasks and the file had three", len(envelope.Data.Tasks))
	}
	for _, task := range envelope.Data.Tasks {
		if task.Urgency == 0 {
			t.Errorf("%s came out with urgency zero:\n%s", task.ID, got.stdout)
		}
		if !contains(task.Changed, "title") {
			t.Errorf("%s says it changed %v, and a task that did not exist changed its title",
				task.ID, task.Changed)
		}
	}
	// The same numbers the board answers for those same tasks.
	for _, task := range envelope.Data.Tasks {
		card := m.run(t, "get", task.ID, "--json").assertCode(t, 0).stdout
		want := fmt.Sprintf(`"urgency": %.1f`, task.Urgency)
		if !strings.Contains(card, want) {
			t.Errorf("the batch said %s for %s and its card says otherwise:\n%s",
				want, task.ID, card)
		}
	}
}

// TestTheTwoLeaseFailuresCarryNoGiven is the rule of
// docs/spec/contrato-json.md#los-errores-en-json for an error that names a
// field and has no value to quote: `given` does not travel, instead of
// travelling empty. Neither half of the lease invariant is about a value
// that was written wrong; both are about which of the two keys the line
// brought, which is why their sentences do not quote anything either.
func TestTheTwoLeaseFailuresCarryNoGiven(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, `{"title":"Not active","leaseHolder":"@claude",`+
		`"leaseExpiresAt":"2026-09-08T14:00:00Z"}`+"\n"+
		`{"title":"Half a lease","status":"In Progress",`+
		`"assignees":["@claude"],"leaseHolder":"@claude"}`+"\n")

	got := m.run(t, "new", "--from", path, "--json").assertCode(t, 7)

	var envelope struct {
		Error struct {
			Details []map[string]any `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(got.stderr), &envelope); err != nil {
		t.Fatalf("the envelope is not JSON: %v\n%s", err, got.stderr)
	}
	if len(envelope.Error.Details) != 2 {
		t.Fatalf("the envelope carries %d failures and not two", len(envelope.Error.Details))
	}
	for _, detail := range envelope.Error.Details {
		if detail["code"] != "invalid_lease" {
			t.Errorf("the failure is %q and not invalid_lease", detail["code"])
		}
		if detail["field"] != "leaseHolder" {
			t.Errorf("the failure names %q and not leaseHolder", detail["field"])
		}
		if _, ok := detail["given"]; ok {
			t.Errorf("the failure carries a given of %#v, and there is none to quote",
				detail["given"])
		}
	}
}
