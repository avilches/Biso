package cli

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file checks the rule of docs/spec/cmd/get.md#salida for a list in the
// card: each value is written with the escape the input uses, read backwards.
// A comma inside a value comes out as `\,`, a backslash as `\\`, nothing else
// is touched, and the separator between values is a bare comma and a space.

func TestTheCardEscapesTheCommaAndTheBackslashOfEveryValueOfAList(t *testing.T) {
	for _, c := range []struct {
		name   string
		values []string
		want   string
	}{
		{"an empty list is a dash", nil, "-"},
		{"a single plain value", []string{"x"}, "x"},
		{"two plain values", []string{"a", "b"}, "a, b"},
		{"a comma in the middle", []string{"a,b", "c.md"}, `a\,b, c.md`},
		{"a comma at the start", []string{",a"}, `\,a`},
		{"a comma at the end", []string{"a,"}, `a\,`},
		{"several commas in one value", []string{"a,b,c"}, `a\,b\,c`},
		{"a comma alone", []string{","}, `\,`},
		{"a backslash alone", []string{`\`}, `\\`},
		{"a backslash followed by a comma", []string{`\,`}, `\\\,`},
		{"a backslash before any other character", []string{`p\q`}, `p\\q`},
		{"a path of Windows", []string{`C:\dir\notes.md`}, `C:\\dir\\notes.md`},
		{"a trailing backslash before the separator", []string{`a\`, "b"}, `a\\, b`},
		{"a comma and a space inside one value", []string{"a, b"}, `a\, b`},
		{"the spaces of a value are kept", []string{"a b", "  c d  "}, "a b,   c d  "},
		{"the order is the one of the list", []string{"z,y", "b", "a,x"}, `z\,y, b, a\,x`},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := joined(c.values); got != c.want {
				t.Errorf("joined(%q) = %q, want %q", c.values, got, c.want)
			}
		})
	}
}

// TestTwoDifferentListsNeverShareALine is the collision that is the reason to
// escape the backslash: without it, the references `a\` and `b` and the
// single reference `a, b` would both print as `a\, b`.
func TestTwoDifferentListsNeverShareALine(t *testing.T) {
	two := joined([]string{`a\`, "b"})
	one := joined([]string{"a, b"})
	if two == one {
		t.Errorf("the lists [a\\ b] and [a, b] print the same line: %q", two)
	}
}

// cardValues is a battery of values, the conflicting ones first, that has to
// survive being written in the card and read back.
var cardValues = []string{
	"a", "a,b", ",", ",,", ",a", "a,", "a,b,c", `\`, `\\`, `\,`, `,\`, `\\,`, `a\`, `\a`, `p\q`,
	`C:\dir\notes.md`, "a, b", " a", "a ", "  c d  ", "docs/bugs/BUG-02.md", "https://x.y/z?a=1,2",
	"é,ñ\\", "x\\\\\\,", "tab\there",
}

// TestWhatTheCardPrintsReadsBackAsTheOriginalValues is the property of the
// whole path: for any list of values, splitting what the card prints on its
// bare commas, dropping the space after each one and undoing the escape gives
// exactly the values it started from. The reader of the card does what the
// specification says, and splitList is the code of the input that undoes the
// escape.
func TestWhatTheCardPrintsReadsBackAsTheOriginalValues(t *testing.T) {
	check := func(values []string) {
		t.Helper()
		line := joined(values)
		if len(values) == 0 {
			if line != "-" {
				t.Errorf("an empty list prints %q", line)
			}
			return
		}
		got := readTheCardLine(line)
		if !reflect.DeepEqual(got, values) {
			t.Errorf("the list %q printed as %q reads back as %q", values, line, got)
		}
	}

	for _, v := range cardValues {
		check([]string{v})
		check([]string{v, v})
		for _, w := range cardValues {
			check([]string{v, w})
		}
	}
	check(cardValues)

	// A random battery over the alphabet that hurts: the two special
	// characters, a space, and a letter.
	rng := rand.New(rand.NewSource(92))
	alphabet := []rune{',', '\\', ' ', 'a'}
	for i := 0; i < 3000; i++ {
		values := make([]string, 1+rng.Intn(4))
		for j := range values {
			n := rng.Intn(7)
			rs := make([]rune, n)
			for k := range rs {
				rs[k] = alphabet[rng.Intn(len(alphabet))]
			}
			values[j] = string(rs)
		}
		// A value that is empty or only spaces cannot be stored (the
		// input discards it), so it is not a value of a list.
		ok := true
		for _, v := range values {
			if isEmpty(v) {
				ok = false
			}
		}
		if ok {
			check(values)
		}
	}
}

// readTheCardLine is the reading the specification gives for a line of a
// list: split on every bare comma, drop the space that follows it, undo the
// escape. It is written here, next to the property, and not in the
// production code, because nothing in the program reads the card.
func readTheCardLine(line string) []string {
	// Cut on the commas that are not escaped, walking the line the way the
	// escape reads it: a backslash takes the next character with it.
	var parts []string
	var cur strings.Builder
	rs := []rune(line)
	for i := 0; i < len(rs); i++ {
		switch {
		case rs[i] == '\\' && i+1 < len(rs):
			cur.WriteRune(rs[i])
			cur.WriteRune(rs[i+1])
			i++
		case rs[i] == ',':
			parts = append(parts, cur.String())
			cur.Reset()
			// The space after the separator is not part of the next value.
			if i+1 < len(rs) && rs[i+1] == ' ' {
				i++
			}
		default:
			cur.WriteRune(rs[i])
		}
	}
	parts = append(parts, cur.String())

	out := make([]string, len(parts))
	for i, p := range parts {
		// The escape of a single value is undone by the input's own rule:
		// a value without a bare comma is one element of splitList.
		got := splitList(p)
		if len(got) != 1 {
			return []string{"<<" + p + " did not read as one value>>"}
		}
		out[i] = got[0]
	}
	return out
}

// TestTheEscapeOfOneValueIsTheInverseOfTheInputRule pins the function itself:
// for every value, reading the escaped text with the input rule gives one
// element, that value.
func TestTheEscapeOfOneValueIsTheInverseOfTheInputRule(t *testing.T) {
	for _, v := range cardValues {
		got := splitList(escapeListValue(v))
		if !reflect.DeepEqual(got, []string{v}) {
			t.Errorf("splitList(escapeListValue(%q)) = %q", v, got)
		}
	}
}

// TestTheListsOfTheCardThatCannotCarryACommaAreTheSameAsBefore checks the
// four other lists of the meta block: their values have no comma and no
// backslash, so the line is what it always was.
func TestTheListsOfTheCardThatCannotCarryACommaAreTheSameAsBefore(t *testing.T) {
	v := ops.TaskView{
		Task: &model.Task{
			ID:           "MYP-1",
			Status:       "todo",
			Assignees:    []string{"@claude", "@avilches"},
			Labels:       []string{"parser", "ui-2"},
			Dependencies: []string{"MYP-2", "MYP-3"},
			References:   []string{"a,b", "c.md"},
		},
		Blocks: []string{"MYP-9", "MYP-10"},
	}
	block := metaBlock(v)
	for _, want := range []string{
		"@claude, @avilches",
		"parser, ui-2",
		"MYP-2, MYP-3",
		"MYP-9, MYP-10",
		"refs       " + `a\,b, c.md` + "\n",
	} {
		if !strings.Contains(block, want) {
			t.Errorf("the meta block lacks %q:\n%s", want, block)
		}
	}
}
