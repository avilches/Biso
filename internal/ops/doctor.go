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
	"biso/internal/walprobe"
)

// This file is `biso doctor` (docs/spec/cmd/doctor.md): the one place where
// "is something wrong with this board" can be asked without having tried an
// operation first, and the one place authorized to write what nobody asked
// for one by one, which is what --fix means.
//
// Two things about its shape follow the page and are easy to get wrong.
// The count of the first line is what the check found, repaired or not, so
// a repaired error moves from `problems` to `fixed` and the two lists add
// up to that number. And an unreadable task is a finding like any other:
// it is reported and the rest of the board is still checked, because the
// only thing that ever stops this command is a database that does not open.

// DoctorParams is one `biso doctor` call.
type DoctorParams struct {
	// Fix is the consent to write what was not asked for one by one.
	Fix bool
	// DryRun previews a --fix: the same report, nothing written.
	DryRun bool
}

// Finding is one thing the check found: the task it is about, empty when it
// is about the board itself, its stable code and its message.
type Finding struct {
	Task    string
	Code    string
	Message string
}

// DoctorResult is the report.
type DoctorResult struct {
	// Problems are the errors that remain, which is not every error that
	// was found: a repaired one moves to Fixed.
	Problems []Finding
	Warnings []Finding
	// Fixed are the repairs that were applied, or that would be applied
	// with --dry-run.
	Fixed  []Finding
	DryRun bool
}

// Doctor checks the board.
func Doctor(env Env, p DoctorParams) (*DoctorResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return DoctorOn(b, env, p)
}

// DoctorOn is Doctor over a board that is already open.
func DoctorOn(b *board.Board, env Env, p DoctorParams) (*DoctorResult, error) {
	// The integrity check runs first and, when it fails, there is no
	// report at all: the whole command aborts with the message and the
	// exit code 21 of
	// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar.
	// Opening a board never runs it, because it reads every page and the
	// startup budget cannot pay for that on every call.
	if err := b.Store.CheckIntegrity(); err != nil {
		return nil, err
	}

	d := &doctor{b: b, env: env, result: &DoctorResult{DryRun: p.DryRun}}
	if err := d.check(); err != nil {
		return nil, err
	}
	if !p.Fix {
		return d.result, nil
	}
	if err := d.fix(p.DryRun); err != nil {
		return nil, err
	}
	return d.result, nil
}

// doctor is one run of the command.
type doctor struct {
	b      *board.Board
	env    Env
	result *DoctorResult

	// repairs are the data repairs the errors found ask for, and marker
	// is the identity marker that has to be written, empty when it is
	// there. They are collected while checking and applied afterwards,
	// which is what lets --dry-run answer the same report without
	// writing.
	repairs board.Repairs
	marker  string
	// problems are the errors found, each with what repairing it would
	// say when --fix can repair it, so that a fixed error moves from one
	// list to the other instead of being listed twice. They are kept
	// together, and not as an index into the result, because the report
	// is sorted before it is answered.
	problems []pending
}

// pending is one error found, and the line the repaired version of it
// would print. repaired is nil when --fix cannot repair it.
type pending struct {
	found    Finding
	repaired *Finding
}

func (d *doctor) problem(task, code, message string) {
	d.problems = append(d.problems, pending{
		found: Finding{Task: task, Code: code, Message: message}})
}

// fixableProblem records an error together with what repairing it would
// say, so --fix can move it from one list to the other.
func (d *doctor) fixableProblem(task, code, problem, repaired string) {
	d.problems = append(d.problems, pending{
		found:    Finding{Task: task, Code: code, Message: problem},
		repaired: &Finding{Code: code, Message: repaired},
	})
}

func (d *doctor) warning(task, code, message string) {
	d.result.Warnings = append(d.result.Warnings, Finding{Task: task, Code: code, Message: message})
}

