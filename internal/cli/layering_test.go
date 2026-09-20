package cli

import (
	goparser "go/parser"
	"go/token"
	"io/fs"
	"strconv"
	"strings"
	"testing"
)

// TestThisLayerImportsOpsAndModelAndNothingElseOfOurs is the dependency rule
// of section 3 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md
// made checkable: `internal/cli` depends on `ops` and on `model`, and a
// package of below is never reached around them. Reading the machine
// configuration used to be the exception, and it was an exception by
// accident and not by design: it now goes through internal/ops like
// everything else a command needs from a board's world.
func TestThisLayerImportsOpsAndModelAndNothingElseOfOurs(t *testing.T) {
	allowed := map[string]bool{"biso/internal/ops": true, "biso/internal/model": true}

	set := token.NewFileSet()
	packages, err := goparser.ParseDir(set, ".", func(f fs.FileInfo) bool {
		return !strings.HasSuffix(f.Name(), "_test.go")
	}, goparser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range packages {
		for name, file := range pkg.Files {
			for _, imported := range file.Imports {
				path, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				if strings.HasPrefix(path, "biso/") && !allowed[path] {
					t.Errorf("%s imports %s, and this layer only imports ops and model",
						name, path)
				}
			}
		}
	}
}
