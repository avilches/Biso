package cli

// This file completes the help of internal/cli/help.go with the
// administrative commands, and holds the catalog `biso help` answers from
// (docs/spec/cmd/help.md).
//
// The catalog is the list of every command biso has, with the one-line
// summary the top-level help prints for it. It is the whole answer of
// `biso help --json`, and the set of names `biso help <command>...`
// resolves against, so a name is in it whether or not this build carries
// the logic behind that name yet: the help of a command is part of the
// interface the specification fixes, and `biso help all` already prints
// every one of these names.

// archiveHelp is `biso archive --help` (docs/spec/cmd/archive.md).
const archiveHelp = "Usage: biso archive <ref>... [options]\n" +
	"\n" +
	"Take tasks off the board without losing them. An archived task still exists,\n" +
	"`biso get` still finds it, `biso ls --archived` lists it, and its id is never\n" +
	"reused.\n" +
	"\n" +
	"Options:\n" +
	"      --unarchive      put them back on the board\n" +
	"      --id / --match   force <ref> to be an id, or free text\n" +
	"  -h, --help           show this help\n" +
	"\n" +
	"Every field flag of `biso set --help` works here too.\n" +
	"\n" +
	"There is no delete command. Archiving is the way.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  archived       3  the task could not be read     7  --dry-run did not pass\n" +
	"  2  bad usage      4  not found                      8  could not be written\n" +
	"                    5  ambiguous                     20  no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso archive MYP-11\n" +
	"  biso archive MYP-11 MYP-12 MYP-13\n" +
	"  biso archive MYP-11 --unarchive\n"

// configHelp is `biso config --help` (docs/spec/cmd/config.md).
const configHelp = "Usage: biso config get <key>\n" +
	"       biso config set <key> <value>\n" +
	"       biso config list [--json]\n" +
	"\n" +
	"Read and change the board configuration. List values are comma-separated.\n" +
	"No configuration change ever touches a task.\n" +
	"\n" +
	"Keys:\n" +
	"  project_name       board name; changing it never moves anything on disk,\n" +
	"                     the board folder keeps whatever name it has\n" +
	"  statuses           the board statuses, in order\n" +
	"  initial_status     status of a new task           (one of statuses)\n" +
	"  active_status      what `biso start` sets         (one of statuses)\n" +
	"  terminal_status    what `biso finish` sets        (one of statuses)\n" +
	"  types              configured task types\n" +
	"  priorities         configured priorities\n" +
	"  labels             labels that filters accept on top of the ones in use\n" +
	"  assignees          assignees that filters accept on top of the ones in use\n" +
	"  extensions         declared external field keys, such as trello.card\n" +
	"  task_prefix        id prefix, letters only (default: derived from\n" +
	"                     project_name); immutable once the board has a task\n" +
	"  finish_strict      make `biso finish` refuse an incomplete task\n" +
	"  lease_minutes      lease duration in minutes (default 240); free to change\n" +
	"                     at any time, it only affects future renewals\n" +
	"  urgency.priority, urgency.active, urgency.blocking, urgency.blocked,\n" +
	"  urgency.due, urgency.criteria, urgency.age\n" +
	"                     the seven urgency coefficients; see `biso get --explain-urgency`\n" +
	"\n" +
	"The three special statuses are stored as explicit values. Changing `statuses`\n" +
	"never moves them; if a change would remove one of them, it fails and says so.\n" +
	"\n" +
	"Removing any configured value that a task still uses is refused, never applied\n" +
	"silently.\n" +
	"\n" +
	"Options:\n" +
	"      --json         machine-readable output, `list` only\n" +
	"  -h, --help         show this help\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  done            4  no such key\n" +
	"  2  bad usage       6  the change would leave the board inconsistent\n" +
	"  3  bad value       8  the configuration could not be written\n" +
	"                     20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso config get active_status\n" +
	"  biso config set statuses \"To Do,In Progress,Done\"\n" +
	"  biso config set finish_strict true\n" +
	"  biso config list --json\n"