// checkRank is the position of every check in the table of
// docs/spec/cmd/doctor.md#qué-comprueba, which is the order the report
// lists its findings in
// (docs/spec/cmd/doctor.md#el-orden-en-que-sale-el-informe). The checks
// themselves do not run in that order, because several of them are asked of
// each task in one pass over the board, so the report is sorted by this map
// before it is answered.
var checkRank = map[string]int{
	"duplicate_id":             1,
	"task_unreadable":          2,
	"undeclared_extension_key": 3,
	"value_not_configured":     4,
	"status_role_unknown":      5,
	"status_role_invalid":      6,
	"dependency_not_found":     7,
	"dependency_cycle":         8,
	"parent_cycle":             9,
	"duplicate_criterion_key":  10,
	"lease_invariant":          11,
	"highest_id_behind":        12,
	"marker_missing":           13,
	"marker_id_mismatch":       14,
	"extra_root_unreadable":    15,
	"unsafe_wal_filesystem":    16,
	"ignore_file_mismatch":     17,
}

// inTableOrder sorts the findings the way the table lists the checks. The
// sort is stable, so two findings of the same check keep the order the
// check produced them in, which is ascending identifier for everything
// asked of a task.
func inTableOrder[T any](items []T, codeOf func(T) string) {
	sort.SliceStable(items, func(i, j int) bool {
		return checkRank[codeOf(items[i])] < checkRank[codeOf(items[j])]
	})
}

// check runs every check of the table of docs/spec/cmd/doctor.md#qué-comprueba.
// They do not run in the order of that table, because the ones that are
// about a task are all asked in a single pass over the board; the report is
// put back into the table's order at the end.
func (d *doctor) check() error {
	if err := d.checkDuplicateIDs(); err != nil {
		return err
	}
	tasks, skipped, err := d.b.Tasks.All()
	if err != nil {
		return err
	}
	d.checkUnreadable(skipped)
	d.checkStatusRoles()
	if err := d.checkTasks(tasks); err != nil {
		return err
	}
	if err := d.checkHighestID(tasks); err != nil {
		return err
	}
	if err := d.checkMarker(); err != nil {
		return err
	}
	d.checkExtraRoots()
	d.checkFilesystem()
	d.checkIgnoreFile()

	inTableOrder(d.problems, func(p pending) string { return p.found.Code })
	inTableOrder(d.result.Warnings, func(w Finding) string { return w.Code })
	d.result.Problems = make([]Finding, 0, len(d.problems))
	for _, p := range d.problems {
		d.result.Problems = append(d.result.Problems, p.found)
	}
	return nil
}

// checkDuplicateIDs and checkDuplicateCriterionKeys ask the two questions
// that today's schema already answers by construction, because what is
// being checked is whether the data is sound and not whether this program
// is the one that wrote it.
func (d *doctor) checkDuplicateIDs() error {
	duplicates, err := d.b.DuplicateIDs()
	if err != nil {
		return err
	}
	for _, id := range sortedKeys(duplicates) {
		d.problem(id, "duplicate_id", fmt.Sprintf(
			"id %q is used by %d tasks, ids must be unique", id, duplicates[id]))
	}
	return nil
}

func (d *doctor) checkDuplicateCriterionKeys() error {
	duplicates, err := d.b.DuplicateCriterionKeys()
	if err != nil {
		return err
	}
	for _, pair := range sortedKeys(duplicates) {
		id, key, _ := strings.Cut(pair, " ")
		d.problem(id, "duplicate_criterion_key", fmt.Sprintf(
			"acceptance criterion key #%s is used by %d criteria, keys must be unique within a task",
			key, duplicates[pair]))
	}
	return nil
}

// checkUnreadable reports every task a set read had to leave out. It is an
// error and never a reason to stop, which is the whole difference between
// one bad task and a database that does not open
// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
func (d *doctor) checkUnreadable(skipped []board.Skipped) {
	for _, s := range skipped {
		d.problem(s.ID, "task_unreadable", fmt.Sprintf(
			"task %q could not be parsed: %s", s.ID, reasonOf(s)))
	}
}

// reasonOf is why a task could not be read, without the identifier the
// message of internal/board already opens with, because this one names the
// task itself.
func reasonOf(s board.Skipped) string {
	if s.Reason == nil {
		return "unknown reason"
	}
	return strings.TrimPrefix(s.Reason.Message, s.ID+" cannot be read: ")
}

