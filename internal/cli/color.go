package cli

// This file is the only thing in biso that looks at the terminal, and it
// looks at it for one reason: the color
// (docs/spec/salida-y-terminal.md#interactividad-terminal-y-color). The
// output of every command is identical byte for byte with a terminal and
// without one, save the color codes, so nothing else may ask this question.
//
// In the 1.0 there are no color codes to save: no output carries any, and
// that page says so and says why, which is that the specification fixes
// when there would be color and never what gets painted. Nothing calls
// UseColor for that reason, and not because it was forgotten: it is the
// answer already written down for whoever paints the first thing.

// ColorWhen is the value of the --color flag.
type ColorWhen string

const (
	ColorAuto   ColorWhen = "auto"
	ColorAlways ColorWhen = "always"
	ColorNever  ColorWhen = "never"
)

// UseColor answers the table of
// docs/spec/salida-y-terminal.md#interactividad-terminal-y-color, read in
// order: the first row that applies decides, which is why --color always
// wins over NO_COLOR.
//
// when is empty when the call did not write --color at all, which is the
// only case NO_COLOR gets to decide: the flag never comes without one of its
// three values, so writing it is already the choice that beats the
// environment variable, and that is what makes
// `NO_COLOR=1 biso ls --color always` print with color. The question is
// answered once per stream, because each one has its own destination.
func UseColor(when ColorWhen, noColor, isTerminal bool) bool {
	switch when {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	case ColorAuto:
		return isTerminal
	}
	if noColor {
		return false
	}
	return isTerminal
}
