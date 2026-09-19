package board

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"biso/internal/model"
)

// This file holds the machine configuration of
// docs/spec/invocacion.md#configuración-de-máquina: the file every board of
// this machine shares, which has to be readable before the first board
// exists and therefore never goes through the board resolution.

// MachineDir is the directory that holds the machine configuration and the
// default boards root, relative to the caller's home directory.
const MachineDir = ".biso"

// MachineFile is the name of the machine configuration file inside
// MachineDir.
const MachineFile = "config.json"

// DefaultLimit is the fallback of the default_limit key, the number of rows
// biso ls prints when neither --limit, nor BISO_LIMIT, nor the key say
// otherwise (docs/spec/invocacion.md#variables-de-entorno).
const DefaultLimit = 30

// Machine is the machine configuration, already resolved: every path is
// absolute and every default is filled in.
//
// The version control keys travel as plain data, without the types of
// internal/vcs, because the dependency rule of the architecture keeps this
// package below that one. internal/ops, which sits above both, is what
// turns them into a vcs.Config.
type Machine struct {
	// BoardsRoot is the directory where `biso init` without --at creates a
	// new board, and the first of the roots a search by id walks.
	BoardsRoot string
	// ExtraRoots are the further roots, in the order they were written.
	ExtraRoots []string
	// VCS is "git", "none" or "custom".
	VCS string
	// VCSCustom is the vcs_custom object, and only with VCS set to
	// "custom".
	VCSCustom MachineVCSCustom
	// Me is the identity of whoever uses biso on this machine, empty when
	// the key is absent.
	Me string
	// DefaultLimit is the default_limit key.
	DefaultLimit int
	// Home is the caller's home directory, which is also the cap of the
	// upward search (docs/spec/resolucion-del-tablero.md#el-tope-de-la-búsqueda-hacia-arriba).
	Home string
}

// MachineVCSCustom is the vcs_custom object of
// docs/spec/invocacion.md#configuración-de-máquina.
type MachineVCSCustom struct {
	Commit     []string
	Publish    []string
	IgnoreFile string
}

// machineFileShape is the file as it is written, with every key optional and
// every unknown one rejected. json.Decoder.DisallowUnknownFields is what
// makes the rule of the specification true without listing the known keys
// twice.
type machineFileShape struct {
	BoardsRoot       *string  `json:"boards_root"`
	BoardsExtraRoots []string `json:"boards_extra_roots"`
	VCS              *string  `json:"vcs"`
	VCSCustom        *struct {
		Commit     []string `json:"commit"`
		Publish    []string `json:"publish"`
		IgnoreFile string   `json:"ignore_file"`
	} `json:"vcs_custom"`
	Me           *string `json:"me"`
	DefaultLimit *int    `json:"default_limit"`
}

// LoadMachine reads ~/.biso/config.json, filling in the defaults of the
// table of docs/spec/invocacion.md#configuración-de-máquina. A file that is
// not there is not an error: it means every key is at its default, which is
// the state of a machine where nobody has configured anything.
func LoadMachine(home string) (Machine, error) {
	m := Machine{
		BoardsRoot:   filepath.Join(home, MachineDir, "boards"),
		VCS:          "git",
		DefaultLimit: DefaultLimit,
		Home:         home,
	}
	path := filepath.Join(home, MachineDir, MachineFile)
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return m, nil
		}
		return Machine{}, unreadableFile(path, err)
	}
	defer f.Close()

	var shape machineFileShape
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&shape); err != nil {
		return Machine{}, badConfigFile(path, err)
	}

	if shape.BoardsRoot != nil {
		m.BoardsRoot = expandHome(*shape.BoardsRoot, home)
	}
	for _, root := range shape.BoardsExtraRoots {
		m.ExtraRoots = append(m.ExtraRoots, expandHome(root, home))
	}
	if shape.VCS != nil {
		switch *shape.VCS {
		case "git", "none", "custom":
			m.VCS = *shape.VCS
		default:
			return Machine{}, badConfigValue(path, "vcs", *shape.VCS,
				[]string{"git", "none", "custom"})
		}
	}
	if shape.VCSCustom != nil {
		m.VCSCustom = MachineVCSCustom{
			Commit:     shape.VCSCustom.Commit,
			Publish:    shape.VCSCustom.Publish,
			IgnoreFile: shape.VCSCustom.IgnoreFile,
		}
	}
	if shape.Me != nil {
		m.Me = *shape.Me
	}
	if shape.DefaultLimit != nil {
		if *shape.DefaultLimit < 0 {
			return Machine{}, badConfigValue(path, "default_limit",
				fmt.Sprint(*shape.DefaultLimit), nil)
		}
		m.DefaultLimit = *shape.DefaultLimit
	}
	return m, nil
}

// Roots are the directories a search by board id walks, in the declared
// order: boards_root first and boards_extra_roots after, each one in the
// order it was written
// (docs/spec/invocacion.md#configuración-de-máquina).
func (m Machine) Roots() []string {
	roots := make([]string, 0, 1+len(m.ExtraRoots))
	roots = append(roots, m.BoardsRoot)
	roots = append(roots, m.ExtraRoots...)
	return roots
}

// expandHome turns a leading "~/" into the caller's home directory, which
// docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero asks for on the
// pointer's path and docs/spec/invocacion.md writes in the default of
// boards_root.
func expandHome(path, home string) string {
	switch {
	case path == "~":
		return home
	case strings.HasPrefix(path, "~/"):
		return filepath.Join(home, path[2:])
	}
	return path
}

// badConfigFile is a configuration file that cannot be read as what it says
// it is: a key the program does not know, or JSON that does not parse. It is
// exit code 3, the direction of that code that covers a stored value the
// program cannot interpret (docs/spec/codigos-de-salida.md#el-código-3-cubre-dos-direcciones),
// and never something ignored in silence.
func badConfigFile(path string, cause error) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message:  fmt.Sprintf("%s cannot be read: %s", path, cause),
		Field:    path,
	}
}

func badConfigValue(path, key, given string, valid []string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message:  fmt.Sprintf("%s: %s cannot be %q", path, key, given),
		Field:    key,
		Given:    given,
		Valid:    valid,
	}
}

// unreadableFile is the environment failing rather than the request: a file
// that is there and that the process cannot open.
func unreadableFile(path string, cause error) *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "file_unreadable",
		Message:  fmt.Sprintf("%s cannot be read: %s", path, cause),
		Field:    path,
	}
}
