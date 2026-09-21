package model

import "testing"

// These are the tables of
// docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito, one case per row
// plus the ones the prose adds. The rule is frozen by
// docs/spec/estabilidad.md, so a change that makes any of them fail is a
// change to the contract and not to an implementation detail.

func TestSplitLabelCutsTheKeyTheSeparatorAndTheValue(t *testing.T) {
	for _, c := range []struct {
		raw, key, separator, value string
	}{
		{"urgent", "", "", ""},
		{"milestone:m1", "milestone", ":", "m1"},
		{"milestone::m1", "milestone", "::", "m1"},
		// The separator is taken as long as it can be up to two, so what
		// is left of a third colon belongs to the value.
		{"trello:card:42", "trello", ":", "card:42"},
		{"size:::m", "size", "::", ":m"},
	} {
		got := SplitLabel(c.raw)
		if got.Key != c.key || got.Separator != c.separator || got.Value != c.value {
			t.Errorf("SplitLabel(%q) = key %q, separator %q, value %q; want %q, %q, %q",
				c.raw, got.Key, got.Separator, got.Value, c.key, c.separator, c.value)
		}
	}
}

func TestAPlainLabelIsNotScoped(t *testing.T) {
	l := SplitLabel("urgent")
	if l.Scoped() || l.Exclusive() || l.IsKey() {
		t.Errorf("the plain label %q answered scoped %v, exclusive %v, key %v",
			l.Raw, l.Scoped(), l.Exclusive(), l.IsKey())
	}
}

func TestTheSeparatorSaysHowManyValuesTheKeyTakes(t *testing.T) {
	if SplitLabel("size:m").Exclusive() {
		t.Error("size:m says the key takes at most one value, and : says it takes several")
	}
	if !SplitLabel("size::m").Exclusive() {
		t.Error("size::m says the key takes several values, and :: says at most one")
	}
}

// TestTheMalformedFormsAreRefusedWithTheirHint is the table of
// docs/spec/valores-de-entrada.md#las-formas-mal-formadas, with the hint of
// each half of the rule.
func TestTheMalformedFormsAreRefusedWithTheirHint(t *testing.T) {
	for _, c := range []struct {
		raw, hint string
	}{
		{"size:", LabelHalvesHint},
		{"size::", LabelHalvesHint},
		{":m", LabelHalvesHint},
		{"::", LabelHalvesHint},
		{"size:::m", LabelColonHint},
		{"size::m:", LabelColonHint},
	} {
		_, err := ParseLabel(c.raw)
		if err == nil {
			t.Errorf("ParseLabel(%q) accepted it, and the rule refuses it", c.raw)
			continue
		}
		if err.ExitCode != 2 || err.Code != "malformed_label" {
			t.Errorf("ParseLabel(%q) = %d/%s, want 2/malformed_label",
				c.raw, err.ExitCode, err.Code)
		}
		if want := "malformed label: \"" + c.raw + "\""; err.Message != want {
			t.Errorf("ParseLabel(%q) said %q, want %q", c.raw, err.Message, want)
		}
		if len(err.Hints) != 1 || err.Hints[0] != c.hint {
			t.Errorf("ParseLabel(%q) hinted %v, want %q", c.raw, err.Hints, c.hint)
		}
	}
}

func TestTheWellFormedFormsAreAccepted(t *testing.T) {
	for _, raw := range []string{"urgent", "milestone:m1", "milestone::m1", "trello:card:42"} {
		if _, err := ParseLabel(raw); err != nil {
			t.Errorf("ParseLabel(%q) = %v, want nil", raw, err)
		}
	}
}

// TestTheKeyFormIsALabelNowhereAndAFilterInTwoPlaces is what keeps the two
// readings apart: `size:` is malformed as a label and legal as the value of
// a filter or as an entry of the `labels` list.
func TestTheKeyFormIsALabelNowhereAndAFilterInTwoPlaces(t *testing.T) {
	for _, raw := range []string{"size:", "milestone::"} {
		if _, err := ParseLabel(raw); err == nil {
			t.Errorf("ParseLabel(%q) accepted a form no task can carry", raw)
		}
		l, err := ParseLabelKeyOrLabel(raw)
		if err != nil {
			t.Errorf("ParseLabelKeyOrLabel(%q) = %v, want nil", raw, err)
			continue
		}
		if !l.IsKey() {
			t.Errorf("ParseLabelKeyOrLabel(%q) did not read it as a key", raw)
		}
	}
	// And the reading that accepts the key form refuses everything else
	// the rule refuses.
	for _, raw := range []string{":m", "::", "size:::m", "size::m:"} {
		if _, err := ParseLabelKeyOrLabel(raw); err == nil {
			t.Errorf("ParseLabelKeyOrLabel(%q) accepted a malformed label", raw)
		}
	}
}

// TestAStoredLabelGoesThroughTheRuleToo is what makes the rule hold wherever
// a task is written and not only where a flag is typed.
func TestAStoredLabelGoesThroughTheRuleToo(t *testing.T) {
	task := &Task{Title: "A task", Labels: []string{"size:"}}
	err := task.Validate()
	if err == nil {
		t.Fatal("a task carrying a malformed scoped label validated")
	}
	var e *Error
	if !asError(err, &e) || e.Code != "malformed_label" {
		t.Errorf("validating it answered %v, want a malformed_label", err)
	}
}

func asError(err error, into **Error) bool {
	e, ok := err.(*Error)
	if ok {
		*into = e
	}
	return ok
}
