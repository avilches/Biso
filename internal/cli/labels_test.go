package cli

import (
	"strings"
	"testing"

	"biso/internal/model"
)

// These are the two readings of the scoped-label rule as the command line
// applies them: the flags that write a label refuse the key form, and the
// two filters accept it because it is how they ask for any value of a key
// (docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito).
//
// They run over the real tables of commands.go and not over the ones
// parse_test.go builds, because what is being checked is which flags carry
// which reading, which is a fact about that table.

// parseArgv runs the real command tables over an argv.
func parseArgv(t *testing.T, argv ...string) (*Parsed, error) {
	t.Helper()
	return Parse(argv, Commands(), Env{})
}

func TestTheFlagsThatWriteALabelRefuseAMalformedOne(t *testing.T) {
	for _, flag := range []string{"--add-labels", "--rm-labels", "--replace-labels"} {
		for _, value := range []string{"size:", "size::", ":m", "::", "size:::m", "size::m:"} {
			_, err := parseArgv(t, "set", "MYP-11", flag, value)
			e := labelError(t, err)
			if e == nil {
				t.Errorf("%s %s was accepted, and the rule refuses it", flag, value)
				continue
			}
			if e.ExitCode != 2 || e.Code != "malformed_label" {
				t.Errorf("%s %s = %d/%s, want 2/malformed_label",
					flag, value, e.ExitCode, e.Code)
			}
			if e.Field != "labels" || e.Given != value {
				t.Errorf("%s %s carries field %q and given %q", flag, value, e.Field, e.Given)
			}
		}
	}
}

// TestRemovingByKeyIsNotASyntaxOfTheWritingFlags is the row of
// docs/spec/familias-de-flags.md: the form with no value is the syntax of a
// filter, there is no "take out every value of a key", and emptying the
// whole list is --clear-labels.
func TestRemovingByKeyIsNotASyntaxOfTheWritingFlags(t *testing.T) {
	_, err := parseArgv(t, "set", "MYP-11", "--rm-labels", "size:")
	if e := labelError(t, err); e == nil || e.Code != "malformed_label" {
		t.Errorf("--rm-labels size: = %v, want a malformed_label", err)
	}
}

func TestTheFlagsThatWriteALabelAcceptTheWellFormedOnes(t *testing.T) {
	for _, value := range []string{"urgent", "size:m", "size::m", "trello:card:42"} {
		if _, err := parseArgv(t, "set", "MYP-11", "--add-labels", value); err != nil {
			t.Errorf("--add-labels %s = %v, want nil", value, err)
		}
	}
}

func TestTheLabelFiltersAcceptTheKeyFormAndRefuseTheRest(t *testing.T) {
	for _, flag := range []string{"--label", "--label-or"} {
		for _, value := range []string{"size:", "size::", "size:m", "size::m", "urgent"} {
			if _, err := parseArgv(t, "ls", flag, value); err != nil {
				t.Errorf("%s %s = %v, want nil", flag, value, err)
			}
		}
		for _, value := range []string{":m", "::", "size:::m"} {
			_, err := parseArgv(t, "ls", flag, value)
			e := labelError(t, err)
			if e == nil || e.ExitCode != 2 || e.Code != "malformed_label" {
				t.Errorf("%s %s = %v, want 2/malformed_label", flag, value, err)
			}
		}
	}
}

// TestTheFilterOfExportTakesTheSameReading is what one shared table of
// filters is for: `biso export` takes exactly the filters of `biso ls`
// (docs/spec/cmd/export.md), so it cannot read a label differently.
func TestTheFilterOfExportTakesTheSameReading(t *testing.T) {
	if _, err := parseArgv(t, "export", "--label", "milestone:"); err != nil {
		t.Errorf("biso export --label milestone: = %v, want nil", err)
	}
	_, err := parseArgv(t, "export", "--label", ":m")
	if e := labelError(t, err); e == nil || e.Code != "malformed_label" {
		t.Errorf("biso export --label :m = %v, want a malformed_label", err)
	}
}

// TestTheHelpOfTheFourCommandsNamesWhatTheRuleAdded is the other half of the
// fixtures of cmd/biso/testdata: those compare the whole block against the
// specification, and this one says which sentence of it is the one that had
// to change.
func TestTheHelpOfTheFourCommandsNamesWhatTheRuleAdded(t *testing.T) {
	for _, c := range []struct{ name, help, sentence string }{
		{"set", setHelp, "A label with a colon is scoped: key:value allows several values of that key"},
		{"set", setHelp, "6  a scoped label already has its one value"},
		{"ls", lsHelp, "The form key:"},
		{"config", configHelp, "a key::value or key:: entry also restricts what that key"},
		{"doctor", doctorHelp, "stored label the labels list does not allow"},
	} {
		if !strings.Contains(c.help, c.sentence) {
			t.Errorf("the help of biso %s no longer says %q", c.name, c.sentence)
		}
	}
}

func labelError(t *testing.T, err error) *model.Error {
	t.Helper()
	if err == nil {
		return nil
	}
	e, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	return e
}
