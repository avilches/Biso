package ops

import (
	"fmt"
	"sort"
	"strings"

	"biso/internal/match"
	"biso/internal/model"
)

// This file is everything about a scoped label that needs more than the
// parsing rule of internal/model: the comparisons that fold the key, the
// `labels` list of the configuration read as what it restricts, and the two
// refusals and the two warnings of
// docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito.
//
// It lives in one file because the same four questions are asked from five
// places that would otherwise each answer them a little differently: `biso
// set` and `biso new` through the write engine, each line of a batch, the
// filters of `biso ls`, `biso config set labels` and `biso doctor`.

// foldKey is how two keys are compared, everywhere: at writing, at removing
// and at querying (docs/spec/valores-de-entrada.md#la-clave-se-compara-plegada-y-el-separador-no-cuenta-al-comparar-valores).
// Folding and not lowercasing, for the reason match.FoldCase gives.
func foldKey(key string) string { return match.FoldCase(key) }

// sameKey answers whether two labels belong to the same key. A plain label
// belongs to none, so it is never the same key as anything.
func sameKey(a, b model.Label) bool {
	return a.Scoped() && b.Scoped() && foldKey(a.Key) == foldKey(b.Key)
}

// labelMatches is the membership test of a reading filter over one stored
// label (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito).
//
// Three shapes of `wanted` answer here, and the difference between them is
// the whole feature. A plain label compares folded, exactly as before. The
// key form `milestone:` matches any value of that key with either
// separator. And a whole scoped label matches by key and value and ignores
// the separator, which is why `--label milestone:m1` finds a task labelled
// `milestone::m1`.
func labelMatches(stored, wanted string) bool {
	w := model.SplitLabel(wanted)
	if !w.Scoped() {
		return match.FoldCase(stored) == match.FoldCase(wanted)
	}
	s := model.SplitLabel(stored)
	if !sameKey(s, w) {
		return false
	}
	if w.IsKey() {
		return true
	}
	return match.FoldCase(s.Value) == match.FoldCase(w.Value)
}

// labelRemoves answers whether a stored label is the one a --rm-labels value
// names. It is the exact comparison every other list field of
// docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma removes by,
// with the one difference a scoped label has: its separator does not count
// and its key is compared folded, so `--rm-labels milestone:m1` takes out a
// stored `milestone::m1`.
func labelRemoves(stored, wanted string) bool {
	w := model.SplitLabel(wanted)
	if !w.Scoped() {
		return stored == wanted
	}
	s := model.SplitLabel(stored)
	return sameKey(s, w) && s.Value == w.Value
}

// labelKeys answers the keys of a set of labels, each one in the spelling it
// first appeared in and each one once, folded keys compared. Nothing is
// sorted here: the caller decides, because the set of the board is sorted
// and the keys of one task are read in the order the task carries them.
func labelKeys(labels []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, raw := range labels {
		l := model.SplitLabel(raw)
		if !l.Scoped() {
			continue
		}
		folded := foldKey(l.Key)
		if seen[folded] {
			continue
		}
		seen[folded] = true
		out = append(out, l.Key)
	}
	return out
}

// labelsOfKey answers the labels of one list that belong to a key, in the
// order the list holds them, which is the order every message that names
// more than one of them prints (the lists of a task are never sorted).
func labelsOfKey(labels []string, key string) []string {
	folded := foldKey(key)
	var out []string
	for _, raw := range labels {
		l := model.SplitLabel(raw)
		if l.Scoped() && foldKey(l.Key) == folded {
			out = append(out, raw)
		}
	}
	return out
}

// labelRule is what the `labels` list of the configuration says about one
// key (docs/spec/cmd/config.md#la-lista-labels).
type labelRule struct {
	// key is the spelling the list used, which is what a message quotes.
	key string
	// separator is the one the key is declared with, and the only one a
	// task may write it with.
	separator string
	// open says the key was declared as `key::` with no values, so it
	// admits any value written with that separator.
	open bool
	// values are the whole labels the key declares, in the order of the
	// list, empty on an open key.
	values []string
}

// labelRules is that list read as what it restricts: one entry per key it
// names, and nothing at all for the plain labels it offers, which restrict
// no one (docs/spec/cmd/config.md#la-lista-labels).
type labelRules struct {
	byKey map[string]*labelRule
}

