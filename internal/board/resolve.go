package board

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"biso/internal/model"
)

// This file is docs/spec/resolucion-del-tablero.md: the two ways of finding
// the board a call refers to, the cap of the upward search, and the three
// bad endings of that search (exit codes 20, 21 and 22 of
// docs/spec/codigos-de-salida.md; the 21 is born in internal/store, when the
// database it points at turns out not to open).

// Way is which of the two ways of
// docs/spec/resolucion-del-tablero.md#el-orden-de-búsqueda found the board.
type Way int

const (
	// WayWorkingDirectory is the first way: the working directory is
	// itself a board directory, recognized by holding board.db. It wins
	// always, because it is the more specific of the two.
	WayWorkingDirectory Way = iota + 1
	// WayPointer is the second way: a .biso.json in the working directory
	// or in one of its ancestors says which board this project has.
	WayPointer
)

// Location is a board that has been found: which one, where, and by which
// way.
type Location struct {
	// ID is the board's identity. It is empty only when the first way
	// found the board and the directory carries no marker to read it from,
	// in which case Open reads it from the database.
	ID string
	// Dir is the board's directory, already resolved and expanded: never
	// the literal text a pointer carried (docs/spec/cmd/where.md).
	Dir string
	Way Way
	// PointerDir is the directory the pointer came from, and empty with
	// the first way.
	PointerDir string
	// Discarded are the candidates this resolution turned down, which
	// today is at most one (docs/spec/cmd/where.md#cuando-hay-más-de-un-candidato).
	Discarded []Discarded
}

// Discarded is one candidate that was not chosen, with the reason in the
// free prose that `biso where` prints under it.
type Discarded struct {
	ID     string
	Dir    string
	Reason string
}

// Source is the sentence of the `source` row of `biso where`, which names
// the way and, with the pointer, the directory that pointer came from.
func (l *Location) Source() string {
	if l.Way == WayWorkingDirectory {
		return "the working directory is this board"
	}
	return "project pointer at " + l.PointerDir
}

// Search is the question the resolution answers: from which directory, on
// which machine.
type Search struct {
	// Dir is the working directory, which is the current directory unless
	// -C or BISO_CWD said another one
	// (docs/spec/cmd/flags-globales.md#flags-globales).
	Dir     string
	Machine Machine
}

// Searched is what the search looked at. It is what `biso where` prints
// when it finds nothing, and what tells the three variants of the
// pointer_unresolved message apart.
type Searched struct {
	// Dir is where the search started and Stop where it stopped going up.
	Dir  string
	Stop string
	// WorkingDirectoryIsBoard says the first way found a board.
	WorkingDirectoryIsBoard bool
	// PointerDir and PointerID are the pointer that was found, if any.
	PointerDir string
	PointerID  string
	// HalfWritten is a directory carrying this board's marker without its
	// database, the one a project that versions its board folder leaves
	// behind, and HalfWrittenHasSnapshot says whether it also carries the
	// two files of `biso snapshot`.
	HalfWritten            string
	HalfWrittenHasSnapshot bool
}

// Resolve finds the board of a call, following
// docs/spec/resolucion-del-tablero.md#el-orden-de-búsqueda.
//
// The error it answers when nothing is found is the generic one every
// command prints; `biso where` replaces it with its own, which says what was
// searched, and recognizes it by its code (docs/spec/cmd/where.md).
func Resolve(s Search) (*Location, Searched, *model.Error) {
	facts := Searched{Dir: s.Dir, Stop: SearchCap(s.Dir, s.Machine.Home)}

	var here *Location
	if HasDatabase(s.Dir) {
		facts.WorkingDirectoryIsBoard = true
		here = &Location{ID: MarkerID(s.Dir), Dir: s.Dir, Way: WayWorkingDirectory}
	}

	// The pointer is read before either way is chosen, because a pointer
	// that cannot be read is exit code 3 whichever way ends up winning
	// (docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero).
	// Discarding a malformed one in silence because the working directory
	// happens to be a board is the very thing that page forbids: there is a
	// pointer here and what happens is that it is written wrong.
	p, pointerDir, readErr := findAndReadPointer(s.Dir, facts.Stop)
	if readErr != nil {
		return nil, facts, readErr
	}

	var viaPointer *Location
	var perr *model.Error
	if p != nil {
		facts.PointerDir, facts.PointerID = pointerDir, p.ID
		viaPointer, perr = boardOfPointer(s, *p, pointerDir, &facts)
	}

	if here != nil {
		// The first way wins always, and the only conflict there can be
		// between the two ways is this one: the pointer would have named
		// a different board (docs/spec/cmd/where.md#cuando-hay-más-de-un-candidato).
		// A pointer that resolves to nothing is not a candidate, so its
		// own error is not this call's business.
		if viaPointer != nil && viaPointer.Dir != here.Dir {
			here.Discarded = append(here.Discarded, Discarded{
				ID:     viaPointer.ID,
				Dir:    viaPointer.Dir,
				Reason: "the project pointer would have named this board instead",
			})
		}
		return here, facts, nil
	}
	if perr != nil {
		return nil, facts, perr
	}
	if viaPointer != nil {
		return viaPointer, facts, nil
	}
	return nil, facts, NoBoardError()
}

