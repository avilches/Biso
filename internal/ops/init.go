package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
	"biso/internal/vcs"
)

// InitParams is one `biso init` call, already read off the command line.
// Each list flag carries a Has* companion because
// docs/spec/cmd/init.md#comportamiento gives "not given" a meaning of its
// own with --overwrite-config: the key keeps whatever the board had, which
// is not the same as going back to the default.
type InitParams struct {
	Name    string
	HasName bool

	// At is the text of --at exactly as it was written, relative or
	// absolute, because that same text is what the pointer stores.
	At    string
	HasAt bool

	Statuses    []string
	HasStatuses bool

	InitialStatus     string
	HasInitialStatus  bool
	ActiveStatus      string
	HasActiveStatus   bool
	TerminalStatus    string
	HasTerminalStatus bool

	Types    []string
	HasTypes bool

	Priorities    []string
	HasPriorities bool

	Extensions    []string
	HasExtensions bool

	Prefix    string
	HasPrefix bool

	OverwriteConfig bool

	From    string
	HasFrom bool

	DryRun bool
}

// InitAction is which of the three endings of `biso init` happened, which is
// what decides the first line of its output.
type InitAction int

const (
	// Created is a board that did not exist before, whether its identity
	// was minted or adopted from a pointer or from a marker.
	Created InitAction = iota + 1
	// Adopted is --at naming a board that was already whole, with this
	// project learning where it is and nothing of the board being touched.
	Adopted
	// Rewrote is --overwrite-config over a board that already existed.
	Rewrote
)

// BoardSummary is the board's vocabulary as a result carries it: the same
// data as board.Config, flat and without the types of internal/board, so
// that internal/cli can render it without importing a package the dependency
// rule keeps two layers below it.
type BoardSummary struct {
	Name           string
	Statuses       []string
	InitialStatus  string
	ActiveStatus   string
	TerminalStatus string
	Types          []string
	Priorities     []string
	Extensions     []string
	TaskPrefix     string
}

func summaryOf(cfg board.Config) BoardSummary {
	return BoardSummary{
		Name:           cfg.ProjectName,
		Statuses:       cfg.Statuses,
		InitialStatus:  cfg.InitialStatus,
		ActiveStatus:   cfg.ActiveStatus,
		TerminalStatus: cfg.TerminalStatus,
		Types:          cfg.Types,
		Priorities:     cfg.Priorities,
		Extensions:     cfg.Extensions,
		TaskPrefix:     cfg.TaskPrefix,
	}
}

// InitResult is what `biso init` answers.
type InitResult struct {
	Action InitAction
	// ID is the board's identity, minted, adopted or already there.
	ID    string
	Board BoardSummary
	// Path is the board directory, resolved.
	Path string
	// StoredPath is the text the pointer holds, empty when it holds none.
	StoredPath string
	// PointerWritten says the pointer was written in this call, which is
	// the pointerCreated key of the JSON envelope. The line of the output
	// that says the project points at the board is printed either way.
	PointerWritten bool
	// Notes are the stderr notes of docs/spec/cmd/init.md, in order. Each
	// one carries its own line breaks.
	Notes []string
	// DryRun says nothing was written, and DryRunPath is the path the note
	// of the preview names.
	DryRun     bool
	DryRunPath string
}

// Init creates a board with its configuration, and writes the project
// pointer that lets every copy of the project find it again
// (docs/spec/cmd/init.md).
//
// It never writes outside the board other than that pointer.
func Init(env Env, p InitParams) (*InitResult, error) {
	env = env.WithDefaults()

	if p.HasFrom {
		return nil, notImplementedFrom()
	}
	cfgFlags, err := readVocabularyFlags(p)
	if err != nil {
		return nil, err
	}

	loc, facts, rerr := board.Resolve(board.Search{Dir: env.Dir, Machine: env.Machine})
	if rerr != nil {
		switch rerr.Code {
		case "no_board", "pointer_unresolved":
			// Neither is an error for the one command whose job is to fix
			// them: the first means there is nothing here yet, and the
			// second that the identity to adopt is already written down
			// (docs/spec/cmd/init.md).
		default:
			return nil, rerr
		}
	}

	target := ""
	if p.HasAt {
		target = filepath.Clean(p.At)
		if !filepath.IsAbs(target) {
			target = filepath.Clean(filepath.Join(env.Dir, target))
		}
	}

	if loc != nil {
		// A board whose database does not open does not count as one for
		// this command, which is the whole remedy of
		// docs/spec/garantias.md#el-segundo-caso-la-base-de-datos-que-no-se-puede-leer:
		// `init` rebuilds it in place instead of answering that there is
		// already one here. It is the same directory, so it becomes the
		// destination, and its marker is the identity that gets adopted.
		readable, err := board.DatabaseReadable(loc.Dir)
		if err != nil {
			return nil, err
		}
		if readable {
			return rewrite(env, p, cfgFlags, facts, loc, target)
		}
		if target == "" {
			target = loc.Dir
		}
	}
	return create(env, p, cfgFlags, facts, target)
}

