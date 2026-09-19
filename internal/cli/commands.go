package cli

// This file is the specification table of the two commands that exist
// today, one row per flag, taken from the table of parameters of
// docs/spec/cmd/init.md and docs/spec/cmd/where.md. The generic parser of
// parse.go walks it; nothing here says what a command does.

// Version is the version this program answers with, and the one the
// top-level help prints on its first line.
const Version = "1.0.0"

// Commands returns the table, fresh on every call so that nobody can change
// the one another caller is reading.
func Commands() []CommandSpec {
	return []CommandSpec{initCommand(), whereCommand()}
}

// vocabularyFlagNames are the flags of `biso init` that describe the board's
// vocabulary. They are the ones --from is incompatible with, because with
// --from every one of them comes from board.json instead
// (docs/spec/cmd/init.md).
var vocabularyFlagNames = []string{
	"statuses", "initial-status", "active-status", "terminal-status",
	"types", "priorities", "extensions", "prefix", "overwrite-config",
}

func initCommand() CommandSpec {
	// Every one of these names --from among its conflicts, and --from names
	// every one of them, because the parser judges a pair when it reads the
	// second of the two: declaring it on one side only would accept the
	// call written the other way round.
	conflictsWithFrom := []string{"from"}
	list := func(name string) FlagSpec {
		return FlagSpec{
			Name: name, Value: PlainValue, Repeatable: true, Comma: true,
			Conflicts: conflictsWithFrom,
		}
	}
	return CommandSpec{
		Name: "init",
		// init writes, so --dry-run applies, and it touches no task that
		// existed before, so --print does not
		// (docs/spec/cmd/flags-globales.md).
		AffectsNoTask: true,
		Flags: []FlagSpec{
			{Name: "at", Value: PlainValue},
			list("statuses"),
			{Name: "initial-status", Value: PlainValue, Conflicts: conflictsWithFrom},
			{Name: "active-status", Value: PlainValue, Conflicts: conflictsWithFrom},
			{Name: "terminal-status", Value: PlainValue, Conflicts: conflictsWithFrom},
			list("types"),
			list("priorities"),
			list("extensions"),
			{Name: "prefix", Value: PlainValue, Conflicts: conflictsWithFrom},
			{Name: "overwrite-config", Conflicts: conflictsWithFrom},
			{Name: "from", Value: PlainValue, Conflicts: vocabularyFlagNames},
		},
	}
}

func whereCommand() CommandSpec {
	return CommandSpec{Name: "where", ReadOnly: true}
}
