package ops

import (
	"fmt"
	"strconv"
	"strings"

	"biso/internal/board"
	"biso/internal/match"
	"biso/internal/model"
)

// This file is `biso config` (docs/spec/cmd/config.md): the twenty keys of
// a board's own configuration, read one by one, written one by one, or
// listed whole.
//
// Nothing here ever touches a task, and nothing here ever touches the file
// system: renaming a board is a write to its database and nothing else,
// because the folder's name is decorative and nobody resolves by it
// (docs/spec/resolucion-del-tablero.md).

// ConfigAction is which of the three subcommands the call named.
type ConfigAction int

const (
	ConfigGet ConfigAction = iota
	ConfigSet
	ConfigList
)

// ConfigParams is one `biso config` call.
type ConfigParams struct {
	Action ConfigAction
	// Key is the key of `get` and of `set`.
	Key string
	// Value is the value of `set` exactly as it was typed, and Values the
	// same value split on the commas the command line reads it by. The
	// two travel because a scalar key takes the first and a list key the
	// second, and the comma rule belongs to the command line and not
	// here.
	Value  string
	Values []string
	DryRun bool
}

// ConfigEntry is one key with the value `get` and `list` print for it: the
// values of a list joined by commas, and the empty string for a list with
// nothing in it.
type ConfigEntry struct {
	Key   string
	Value string
}

// ConfigResult is what one call answers.
type ConfigResult struct {
	// Entries are the twenty keys in the order of the table of
	// docs/spec/cmd/config.md#las-claves, filled by `list`.
	Entries []ConfigEntry
	// Value is what `get` answers.
	Value string
	// View is the configuration as the JSON envelope of `list` carries
	// it, with each value in its own type.
	View ConfigView
	// Note is the line `set` writes on stderr, already in the tense the
	// call earned: "key = value", or "key would be set to value
	// (--dry-run)".
	Note string
}

// ConfigView is a board's configuration with every value in its own type,
// which is what the envelope of `biso config list --json` needs and what a
// list of strings could not carry.
type ConfigView struct {
	ProjectName    string
	Statuses       []string
	InitialStatus  string
	ActiveStatus   string
	TerminalStatus string
	Types          []string
	Priorities     []string
	Labels         []string
	Assignees      []string
	Extensions     []string
	TaskPrefix     string
	FinishStrict   bool
	LeaseMinutes   int
	Urgency        model.UrgencyCoefficients
}

// Config reads or changes the board configuration.
func Config(env Env, p ConfigParams) (*ConfigResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return ConfigOn(b, env, p)
}

// ConfigOn is Config over a board that is already open.
func ConfigOn(b *board.Board, env Env, p ConfigParams) (*ConfigResult, error) {
	switch p.Action {
	case ConfigList:
		return &ConfigResult{Entries: configEntries(b.Config), View: viewOf(b.Config)}, nil
	case ConfigGet:
		if err := knownConfigKey(p.Key); err != nil {
			return nil, err
		}
		return &ConfigResult{Value: configValue(b.Config, p.Key)}, nil
	}
	return setConfig(b, p)
}

// setConfig writes one key. The value is checked in the two steps the case
// table of docs/spec/cmd/config.md separates: first whether it is a value
// of that key's type and domain at all, which is exit code 3, and then
// whether the board would still hold together with it, which is exit code
// 6 and needs to read the tasks.
func setConfig(b *board.Board, p ConfigParams) (*ConfigResult, error) {
	if err := knownConfigKey(p.Key); err != nil {
		return nil, err
	}
	cfg := cloneConfig(b.Config)
	if err := writeConfigKey(&cfg, p); err != nil {
		return nil, err
	}
	if err := checkConfigConsistency(b, cfg, p.Key); err != nil {
		return nil, err
	}

	value := configValue(cfg, p.Key)
	if p.DryRun {
		// `config set` has no status line to mark as hypothetical, so the
		// preview is the same note in the conditional
		// (docs/spec/cmd/config.md).
		return &ConfigResult{Note: previewNote(p.Key, value)}, nil
	}
	if err := b.Rewrite(cfg); err != nil {
		return nil, err
	}
	return &ConfigResult{Note: ConfigLine(p.Key, value)}, nil
}

// ConfigLine is how a key and its value are written wherever the two travel
// together: one line of `biso config list`, the answer of `biso config get`
// and the note of a `biso config set`. A list with nothing in it is the key
// and the equals sign and nothing after, with no space left dangling
// (docs/spec/cmd/config.md#salida).
func ConfigLine(key, value string) string {
	if value == "" {
		return key + " ="
	}
	return key + " = " + value
}

// previewNote is that same note in the conditional. The empty list has its
// own wording, because "would be set to " with nothing behind it would end
// the sentence mid air; what a vocabulary emptied on purpose does is be
// emptied, and the note says that (docs/spec/cmd/config.md).
func previewNote(key, value string) string {
	if value == "" {
		return key + " would be emptied (--dry-run)"
	}
	return fmt.Sprintf("%s would be set to %s (--dry-run)", key, value)
}