// doctorHelp is `biso doctor --help` (docs/spec/cmd/doctor.md).
const doctorHelp = "Usage: biso doctor [options]\n" +
	"\n" +
	"Check the board for duplicate ids, unreadable tasks, undeclared extension keys,\n" +
	"values that are no longer configured, a broken status-role invariant, broken\n" +
	"dependencies, dependency cycles, parent cycles, repeated criterion keys, a lease\n" +
	"on a task that is not both active and assigned, a recorded highest id that has\n" +
	"fallen behind, a database that fails its integrity check, a missing or\n" +
	"mismatched <id>.id marker, an extra board root that cannot be read, an exclusion\n" +
	"file that no longer matches the configured vcs, and a board directory on a\n" +
	"filesystem where SQLite's WAL mode is not safe.\n" +
	"\n" +
	"Options:\n" +
	"      --fix      repair what can be repaired without a decision\n" +
	"  -h, --help     show this help\n" +
	"\n" +
	"Without --fix this is a read-only command: --print and --dry-run are bad usage\n" +
	"here, same as in any other read-only command. With --fix, --dry-run reports\n" +
	"what would be fixed without fixing it, same report as a real run, same exit\n" +
	"code too: 6 if an error would remain unfixed, 0 otherwise. Never its own 7,\n" +
	"doctor has none.\n" +
	"\n" +
	"Findings come in two levels: errors, which leave the board inconsistent or\n" +
	"unreliable, and warnings, which are true and worth knowing but fix nothing.\n" +
	"Only remaining errors produce exit code 6.\n" +
	"\n" +
	"An unreadable task is reported and skipped, never a reason to stop. A database\n" +
	"that cannot be opened, or that fails its integrity check, is not a finding: the\n" +
	"whole command fails instead, with exit code 21.\n" +
	"Gaps in the id sequence are normal and are not reported.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  nothing wrong, or every error found was fixed\n" +
	"  2  bad usage\n" +
	"  6  errors remain\n" +
	"  8  cannot write while fixing\n" +
	"  20 no board here\n" +
	"  21 its database could not be read\n" +
	"\n" +
	"Examples:\n" +
	"  biso doctor\n" +
	"  biso doctor --fix\n"

// helpHelp is `biso help --help` (docs/spec/cmd/help.md).
const helpHelp = "Usage: biso help [command...|all]\n" +
	"\n" +
	"Print the top-level help, the help of one or more commands, or the top-level\n" +
	"help plus the nine administrative commands with `all`. Works without a board.\n" +
	"\n" +
	"Given more than one command name, prints the full help of each one, in the\n" +
	"order given, in the same call, as if `biso <cmd> --help` had been called once\n" +
	"per name, with a blank line between one command's help and the next. A name\n" +
	"that does not exist fails the whole call before anything is printed, even the\n" +
	"help of the names that do exist earlier in the list. `all` is not a command\n" +
	"name and does not combine with one.\n" +
	"\n" +
	"With --json this prints the command list and its one-line summaries, never\n" +
	"the prose help: a help text is written to be read, and the list is the part\n" +
	"that is data. Given several names, `commands` carries one element per name,\n" +
	"in the same order.\n" +
	"\n" +
	"Arguments:\n" +
	"  command        one or more command names, or `all` on its own\n" +
	"\n" +
	"Options:\n" +
	"      --json     machine-readable envelope with the command list\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  help printed\n" +
	"  2  usage error, such as `all` combined with a command name\n" +
	"  4  no such command\n" +
	"\n" +
	"Examples:\n" +
	"  biso help\n" +
	"  biso help finish\n" +
	"  biso help finish note\n" +
	"  biso help all\n"

