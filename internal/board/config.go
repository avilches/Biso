package board

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"

	"biso/internal/model"
)

// This file is the board's own configuration: the nineteen keys of
// docs/spec/cmd/config.md, which live inside the board's database and
// therefore travel with it wherever the directory goes.

// The default vocabulary of a board created without the flags that set it
// (docs/spec/cmd/init.md).
var (
	DefaultStatuses   = []string{"To Do", "In Progress", "Done"}
	DefaultTypes      = []string{"task", "bug", "docs"}
	DefaultPriorities = []string{"high", "medium", "low"}
)

// DefaultLeaseMinutes is the lease duration of
// docs/spec/cmd/config.md: long enough that the harmful error, a lease that
// looks expired while someone is really working, is the rare one.
const DefaultLeaseMinutes = 240

// Config is a board's configuration, whole. Every key of
// docs/spec/cmd/config.md is here, in the order that page lists them, which
// is also the order `biso config list` prints.
type Config struct {
	ProjectName    string
	Statuses       []string
	InitialStatus  string
	ActiveStatus   string
	TerminalStatus string
	Types          []string
	Priorities     []string
	Labels         []string
	Assignees      []string
	TaskPrefix     string
	FinishStrict   bool
	LeaseMinutes   int
	Urgency        model.UrgencyCoefficients
}

// DefaultConfig is the configuration of a board created with nothing but a
// name and a prefix.
func DefaultConfig(name, prefix string) Config {
	return Config{
		ProjectName:    name,
		Statuses:       append([]string(nil), DefaultStatuses...),
		InitialStatus:  DefaultStatuses[0],
		ActiveStatus:   DefaultStatuses[1],
		TerminalStatus: DefaultStatuses[2],
		Types:          append([]string(nil), DefaultTypes...),
		Priorities:     append([]string(nil), DefaultPriorities...),
		Labels:         []string{},
		Assignees:      []string{},
		TaskPrefix:     prefix,
		FinishStrict:   false,
		LeaseMinutes:   DefaultLeaseMinutes,
		Urgency:        model.DefaultUrgencyCoefficients,
	}
}

// The names of the configuration keys, as `biso config` spells them.
const (
	KeyProjectName     = "project_name"
	KeyStatuses        = "statuses"
	KeyInitialStatus   = "initial_status"
	KeyActiveStatus    = "active_status"
	KeyTerminalStatus  = "terminal_status"
	KeyTypes           = "types"
	KeyPriorities      = "priorities"
	KeyLabels          = "labels"
	KeyAssignees       = "assignees"
	KeyTaskPrefix      = "task_prefix"
	KeyFinishStrict    = "finish_strict"
	KeyLeaseMinutes    = "lease_minutes"
	KeyUrgencyPriority = "urgency.priority"
	KeyUrgencyActive   = "urgency.active"
	KeyUrgencyBlocking = "urgency.blocking"
	KeyUrgencyBlocked  = "urgency.blocked"
	KeyUrgencyDue      = "urgency.due"
	KeyUrgencyCriteria = "urgency.criteria"
	KeyUrgencyAge      = "urgency.age"
)

// ConfigKeys are the nineteen keys, in the order of the table of
// docs/spec/cmd/config.md#las-claves.
var ConfigKeys = []string{
	KeyProjectName, KeyStatuses, KeyInitialStatus, KeyActiveStatus,
	KeyTerminalStatus, KeyTypes, KeyPriorities, KeyLabels, KeyAssignees,
	KeyTaskPrefix, KeyFinishStrict, KeyLeaseMinutes,
	KeyUrgencyPriority, KeyUrgencyActive, KeyUrgencyBlocking,
	KeyUrgencyBlocked, KeyUrgencyDue, KeyUrgencyCriteria, KeyUrgencyAge,
}

// rows turns a configuration into the rows that hold it, one per key. A list
// is written as a JSON array, so that reading it back can never depend on
// what its values contain.
func (c Config) rows() map[string]string {
	return map[string]string{
		KeyProjectName:     c.ProjectName,
		KeyStatuses:        encodeList(c.Statuses),
		KeyInitialStatus:   c.InitialStatus,
		KeyActiveStatus:    c.ActiveStatus,
		KeyTerminalStatus:  c.TerminalStatus,
		KeyTypes:           encodeList(c.Types),
		KeyPriorities:      encodeList(c.Priorities),
		KeyLabels:          encodeList(c.Labels),
		KeyAssignees:       encodeList(c.Assignees),
		KeyTaskPrefix:      c.TaskPrefix,
		KeyFinishStrict:    strconv.FormatBool(c.FinishStrict),
		KeyLeaseMinutes:    strconv.Itoa(c.LeaseMinutes),
		KeyUrgencyPriority: encodeFloat(c.Urgency.Priority),
		KeyUrgencyActive:   encodeFloat(c.Urgency.Active),
		KeyUrgencyBlocking: encodeFloat(c.Urgency.Blocking),
		KeyUrgencyBlocked:  encodeFloat(c.Urgency.Blocked),
		KeyUrgencyDue:      encodeFloat(c.Urgency.Due),
		KeyUrgencyCriteria: encodeFloat(c.Urgency.Criteria),
		KeyUrgencyAge:      encodeFloat(c.Urgency.Age),
	}
}

