package ops

import (
	"errors"
	"fmt"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the second half of `biso config set`: the checks that answer
// exit code 6, the ones that need to know what the board holds and not just
// what the key accepts (docs/spec/cmd/config.md#comportamiento-caso-a-caso).
//
// They all say the same thing in different words: a value that is in use is
// never removed in silence, and the three status roles never stop being
// three distinct statuses of the list.

// minimumStatuses is what the table of keys fixes for `statuses`, and the
// only list of the configuration that has a minimum at all.
const minimumStatuses = 3

func checkConfigConsistency(b *board.Board, cfg board.Config, key string) error {
	switch key {
	case board.KeyStatuses:
		return checkStatuses(b, cfg)
	case board.KeyInitialStatus, board.KeyActiveStatus, board.KeyTerminalStatus:
		return checkDistinctRoles(cfg, key)
	case board.KeyTypes:
		return checkVocabularyInUse(b, key, "type", cfg.Types, func(t *model.Task) []string {
			return oneOrNone(t.Type)
		})
	case board.KeyPriorities:
		return checkVocabularyInUse(b, key, "priority", cfg.Priorities, func(t *model.Task) []string {
			return oneOrNone(t.Priority)
		})
	case board.KeyExtensions:
		return checkVocabularyInUse(b, key, "ext key", cfg.Extensions, func(t *model.Task) []string {
			keys := make([]string, 0, len(t.Ext))
			for k := range t.Ext {
				keys = append(keys, k)
			}
			return keys
		})
	case board.KeyTaskPrefix:
		return checkPrefixChange(b, cfg)
	}
	return nil
}

// checkStatuses is the three refusals of changing the list of statuses: too
// few, one that a task still uses, and one that is a role.
func checkStatuses(b *board.Board, cfg board.Config) error {
	if len(cfg.Statuses) < minimumStatuses {
		return inconsistent(board.KeyStatuses, fmt.Sprintf(
			"statuses would have %d, and at least %d are required",
			len(cfg.Statuses), minimumStatuses))
	}
	for _, role := range []struct{ key, value string }{
		{board.KeyInitialStatus, cfg.InitialStatus},
		{board.KeyActiveStatus, cfg.ActiveStatus},
		{board.KeyTerminalStatus, cfg.TerminalStatus},
	} {
		if containsString(cfg.Statuses, role.value) {
			continue
		}
		// The three roles are explicit values and changing `statuses`
		// never moves them, so removing the status one of them names is
		// refused instead of quietly pointing the role somewhere else
		// (docs/spec/cmd/config.md).
		return inconsistent(board.KeyStatuses, fmt.Sprintf(
			"%q is the %s, and it is not in the new statuses", role.value, role.key),
			fmt.Sprintf("change %s first, with `biso config set %s <status>`",
				role.key, role.key))
	}
	return checkVocabularyInUse(b, board.KeyStatuses, "status", cfg.Statuses,
		func(t *model.Task) []string { return oneOrNone(t.Status) })
}

// checkDistinctRoles is the refusal of giving a role the status another
// role already has. The three name three different statuses, which is what
// keeps `biso start` and `biso finish` from meaning the same thing.
func checkDistinctRoles(cfg board.Config, key string) error {
	roles := []struct{ key, value string }{
		{board.KeyInitialStatus, cfg.InitialStatus},
		{board.KeyActiveStatus, cfg.ActiveStatus},
		{board.KeyTerminalStatus, cfg.TerminalStatus},
	}
	for i := range roles {
		for j := i + 1; j < len(roles); j++ {
			if roles[i].value != roles[j].value {
				continue
			}
			return inconsistent(key, fmt.Sprintf(
				"%s and %s would both be %q, and the three roles have to be distinct",
				roles[i].key, roles[j].key, roles[i].value))
		}
	}
	return nil
}

// checkVocabularyInUse refuses to drop a value that some task still
// carries, saying how many tasks use it and which ones. It is the same
// refusal for the four closed vocabularies of the configuration, because it
// is the same fact: a task would be left naming something the board no
// longer declares, which is exactly what `biso doctor` reports as an error.
func checkVocabularyInUse(b *board.Board, key, noun string, allowed []string,
	used func(*model.Task) []string) error {
	tasks, _, err := b.Tasks.All()
	if err != nil {
		return err
	}
	// The order is the order of the board, so the message names the same
	// tasks in the same order whatever the map iteration did.
	var missing []string
	byValue := map[string][]string{}
	for _, t := range tasks {
		for _, value := range used(t) {
			if containsString(allowed, value) {
				continue
			}
			if _, seen := byValue[value]; !seen {
				missing = append(missing, value)
			}
			byValue[value] = append(byValue[value], t.ID)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	value := missing[0]
	ids := byValue[value]
	return inconsistent(key, fmt.Sprintf("%s %q is used by %s: %s",
		noun, value, countOfTasks(len(ids)), strings.Join(ids, ", ")))
}

// checkPrefixChange is the one key whose value is baked into data that
// already exists: every identifier ever handed out carries it, so it can
// only change while the board has no task at all.
func checkPrefixChange(b *board.Board, cfg board.Config) error {
	if cfg.TaskPrefix == b.Config.TaskPrefix {
		return nil
	}
	counts, err := b.Counts()
	if err != nil {
		return err
	}
	if counts.NotArchived+counts.Archived == 0 && counts.HighestEverAssigned == 0 {
		return nil
	}
	return inconsistent(board.KeyTaskPrefix,
		"task_prefix cannot change on a board that has already handed out identifiers",
		"export the board, rewrite the identifiers and import them into a new board")
}

func countOfTasks(n int) string {
	if n == 1 {
		return "1 task"
	}
	return fmt.Sprintf("%d tasks", n)
}

func oneOrNone(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}

// asModelError is errors.As for the one error type of the specification,
// written once so that a caller that has to add a field to an error born
// several layers down does not repeat the dance.
func asModelError(err error, into **model.Error) bool {
	return errors.As(err, into)
}
