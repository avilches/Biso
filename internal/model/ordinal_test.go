package model

import (
	"math/rand"
	"strings"
	"testing"
)

// TestTheCasesOfTheSpecification is the table of
// docs/spec/modelo-de-datos/orden-manual.md#el-algoritmo-del-punto-medio,
// character for character. It is the one test of this file that pins the
// exact key the algorithm picks; the ones below pin the properties the page
// promises, which is what a caller actually depends on.
func TestTheCasesOfTheSpecification(t *testing.T) {
	for _, c := range []struct {
		prev, next, want string
		exercises        string
	}{
		{"", "", "i", "the board has no key at all yet"},
		{"", "i", "9", "placing before every other"},
		{"i", "", "r", "placing after every other"},
		{"i", "r", "m", "the ordinary midpoint"},
		{"a", "c", "b", "one symbol fits between the first two"},
		{"a", "b", "ai", "consecutive symbols: the key grows by one"},
		{"9", "a", "9i", "the jump from the last digit to the first letter"},
		{"mm", "mn", "mmi", "a common prefix, and the case above behind it"},
		{"i", "j", "ii", "next is a single symbol"},
		{"zz", "", "zzi", "the top end, over the highest key of its length"},
		{"", "01", "00i", "the bottom end, over a key that begins with zero"},
	} {
		got := KeyBetween(c.prev, c.next)
		if got != c.want {
			t.Errorf("KeyBetween(%q, %q) = %q, want %q (%s)",
				c.prev, c.next, got, c.want, c.exercises)
		}
	}
}

// TestTheWorstCaseGrowsBySymbolEveryFiveInsertions is the paragraph of that
// same section: inserting always in front of everything walks the keys down
// one symbol at a time and only then lengthens them.
func TestTheWorstCaseGrowsBySymbolEveryFiveInsertions(t *testing.T) {
	want := []string{"i", "9", "4", "2", "1", "0i", "09", "04", "02", "01", "00i"}
	key := ""
	for i, expected := range want {
		key = KeyBetween("", key)
		if key != expected {
			t.Fatalf("insertion %d gave %q, want %q", i+1, key, expected)
		}
	}
}

// TestValidOrdinal is the definition of a key: not empty, the alphabet
// 0-9a-z, and never a 0 at the end.
func TestValidOrdinal(t *testing.T) {
	for _, c := range []struct {
		key  string
		want bool
	}{
		{"i", true},
		{"00i", true},
		{"m8", true},
		{"zzzz1", true},
		{"", false},
		{"m0", false},
		{"0", false},
		{"M", false},
		{"m-8", false},
		{"m 8", false},
		{"ñ", false},
		{"3000", false},
		{"30", false},
	} {
		if got := ValidOrdinal(c.key); got != c.want {
			t.Errorf("ValidOrdinal(%q) = %v, want %v", c.key, got, c.want)
		}
	}
}

// TestValidateOrdinalIsTheMessageOfTheBatch pins the literal of
// docs/spec/cmd/new.md#el-modo-lote, which the batch prints with its line
// number in front.
func TestValidateOrdinalIsTheMessageOfTheBatch(t *testing.T) {
	err := ValidateOrdinal("3000")
	if err == nil {
		t.Fatal("a key of 3000 was accepted")
	}
	want := `malformed ordinal: "3000" (an ordinal key is made of 0-9 and a-z, and never ends in 0)`
	if err.Message != want {
		t.Errorf("message is %q, want %q", err.Message, want)
	}
	if err.ExitCode != 2 || err.Code != "malformed_ordinal" {
		t.Errorf("ValidateOrdinal answered %d/%s, want 2/malformed_ordinal", err.ExitCode, err.Code)
	}
	if err.Field != "ordinal" || err.Given != "3000" {
		t.Errorf("field/given are %q/%q, want ordinal/3000", err.Field, err.Given)
	}
	if ValidateOrdinal("m8") != nil {
		t.Error("a well formed key was refused")
	}
}

// TestBetweenTwoKeysThereIsAlwaysAnother is the first promise of the page,
// asked over every pair of the table above and over the pairs a gap of one
// symbol produces, which are the ones where an implementation that only
// averaged would have nowhere to go.
func TestBetweenTwoKeysThereIsAlwaysAnother(t *testing.T) {
	keys := []string{"1", "9", "a", "ai", "b", "i", "ii", "j", "m", "mm", "mn", "r", "z", "zz", "00i", "01"}
	for _, prev := range keys {
		for _, next := range keys {
			if prev >= next {
				continue
			}
			assertInside(t, prev, next, KeyBetween(prev, next))
		}
	}
}

// TestThereIsAlwaysOneBelowTheLowestAndOneAboveTheHighest is the other half
// of that promise, which is what makes --ordinal first, --ordinal last,
// --above over the lowest and --below over the highest never fail.
func TestThereIsAlwaysOneBelowTheLowestAndOneAboveTheHighest(t *testing.T) {
	lowest, highest := "i", "i"
	for i := 0; i < 200; i++ {
		below := KeyBetween("", lowest)
		assertKey(t, below)
		if below >= lowest {
			t.Fatalf("round %d: %q is not below %q", i, below, lowest)
		}
		lowest = below

		above := KeyBetween(highest, "")
		assertKey(t, above)
		if above <= highest {
			t.Fatalf("round %d: %q is not above %q", i, above, highest)
		}
		highest = above
	}
}

