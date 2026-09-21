package ops

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the configuration half of the interchange: board.json, the
// file `biso snapshot` writes next to snapshot.ndjson and `biso init --from`
// reads back (docs/spec/cmd/snapshot.md, docs/spec/cmd/init.md).
//
// Like the task half, it is one struct and therefore one list of keys for
// the two directions, so that a key that gets written is a key that gets
// read. Its shape is the `config` object of `biso config list --json`
// (docs/spec/cmd/config.md#el-esquema-json) minus nothing: `me` and
// `default_limit` are not in it because they are not keys of a board's
// configuration but of a machine's, which is why restoring somebody else's
// snapshot never inherits their identity or their default limit.

// wireConfig is board.json. The field order is the order of the table of
// docs/spec/cmd/config.md#las-claves, which is the order the keys come out
// in.
type wireConfig struct {
	ProjectName    string      `json:"project_name"`
	Statuses       []string    `json:"statuses"`
	InitialStatus  string      `json:"initial_status"`
	ActiveStatus   string      `json:"active_status"`
	TerminalStatus string      `json:"terminal_status"`
	Types          []string    `json:"types"`
	Priorities     []string    `json:"priorities"`
	Labels         []string    `json:"labels"`
	Assignees      []string    `json:"assignees"`
	TaskPrefix     string      `json:"task_prefix"`
	FinishStrict   bool        `json:"finish_strict"`
	LeaseMinutes   int         `json:"lease_minutes"`
	Urgency        wireUrgency `json:"urgency"`
}

// boardConfigKeys are the keys board.json may carry, taken from the struct
// tags themselves so that the check and the format cannot drift apart.
var boardConfigKeys = keysOf(reflect.TypeOf(wireConfig{}))

// wireUrgency is the nested object of the seven coefficients.
type wireUrgency struct {
	Priority decimal `json:"priority"`
	Active   decimal `json:"active"`
	Blocking decimal `json:"blocking"`
	Blocked  decimal `json:"blocked"`
	Due      decimal `json:"due"`
	Criteria decimal `json:"criteria"`
	Age      decimal `json:"age"`
}

// decimal is a coefficient written the way `biso config list --json` prints
// one: always with a point and at least one digit behind it, so that the
// default 6 reads 6.0 and a coefficient of 0.25 keeps its two digits.
type decimal float64

func (d decimal) MarshalJSON() ([]byte, error) {
	text := strconv.FormatFloat(float64(d), 'f', -1, 64)
	if !strings.ContainsAny(text, ".eE") {
		text += ".0"
	}
	return []byte(text), nil
}

// encodeBoardConfig writes board.json: indented, one key per line, because
// it is a file meant to be read in a diff (docs/spec/cmd/snapshot.md).
func encodeBoardConfig(cfg board.Config) ([]byte, error) {
	w := wireConfig{
		ProjectName:    cfg.ProjectName,
		Statuses:       listOrEmpty(cfg.Statuses),
		InitialStatus:  cfg.InitialStatus,
		ActiveStatus:   cfg.ActiveStatus,
		TerminalStatus: cfg.TerminalStatus,
		Types:          listOrEmpty(cfg.Types),
		Priorities:     listOrEmpty(cfg.Priorities),
		Labels:         listOrEmpty(cfg.Labels),
		Assignees:      listOrEmpty(cfg.Assignees),
		TaskPrefix:     cfg.TaskPrefix,
		FinishStrict:   cfg.FinishStrict,
		LeaseMinutes:   cfg.LeaseMinutes,
		Urgency: wireUrgency{
			Priority: decimal(cfg.Urgency.Priority),
			Active:   decimal(cfg.Urgency.Active),
			Blocking: decimal(cfg.Urgency.Blocking),
			Blocked:  decimal(cfg.Urgency.Blocked),
			Due:      decimal(cfg.Urgency.Due),
			Criteria: decimal(cfg.Urgency.Criteria),
			Age:      decimal(cfg.Urgency.Age),
		},
	}
	out, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(out, '\n'), nil
}

