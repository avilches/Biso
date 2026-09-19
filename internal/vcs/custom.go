package vcs

import (
	"strings"

	"biso/internal/model"
)

// Placeholders of the orders of a custom system, from the table of
// docs/spec/invocacion.md.
const (
	// placeholderMessage is replaced, wherever it appears inside an argument,
	// by the revision message biso composes.
	placeholderMessage = "{message}"
	// placeholderFiles is replaced, only when it is a whole argument, by the
	// three files of the revision, each one as an argument of its own.
	placeholderFiles = "{files}"
)

// runCustom executes the orders of a system biso does not know. It asks the
// repository nothing: it runs the order in the board's directory and looks at
// its exit code, and those are the only two things it can say afterwards, so a
// snapshot with custom never carries a revision identifier and never tells the
// case of having nothing to record.
func (r *Runner) runCustom(req Request) (*Result, *model.Error) {
	result := &Result{VCS: string(KindCustom)}
	var lines collector

	args := expand(r.config.Custom.Commit, req)
	if len(args) == 0 {
		return result, commitFailed(nil)
	}
	committed := order(&lines, req.BoardDir, args[0], args[1:]...)
	result.Output = lines.take()
	switch {
	case !committed.started:
		// The configured system is not installed, which is never a failure of
		// this command: the two files are written and the revision is skipped.
		result.Outcome = OutcomeUnavailable
		return result, nil
	case committed.exit != 0:
		return result, commitFailed(result.Output)
	}
	result.Outcome = OutcomeCommitted

	if r.mode == ModePush {
		// New already rejected a push with no publish order declared, so
		// reaching here with one means the order exists and is expected to
		// work: a program that cannot even be launched is a failed push, not a
		// system that is not installed, because the commit order did run.
		publish := expand(r.config.Custom.Publish, req)
		published := order(&lines, req.BoardDir, publish[0], publish[1:]...)
		result.Output = lines.take()
		if !published.ok() {
			return result, pushFailed(result.Output)
		}
		result.Pushed = true
	}
	return result, nil
}

// expand substitutes the two placeholders of a declared order.
func expand(args []string, req Request) []string {
	expanded := make([]string, 0, len(args)+len(req.Files))
	for _, arg := range args {
		if arg == placeholderFiles {
			expanded = append(expanded, req.Files...)
			continue
		}
		expanded = append(expanded, strings.ReplaceAll(arg, placeholderMessage, req.Message))
	}
	return expanded
}
