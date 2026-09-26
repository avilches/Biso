// Package source reads a Backlog.md board (the directory that holds
// tasks/, completed/, and the rest of the layout Backlog.md writes) into an
// in-memory representation that keeps every known field and section
// verbatim, exactly as docs/especificacion.md, "Qué lee del origen" and "El
// mapeo de campos", describe them.
//
// This package only reads and structures. It never rewrites an id, never
// adds a milestone:: or project:: label, never converts a date to UTC, and
// never picks a destination vocabulary: all of that is the conversion
// engine's job, built on top of what Read returns here. Anything this
// package cannot carry over as-is (an unparseable frontmatter, an
// unrecognized key or section, a date with the wrong shape, and so on)
// becomes a Finding instead of being silently dropped or aborting the read.
//
// Read never prints anything and never decides a process exit code; both of
// those belong to a later phase.
package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Board is everything Read gathers from a Backlog.md board.
type Board struct {
	// Tasks holds every task read successfully from tasks/, completed/,
	// and archive/tasks/. A task whose frontmatter could not be parsed
	// never appears here; it only produces a Finding. The slice order
	// is stable across two reads of the same board but carries no other
	// meaning: a later phase sorts by the source id's number.
	Tasks []Task
	// Milestones maps a milestone id (as it appears in a task's
	// milestone field, for example "m-4") to its title, read from
	// milestones/ and archive/milestones/.
	Milestones map[string]string
	// Findings accumulates every finding raised while reading the
	// board: broken frontmatter, unrecognized frontmatter keys or body
	// sections, invalid date shapes, joined multi-line criteria,
	// comments without a created date, and non-empty drafts/docs/
	// decisions folders.
	Findings []Finding
}

// Read reads the Backlog.md board rooted at backlogDir (the directory that
// contains tasks/, the one usually called "backlog") and returns its Board.
//
// backlogDir itself must exist; Read returns an error otherwise. Every
// subdirectory Read looks at (tasks/, completed/, archive/tasks/,
// milestones/, archive/milestones/, drafts/, archive/drafts/, docs/,
// decisions/) is optional: a missing one is treated as empty, not as an
// error, matching "vacía o ausente, no produce nada" for the folders that
// only produce a count.
//
// Read never touches backlogDir's config.yml or a sibling
// backlog.config.yml: docs/especificacion.md, "Qué lee del origen", says
// nothing from there is needed.
func Read(backlogDir string) (Board, error) {
	info, err := os.Stat(backlogDir)
	if err != nil {
		return Board{}, fmt.Errorf("source: cannot read backlog directory: %w", err)
	}
	if !info.IsDir() {
		return Board{}, fmt.Errorf("source: %s is not a directory", backlogDir)
	}

	board := Board{
		Milestones: make(map[string]string),
	}

	milestoneDirs := []string{
		filepath.Join(backlogDir, "milestones"),
		filepath.Join(backlogDir, "archive", "milestones"),
	}
	for _, dir := range milestoneDirs {
		titles, findings, err := readMilestones(dir)
		if err != nil {
			return Board{}, err
		}
		for id, title := range titles {
			board.Milestones[id] = title
		}
		board.Findings = append(board.Findings, findings...)
	}

	taskDirs := []struct {
		dir      string
		archived bool
	}{
		{filepath.Join(backlogDir, "tasks"), false},
		{filepath.Join(backlogDir, "completed"), false},
		{filepath.Join(backlogDir, "archive", "tasks"), true},
	}
	for _, td := range taskDirs {
		tasks, findings, err := readTaskDir(td.dir, td.archived)
		if err != nil {
			return Board{}, err
		}
		board.Tasks = append(board.Tasks, tasks...)
		board.Findings = append(board.Findings, findings...)
	}

	unreadDirs := []struct {
		dir   string
		field string
		label string
	}{
		{filepath.Join(backlogDir, "drafts"), "drafts", "drafts/"},
		{filepath.Join(backlogDir, "archive", "drafts"), "drafts", "archive/drafts/"},
		{filepath.Join(backlogDir, "docs"), "docs", "docs/"},
		{filepath.Join(backlogDir, "decisions"), "decisions", "decisions/"},
	}
	for _, ud := range unreadDirs {
		count, err := countFiles(ud.dir)
		if err != nil {
			return Board{}, err
		}
		if count > 0 {
			board.Findings = append(board.Findings, Finding{
				File:  "-",
				Field: ud.field,
				Message: fmt.Sprintf(
					"%d file(s) found in %s, not read (not a task, no biso equivalent)",
					count, ud.label,
				),
			})
		}
	}

	return board, nil
}

// countFiles returns how many regular files sit directly inside dir. A
// missing dir counts as zero files, not an error.
func countFiles(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("source: %w", err)
	}
	count := 0
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			count++
		}
	}
	return count, nil
}

// readTaskDir reads every ".md" file directly inside dir as a task. A
// missing dir yields no tasks and no findings.
func readTaskDir(dir string, archived bool) ([]Task, []Finding, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil
		}
		return nil, nil, fmt.Errorf("source: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if entry.Type().IsRegular() && filepath.Ext(entry.Name()) == ".md" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	var tasks []Task
	var findings []Finding
	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, nil, fmt.Errorf("source: %w", err)
		}
		task, taskFindings, skipped := parseTask(name, string(content), archived)
		findings = append(findings, taskFindings...)
		if !skipped {
			tasks = append(tasks, task)
		}
	}
	return tasks, findings, nil
}
