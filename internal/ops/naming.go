package ops

import (
	"fmt"
	"strings"

	"biso/internal/match"
	"biso/internal/model"
)

// This file holds the two things a board's name derives, both of them once
// and at creation: the slug that names its folder
// (docs/spec/resolucion-del-tablero.md#cómo-se-deriva-el-nombre-de-la-carpeta)
// and the prefix its task identifiers carry
// (docs/spec/modelo-de-datos/identificadores.md#identificador-de-tarea).
//
// They start from the same project_name and fail for different reasons, so
// both are checked: "2026" gives a valid slug and no prefix at all, and
// "///" gives neither.

// Slug is the folder name's first half: the name folded and stripped of its
// diacritics, with every run of characters that are neither an ASCII letter
// nor an ASCII digit collapsed into one hyphen, and the hyphens at either
// end trimmed. So "Kex" gives "kex", "Mi Proyecto" gives "mi-proyecto" and
// "Peña 2026" gives "pena-2026".
//
// No character of project_name, a path separator included, can break the
// folder name, because any run of them collapses into the same hyphen.
func Slug(name string) string {
	var b strings.Builder
	pending := false
	for _, r := range match.Simplify(name) {
		if isASCIILetter(r) || (r >= '0' && r <= '9') {
			if pending && b.Len() > 0 {
				b.WriteByte('-')
			}
			pending = false
			b.WriteRune(r)
			continue
		}
		pending = true
	}
	return b.String()
}

// FolderName is what `biso init` calls the folder it creates inside a root:
// the slug and the id, which is what makes a listing of the boards root
// readable. Nothing ever resolves by it.
func FolderName(name, id string) string { return Slug(name) + "-" + id }

// DerivePrefix is task_prefix when no --prefix says otherwise: the name
// folded and stripped of its diacritics, with everything that is not an
// ASCII letter dropped, uppercased. "mi-proyecto-2" gives "MIPROYECTO" and
// "Café" gives "CAFE".
//
// A name that leaves no letter at all has no prefix to derive, and biso
// invents none: it is the invalid_prefix error, which asks for an explicit
// --prefix.
func DerivePrefix(name string) (string, *model.Error) {
	var b strings.Builder
	for _, r := range match.Simplify(name) {
		if isASCIILetter(r) {
			b.WriteRune(r - 'a' + 'A')
		}
	}
	if b.Len() == 0 {
		return "", &model.Error{
			ExitCode: 2,
			Code:     "invalid_prefix",
			Message: fmt.Sprintf(
				"the board name %q leaves no letter to derive a task prefix from", name),
			Field: "prefix",
			Given: name,
			Hints: []string{"give one with --prefix, letters only"},
		}
	}
	return b.String(), nil
}

// ValidatePrefix is the rule of the --prefix flag: letters only.
func ValidatePrefix(prefix string) *model.Error {
	if prefix == "" {
		return invalidPrefix(prefix)
	}
	for _, r := range prefix {
		if !isASCIILetter(r) && !isASCIIUpper(r) {
			return invalidPrefix(prefix)
		}
	}
	return nil
}

func invalidPrefix(given string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "invalid_prefix",
		Message:  fmt.Sprintf("a task prefix is letters only: %q", given),
		Field:    "prefix",
		Given:    given,
	}
}

// ValidateSlug rejects a board name that leaves nothing to name its folder
// with. The slug is a datum of the board like any other, so a value that is
// not valid is never accepted (docs/spec/cmd/config.md).
func ValidateSlug(name string) *model.Error {
	if Slug(name) != "" {
		return nil
	}
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message: fmt.Sprintf(
			"the board name %q leaves no letter or digit to name its folder with", name),
		Field: "project_name",
		Given: name,
	}
}

func isASCIILetter(r rune) bool { return r >= 'a' && r <= 'z' }
func isASCIIUpper(r rune) bool  { return r >= 'A' && r <= 'Z' }
