package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file walks the scenarios of docs/spec/resolucion-del-tablero.md from
// outside, through the program: it creates boards with `biso init` and asks
// `biso where` which one a call would use. They are the cases that page
// names one by one, and each test is one of them.

func TestABoardWithNoAtGoesToTheDefaultRootAndThePointerCarriesNoPath(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	m.run("init", "Mi Proyecto").assertCode(t, 0)

	// The folder is <slug>-<id>, and the slug collapses every run of
	// characters that are neither an ASCII letter nor an ASCII digit.
	dir := filepath.Join(m.boardsRoot(), "mi-proyecto-3f9a2b1c")
	if !m.exists(filepath.Join(dir, "board.db")) {
		t.Fatalf("the board is not at %s", dir)
	}
	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n",
		"the pointer of a board in the default root")

	// And it is found again from any subdirectory of the project.
	deep := m.at(filepath.Join(m.dir, "src", "cli"))
	got := deep.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "path     "+dir)
	assertContains(t, got.stdout, "source   project pointer at "+m.dir)
}

func TestTheSlugDropsDiacriticsAndCollapsesSeparators(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	m.run("init", "Peña 2026", "--prefix", "PENA").assertCode(t, 0)

	if !m.exists(filepath.Join(m.boardsRoot(), "pena-2026-3f9a2b1c", "board.db")) {
		t.Fatalf("the folder is not pena-2026-3f9a2b1c: %s",
			strings.Join(names(t, m.boardsRoot()), ", "))
	}
}

func TestAnAbsoluteAtIsFoundFromAnywhereWithTheCwdFlag(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	outside := filepath.Join(m.home, "boards", "elsewhere")

	m.run("init", "My project", "--at", outside).assertCode(t, 0)

	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \""+outside+"\" }\n",
		"the pointer of an absolute --at")

	// From a directory that shares no ancestor with the project, -C is all
	// it takes to work against that project's board.
	far := m.at(filepath.Join(m.home, "somewhere-else"))
	got := far.run("-C", m.dir, "where").assertCode(t, 0)
	assertContains(t, got.stdout, "path     "+outside)
}

func TestARelativeAtIsFoundFromAWorkingCopyInsideTheProject(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project", "--at", "board").assertCode(t, 0)

	// A git worktree under the project: it has the pointer, because that is
	// versioned, and not the board folder, because that is ignored. The
	// upward walk resolves the relative path against the ancestors of the
	// pointer's directory, which is what finds the board a level up.
	worktree := filepath.Join(m.dir, ".claude", "worktrees", "feature")
	m.at(worktree).writeFile(filepath.Join(worktree, ".biso.json"),
		m.read(filepath.Join(m.dir, ".biso.json")))

	got := m.at(worktree).run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "path     "+filepath.Join(m.dir, "board"))
	assertContains(t, got.stdout, "source   project pointer at "+worktree)
}

func TestARelativeAtIsNotFoundFromAWorkingCopyOutsideTheProject(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project", "--at", "board").assertCode(t, 0)

	outside := filepath.Join(m.home, "worktrees", "feature")
	other := m.at(outside)
	other.writeFile(filepath.Join(outside, ".biso.json"),
		m.read(filepath.Join(m.dir, ".biso.json")))

	got := other.run("where").assertCode(t, 20)
	assertContains(t, got.stderr, "this project's pointer names board 3f9a2b1c, which is not on this machine")
}

