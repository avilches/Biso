package main

import (
	"strings"
	"testing"
)

// This file is the rule of
// docs/spec/codigos-de-salida.md#el-código-7-garantiza-que-no-se-ha-escrito-nada-y-el-código-específico-siempre-gana-sobre-él
// seen from outside the program: a preview answers the code of the failure
// it found, so the exit code table of a command that writes tasks and takes
// no batch has no row for the 7 at all, and no call can reach one.

func TestADryRunThatDoesNotPassKeepsTheCodeOfItsFailure(t *testing.T) {
	for _, c := range []struct {
		name string
		argv []string
		code int
	}{
		{"a reference that does not exist",
			[]string{"set", "MYP-999", "--priority", "high", "--dry-run"}, 4},
		{"a value outside a closed vocabulary",
			[]string{"set", "MYP-1", "--priority", "urgent", "--dry-run"}, 3},
		{"nothing to change",
			[]string{"set", "MYP-1", "--dry-run"}, 2},
		{"a task that cannot be started twice",
			[]string{"finish", "MYP-999", "--dry-run"}, 4},
		{"a verb that needs a text it was not given",
			[]string{"note", "MYP-1", "--dry-run"}, 2},
		{"an answer with no question to answer",
			[]string{"answer", "MYP-1", "Anything", "--dry-run"}, 6},
		{"archiving something that is not there",
			[]string{"archive", "MYP-999", "--dry-run"}, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := smallBoard(t)

			got := m.run(t, c.argv...).assertCode(t, c.code)

			if strings.Contains(got.stderr, "dry_run_failed") {
				t.Errorf("the failure named an identifier that does not exist:\n%s", got.stderr)
			}
		})
	}
}

// TestNoHelpOfAWritingCommandPromisesASeven is the other half: the eight
// exit code tables of those commands, as `biso <cmd> --help` prints them,
// no longer list a 7 nobody can reach.
func TestNoHelpOfAWritingCommandPromisesASeven(t *testing.T) {
	m := smallBoard(t)
	for _, command := range []string{
		"set", "archive", "start", "note", "comment", "finish", "ask", "answer",
	} {
		got := m.run(t, command, "--help").assertCode(t, 0)

		codes, found := cutExitCodes(got.stdout)
		if !found {
			t.Fatalf("biso %s --help prints no exit codes", command)
		}
		if strings.Contains(codes, "7 ") {
			t.Errorf("biso %s --help still promises a 7:\n%s", command, codes)
		}
	}
}

// cutExitCodes is the block of a help text that lists the exit codes.
func cutExitCodes(help string) (string, bool) {
	_, rest, found := strings.Cut(help, "Exit codes:\n")
	if !found {
		return "", false
	}
	block, _, _ := strings.Cut(rest, "\n\n")
	return block, true
}