// findAndReadPointer answers the project's pointer, the directory it came
// from, and the error of one that is there and cannot be read. A project
// with no pointer at all answers three zero values, which is not an error.
func findAndReadPointer(dir, stop string) (*Pointer, string, *model.Error) {
	pointerDir, path := findPointer(dir, stop)
	if path == "" {
		return nil, "", nil
	}
	p, err := ReadPointer(path)
	if err != nil {
		return nil, "", err.(*model.Error)
	}
	return p, pointerDir, nil
}

// boardOfPointer is the second half of the second way: the board the
// pointer names, looked for where its path says and then in the machine's
// roots.
func boardOfPointer(s Search, p Pointer, dir string, facts *Searched) (*Location, *model.Error) {
	if p.Path != "" {
		for _, candidate := range pointerCandidates(p.Path, dir, facts.Stop, s.Machine.Home) {
			if HasMarker(candidate, p.ID) {
				if HasDatabase(candidate) {
					return &Location{ID: p.ID, Dir: candidate, Way: WayPointer, PointerDir: dir}, nil
				}
				// A directory with the marker and no database is not a
				// board, and the search goes on. Remembering the first one
				// is what lets the error say where the tasks went
				// (docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero).
				if facts.HalfWritten == "" {
					facts.HalfWritten = candidate
					facts.HalfWrittenHasSnapshot = HasSnapshot(candidate)
				}
			}
		}
	}

	// Either the pointer carries no path, or the one it carries did not
	// resolve: both end up walking the machine's roots looking for the
	// marker (docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero).
	hits, err := searchRoots(s.Machine, p.ID, true)
	if err != nil {
		return nil, err.(*model.Error)
	}
	if len(hits) == 1 {
		return &Location{ID: p.ID, Dir: hits[0], Way: WayPointer, PointerDir: dir}, nil
	}
	return nil, PointerUnresolvedError(*facts)
}

// findPointer walks from dir up to stop and answers the first directory that
// holds a pointer, with the path of the file itself.
func findPointer(dir, stop string) (string, string) {
	for _, d := range Ancestors(dir, stop) {
		path := filepath.Join(d, PointerFile)
		if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
			return d, path
		}
	}
	return "", ""
}

// pointerCandidates are the directories a pointer's path names, in the
// order they are tried. An absolute path names exactly one. A relative one
// is resolved against the directory that holds the pointer, and then,
// unchanged, against each of that directory's ancestors up to the cap: that
// last step is what keeps a board that lives inside the project reachable
// from a working copy that does not have the folder, such as a git worktree
// (docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero).
func pointerCandidates(path, pointerDir, stop, home string) []string {
	path = expandHome(path, home)
	if filepath.IsAbs(path) {
		return []string{filepath.Clean(path)}
	}
	var out []string
	for _, d := range Ancestors(pointerDir, stop) {
		out = append(out, filepath.Clean(filepath.Join(d, path)))
	}
	return out
}

// searchRoots walks the machine's roots looking for the directory that
// carries the marker of id, and answers every one it finds. A root that
// does not exist or cannot be read is not an error: a machine can have a
// disk configured that is not mounted today
// (docs/spec/invocacion.md#configuración-de-máquina).
//
// withDatabase says whether a directory has to carry the database as well
// to count, and the two callers of this function want different answers.
// Choosing a board to work on asks for both, because a directory with the
// marker and no database is not a board and the search goes on past it.
// Asking whether an identity is already taken settles for the marker,
// because that half-written directory is already claiming that id, and
// minting it again would make the two of them the duplicate of
// docs/spec/resolucion-del-tablero.md#el-mismo-id-en-dos-sitios.
func searchRoots(m Machine, id string, withDatabase bool) ([]string, error) {
	var hits []string
	for _, root := range m.Roots() {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			dir := filepath.Join(root, e.Name())
			if HasMarker(dir, id) && (!withDatabase || HasDatabase(dir)) {
				hits = append(hits, dir)
			}
		}
	}
	if len(hits) > 1 {
		return nil, AmbiguousBoardError(id, hits)
	}
	return hits, nil
}

// FindID answers where a board id already lives on this machine, which is
// what `biso init` asks before minting or adopting one
// (docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero).
// The marker alone is enough to answer yes, per the rule above.
func FindID(m Machine, id string) ([]string, error) { return searchRoots(m, id, false) }

