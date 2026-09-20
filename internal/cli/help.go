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
