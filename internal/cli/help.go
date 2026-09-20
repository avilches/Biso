package cli

// This file holds the help text of each command, transcribed character for
// character from its page under docs/spec/cmd/. It is not derived from the
// specification table of spec.go on purpose: generating it would add a layer
// that would have to be kept in step by hand anyway, per section 6 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md.
//
// A test reads the same blocks out of docs/spec/ and compares them with
// these constants, so the two can never drift apart in silence.

// topLevelHelp is what `biso --help` and `biso help` print, and only that
// (docs/spec/cmd/help.md#la-ayuda-de-primer-nivel).
const topLevelHelp = "biso 1.0.0 - the task board of this project.\n" +
	"\n" +
	"Usage: biso [global options] <command> [options]\n" +
	"\n" +
	"Start here:\n" +
	"  prime              everything you need to work on this board, in one message\n" +
	"\n" +
	"Daily work:\n" +
	"  ls                 list tasks, most urgent first\n" +
	"  get <ref>          show one task\n" +
	"  new \"TITLE\"        create a task and print its id\n" +
	"  set <ref>...       change any field\n" +
	"  start <ref>...     take a task\n" +
	"  note <ref> TEXT    append an implementation note\n" +
	"  comment <ref> TEXT append a discussion comment\n" +
	"  finish <ref>...    close a task\n" +
	"  ask <ref> TEXT     park on a question\n" +
	"  answer <ref> TEXT  answer it and unpark\n" +
	"\n" +
	"Global options:\n" +
	"  -C, --cwd <path>   resolve the board from there, instead of cd-ing\n" +
	"      --json         machine-readable output\n" +
	"  -q, --quiet        print only ids\n" +
	"      --print        print the whole record after writing\n" +
	"      --color <when> auto (default), always, never\n" +
	"      --dry-run      validate, write nothing (writing commands only)\n" +
	"  -V, --version      print the version\n" +
	"  -h, --help         this, or the help of a command\n" +
	"\n" +
	"More: `biso <command> --help`, and `biso help all` for the administrative\n" +
	"commands (init, where, archive, export, config, doctor, board, help, snapshot).\n"

// initHelp is `biso init --help` (docs/spec/cmd/init.md).
const initHelp = "Usage: biso init [name] [options]\n" +
	"\n" +
	"Create a task board for this project. It writes the board and a pointer\n" +
	"inside the project so every copy of the project finds the same board. It\n" +
	"never writes outside the board otherwise.\n" +
	"\n" +
	"Arguments:\n" +
	"  name                   board name (default: the project directory name)\n" +
	"\n" +
	"Options:\n" +
	"  --at <dir>                  the board's own directory, not where to put it\n" +
	"                              (default: a new folder in the machine's default\n" +
	"                              boards root). A relative path is stored relative\n" +
	"                              to the pointer; an absolute one is stored as is\n" +
	"  --statuses <list>           comma-separated, at least three\n" +
	"                              (default: \"To Do,In Progress,Done\")\n" +
	"  --initial-status <status>   status of a new task (default: \"To Do\")\n" +
	"  --active-status <status>    what `biso start` sets (default: \"In Progress\")\n" +
	"  --terminal-status <status>  what `biso finish` sets (default: \"Done\")\n" +
	"  --types <list>              comma-separated (default: \"task,bug,docs\")\n" +
	"  --priorities <list>         comma-separated (default: \"high,medium,low\")\n" +
	"  --extensions <list>         comma-separated declared external field keys,\n" +
	"                              such as trello.card (default: none)\n" +
	"  --prefix <text>             task id prefix, letters only (default: derived\n" +
	"                              from the board name, uppercased)\n" +
	"  --overwrite-config          replace the configuration of an existing board,\n" +
	"                              keeping every task\n" +
	"  --from <location>           restore a snapshot: the directory where `biso\n" +
	"                              snapshot` wrote snapshot.ndjson and board.json.\n" +
	"                              Incompatible with name and with every vocabulary\n" +
	"                              option (which all come from board.json\n" +
	"                              instead), and with --overwrite-config: restoring\n" +
	"                              always creates a new board\n" +
	"  -h, --help                  show this help\n" +
	"\n" +
	"`--initial-status`, `--active-status` and `--terminal-status` each name one of\n" +
	"`--statuses`, all three distinct. Giving `--statuses` requires the three\n" +
	"together; giving any of them without `--statuses` is bad usage. They are then\n" +
	"stored as explicit values and never move again.\n" +
	"\n" +
	"With --at the board can live inside the project itself, which is fine. Two\n" +
	"things follow, and `init` says both when it applies. Decide whether the\n" +
	"project ignores that folder or versions it: ignore it and the board keeps a\n" +
	"history of its own, version it and its snapshot travels with your code. The\n" +
	"database stays out either way. And mind the form of the path: a relative --at\n" +
	"is stored relative to the pointer and resolves from any working copy that has\n" +
	"the folder inside it or above it, which is the case for worktrees kept under\n" +
	"the project; a working copy that lives outside the project has no such\n" +
	"folder, precisely because it is ignored, so pass an absolute --at if you work\n" +
	"that way.\n" +
	"\n" +
	"The board directory can also become a repository of its own, but `init` does\n" +
	"not create it: `init` only writes the ignore file of the version control\n" +
	"system this machine is configured for (the vcs key, .gitignore with git),\n" +
	"holding the database file and its WAL auxiliaries, so that once a repository\n" +
	"exists only snapshot.ndjson, board.json and the <id>.id marker are ever\n" +
	"versioned. `biso snapshot` is the one that creates that repository, lazily,\n" +
	"the first time it runs against a directory that is in none (see `biso\n" +
	"snapshot --help`). Missing version control never fails `init` or `snapshot`;\n" +
	"the board works the same, only its history is lost.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  board created, or restored with --from\n" +
	"  2  bad usage, or a board is already reachable from here\n" +
	"  4  --from points at a directory missing snapshot.ndjson, board.json, or both\n" +
	"  6  --overwrite-config would change task_prefix on a board with tasks\n" +
	"  7  --from's snapshot.ndjson failed batch validation\n" +
	"  8  cannot write there\n" +
	"\n" +
	"Examples:\n" +
	"  biso init\n" +
	"  biso init \"My project\" --statuses \"Ideas,To Do,In Progress,Done\" \\\n" +
	"      --initial-status Ideas --active-status \"In Progress\" \\\n" +
	"      --terminal-status Done\n" +
	"  biso init \"My project\" --prefix MYP --at my-project-board --extensions trello.card\n" +
	"  biso init --at /tmp/tablero-nuevo --from ~/.biso/boards/my-project-3f9a2b1c\n"