// checkStatusRoles is the two rows about the configuration: a role that
// names a status the board does not have, and a list of statuses that
// cannot hold three distinct roles.
func (d *doctor) checkStatusRoles() {
	cfg := d.b.Config
	roles := []struct{ key, value string }{
		{board.KeyInitialStatus, cfg.InitialStatus},
		{board.KeyActiveStatus, cfg.ActiveStatus},
		{board.KeyTerminalStatus, cfg.TerminalStatus},
	}
	for _, role := range roles {
		if containsString(cfg.Statuses, role.value) {
			continue
		}
		d.problem("", "status_role_unknown", fmt.Sprintf(
			"%s %q is not one of the configured statuses %q",
			role.key, role.value, strings.Join(cfg.Statuses, ", ")))
	}
	if len(cfg.Statuses) < minimumStatuses {
		d.problem("", "status_role_invalid", fmt.Sprintf(
			"statuses has %d elements, at least %d are required",
			len(cfg.Statuses), minimumStatuses))
	}
	for i := range roles {
		for j := i + 1; j < len(roles); j++ {
			if roles[i].value != roles[j].value {
				continue
			}
			d.problem("", "status_role_invalid", fmt.Sprintf(
				"%s and %s are both %q, the three roles must be distinct",
				roles[i].key, roles[j].key, roles[i].value))
		}
	}
}

// checkTasks is every row of the table that is about one task, asked of
// each task in ascending identifier order so that the report reads in the
// order of the board.
func (d *doctor) checkTasks(tasks []*model.Task) error {
	byID := make(map[string]*model.Task, len(tasks))
	for _, t := range tasks {
		byID[t.ID] = t
	}
	cfg := d.b.Config
	reportedDependency := map[string]bool{}
	reportedParent := map[string]bool{}

	for _, t := range tasks {
		d.checkVocabulary(t, "status", t.Status, cfg.Statuses)
		d.checkVocabulary(t, "type", t.Type, cfg.Types)
		d.checkVocabulary(t, "priority", t.Priority, cfg.Priorities)
		for _, key := range sortedKeysOfStrings(t.Ext) {
			if containsString(cfg.Extensions, key) {
				continue
			}
			d.problem(t.ID, "undeclared_extension_key", undeclaredKeyMessage(key, cfg.Extensions))
		}
		for _, dep := range t.Dependencies {
			if _, ok := byID[dep]; !ok {
				d.problem(t.ID, "dependency_not_found",
					fmt.Sprintf("dependency %s does not exist", dep))
			}
		}
		d.checkCycle(t, byID, reportedDependency, "dependency_cycle", "dependency cycle",
			func(x *model.Task) []string { return x.Dependencies })
		d.checkCycle(t, byID, reportedParent, "parent_cycle", "parent cycle",
			func(x *model.Task) []string {
				if x.Parent == "" {
					return nil
				}
				return []string{x.Parent}
			})
		d.checkLease(t)
	}
	return d.checkDuplicateCriterionKeys()
}

// checkVocabulary is the row about a value the board no longer configures.
// An empty value is no value at all and is never checked against anything.
func (d *doctor) checkVocabulary(t *model.Task, field, value string, configured []string) {
	if value == "" || containsString(configured, value) {
		return
	}
	noun := "statuses"
	switch field {
	case "type":
		noun = "types"
	case "priority":
		noun = "priorities"
	}
	d.problem(t.ID, "value_not_configured", notConfiguredMessage(field, value, noun, configured))
}

// notConfiguredMessage and undeclaredKeyMessage are the two messages that
// quote a list the board may have emptied. `types`, `priorities` and
// `extensions` can all be left with nothing in them
// (docs/spec/cmd/config.md), and quoting an empty list would print a pair
// of empty quotes where a value should be, which says nothing and reads
// like a bug. The message says there is none instead
// (docs/spec/cmd/doctor.md#el-code-y-el-mensaje-de-cada-comprobación).
func notConfiguredMessage(field, value, noun string, configured []string) string {
	if len(configured) == 0 {
		return fmt.Sprintf("%s %q is not one of the configured %s, and the board configures none",
			field, value, noun)
	}
	return fmt.Sprintf("%s %q is not one of the configured %s %q",
		field, value, noun, strings.Join(configured, ", "))
}

