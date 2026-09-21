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
		listCommand(), getCommand(), setCommand(), primeCommand(),
		startCommand(), noteCommand(), commentCommand(),
		finishCommand(), askCommand(), answerCommand(),
		archiveCommand(), configCommand(), doctorCommand(), helpCommand(),
		exportCommand(), snapshotCommand(),
	}
}

// vocabularyFlagNames are the flags of `biso init` that describe the board's
// vocabulary. They are the ones --from is incompatible with, because with
// --from every one of them comes from board.json instead
// (docs/spec/cmd/init.md).
var vocabularyFlagNames = []string{
	"statuses", "initial-status", "active-status", "terminal-status",
	"types", "priorities", "prefix", "overwrite-config",
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
			{Name: "prefix", Value: PlainValue, Conflicts: conflictsWithFrom},
			{Name: "overwrite-config", Conflicts: conflictsWithFrom},
			{Name: "from", Value: PlainValue, Conflicts: vocabularyFlagNames},
		},
	}
}

// archiveCommand is the table of docs/spec/cmd/archive.md: the field flags,
// the pair that forces how a <ref> is read, and its own switch.
func archiveCommand() CommandSpec {
	flags := append(fieldFlags(), resolution()...)
	flags = append(flags, FlagSpec{Name: "unarchive"})
	return CommandSpec{Name: "archive", Flags: flags}
}

// configCommand is the table of docs/spec/cmd/config.md. Its three
// subcommands are positional arguments and not flags, so the table carries
// none of its own; --json is global and this command judges where it
// applies itself, because only `config list` takes it.
//
// AffectsNoTask is true because `config set` writes without touching any
// task that existed before, which is what makes --print a usage error
// (docs/spec/cmd/flags-globales.md). ReadOnly is false for the same reason
// it is false in `init`: --dry-run does apply, and `config get` and
// `config list` reject it themselves, where the subcommand is known.
func configCommand() CommandSpec {
	return CommandSpec{Name: "config", AffectsNoTask: true}
}

// doctorCommand is the table of docs/spec/cmd/doctor.md: one switch, which
// is also the flag that turns a read-only command into a writing one.
//
// AffectsNoTask is true in both halves of that: what --fix repairs is a
// lease that no task justifies, a counter and a marker file, and none of
// them is a task to print a record of, so --print has nothing to do here
// with --fix or without it and is a usage error either way
// (docs/spec/cmd/flags-globales.md#flags-globales).
func doctorCommand() CommandSpec {
	return CommandSpec{
		Name:          "doctor",
		ReadOnly:      true,
		WriteFlags:    []string{"fix"},
		AffectsNoTask: true,
		Flags:         []FlagSpec{{Name: "fix"}},
	}
}

// helpCommand is the table of docs/spec/cmd/help.md. Its arguments are
// command names, which are positional, so it declares no flag of its own.
func helpCommand() CommandSpec {
	return CommandSpec{Name: "help", ReadOnly: true}
}

func whereCommand() CommandSpec {
	return CommandSpec{Name: "where", ReadOnly: true}
}

// primeCommand is the table of docs/spec/cmd/prime.md. It does not declare
// --full incompatible with --json, although they are: that pair has a
// message of its own, plain text on stderr even with --json, and the
// generic rejection of the parser would print the wrong one.
func primeCommand() CommandSpec {
	return CommandSpec{
		Name:     "prime",
		ReadOnly: true,
		Flags: []FlagSpec{
			{Name: "full"},
			{Name: "limit", Value: PlainValue},
		},
	}
}