// vocabularyFlags is the half of the parameters that describes the board's
// vocabulary, already checked among themselves.
type vocabularyFlags struct {
	statuses                             []string
	initial, active, terminal            string
	types, priorities, extensions        []string
	hasStatuses, hasTypes, hasPriorities bool
	hasExtensions, hasPrefix, hasName    bool
	prefix, name                         string
	rolesGiven                           bool
}

// readVocabularyFlags applies the rules that relate the vocabulary flags to
// each other, which hold whether the board is new or is being rewritten.
func readVocabularyFlags(p InitParams) (vocabularyFlags, *model.Error) {
	v := vocabularyFlags{
		statuses: p.Statuses, initial: p.InitialStatus, active: p.ActiveStatus,
		terminal: p.TerminalStatus, types: p.Types, priorities: p.Priorities,
		extensions: p.Extensions, hasStatuses: p.HasStatuses, hasTypes: p.HasTypes,
		hasPriorities: p.HasPriorities, hasExtensions: p.HasExtensions,
		hasPrefix: p.HasPrefix, prefix: p.Prefix, hasName: p.HasName, name: p.Name,
	}
	v.rolesGiven = p.HasInitialStatus || p.HasActiveStatus || p.HasTerminalStatus

	if !p.HasStatuses {
		if v.rolesGiven {
			// It names a flag, so it carries field and given, like every
			// other error of code 2 that names one: the only exception
			// declared by docs/spec/contrato-json.md#los-errores-en-json is
			// incompatible_flags. The one named is the first role written
			// down in the order of the table of docs/spec/cmd/init.md,
			// which is the same canonical order a pair of flags is named
			// in, and not the order of argv.
			role, given := firstRoleGiven(p)
			return v, &model.Error{
				ExitCode: 2,
				Code:     "invalid_status_roles",
				Message:  "the status roles are only set together with --statuses",
				Field:    role,
				Given:    given,
				Hints: []string{
					"give --statuses, and then --initial-status, --active-status and --terminal-status",
				},
			}
		}
		return v, checkPrefixFlag(p)
	}

	if len(p.Statuses) < 3 {
		return v, &model.Error{
			ExitCode: 2,
			Code:     "too_few_statuses",
			Message: fmt.Sprintf(
				"--statuses needs at least three statuses, and %d were given",
				len(p.Statuses)),
			Field: "statuses",
			Given: strings.Join(p.Statuses, ","),
			Hints: []string{"a board needs one status for a new task, one for an active one and one for a finished one"},
		}
	}

	var missing []string
	if !p.HasInitialStatus {
		missing = append(missing, "--initial-status")
	}
	if !p.HasActiveStatus {
		missing = append(missing, "--active-status")
	}
	if !p.HasTerminalStatus {
		missing = append(missing, "--terminal-status")
	}
	if len(missing) > 0 {
		return v, &model.Error{
			ExitCode: 2,
			Code:     "invalid_status_roles",
			Message: fmt.Sprintf(
				"--statuses needs --initial-status, --active-status and --terminal-status, and %s missing",
				listAndIsAre(missing)),
			// The flag this one names is --statuses, which is the one that
			// was written and the one that demands the other three.
			Field: "statuses",
			Given: strings.Join(p.Statuses, ","),
		}
	}

	roles := []struct{ flag, value string }{
		{"--initial-status", p.InitialStatus},
		{"--active-status", p.ActiveStatus},
		{"--terminal-status", p.TerminalStatus},
	}
	for _, r := range roles {
		if !contains(p.Statuses, r.value) {
			return v, &model.Error{
				ExitCode: 2,
				Code:     "unknown_status_role",
				Message: fmt.Sprintf("%s names %q, which is not one of --statuses",
					r.flag, r.value),
				Field: strings.TrimPrefix(r.flag, "--"),
				Given: r.value,
				Valid: append([]string(nil), p.Statuses...),
			}
		}
	}
	for i := range roles {
		for j := i + 1; j < len(roles); j++ {
			if roles[i].value == roles[j].value {
				return v, &model.Error{
					ExitCode: 2,
					Code:     "invalid_status_roles",
					Message: fmt.Sprintf("%s and %s both name %q, and the three roles are distinct",
						roles[i].flag, roles[j].flag, roles[i].value),
					Field: strings.TrimPrefix(roles[j].flag, "--"),
					Given: roles[j].value,
				}
			}
		}
	}
	return v, checkPrefixFlag(p)
}

