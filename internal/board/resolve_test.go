package board

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/model"
)

func TestSearchCap(t *testing.T) {
	for _, c := range []struct {
		name string
		dir  string
		home string
		want string
	}{
		{
			"inside the home directory it stops there",
			"/Users/avilches/Hub/Projects/Biso/src", "/Users/avilches", "/Users/avilches",
		},
		{
			"the home directory itself is already the cap",
			"/Users/avilches", "/Users/avilches", "/Users/avilches",
		},
		{
			"outside it, the second component of the path",
			"/Volumes/disco/proyecto", "/Users/avilches", "/Volumes/disco",
		},
		{
			"and the same with no home at all",
			"/opt/proyecto/src", "", "/opt/proyecto",
		},
		{
			"a path with a single component is its own cap",
			"/opt", "", "/opt",
		},
		{
			"a directory whose name merely starts like the home directory is not inside it",
			"/Users/avilches2/work", "/Users/avilches", "/Users/avilches2",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := SearchCap(c.dir, c.home); got != c.want {
				t.Errorf("SearchCap(%q, %q) = %q, want %q", c.dir, c.home, got, c.want)
			}
		})
	}
}

func TestAncestorsWalkUpToTheCapAndNoFurther(t *testing.T) {
	got := Ancestors("/Users/avilches/Hub/Projects/Biso/src", "/Users/avilches")
	want := []string{
		"/Users/avilches/Hub/Projects/Biso/src",
		"/Users/avilches/Hub/Projects/Biso",
		"/Users/avilches/Hub/Projects",
		"/Users/avilches/Hub",
		"/Users/avilches",
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("Ancestors = %v, want %v", got, want)
	}
}

func TestAncestorsOfADirectoryOutsideTheCapIsItself(t *testing.T) {
	got := Ancestors("/opt/other", "/Users/avilches")
	if len(got) != 1 || got[0] != "/opt/other" {
		t.Errorf("Ancestors = %v, want just the directory itself", got)
	}
}

func TestValidID(t *testing.T) {
	for _, c := range []struct {
		id   string
		want bool
	}{
		{"3f9a2b1c", true},
		{"00000000", true},
		{"3F9A2B1C", false}, // uppercase is invalid and never normalized
		{"3f9a2b1", false},
		{"3f9a2b1cc", false},
		{"3f9a2b1g", false},
		{"", false},
	} {
		if got := ValidID(c.id); got != c.want {
			t.Errorf("ValidID(%q) = %v, want %v", c.id, got, c.want)
		}
	}
}

func TestLoadMachineFillsInTheDefaults(t *testing.T) {
	home := t.TempDir()
	m, err := LoadMachine(home)
	if err != nil {
		t.Fatal(err)
	}
	if m.BoardsRoot != filepath.Join(home, ".biso", "boards") {
		t.Errorf("BoardsRoot = %q", m.BoardsRoot)
	}
	if m.VCS != "git" || m.DefaultLimit != 30 || m.Me != "" || len(m.ExtraRoots) != 0 {
		t.Errorf("machine = %+v, want the defaults of the specification", m)
	}
}

func TestLoadMachineExpandsTheTildeOfEveryRoot(t *testing.T) {
	home := t.TempDir()
	write(t, filepath.Join(home, ".biso", "config.json"),
		`{"boards_root": "~/tableros", "boards_extra_roots": ["~/otros", "/mnt/x"]}`)

	m, err := LoadMachine(home)
	if err != nil {
		t.Fatal(err)
	}
	if m.BoardsRoot != filepath.Join(home, "tableros") {
		t.Errorf("BoardsRoot = %q", m.BoardsRoot)
	}
	want := []string{filepath.Join(home, "otros"), "/mnt/x"}
	if strings.Join(m.ExtraRoots, "|") != strings.Join(want, "|") {
		t.Errorf("ExtraRoots = %v, want %v", m.ExtraRoots, want)
	}
	roots := m.Roots()
	if len(roots) != 3 || roots[0] != m.BoardsRoot {
		t.Errorf("Roots = %v, want the default root first", roots)
	}
}

func TestLoadMachineRejectsAnUnknownKeyAndAnUnknownVcs(t *testing.T) {
	for _, content := range []string{
		`{"boards_roots": "/tmp"}`,
		`{"vcs": "fossil"}`,
		`{"default_limit": -1}`,
		`not json at all`,
	} {
		home := t.TempDir()
		write(t, filepath.Join(home, ".biso", "config.json"), content)
		_, err := LoadMachine(home)
		if err == nil {
			t.Fatalf("%s was accepted", content)
		}
		assertExit(t, err, 3, "bad_config_value")
	}
}

func TestReadPointerRejectsWhatTheSpecificationRejects(t *testing.T) {
	for _, content := range []string{
		`{"version": 1, "id": "3f9a2b1c", "extra": 1}`,
		`{"version": 1, "id": "3F9A2B1C"}`,
		`{"id": "3f9a2b1c"}`,
		`{"version": 1}`,
	} {
		dir := t.TempDir()
		path := filepath.Join(dir, PointerFile)
		write(t, path, content)
		if _, err := ReadPointer(path); err == nil {
			t.Fatalf("%s was accepted", content)
		}
	}
}

func TestWriteAndReadPointerRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := WritePointer(dir, Pointer{Version: 1, ID: "3f9a2b1c", Path: "board"}); err != nil {
		t.Fatal(err)
	}
	p, err := ReadPointer(filepath.Join(dir, PointerFile))
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != 1 || p.ID != "3f9a2b1c" || p.Path != "board" {
		t.Errorf("pointer = %+v", p)
	}
}

func TestMarkerIDAnswersNothingWhenThereIsMoreThanOne(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "3f9a2b1c.id"), MarkerContent)
	if got := MarkerID(dir); got != "3f9a2b1c" {
		t.Errorf("MarkerID = %q", got)
	}
	write(t, filepath.Join(dir, "7a1b2c3d.id"), MarkerContent)
	if got := MarkerID(dir); got != "" {
		t.Errorf("MarkerID = %q, want nothing with two markers", got)
	}
}

// write puts a file on disk, making its directory if it is not there.
func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// assertExit checks an error is the case of the specification it should be,
// by its exit code and its stable identifier and never by its prose.
func assertExit(t *testing.T, err error, code int, identifier string) {
	t.Helper()
	e, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	if e.ExitCode != code || e.Code != identifier {
		t.Errorf("error = %d/%s, want %d/%s", e.ExitCode, e.Code, code, identifier)
	}
}
