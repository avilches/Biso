package cli

import (
	"fmt"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is `biso config` (docs/spec/cmd/config.md): the three
// subcommands read off the positional arguments, and the three shapes of
// output they answer with.

// runConfig dispatches the call and prints what it answers.
func runConfig(s Streams, p *Parsed, env ops.Env) int {
	params, err := configParams(p)
	if err != nil {
		// --json is not accepted outside `config list`, and the refusal
		// of the flag is plain text on stderr even here, because it is
		// --json itself that is the invalid part of the call
		// (docs/spec/cmd/config.md#parámetros).
		return fail(s, p.Has("json") && params.Action == ops.ConfigList, err, warningsOf(p))
	}
	asJSON := p.Has("json")
	result, err := ops.Config(env, params)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)

	switch params.Action {
	case ops.ConfigList:
		if asJSON {
			writeEnvelope(s, env, "config", configData(result.View))
			return 0
		}
		for _, entry := range result.Entries {
			fmt.Fprintln(s.Stdout, configLine(entry))
		}
	case ops.ConfigGet:
		fmt.Fprintln(s.Stdout, result.Value)
	default:
		// A correct `set` writes nothing on stdout: its whole answer is
		// the note (docs/spec/cmd/config.md#comportamiento-caso-a-caso).
		printNote(s, p, result.Note)
	}
	return 0
}

// configLine is one line of `biso config list` and of the note of a `set`:
// the key, the equals sign, and the value behind it. A list with nothing in
// it prints the key and the equals sign and nothing after, which is what
// `biso config get` of that same key prints too.
func configLine(entry ops.ConfigEntry) string {
	if entry.Value == "" {
		return entry.Key + " ="
	}
	return entry.Key + " = " + entry.Value
}

// configParams reads the subcommand and its arguments off the positional
// arguments of the call. The three of them are positional and not flags, so
// the generic table of commands.go carries none of them and this is where
// they are judged.
func configParams(p *Parsed) (ops.ConfigParams, error) {
	if len(p.Positionals) == 0 {
		return ops.ConfigParams{}, &model.Error{
			ExitCode: 2,
			Code:     "unknown_command",
			Message:  "biso config needs one of get, set or list",
			Hints:    []string{"biso config list"},
			Field:    "subcommand",
		}
	}
	params := ops.ConfigParams{DryRun: p.Has("dry-run")}
	switch p.Positionals[0] {
	case "list":
		params.Action = ops.ConfigList
		if len(p.Positionals) > 1 {
			return params, errUnexpectedArgument(p.Positionals[1])
		}
	case "get":
		params.Action = ops.ConfigGet
		if len(p.Positionals) < 2 {
			return params, errMissingConfigArgument("get", "a key")
		}
		if len(p.Positionals) > 2 {
			return params, errUnexpectedArgument(p.Positionals[2])
		}
		params.Key = p.Positionals[1]
	case "set":
		params.Action = ops.ConfigSet
		if len(p.Positionals) < 2 {
			return params, errMissingConfigArgument("set", "a key")
		}
		if len(p.Positionals) < 3 {
			return params, errMissingConfigArgument("set", "a value")
		}
		if len(p.Positionals) > 3 {
			return params, errUnexpectedArgument(p.Positionals[3])
		}
		params.Key, params.Value = p.Positionals[1], p.Positionals[2]
		// The comma rule belongs to the command line, so the value
		// arrives at internal/ops both whole and already split, and that
		// package picks the half its key takes
		// (docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas).
		params.Values = splitList(params.Value)
	default:
		return params, &model.Error{
			ExitCode: 2,
			Code:     "unknown_command",
			Message:  fmt.Sprintf("unknown config subcommand: %q", p.Positionals[0]),
			Hints:    []string{"biso config takes get, set or list"},
			Field:    "subcommand",
			Given:    p.Positionals[0],
		}
	}
	if p.Has("json") && params.Action != ops.ConfigList {
		return params, &model.Error{
			ExitCode: 2,
			Code:     "read_only_flag",
			Message:  "--json only applies to config list",
			Field:    "json",
		}
	}
	if p.Has("dry-run") && params.Action != ops.ConfigSet {
		// `config get` and `config list` write nothing, so --dry-run has
		// nothing to do there and is never ignored in silence
		// (docs/spec/cmd/flags-globales.md).
		return params, errReadOnlyFlag(lookupLong(nil, "dry-run"),
			"--dry-run does not apply to a read-only command")
	}
	return params, nil
}

func errMissingConfigArgument(subcommand, what string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "missing_value",
		Message:  "biso config " + subcommand + " needs " + what,
		Hints:    []string{"biso config " + subcommand + " lease_minutes"},
		Field:    "subcommand",
		Given:    subcommand,
	}
}

// The config envelope of docs/spec/cmd/config.md#el-esquema-json: the whole
// configuration, with each value in its own type and the seven
// coefficients under one key.
type configEnvelopeData struct {
	Config configEnvelopeConfig `json:"config"`
}

type configEnvelopeConfig struct {
	ProjectName    string                `json:"project_name"`
	Statuses       []string              `json:"statuses"`
	InitialStatus  string                `json:"initial_status"`
	ActiveStatus   string                `json:"active_status"`
	TerminalStatus string                `json:"terminal_status"`
	Types          []string              `json:"types"`
	Priorities     []string              `json:"priorities"`
	Labels         []string              `json:"labels"`
	Assignees      []string              `json:"assignees"`
	Extensions     []string              `json:"extensions"`
	TaskPrefix     string                `json:"task_prefix"`
	FinishStrict   bool                  `json:"finish_strict"`
	LeaseMinutes   int                   `json:"lease_minutes"`
	Urgency        configEnvelopeUrgency `json:"urgency"`
}

// configEnvelopeUrgency carries the seven coefficients with the one decimal
// digit of docs/spec/contrato-json.md#números-fechas-y-ausencias, so that a
// whole one is 6.0 and never 6.
type configEnvelopeUrgency struct {
	Priority urgency `json:"priority"`
	Active   urgency `json:"active"`
	Blocking urgency `json:"blocking"`
	Blocked  urgency `json:"blocked"`
	Due      urgency `json:"due"`
	Criteria urgency `json:"criteria"`
	Age      urgency `json:"age"`
}

func configData(v ops.ConfigView) configEnvelopeData {
	return configEnvelopeData{Config: configEnvelopeConfig{
		ProjectName:    v.ProjectName,
		Statuses:       list(v.Statuses),
		InitialStatus:  v.InitialStatus,
		ActiveStatus:   v.ActiveStatus,
		TerminalStatus: v.TerminalStatus,
		Types:          list(v.Types),
		Priorities:     list(v.Priorities),
		Labels:         list(v.Labels),
		Assignees:      list(v.Assignees),
		Extensions:     list(v.Extensions),
		TaskPrefix:     v.TaskPrefix,
		FinishStrict:   v.FinishStrict,
		LeaseMinutes:   v.LeaseMinutes,
		Urgency: configEnvelopeUrgency{
			Priority: urgency(v.Urgency.Priority),
			Active:   urgency(v.Urgency.Active),
			Blocking: urgency(v.Urgency.Blocking),
			Blocked:  urgency(v.Urgency.Blocked),
			Due:      urgency(v.Urgency.Due),
			Criteria: urgency(v.Urgency.Criteria),
			Age:      urgency(v.Urgency.Age),
		},
	}}
}
