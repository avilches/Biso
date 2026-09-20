package cli

// This file holds the help text of the six verbs of the cycle, transcribed
// character for character from docs/spec/cmd/verbos-del-ciclo.md, exactly as
// help.go holds the ones of the commands of the steps before this one. The
// golden test of cmd/biso reads the same blocks out of that page and compares
// them with what the compiled program prints.

// startHelp is `biso start --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-start).
const startHelp = "Usage: biso start <ref>... [options]\n" +
	"\n" +
	"Take one or more tasks: move them to the active status, claim the lease for\n" +
	"you, assign them to you if nobody has them, and record a plan. One call.\n" +
	"\n" +
	"Options:\n" +
	"      --append-plan <text>       add to the implementation plan; repeatable,\n" +
	"                                 and takes @file and - like every text option\n" +
	"  -a, --add-assignees <@who>     add an assignee (--replace-assignees replaces\n" +
	"                                 the list)\n" +
	"  -s, --status <value>           use another status instead of the active one;\n" +
	"                                 no lease is claimed then, a lease only exists\n" +
	"                                 on an active task\n" +
	"      --reopen           allow starting a task that is already finished\n" +
	"      --id / --match     force <ref> to be an id, or free text\n" +
	"  -h, --help             show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too.\n" +
	"\n" +
	"Unresolved dependencies produce a warning, not an error: you decide. Taking\n" +
	"over a live lease held by someone else is the same: it warns, it does not\n" +
	"refuse. An archived task refuses instead: `biso archive --unarchive` it first.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  started        4  not found        8  the board could not be written\n" +
	"  2  bad usage      5  ambiguous        20 no board here\n" +
	"  3  unknown value  6  already finished, or archived\n" +
	"\n" +
	"Examples:\n" +
	"  biso start MYP-11 --append-plan \"1. Read the parser. 2. Add the CRLF case.\"\n" +
	"  biso start 11\n" +
	"  biso start MYP-11 MYP-12\n"

// noteHelp is `biso note --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-note).
const noteHelp = "Usage: biso note <ref> <text>... [options]\n" +
	"\n" +
	"Append one or more paragraphs to the implementation notes of ONE task. It never\n" +
	"replaces anything; clearing and appending in the same `biso set <ref>` call does.\n" +
	"\n" +
	"Arguments:\n" +
	"  ref            one task: an id, a bare number or free text\n" +
	"  text           one paragraph per argument; @file and - work here too\n" +
	"\n" +
	"Options:\n" +
	"      --id / --match   force <ref> to be an id, or free text\n" +
	"  -h, --help           show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too. Use `--append-note <text>`\n" +
	"for a paragraph that is not checked against the id grammar, for when the note\n" +
	"itself looks like an id.\n" +
	"\n" +
	"To note the same thing on several tasks, use `biso set A B --append-note \"...\"`.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  appended       3  the task could not be read      8  could not be written\n" +
	"  2  bad usage      4  not found                      20  no board here\n" +
	"                    5  ambiguous\n" +
	"\n" +
	"Examples:\n" +
	"  biso note MYP-11 \"The parser already normalized LF, CRLF was missing\"\n" +
	"  biso note 11 \"First finding\" \"Second finding\"\n" +
	"  biso note MYP-11 @/tmp/benchmark-output.txt\n"

// commentHelp is `biso comment --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
const commentHelp = "Usage: biso comment <ref> <text>... [options]\n" +
	"\n" +
	"Append a discussion comment to ONE task, with an author and a timestamp. This\n" +
	"is not `biso note`, which records what you did while implementing.\n" +
	"\n" +
	"Arguments:\n" +
	"  ref                      one task: an id, a bare number or free text\n" +
	"  text                     one comment per argument; @file and - work too\n" +
	"\n" +
	"Options:\n" +
	"      --comment-author <@who>  free text author (default: you). An external\n" +
	"                               system can use its own convention, such as\n" +
	"                               @trello:juan. The @ is never a file reference\n" +
	"      --id / --match           force <ref> to be an id, or free text\n" +
	"  -h, --help                   show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too. Use `--comment <text>`\n" +
	"for a comment that is not checked against the id grammar.\n" +
	"\n" +
	"A comment's body and author are never edited, by any flag. The whole comment\n" +
	"can be removed with --rm-comment, and only its date corrected with\n" +
	"--set-comment-date, both in `biso set --help`.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  appended       3  the task could not be read      8  could not be written\n" +
	"  2  bad usage      4  not found                      20  no board here\n" +
	"                    5  ambiguous\n" +
	"\n" +
	"Examples:\n" +
	"  biso comment MYP-11 \"A user with a Windows clone reported this\"\n" +
	"  biso comment 11 \"Moved to Doing from the phone\" --comment-author @trello:avilches\n"