// firstRoleGiven answers the first status role this call wrote, in the
// order of the table of docs/spec/cmd/init.md, with its value.
func firstRoleGiven(p InitParams) (string, string) {
	switch {
	case p.HasInitialStatus:
		return "initial-status", p.InitialStatus
	case p.HasActiveStatus:
		return "active-status", p.ActiveStatus
	default:
		return "terminal-status", p.TerminalStatus
	}
}

func checkPrefixFlag(p InitParams) *model.Error {
	if !p.HasPrefix {
		return nil
	}
	return ValidatePrefix(p.Prefix)
}

// create is `biso init` where no board is reachable from here: it mints or
// adopts an identity, writes the board, and writes the pointer when what is
// written down would not find it.
func create(env Env, p InitParams, v vocabularyFlags, facts board.Searched, target string) (*InitResult, error) {
	name := v.name
	if !v.hasName {
		name = filepath.Base(env.Dir)
	}
	if err := ValidateSlug(name); err != nil {
		return nil, err
	}
	prefix := v.prefix
	if !v.hasPrefix {
		derived, err := DerivePrefix(name)
		if err != nil {
			return nil, err
		}
		prefix = derived
	}

	cfg := board.DefaultConfig(name, prefix)
	if v.hasStatuses {
		cfg.Statuses = v.statuses
		cfg.InitialStatus, cfg.ActiveStatus, cfg.TerminalStatus = v.initial, v.active, v.terminal
	}
	if v.hasTypes {
		cfg.Types = v.types
	}
	if v.hasPriorities {
		cfg.Priorities = v.priorities
	}
	if v.hasExtensions {
		cfg.Extensions = v.extensions
	}

	// The identity: the one the marker of the destination already carries,
	// the one the pointer already names, or a new one. The two adoptions
	// are what keeps two machines talking about the same board
	// (docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero).
	//
	// The destination wins over the pointer, and there is no third answer
	// where the two disagree: getting here at all means this project
	// resolves to no accessible board of its own, and that is precisely
	// when docs/spec/cmd/init.md hands the identity to the marker of the
	// destination and has the pointer rewritten ("cuando no hay ningún
	// puntero aquí, o cuando el que hay no resuelve a nada en esta
	// máquina"). A pointer that does resolve never reaches this function.
	destinationID := ""
	if target != "" {
		destinationID = board.MarkerID(target)
	}
	id, adopted := destinationID, destinationID != ""
	if !adopted && facts.PointerID != "" {
		id, adopted = facts.PointerID, true
	}
	if adopted {
		// The roots are walked for an adopted identity too, and not only
		// for a minted one, because the same id in two of them is the
		// error of docs/spec/resolucion-del-tablero.md#el-mismo-id-en-dos-sitios
		// however this call came by it (the row of code 22 of
		// docs/spec/cmd/init.md says "acuñar o adoptar").
		if _, err := board.FindID(env.Machine, id); err != nil {
			return nil, err
		}
	} else {
		fresh, err := mintID(env)
		if err != nil {
			return nil, err
		}
		id = fresh
	}

	dir := target
	if dir == "" {
		dir = filepath.Join(env.Machine.BoardsRoot, FolderName(name, id))
	}

	// A destination that is already a whole board is adopted and never
	// touched: the operation is local to this project, telling it where its
	// board is (docs/spec/cmd/init.md). One whose database does not open is
	// not whole, and is rebuilt in place instead.
	adoptWhole := false
	if target != "" && destinationID != "" {
		readable, err := board.DatabaseReadable(target)
		if err != nil {
			return nil, err
		}
		adoptWhole = readable
	}

	// The text the pointer stores is the one --at was given, and the one
	// a board that is being rebuilt in place needs to be found again when
	// no --at named it (docs/spec/cmd/init.md).
	stored := ""
	switch {
	case p.HasAt:
		stored = p.At
	case target != "":
		stored = pointerPathFor(env, p, target)
	}
	notes := initNotes(env, p, dir, stored)

	if p.DryRun {
		return &InitResult{
			Action: Created, ID: id, Board: summaryOf(cfg), Path: dir, StoredPath: stored,
			Notes: notes, DryRun: true, DryRunPath: dryRunPath(p, dir),
		}, nil
	}

	result := &InitResult{
		Action: Created, ID: id, Board: summaryOf(cfg),
		Path: dir, StoredPath: stored, Notes: notes,
	}
	if adoptWhole {
		b, err := board.Open(&board.Location{ID: id, Dir: dir, Way: board.WayWorkingDirectory}, env.Machine)
		if err != nil {
			return nil, err
		}
		defer b.Close()
		result.Action, result.Board = Adopted, summaryOf(b.Config)
	} else {
		if err := clearUnreadableDatabase(dir); err != nil {
			return nil, err
		}
		b, err := board.Create(dir, id, cfg, env.Machine)
		if err != nil {
			return nil, err
		}
		defer b.Close()
		if err := board.WriteMarker(dir, id); err != nil {
			return nil, err
		}
		if err := writeIgnoreFile(env, dir); err != nil {
			return nil, err
		}
	}

	written, err := writePointer(env, facts, id, stored)
	if err != nil {
		return nil, err
	}
	result.PointerWritten = written
	return result, nil
}