func undeclaredKeyMessage(key string, declared []string) string {
	if len(declared) == 0 {
		return fmt.Sprintf("ext key %q is not declared, and the board declares none", key)
	}
	return fmt.Sprintf("ext key %q is not declared, declared keys are %q",
		key, strings.Join(declared, ", "))
}

// checkCycle reports one finding per cycle and not one per task in it: the
// second task of a cycle says nothing the first did not.
func (d *doctor) checkCycle(t *model.Task, byID map[string]*model.Task,
	reported map[string]bool, code, noun string, edges func(*model.Task) []string) {
	if reported[t.ID] {
		return
	}
	for _, next := range edges(t) {
		path := reaches(byID, next, t.ID, edges)
		if path == nil {
			continue
		}
		whole := append([]string{t.ID}, path...)
		for _, id := range whole {
			reported[id] = true
		}
		d.problem(t.ID, code, fmt.Sprintf("%s is part of a %s: %s",
			t.ID, noun, strings.Join(whole, " -> ")))
		return
	}
}

// checkLease is the invariant of docs/spec/lease.md#el-vaciado: the two
// fields only have a value on a task that is at once in the active status
// and assigned, and either one alone is as broken as both on a task that
// is neither.
//
// It repairs in one direction only, and that is why --fix may do it on its
// own: the two lease fields are what is left over, and the status and the
// people assigned are the data.
func (d *doctor) checkLease(t *model.Task) {
	hasExpiry := !t.LeaseExpiresAt.IsZero()
	hasHolder := t.LeaseHolder != ""
	if !hasExpiry && !hasHolder {
		return
	}
	holds := t.Status == d.b.Config.ActiveStatus && len(t.Assignees) > 0 && !t.Archived
	if holds && hasExpiry && hasHolder {
		return
	}
	d.repairs.ClearLease = append(d.repairs.ClearLease, t.ID)
	d.fixableProblem(t.ID, "lease_invariant",
		"has a lease but is not both active and assigned",
		fmt.Sprintf(
			"%s had a lease but was not both active and assigned; cleared leaseExpiresAt and leaseHolder",
			t.ID))
}

// checkHighestID is the counter of
// docs/spec/modelo-de-datos/identificadores.md falling behind the tasks
// that exist. It can only have one reading, so --fix repairs it alone.
func (d *doctor) checkHighestID(tasks []*model.Task) error {
	recorded, err := d.b.Tasks.LastAllocated()
	if err != nil {
		return err
	}
	highest := 0
	for _, t := range tasks {
		if n := taskNumber(t.ID); n > highest {
			highest = n
		}
	}
	if highest <= recorded {
		return nil
	}
	prefix := d.b.Config.TaskPrefix
	// A counter at zero is not an identifier: the board has never handed
	// one out, so there is no MYP-0 to name and the message says that
	// instead of inventing one
	// (docs/spec/cmd/doctor.md#el-code-y-el-mensaje-de-cada-comprobación).
	behind := fmt.Sprintf("the highest recorded id was %s-%d and tasks go up to %s-%d",
		prefix, recorded, prefix, highest)
	if recorded == 0 {
		behind = fmt.Sprintf("no id is recorded as handed out and tasks go up to %s-%d",
			prefix, highest)
	}
	d.repairs.HighestID = highest
	d.fixableProblem("", "highest_id_behind", behind,
		fmt.Sprintf("%s; recorded %s-%d", behind, prefix, highest))
	return nil
}

// checkMarker is the two rows about the <id>.id marker, which look alike
// and repair the opposite way. A missing one has a single reading, because
// the identifier of the database is the real one and the marker is its copy
// in the file system, so --fix writes it. One that names another identifier
// has two readings and loses something either way, so it is decided by hand
// (docs/spec/cmd/doctor.md#qué-comprueba).
// The identifier the two are compared against is the one the database
// holds, read here on purpose: the one of the location may have been taken
// off the marker itself, which would make the mismatch invisible by
// comparing the marker with a copy of itself.
func (d *doctor) checkMarker() error {
	dir := d.b.Location.Dir
	id, err := board.ReadIdentity(d.b.Store)
	if err != nil {
		return err
	}
	found := board.MarkerID(dir)
	switch {
	case found == id:
	case found == "":
		d.marker = id
		d.fixableProblem("", "marker_missing", fmt.Sprintf(
			"board directory has no <id>.id marker, the database says id is %q", id),
			fmt.Sprintf("wrote the <id>.id marker, the database says id is %q", id))
	default:
		d.problem("", "marker_id_mismatch", fmt.Sprintf(
			"marker file names id %q, the database says id is %q", found, id))
	}
	return nil
}

