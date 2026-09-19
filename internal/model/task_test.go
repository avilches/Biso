package model

import (
	"testing"
	"time"
)

func TestCriterionKeysAreStableWhenOneInTheMiddleIsRemoved(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	first := task.AddCriterion("The diff ignores CRLF")
	second := task.AddCriterion("There is a test that covers it")
	third := task.AddCriterion("The changelog mentions it")

	if first.Key != 1 || second.Key != 2 || third.Key != 3 {
		t.Fatalf("keys = %d, %d, %d, want 1, 2, 3", first.Key, second.Key, third.Key)
	}

	if !task.RemoveCriterion(2) {
		t.Fatalf("RemoveCriterion(2) said there was no such criterion")
	}

	keys := task.CriterionKeys()
	if len(keys) != 2 || keys[0] != 1 || keys[1] != 3 {
		t.Fatalf("keys after removing #2 = %v, want [1 3]", keys)
	}

	// docs/spec/modelo-de-datos/criterios.md: the counter only grows, so the
	// next criterion is #4 and never reuses the key that was freed.
	fourth := task.AddCriterion("It is documented")
	if fourth.Key != 4 {
		t.Fatalf("the criterion added after removing #2 got key %d, want 4", fourth.Key)
	}
}

func TestAcTotalCountsPresentCriteriaAndNotTheHighestKey(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	task.AddCriterion("one")
	task.AddCriterion("two")
	task.AddCriterion("three")
	task.RemoveCriterion(2)

	// docs/spec/modelo-de-datos/criterios.md: a task with criteria #1 and #3
	// has acTotal 2, not 3.
	if got := task.AcTotal(); got != 2 {
		t.Fatalf("AcTotal() = %d, want 2", got)
	}
	if got := task.AcDone(); got != 0 {
		t.Fatalf("AcDone() = %d, want 0", got)
	}

	task.Criterion(3).Checked = true
	if got := task.AcDone(); got != 1 {
		t.Fatalf("AcDone() after checking #3 = %d, want 1", got)
	}
}

func TestCriteriaKeepTheOrderTheyWereCreatedIn(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	task.AddCriterion("one")
	task.AddCriterion("two")
	task.AddCriterion("three")
	task.RemoveCriterion(1)

	var texts []string
	for _, c := range task.AcceptanceCriteria {
		texts = append(texts, c.Text)
	}
	if len(texts) != 2 || texts[0] != "two" || texts[1] != "three" {
		t.Fatalf("texts = %v, want [two three]", texts)
	}
}

func TestAnUnknownCriterionKeyIsRejectedWithItsOwnError(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	task.AddCriterion("one")
	task.AddCriterion("two")
	task.AddCriterion("three")
	task.RemoveCriterion(2)

	c, err := task.CriterionOrError(7)
	if c != nil {
		t.Fatalf("CriterionOrError(7) returned a criterion: %+v", c)
	}
	if err == nil {
		t.Fatalf("CriterionOrError(7) returned no error")
	}
	// docs/spec/familias-de-flags.md#selectores-de-criterios
	if err.ExitCode != 4 || err.Code != "criterion_not_found" {
		t.Fatalf("error = %d/%s, want 4/criterion_not_found", err.ExitCode, err.Code)
	}
	want := "no acceptance criterion #7 on MYP-11 (keys: 1, 3)"
	if err.Message != want {
		t.Fatalf("message = %q, want %q", err.Message, want)
	}
}

func TestCommentKeysAreStableAndOrderIsCreationOrder(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	at := time.Date(2026, 9, 6, 10, 2, 11, 0, time.UTC)
	first := task.AddComment("@avilches", at, "A user with a Windows clone...")
	second := task.AddComment("@claude", at.Add(time.Hour), "Reproduced it")

	if first.Key != 1 || second.Key != 2 {
		t.Fatalf("keys = %d, %d, want 1, 2", first.Key, second.Key)
	}
	if !task.RemoveComment(1) {
		t.Fatalf("RemoveComment(1) said there was no such comment")
	}
	if got := task.CommentCount(); got != 1 {
		t.Fatalf("CommentCount() = %d, want 1", got)
	}

	// docs/spec/modelo-de-datos/comentarios.md: a comment added later takes the
	// next key of the counter, never the one that was freed, and goes last.
	third := task.AddComment("@sara", at.Add(2*time.Hour), "Fixed")
	if third.Key != 3 {
		t.Fatalf("the comment added after removing #1 got key %d, want 3", third.Key)
	}
	if task.Comments[len(task.Comments)-1].Key != 3 {
		t.Fatalf("the new comment is not the last one: %+v", task.Comments)
	}
}

func TestAnUnknownCommentKeyIsRejectedWithItsOwnError(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	task.AddComment("@avilches", time.Now().UTC(), "one")

	c, err := task.CommentOrError(7)
	if c != nil {
		t.Fatalf("CommentOrError(7) returned a comment: %+v", c)
	}
	if err == nil {
		t.Fatalf("CommentOrError(7) returned no error")
	}
	// docs/spec/familias-de-flags.md#comentarios
	if err.ExitCode != 4 || err.Code != "comment_not_found" {
		t.Fatalf("error = %d/%s, want 4/comment_not_found", err.ExitCode, err.Code)
	}
	want := "no comment #7 on MYP-11 (keys: 1)"
	if err.Message != want {
		t.Fatalf("message = %q, want %q", err.Message, want)
	}
}