// readLabelRules builds the rules and answers the two ways the list can
// contradict itself, which is exit code 3 and the `code` of any other
// configuration value the board does not accept.
//
// It is not where a malformed entry is caught: that is the form of the
// token, exit code 2, and it is asked before this, by the layer that writes
// the key.
func readLabelRules(entries []string) (*labelRules, *model.Error) {
	rules := &labelRules{byKey: map[string]*labelRule{}}
	for _, entry := range entries {
		l, err := model.ParseLabelKeyOrLabel(entry)
		if err != nil {
			return nil, err
		}
		if !l.Scoped() {
			// A plain label offers itself and restricts nothing.
			continue
		}
		folded := foldKey(l.Key)
		rule, known := rules.byKey[folded]
		if !known {
			rules.byKey[folded] = &labelRule{
				key: l.Key, separator: l.Separator, open: l.IsKey(),
				values: valuesOf(l),
			}
			continue
		}
		if rule.separator != l.Separator {
			return nil, &model.Error{
				ExitCode: 3,
				Code:     "bad_config_value",
				Message: fmt.Sprintf("labels: the key %q is declared with both : and ::",
					rule.key),
				Hints: []string{
					"a key takes one separator, : for several values or :: for at most one",
				},
				Field: "labels",
				Given: entry,
			}
		}
		if rule.open != l.IsKey() {
			return nil, &model.Error{
				ExitCode: 3,
				Code:     "bad_config_value",
				Message: fmt.Sprintf(
					"labels: the key %q is declared both as an open key and with exact values",
					rule.key),
				Hints: []string{fmt.Sprintf(
					"declare %s%s on its own, or its values, but not both",
					rule.key, rule.separator)},
				Field: "labels",
				Given: entry,
			}
		}
		rule.values = append(rule.values, valuesOf(l)...)
	}
	return rules, nil
}

func valuesOf(l model.Label) []string {
	if l.IsKey() {
		return nil
	}
	return []string{l.Raw}
}

// rule answers what the list says about the key of a label, or nil when the
// list names no key of that label: an unnamed key is free, and so is every
// plain label.
func (rules *labelRules) rule(l model.Label) *labelRule {
	if rules == nil || !l.Scoped() {
		return nil
	}
	return rules.byKey[foldKey(l.Key)]
}

// allows answers why the list would not accept one label on a task, and nil
// when it would. It is the one predicate behind the three places that ask
// it: a label being written, `biso config set labels` judging a label
// already stored, and `biso doctor` reporting one
// (docs/spec/cmd/config.md#la-lista-labels).
//
// The value is compared folded, the same way a label is compared wherever it
// is read: a board that declares `size::m` and is written `size::M` has been
// written one of its own values in another spelling, not a value it does not
// know.
func (rules *labelRules) allows(raw string) *labelRule {
	l := model.SplitLabel(raw)
	rule := rules.rule(l)
	if rule == nil {
		return nil
	}
	if rule.open {
		if rule.separator == l.Separator {
			return nil
		}
		return rule
	}
	for _, declared := range rule.values {
		d := model.SplitLabel(declared)
		if d.Separator == l.Separator && match.FoldCase(d.Value) == match.FoldCase(l.Value) {
			return nil
		}
	}
	return rule
}

// refuse is the error the rule above deserves when it refuses: the open key
// answers about the separator and has no set of values to offer, and the key
// with exact values answers with them (docs/spec/cmd/config.md#la-lista-labels).
func (rule *labelRule) refuse(raw string) *model.Error {
	if rule.open {
		return &model.Error{
			ExitCode: 3,
			Code:     "wrong_label_separator",
			Message: fmt.Sprintf("wrong separator for the label key %q: %q",
				rule.key, raw),
			Hints: []string{fmt.Sprintf("this board declares %s%s, %s",
				rule.key, rule.separator, rule.cardinality())},
			Field: "labels",
			Given: raw,
		}
	}
	return &model.Error{
		ExitCode: 3,
		Code:     "unknown_label_value",
		Message:  fmt.Sprintf("unknown label value: %q", raw),
		// The second display line of the message, aligned under it with the
		// seven spaces of "error: ", the way
		// docs/spec/cmd/config.md#la-lista-labels prints it. The list it
		// names travels in `valid` too, which is what a caller reading JSON
		// branches on.
		Detail: []string{fmt.Sprintf("       valid labels for the key %q on this board: %s",
			rule.key, strings.Join(rule.values, ", "))},
		Valid: append([]string(nil), rule.values...),
		Field: "labels",
		Given: raw,
	}
}

// cardinality is how the hint of a wrong separator says what the declared
// one promises.
func (rule *labelRule) cardinality() string {
	if rule.separator == model.LabelSeparatorExclusive {
		return "at most one value per task"
	}
	return "several values per task"
}

