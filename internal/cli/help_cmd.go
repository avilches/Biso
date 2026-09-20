package cli

import (
	"fmt"
	"strings"

	"biso/internal/match"
	"biso/internal/model"
	"biso/internal/ops"
)

// This file is `biso help` (docs/spec/cmd/help.md): the one command whose
// whole answer lives in this layer, because what it prints is the text of
// this layer and nothing of a board.
//
// It takes several names in one call on purpose: a caller that wants the
// detail of three commands pays one invocation and not three.

// helpSuggestions is how many names the error of a command that does not
// exist offers, which is the N of
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-más-parecidas for
// this one of its five places.
const helpSuggestions = 3

func runHelp(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	names, wantsAll, err := helpNames(p.Positionals)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)

	if asJSON {
		writeEnvelope(s, env, "help", helpData(names))
		return 0
	}
	if len(names) == 0 {
		fmt.Fprint(s.Stdout, topLevelHelp)
		if wantsAll {
			// `all` adds the administrative commands under the
			// top-level help, with the blank line that separates every
			// other block of that text from the next
			// (docs/spec/cmd/help.md#salida-de-biso-help-all).
			fmt.Fprint(s.Stdout, "\n"+adminBlock)
		}
		return 0
	}
	// One command's help, then a blank line, then the next one's: each
	// block is exactly what `biso <cmd> --help` prints.
	blocks := make([]string, 0, len(names))
	for _, name := range names {
		blocks = append(blocks, helpOf(name))
	}
	fmt.Fprint(s.Stdout, strings.Join(blocks, "\n"))
	return 0
}

// helpNames reads the positional arguments of the call: the command names
// it asked for, and whether it asked for `all`.
//
// The validation is whole and happens here, before anything is printed: a
// name that does not exist fails the call, and not even the help of the
// names before it in the list is written
// (docs/spec/cmd/help.md#varios-comandos-en-una-sola-llamada).
func helpNames(positionals []string) ([]string, bool, error) {
	wantsAll := false
	for _, arg := range positionals {
		if arg == "all" {
			wantsAll = true
		}
	}
	if wantsAll && len(positionals) > 1 {
		// `all` asks for a list and a name asks for a text, and there is
		// no order between the two that means one thing, so the pair is
		// a usage error instead of a guess.
		return nil, false, &model.Error{
			ExitCode: 2,
			Code:     "unexpected_argument",
			Message:  `help: "all" cannot be combined with a command name`,
			Hints: []string{
				"run `biso help all` on its own, or list only command names",
			},
			Field: "command",
			Given: "all",
		}
	}
	if wantsAll {
		return nil, true, nil
	}
	for _, name := range positionals {
		if !inCatalog(name) {
			return nil, false, errNoSuchCommand(name)
		}
	}
	return positionals, false, nil
}

// errNoSuchCommand is the error 4 of a name `biso help` was asked for and
// biso does not have. It is not the `unknown_command` of the parser, which
// is exit code 2: there the call cannot be carried out at all, and here the
// call is well formed and names something that is not there, which is what
// exit code 4 means (docs/spec/codigos-de-salida.md).
func errNoSuchCommand(name string) *model.Error {
	e := &model.Error{
		ExitCode: 4,
		Code:     "no_such_command",
		Message:  fmt.Sprintf("no such command: %q", name),
		Field:    "command",
		Given:    name,
	}
	if closest := match.Suggest(name, catalogNames(), helpSuggestions); len(closest) > 0 {
		e.Hints = []string{"did you mean: " + strings.Join(closest, ", ") + "?"}
	}
	return e
}

// helpEnvelopeData is the data of the help envelope
// (docs/spec/cmd/help.md#el-esquema-json): the command list and nothing
// else. The prose help never travels in it.
type helpEnvelopeData struct {
	Commands []helpEnvelopeCommand `json:"commands"`
}

type helpEnvelopeCommand struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// helpData is the whole catalog when the call named no command, and one
// element per name, in the order asked for, when it named some. The key is
// always there and always a list, so nobody has to read the argument to
// know the shape of the answer.
func helpData(names []string) helpEnvelopeData {
	data := helpEnvelopeData{Commands: []helpEnvelopeCommand{}}
	if len(names) == 0 {
		for _, c := range commandCatalog {
			data.Commands = append(data.Commands,
				helpEnvelopeCommand{Name: c.Name, Summary: c.Summary})
		}
		return data
	}
	for _, name := range names {
		for _, c := range commandCatalog {
			if c.Name == name {
				data.Commands = append(data.Commands,
					helpEnvelopeCommand{Name: c.Name, Summary: c.Summary})
				break
			}
		}
	}
	return data
}
