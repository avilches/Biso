package match

import (
	"fmt"
	"strings"

	"biso/internal/model"
)

// Field is a field whose vocabulary is closed: the board configures its
// values and nothing else is accepted, when writing and when filtering alike.
// docs/spec/vocabularios.md names three of them, and its value is the word
// that appears in the error message and, prefixed with "unknown_", in the
// error code of docs/spec/contrato-json.md#los-identificadores-de-error.
type Field string

const (
	Status   Field = "status"
	Type     Field = "type"
	Priority Field = "priority"
)

// Match resolves value against the values the board has configured for field,
// and returns the configured value it resolves to, following
// docs/spec/vocabularios.md#el-algoritmo-de-coincidencia step by step:
//
//   - an exactly equal configured value wins, even when the board has two
//     values that normalize the same;
//   - otherwise the normalized forms are compared, and exactly one match wins;
//     a configured value repeated letter for letter counts once, because it is
//     the same value listed twice and not two values to disambiguate;
//   - no match is exit code 3 with the code "unknown_<field>";
//   - two or more matches is exit code 3 with the code "ambiguous_vocabulary",
//     because the board itself is what has to be disambiguated.
//
// The result is the configured spelling, never what was typed, so whatever
// stores or filters by it always sees the board's own vocabulary. The error,
// when there is one, is always a *model.Error and already carries the values
// the layer that prints it needs: the caller neither reinterprets nor wraps it.
//
// The value arrives already resolved by the layer that read it, and Match
// does not second-guess it. In particular, a value of nothing but spaces is
// the empty value by docs/spec/valores-de-entrada.md#el-valor-vacío, and it is
// that layer, the one that reads a flag, a file or standard input, that turns
// it into the empty string. Match quotes in its message whatever it was
// given, so a board that ever printed unknown status: "   " would be
// reporting a caller that skipped that conversion, not a rule of its own.
//
// The same call serves writing and filtering. That is not a convention that
// each command follows on its own, it is the only implementation there is,
// which is what makes the table of
// docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos
// true by construction.
func Match(field Field, value string, configured []string) (string, error) {
	for _, c := range configured {
		if c == value {
			return c, nil
		}
	}

	normalized := Normalize(value)
	var matches []string
	for _, c := range configured {
		if Normalize(c) != normalized {
			continue
		}
		if contains(matches, c) {
			continue
		}
		matches = append(matches, c)
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", &model.Error{
			ExitCode: 3,
			Code:     "unknown_" + string(field),
			Message:  fmt.Sprintf("unknown %s: %q", field, value),
			Field:    string(field),
			Given:    value,
			Valid:    copyOf(configured),
		}
	default:
		return "", &model.Error{
			ExitCode: 3,
			Code:     "ambiguous_vocabulary",
			Message: fmt.Sprintf("ambiguous %s: %q matches %d configured values: %s",
				field, value, len(matches), strings.Join(matches, ", ")),
			Hints: []string{"type one of them exactly, or rename one so the two no longer normalize the same"},
			Field: string(field),
			Given: value,
			Valid: matches,
		}
	}
}

// contains says whether values already holds this exact spelling. It is the
// deduplication of step b: a board that lists the same value twice has one
// value, so it can never be the ambiguity of step e.
func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// copyOf returns a copy of the given values, so that an error travelling up
// the stack can never be a window onto the caller's own slice.
func copyOf(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