// Ancestors is the walk upward: dir, its parent, and so on, up to and
// including stop. A dir that is not under stop answers just itself, so the
// walk can never leave the area the cap protects.
func Ancestors(dir, stop string) []string {
	dir = filepath.Clean(dir)
	stop = filepath.Clean(stop)
	out := []string{dir}
	if dir == stop {
		return out
	}
	if !within(dir, stop) {
		return out
	}
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return out
		}
		out = append(out, parent)
		if parent == stop {
			return out
		}
		dir = parent
	}
}

// SearchCap is docs/spec/resolucion-del-tablero.md#el-tope-de-la-búsqueda-hacia-arriba:
// the caller's home directory when the working directory is inside it, and
// otherwise the second component of the path.
//
// The second component is not a habit of counting levels: the first
// component of an absolute path is always a system directory or the one
// that holds everybody's home directories, and a pointer there cannot be on
// purpose, only by accident.
func SearchCap(dir, home string) string {
	dir = filepath.Clean(dir)
	if home != "" {
		home = filepath.Clean(home)
		if dir == home || within(dir, home) {
			return home
		}
	}
	volume := filepath.VolumeName(dir)
	rest := strings.TrimPrefix(dir[len(volume):], string(filepath.Separator))
	parts := strings.Split(rest, string(filepath.Separator))
	if len(parts) < 2 {
		return dir
	}
	return filepath.Join(volume+string(filepath.Separator), parts[0], parts[1])
}

// within answers whether dir is strictly below root.
func within(dir, root string) bool {
	return strings.HasPrefix(dir, root+string(filepath.Separator))
}

// NoBoardError is the message every command but the five exempt ones prints
// when the search finds nothing
// (docs/spec/resolucion-del-tablero.md#cuando-no-hay-tablero-que-encontrar).
func NoBoardError() *model.Error {
	return &model.Error{
		ExitCode: 20,
		Code:     "no_board",
		Message:  "no board here, and none configured for this project",
		Hints:    []string{"`biso init` creates one, `biso where` explains what was searched"},
	}
}

// PointerUnresolvedError is the message of
// docs/spec/resolucion-del-tablero.md#el-puntero-nombra-un-tablero-que-no-está-en-esta-máquina,
// in its three variants: the generic one, and the two that name a directory
// carrying the marker without the database, which differ in whether there
// is anything to restore from it.
//
// The line breaks inside each hint are the specification's own and do not
// depend on how long the interpolated path is, the same way the two notes of
// docs/spec/cmd/init.md break between the same words whatever folder name
// they name. Whoever prints them aligns every line after the first under the
// "hint: " prefix.
func PointerUnresolvedError(facts Searched) *model.Error {
	e := &model.Error{
		ExitCode: 20,
		Code:     "pointer_unresolved",
		Message: fmt.Sprintf(
			"this project's pointer names board %s, which is not on this machine",
			facts.PointerID),
	}
	switch {
	case facts.HalfWritten == "":
		e.Hints = []string{fmt.Sprintf("`biso init` creates it here, adopting id %s", facts.PointerID)}
	case facts.HalfWrittenHasSnapshot:
		e.Hints = []string{
			fmt.Sprintf("%s has this board's marker\nand snapshot, versioned without its database",
				facts.HalfWritten),
			fmt.Sprintf("`biso init --from %q\n--at %q` restores it\nthere, adopting id %s",
				facts.HalfWritten, facts.HalfWritten, facts.PointerID),
		}
	default:
		e.Hints = []string{
			fmt.Sprintf("%s has this board's marker,\nversioned before any snapshot was ever taken",
				facts.HalfWritten),
			fmt.Sprintf("`biso init --at %q`\ncreates it there, adopting id %s",
				facts.HalfWritten, facts.PointerID),
		}
	}
	return e
}

// AmbiguousBoardError is docs/spec/resolucion-del-tablero.md#el-mismo-id-en-dos-sitios:
// the same identity in more than one place, which biso will not choose
// between, because choosing would be writing into a board nobody named.
func AmbiguousBoardError(id string, dirs []string) *model.Error {
	sorted := append([]string(nil), dirs...)
	sort.Strings(sorted)

	places := "two places"
	if len(sorted) != 2 {
		places = fmt.Sprintf("%d places", len(sorted))
	}
	detail := make([]string, 0, len(sorted))
	for _, dir := range sorted {
		detail = append(detail, "        "+dir)
	}
	hint := "rename or remove one of the two directories"
	if len(sorted) != 2 {
		hint = "rename or remove all but one of those directories"
	}
	return &model.Error{
		ExitCode: 22,
		Code:     "ambiguous_board_id",
		Message: fmt.Sprintf("board %s is in %s, and biso will not choose between them",
			id, places),
		Detail: detail,
		Hints:  []string{hint},
	}
}