// clearUnreadableDatabase takes the database out of a directory that is
// about to hold a board again and whose current one cannot be read. That is
// the only case where this command touches a file it did not write: the
// case table of docs/spec/cmd/init.md answers 0 and a board created here
// for a destination "con el marcador pero sin una base de datos legible",
// and a file the program cannot read is not one it can keep either.
func clearUnreadableDatabase(dir string) error {
	readable, err := board.DatabaseReadable(dir)
	if err != nil || readable || !board.HasDatabase(dir) {
		return err
	}
	return board.RemoveDatabase(dir)
}

// rewrite is `biso init` where a board is already reachable: the most likely
// ending of this command, and an error unless --overwrite-config says to
// replace its configuration, which never touches a task.
func rewrite(env Env, p InitParams, v vocabularyFlags, facts board.Searched, loc *board.Location, target string) (*InitResult, error) {
	if !p.OverwriteConfig {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "board_exists",
			Message:  fmt.Sprintf("this project already has board %s, at %s", loc.ID, loc.Dir),
			Hints: []string{
				"`biso where` says which rule picked it",
				"--overwrite-config rewrites its configuration and never touches its tasks",
			},
		}
	}
	if target != "" && target != loc.Dir {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "board_exists",
			Message: fmt.Sprintf("this project already has board %s, at %s, and --at names %s",
				loc.ID, loc.Dir, target),
			Hints: []string{"--overwrite-config rewrites the configuration of the board this project already has"},
		}
	}

	b, err := board.Open(loc, env.Machine)
	if err != nil {
		return nil, err
	}
	defer b.Close()

	cfg := b.Config
	if v.hasName {
		if err := ValidateSlug(v.name); err != nil {
			return nil, err
		}
		cfg.ProjectName = v.name
	}
	if v.hasStatuses {
		cfg.Statuses = v.statuses
		cfg.InitialStatus, cfg.ActiveStatus, cfg.TerminalStatus = v.initial, v.active, v.terminal
	}
	if v.hasTypes {
		cfg.Types = v.types
	}
	if v.hasPriorities {
		cfg.Priorities = v.priorities
	}
	if v.hasExtensions {
		cfg.Extensions = v.extensions
	}
	if v.hasPrefix {
		cfg.TaskPrefix = v.prefix
	}

	if err := checkNothingInUseIsRemoved(b, cfg, v); err != nil {
		return nil, err
	}

	result := &InitResult{Action: Rewrote, ID: b.Location.ID, Board: summaryOf(cfg), Path: b.Location.Dir}
	if p.DryRun {
		result.DryRun = true
		result.DryRunPath = b.Location.Dir
		return result, nil
	}
	if err := b.Rewrite(cfg); err != nil {
		return nil, err
	}

	// The output says the project points at this board, and that sentence
	// is only true if something points at it: docs/spec/cmd/init.md prints
	// it "se escriba el puntero en esta llamada o ya estuviera escrito de
	// antes", which leaves no third case where nobody wrote one. It happens
	// when the board was reached by the first way, standing inside it, with
	// no pointer anywhere between here and the cap of the search, and the
	// answer is to write it, the same way a creation does.
	if facts.PointerDir == "" {
		written, err := writePointer(env, facts, b.Location.ID, pointerPathFor(env, p, b.Location.Dir))
		if err != nil {
			return nil, err
		}
		result.PointerWritten = written
	}
	return result, nil
}

