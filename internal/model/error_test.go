package model

import (
	"errors"
	"testing"
)

func TestErrorSatisfiesErrorInterface(t *testing.T) {
	e := &Error{
		ExitCode: 8,
		Code:     "busy",
		Message:  "the board is busy, another process is writing to it",
	}

	var target error = e
	if target.Error() != "the board is busy, another process is writing to it" {
		t.Fatalf("Error() = %q, want the message", target.Error())
	}

	var asErr *Error
	if !errors.As(target, &asErr) {
		t.Fatalf("errors.As did not recover the concrete *Error")
	}
	if asErr.ExitCode != 8 || asErr.Code != "busy" {
		t.Fatalf("the recovered error lost its fields: %+v", asErr)
	}
}
