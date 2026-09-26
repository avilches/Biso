// Package destination talks to the biso board that is the target of an
// import, by running the compiled biso binary exactly the way any user
// would, never by importing biso's own internal packages. It reads exactly
// what docs/especificacion.md, "Qué le pregunta al destino", says this phase
// needs: the destination's vocabulary and the tasks that already exist on
// it.
//
// This package only reads and structures. It never converts a task, never
// decides an id collision, and never picks a destination vocabulary for a
// value that does not match: all of that is the conversion engine's job,
// built on top of what Read returns here.
package destination

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// Config is the slice of the destination's configuration this phase needs,
// read from "biso config list --json". docs/especificacion.md, "Qué le
// pregunta al destino", only asks for these four keys; every other key of
// the "data.config" object (project_name, labels, assignees, finish_strict,
// lease_minutes, urgency.*, and any key a future biso adds) is ignored, not
// an error.
type Config struct {
	TaskPrefix string
	Statuses   []string
	Types      []string
	Priorities []string
}

// Task is one task that already exists on the destination board, read from
// "biso export --out -". Only the four fields docs/especificacion.md, "Qué
// le pregunta al destino", names are kept.
type Task struct {
	ID        string
	Title     string
	CreatedAt string
	// Ordinal is the destination's manual order key, a base 36 string that
	// never has an arithmetic relationship with a Backlog.md ordinal
	// (docs/especificacion.md, "Orden manual"). It is empty when the task
	// has no manual order at all, which "biso export" represents as a JSON
	// null.
	Ordinal string
}

// Board is everything Read gathers from the destination.
type Board struct {
	Config Config
	Tasks  []Task
}

// Read runs biso (the path or name of the binary to execute) against the
// board that project resolves to, exactly the way biso --cwd <project>
// would, and returns its configuration and its tasks.
//
// It never writes to the destination board: this is a read-only snapshot of
// the moment it runs, as docs/especificacion.md, "Qué le pregunta al
// destino", describes.
//
// The three error shapes docs/decisiones and the phase 3 brief distinguish
// are all wrapped so the message names which one it is: biso could not be
// run at all (not found, not on PATH, not executable), biso ran but exited
// with a non-zero code (the message then includes that code and biso's own
// stderr), or biso's output could not be parsed as the JSON or NDJSON this
// package expects.
func Read(biso, project string) (Board, error) {
	configOutput, err := run(biso, project, "config", "list", "--json")
	if err != nil {
		return Board{}, fmt.Errorf("destination: reading the destination configuration: %w", err)
	}
	config, err := parseConfig(configOutput)
	if err != nil {
		return Board{}, fmt.Errorf("destination: reading the destination configuration: %w", err)
	}

	exportOutput, err := run(biso, project, "export", "--out", "-")
	if err != nil {
		return Board{}, fmt.Errorf("destination: reading the destination tasks: %w", err)
	}
	tasks, err := parseExport(exportOutput)
	if err != nil {
		return Board{}, fmt.Errorf("destination: reading the destination tasks: %w", err)
	}

	return Board{Config: config, Tasks: tasks}, nil
}

// run executes "<biso> --cwd <project> <args...>" and returns its standard
// output. It distinguishes, in the returned error's message, between biso
// not running at all and biso running but failing: the former wraps the
// original error with %w, the latter has no single error to wrap and
// instead states the exit code and biso's own stderr, trimmed, so the
// caller does not have to go looking for it.
func run(biso, project string, args ...string) ([]byte, error) {
	fullArgs := append([]string{"--cwd", project}, args...)
	cmd := exec.Command(biso, fullArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	command := "biso " + strings.Join(fullArgs, " ")

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf(
				"%s exited with code %d: %s",
				command, exitErr.ExitCode(), strings.TrimSpace(stderr.String()),
			)
		}
		return nil, fmt.Errorf("could not run %s: %w", command, err)
	}

	return stdout.Bytes(), nil
}

// configEnvelope is the shape of "biso config list --json" from
// docs/spec/cmd/config.md, "El esquema JSON", holding only the four keys
// this package needs. encoding/json ignores every key it does not know
// about, which is exactly what lets biso add new ones (project_name,
// urgency.*, and so on) without breaking this parser.
type configEnvelope struct {
	Data struct {
		Config struct {
			TaskPrefix string   `json:"task_prefix"`
			Statuses   []string `json:"statuses"`
			Types      []string `json:"types"`
			Priorities []string `json:"priorities"`
		} `json:"config"`
	} `json:"data"`
}

// parseConfig parses the output of "biso config list --json" into a Config.
func parseConfig(data []byte) (Config, error) {
	var envelope configEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return Config{}, fmt.Errorf(
			"could not parse the output of biso config list --json: %w", err,
		)
	}
	return Config{
		TaskPrefix: envelope.Data.Config.TaskPrefix,
		Statuses:   envelope.Data.Config.Statuses,
		Types:      envelope.Data.Config.Types,
		Priorities: envelope.Data.Config.Priorities,
	}, nil
}

// exportedTask is one line of "biso export --out -", holding only the four
// fields docs/especificacion.md, "Qué le pregunta al destino", names.
// Ordinal is a *string, never a number: docs/spec/cmd/export.md, "La
// garantía de simetría", says export writes it exactly as biso stores it, a
// base 36 string, and it comes out as a JSON null when the task has no
// manual order.
type exportedTask struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	CreatedAt string  `json:"createdAt"`
	Ordinal   *string `json:"ordinal"`
}

// parseExport parses the NDJSON of "biso export --out -" into a slice of
// Task, one per non-empty line, in the order biso wrote them.
func parseExport(data []byte) ([]Task, error) {
	var tasks []Task

	scanner := bufio.NewScanner(bytes.NewReader(data))
	// biso export can write a task with an arbitrarily long description or
	// body of comments on a single NDJSON line; the scanner's default
	// 64 KiB limit is not enough for that, so raise it well above what a
	// real board could produce.
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var t exportedTask
		if err := json.Unmarshal([]byte(line), &t); err != nil {
			return nil, fmt.Errorf(
				"could not parse line %d of the output of biso export as JSON: %w",
				lineNumber, err,
			)
		}

		ordinal := ""
		if t.Ordinal != nil {
			ordinal = *t.Ordinal
		}

		tasks = append(tasks, Task{
			ID:        t.ID,
			Title:     t.Title,
			CreatedAt: t.CreatedAt,
			Ordinal:   ordinal,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not read the output of biso export: %w", err)
	}

	return tasks, nil
}