// whereHelp is `biso where --help` (docs/spec/cmd/where.md).
const whereHelp = "Usage: biso where [options]\n" +
	"\n" +
	"Say which board is in use and which rule picked it. Run it when a command\n" +
	"answers \"no board here\" and you expected one.\n" +
	"\n" +
	"Options:\n" +
	"      --json     machine-readable envelope\n" +
	"  -h, --help     show this help\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  a board is in use\n" +
	"  20 no board here\n" +
	"  21 its database could not be read\n" +
	"  22 the same board id is in two places\n" +
	"\n" +
	"Examples:\n" +
	"  biso where\n" +
	"  biso -C ~/work/my-project where\n"

// newHelp is `biso new --help` (docs/spec/cmd/new.md).
const newHelp = "Usage: biso new <title> [options]\n" +
	"       biso new --from <file|-> [options]\n" +
	"\n" +
	"Create a task and print its id. Every field flag of `biso set` works here.\n" +
	"\n" +
	"Arguments:\n" +
	"  title                      task title (required unless --from is given)\n" +
	"\n" +
	"Most used:\n" +
	"  -d, --append-desc <text>    description; repeat to append paragraphs\n" +
	"      --add-ac <text>         add an acceptance criterion; repeatable\n" +
	"      --type <value>          configured type\n" +
	"      --priority <value>      configured priority\n" +
	"  -s, --status <value>        configured status (default: the initial one)\n" +
	"  -l, --add-labels <value>    add a label; repeatable or comma-separated\n" +
	"  -a, --add-assignees <@who>  add an assignee; repeatable or comma-separated\n" +
	"      --add-deps <ref>        add a dependency; validated, repeatable\n" +
	"      --due <YYYY-MM-DD>      due date\n" +
	"      --comment <text>        add a discussion comment; repeatable\n" +
	"      --append-plan <text>    implementation plan\n" +
	"      --start                 create it already in the active status, assigned\n" +
	"                              to you, with the lease claimed for you\n" +
	"\n" +
	"Every other field flag of `biso set --help` is accepted too.\n" +
	"\n" +
	"Batch:\n" +
	"      --from <file|->        NDJSON, one task object per line. The only place\n" +
	"                             where id, createdAt, updatedAt, criterion keys,\n" +
	"                             comment timestamps and question timestamps can be\n" +
	"                             given. Validated whole before anything is written.\n" +
	"\n" +
	"Any text option also takes @file to read a file, or - to read stdin.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  created            4  a referenced task or file does not exist\n" +
	"  2  bad usage          5  a text reference matched several tasks\n" +
	"  3  unknown value      8  the board could not be written\n" +
	"  7  batch or --dry-run validation failed, nothing was written\n" +
	"                        20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso new \"Normalize CRLF in the diff\" --type bug --priority high\n" +
	"  biso new \"Add OAuth\" --add-ac \"Login succeeds\" --add-ac \"Token refreshes\"\n" +
	"  biso new \"Rewrite the installer\" -d @docs/installer.md --start\n" +
	"  biso new --from tasks.ndjson --dry-run\n"