// TestAKeyNeverEndsInZero asks the one rule that has no second chance: two
// keys that differ only by a trailing zero would leave a place in the order
// where nothing can be inserted.
func TestAKeyNeverEndsInZero(t *testing.T) {
	prev := ""
	for i := 0; i < 500; i++ {
		key := KeyBetween(prev, "")
		assertKey(t, key)
		prev = key
	}
	next := "zzzz"
	for i := 0; i < 500; i++ {
		key := KeyBetween("", next)
		assertKey(t, key)
		next = key
	}
}

// TestABlockOfTasksKeepsTheOrderItWasWrittenIn is the repeated midpoint of
// docs/spec/modelo-de-datos/orden-manual.md#varias-tareas-en-la-misma-llamada,
// including the example that page works out: with m above and t below,
// three tasks take p, r and s.
func TestABlockOfTasksKeepsTheOrderItWasWrittenIn(t *testing.T) {
	got := blockOf(t, "m", "t", 3)
	want := []string{"p", "r", "s"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("the block gave %v, want %v", got, want)
		}
	}

	// And the same rule over a gap that has no room to spare: the keys stay
	// ascending and stay inside the gap however many tasks the call moves.
	for _, gap := range [][2]string{{"a", "b"}, {"", "01"}, {"zz", ""}, {"", ""}, {"i", "j"}} {
		keys := blockOf(t, gap[0], gap[1], 12)
		for i, key := range keys {
			assertKey(t, key)
			if gap[0] != "" && key <= gap[0] {
				t.Errorf("key %d of the block over (%q, %q) is %q, not above the gap",
					i, gap[0], gap[1], key)
			}
			if gap[1] != "" && key >= gap[1] {
				t.Errorf("key %d of the block over (%q, %q) is %q, not below the gap",
					i, gap[0], gap[1], key)
			}
			if i > 0 && key <= keys[i-1] {
				t.Errorf("the block over (%q, %q) is not ascending: %v", gap[0], gap[1], keys)
			}
		}
	}
}

// blockOf is the rule of a call that moves several tasks: the gap is
// computed once and each task takes the midpoint of what is left.
func blockOf(t *testing.T, prev, next string, count int) []string {
	t.Helper()
	keys := make([]string, 0, count)
	for i := 0; i < count; i++ {
		key := KeyBetween(prev, next)
		keys = append(keys, key)
		prev = key
	}
	return keys
}

// TestRandomInsertionsKeepTheOrder is the worst case of the page done for
// real: many insertions at positions the seed picks, plus a long run that
// always lands in the same gap, checking after every one of them that the
// list of keys is still sorted exactly as the list of tasks is.
//
// The seed is fixed so that a failure can be reproduced. What it checks is
// not any particular key but the invariant: every key is well formed, and
// reading the keys in the order the tasks are in gives an ascending list.
func TestRandomInsertionsKeepTheOrder(t *testing.T) {
	rng := rand.New(rand.NewSource(20260921))

	// The list of keys, in the order the tasks are in. Inserting at
	// position i means the new task ends up between i-1 and i.
	order := []string{}
	insert := func(at int) {
		prev, next := "", ""
		if at > 0 {
			prev = order[at-1]
		}
		if at < len(order) {
			next = order[at]
		}
		key := KeyBetween(prev, next)
		assertKey(t, key)
		assertInside(t, prev, next, key)
		order = append(order, "")
		copy(order[at+1:], order[at:])
		order[at] = key
		assertAscending(t, order)
	}

	for i := 0; i < 400; i++ {
		insert(rng.Intn(len(order) + 1))
	}
	// The worst case of the page: always the same spot, which is the one
	// that makes the keys grow. Somewhere in the middle, so that both
	// neighbours exist and neither end can absorb it.
	for i := 0; i < 400; i++ {
		insert(len(order) / 2)
	}
	if len(order) != 800 {
		t.Fatalf("the list holds %d keys, want 800", len(order))
	}
	// The growth is the one the page describes and not the one a successor
	// would give: 800 insertions, most of them in the same gap, stay far
	// from one symbol per insertion.
	longest := 0
	for _, key := range order {
		if len(key) > longest {
			longest = len(key)
		}
	}
	if longest > 100 {
		t.Errorf("the longest key measures %d symbols after 800 insertions", longest)
	}
}

func assertKey(t *testing.T, key string) {
	t.Helper()
	if !ValidOrdinal(key) {
		t.Fatalf("%q is not a key", key)
	}
}

func assertInside(t *testing.T, prev, next, key string) {
	t.Helper()
	assertKey(t, key)
	if prev != "" && key <= prev {
		t.Errorf("KeyBetween(%q, %q) = %q, which is not above %q", prev, next, key, prev)
	}
	if next != "" && key >= next {
		t.Errorf("KeyBetween(%q, %q) = %q, which is not below %q", prev, next, key, next)
	}
}

func assertAscending(t *testing.T, keys []string) {
	t.Helper()
	for i := 1; i < len(keys); i++ {
		if keys[i-1] >= keys[i] {
			t.Fatalf("the keys are out of order at %d: %s", i,
				strings.Join(keys[max(0, i-3):min(len(keys), i+3)], ", "))
		}
	}
}