// checkExtraRoots is the warning of a further boards root the machine
// declares and this process cannot read: nothing is broken, and a board
// living there would not be found by its identifier while that is true.
func (d *doctor) checkExtraRoots() {
	for _, root := range d.env.Machine.ExtraRoots {
		if _, err := os.ReadDir(root); err == nil {
			continue
		}
		d.warning("", "extra_root_unreadable", fmt.Sprintf(
			"extra board root %q cannot be read (skipped when looking up boards by id)", root))
	}
}

// checkFilesystem is the warning of a board directory where SQLite's WAL
// mode cannot be relied on. It is a warning and not an error because it
// reports a risk and not damage: the board in front of it may be perfectly
// sound (docs/spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros).
func (d *doctor) checkFilesystem() {
	dir := d.b.Location.Dir
	if walprobe.Safe(dir) {
		return
	}
	d.warning("", "unsafe_wal_filesystem", fmt.Sprintf(
		"board directory %q is on a filesystem where SQLite's WAL mode is not safe "+
			"(the byte-range lock or the shared mmap probe failed)", dir))
}

// checkIgnoreFile is the one case of an exclusion file left over from
// another version control system that can be recognized with certainty:
// a .gitignore in the board directory while the configured vcs is neither
// git nor a custom one whose ignore_file is .gitignore. The opposite case,
// an orphaned ignore_file of a custom system that changed, is not
// detectable, because biso remembers no history of that key.
//
// It is not repairable with --fix for the reason docs/spec/cmd/init.md
// already fixes: biso never renames or rewrites an exclusion file of its
// own accord when vcs changes.
func (d *doctor) checkIgnoreFile() {
	const gitIgnore = ".gitignore"
	if vcs.IgnoreFile(VCSConfig(d.env.Machine)) == gitIgnore {
		return
	}
	if _, err := os.Stat(filepath.Join(d.b.Location.Dir, gitIgnore)); err != nil {
		return
	}
	d.warning("", "ignore_file_mismatch", fmt.Sprintf(
		"%s does not match the configured vcs %q "+
			"(left over from git, biso does not rewrite it automatically)",
		gitIgnore, d.env.Machine.VCS))
}

// fix applies the repairs the checks collected, and moves every error it
// repaired out of `problems` and into `fixed`.
//
// The order is fixed and it matters: the data repairs go in one
// transaction, all or nothing, and the marker is written after and apart,
// because writing a file cannot be inside a transaction of SQLite. If the
// marker fails, the data repairs stay applied, the call ends with the exit
// code of not being able to write, and the missing marker shows up again
// next time, because it is still true
// (docs/spec/cmd/doctor.md#atomicidad-de---fix-con-varias-reparaciones).
func (d *doctor) fix(dryRun bool) error {
	if !dryRun {
		if err := d.b.Repair(d.repairs); err != nil {
			return err
		}
		if d.marker != "" {
			if err := board.WriteMarker(d.b.Location.Dir, d.marker); err != nil {
				return err
			}
		}
	}
	kept := make([]Finding, 0, len(d.problems))
	for _, p := range d.problems {
		if p.repaired == nil {
			kept = append(kept, p.found)
			continue
		}
		d.result.Fixed = append(d.result.Fixed, *p.repaired)
	}
	d.result.Problems = kept
	return nil
}

// sortedKeys and sortedKeysOfStrings keep every list of this report in one
// fixed order, so that two runs over the same board print the same thing.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func sortedKeysOfStrings(m map[string]string) []string { return sortedKeys(m) }

func sortStrings(values []string) { sort.Strings(values) }