// setHelp is `biso set --help` (docs/spec/cmd/set.md).
const setHelp = "Usage: biso set <ref>... [options]\n" +
	"\n" +
	"Change any field of one or more tasks, all or nothing. Every flag here means\n" +
	"the same in `biso new`, `biso start`, `biso note`, `biso comment`, `biso ask`,\n" +
	"`biso answer`, `biso finish` and `biso archive`. Every flag name says what it\n" +
	"does; there is no rule to learn beyond the name.\n" +
	"\n" +
	"List fields that take comma-separated values have four shapes, and there is no\n" +
	"field that breaks them:\n" +
	"  --add-labels X      add one or more       --replace-labels X   replace the whole list\n" +
	"  --rm-labels X       remove one or more    --clear-labels       empty the list\n" +
	"The same works for --assignees, --refs, --docs, --deps and --files.\n" +
	"\n" +
	"Criteria have three, because a criterion's text can contain a comma and so is\n" +
	"never split on one. There is no whole-list replace; do it by clearing and\n" +
	"adding in the same call.\n" +
	"      --add-ac <text>        add a criterion; repeatable\n" +
	"      --rm-ac <sel>          remove by selector; sel is all, 3, 1-4, 1,3,7 or\n" +
	"                             the criterion text. The numbers are stable #N\n" +
	"                             keys. With several tasks, sel has to be all\n" +
	"      --clear-acs            empty the list\n" +
	"      --check-ac <sel>       check criteria, by the same kind of selector\n" +
	"      --uncheck-ac <sel>     the opposite\n" +
	"\n" +
	"Prose fields have two, because a block of text has no single item to remove.\n" +
	"Replace by clearing and appending in the same call.\n" +
	"      --append-desc X (-d)   append a paragraph\n" +
	"      --append-plan X\n" +
	"      --append-note X\n" +
	"      --append-summary X\n" +
	"      --clear-desc / --clear-plan / --clear-notes / --clear-summary\n" +
	"\n" +
	"External fields have three: --ext key=value sets that one key, --rm-ext key\n" +
	"drops it, --clear-ext empties the map. There is no --replace-ext: setting a\n" +
	"key already replaces its value.\n" +
	"\n" +
	"Scalars just take a value: -t/--title, -s/--status, --type, --priority,\n" +
	"-p/--parent, --due, --ordinal, --author. Each has a --clear-<field>. An\n" +
	"empty string is never a way to clear anything.\n" +
	"\n" +
	"Comments:\n" +
	"      --comment <text>            append a comment; repeatable\n" +
	"      --comment-author <@w>       who wrote it (default: you)\n" +
	"      --rm-comment <sel>          remove one or more, by the same kind of\n" +
	"                                   selector as --rm-ac; body and author are\n" +
	"                                   never edited, by any flag\n" +
	"      --set-comment-date <sel>=<instant>\n" +
	"                                   correct only the date of one or more,\n" +
	"                                   instant is YYYY-MM-DDTHH:MM:SSZ\n" +
	"\n" +
	"Resolution:\n" +
	"      --id / --match         force <ref> to be an id, or free text\n" +
	"\n" +
	"Within one call, every --rm-*/--clear-* is applied before every --add-*/\n" +
	"--append-*, regardless of the order they were written in. A --replace-* over\n" +
	"a non-empty list is allowed and warns on stderr with how many items it\n" +
	"replaced.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  done                    5  something matched more than one thing\n" +
	"  2  bad usage               7  --dry-run did not pass\n" +
	"  3  unknown value           8  the board could not be written\n" +
	"  4  a task, criterion or comment was not found\n" +
	"                             20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso set MYP-11 --priority high --add-labels parser\n" +
	"  biso set MYP-11 --check-ac 1,3 --append-note \"Both covered by diff_test.rs\"\n" +
	"  biso set MYP-11 MYP-12 --due 2026-09-20\n" +
	"  biso set \"CRLF\" --clear-desc --append-desc @docs/bugs/BUG-02.md\n"

