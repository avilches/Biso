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
// above, word for word, which help_admin_test.go walks in both directions
// so that neither copy can be reworded on its own.
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
