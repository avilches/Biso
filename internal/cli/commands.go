package cli

import "biso/internal/ops"

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
	return []CommandSpec{
		initCommand(), whereCommand(), newCommand(),
		listCommand(), getCommand(), setCommand(),
	}
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

// listCommand is the table of docs/spec/cmd/ls.md: the filters, with the
// incompatibilities its table of parameters declares on both sides of every
// pair, and the flags that shape the answer.
//
// None of the filters is a field flag: `biso ls` writes nothing, so every
// one of them is a switch or a value that belongs to this command alone.
func listCommand() CommandSpec {
	// A filter that repeats accumulates values and admits a list separated
	// by commas, which is what the head of its table says
	// ("repeat or comma-separate").
	filter := func(name, short string, conflicts ...string) FlagSpec {
		return FlagSpec{
			Name: name, Short: short, Value: PlainValue,
			Repeatable: true, Comma: true, Conflicts: conflicts,
		}
	}
	// A filter over a configured vocabulary declares it, so that an empty
	// value travels down to the matching algorithm instead of being
	// rejected here: the empty value is an unknown value of that
	// vocabulary, exit code 3, in both directions
	// (docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos).
	vocabulary := func(name, short string, conflicts ...string) FlagSpec {
		f := filter(name, short, conflicts...)
		f.ClosedVocabulary = true
		return f
	}
	// The two halves of a pair of opposites each name the other, because
	// the parser judges a pair when it reads the second of the two.
	opposites := func(yes, no string) []FlagSpec {
		return []FlagSpec{
			{Name: yes, Conflicts: []string{no}},
			{Name: no, Conflicts: []string{yes}},
		}
	}
	flags := []FlagSpec{
		vocabulary("status", "s", "any-status"),
		vocabulary("not-status", "", "any-status"),
		{Name: "any-status", Conflicts: []string{"status", "not-status"}},
		{Name: "archived", Conflicts: []string{"only-archived"}},
		{Name: "only-archived", Conflicts: []string{"archived"}},
		vocabulary("type", ""),
		vocabulary("priority", ""),
		filter("label", "l"),
		filter("label-or", ""),
		filter("assignee", "a", "mine", "unassigned"),
		{Name: "mine", Conflicts: []string{"assignee", "unassigned"}},
		{Name: "unassigned", Conflicts: []string{"assignee", "mine"}},
		{Name: "parent", Short: "p", Value: PlainValue},
	}
	flags = append(flags, opposites("blocked", "not-blocked")...)
	flags = append(flags, opposites("waiting", "not-waiting")...)
	flags = append(flags, opposites("active", "not-active")...)
	flags = append(flags,
		FlagSpec{Name: "overdue"},
		FlagSpec{Name: "due-before", Value: PlainValue, Field: "dueBefore"},
		FlagSpec{Name: "search", Value: TextValue},
		FlagSpec{Name: "unchecked"},
		FlagSpec{
			Name: "sort", Value: PlainValue,
			Domain: sortFields, DomainCode: "unknown_sort_field",
		},
		FlagSpec{Name: "reverse"},
		FlagSpec{Name: "limit", Value: PlainValue, Conflicts: []string{"all"}},
		FlagSpec{Name: "all", Conflicts: []string{"limit"}},
		FlagSpec{Name: "ids", Conflicts: []string{"count"}},
		FlagSpec{Name: "count", Conflicts: []string{"ids"}},
	)
	return CommandSpec{Name: "ls", ReadOnly: true, Flags: flags}
}

// getCommand is the table of docs/spec/cmd/get.md: the two flags that force
// how a <ref> is read, the sections and the explanation of the urgency.
func getCommand() CommandSpec {
	return CommandSpec{
		Name:     "get",
		ReadOnly: true,
		Flags: []FlagSpec{
			{Name: "id", Conflicts: []string{"match"}},
			{Name: "match", Conflicts: []string{"id"}},
			{
				Name: "section", Value: PlainValue, Repeatable: true, Comma: true,
				Domain: sections, DomainCode: "unknown_section",
			},
			{Name: "explain-urgency"},
		},
	}
}

// sortFields and sections are the two closed domains of these commands,
// taken from internal/ops so that the table the parser checks and the one
// the logic answers with are the same list.
var (
	sortFields = ops.SortFields
	sections   = ops.Sections
)