// pointerPathFor is the `path` key a pointer needs to find a board that is
// already there: none when the board sits directly under one of the
// machine's roots, because the search by marker finds it
// (docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero),
// the text of --at when this call gave one, and the board's own directory
// otherwise.
func pointerPathFor(env Env, p InitParams, dir string) string {
	if p.HasAt {
		return p.At
	}
	for _, root := range env.Machine.Roots() {
		if filepath.Dir(dir) == filepath.Clean(root) {
			return ""
		}
	}
	return dir
}

// checkNothingInUseIsRemoved is the exit code 6 of
// docs/spec/cmd/init.md: rewriting a configuration can never take away a
// value that a task already carries, nor change the prefix that identifiers
// already have embedded in them.
func checkNothingInUseIsRemoved(b *board.Board, cfg board.Config, v vocabularyFlags) *model.Error {
	tasks, _, err := b.Tasks.All()
	if err != nil {
		if e, ok := err.(*model.Error); ok {
			return e
		}
		return nil
	}
	if len(tasks) == 0 {
		return nil
	}
	if v.hasPrefix && cfg.TaskPrefix != b.Config.TaskPrefix {
		return &model.Error{
			ExitCode: 6,
			Code:     "board_inconsistent",
			Message: fmt.Sprintf(
				"task_prefix cannot change from %q to %q on a board that already has %s",
				b.Config.TaskPrefix, cfg.TaskPrefix, plural(len(tasks), "task")),
			Field: "task_prefix",
			Given: cfg.TaskPrefix,
			Hints: []string{"export the board, rewrite the identifiers and import them into a new one"},
		}
	}

	inUse := map[string]map[string][]string{
		"statuses": {}, "types": {}, "priorities": {}, "extensions": {},
	}
	for _, t := range tasks {
		inUse["statuses"][t.Status] = append(inUse["statuses"][t.Status], t.ID)
		inUse["types"][t.Type] = append(inUse["types"][t.Type], t.ID)
		inUse["priorities"][t.Priority] = append(inUse["priorities"][t.Priority], t.ID)
		for key := range t.Ext {
			inUse["extensions"][key] = append(inUse["extensions"][key], t.ID)
		}
	}
	checks := []struct {
		key     string
		given   bool
		kept    []string
		subject string
	}{
		{"statuses", v.hasStatuses, cfg.Statuses, "status"},
		{"types", v.hasTypes, cfg.Types, "type"},
		{"priorities", v.hasPriorities, cfg.Priorities, "priority"},
		{"extensions", v.hasExtensions, cfg.Extensions, "extension key"},
	}
	for _, c := range checks {
		if !c.given {
			continue
		}
		for value, ids := range inUse[c.key] {
			if value == "" || contains(c.kept, value) {
				continue
			}
			sort.Strings(ids)
			return &model.Error{
				ExitCode: 6,
				Code:     "board_inconsistent",
				Message: fmt.Sprintf("%s %q is still in use by %s: %s",
					c.subject, value, plural(len(ids), "task"), strings.Join(ids, ", ")),
				Field: c.key,
				Given: value,
			}
		}
	}
	return nil
}

// mintID generates an identity and makes sure this machine does not have it
// already, which is one directory listing per root
// (docs/spec/resolucion-del-tablero.md#cómo-biso-init-genera-el-id-y-escribe-el-puntero).
func mintID(env Env) (string, error) {
	for attempt := 0; attempt < 100; attempt++ {
		id, err := env.NewID()
		if err != nil {
			return "", err
		}
		hits, err := board.FindID(env.Machine, id)
		if err != nil {
			return "", err
		}
		if len(hits) == 0 {
			return id, nil
		}
	}
	return "", &model.Error{
		ExitCode: 8,
		Code:     "io_error",
		Message:  "a board id that this machine does not already have could not be generated",
	}
}

