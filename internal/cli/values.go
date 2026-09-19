package cli

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"
)

// This file holds the rules of docs/spec/valores-de-entrada.md that are about
// a value and not about the shape of the command line: the three forms of
// giving a long value, the commas that split a list, the alphabet a token
// field closes, and what a value of nothing but spaces means.

// Env is everything the parser reads from outside itself. A test gives it its
// own standard input and its own reader, so that no rule of this package needs
// a real terminal or a real file to be exercised.
type Env struct {
	Stdin    io.Reader
	ReadFile func(name string) ([]byte, error)
}

func (e Env) withDefaults() Env {
	if e.Stdin == nil {
		e.Stdin = os.Stdin
	}
	if e.ReadFile == nil {
		e.ReadFile = os.ReadFile
	}
	return e
}

// readValue turns what was typed behind a flag into the value itself,
// following docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo.
// Only a flag of kind TextValue admits the three forms: a person field takes
// its value exactly as it is, so --comment-author @trello:juan stores that text
// and reads no file.
func (st *parser) readValue(f *FlagSpec, raw string) (string, error) {
	if f.Value != TextValue {
		return raw, nil
	}
	switch {
	case raw == "-":
		return st.readStdin(f)
	case strings.HasPrefix(raw, "@@"):
		// The only escape sequence the program has: the first "@" is
		// dropped and the rest is literal.
		return raw[1:], nil
	case strings.HasPrefix(raw, "@"):
		return st.readFile(f, raw[1:])
	}
	return raw, nil
}

// readStdin reads standard input whole, once. A second flag asking for it is a
// usage error and not a second read: the stream is already exhausted, so it
// would store the empty value without anyone noticing.
func (st *parser) readStdin(f *FlagSpec) (string, error) {
	if st.stdinTakenBy != nil {
		return "", errTwoStdin(st.stdinTakenBy, f)
	}
	st.stdinTakenBy = f
	b, err := io.ReadAll(st.env.Stdin)
	if err != nil {
		return "", errStdinUnreadable(f, err)
	}
	return string(b), nil
}

func (st *parser) readFile(f *FlagSpec, path string) (string, error) {
	b, err := st.env.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", errFileNotFound(f, path)
		}
		return "", errFileUnreadable(f, path)
	}
	return string(b), nil
}

// normalizeNewlines applies the rule of
// docs/spec/salida-y-terminal.md#codificación-y-texto: on reading an input,
// "\r\n" and "\n" are accepted alike and normalized to "\n".
func normalizeNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// invalidUTF8At returns the offset of the first byte that is not valid UTF-8,
// or -1 when the whole string is. The input and the output are UTF-8 always,
// whatever the locale of the system.
func invalidUTF8At(s string) int {
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			return i
		}
		i += size
	}
	return -1
}

// splitList splits a value on its commas, where "\," is a literal comma, per
// docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas. A
// backslash escapes a comma and nothing else, so any other backslash is part
// of the value: a path or a URL keeps whatever it carries.
func splitList(s string) []string {
	var (
		out  []string
		cur  []rune
		prev rune
	)
	for _, r := range s {
		switch {
		case r == ',' && prev == '\\':
			// The backslash was written to escape this comma, so it is
			// not part of the value.
			cur = append(cur[:len(cur)-1], ',')
		case r == ',':
			out = append(out, string(cur))
			cur = nil
		default:
			cur = append(cur, r)
		}
		prev = r
	}
	return append(out, string(cur))
}

// isEmpty is the definition of docs/spec/valores-de-entrada.md#el-valor-vacío:
// a string with no character at all, or with nothing but spaces, wherever it
// came from.
func isEmpty(s string) bool { return strings.TrimSpace(s) == "" }

// allowed answers whether every character of a value is inside the alphabet
// the field closes. Neither alphabet admits a space.
func (a Alphabet) allowed(s string) bool {
	if a == AnyText {
		return true
	}
	symbols := "-_."
	if a == TokenAlphabet {
		symbols = "-_.:@"
	}
	for _, r := range s {
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

// underscored turns the noun of a field into the tail of its error code, so
// that "extension key" gives malformed_extension_key.
func underscored(noun string) string {
	return strings.ReplaceAll(noun, " ", "_")
}