// lsHelp is the help block of docs/spec/cmd/ls.md.
const lsHelp = "Usage: biso ls [options]\n" +
	"\n" +
	"List tasks, one per line. Shows 30 by default, hides the Done ones and the\n" +
	"archived ones, and says on stderr what it left out. A filter value the board\n" +
	"does not have is an error, never an empty list, so an empty list is a fact.\n" +
	"\n" +
	"Filters (repeat or comma-separate; same field is OR, different fields are AND):\n" +
	"  -s, --status <value>       configured status (default: all but the terminal)\n" +
	"      --not-status <value>   exclude a status\n" +
	"      --any-status           include the terminal status too\n" +
	"      --archived             include archived tasks\n" +
	"      --only-archived        only archived tasks\n" +
	"      --type <value>         configured type\n" +
	"      --priority <value>     configured priority\n" +
	"  -l, --label <value>        label; several labels are ANDed\n" +
	"      --label-or <value>     label; several are ORed\n" +
	"  -a, --assignee <@who>      assignee\n" +
	"      --mine                 assigned to you\n" +
	"      --unassigned           assigned to nobody\n" +
	"  -p, --parent <ref>         subtasks of this task\n" +
	"      --blocked              something unfinished blocks it\n" +
	"      --not-blocked          nothing unfinished blocks it; it may still be\n" +
	"                             waiting on an answer, so add --not-waiting\n" +
	"      --waiting              has an open question\n" +
	"      --not-waiting          has no open question\n" +
	"      --active               in the board's active status\n" +
	"      --not-active           not in the active status\n" +
	"      --overdue              past its due date\n" +
	"      --due-before <date>    due before YYYY-MM-DD\n" +
	"      --search <text>        free text; see `biso get --help` for the scope\n" +
	"      --unchecked            do not check that the labels and assignees you\n" +
	"                             filter by exist on the board; nothing else\n" +
	"                             changes\n" +
	"\n" +
	"Shape:\n" +
	"      --sort <field>         urgency, id, ordinal, due, updated, created, title\n" +
	"      --reverse              flip the whole order, tie-breaks included\n" +
	"      --limit <n>            how many rows to print (default 30, 0 prints none)\n" +
	"      --all                  print every match\n" +
	"      --ids                  print only ids, one per line\n" +
	"      --count                print only how many match\n" +
	"\n" +
	"Columns: id, status, type, priority, title, criteria, assignee, due. Empty\n" +
	"cells print a dash. The title is cut at 100 characters, always, before any\n" +
	"column width is computed. Columns 1 to 7 are padded to the widest value\n" +
	"printed; column 8 never is. Two spaces always separate columns.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  listed, even when empty      5  --parent matched several tasks\n" +
	"  2  bad usage                    6  --mine with no identity configured\n" +
	"  3  a filter value does not exist here\n" +
	"  4  --parent does not exist      8  the board could not respond\n" +
	"                                  20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso ls\n" +
	"  biso ls -s \"In Progress\" --mine\n" +
	"  biso ls --type bug --priority high --limit 10\n" +
	"  biso ls --not-blocked --not-waiting --ids\n" +
	"  biso ls --any-status --archived --all\n"

// getHelp is the help block of docs/spec/cmd/get.md.
const getHelp = "Usage: biso get <ref> [options]\n" +
	"\n" +
	"Show one task. <ref> is an id (MYP-11), a bare number (11) or free text\n" +
	"(\"CRLF\"). Free text that matches several tasks lists them and exits 5; it\n" +
	"never picks one for you.\n" +
	"\n" +
	"Free text searches the title, description, plan, notes, final summary, the\n" +
	"text of the criteria, the body of the comments, the body of the open question\n" +
	"and the labels. A match in the title always\n" +
	"wins over a match anywhere else. `biso ls --search` uses this same scope.\n" +
	"\n" +
	"Options:\n" +
	"      --id                   force <ref> to be read as an id\n" +
	"      --match                force <ref> to be read as free text\n" +
	"      --section <name>       print only these sections; repeatable or comma\n" +
	"                             separated. One of: meta, desc, ac, plan, notes,\n" +
	"                             summary, comments, question\n" +
	"      --explain-urgency      show how the urgency number is built\n" +
	"  -h, --help                 show this help\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  printed          4  not on this board\n" +
	"  2  bad usage        5  the text matched several tasks\n" +
	"  3  the task could not be read\n" +
	"  8  the board could not respond\n" +
	"  20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso get MYP-11\n" +
	"  biso get 11 --section ac\n" +
	"  biso get \"CRLF\"\n" +
	"  biso get MYP-11 --explain-urgency\n"