// boardHelp is `biso board --help` (docs/spec/cmd/board.md). The command
// itself is out of the scope of version 1.0 by an explicit decision, and its
// help is here all the same: `biso help all` names `board` among the
// administrative commands, so `biso help board` has to answer with it.
const boardHelp = "Usage: biso board [options]\n" +
	"\n" +
	"Open the interactive board in a browser. This is the only command that opens\n" +
	"an interface: every other one prints text and exits, with or without a\n" +
	"terminal.\n" +
	"\n" +
	"Options:\n" +
	"      --port <n>   port to listen on, 1024 to 65535 (default 6420)\n" +
	"      --no-open    print the address, do not open a browser, and do not\n" +
	"                   require a terminal\n" +
	"      --json       machine-readable envelope, printed when the server starts\n" +
	"  -h, --help       show this help\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  stopped cleanly\n" +
	"  2  bad usage\n" +
	"  8  no terminal without --no-open, or the port is taken\n" +
	"  20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso board\n" +
	"  biso board --port 7000 --no-open\n"

// exportHelp is `biso export --help` (docs/spec/cmd/export.md), and
// snapshotHelp is `biso snapshot --help` (docs/spec/cmd/snapshot.md). The
// two commands belong to another step and are not in the table of
// commands.go yet; their help is here for the same reason board's is, so
// that every name `biso help all` prints answers `biso help <name>`.
const exportHelp = "Usage: biso export [options]\n" +
	"\n" +
	"Write the board as NDJSON, one task per line, in exactly the shape that\n" +
	"`biso new --from` reads back. Round-tripping every non-derived field is a\n" +
	"tested guarantee: ids, dates, criterion and comment keys, and checkmarks\n" +
	"included.\n" +
	"\n" +
	"With no filters it exports everything, the finished and the archived included.\n" +
	"It never inherits the default limit or the default status filter of `biso ls`.\n" +
	"If a task cannot be read, the rest is still written and the exit code is 6,\n" +
	"not 0: this is the one command whose purpose is to lose nothing.\n" +
	"\n" +
	"Options:\n" +
	"  -o, --out <file|->   where to write (default: stdout)\n" +
	"      --no-archived    leave the archived tasks out\n" +
	"  -h, --help           show this help\n" +
	"\n" +
	"Every filter of `biso ls` works here except --archived and --only-archived,\n" +
	"which do not apply because archived tasks are already included by default.\n" +
	"Its shaping flags (--sort, --limit, --all, --ids, --count) do not apply either.\n" +
	"--json is rejected with code 2: this output is already one JSON object per\n" +
	"line, while --json means the single envelope every other command prints.\n" +
	"\n" +
	"Derived fields are never written: urgency, acDone, acTotal, commentCount,\n" +
	"blocks, blocked, waiting, leaseExpired.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  exported       3  a filter value does not exist here\n" +
	"  2  bad usage      6  some task was skipped, unreadable\n" +
	"  8  cannot write there                20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso export -o backup.ndjson\n" +
	"  biso export -s Done --no-archived -o done.ndjson\n"