// finishHelp is `biso finish --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
const finishHelp = "Usage: biso finish <ref>... [options]\n" +
	"\n" +
	"Close one or more tasks: check criteria, add the last note, write the final\n" +
	"summary and move to the terminal status. One call.\n" +
	"\n" +
	"Options:\n" +
	"      --append-summary <text>  add to the final summary; repeatable, takes\n" +
	"                               @file and -\n" +
	"      --check-ac <sel>         check criteria: all, 3, 1-4, 1,3,7 or the text.\n" +
	"                               With several tasks the selector has to be `all`\n" +
	"      --append-note <text>     one last implementation note; repeatable\n" +
	"      --add-files <path>       record a modified file; repeatable\n" +
	"  -s, --status <value>         use another status instead of the terminal one\n" +
	"      --strict           refuse to finish with unchecked criteria, unfinished\n" +
	"                         subtasks or no summary\n" +
	"                         (default: warn and go on; see finish_strict)\n" +
	"      --no-checks        skip every check and every warning\n" +
	"      --id / --match     force <ref> to be an id, or free text\n" +
	"  -h, --help             show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  finished       3  unknown value    6  --strict and something is missing\n" +
	"  2  bad usage      4  not found        8  the board could not be written\n" +
	"                    5  ambiguous        20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso finish MYP-11 --check-ac all --append-summary \"Normalizes CRLF\"\n" +
	"  biso finish MYP-11 --check-ac \"covers CRLF\" --append-note \"313 tests green\"\n" +
	"  biso finish MYP-11 MYP-12 --check-ac all --append-summary \"Both closed by PR 42\"\n"

// askHelp is `biso ask --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-ask).
const askHelp = "Usage: biso ask <ref> <text>... [options]\n" +
	"\n" +
	"Park ONE task on a question for a person. The task keeps its status, but it\n" +
	"leaves the IN PROGRESS block of `biso prime` and shows up under NEEDS ANSWER\n" +
	"until somebody runs `biso answer`.\n" +
	"\n" +
	"Arguments:\n" +
	"  ref                one task: an id, a bare number or free text\n" +
	"  text               the question; @file and - work too\n" +
	"\n" +
	"Options:\n" +
	"      --id / --match force <ref> to be an id, or free text\n" +
	"  -h, --help         show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too, but nothing except this\n" +
	"command and `biso answer` ever writes the question itself.\n" +
	"\n" +
	"A task holds one open question at a time. Answer it before asking another.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  asked          4  not found        8  could not be written\n" +
	"  2  bad usage      5  ambiguous        20 no board here\n" +
	"  3  empty question, or unreadable\n" +
	"  6  already asking, or already finished\n" +
	"\n" +
	"Examples:\n" +
	"  biso ask MYP-11 \"Do we normalize binary files too, or only text?\"\n" +
	"  biso ask 11 @/tmp/question.md\n"

// answerHelp is `biso answer --help` (docs/spec/cmd/verbos-del-ciclo.md#biso-answer).
const answerHelp = "Usage: biso answer <ref> <text>... [options]\n" +
	"\n" +
	"Answer the open question of ONE task and unpark it. In a single write this\n" +
	"moves the question into the comments with its original author and time, adds\n" +
	"your answer behind it, and clears the question.\n" +
	"\n" +
	"Arguments:\n" +
	"  ref                one task: an id, a bare number or free text\n" +
	"  text               the answer; @file and - work too\n" +
	"\n" +
	"Options:\n" +
	"      --id / --match force <ref> to be an id, or free text\n" +
	"  -h, --help         show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too, so you can answer and\n" +
	"refine in one call.\n" +
	"\n" +
	"Only the answer comment is signed with your configured identity; the question\n" +
	"comment keeps its original author and time. This command does not take\n" +
	"--comment-author.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  answered       4  not found        8  could not be written\n" +
	"  2  bad usage      5  ambiguous        20 no board here\n" +
	"  3  empty answer, or unreadable\n" +
	"  6  no open question\n" +
	"\n" +
	"Examples:\n" +
	"  biso answer MYP-11 \"Only text files. Binary ones are skipped entirely.\"\n" +
	"  biso answer 11 \"Yes\" --add-ac \"A binary file is never touched\"\n"