func TestListFieldsAreAddressedByNameAndAnUnknownNameIsRejected(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	if err := task.SetListField(FieldLabels, []string{"parser", "crlf"}); err != nil {
		t.Fatalf("SetListField(labels): %v", err)
	}
	got, err := task.ListField(FieldLabels)
	if err != nil {
		t.Fatalf("ListField(labels): %v", err)
	}
	if len(got) != 2 || got[0] != "parser" || got[1] != "crlf" {
		t.Fatalf("labels = %v, want [parser crlf]", got)
	}

	// The list fields are a closed vocabulary: a name that is not one of them
	// is an error, never a silently empty list.
	if _, err := task.ListField("tags"); err == nil {
		t.Fatalf("ListField(\"tags\") returned no error")
	}
	if err := task.SetListField("tags", []string{"x"}); err == nil {
		t.Fatalf("SetListField(\"tags\") returned no error")
	}
}

func TestListFieldsCoversEveryListFieldOfTheModel(t *testing.T) {
	task := &Task{ID: "MYP-11"}
	for _, f := range ListFields() {
		if err := task.SetListField(f, []string{string(f) + "-value"}); err != nil {
			t.Fatalf("SetListField(%q): %v", f, err)
		}
		got, err := task.ListField(f)
		if err != nil {
			t.Fatalf("ListField(%q): %v", f, err)
		}
		if len(got) != 1 || got[0] != string(f)+"-value" {
			t.Fatalf("%s = %v, want one value", f, got)
		}
	}
	if len(ListFields()) != 6 {
		t.Fatalf("ListFields() has %d entries, want the 6 of docs/spec/modelo-de-datos/index.md", len(ListFields()))
	}
}

func TestUnknownExtensionKeyIsRejectedAgainstTheDeclaredOnes(t *testing.T) {
	declared := []string{"trello.card", "github.issue"}

	if err := ValidateExtensionKey("trello.card", declared); err != nil {
		t.Fatalf("a declared key was rejected: %v", err)
	}

	err := ValidateExtensionKey("jira.key", declared)
	if err == nil {
		t.Fatalf("an undeclared key was accepted")
	}
	// docs/spec/modelo-de-datos/campos-externos.md
	if err.ExitCode != 3 || err.Code != "unknown_extension_key" {
		t.Fatalf("error = %d/%s, want 3/unknown_extension_key", err.ExitCode, err.Code)
	}
	if err.Message != `unknown extension key: "jira.key"` {
		t.Fatalf("message = %q", err.Message)
	}
	if err.Field != "ext" || err.Given != "jira.key" {
		t.Fatalf("field/given = %q/%q, want ext/jira.key", err.Field, err.Given)
	}
	if len(err.Valid) != 2 || err.Valid[0] != "trello.card" || err.Valid[1] != "github.issue" {
		t.Fatalf("valid = %v, want the declared keys in their configured order", err.Valid)
	}
}

func TestExtensionKeyAlphabetIsChecked(t *testing.T) {
	// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token
	for _, good := range []string{"trello.card", "github_issue", "a-b", "año.1"} {
		if err := ValidateExtensionKeySyntax(good); err != nil {
			t.Fatalf("ValidateExtensionKeySyntax(%q) = %v, want nil", good, err)
		}
	}
	for _, bad := range []string{"trello=card", "trello card", "trello:card", "trello@card", ""} {
		err := ValidateExtensionKeySyntax(bad)
		if err == nil {
			t.Fatalf("ValidateExtensionKeySyntax(%q) = nil, want an error", bad)
		}
		if err.ExitCode != 2 || err.Code != "malformed_extension_key" {
			t.Fatalf("%q gave %d/%s, want 2/malformed_extension_key", bad, err.ExitCode, err.Code)
		}
		if len(err.Hints) != 1 || err.Hints[0] != "an extension key may contain letters, digits, and - _ ." {
			t.Fatalf("%q gave hints %v", bad, err.Hints)
		}
	}
}

func TestLabelAndAssigneeAlphabetIsChecked(t *testing.T) {
	// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token: the
	// label and assignee alphabet adds @ and : to the one of an extension key.
	for _, good := range []string{"urgent", "team:core", "@sara", "a-b_c.d"} {
		if err := ValidateLabel(good); err != nil {
			t.Fatalf("ValidateLabel(%q) = %v, want nil", good, err)
		}
	}
	err := ValidateLabel("urgent!")
	if err == nil {
		t.Fatalf("ValidateLabel(\"urgent!\") = nil, want an error")
	}
	if err.ExitCode != 2 || err.Code != "malformed_label" {
		t.Fatalf("error = %d/%s, want 2/malformed_label", err.ExitCode, err.Code)
	}
	if err.Message != `malformed label: "urgent!"` {
		t.Fatalf("message = %q", err.Message)
	}
	if len(err.Hints) != 1 || err.Hints[0] != "a label may contain letters, digits, and - _ . : @" {
		t.Fatalf("hints = %v", err.Hints)
	}

	err = ValidateAssignee("sara smith")
	if err == nil || err.Code != "malformed_assignee" {
		t.Fatalf("ValidateAssignee(\"sara smith\") = %v", err)
	}
	if err.Message != `malformed assignee: "sara smith"` {
		t.Fatalf("message = %q", err.Message)
	}
}