// writePointer writes the project pointer when what is written down would
// not find the board that was just created or adopted, and leaves it alone
// when it would.
//
// docs/spec/cmd/init.md spells out the two ends of this rule: a pointer that
// already names the right identity is not rewritten, because it was already
// correct, and a project that resolves to nothing gets one written. The
// middle, a correct identity whose path no longer names where the board
// went, is the same question answered the same way: the pointer is only
// touched when leaving it would leave the project unable to find its board.
func writePointer(env Env, facts board.Searched, id, stored string) (bool, error) {
	if facts.PointerDir != "" && facts.PointerID == id {
		current, err := board.ReadPointer(filepath.Join(facts.PointerDir, board.PointerFile))
		if err != nil {
			return false, err
		}
		if current.Path == stored {
			return false, nil
		}
	}
	p := board.Pointer{Version: board.PointerVersion, ID: id, Path: stored}
	if err := board.WritePointer(env.Dir, p); err != nil {
		return false, err
	}
	return true, nil
}

// writeIgnoreFile writes the exclusion file of the version control system
// this machine is configured for, holding the database and its two
// auxiliaries, so that the day the directory becomes a repository only
// snapshot.ndjson, board.json and the marker are ever versioned
// (docs/spec/cmd/init.md). With vcs none there is none to write.
func writeIgnoreFile(env Env, dir string) error {
	name := vcs.IgnoreFile(VCSConfig(env.Machine))
	if name == "" {
		return nil
	}
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		// The directory already had one, which is the case of a board
		// rebuilt where a versioned folder used to be. Rewriting somebody
		// else's exclusion file would be going further than this command
		// ever goes.
		return nil
	}
	content := board.DatabaseFile + "\n" +
		board.DatabaseFile + "-wal\n" +
		board.DatabaseFile + "-shm\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return &model.Error{
			ExitCode: 8,
			Code:     "io_error",
			Message:  fmt.Sprintf("%s cannot be written: %s", path, err),
			Field:    path,
		}
	}
	return nil
}

// initNotes are the two notes of docs/spec/cmd/init.md, in the order that
// page prints them. Their line breaks are fixed text and not a width: the
// same note breaks between the same words whatever folder name it carries.
func initNotes(env Env, p InitParams, dir, stored string) []string {
	if !p.HasAt {
		return nil
	}
	var notes []string
	if dir != env.Dir && strings.HasPrefix(dir, env.Dir+string(filepath.Separator)) {
		folder := strings.TrimPrefix(dir, env.Dir+string(filepath.Separator))
		note := fmt.Sprintf(
			"the board lives inside this project. Ignore %s/ and the board keeps\n"+
				"its own history; version it and the snapshot travels with your code.",
			folder)
		if ignore := vcs.IgnoreFile(VCSConfig(env.Machine)); ignore != "" {
			note += fmt.Sprintf(
				"\nEither way the database stays out, %s/%s excludes it", folder, ignore)
		}
		notes = append(notes, note)
	}
	if stored != "" && !filepath.IsAbs(stored) {
		notes = append(notes, fmt.Sprintf(
			"the location is stored as the relative path %q. A working copy\n"+
				"outside this project will not have that folder while git ignores it, so\n"+
				"it will not find the board: use an absolute --at if you work that way",
			stored))
	}
	return notes
}

// dryRunPath is the path the preview's note names: the one --at gave, in the
// same form it gave it, and the resolved one otherwise.
func dryRunPath(p InitParams, dir string) string {
	if p.HasAt {
		return p.At
	}
	return dir
}

// notImplementedFrom is the one branch of this command that step 4 of the
// implementation does not cover: restoring a snapshot needs the interchange
// format, which arrives with `biso export` and `biso new --from`
// (docs/spec/estado-de-implementacion.md).
func notImplementedFrom() *model.Error {
	return &model.Error{
		ExitCode: 1,
		Code:     "internal",
		Message:  "biso init --from is not implemented yet",
		Hints:    []string{"it arrives with `biso export` and `biso snapshot`"},
	}
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// listAndIsAre writes a list of flags and the verb that agrees with it, so
// that one missing flag reads "--active-status is missing" and two read
// "--active-status and --terminal-status are missing".
func listAndIsAre(items []string) string {
	verb := " are"
	if len(items) == 1 {
		verb = " is"
	}
	switch len(items) {
	case 1:
		return items[0] + verb
	case 2:
		return items[0] + " and " + items[1] + verb
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1] + verb
	}
}