func TestTheUpwardSearchStopsAtTheHomeDirectory(t *testing.T) {
	m := newMachine(t)
	// A pointer one level above the home directory must never be reached.
	m.writeFile(filepath.Join(filepath.Dir(m.home), ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")

	got := m.run("where").assertCode(t, 20)
	assertContains(t, got.stderr, "no board here, and none configured for this project")
	assertContains(t, got.stderr, "pointer:        not found between this directory and "+m.home+",")
}

func TestADirectoryWithTheMarkerAndNoDatabaseIsSkipped(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	// What a working copy of a project that versions its board folder gets:
	// the marker, without the database that the exclusion file keeps out.
	half := filepath.Join(m.dir, "board")
	m.writeFile(filepath.Join(half, "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")

	// The pointer of this project has no path, so the search goes to the
	// roots and finds the real board: the half-written directory changes
	// nothing.
	got := m.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "path     "+filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c"))
}

func TestAHalfWrittenDirectoryIsNamedByTheErrorWhenNothingElseResolves(t *testing.T) {
	m := newMachine(t)
	half := filepath.Join(m.dir, "board")
	m.writeFile(filepath.Join(half, "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")
	m.writeFile(filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \"board\" }\n")

	got := m.run("where").assertCode(t, 20)
	assertContains(t, got.stderr, half+" has this board's marker,")
	assertContains(t, got.stderr, "versioned before any snapshot was ever taken")
	assertContains(t, got.stderr, "`biso init --at \""+half+"\"`")

	// With the two files of `biso snapshot` next to it, there are tasks to
	// recover and the remedy is the other one.
	m.writeFile(filepath.Join(half, "snapshot.ndjson"), "")
	m.writeFile(filepath.Join(half, "board.json"), "{}")
	got = m.run("where").assertCode(t, 20)
	assertContains(t, got.stderr, "and snapshot, versioned without its database")
	assertContains(t, got.stderr, "`biso init --from \""+half+"\"")
}

func TestTheSameIdInTwoRootsIsAnErrorThatChoosesNeither(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	// A second root, with a copy of the same board in it.
	extra := filepath.Join(m.home, "extra-root")
	copyDir(t, filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c"),
		filepath.Join(extra, "my-project-3f9a2b1c"))
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"),
		"{\"boards_extra_roots\": [\""+extra+"\"]}\n")

	got := m.run("where").assertCode(t, 22)
	assertContains(t, got.stderr, "board 3f9a2b1c is in two places, and biso will not choose between them")
	assertContains(t, got.stderr, "        "+filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c"))
	assertContains(t, got.stderr, "        "+filepath.Join(extra, "my-project-3f9a2b1c"))
	assertContains(t, got.stderr, "hint: rename or remove one of the two directories")
}

func TestInitAdoptsTheIdOfAPointerThatNamesAnAbsentBoard(t *testing.T) {
	m := newMachine(t).withIDs("ffffffff")
	pointer := "{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n"
	m.writeFile(filepath.Join(m.dir, ".biso.json"), pointer)

	m.run("init", "My project").assertCode(t, 0)

	// The identity is the one the pointer already named, and the pointer is
	// not rewritten, because it was already correct.
	if !m.exists(filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c", "board.db")) {
		t.Fatal("the board did not adopt the id of the pointer")
	}
	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")), pointer, "the pointer")
}

func TestInitAdoptsAWholeBoardAtWhenThisProjectResolvesToNone(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	// The pointer is lost, which is what a clone without it looks like.
	if err := os.Remove(filepath.Join(m.dir, ".biso.json")); err != nil {
		t.Fatal(err)
	}

	got := m.run("init", "--at", dir).assertCode(t, 0)
	assertContains(t, got.stdout, "Adopted board \"My project\"")
	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \""+dir+"\" }\n",
		"the pointer written by the recovery")
}

func TestInitOverAnAccessibleBoardIsTheBoardExistsError(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	got := m.run("init", "Another name").assertCode(t, 2)
	assertEqual(t, got.stderr,
		"error: this project already has board 3f9a2b1c, at "+dir+"\n"+
			"hint: `biso where` says which rule picked it\n"+
			"hint: --overwrite-config rewrites its configuration and never touches its tasks\n",
		"the board_exists message")
}

func TestTwoProjectsMayPointAtTheSameBoard(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	other := m.at(filepath.Join(m.home, "another-project"))
	other.writeFile(filepath.Join(other.dir, ".biso.json"),
		m.read(filepath.Join(m.dir, ".biso.json")))

	got := other.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "id       3f9a2b1c")
	assertContains(t, got.stdout, "source   project pointer at "+other.dir)
}

func TestTheWorkingDirectoryThatIsABoardWinsAndSaysWhatItDiscarded(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c", "7a1b2c3d")
	m.run("init", "My project").assertCode(t, 0)

	// A second board, and a call made from inside it while the project's
	// pointer names the first one.
	second := m.at(filepath.Join(m.home, "second-project"))
	second.run("init", "Other project").assertCode(t, 0)
	secondDir := filepath.Join(m.boardsRoot(), "other-project-7a1b2c3d")
	inside := m.at(secondDir)
	inside.writeFile(filepath.Join(secondDir, ".biso.json"),
		m.read(filepath.Join(m.dir, ".biso.json")))

	got := inside.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "id       7a1b2c3d")
	assertContains(t, got.stdout, "source   the working directory is this board")
	assertContains(t, got.stdout, "chosen     7a1b2c3d at "+secondDir)
	assertContains(t, got.stdout, "           because the working directory is itself a board")
	assertContains(t, got.stdout, "discarded  3f9a2b1c at "+
		filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c"))
	assertContains(t, got.stdout, "           the project pointer would have named this board instead")
}

func TestTheFiveExemptCallsWorkWithoutABoard(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	for _, argv := range [][]string{{"--help"}, {"--version"}, {"where", "--help"}, {"init", "--help"}} {
		m.run(argv...).assertCode(t, 0)
	}
	// where needs a board and answers for itself, with exit code 20.
	m.run("where").assertCode(t, 20)
	// init needs none, which is the whole point of it.
	m.run("init", "My project").assertCode(t, 0)
}

func names(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func copyDir(t *testing.T, from, to string) {
	t.Helper()
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(from)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(from, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(to, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("the output does not carry the line %q.\n--- got ---\n%s", want, got)
	}
}

func TestABoardIsNeverCreatedInsideAnother(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	// Called from inside the board itself, which is the first way of
	// choosing one: the board is that one, and creating another there is
	// the same error as anywhere else.
	inside := m.at(dir)
	got := inside.run("init", "Another").assertCode(t, 2)
	assertContains(t, got.stderr, "error: this project already has board 3f9a2b1c, at "+dir)

	// And --overwrite-config over it rewrites that board's configuration,
	// which is exactly what the flag means.
	got = inside.run("init", "--overwrite-config", "--types", "task").assertCode(t, 0)
	assertContains(t, got.stdout, "Rewrote the configuration of board \"My project\"")
}

func TestAtADirectoryWithTheMarkerAndNoDatabaseCreatesItThereAdoptingTheId(t *testing.T) {
	m := newMachine(t).withIDs("ffffffff")
	half := filepath.Join(m.dir, "board")
	m.writeFile(filepath.Join(half, "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")

	got := m.run("init", "My project", "--at", "board").assertCode(t, 0)

	assertContains(t, got.stdout, "Created board \"My project\"")
	if !m.exists(filepath.Join(half, "board.db")) {
		t.Fatal("the board was not created where --at named")
	}
	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \"board\" }\n",
		"the pointer of a board that adopted the id of a marker")
}

func TestARelativeAtThatLeavesTheProjectIsStoredAsItWasWritten(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	got := m.run("init", "My project", "--at", "../boards/my-project").assertCode(t, 0)

	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \"../boards/my-project\" }\n",
		"the pointer of a relative --at that leaves the project")
	if !m.exists(filepath.Join(m.home, "boards", "my-project", "board.db")) {
		t.Fatal("the board is not where --at named")
	}
	// It is outside the project, so only the note about the relative path
	// comes out, and not the one about living inside it.
	if strings.Contains(got.stderr, "the board lives inside this project") {
		t.Errorf("stderr = %q", got.stderr)
	}
	assertContains(t, got.stderr, "note: the location is stored as the relative path \"../boards/my-project\".")

	// And it resolves, because the relative position between the pointer
	// and the board is what it was.
	assertContains(t, m.run("where").assertCode(t, 0).stdout,
		"path     "+filepath.Join(m.home, "boards", "my-project"))
}

func TestMovingABoardByHandIsFixedByInitAtItsNewPath(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project", "--at", filepath.Join(m.home, "old")).assertCode(t, 0)

	moved := filepath.Join(m.home, "new")
	if err := os.Rename(filepath.Join(m.home, "old"), moved); err != nil {
		t.Fatal(err)
	}
	m.run("where").assertCode(t, 20)

	got := m.run("init", "--at", moved).assertCode(t, 0)
	assertContains(t, got.stdout, "Adopted board \"My project\"")
	assertContains(t, m.run("where").assertCode(t, 0).stdout, "path     "+moved)
}

func TestABoardWithNoMarkerStillOpensAndSaysItsId(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	// The first way of choosing a board settles for the database and does
	// not ask for the marker, precisely so that a board missing one can be
	// opened and repaired. The id then comes from inside the database,
	// where it is kept as well.
	if err := os.Remove(filepath.Join(dir, "3f9a2b1c.id")); err != nil {
		t.Fatal(err)
	}

	got := m.at(dir).run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "id       3f9a2b1c")
	assertContains(t, got.stdout, "source   the working directory is this board")
}

// TestAtAWholeBoardWhileThePointerNamesAnAbsentOneAdoptsTheDestination is
// the second exception of
// docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero
// in its second half: the exception applies "cuando no hay ningún puntero
// aquí, **o cuando el que hay no resuelve a nada en esta máquina**", and this
// is the second of those two. The destination's marker wins, and the pointer
// is rewritten, because there was no correct one to keep.
func TestAtAWholeBoardWhileThePointerNamesAnAbsentOneAdoptsTheDestination(t *testing.T) {
	m := newMachine(t).withIDs("7a1b2c3d")
	other := m.at(filepath.Join(m.home, "other-project"))
	other.run("init", "Other project", "--at", filepath.Join(m.home, "board")).assertCode(t, 0)
	dir := filepath.Join(m.home, "board")

	// This project's pointer names a board that is not on this machine.
	m.writeFile(filepath.Join(m.dir, ".biso.json"), "{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	m.run("where").assertCode(t, 20)

	got := m.run("init", "--at", dir).assertCode(t, 0)

	assertContains(t, got.stdout, "Adopted board \"Other project\"")
	assertEqual(t, m.read(filepath.Join(m.dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"7a1b2c3d\", \"path\": \""+dir+"\" }\n",
		"the pointer rewritten with the id of the destination")
	assertContains(t, m.run("where").assertCode(t, 0).stdout, "id       7a1b2c3d")
}

// TestAnIdIsNotMintedOverADirectoryThatOnlyHasTheMarker is the check
// docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero
// asks `biso init` to make before minting: an id that already exists in a
// root is taken, and a directory with the marker and no database is already
// claiming it, however little of a board it is.
func TestAnIdIsNotMintedOverADirectoryThatOnlyHasTheMarker(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c", "7a1b2c3d")
	m.writeFile(filepath.Join(m.boardsRoot(), "half-written", "3f9a2b1c.id"),
		"{ \"storeVersion\": 1 }\n")

	m.run("init", "My project").assertCode(t, 0)

	if !m.exists(filepath.Join(m.boardsRoot(), "my-project-7a1b2c3d", "board.db")) {
		t.Fatal("the minted id was the one a half-written directory already claimed")
	}
}

// TestAnAdoptedIdWalksTheRootsToo is the row of code 22 of
// docs/spec/cmd/init.md, which says `init` walks the roots for the id it is
// about to mint "o a adoptar": adopting the marker of a destination is no
// excuse for not looking.
func TestAnAdoptedIdWalksTheRootsToo(t *testing.T) {
	m := newMachine(t).withIDs("ffffffff")
	extra := filepath.Join(m.home, "extra-root")
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"),
		"{\"boards_extra_roots\": [\""+extra+"\"]}\n")
	for _, root := range []string{m.boardsRoot(), extra} {
		m.writeFile(filepath.Join(root, "copy-3f9a2b1c", "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")
	}

	// The destination carries that same identity, so adopting it would make
	// a third place with it.
	dir := filepath.Join(m.dir, "board")
	m.writeFile(filepath.Join(dir, "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")

	got := m.run("init", "My project", "--at", "board").assertCode(t, 22)
	assertContains(t, got.stderr, "board 3f9a2b1c is in two places, and biso will not choose between them")
}
