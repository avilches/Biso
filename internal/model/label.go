package model

import (
	"fmt"
	"strings"
)

// This file is docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito: the
// one rule that splits a label into a key and a value, and the forms of it
// that no board ever stores.
//
// It is pure text and knows nothing about a board, which is what lets the
// same rule run over the value of a flag, over a line of a batch, over an
// entry of the configuration and over a label that is already stored. The
// rule is frozen by docs/spec/estabilidad.md, so the shape of this file is
// part of the contract and not an implementation detail.

// The two separators of a scoped label. The one a key is written with says
// how many values of that key one task admits, and nothing else: ":" admits
// as many as it is given, "::" at most one.
const (
	LabelSeparatorSeveral   = ":"
	LabelSeparatorExclusive = "::"
)

// The two hints of a malformed scoped label, one per half of the rule. They
// are constants because two layers build the same error: the model, over a
// task that is about to be written, and internal/cli, over a value that has
// just been typed.
const (
	LabelHalvesHint = "a scoped label is key:value or key::value, and neither side can be empty"
	LabelColonHint  = "the value of a scoped label cannot start or end with a colon"
)

// Label is one label already split by the rule.
type Label struct {
	// Raw is the label exactly as it was written or stored, which is what
	// every message quotes.
	Raw string
	// Key is the text before the first colon, and is empty on a plain
	// label.
	Key string
	// Separator is ":" or "::", and is empty on a plain label.
	Separator string
	// Value is everything behind the separator, colons included. It is
	// empty only on the key form `key:`, which is a filter and never a
	// label.
	Value string
}

// Scoped answers whether the label carries a key at all. A label with no
// colon in it is a plain label, and none of the rules of this file touches
// it.
func (l Label) Scoped() bool { return l.Separator != "" }

// Exclusive answers whether the key was written with the separator that
// admits at most one value of it per task.
func (l Label) Exclusive() bool { return l.Separator == LabelSeparatorExclusive }

// IsKey answers the form `key:` or `key::`, with nothing behind the
// separator. It is how a filter asks for any value of a key and how the
// configuration declares an open key, and it is never a label a task can
// carry (docs/spec/valores-de-entrada.md#las-formas-mal-formadas).
func (l Label) IsKey() bool { return l.Scoped() && l.Value == "" }

// SplitLabel applies the rule without judging the result: the key is the
// text before the first colon, the separator is the run of colons right
// behind it taken as long as it can be up to two, and the value is
// everything that is left, colons included.
func SplitLabel(raw string) Label {
	at := strings.IndexByte(raw, ':')
	if at < 0 {
		return Label{Raw: raw}
	}
	separator := LabelSeparatorSeveral
	if strings.HasPrefix(raw[at:], LabelSeparatorExclusive) {
		separator = LabelSeparatorExclusive
	}
	return Label{
		Raw:       raw,
		Key:       raw[:at],
		Separator: separator,
		Value:     raw[at+len(separator):],
	}
}

// ParseLabel is SplitLabel plus the refusals of
// docs/spec/valores-de-entrada.md#las-formas-mal-formadas, for a label that
// is about to be stored on a task: neither half can be empty and the value
// cannot start or end in a colon. The key form `key:` is a filter and not a
// label, so it is refused here like any other empty half.
func ParseLabel(raw string) (Label, *Error) {
	l := SplitLabel(raw)
	if !l.Scoped() {
		return l, nil
	}
	if l.Key == "" || l.Value == "" {
		return l, MalformedLabel(raw, LabelHalvesHint)
	}
	return l, colonEnds(l)
}

// ParseLabelKeyOrLabel is ParseLabel for the two places where the key form
// is legal instead of malformed: the value of a reading filter
// (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito)
// and an entry of the `labels` list of the configuration
// (docs/spec/cmd/config.md#la-lista-labels). Everything else about the rule
// is the same, which is the point: the two do not have a grammar of their
// own, they only accept one more form of the same one.
func ParseLabelKeyOrLabel(raw string) (Label, *Error) {
	l := SplitLabel(raw)
	if !l.Scoped() {
		return l, nil
	}
	if l.Key == "" {
		return l, MalformedLabel(raw, LabelHalvesHint)
	}
	if l.Value == "" {
		return l, nil
	}
	return l, colonEnds(l)
}

func colonEnds(l Label) *Error {
	if strings.HasPrefix(l.Value, LabelSeparatorSeveral) ||
		strings.HasSuffix(l.Value, LabelSeparatorSeveral) {
		return MalformedLabel(l.Raw, LabelColonHint)
	}
	return nil
}

// MalformedLabel is the error of a label whose form the rule refuses. It
// carries the same `code` as a label that steps outside the token alphabet,
// because it is the same kind of failure: the shape of the token, and not a
// vocabulary the board does not recognize
// (docs/spec/valores-de-entrada.md#las-formas-mal-formadas).
func MalformedLabel(raw, hint string) *Error {
	return &Error{
		ExitCode: 2,
		Code:     "malformed_label",
		Message:  fmt.Sprintf("malformed label: %q", raw),
		Hints:    []string{hint},
		Field:    "labels",
		Given:    raw,
	}
}
