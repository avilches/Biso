package cli

import "testing"

// A short form exists only if no other flag of the same command starts with
// its letter (docs/decisiones/comandos-y-flags.md#una-forma-corta-solo-existe-si-nadie-mas-reclama-su-inicial).
// Letters are compared as written, so -C and -c are different letters.
func TestAShortFormIsNeverClaimedByAnotherFlagOfTheSameCommand(t *testing.T) {
	for _, cmd := range Commands() {
		flags := append(GlobalFlags(), cmd.Flags...)
		for _, f := range flags {
			if f.Short == "" {
				continue
			}
			for _, other := range flags {
				if other.Name == f.Name || other.Name[:1] != f.Short {
					continue
				}
				t.Errorf("biso %s: -%s is the short form of --%s, but --%s starts with the same letter",
					cmd.Name, f.Short, f.Name, other.Name)
			}
		}
	}
}

// Only the flags every command shares keep a short form: the four of
// docs/spec/cmd/flags-globales.md. A command's own flag never has one, and
// this test says so in the plainest way, so that adding one is a decision
// somebody has to reverse here.
func TestNoFlagOfACommandHasAShortForm(t *testing.T) {
	for _, cmd := range Commands() {
		for _, f := range cmd.Flags {
			if f.Short != "" {
				t.Errorf("biso %s: --%s has the short form -%s", cmd.Name, f.Name, f.Short)
			}
		}
	}
}
