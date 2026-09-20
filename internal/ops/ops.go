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
	// ListLimit is how many rows `biso ls` prints when the call writes no
	// --limit: what BISO_LIMIT said, and otherwise the machine's
	// `default_limit` key, which is already thirty when the file does not
	// set it. It is a pointer so that an Env built by hand, in a test or
	// in a caller that has no machine configuration, still answers the
	// built-in thirty instead of a limit of zero rows
	// (docs/spec/invocacion.md#variables-de-entorno).
	ListLimit *int
	// Now is the clock, and NewID the source of board identifiers.
	Now   func() time.Time
	NewID func() (string, error)
}

// Call is what the layer above knows about an invocation before any board
// exists: where it is being made from, whose machine it is, and the
// identity the environment declared, empty when it declared none
// (docs/spec/invocacion.md#variables-de-entorno).
//
// It exists so that reading the machine's configuration is this layer's
// job and not the command line's. internal/cli depends on ops and on
// model, and on nothing below them, per section 3 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md:
// loading ~/.biso/config.json is opening a file of a board's world, so the
// call that does it belongs here.
type Call struct {
	// Dir is the working directory, absolute and clean.
	Dir string
	// Home is the caller's home directory, where the machine
	// configuration lives and where the upward search stops.
	Home string
	// Me is what BISO_ME said, and empty when it said nothing, in which
	// case the machine's `me` key answers instead.
	Me string
	// Limit is what BISO_LIMIT said, already read as a whole number of
	// rows, and nil when it said nothing, in which case the machine's
	// `default_limit` key answers instead.
	Limit *int

	Now   func() time.Time
	NewID func() (string, error)
}

// NewEnv reads this machine's configuration and answers the environment
// every command works in.
func NewEnv(c Call) (Env, error) {
	machine, err := board.LoadMachine(c.Home)
	if err != nil {
		return Env{}, err
	}
	me := machine.Me
	if c.Me != "" {
		me = c.Me
	}
	// The variable wins over the key, and the key is already filled in
	// with the built-in thirty when the file does not set it, so this one
	// assignment is the whole of the last two rungs of the precedence of
	// docs/spec/invocacion.md#variables-de-entorno. The first, --limit,
	// belongs to the call and not to the environment, so it is read where
	// the command line is.
	limit := machine.DefaultLimit
	if c.Limit != nil {
		limit = *c.Limit
	}
	return Env{
		Dir: c.Dir, Machine: machine, Me: me, ListLimit: &limit,
		Now: c.Now, NewID: c.NewID,
	}.WithDefaults(), nil
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