// knownConfigKey is the error 4 of a key that does not exist, with up to
// three of the closest ones
// (docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas).
func knownConfigKey(key string) *model.Error {
	for _, k := range board.ConfigKeys {
		if k == key {
			return nil
		}
	}
	e := &model.Error{
		ExitCode: 4,
		Code:     "unknown_config_key",
		Message:  fmt.Sprintf("no such key: %q", key),
		Field:    "key",
		Given:    key,
	}
	if closest := match.Suggest(key, board.ConfigKeys, configSuggestions); len(closest) > 0 {
		e.Hints = []string{"did you mean: " + strings.Join(closest, ", ") + "?"}
	}
	return e
}

// configSuggestions is the N of that algorithm for this one of its places.
const configSuggestions = 3

// writeConfigKey puts the value of the call into the configuration, with
// the type and the domain of its key. Everything it refuses is exit code 3:
// the value arrived well formed and the board does not accept it, which is
// what that code covers in the direction of the input
// (docs/spec/codigos-de-salida.md#el-código-3-cubre-dos-direcciones).
func writeConfigKey(cfg *board.Config, p ConfigParams) *model.Error {
	switch p.Key {
	case board.KeyProjectName:
		if p.Value == "" {
			return emptyConfigValue(p.Key)
		}
		if err := ValidateSlug(p.Value); err != nil {
			return err
		}
		// task_prefix was derived once, when the board was created, and
		// renaming never recalculates it: a name that would leave no
		// letter to derive one from is not an error here.
		cfg.ProjectName = p.Value
	case board.KeyStatuses:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Statuses = values
	case board.KeyInitialStatus:
		return configStatusRole(p, cfg, &cfg.InitialStatus)
	case board.KeyActiveStatus:
		return configStatusRole(p, cfg, &cfg.ActiveStatus)
	case board.KeyTerminalStatus:
		return configStatusRole(p, cfg, &cfg.TerminalStatus)
	case board.KeyTypes:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Types = values
	case board.KeyPriorities:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Priorities = values
	case board.KeyLabels:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Labels = values
	case board.KeyAssignees:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Assignees = values
	case board.KeyExtensions:
		values, err := configList(p)
		if err != nil {
			return err
		}
		cfg.Extensions = values
	case board.KeyTaskPrefix:
		if p.Value == "" {
			return emptyConfigValue(p.Key)
		}
		// The alphabet of a prefix is letters only, and here that is a
		// value the board does not accept and not a malformed command
		// line, so it is exit code 3 and not the 2 of the --prefix flag
		// of `biso init` (docs/spec/cmd/config.md#códigos-de-salida).
		if ValidatePrefix(p.Value) != nil {
			return badConfigValue(p.Key, p.Value, "a task prefix is letters only")
		}
		cfg.TaskPrefix = strings.ToUpper(p.Value)
	case board.KeyFinishStrict:
		b, err := strconv.ParseBool(p.Value)
		if err != nil {
			return badConfigValue(p.Key, p.Value, "finish_strict is true or false")
		}
		cfg.FinishStrict = b
	case board.KeyLeaseMinutes:
		n, err := strconv.Atoi(p.Value)
		if err != nil || n <= 0 {
			return badConfigValue(p.Key, p.Value,
				"lease_minutes is a whole number of minutes, greater than zero")
		}
		cfg.LeaseMinutes = n
	default:
		return writeUrgencyKey(cfg, p)
	}
	return nil
}

// writeUrgencyKey is the seven coefficients, which are the rest of the
// table and are all one type (docs/spec/modelo-de-datos/urgencia.md).
func writeUrgencyKey(cfg *board.Config, p ConfigParams) *model.Error {
	into := map[string]*float64{
		board.KeyUrgencyPriority: &cfg.Urgency.Priority,
		board.KeyUrgencyActive:   &cfg.Urgency.Active,
		board.KeyUrgencyBlocking: &cfg.Urgency.Blocking,
		board.KeyUrgencyBlocked:  &cfg.Urgency.Blocked,
		board.KeyUrgencyDue:      &cfg.Urgency.Due,
		board.KeyUrgencyCriteria: &cfg.Urgency.Criteria,
		board.KeyUrgencyAge:      &cfg.Urgency.Age,
	}[p.Key]
	if into == nil {
		// knownConfigKey already ran, so this cannot happen; answering
		// keeps the function total.
		return badConfigValue(p.Key, p.Value, "no such key")
	}
	f, err := strconv.ParseFloat(p.Value, 64)
	if err != nil {
		return badConfigValue(p.Key, p.Value, p.Key+" is a decimal number")
	}
	*into = f
	return nil
}

// configStatusRole writes one of the three roles. Its value has to name one
// of the board's statuses, matched with the same algorithm every other
// status value of the program goes through, so that the same text is worth
// the same here as behind --status
// (docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos).
func configStatusRole(p ConfigParams, cfg *board.Config, into *string) *model.Error {
	status, err := match.Match(match.Status, p.Value, cfg.Statuses)
	if err != nil {
		var e *model.Error
		if !asModelError(err, &e) {
			return &model.Error{ExitCode: 3, Code: "bad_config_value", Message: err.Error()}
		}
		e.Field = p.Key
		return e
	}
	*into = status
	return nil
}

