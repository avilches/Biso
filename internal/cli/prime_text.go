package cli

import "strings"

// This file holds the part of the startup message that does not depend on
// the board, transcribed character for character from
// docs/spec/cmd/prime.md#la-salida-literal, plus the block --full adds.
//
// It is fixed on purpose, examples included: the `MYP-12` of rule 2 and the
// `Done` of rule 4 illustrate the shape of an identifier and of a terminal
// status and do not describe this board, which is what
// docs/spec/cmd/prime.md#lo-que-no-depende-del-tablero says and what makes
// the arithmetic of docs/spec/presupuestos.md#el-presupuesto-de-tamaño
// exact: the fixed part always measures the same, so all the room left up
// to the cap belongs to the summary.

// primeCommands is the COMMANDS block: the ten orders of the cycle with
// their shape.
const primeCommands = "COMMANDS  (`biso help <cmd>...` for the detail of any, several at once)\n" +
	"  biso ls [--status STATUS] [--type T] [--label LABEL] [--mine] [--search TEXT]\n" +
	"  biso get <ref> [--section ac]\n" +
	"  biso new \"TITLE\" [--append-desc TEXT] [--add-ac TEXT]... [--type T] [--priority P]\n" +
	"  biso start <ref>... [--append-plan TEXT]\n" +
	"  biso note <ref> \"TEXT\"\n" +
	"  biso ask <ref> \"QUESTION\"\n" +
	"  biso answer <ref> \"TEXT\"\n" +
	"  biso finish <ref>... [--append-summary \"TEXT\"] [--check-ac all]\n" +
	"  biso set <ref>... [any field flag]\n" +
	"  biso comment <ref> \"TEXT\" [--comment-author @who]\n"

// primeFieldFlags is the FIELD FLAGS grid: the names of every field flag,
// in twelve lines. `--full` prints the same flags grouped by the field
// they write, and neither block replaces the other.
const primeFieldFlags = "FIELD FLAGS  (same names, same meaning, in every command above that writes)\n" +
	"  --title  --status  --type --clear-type  --priority --clear-priority\n" +
	"  --parent --clear-parent  --due --clear-due  --ordinal --clear-ordinal  --author --clear-author\n" +
	"  --add-labels --rm-labels --clear-labels --replace-labels\n" +
	"  --add-assignees --rm-assignees --clear-assignees --replace-assignees\n" +
	"  --add-refs --rm-refs --clear-refs --replace-refs\n" +
	"  --add-deps --rm-deps --clear-deps --replace-deps\n" +
	"  --add-files --rm-files --clear-files --replace-files\n" +
	"  --add-ac --rm-ac --clear-acs   --check-ac --uncheck-ac\n" +
	"  --append-desc --clear-desc  --append-plan --clear-plan\n" +
	"  --append-note --clear-notes  --append-summary --clear-summary\n" +
	"  --comment --rm-comment --set-comment-date\n" +
	"  --ext K=V --rm-ext --clear-ext\n"

