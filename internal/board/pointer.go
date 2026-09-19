package board

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"biso/internal/model"
)

// This file holds the two small files that make a board findable: the
// project pointer of
// docs/spec/resolucion-del-tablero.md#el-fichero-bisojson-y-sus-claves, and
// the identity marker that sits next to the database.

const (
	// PointerFile is the name of the project pointer, the file `biso init`
	// writes in the project and that is versioned with it.
	PointerFile = ".biso.json"
	// PointerVersion is the version of the pointer's format that this
	// program writes.
	PointerVersion = 1
	// DatabaseFile is the name of a board's database. It is fixed and part
	// of the interface, because it is what makes a directory recognizable
	// as a board (docs/spec/cmd/init.md).
	DatabaseFile = "board.db"
	// MarkerContent is what a board's <id>.id marker holds: the version of
	// the store's format, which does not repeat the identifier so that
	// there are never two places that could say different things.
	MarkerContent = "{ \"storeVersion\": 1 }\n"
	// SnapshotTasksFile and SnapshotConfigFile are the two files
	// `biso snapshot` writes next to the marker. This package only ever
	// asks whether they are there, which is what tells the two variants of
	// the pointer_unresolved message apart.
	SnapshotTasksFile  = "snapshot.ndjson"
	SnapshotConfigFile = "board.json"
)

// Pointer is the .biso.json file of a project.
type Pointer struct {
	Version int
	// ID is the board's identity, eight lowercase hexadecimal characters.
	ID string
	// Path is the board directory, absolute or relative, and empty when the
	// board lives in one of the machine's roots. A relative one is resolved
	// against the directory that holds the pointer, never against the
	// working directory.
	Path string
}

// pointerShape is the file as it is written. Unknown keys are rejected,
// which is the rule of
// docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero.
type pointerShape struct {
	Version *int    `json:"version"`
	ID      *string `json:"id"`
	Path    string  `json:"path,omitempty"`
}

// ReadPointer reads the pointer at path.
func ReadPointer(path string) (*Pointer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, unreadableFile(path, err)
	}
	defer f.Close()

	var shape pointerShape
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&shape); err != nil {
		return nil, badConfigFile(path, err)
	}
	if shape.Version == nil {
		return nil, badConfigFile(path, fmt.Errorf("it has no version"))
	}
	if shape.ID == nil {
		return nil, badConfigFile(path, fmt.Errorf("it has no id"))
	}
	if !ValidID(*shape.ID) {
		return nil, badConfigValue(path, "id", *shape.ID, nil)
	}
	return &Pointer{Version: *shape.Version, ID: *shape.ID, Path: shape.Path}, nil
}

// WritePointer writes the pointer into dir.
//
// It is written on one line, exactly as
// docs/spec/resolucion-del-tablero.md prints it, because the file is
// versioned with the project: a shape that never changes is one line of diff
// when the board moves and none when it does not.
func WritePointer(dir string, p Pointer) error {
	text := fmt.Sprintf("{ \"version\": %d, \"id\": %q", p.Version, p.ID)
	if p.Path != "" {
		text += fmt.Sprintf(", \"path\": %q", p.Path)
	}
	text += " }\n"

	path := filepath.Join(dir, PointerFile)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return cannotWrite(path, err)
	}
	return nil
}

// ValidID answers whether s is a board identifier: eight lowercase
// hexadecimal characters. An uppercase spelling is invalid and is never
// normalized (docs/spec/resolucion-del-tablero.md#el-fichero-bisojson-y-sus-claves).
func ValidID(s string) bool {
	if len(s) != 8 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

// MarkerName is the name of the identity marker of a board: its id and the
// ".id" extension. Its name is the data.
func MarkerName(id string) string { return id + ".id" }

// WriteMarker writes the identity marker of id into dir.
func WriteMarker(dir, id string) error {
	path := filepath.Join(dir, MarkerName(id))
	if err := os.WriteFile(path, []byte(MarkerContent), 0o644); err != nil {
		return cannotWrite(path, err)
	}
	return nil
}

// HasDatabase answers whether dir holds a board's database, which is what
// the first way of docs/spec/resolucion-del-tablero.md#el-orden-de-búsqueda
// recognizes a board by.
func HasDatabase(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, DatabaseFile))
	return err == nil && info.Mode().IsRegular()
}

// HasMarker answers whether dir holds the identity marker of id.
func HasMarker(dir, id string) bool {
	info, err := os.Stat(filepath.Join(dir, MarkerName(id)))
	return err == nil && info.Mode().IsRegular()
}

// HasSnapshot answers whether dir holds the two files `biso snapshot`
// writes, which is what tells apart a half-written directory that has tasks
// to restore from one that was versioned before any snapshot was ever taken
// (docs/spec/resolucion-del-tablero.md#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina).
func HasSnapshot(dir string) bool {
	for _, name := range []string{SnapshotTasksFile, SnapshotConfigFile} {
		if info, err := os.Stat(filepath.Join(dir, name)); err != nil || !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}

// MarkerID answers the identity a directory claims through its marker, and
// the empty string when it has none or has more than one. More than one is
// not a board this function can name, and saying nothing here leaves the
// question to whoever opens the database, which holds the same id
// (docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero).
func MarkerID(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	found := ""
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".id") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".id")
		if !ValidID(id) {
			continue
		}
		if found != "" {
			return ""
		}
		found = id
	}
	return found
}

// cannotWrite is the environment failing on a write: exit code 8 of
// docs/spec/codigos-de-salida.md, which is what `biso init` answers when the
// directory it was pointed at cannot be written to.
func cannotWrite(path string, cause error) *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "io_error",
		Message:  fmt.Sprintf("%s cannot be written: %s", path, cause),
		Field:    path,
	}
}