// configList reads the value of a key of type list: the values the command
// line split on commas, with the one empty value meaning the empty list,
// which is how a vocabulary is emptied
// (docs/spec/cmd/config.md#comportamiento-caso-a-caso).
func configList(p ConfigParams) ([]string, *model.Error) {
	if len(p.Values) == 1 && p.Values[0] == "" {
		return []string{}, nil
	}
	out := make([]string, 0, len(p.Values))
	for _, v := range p.Values {
		if v == "" {
			return nil, badConfigValue(p.Key, p.Value,
				p.Key+" cannot hold an empty value")
		}
		out = append(out, v)
	}
	return out, nil
}

func emptyConfigValue(key string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "empty_scalar_value",
		Message:  key + " cannot be empty",
		Field:    key,
		Given:    "",
	}
}

func badConfigValue(key, given, expected string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message:  fmt.Sprintf("%s cannot be %q: %s", key, given, expected),
		Field:    key,
		Given:    given,
	}
}

// inconsistent is the exit code 6 of this command: the value is a fine
// value of its key and the board would not hold together with it
// (docs/spec/cmd/config.md#códigos-de-salida).
func inconsistent(key, message string, hints ...string) *model.Error {
	return &model.Error{
		ExitCode: 6,
		Code:     "board_inconsistent",
		Message:  message,
		Hints:    hints,
		Field:    key,
	}
}

func cloneConfig(c board.Config) board.Config {
	clone := c
	clone.Statuses = append([]string(nil), c.Statuses...)
	clone.Types = append([]string(nil), c.Types...)
	clone.Priorities = append([]string(nil), c.Priorities...)
	clone.Labels = append([]string(nil), c.Labels...)
	clone.Assignees = append([]string(nil), c.Assignees...)
	clone.Extensions = append([]string(nil), c.Extensions...)
	return clone
}

// configEntries is the twenty lines `biso config list` prints, in the order
// of the table of keys.
func configEntries(c board.Config) []ConfigEntry {
	entries := make([]ConfigEntry, 0, len(board.ConfigKeys))
	for _, key := range board.ConfigKeys {
		entries = append(entries, ConfigEntry{Key: key, Value: configValue(c, key)})
	}
	return entries
}

// configValue is how one key is written as text, which is the same in `get`,
// in `list` and in the note of a `set`.
func configValue(c board.Config, key string) string {
	switch key {
	case board.KeyProjectName:
		return c.ProjectName
	case board.KeyStatuses:
		return strings.Join(c.Statuses, ",")
	case board.KeyInitialStatus:
		return c.InitialStatus
	case board.KeyActiveStatus:
		return c.ActiveStatus
	case board.KeyTerminalStatus:
		return c.TerminalStatus
	case board.KeyTypes:
		return strings.Join(c.Types, ",")
	case board.KeyPriorities:
		return strings.Join(c.Priorities, ",")
	case board.KeyLabels:
		return strings.Join(c.Labels, ",")
	case board.KeyAssignees:
		return strings.Join(c.Assignees, ",")
	case board.KeyExtensions:
		return strings.Join(c.Extensions, ",")
	case board.KeyTaskPrefix:
		return c.TaskPrefix
	case board.KeyFinishStrict:
		return strconv.FormatBool(c.FinishStrict)
	case board.KeyLeaseMinutes:
		return strconv.Itoa(c.LeaseMinutes)
	}
	return FormatCoefficient(urgencyCoefficient(c, key))
}

func urgencyCoefficient(c board.Config, key string) float64 {
	switch key {
	case board.KeyUrgencyPriority:
		return c.Urgency.Priority
	case board.KeyUrgencyActive:
		return c.Urgency.Active
	case board.KeyUrgencyBlocking:
		return c.Urgency.Blocking
	case board.KeyUrgencyBlocked:
		return c.Urgency.Blocked
	case board.KeyUrgencyDue:
		return c.Urgency.Due
	case board.KeyUrgencyCriteria:
		return c.Urgency.Criteria
	}
	return c.Urgency.Age
}

// FormatCoefficient writes a coefficient with the one decimal digit that
// docs/spec/contrato-json.md#números-fechas-y-ausencias fixes for every
// number of the urgency, so that a whole one reads 6.0 and never 6.
func FormatCoefficient(f float64) string { return strconv.FormatFloat(f, 'f', 1, 64) }

func viewOf(c board.Config) ConfigView {
	return ConfigView{
		ProjectName:    c.ProjectName,
		Statuses:       c.Statuses,
		InitialStatus:  c.InitialStatus,
		ActiveStatus:   c.ActiveStatus,
		TerminalStatus: c.TerminalStatus,
		Types:          c.Types,
		Priorities:     c.Priorities,
		Labels:         c.Labels,
		Assignees:      c.Assignees,
		Extensions:     c.Extensions,
		TaskPrefix:     c.TaskPrefix,
		FinishStrict:   c.FinishStrict,
		LeaseMinutes:   c.LeaseMinutes,
		Urgency:        c.Urgency,
	}
}
