package cli

// This file is the specification table of the commands that exist today,
// one row per flag, taken from each one's table of parameters. The generic
// parser of parse.go walks it; nothing here says what a command does.
//
// The field flags of docs/spec/familias-de-flags.md are not repeated per
// command: they live once in fields.go and every writing command takes the
// same table, which is what makes that page's promise true that a flag means
// the same wherever it appears.

// Version is the version this program answers with, and the one the
// top-level help prints on its first line.
const Version = "1.0.0"

// Commands returns the table, fresh on every call so that nobody can change
// the one another caller is reading.
func Commands() []CommandSpec {
	return []CommandSpec{initCommand(), whereCommand(), newCommand(), setCommand()}
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

// newCommand is the table of docs/spec/cmd/new.md: the field flags, plus
// --start. The batch of --from belongs to a later step and is not here yet
// (docs/spec/estado-de-implementacion.md).
func newCommand() CommandSpec {
	flags := fieldFlags()
	for i := range flags {
		if flags[i].Name == "status" {
			// The two say different things about the status of a task
			// being created, so writing both is a usage error. The pair
			// is declared on both sides because the parser judges it when
			// it reads the second of the two.
			flags[i].Conflicts = append(flags[i].Conflicts, "start")
		}
	}
	flags = append(flags, FlagSpec{Name: "start", Conflicts: []string{"status"}})
	return CommandSpec{Name: "new", Flags: flags}
}

// setCommand is the table of docs/spec/cmd/set.md: the field flags, plus the
// two that force how a <ref> is read.
func setCommand() CommandSpec {
	flags := append(fieldFlags(),
		FlagSpec{Name: "id", Conflicts: []string{"match"}},
		FlagSpec{Name: "match", Conflicts: []string{"id"}},
	)
	return CommandSpec{Name: "set", Flags: flags}
}