// primeRules is the RULES block: the ten rules that cannot be guessed.
const primeRules = "RULES  (none of these are guessable; they are the whole learning curve)\n" +
	"  1. Every write goes through biso. Nothing else touches the board.\n" +
	"  2. <ref> is an id (MYP-12), a bare number (12) or free text (\"CRLF\"). Text\n" +
	"     matching several tasks is an error that lists them, never a guess. `note`,\n" +
	"     `comment`, `ask` and `answer` take one <ref>; `set`, `start` and `finish`\n" +
	"     take several.\n" +
	"  3. Filters reject values this board does not have: `--status Pending` is an error,\n" +
	"     not an empty list. Case, spaces, hyphens and underscores are ignored, so\n" +
	"     `--status todo`, `--status \"To Do\"` and `--status TO_DO` are one and the same filter. An\n" +
	"     empty list is therefore a fact about the board that you can act on.\n" +
	"  4. `biso ls` prints 30 tasks by urgency and leaves out the Done ones. It says\n" +
	"     on stderr what it left out. --all lifts the limit, --any-status includes\n" +
	"     Done, --archived reaches the archive.\n" +
	"  5. --check-ac and --uncheck-ac take all, 3, 1-4, 1,3,7 or the criterion text. The\n" +
	"     numbers are the stable #N keys that `biso get` shows, and they never shift\n" +
	"     when one criterion is removed.\n" +
	"  6. `biso new` prints the new id and nothing else. Every other write prints one\n" +
	"     line per task: id, status, criteria, urgency. Add --print for the whole\n" +
	"     record, or --json for a versioned envelope.\n" +
	"  7. Write `biso -C <dir> ...`, never `cd <dir> && biso ...`.\n" +
	"  8. Long text: a real newline works, and so do --append-desc @file.md and --append-desc - for stdin.\n" +
	"  9. Exit codes: 0 ok, 2 bad usage, 3 bad value, 4 not found, 5 ambiguous,\n" +
	"     6 precondition not met, 7 nothing written, 8 environment, 20 no board here.\n" +
	" 10. `biso ask <ref> \"...\"` parks a task on a question and `biso answer` unparks\n" +
	"     it, writing both into the comments. Ask instead of guessing. A task\n" +
	"     assigned to you is one a person decided you should do.\n"

// primeClosing is the paragraph that closes the message.
const primeClosing = "Pick one, `biso start <ref> --append-plan \"...\"`, work, `biso note <ref> \"...\"` as you go,\n" +
	"and close with `biso finish <ref> --check-ac all --append-summary \"...\"`.\n" +
	"That is the loop. Create a task when the work needs planning or review; do small\n" +
	"edits directly.\n"

// primeEmptyBoard replaces the four blocks of tasks on a board that holds
// none (docs/spec/cmd/prime.md#tablero-vacío).
const primeEmptyBoard = "THE BOARD IS EMPTY\n" +
	"  Create the first one:\n" +
	"  biso new \"Title\" --append-desc \"What and why\" --add-ac \"How we will know it works\"\n"

// primeTitle is the first line, and the one thing of the fixed part that
// changes: the version string is the one of whoever runs it, and the
// stability contract lets it change between versions.
func primeTitle(version string) string {
	return "biso " + version +
		" - the task board of this project. This message is all you need to start.\n"
}

// primeFullBlock is what --full adds at the end
// (docs/spec/cmd/prime.md#--full): every field flag grouped by the field it
// writes, in the order a write applies them.
//
// It is built from the one table of fields.go and not written out by hand,
// because a second copy of seventy flag names is exactly the thing that
// drifts: a flag added there appears here without anybody remembering to
// add it.
func primeFullBlock() string {
	var b strings.Builder
	b.WriteString("ALL FIELD FLAGS  (--full: by the field they write, in the order a write applies them)\n")

	groups, order := fieldFlagsByField()
	width := 0
	for _, field := range order {
		if len(field) > width {
			width = len(field)
		}
	}
	for _, field := range order {
		b.WriteString("  " + field + strings.Repeat(" ", width-len(field)) + "  ")
		b.WriteString(strings.Join(groups[field], " "))
		b.WriteString("\n")
	}
	b.WriteString("\n  `biso new` adds --start, and `biso set` adds --id and --match.\n")
	return b.String()
}

// fieldFlagsByField groups the field flags by the field of the JSON
// contract they write, keeping the order of the table, which is the order
// of docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura.
//
// Two entries of the table carry no field of their own and are placed by
// hand, because there is no other honest place for them: --check-ac and
// --uncheck-ac write the criteria without being one of the four shapes of
// a list, and --comment-author signs a comment instead of writing a field.
func fieldFlagsByField() (map[string][]string, []string) {
	groups := map[string][]string{}
	var order []string
	for _, f := range fieldFlags() {
		field := f.Field
		switch {
		case f.Category == CheckAC:
			field = "acceptanceCriteria"
		case field == "comment-author":
			field = "comments"
		}
		if _, seen := groups[field]; !seen {
			order = append(order, field)
		}
		groups[field] = append(groups[field], "--"+f.Name)
	}
	return groups, order
}
