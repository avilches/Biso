package cli

import (
	"fmt"
	"sort"
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
//
// For the property to be worth anything, render has to write back everything
// the parser reads: it escapes what the parser unescapes, it walks nothing
// unordered, and it rotates through the four spellings of a flag instead of
// always writing the one that is easiest to get right. The one value it cannot
// write back is a text value that is exactly "-", because that is how the
// specification spells standard input and there is no escape for it
// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo), so
// the table below does not contain one.

// spelling is one of the four ways a flag and its value can be written, from
// docs/spec/valores-de-entrada.md#cómo-se-lee-la-línea-de-comandos.
type spelling int

const (
	longAttached spelling = iota
	longSeparate
	shortAttached
	shortSeparate
	spellings
)

func (s spelling) String() string {
	return [...]string{"--flag=value", "--flag value", "-f=value", "-f value"}[s]
}

// write turns one flag and one already escaped value into the argv fragment of
// this spelling, falling back to the long form for a flag that has no short
// name.
func (s spelling) write(f *FlagSpec, value string) []string {
	switch s {
	case longSeparate:
		return []string{f.long(), value}
	case shortAttached:
		if f.Short != "" {
			return []string{"-" + f.Short + "=" + value}
		}
	case shortSeparate:
		if f.Short != "" {
			return []string{"-" + f.Short, value}
		}
		return []string{f.long(), value}
	}
	return []string{f.long() + "=" + value}
}

// escapeList puts back the escapes of
// docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas, in
// the reverse order of the parser: the backslash first, so that the one this
// call adds in front of a comma is not escaped in turn.
func escapeList(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	return strings.ReplaceAll(v, ",", `\,`)
}

// render rebuilds a canonical argv from what the parser understood, using the
// same specification table the parser walked. The positional arguments go in
// front of the flags, which is where
// docs/spec/valores-de-entrada.md#cómo-se-lee-la-línea-de-comandos puts them,
// and the escapes go back in so that the second parse reads the same values:
// the comma and the backslash of a list, and the "@" that would otherwise name
// a file.
func render(p *Parsed) []string {
	argv := []string{}
	if p.Command != "" {
		argv = append(argv, p.Command)
	}
	argv = append(argv, p.Positionals...)
	used := 0
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
		if o.flag.Comma {
			value = escapeList(value)
		}
		if o.flag.Value == TextValue && strings.HasPrefix(value, "@") {
			value = "@" + value
		}
		argv = append(argv, spelling(used%int(spellings)).write(o.flag, value)...)
		used++
	}
	// A map has no order, and an argv that changes between two runs would
	// turn this property into a coin toss.
	emptied := make([]string, 0, len(p.emptied))
	for name := range p.emptied {
		emptied = append(emptied, name)
	}
	sort.Strings(emptied)
	for _, name := range emptied {
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

func roundTripInvocations() [][]string {
	return [][]string{
		{"set", "MYP-1"},
		{"set", "MYP-1", "MYP-2", "--add-labels", "a,b", "--add-labels", "c"},
		{"set", "MYP-1", "-l", "urgent", "--status", "Done", "--clear-desc"},
		{"set", "MYP-1", "--append-note", "-x", "--append-desc", "several\nlines"},
		{"set", "MYP-1", "--ext", "trello=card/9", "--rm-ext", "old"},
		{"set", "MYP-1", "--set-comment-date", "a=b=2026-08-14T10:22:00Z"},
		{"set", "MYP-1", "--comment", "said something", "--comment-author", "@sara"},
		{"set", "MYP-1", "--replace-labels", ""},
		{"set", "MYP-1", "--append-plan", "@@literal at sign"},
		// The example of docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma:
		// one reference with a comma inside, and one with a backslash of its own.
		{"set", "MYP-1", "--add-refs", `notes/a\,b.md`},
		{"set", "MYP-1", "--add-refs", `C:\dir\,notes/b.md`, "--add-refs", `ends-in\\`},
		{"set", "MYP-1", "-d", "-5 degrees", "-t", "A title"},
		// Four flags with a value, so that the fourth one is rendered as
		// "-f value", the last of the four spellings.
		{"set", "MYP-1", "--append-plan", "p", "--append-note", "n",
			"--add-refs", "r", "-l", "x"},
		{"ls", "--json"},
		{"ls", "--limit", "10", "--mine"},
		{"init", "--at", "../boards/mine", "--dry-run"},
	}
}

func TestParseRenderRoundTrip(t *testing.T) {
	for _, argv := range roundTripInvocations() {
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
			// Rendering the same parse twice has to give the same argv,
			// or the property above would be answering a different
			// question on every run.
			if repeat := render(first); strings.Join(repeat, "\x00") != strings.Join(again, "\x00") {
				t.Errorf("render is not deterministic:\n%q\n%q", again, repeat)
			}
		})
	}
}

// TestRoundTripCoversEverySpelling is what keeps the property above from
// checking one single form of writing a flag: the table has to exercise the
// four spellings, or a bug in three of them would never be seen.
func TestRoundTripCoversEverySpelling(t *testing.T) {
	seen := map[spelling]bool{}
	for _, argv := range roundTripInvocations() {
		p, err := Parse(argv, testCommands(), Env{})
		if err != nil {
			t.Fatalf("parse(%q): %v", argv, err)
		}
		used := 0
		for _, o := range p.occs {
			if o.dropped || o.flag.Value == NoValue {
				continue
			}
			s := spelling(used % int(spellings))
			if s == shortAttached || s == shortSeparate {
				if o.flag.Short == "" {
					used++
					continue
				}
			}
			seen[s] = true
			used++
		}
	}
	for s := spelling(0); s < spellings; s++ {
		if !seen[s] {
			t.Errorf("no invocation of the table is rendered as %s", s)
		}
	}
}
