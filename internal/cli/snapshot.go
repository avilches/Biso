package cli

import (
	"fmt"

	"biso/internal/ops"
)

// This file is the output of docs/spec/cmd/snapshot.md: the lines that say
// what was written and where the revision went, the forwarding of whatever
// the version control system printed, and the `snapshot` envelope.

func runSnapshot(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	if len(p.Positionals) > 0 {
		return fail(s, asJSON, errUnexpectedArgument(p.Positionals[0]), warningsOf(p))
	}
	mode := ops.DefaultVCSMode
	if v, ok := p.Value("vcs"); ok {
		parsed, err := ops.ParseVCSMode(v)
		if err != nil {
			return fail(s, asJSON, err, warningsOf(p))
		}
		mode = parsed
	}

	result, err := ops.Snapshot(env, ops.SnapshotParams{Mode: mode})
	if err != nil {
		if result != nil && !asJSON {
			// Whatever the orders wrote comes out before the error line,
			// which is the order docs/spec/cmd/snapshot.md prints them in.
			printVCSOutput(s, result)
		}
		return fail(s, asJSON, err, opsWarnings(vcsWarnings(result)))
	}
	printWarnings(s, p)
	printOpsWarnings(s, result.Warnings)
	if !asJSON {
		printVCSOutput(s, result)
	}

	if asJSON {
		writeEnvelope(s, env, "snapshot", snapshotData(result))
	} else {
		fmt.Fprint(s.Stdout, renderSnapshot(result))
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	if len(result.Skipped) > 0 {
		return 6
	}
	return 0
}

// printVCSOutput forwards, line by line, what the orders of the version
// control system wrote, each one behind the name of the system and a colon.
// It is never suppressed, not by --quiet and not by anything else: it is the
// output of a program that is not biso
// (docs/spec/cmd/snapshot.md#cómo-se-ejecutan-las-órdenes).
func printVCSOutput(s Streams, result *ops.SnapshotResult) {
	if result == nil {
		return
	}
	for _, line := range result.VCSOutput {
		fmt.Fprintf(s.Stderr, "%s: %s\n", result.VCS, line)
	}
}

// vcsWarnings are the warnings a failed snapshot had already earned, which
// a failure never swallows.
func vcsWarnings(result *ops.SnapshotResult) []ops.Warning {
	if result == nil {
		return nil
	}
	return result.Warnings
}

func opsWarnings(warnings []ops.Warning) []Warning {
	out := make([]Warning, 0, len(warnings))
	for _, w := range warnings {
		out = append(out, Warning{Code: w.Code, Message: w.Message, Hints: w.Hints, Fields: w.Fields})
	}
	return out
}

// renderSnapshot is the output block of docs/spec/cmd/snapshot.md#salida:
// one line for the two files, one for the revision when there was one, and
// one for the publication when it was asked for and happened.
func renderSnapshot(r *ops.SnapshotResult) string {
	out := fmt.Sprintf("Snapshot written: %s, %s (%s)\n",
		r.Files[0], r.Files[1], plural(r.Tasks, "task"))
	if !r.Committed {
		return out
	}
	// The identifier is abbreviated in the text and whole in the JSON, and
	// a system that returns none leaves the line without one.
	where := "the board's own repository"
	if !r.OwnRepository {
		where = r.Repository + ", the repository this board lives in"
	}
	if r.Commit == "" {
		out += "Committed to " + where + "\n"
	} else {
		out += fmt.Sprintf("Committed %s to %s\n", abbreviate(r.Commit), where)
	}
	if r.Pushed {
		out += "Pushed that branch to its remote\n"
	}
	return out
}

// abbreviate is the short form of a revision identifier that the text
// output prints, seven characters like the one of the example.
func abbreviate(commit string) string {
	if len(commit) <= 7 {
		return commit
	}
	return commit[:7]
}

// plural writes a count with its noun, in the singular when it is one.
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// The `snapshot` envelope of docs/spec/cmd/snapshot.md#el-esquema-json.
type snapshotEnvelopeData struct {
	Tasks              int      `json:"tasks"`
	Files              []string `json:"files"`
	VCS                string   `json:"vcs"`
	Committed          bool     `json:"committed"`
	Commit             *string  `json:"commit"`
	Repository         *string  `json:"repository"`
	Pushed             bool     `json:"pushed"`
	VCSOutput          []string `json:"vcsOutput"`
	StagedOutsideBoard int      `json:"stagedOutsideBoard"`
	Skipped            []string `json:"skipped"`
}

func snapshotData(r *ops.SnapshotResult) snapshotEnvelopeData {
	data := snapshotEnvelopeData{
		Tasks:              r.Tasks,
		Files:              r.Files,
		VCS:                r.VCS,
		Committed:          r.Committed,
		Pushed:             r.Pushed,
		VCSOutput:          list(r.VCSOutput),
		StagedOutsideBoard: r.StagedOutsideBoard,
		Skipped:            list(r.Skipped),
	}
	// commit and repository are null when no revision was recorded, and
	// commit also with a system that returns no identifier.
	if r.Commit != "" {
		commit := r.Commit
		data.Commit = &commit
	}
	if r.Committed && r.Repository != "" {
		repository := r.Repository
		data.Repository = &repository
	}
	return data
}
