package ops

import (
	"fmt"
	"strconv"
	"strings"

	"biso/internal/model"
)

// This file is docs/spec/familias-de-flags.md#selectores-de-criterios: the
// value that --check-ac, --uncheck-ac, --rm-ac, --rm-comment and
// --set-comment-date take, and how it chooses elements of a list.
//
// One implementation serves the criteria and the comments, because the
// specification says the two take the same selector with the key of a
// comment in place of the key of a criterion and the body of a comment in
// place of the text of a criterion. Writing it twice is exactly the drift
// that page exists to prevent.

// listKind is which of the two lists a selector is being resolved against.
// It carries the four words a message of this family needs, so that the same
// code prints "no acceptance criterion #7 on MYP-11" and "no comment #7 on
// MYP-11" without a branch at every sentence.
type listKind struct {
	singular  string // "acceptance criterion"
	plural    string // "acceptance criteria"
	notFound  string // the error code of a key that is not there
	ambiguous string // the error code of a text that matches more than one
	emptyWarn string // the warning code of `all` over an empty list
}

var (
	criteriaList = listKind{
		singular:  "acceptance criterion",
		plural:    "acceptance criteria",
		notFound:  "criterion_not_found",
		ambiguous: "criterion_ambiguous",
		emptyWarn: "no_acceptance_criteria",
	}
	commentsList = listKind{
		singular:  "comment",
		plural:    "comments",
		notFound:  "comment_not_found",
		ambiguous: "comment_ambiguous",
		emptyWarn: "no_comments",
	}
)

// element is one addressable element of either list: its stable key and the
// text a text selector compares against.
type element struct {
	Key  int
	Text string
}

func criteriaElements(t *model.Task) []element {
	out := make([]element, 0, len(t.AcceptanceCriteria))
	for _, c := range t.AcceptanceCriteria {
		out = append(out, element{Key: c.Key, Text: c.Text})
	}
	return out
}

func commentElements(t *model.Task) []element {
	out := make([]element, 0, len(t.Comments))
	for _, c := range t.Comments {
		out = append(out, element{Key: c.Key, Text: c.Body})
	}
	return out
}

// selector is one value of one of those five flags, already told apart into
// the two things it can be.
type selector struct {
	raw  string
	all  bool
	keys []keyPart
	text string
}

// keyPart is one comma-separated piece of a key selector: one key, or a
// range. The two are kept apart because a key that does not exist is an
// error and a range with gaps is not
// (docs/spec/familias-de-flags.md#selectores-de-criterios).
type keyPart struct {
	from, to int
	isRange  bool
}

// isKeySelector is the disambiguation rule, to be implemented exactly as it
// is written: the value is a list of keys only if the value as a whole
// matches ^(all|\d+(-\d+)?)(,\d+(-\d+)?)*$. Anything else is literal text,
// commas included, so --check-ac "1, 2 and the last one" is a text search
// that finds nothing rather than something half done.
func isKeySelector(raw string) bool {
	if raw == "" {
		return false
	}
	for i, part := range strings.Split(raw, ",") {
		if i == 0 && part == "all" {
			continue
		}
		if !isKeyOrRange(part) {
			return false
		}
	}
	return true
}

