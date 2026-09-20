package cli

import "biso/internal/ops"

// runArchive is `biso archive` (docs/spec/cmd/archive.md). It prints what
// every other writing command over an existing task prints, the status line
// of docs/spec/cmd/set.md#salida, which is why there is nothing of its own
// here beyond its two flags.
func runArchive(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		return ops.Archive(env, ops.ArchiveParams{
			Refs: p.Positionals, Mode: refMode(p),
			Unarchive: p.Has("unarchive"),
			Changes:   changesOf(p), DryRun: p.Has("dry-run"), Print: p.Has("print"),
		})
	})
}