const snapshotHelp = "Usage: biso snapshot [options]\n" +
	"\n" +
	"Write snapshot.ndjson and board.json into the board's own directory, then\n" +
	"record them with the version control system this machine is configured for\n" +
	"(the vcs key, git by default). Those two files and the <id>.id marker are\n" +
	"what `biso init --from` reads back to rebuild a board whole: its tasks, in\n" +
	"the same shape `biso export` writes, its configuration, in the same shape\n" +
	"`biso config list --json` prints, and its identity.\n" +
	"\n" +
	"Where the revision lands follows the board's directory. If it is a\n" +
	"repository of its own, there. If it sits inside another repository that does\n" +
	"not ignore it, in that one, beside the code, which is what makes the\n" +
	"snapshot travel to other machines on its own. Otherwise this command creates\n" +
	"the board its own repository, lazily, the first time it runs there.\n" +
	"\n" +
	"biso snapshot is the only command that ever runs another program, and the\n" +
	"only one that creates a repository. Invoking version control costs about\n" +
	"12ms, more than the 25ms startup budget for biso ls and biso prime allows on\n" +
	"the hot path.\n" +
	"\n" +
	"Options:\n" +
	"      --vcs <mode>   none, commit or push (default commit)\n" +
	"  -h, --help         show this help\n" +
	"\n" +
	"The two files are always written, whatever version control does. If it is\n" +
	"not installed, or creating the repository fails, the commit is skipped with\n" +
	"a note: it is optional, and its absence never fails this command. A commit\n" +
	"that fails once it is really attempted is an error, nothing is lost, and\n" +
	"running this command again after fixing the reason is all it takes.\n" +
	"\n" +
	"Whatever those commands print is forwarded on stderr, prefixed with the\n" +
	"system name, and never suppressed. There is no timeout: interrupting a\n" +
	"half-written revision is not safe, and nothing is lost by killing this\n" +
	"command, since both files are on disk before the first one runs.\n" +
	"\n" +
	"Exit codes:\n" +
	"  0  written, and recorded if that applied\n" +
	"  2  bad usage, or --vcs push with no publish command configured\n" +
	"  6  some task was skipped, unreadable\n" +
	"  8  cannot write there, or the commit or the push failed\n" +
	"  20 no board here\n" +
	"\n" +
	"Examples:\n" +
	"  biso snapshot\n" +
	"  biso snapshot --vcs none\n" +
	"  biso snapshot --vcs push\n"

// adminBlock is the second half of `biso help all`: the nine administrative
// commands, each one with its line, under the top-level help
// (docs/spec/cmd/help.md#salida-de-biso-help-all).
const adminBlock = "Administration:\n" +
	"  init               create a board for this project\n" +
	"  where              say which board is in use and why\n" +
	"  archive <ref>      take a task off the board (there is no delete)\n" +
	"  export             dump the board as NDJSON that `biso new --from` reads back\n" +
	"  config             read and change the board configuration\n" +
	"  doctor             check the board, and repair what can be repaired\n" +
	"  board              open the interactive board\n" +
	"  help [cmd|all]     this\n" +
	"  snapshot           write snapshot.ndjson and board.json, then record them\n"

// CommandEntry is one row of the catalog: the command's name and the
// one-line summary the two help blocks print beside it.
type CommandEntry struct {
	Name    string
	Summary string
}

// commandCatalog is every command of biso, in the order the help prints
// them: the one to start with, the daily ones, and then the administrative
// ones. The summaries are the ones of the top-level help and of the block
// above, word for word; a test of this package checks that they still are.
var commandCatalog = []CommandEntry{
	{"prime", "everything you need to work on this board, in one message"},
	{"ls", "list tasks, most urgent first"},
	{"get", "show one task"},
	{"new", "create a task and print its id"},
	{"set", "change any field"},
	{"start", "take a task"},
	{"note", "append an implementation note"},
	{"comment", "append a discussion comment"},
	{"finish", "close a task"},
	{"ask", "park on a question"},
	{"answer", "answer it and unpark"},
	{"init", "create a board for this project"},
	{"where", "say which board is in use and why"},
	{"archive", "take a task off the board (there is no delete)"},
	{"export", "dump the board as NDJSON that `biso new --from` reads back"},
	{"config", "read and change the board configuration"},
	{"doctor", "check the board, and repair what can be repaired"},
	{"board", "open the interactive board"},
	{"help", "this"},
	{"snapshot", "write snapshot.ndjson and board.json, then record them"},
}

// Catalog answers a copy of it, for a test that walks it.
func Catalog() []CommandEntry {
	out := make([]CommandEntry, len(commandCatalog))
	copy(out, commandCatalog)
	return out
}

// catalogNames is the names alone, in the same order, which is what the
// suggestion of a name that does not exist is drawn from.
func catalogNames() []string {
	names := make([]string, 0, len(commandCatalog))
	for _, c := range commandCatalog {
		names = append(names, c.Name)
	}
	return names
}

// inCatalog answers whether that name is a command of biso.
func inCatalog(name string) bool {
	for _, c := range commandCatalog {
		if c.Name == name {
			return true
		}
	}
	return false
}