// declares answers the whole labels the key admits, which is what the
// finding of `biso doctor` quotes.
func (rule *labelRule) declares() string { return strings.Join(rule.values, ", ") }

// whyNot is that same refusal in the half sentence the report of
// `biso config set labels` puts after each offending label
// (docs/spec/cmd/config.md#la-lista-labels).
func (rule *labelRule) whyNot() string {
	if rule.open {
		return fmt.Sprintf("the new list declares the key %q with %s",
			rule.key, rule.separator)
	}
	return fmt.Sprintf("the key %q allows %s", rule.key, rule.declares())
}

// mixedSeparators answers the first two values of one key written with the
// two separators, in the order they were written, and nil when there is no
// such pair (docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito).
func mixedSeparators(values []string) []model.Label {
	first := map[string]model.Label{}
	for _, raw := range values {
		l := model.SplitLabel(raw)
		if !l.Scoped() {
			continue
		}
		folded := foldKey(l.Key)
		seen, known := first[folded]
		if !known {
			first[folded] = l
			continue
		}
		if seen.Separator != l.Separator {
			return []model.Label{seen, l}
		}
	}
	return nil
}

// mixedSeparatorsError is the refusal that pair deserves. It names the two
// values in the order they were written and blames neither, because neither
// is more at fault than the other, which is why it carries `field` and no
// `given` (docs/spec/contrato-json.md#los-errores-en-json).
func mixedSeparatorsError(message string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "mixed_label_separators",
		Message:  message,
		Hints:    []string{"a key takes either several values with :, or at most one with ::"},
		Field:    "labels",
		NoGiven:  true,
	}
}

// exclusiveConflict is the error of writing a key with `:` on a task that
// keeps a value of it written with `::`: the stored state says the key takes
// one value, and `k:v` asks for a key that takes several
// (docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito).
func exclusiveConflict(taskID, present, written string) *model.Error {
	l := model.SplitLabel(present)
	return &model.Error{
		ExitCode: 6,
		Code:     "exclusive_label_conflict",
		Message: fmt.Sprintf("%s already has %q, and :: allows at most one value of the key %q",
			taskID, present, l.Key),
		Hints: []string{fmt.Sprintf("drop it first, as in --rm-labels %s --add-labels %s",
			present, written)},
	}
}

// exclusiveLabelReplaced is the warning of
// docs/spec/salida-y-terminal.md#notas-y-avisos for the labels a `::` write
// took off the task, named in the order the task held them.
func exclusiveLabelReplaced(flag, value, taskID string, replaced []string) Warning {
	return Warning{
		Code: "exclusive_label_replaced",
		Message: fmt.Sprintf("--%s: %q replaced %s on %s",
			flag, value, strings.Join(replaced, ", "), taskID),
		Fields: map[string]any{
			"flag": "--" + flag, "value": value,
			"replaced": append([]string(nil), replaced...), "task": taskID,
		},
	}
}

// exclusiveLabelLastWins is the other one: two or more values of the same
// `::` key in one call, where the last one written is the intention and the
// warning says which one stayed. It counts the appearances past two exactly
// as `duplicate_flag_value` does.
func exclusiveLabelLastWins(flag, key, kept string, times int) Warning {
	written := "twice"
	if times != 2 {
		written = fmt.Sprintf("%d times", times)
	}
	return Warning{
		Code: "exclusive_label_last_wins",
		Message: fmt.Sprintf("--%s: key %q given %s with ::, kept %q",
			flag, key, written, kept),
		Fields: map[string]any{"flag": "--" + flag, "key": key, "kept": kept},
	}
}

// exclusiveKeyViolations answers the keys of one list of labels that carry
// more than one value while at least one of them is written `::`, each with
// the labels that break it. It is the shape `biso doctor` reports and the
// shape a line of a batch is refused by, so the two cannot disagree about
// what breaking the exclusivity is.
//
// The keys come out in the order the list first names them, which is the
// order of the task.
func exclusiveKeyViolations(labels []string) [][]string {
	var out [][]string
	for _, key := range labelKeys(labels) {
		of := labelsOfKey(labels, key)
		if len(of) < 2 {
			continue
		}
		exclusive := false
		for _, raw := range of {
			if model.SplitLabel(raw).Exclusive() {
				exclusive = true
			}
		}
		if exclusive {
			out = append(out, of)
		}
	}
	return out
}

// sortedCopy answers the values in the one fixed order every set of this
// program is put in before it is printed or suggested from, without
// disturbing the list it was given.
func sortedCopy(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