// newCommand is the table of docs/spec/cmd/new.md: the field flags, plus
// --start and the --from of the batch.
//
// --from names every one of the others and every one of the others names
// --from, because a batch takes no field flag at all: everything it writes
// travels in the file (docs/spec/cmd/new.md#el-modo-lote). The pair is
// declared on both sides because the parser judges it when it reads the
// second of the two, so declaring it once would accept the call written the
// other way round.
func newCommand() CommandSpec {
	flags := fieldFlags()
	names := make([]string, 0, len(flags)+1)
	for i := range flags {
		if flags[i].Name == "status" {
			// The two say different things about the status of a task
			// being created, so writing both is a usage error.
			flags[i].Conflicts = append(flags[i].Conflicts, "start")
		}
		flags[i].Conflicts = append(flags[i].Conflicts, "from")
		names = append(names, flags[i].Name)
	}
	names = append(names, "start")
	flags = append(flags,
		FlagSpec{Name: "start", Conflicts: []string{"status", "from"}},
		FlagSpec{Name: "from", Value: PlainValue, Conflicts: names},
	)
	return CommandSpec{Name: "new", Flags: flags}
}

// exportCommand is the table of docs/spec/cmd/export.md: every filter of
// `biso ls` and none of its shaping flags, because a dump has no shape to
// choose, plus the two of its own.
//
// --archived and --only-archived are not here either, and their absence is
// not an oversight: archived tasks come out by default, so the only flag
// this command has about the archive is --no-archived.
func exportCommand() CommandSpec {
	flags := append(filterFlags(),
		FlagSpec{Name: "out", Value: PlainValue},
		FlagSpec{Name: "no-archived"},
	)
	return CommandSpec{Name: "export", ReadOnly: true, AffectsNoTask: true, Flags: flags}
}

// snapshotCommand is the table of docs/spec/cmd/snapshot.md. The value of
// --vcs is not closed here but in internal/vcs, which is the one place that
// knows the three modes and the message that rejects a fourth.
func snapshotCommand() CommandSpec {
	return CommandSpec{
		Name: "snapshot", ReadOnly: true, AffectsNoTask: true,
		Flags: []FlagSpec{{Name: "vcs", Value: PlainValue}},
	}
}

// setCommand is the table of docs/spec/cmd/set.md: the field flags, plus the
// two that force how a <ref> is read.
func setCommand() CommandSpec {
	return CommandSpec{Name: "set", Flags: append(fieldFlags(), resolution()...)}
}

// The six verbs of the cycle, each one the field flags of fields.go plus its
// own row or two (docs/spec/cmd/verbos-del-ciclo.md). They take all of them
// and not a subset, which is the promise that page opens with, so the table
// is built from the same function `biso set` builds its own from.

// resolution is the pair that forces how a <ref> is read, which every
// command with a positional reference carries
// (docs/spec/referencias.md#la-gramática).
func resolution() []FlagSpec {
	return []FlagSpec{
		{Name: "id", Conflicts: []string{"match"}},
		{Name: "match", Conflicts: []string{"id"}},
	}
}

func startCommand() CommandSpec {
	flags := append(fieldFlags(), resolution()...)
	flags = append(flags, FlagSpec{Name: "reopen"})
	return CommandSpec{Name: "start", Flags: flags}
}

// noteCommand and the three verbs below take exactly one reference and read
// every positional after it as a long text, which is what makes
// `biso note MYP-11 @findings.md` work.
func noteCommand() CommandSpec {
	return CommandSpec{
		Name:                "note",
		Flags:               append(fieldFlags(), resolution()...),
		TextPositionalsFrom: 1,
	}
}

// commentCommand is the one command where --comment-author stands on its
// own: everywhere else it is the author of a --comment, and here the
// positional text is the comment (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
func commentCommand() CommandSpec {
	flags := append(fieldFlags(), resolution()...)
	for i := range flags {
		if flags[i].Name == "comment-author" {
			flags[i].Requires = nil
		}
	}
	return CommandSpec{Name: "comment", Flags: flags, TextPositionalsFrom: 1}
}

