package cli

import (
	"sort"
	"unicode"
	"unicode/utf8"
)

// This file is the unit of the column algorithm of
// docs/spec/cmd/ls.md#salida: the cell of a monospaced terminal, which is
// not the byte and not the code point either. A combining mark measures
// zero, an East Asian ideograph or an emoji measures two, and everything
// else measures one.
//
// Measuring in cells is not looking at the terminal: the width of a
// character is a property of Unicode, the same on every machine and in
// every window, so the output still does not depend on where the program
// runs (docs/spec/salida-y-terminal.md#interactividad-terminal-y-color).

// cells is the width of s in terminal cells.
func cells(s string) int {
	total := 0
	for _, r := range s {
		total += runeCells(r)
	}
	return total
}

// runeCells is the width of one code point, by the two Unicode tables
// docs/spec/cmd/ls.md names: the East Asian width for the wide ones and the
// category of the combining marks for the ones that measure nothing. The
// format characters (category Cf) measure nothing either, which that page
// spells out: the zero width joiner of a composed emoji is not drawn.
func runeCells(r rune) int {
	if r == 0 {
		return 0
	}
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) {
		return 0
	}
	if isWide(r) {
		return 2
	}
	return 1
}

// truncateCells cuts s to at most max cells without ever splitting a
// grapheme. The promise is the cap and never the exact length: a cut that
// would separate a letter from its combining accent, or halve a composed
// emoji, lands on the boundary before it, so the result can measure one or
// two cells less than the cap.
func truncateCells(s string, max int) string {
	width, end := 0, 0
	for i := 0; i < len(s); {
		size := graphemeLen(s[i:])
		w := cells(s[i : i+size])
		if width+w > max {
			break
		}
		width += w
		i += size
		end = i
	}
	return s[:end]
}

// cutTitle is step 1 of the format of docs/spec/cmd/ls.md#salida: the title
// is cut to a hundred cells, counting the three dots, before any column
// width is computed. So a longer title keeps ninety-seven cells of its own,
// or fewer when the cut lands on a grapheme boundary, and the string that
// gets printed never measures more than a hundred.
func cutTitle(title string) string {
	if cells(title) <= titleCells {
		return title
	}
	return truncateCells(title, titleCells-len(ellipsis)) + ellipsis
}

// titleCells is the cap the specification fixes for column 5, and ellipsis
// the three dots that say a title was cut. The dots are ASCII, so each one
// measures one cell and subtracting their length is subtracting their width.
const (
	titleCells = 100
	ellipsis   = "..."
)

// graphemeLen is the length in bytes of the grapheme cluster that starts at
// the beginning of s.
//
// It is not the whole of the Unicode text segmentation algorithm: it covers
// the two cases docs/spec/cmd/ls.md names, a letter with its combining
// accent and a composed emoji, plus the pair of regional indicators that
// make a flag. A base character takes with it the marks that are painted on
// it, the variation selectors and the emoji modifiers, and a zero width
// joiner takes with it whatever it joins.
func graphemeLen(s string) int {
	if s == "" {
		return 0
	}
	first, n := utf8.DecodeRuneInString(s)
	i := n
	if isRegionalIndicator(first) {
		if r, size := utf8.DecodeRuneInString(s[i:]); isRegionalIndicator(r) {
			i += size
		}
		return i
	}
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case isExtend(r):
			i += size
		case r == zeroWidthJoiner:
			// The joiner belongs to this cluster, and so does whatever
			// it joins: an emoji sequence is one grapheme.
			i += size
			if i < len(s) {
				_, next := utf8.DecodeRuneInString(s[i:])
				i += next
			}
		default:
			return i
		}
	}
	return i
}

const zeroWidthJoiner = '‍'

// isExtend says whether a code point attaches to the one before it: a
// combining mark, a variation selector or one of the five emoji skin tone
// modifiers.
func isExtend(r rune) bool {
	switch {
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r), unicode.Is(unicode.Mc, r):
		return true
	case r >= 0x1F3FB && r <= 0x1F3FF:
		return true
	}
	return false
}

func isRegionalIndicator(r rune) bool { return r >= 0x1F1E6 && r <= 0x1F1FF }

// isWide answers the East Asian width table: a code point of class W (wide)
// or F (fullwidth) measures two cells.
func isWide(r rune) bool {
	i := sort.Search(len(wideRanges), func(i int) bool { return wideRanges[i].hi >= r })
	return i < len(wideRanges) && wideRanges[i].lo <= r
}

type runeRange struct{ lo, hi rune }

