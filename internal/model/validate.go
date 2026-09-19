package model

import (
	"fmt"
	"strings"
	"unicode"
)

// The two token alphabets of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token. Both
// take Unicode letters and digits; they differ only in which symbols they
// add. An extension key leaves out @ and :, which have no documented use
// there, and =, which --ext <key>=<value> already spends on separating the
// two halves.
const (
	tokenSymbols        = "-_.:@"
	extensionKeySymbols = "-_."
)

// ValidateLabel answers the error of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token when
// a label steps outside its alphabet, and nil when it does not. It is a
// problem of form, not of the board not recognizing the value, so the exit
// code is 2 and not the 3 of an unknown vocabulary value: neither labels
// nor assignees have a closed vocabulary when written.
func ValidateLabel(label string) *Error {
	return validateToken(label, tokenSymbols, "label", "malformed_label", "labels",
		"a label may contain letters, digits, and - _ . : @")
}

// ValidateAssignee is ValidateLabel for an assignee, with the same
// alphabet and the same exit code.
func ValidateAssignee(assignee string) *Error {
	return validateToken(assignee, tokenSymbols, "assignee", "malformed_assignee", "assignees",
		"an assignee may contain letters, digits, and - _ . : @")
}

// ValidateExtensionKeySyntax answers the same kind of error for the key of
// an extension field, whose alphabet is the narrower one
// (docs/spec/modelo-de-datos/campos-externos.md). It says nothing about
// whether the board declares that key: that is ValidateExtensionKey, and
// the two are separate because they are different failures with different
// exit codes.
func ValidateExtensionKeySyntax(key string) *Error {
	return validateToken(key, extensionKeySymbols, "extension key", "malformed_extension_key", "ext",
		"an extension key may contain letters, digits, and - _ .")
}

// validateToken is the shared body of the three above.
func validateToken(value, symbols, noun, code, field, hint string) *Error {
	if value != "" && allowedToken(value, symbols) {
		return nil
	}
	return &Error{
		ExitCode: 2,
		Code:     code,
		Message:  fmt.Sprintf("malformed %s: %q", noun, value),
		Hints:    []string{hint},
		Field:    field,
		Given:    value,
	}
}

// allowedToken answers whether every rune of value is a Unicode letter, a
// Unicode digit, or one of the symbols the alphabet adds.
func allowedToken(value, symbols string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if strings.ContainsRune(symbols, r) {
			continue
		}
		return false
	}
	return true
}

// ValidateExtensionKey answers the error of
// docs/spec/modelo-de-datos/campos-externos.md when a key is not one of
// the ones the board declares in its `extensions` list, and nil when it
// is. Writing an undeclared key is exit code 3.
//
// The second line the specification prints under the message,
//
//	error: unknown extension key: "jira.key"
//	       declared keys on this board: trello.card, github.issue
//
// is the rendering of Valid, the same shape as the unknown-status error of
// docs/spec/vocabularios.md, so it is not a hint and does not live here:
// this builds the error, and the layer that writes to stderr writes both
// lines from it.
func ValidateExtensionKey(key string, declared []string) *Error {
	for _, d := range declared {
		if d == key {
			return nil
		}
	}
	valid := make([]string, len(declared))
	copy(valid, declared)
	return &Error{
		ExitCode: 3,
		Code:     "unknown_extension_key",
		Message:  fmt.Sprintf("unknown extension key: %q", key),
		Field:    "ext",
		Given:    key,
		Valid:    valid,
	}
}