func encodeList(values []string) string {
	if values == nil {
		values = []string{}
	}
	b, err := json.Marshal(values)
	if err != nil {
		// A list of strings always marshals, so this cannot happen; the
		// empty array keeps the function total all the same.
		return "[]"
	}
	return string(b)
}

func decodeList(key, value string) ([]string, *model.Error) {
	var out []string
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		return nil, undecodableConfig(key, value)
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func encodeFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

// WriteConfig writes the whole configuration inside the caller's
// transaction, replacing whatever was there. It is what `biso init` does,
// and what `biso init --overwrite-config` does over a board that already
// exists.
func WriteConfig(tx *sql.Tx, c Config) error {
	for key, value := range c.rows() {
		if _, err := tx.Exec(
			`INSERT INTO board_config (key, value) VALUES (?, ?)
			 ON CONFLICT (key) DO UPDATE SET value = excluded.value`,
			key, value,
		); err != nil {
			return err
		}
	}
	return nil
}

// ReadConfig reads the whole configuration back.
//
// A key the program does not know, or a value it cannot interpret, is exit
// code 3 in the direction of docs/spec/codigos-de-salida.md#el-código-3-cubre-dos-direcciones
// that covers stored data: never something read past in silence.
func ReadConfig(q queryer) (Config, error) {
	rows, err := q.Query(`SELECT key, value FROM board_config`)
	if err != nil {
		return Config{}, err
	}
	defer rows.Close()

	c := Config{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return Config{}, err
		}
		if err := c.set(key, value); err != nil {
			return Config{}, err
		}
	}
	if err := rows.Err(); err != nil {
		return Config{}, err
	}
	return c, nil
}

// queryer is the little that ReadConfig needs, so that it can read through
// the store's handle or through an open transaction.
type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

func (c *Config) set(key, value string) *model.Error {
	var err *model.Error
	switch key {
	case KeyProjectName:
		c.ProjectName = value
	case KeyStatuses:
		c.Statuses, err = decodeList(key, value)
	case KeyInitialStatus:
		c.InitialStatus = value
	case KeyActiveStatus:
		c.ActiveStatus = value
	case KeyTerminalStatus:
		c.TerminalStatus = value
	case KeyTypes:
		c.Types, err = decodeList(key, value)
	case KeyPriorities:
		c.Priorities, err = decodeList(key, value)
	case KeyLabels:
		c.Labels, err = decodeList(key, value)
	case KeyAssignees:
		c.Assignees, err = decodeList(key, value)
	case KeyTaskPrefix:
		c.TaskPrefix = value
	case KeyFinishStrict:
		b, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return undecodableConfig(key, value)
		}
		c.FinishStrict = b
	case KeyLeaseMinutes:
		n, parseErr := strconv.Atoi(value)
		if parseErr != nil {
			return undecodableConfig(key, value)
		}
		c.LeaseMinutes = n
	case KeyUrgencyPriority:
		c.Urgency.Priority, err = decodeFloat(key, value)
	case KeyUrgencyActive:
		c.Urgency.Active, err = decodeFloat(key, value)
	case KeyUrgencyBlocking:
		c.Urgency.Blocking, err = decodeFloat(key, value)
	case KeyUrgencyBlocked:
		c.Urgency.Blocked, err = decodeFloat(key, value)
	case KeyUrgencyDue:
		c.Urgency.Due, err = decodeFloat(key, value)
	case KeyUrgencyCriteria:
		c.Urgency.Criteria, err = decodeFloat(key, value)
	case KeyUrgencyAge:
		c.Urgency.Age, err = decodeFloat(key, value)
	default:
		return unknownStoredConfigKey(key)
	}
	return err
}

func decodeFloat(key, value string) (float64, *model.Error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, undecodableConfig(key, value)
	}
	return f, nil
}

func undecodableConfig(key, value string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message:  fmt.Sprintf("the board's %s cannot be read: %q", key, value),
		Field:    key,
		Given:    value,
	}
}

func unknownStoredConfigKey(key string) *model.Error {
	return &model.Error{
		ExitCode: 3,
		Code:     "bad_config_value",
		Message:  fmt.Sprintf("the board's configuration holds a key biso does not know: %q", key),
		Field:    key,
		Given:    key,
	}
}