func finishCommand() CommandSpec {
	flags := append(fieldFlags(), resolution()...)
	flags = append(flags,
		FlagSpec{Name: "strict", Conflicts: []string{"no-checks"}},
		FlagSpec{Name: "no-checks", Conflicts: []string{"strict"}},
	)
	return CommandSpec{Name: "finish", Flags: flags}
}

// askCommand and answerCommand drop --comment-author from the table: the
// author of a question is always the configured identity, and of the two
// comments `biso answer` writes only the answer is signed by the caller
// (docs/spec/cmd/verbos-del-ciclo.md#biso-ask).
func askCommand() CommandSpec {
	return CommandSpec{
		Name:                "ask",
		Flags:               withoutCommentAuthor(append(fieldFlags(), resolution()...)),
		TextPositionalsFrom: 1,
	}
}

func answerCommand() CommandSpec {
	return CommandSpec{
		Name:                "answer",
		Flags:               withoutCommentAuthor(append(fieldFlags(), resolution()...)),
		TextPositionalsFrom: 1,
	}
}

func withoutCommentAuthor(flags []FlagSpec) []FlagSpec {
	out := make([]FlagSpec, 0, len(flags))
	for _, f := range flags {
		if f.Name == "comment-author" {
			continue
		}
		out = append(out, f)
	}
	return out
}

// listCommand is the table of docs/spec/cmd/ls.md: the filters, with the
// incompatibilities its table of parameters declares on both sides of every
// pair, and the flags that shape the answer.
//
// None of the filters is a field flag: `biso ls` writes nothing, so every
// one of them is a switch or a value that belongs to this command alone.
func listCommand() CommandSpec {
	flags := append(filterFlags(),
		FlagSpec{Name: "archived", Conflicts: []string{"only-archived"}},
		FlagSpec{Name: "only-archived", Conflicts: []string{"archived"}},
	)
	flags = append(flags,
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

// filterFlags are the filters of docs/spec/cmd/ls.md, the ones that narrow
// a listing instead of shaping it. They live apart from listCommand because
// `biso export` takes exactly these and none of the others
// (docs/spec/cmd/export.md), so writing them twice is what would let the
// two lists drift apart.
//
// --archived and --only-archived are not among them: they are the two that
// `export` does not take, so each command adds its own.
func filterFlags() []FlagSpec {
	// A filter that repeats accumulates values and admits a list separated
	// by commas, which is what the head of its table says
	// ("repeat or comma-separate").
	filter := func(name string, conflicts ...string) FlagSpec {
		return FlagSpec{
			Name: name, Value: PlainValue,
			Repeatable: true, Comma: true, Conflicts: conflicts,
		}
	}
	// A filter over a configured vocabulary declares it, so that an empty
	// value travels down to the matching algorithm instead of being
	// rejected here: the empty value is an unknown value of that
	// vocabulary, exit code 3, in both directions
	// (docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos).
	vocabulary := func(name string, conflicts ...string) FlagSpec {
		f := filter(name, conflicts...)
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
		vocabulary("status", "any-status"),
		vocabulary("not-status", "any-status"),
		{Name: "any-status", Conflicts: []string{"status", "not-status"}},
		vocabulary("type"),
		vocabulary("priority"),
		filter("label"),
		filter("label-or"),
		filter("assignee", "mine", "unassigned"),
		{Name: "mine", Conflicts: []string{"assignee", "unassigned"}},
		{Name: "unassigned", Conflicts: []string{"assignee", "mine"}},
		{Name: "parent", Value: PlainValue},
	}
	flags = append(flags, opposites("blocked", "not-blocked")...)
	flags = append(flags, opposites("waiting", "not-waiting")...)
	flags = append(flags, opposites("active", "not-active")...)
	return append(flags,
		FlagSpec{Name: "overdue"},
		FlagSpec{Name: "due-before", Value: PlainValue, Field: "dueBefore"},
		FlagSpec{Name: "search", Value: TextValue},
		FlagSpec{Name: "unchecked"},
	)
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
