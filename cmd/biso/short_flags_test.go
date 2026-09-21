package main

import (
	"strings"
	"testing"
)

// This file runs the compiled program against the rule of
// docs/decisiones/comandos-y-flags.md#una-forma-corta-solo-existe-si-nadie-más-reclama-su-inicial:
// a short form exists only if no other flag of the same command starts with
// its letter, so the seven that used to exist are unknown flags now, and the
// call that used to take one in silence fails with its code.

func TestTheSevenRetiredShortFormsAreUnknownFlags(t *testing.T) {
	m := exportBoard(t)

	for _, tc := range []struct {
		short string
		argv  []string
	}{
		{"-a", []string{"new", "A task", "-a", "@sara"}},
		{"-d", []string{"new", "A task", "-d", "text"}},
		{"-l", []string{"ls", "-l", "backend"}},
		{"-o", []string{"export", "-o", "backup.ndjson"}},
		{"-p", []string{"ls", "-p", "MYP-1"}},
		{"-s", []string{"ls", "-s", "Done"}},
		{"-t", []string{"new", "A task", "-t", "feature"}},
	} {
		t.Run(tc.short, func(t *testing.T) {
			got := m.run(t, tc.argv...).assertCode(t, 2)

			assertEqual(t, got.stderr, "error: unknown flag: "+tc.short+"\n", "the message of "+tc.short)
			if got.stdout != "" {
				t.Errorf("%s wrote on stdout:\n%s", tc.short, got.stdout)
			}
		})
	}
}

func TestARetiredShortFormIsAnErrorInTheJsonEnvelopeToo(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "ls", "-s", "Done", "--json").assertCode(t, 2)

	for _, want := range []string{`"code": "unknown_flag"`, `"message": "unknown flag: -s"`} {
		if !strings.Contains(got.stderr, want) {
			t.Errorf("the error envelope lacks %s:\n%s", want, got.stderr)
		}
	}
}

func TestTheFourGlobalShortFormsStillWork(t *testing.T) {
	m := exportBoard(t)

	m.run(t, "-C", m.dir, "ls", "--ids").assertCode(t, 0)
	m.run(t, "ls", "-q").assertCode(t, 0)
	m.run(t, "-V").assertCode(t, 0)
	m.run(t, "ls", "-h").assertCode(t, 0)
}

func TestATitleGivenTwiceIsAnErrorAndNotAnOverwrite(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "new", "Fix the parser", "--title", "feature").assertCode(t, 2)

	assertEqual(t, got.stderr,
		"error: --title and the title argument cannot be used together\n", "the message of the collision")
	if got.stdout != "" {
		t.Errorf("a task was created:\n%s", got.stdout)
	}
	// Even when both texts are the same: what is refused is the two entries,
	// not the disagreement between them.
	m.run(t, "new", "Fix the parser", "--title", "Fix the parser").assertCode(t, 2)
}

func TestATitleGivenOnlyWithTheFlagCreatesTheTask(t *testing.T) {
	m := exportBoard(t)

	id := strings.TrimSpace(m.run(t, "new", "--title", "Only the flag").assertCode(t, 0).stdout)

	title := m.run(t, "get", id, "--json").assertCode(t, 0).stdout
	if !strings.Contains(title, `"title": "Only the flag"`) {
		t.Errorf("the task was not created with the title of the flag:\n%s", title)
	}
}

func TestANewWithNoTitleAtAllStillNamesTheMissingTitle(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "new", "--type", "task").assertCode(t, 2)

	assertEqual(t, got.stderr, "error: title cannot be empty\n", "the message without a title")
}

func TestTheHelpOfNewListsTheTitleFlag(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "new", "--help").assertCode(t, 0)

	if !strings.Contains(got.stdout, "--title <text>") {
		t.Errorf("biso new --help does not list --title:\n%s", got.stdout)
	}
}

// No help text of the program names a short form that does not exist.
func TestNoHelpTextNamesARetiredShortForm(t *testing.T) {
	m := exportBoard(t)

	for _, cmd := range []string{"new", "set", "ls", "get", "export", "start", "finish", "note", "comment", "ask", "answer", "prime"} {
		out := m.run(t, cmd, "--help").assertCode(t, 0).stdout
		for _, retired := range []string{"-a,", "-d,", "-l,", "-o,", "-p,", "-s,", "-t,", "-t/", "-s/", "-p/"} {
			if strings.Contains(out, " "+retired) {
				t.Errorf("biso %s --help still names %q", cmd, retired)
			}
		}
	}
}

// TestTheFlagsOfTheRetiredFilesFieldAreUnknownFlags is the acceptance of
// docs/decisiones/detalles.md#se-retira-modifiedfiles: the four flags of the
// modified files family, and the one `biso finish` used to announce, fail the
// way any flag that does not exist fails, in every command that writes.
func TestTheFlagsOfTheRetiredFilesFieldAreUnknownFlags(t *testing.T) {
	m := exportBoard(t)

	for _, argv := range [][]string{
		{"set", "MYP-2", "--add-files", "a.go"},
		{"set", "MYP-2", "--rm-files", "a.go"},
		{"set", "MYP-2", "--clear-files"},
		{"set", "MYP-2", "--replace-files", "a.go"},
		{"new", "A task", "--add-files", "a.go"},
		{"finish", "MYP-2", "--add-files", "a.go"},
	} {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			flag := argv[len(argv)-1]
			if strings.HasPrefix(flag, "--") == false {
				flag = argv[len(argv)-2]
			}
			got := m.run(t, argv...).assertCode(t, 2)

			assertEqual(t, got.stderr, "error: unknown flag: "+flag+"\n", "the message")
		})
	}
}
