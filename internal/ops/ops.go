// Package ops holds the logic of each command: typed parameters in, a typed
// result out, and no terminal I/O of any kind. Nothing here prints text,
// builds JSON or knows what --json is; internal/cli does all of that with
// what these functions answer.
//
// Every function takes the board it works on as its first argument, per
// section 3.5 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md.
// Init and Where are the two exceptions, and they are the reason the type
// exists: one creates the board and the other explains how it was found, so
// neither can be handed one already open. They take an Env instead, which is
// the same environment internal/cli would have used to open it.
package ops

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"biso/internal/board"
	"biso/internal/model"
	"biso/internal/vcs"
)

// Env is everything a command needs from outside itself that is not the
// board: where the call is being made from, what this machine is configured
// with, who is calling, and the two sources a test has to be able to replace
// to stay deterministic.
type Env struct {
	// Dir is the working directory, absolute and clean: the current
	// directory unless -C or BISO_CWD said another one.
	Dir string
	// Machine is the machine configuration, already loaded.
	Machine board.Machine
	// Me is the identity of whoever is calling, from BISO_ME or from the
	// machine's `me` key, and empty when neither is set
	// (docs/spec/invocacion.md#variables-de-entorno).
	Me string
	// Now is the clock, and NewID the source of board identifiers.
	Now   func() time.Time
	NewID func() (string, error)
}

// WithDefaults fills in the two sources with the real ones.
func (e Env) WithDefaults() Env {
	if e.Now == nil {
		e.Now = func() time.Time { return time.Now().UTC() }
	}
	if e.NewID == nil {
		e.NewID = RandomID
	}
	return e
}

// RandomID mints a board identifier from the system's source of random
// numbers: the eight lowercase hexadecimal characters of
// docs/spec/resolucion-del-tablero.md. Whether it already exists on this
// machine is a separate question, which `biso init` asks of the roots.
func RandomID() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", &model.Error{
			ExitCode: 8,
			Code:     "io_error",
			Message:  "a board id could not be generated: " + err.Error(),
		}
	}
	return hex.EncodeToString(b[:]), nil
}

// VCSConfig turns the machine's version control keys into the configuration
// internal/vcs understands. It lives here and not in internal/board because
// the dependency rule keeps that package below internal/vcs.
func VCSConfig(m board.Machine) vcs.Config {
	return vcs.Config{
		Kind: vcs.Kind(m.VCS),
		Custom: vcs.Custom{
			Commit:     m.VCSCustom.Commit,
			Publish:    m.VCSCustom.Publish,
			IgnoreFile: m.VCSCustom.IgnoreFile,
		},
	}
}
