package cli

import (
	"fmt"
	"strings"
	"testing"
)

// This file holds the second property of section 7.1 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md: a
// table of representative invocations per command is parsed, an argv is built
// back from the result, and parsing that argv has to give the same thing
// again. render lives only here, nothing in production uses it, and that is
// the point: it checks that the parser is deterministic without needing a
// second implementation to compare against, and it catches the class of bug
// where two equivalent invocations resolve to different values.

// render rebuilds a canonical argv from what the parser understood, using the
// same specification table the parser walked. It writes every value attached
// with "=", which is the form that always works
// (docs/spec/valores-de-entrada.md#valores-que-empiezan-por-guion), and it
// escapes a value that begins with "@" so that the second parse reads it as
// the text it is and not as the name of a file.
func render(p *Parsed) []string {
	argv := []string{}
	if p.Command != "" {
		argv = append(argv, p.Command)
	}
	argv = append(argv, p.Positionals...)
	for _, o := range p.occs {
		if o.dropped {
			continue
		}
		if o.flag.Value == NoValue {
			argv = append(argv, o.flag.long())
			continue
		}
		value := o.value
		if o.key != "" {
			value = o.key + "=" + value
		}
		if o.flag.Value == TextValue && strings.HasPrefix(value, "@") {
			value = "@" + value
		}
		argv = append(argv, o.flag.long()+"="+value)
	}
	for name := range p.emptied {
		argv = append(argv, "--"+name+"=")
	}
	return argv
}

// fingerprint is everything two parses have to agree on: the command, the
// positional arguments, and every change with its step, its key and its value,
// in the order the caller is going to apply them.
func fingerprint(p *Parsed) string {
	var b strings.Builder
	fmt.Fprintf(&b, "command=%s\n", p.Command)
	fmt.Fprintf(&b, "positionals=%s\n", strings.Join(p.Positionals, "|"))
	for _, c := range p.Changes() {
		fmt.Fprintf(&b, "change %d %s %q %q\n", c.Category, c.Flag.Name, c.Key, c.Value)
	}
	for _, o := range p.occs {
		if !o.dropped && o.flag.Category == NotAChange {
			fmt.Fprintf(&b, "flag %s %q\n", o.flag.Name, o.value)
		}
	}
	return b.String()
}

func TestParseRenderRoundTrip(t *testing.T) {
	invocations := [][]string{
		{"set", "MYP-1"},
		{"set", "MYP-1", "MYP-2", "--add-labels", "a,b", "--add-labels", "c"},
		{"set", "MYP-1", "-l", "urgent", "--status", "Done", "--clear-desc"},
		{"set", "MYP-1", "--append-note", "-x", "--append-desc", "several\nlines"},
		{"set", "MYP-1", "--ext", "trello=card/9", "--rm-ext", "old"},
		{"set", "MYP-1", "--set-comment-date", "a=b=2026-08-14T10:22:00Z"},
		{"set", "MYP-1", "--comment", "said something", "--comment-author", "@sara"},
		{"set", "MYP-1", "--replace-labels", ""},
		{"set", "MYP-1", "--append-plan", "@@literal at sign"},
		{"ls", "--json"},
		{"ls", "--limit", "10", "--mine"},
		{"init", "--at", "../boards/mine", "--dry-run"},
	}
	for _, argv := range invocations {
		t.Run(strings.Join(argv, " "), func(t *testing.T) {
			first, err := Parse(argv, testCommands(), Env{})
			if err != nil {
				t.Fatalf("parse(%q): %v", argv, err)
			}
			again := render(first)
			second, err := Parse(again, testCommands(), Env{})
			if err != nil {
				t.Fatalf("parse(render(...)) of %q: %v", again, err)
			}
			if got, want := fingerprint(second), fingerprint(first); got != want {
				t.Errorf("round trip through %q changed the invocation:\ngot\n%s\nwant\n%s", again, got, want)
			}
		})
	}
}