func isKeyOrRange(part string) bool {
	from, to, isRange := strings.Cut(part, "-")
	if !allDigits(from) {
		return false
	}
	return !isRange || allDigits(to)
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parseSelector reads one value. The only failure it can answer is the
// inverted range, which is a malformed command line and therefore exit
// code 2.
func parseSelector(flag, raw string) (selector, *model.Error) {
	if !isKeySelector(raw) {
		return selector{raw: raw, text: raw}, nil
	}
	s := selector{raw: raw}
	for _, part := range strings.Split(raw, ",") {
		if part == "all" {
			s.all = true
			continue
		}
		from, to, isRange := strings.Cut(part, "-")
		first, _ := strconv.Atoi(from)
		if !isRange {
			s.keys = append(s.keys, keyPart{from: first, to: first})
			continue
		}
		last, _ := strconv.Atoi(to)
		if last < first {
			return selector{}, &model.Error{
				ExitCode: 2,
				Code:     "inverted_range",
				Message:  fmt.Sprintf("--%s: inverted range: %q", flag, part),
				Hints:    []string{"a range goes from the lower key to the higher one, as in 1-4"},
				Field:    flag,
				Given:    raw,
			}
		}
		s.keys = append(s.keys, keyPart{from: first, to: last, isRange: true})
	}
	return s, nil
}

// isAll says whether this selector chooses the whole list, which is the one
// shape that is valid over several tasks at once.
func (s selector) isAll() bool { return s.all }

// resolve answers the keys this selector chooses out of the elements a task
// had when the call started, and the warning of `all` over an empty list.
func (s selector) resolve(flag string, taskID string, kind listKind, elements []element) ([]int, *Warning, *model.Error) {
	if s.all {
		if len(elements) == 0 {
			return nil, &Warning{
				Code:    kind.emptyWarn,
				Message: fmt.Sprintf("%s has no %s", taskID, kind.plural),
				Fields:  map[string]any{"task": taskID},
			}, nil
		}
		keys := make([]int, 0, len(elements))
		for _, e := range elements {
			keys = append(keys, e.Key)
		}
		return keys, nil, nil
	}
	if len(s.keys) > 0 {
		return s.resolveKeys(taskID, kind, elements)
	}
	return s.resolveText(flag, taskID, kind, elements)
}

func (s selector) resolveKeys(taskID string, kind listKind, elements []element) ([]int, *Warning, *model.Error) {
	present := make(map[int]bool, len(elements))
	for _, e := range elements {
		present[e.Key] = true
	}
	var keys []int
	seen := map[int]bool{}
	for _, part := range s.keys {
		for key := part.from; key <= part.to; key++ {
			if !present[key] {
				if part.isRange {
					// A range applies the keys that are there, with no
					// warning: asking for 1-4 on a task whose criteria
					// are #1 and #3 is not asking for a #2.
					continue
				}
				return nil, nil, missingKeyError(taskID, kind, key, elements)
			}
			if !seen[key] {
				seen[key] = true
				keys = append(keys, key)
			}
		}
	}
	return keys, nil, nil
}

func missingKeyError(taskID string, kind listKind, key int, elements []element) *model.Error {
	return &model.Error{
		ExitCode: 4,
		Code:     kind.notFound,
		Message: fmt.Sprintf("no %s #%d on %s (keys: %s)",
			kind.singular, key, taskID, joinKeys(elements)),
		Field: "selector",
		Given: strconv.Itoa(key),
	}
}

// resolveText applies the fragment of text, with the case folded and the
// diacritics dropped on both sides, and ends in the two errors of the table:
// nothing found is exit code 4 and more than one is exit code 5, and both
// list the elements so that the next call does not have to ask what is there.
func (s selector) resolveText(flag, taskID string, kind listKind, elements []element) ([]int, *Warning, *model.Error) {
	var keys []int
	for _, e := range elements {
		if TextMatches(e.Text, s.text) {
			keys = append(keys, e.Key)
		}
	}
	switch len(keys) {
	case 1:
		return keys, nil, nil
	case 0:
		return nil, nil, &model.Error{
			ExitCode: 4,
			Code:     kind.notFound,
			Message: fmt.Sprintf("no %s of %s matches %q",
				kind.singular, taskID, s.text),
			Detail: listing(kind, elements),
			Field:  flag,
			Given:  s.raw,
		}
	}
	return nil, nil, &model.Error{
		ExitCode: 5,
		Code:     kind.ambiguous,
		Message: fmt.Sprintf("%q matches %d %s of %s",
			s.text, len(keys), kind.plural, taskID),
		Detail: listing(kind, matching(elements, keys)),
		Field:  flag,
		Given:  s.raw,
	}
}

// listing writes the elements the two text errors print under the message,
// one per line and indented, with the key each one answers to.
func listing(kind listKind, elements []element) []string {
	if len(elements) == 0 {
		return []string{fmt.Sprintf("  it has no %s", kind.plural)}
	}
	out := make([]string, 0, len(elements))
	for _, e := range elements {
		out = append(out, fmt.Sprintf("  #%d %s", e.Key, e.Text))
	}
	return out
}

func matching(elements []element, keys []int) []element {
	chosen := map[int]bool{}
	for _, key := range keys {
		chosen[key] = true
	}
	var out []element
	for _, e := range elements {
		if chosen[e.Key] {
			out = append(out, e)
		}
	}
	return out
}

func joinKeys(elements []element) string {
	parts := make([]string, 0, len(elements))
	for _, e := range elements {
		parts = append(parts, strconv.Itoa(e.Key))
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, ", ")
}
