package cli

import (
	"fmt"

	"biso/internal/model"
)

// This file holds every error the analysis of the command line can produce.
// They are born here already complete, with the exit code of
// docs/spec/codigos-de-salida.md, the code of
// docs/spec/contrato-json.md#los-identificadores-de-error and the detail keys
// that code is documented to carry, and they travel up unchanged: no layer in
// between reinterprets or wraps them.

// usage builds an exit code 2 error, the code of a command line that is
// malformed.
func usage(code, message string) *model.Error {
	return &model.Error{ExitCode: 2, Code: code, Message: message}
}

// flagUsage is usage for an error that names a flag, which per the table of
// docs/spec/contrato-json.md#los-errores-en-json carries field and given.
func flagUsage(code string, f *FlagSpec, given, message string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     code,
		Message:  message,
		Field:    f.field(),
		Given:    given,
	}
}

func errUnknownCommand(name string, known []string) *model.Error {
	e := usage("unknown_command", fmt.Sprintf("unknown command: %q", name))
	e.Valid = known
	e.Given = name
	return e
}

func errUnknownFlag(spelling string) *model.Error {
	e := usage("unknown_flag", "unknown flag: "+spelling)
	e.Given = spelling
	return e
}

func errMissingValue(f *FlagSpec) *model.Error {
	return flagUsage("missing_value", f, "", f.long()+" requires a value")
}

func errTakesNoValue(f *FlagSpec, given string) *model.Error {
	return flagUsage("unexpected_argument", f, given, f.long()+" takes no value")
}

func errUnexpectedArgument(arg string) *model.Error {
	e := usage("unexpected_argument", "unexpected argument: "+arg)
	e.Given = arg
	return e
}

func errDuplicateScalar(f *FlagSpec, first, second string) *model.Error {
	return flagUsage("duplicate_scalar_flag", f, second, fmt.Sprintf(
		"%s given twice with different values: %q and %q", f.long(), first, second))
}

func errMalformedPair(f *FlagSpec, given string) *model.Error {
	return flagUsage("unexpected_argument", f, given, fmt.Sprintf(
		"%s: expected %s, got %q", f.long(), f.PairSyntax, given))
}

// errMalformedToken is the error of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token: a
// character outside the alphabet a field closes is a matter of form, not of a
// vocabulary the board does not recognize, so it is exit code 2 and not 3.
func errMalformedToken(f *FlagSpec, given string) *model.Error {
	code := "malformed_" + underscored(f.Noun)
	e := flagUsage(code, f, given, fmt.Sprintf("malformed %s: %q", f.Noun, given))
	e.Hints = []string{fmt.Sprintf(
		"%s %s may contain letters, digits, and %s",
		article(f.Noun), f.Noun, f.Alphabet.symbols())}
	return e
}

// errMalformedString is the error of
// docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string, for a
// one-line field that arrives carrying a newline or a carriage return.
func errMalformedString(f *FlagSpec, given string) *model.Error {
	e := flagUsage("malformed_string_value", f, given,
		fmt.Sprintf("malformed %s: %q", f.field(), given))
	e.Hints = []string{"a string field cannot contain a newline or a carriage return"}
	return e
}

// errEmptyScalar is the third row of
// docs/spec/valores-de-entrada.md#el-valor-vacío: the empty string is never
// the way to clear a scalar, and the hint names the flag that is.
func errEmptyScalar(f *FlagSpec) *model.Error {
	e := &model.Error{
		ExitCode: 3,
		Code:     "empty_scalar_value",
		Message:  f.long() + " cannot be empty",
		Field:    f.field(),
		Given:    "",
	}
	if f.ClearFlag != "" {
		e.Hints = []string{"to clear it, use --" + f.ClearFlag}
	}
	return e
}

func errTwoStdin(first, second *FlagSpec) *model.Error {
	return flagUsage("two_stdin", second, "-", fmt.Sprintf(
		"- can be given only once per invocation; %s and %s both read stdin",
		first.long(), second.long()))
}

func errFileNotFound(f *FlagSpec, path string) *model.Error {
	return &model.Error{
		ExitCode: 4,
		Code:     "file_not_found",
		Message:  fmt.Sprintf("%s: file not found: %s", f.long(), path),
		Field:    f.field(),
		Given:    path,
	}
}

func errFileUnreadable(f *FlagSpec, path string) *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "file_unreadable",
		Message:  fmt.Sprintf("%s: file cannot be read: %s", f.long(), path),
		Field:    f.field(),
		Given:    path,
	}
}

// errStdinUnreadable is the input stream itself failing, which is the
// environment failing and not the request: exit code 8 of
// docs/spec/codigos-de-salida.md.
func errStdinUnreadable(f *FlagSpec, err error) *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "io_error",
		Message:  fmt.Sprintf("%s: stdin cannot be read: %v", f.long(), err),
		Field:    f.field(),
		Given:    "-",
	}
}

// errInvalidEncoding is the rule of
// docs/spec/salida-y-terminal.md#codificación-y-texto: an invalid byte
// sequence in an argument or in an input file is exit code 3, and the message
// points at the byte.
func errInvalidEncoding(f *FlagSpec, offset int) *model.Error {
	prefix := ""
	if f != nil {
		prefix = f.long() + ": "
	}
	e := &model.Error{
		ExitCode: 3,
		Code:     "invalid_encoding",
		Message:  fmt.Sprintf("%sinvalid UTF-8 at byte %d", prefix, offset),
	}
	if f != nil {
		e.Field = f.field()
	}
	return e
}

func errOutsideDomain(f *FlagSpec, given string) *model.Error {
	e := flagUsage(f.DomainCode, f, given, fmt.Sprintf(
		"%s: unknown value: %q", f.long(), given))
	e.Valid = f.Domain
	return e
}

func errIncompatible(first, second *FlagSpec) *model.Error {
	return usage("incompatible_flags", fmt.Sprintf(
		"%s and %s cannot be used together", first.long(), second.long()))
}

func errRequires(f *FlagSpec, required string) *model.Error {
	return usage("incompatible_flags", fmt.Sprintf(
		"%s requires --%s", f.long(), required))
}

// errReadOnlyFlag carries the two literal messages of
// docs/spec/cmd/flags-globales.md for a global flag written where it has
// nothing to do.
func errReadOnlyFlag(message string) *model.Error {
	return usage("read_only_flag", message)
}
