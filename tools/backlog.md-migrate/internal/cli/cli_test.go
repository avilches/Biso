package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(args ...string) (stdout, stderr string, code int) {
	var out, err bytes.Buffer
	code = Run(args, &out, &err)
	return out.String(), err.String(), code
}

func TestImportValidArgsMinimal(t *testing.T) {
	opts, code, ok := ParseImportArgs([]string{"./backlog", "--project", "."}, &bytes.Buffer{}, &bytes.Buffer{})
	if !ok {
		t.Fatalf("expected ok=true, got exit code %d", code)
	}
	want := ImportOptions{BacklogDir: "./backlog", Project: ".", Out: "-", Biso: "biso", Strict: false}
	if opts != want {
		t.Fatalf("got %+v, want %+v", opts, want)
	}
}

func TestImportValidArgsAllFlags(t *testing.T) {
	opts, code, ok := ParseImportArgs(
		[]string{"./backlog", "--project", "/tmp/proj", "--out", "tasks.ndjson", "--biso", "/usr/local/bin/biso", "--strict"},
		&bytes.Buffer{}, &bytes.Buffer{},
	)
	if !ok {
		t.Fatalf("expected ok=true, got exit code %d", code)
	}
	want := ImportOptions{
		BacklogDir: "./backlog",
		Project:    "/tmp/proj",
		Out:        "tasks.ndjson",
		Biso:       "/usr/local/bin/biso",
		Strict:     true,
	}
	if opts != want {
		t.Fatalf("got %+v, want %+v", opts, want)
	}
}

func TestImportValidArgsFlagsBeforePositional(t *testing.T) {
	opts, code, ok := ParseImportArgs([]string{"--project", ".", "./backlog"}, &bytes.Buffer{}, &bytes.Buffer{})
	if !ok {
		t.Fatalf("expected ok=true, got exit code %d", code)
	}
	if opts.BacklogDir != "./backlog" || opts.Project != "." {
		t.Fatalf("got %+v", opts)
	}
}

func TestImportMissingBacklogDir(t *testing.T) {
	_, code, ok := ParseImportArgs([]string{"--project", "."}, &bytes.Buffer{}, &bytes.Buffer{})
	if ok {
		t.Fatal("expected ok=false")
	}
	if code != 2 {
		t.Fatalf("got exit code %d, want 2", code)
	}
}

func TestImportMissingProject(t *testing.T) {
	_, code, ok := ParseImportArgs([]string{"./backlog"}, &bytes.Buffer{}, &bytes.Buffer{})
	if ok {
		t.Fatal("expected ok=false")
	}
	if code != 2 {
		t.Fatalf("got exit code %d, want 2", code)
	}
}

func TestImportUnknownFlag(t *testing.T) {
	_, code, ok := ParseImportArgs([]string{"./backlog", "--project", ".", "--bogus"}, &bytes.Buffer{}, &bytes.Buffer{})
	if ok {
		t.Fatal("expected ok=false")
	}
	if code != 2 {
		t.Fatalf("got exit code %d, want 2", code)
	}
}

func TestImportHelpLong(t *testing.T) {
	var out, errBuf bytes.Buffer
	_, code, ok := ParseImportArgs([]string{"--help"}, &out, &errBuf)
	if ok {
		t.Fatal("expected ok=false for --help")
	}
	if code != 0 {
		t.Fatalf("got exit code %d, want 0", code)
	}
	if errBuf.Len() != 0 {
		t.Fatalf("expected nothing on stderr, got %q", errBuf.String())
	}
	assertHelpText(t, out.String())
}

func TestImportHelpShort(t *testing.T) {
	var out, errBuf bytes.Buffer
	_, code, ok := ParseImportArgs([]string{"-h"}, &out, &errBuf)
	if ok {
		t.Fatal("expected ok=false for -h")
	}
	if code != 0 {
		t.Fatalf("got exit code %d, want 0", code)
	}
	assertHelpText(t, out.String())
}

func assertHelpText(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(text, "backlog.md-migrate import <backlog-dir> --project <dir> [--out <file|->] [--biso <path>] [--strict]") {
		t.Fatalf("help text is missing the signature:\n%s", text)
	}
	for _, flagName := range []string{"--project", "--out", "--biso", "--strict"} {
		if !strings.Contains(text, flagName) {
			t.Fatalf("help text is missing flag %s:\n%s", flagName, text)
		}
	}
}

func TestRunImportNotImplementedYet(t *testing.T) {
	_, errOut, code := run("import", "./backlog", "--project", ".")
	if code != 1 {
		t.Fatalf("got exit code %d, want 1", code)
	}
	if !strings.Contains(errOut, "not implemented yet") {
		t.Fatalf("expected a not-implemented message on stderr, got %q", errOut)
	}
}

func TestRunNoSubcommand(t *testing.T) {
	_, errOut, code := run()
	if code != 2 {
		t.Fatalf("got exit code %d, want 2", code)
	}
	if errOut == "" {
		t.Fatal("expected a usage message on stderr")
	}
}

func TestRunUnknownSubcommand(t *testing.T) {
	_, errOut, code := run("export")
	if code != 2 {
		t.Fatalf("got exit code %d, want 2", code)
	}
	if errOut == "" {
		t.Fatal("expected a usage message on stderr")
	}
}