// wideRanges are the ranges of the East Asian width table whose class is W
// or F, sorted and non overlapping. They are written here, and not taken
// from a module, because the only external modules biso may link are
// modernc.org/sqlite and golang.org/x/term (section 5 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md),
// and the standard library carries no East Asian width table.
var wideRanges = []runeRange{
	{0x1100, 0x115F}, {0x231A, 0x231B}, {0x2329, 0x232A}, {0x23E9, 0x23EC},
	{0x23F0, 0x23F0}, {0x23F3, 0x23F3}, {0x25FD, 0x25FE}, {0x2614, 0x2615},
	{0x2648, 0x2653}, {0x267F, 0x267F}, {0x2693, 0x2693}, {0x26A1, 0x26A1},
	{0x26AA, 0x26AB}, {0x26BD, 0x26BE}, {0x26C4, 0x26C5}, {0x26CE, 0x26CE},
	{0x26D4, 0x26D4}, {0x26EA, 0x26EA}, {0x26F2, 0x26F3}, {0x26F5, 0x26F5},
	{0x26FA, 0x26FA}, {0x26FD, 0x26FD}, {0x2705, 0x2705}, {0x270A, 0x270B},
	{0x2728, 0x2728}, {0x274C, 0x274C}, {0x274E, 0x274E}, {0x2753, 0x2755},
	{0x2757, 0x2757}, {0x2795, 0x2797}, {0x27B0, 0x27B0}, {0x27BF, 0x27BF},
	{0x2B1B, 0x2B1C}, {0x2B50, 0x2B50}, {0x2B55, 0x2B55},
	{0x2E80, 0x2E99}, {0x2E9B, 0x2EF3}, {0x2F00, 0x2FD5}, {0x2FF0, 0x2FFB},
	{0x3000, 0x303E}, {0x3041, 0x3096}, {0x3099, 0x30FF}, {0x3105, 0x312F},
	{0x3131, 0x318E}, {0x3190, 0x31E3}, {0x31F0, 0x321E}, {0x3220, 0x3247},
	{0x3250, 0x4DBF}, {0x4E00, 0xA48C}, {0xA490, 0xA4C6}, {0xA960, 0xA97C},
	{0xAC00, 0xD7A3}, {0xF900, 0xFAFF}, {0xFE10, 0xFE19}, {0xFE30, 0xFE52},
	{0xFE54, 0xFE66}, {0xFE68, 0xFE6B}, {0xFF01, 0xFF60}, {0xFFE0, 0xFFE6},
	{0x16FE0, 0x16FE4}, {0x16FF0, 0x16FF1}, {0x17000, 0x187F7},
	{0x18800, 0x18CD5}, {0x18D00, 0x18D08}, {0x1B000, 0x1B11E},
	{0x1B150, 0x1B152}, {0x1B164, 0x1B167}, {0x1B170, 0x1B2FB},
	{0x1F004, 0x1F004}, {0x1F0CF, 0x1F0CF}, {0x1F18E, 0x1F18E},
	{0x1F191, 0x1F19A}, {0x1F200, 0x1F202}, {0x1F210, 0x1F23B},
	{0x1F240, 0x1F248}, {0x1F250, 0x1F251}, {0x1F260, 0x1F265},
	{0x1F300, 0x1F320}, {0x1F32D, 0x1F335}, {0x1F337, 0x1F37C},
	{0x1F37E, 0x1F393}, {0x1F3A0, 0x1F3CA}, {0x1F3CF, 0x1F3D3},
	{0x1F3E0, 0x1F3F0}, {0x1F3F4, 0x1F3F4}, {0x1F3F8, 0x1F43E},
	{0x1F440, 0x1F440}, {0x1F442, 0x1F4FC}, {0x1F4FF, 0x1F53D},
	{0x1F54B, 0x1F54E}, {0x1F550, 0x1F567}, {0x1F57A, 0x1F57A},
	{0x1F595, 0x1F596}, {0x1F5A4, 0x1F5A4}, {0x1F5FB, 0x1F64F},
	{0x1F680, 0x1F6C5}, {0x1F6CC, 0x1F6CC}, {0x1F6D0, 0x1F6D2},
	{0x1F6D5, 0x1F6D7}, {0x1F6EB, 0x1F6EC}, {0x1F6F4, 0x1F6FC},
	{0x1F7E0, 0x1F7EB}, {0x1F90C, 0x1F93A}, {0x1F93C, 0x1F945},
	{0x1F947, 0x1F978}, {0x1F97A, 0x1F9CB}, {0x1F9CD, 0x1F9FF},
	{0x1FA70, 0x1FA74}, {0x1FA78, 0x1FA7A}, {0x1FA80, 0x1FA86},
	{0x1FA90, 0x1FAA8}, {0x1FAB0, 0x1FAB6}, {0x1FAC0, 0x1FAC2},
	{0x1FAD0, 0x1FAD6}, {0x20000, 0x2FFFD}, {0x30000, 0x3FFFD},
}