// decodeBoardConfig reads board.json back into a configuration.
//
// A file that is not interpretable as JSON, and one that carries a key this
// program does not know, are both the invalid_snapshot_config of
// docs/spec/cmd/init.md: never a key read past in silence, because a
// restored board that quietly lost a key would not be the board that was
// snapshotted.
func decodeBoardConfig(data []byte) (board.Config, *model.Error) {
	if err := objectKeys(data, boardConfigKeys); err != nil {
		return board.Config{}, invalidSnapshotConfig(err.Error())
	}
	var w wireConfig
	if err := json.Unmarshal(data, &w); err != nil {
		return board.Config{}, invalidSnapshotConfig(unwrapJSONError(err).Error())
	}
	cfg := board.Config{
		ProjectName:    w.ProjectName,
		Statuses:       w.Statuses,
		InitialStatus:  w.InitialStatus,
		ActiveStatus:   w.ActiveStatus,
		TerminalStatus: w.TerminalStatus,
		Types:          w.Types,
		Priorities:     w.Priorities,
		Labels:         listOrEmpty(w.Labels),
		Assignees:      listOrEmpty(w.Assignees),
		TaskPrefix:     w.TaskPrefix,
		FinishStrict:   w.FinishStrict,
		LeaseMinutes:   w.LeaseMinutes,
		Urgency: model.UrgencyCoefficients{
			Priority: float64(w.Urgency.Priority),
			Active:   float64(w.Urgency.Active),
			Blocking: float64(w.Urgency.Blocking),
			Blocked:  float64(w.Urgency.Blocked),
			Due:      float64(w.Urgency.Due),
			Criteria: float64(w.Urgency.Criteria),
			Age:      float64(w.Urgency.Age),
		},
	}
	return cfg, nil
}

// checkSnapshotConfig applies to board.json the same rules the vocabulary
// flags of `biso init` apply to what a caller types, with the same `code`
// each of those flags would answer: a snapshot with fewer than three
// statuses or a prefix that is not letters is as impossible a board as one
// asked for on the command line (docs/spec/cmd/init.md).
func checkSnapshotConfig(cfg board.Config) *model.Error {
	if len(cfg.Statuses) < 3 {
		return &model.Error{
			ExitCode: 2,
			Code:     "too_few_statuses",
			Message: fmt.Sprintf(
				"board.json declares %d statuses, and a board needs at least three",
				len(cfg.Statuses)),
			Field: "statuses",
			Given: strings.Join(cfg.Statuses, ","),
			Hints: []string{"a board needs one status for a new task, one for an active one and one for a finished one"},
		}
	}
	roles := []struct{ key, value string }{
		{"initial_status", cfg.InitialStatus},
		{"active_status", cfg.ActiveStatus},
		{"terminal_status", cfg.TerminalStatus},
	}
	for _, r := range roles {
		if !contains(cfg.Statuses, r.value) {
			return &model.Error{
				ExitCode: 2,
				Code:     "unknown_status_role",
				Message: fmt.Sprintf("board.json's %s names %q, which is not one of its statuses",
					r.key, r.value),
				Field: r.key,
				Given: r.value,
				Valid: append([]string(nil), cfg.Statuses...),
			}
		}
	}
	for i := range roles {
		for j := i + 1; j < len(roles); j++ {
			if roles[i].value == roles[j].value {
				return &model.Error{
					ExitCode: 2,
					Code:     "invalid_status_roles",
					Message: fmt.Sprintf("board.json's %s and %s both name %q, and the three roles are distinct",
						roles[i].key, roles[j].key, roles[i].value),
					Field: roles[j].key,
					Given: roles[j].value,
				}
			}
		}
	}
	if err := ValidatePrefix(cfg.TaskPrefix); err != nil {
		return err
	}
	if err := ValidateSlug(cfg.ProjectName); err != nil {
		return err
	}
	if cfg.LeaseMinutes <= 0 {
		return &model.Error{
			ExitCode: 2,
			Code:     "invalid_snapshot_config",
			Message: fmt.Sprintf("board.json's lease_minutes is %d, and it has to be greater than zero",
				cfg.LeaseMinutes),
			Field: "lease_minutes",
			Given: strconv.Itoa(cfg.LeaseMinutes),
		}
	}
	return nil
}

func invalidSnapshotConfig(reason string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "invalid_snapshot_config",
		Message:  "board.json cannot be read: " + reason,
		Field:    "from",
	}
}
