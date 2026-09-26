package source

import (
	"regexp"
	"strings"
)

// knownSectionNames are the free text sections docs/especificacion.md, "El
// mapeo de campos", recognizes, each delimited by
// "<!-- SECTION:<name>:BEGIN -->" and "<!-- SECTION:<name>:END -->".
var knownSectionNames = map[string]bool{
	"DESCRIPTION":   true,
	"PLAN":          true,
	"NOTES":         true,
	"FINAL_SUMMARY": true,
}

// knownBlockNames are the other three known sections, each delimited by
// "<!-- <name>:BEGIN -->" and "<!-- <name>:END -->" without the "SECTION:"
// prefix: acceptance criteria, Definition of Done, and comments.
var knownBlockNames = map[string]bool{
	"AC":       true,
	"DOD":      true,
	"COMMENTS": true,
}

// markerRe matches both marker styles Backlog.md uses in a task body:
// "<!-- SECTION:NAME:BEGIN -->" / "END", and "<!-- NAME:BEGIN -->" / "END".
// Group 1 is the section name, group 2 is BEGIN or END.
var markerRe = regexp.MustCompile(`<!--\s*(?:SECTION:)?([A-Za-z0-9_]+):(BEGIN|END)\s*-->`)

// splitFrontmatter separates a task or milestone file's YAML frontmatter
// from its body. It looks for a first line that is exactly "---" and a
// later line that is also exactly "---", the way Backlog.md always writes
// it; it does not try to understand the YAML in between, so an invalid
// frontmatter (an unterminated quote, a duplicate key) still splits
// correctly and is left for the YAML parser to reject.
//
// ok is false when the file does not even have that shape (no opening
// "---" line, or no closing one), which this package treats the same as a
// broken frontmatter.
func splitFrontmatter(content string) (frontmatter, body string, ok bool) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != "---" {
		return "", content, false
	}
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return "", content, false
}

// extractSection returns the text of the free text section named name
// (DESCRIPTION, PLAN, NOTES, or FINAL_SUMMARY) from body, and whether that
// section is present at all. The text is exactly what sits between the
// BEGIN and END markers, minus a single leading and a single trailing
// newline; a "##" heading or a "---" line inside it is part of the text,
// not a new section, because the section is delimited by its markers, not
// by its content.
func extractSection(body, name string) (text string, found bool) {
	begin := "<!-- SECTION:" + name + ":BEGIN -->"
	end := "<!-- SECTION:" + name + ":END -->"

	start := strings.Index(body, begin)
	if start < 0 {
		return "", false
	}
	start += len(begin)
	stop := strings.Index(body[start:], end)
	if stop < 0 {
		return "", false
	}
	inner := body[start : start+stop]
	inner = strings.TrimPrefix(inner, "\n")
	inner = strings.TrimSuffix(inner, "\n")
	return inner, true
}

// extractBlock returns the raw text between "<!-- name:BEGIN -->" and
// "<!-- name:END -->" (no "SECTION:" prefix), the marker style acceptance
// criteria, Definition of Done, and comments use, with the same leading and
// trailing newline trimmed as extractSection.
func extractBlock(body, name string) (text string, found bool) {
	begin := "<!-- " + name + ":BEGIN -->"
	end := "<!-- " + name + ":END -->"

	start := strings.Index(body, begin)
	if start < 0 {
		return "", false
	}
	start += len(begin)
	stop := strings.Index(body[start:], end)
	if stop < 0 {
		return "", false
	}
	inner := body[start : start+stop]
	inner = strings.TrimPrefix(inner, "\n")
	inner = strings.TrimSuffix(inner, "\n")
	return inner, true
}

// findUnknownSections scans body for every BEGIN marker (in either style)
// and returns the names that are not one of the sections
// docs/especificacion.md, "El mapeo de campos", recognizes. Each name is
// reported once, in the order it first appears.
func findUnknownSections(body string) []string {
	seen := make(map[string]bool)
	var unknown []string
	for _, m := range markerRe.FindAllStringSubmatch(body, -1) {
		name, kind := m[1], m[2]
		if kind != "BEGIN" {
			continue
		}
		if knownSectionNames[name] || knownBlockNames[name] {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true
		unknown = append(unknown, name)
	}
	return unknown
}

// checkboxLineRe matches a checkbox line such as "- [ ] #1 text" or
// "- [x] #12 text". It requires the "#n" marker right after the checkbox,
// the same way docs/especificacion.md, "Criterios de aceptación", defines
// one; a "#" or "[x]" appearing later in a line's text is just text.
var checkboxLineRe = regexp.MustCompile(`^- \[([ xX])\] #(\d+) (.*)$`)

// extractCheckboxes parses the checkbox lines out of a checklist section's
// raw text (as returned by extractBlock), in the format acceptance criteria
// and Definition of Done share. A line following a checkbox line that is
// not itself a checkbox line is the continuation of the previous
// checkbox's text and is joined to it with a single space; joined reports
// whether at least one such join happened, so the caller can raise the
// "multiple lines joined" finding.
func extractCheckboxes(text string) (boxes []Checkbox, joined bool) {
	if text == "" {
		return nil, false
	}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if m := checkboxLineRe.FindStringSubmatch(line); m != nil {
			checked := m[1] == "x" || m[1] == "X"
			number := parseInt(m[2])
			boxes = append(boxes, Checkbox{
				Number:  number,
				Checked: checked,
				Text:    m[3],
			})
			continue
		}
		if len(boxes) == 0 {
			// A line before the first checkbox is not valid input for
			// this format; there is nothing to attach it to.
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		last := &boxes[len(boxes)-1]
		last.Text = last.Text + " " + trimmed
		joined = true
	}
	return boxes, joined
}

// parseInt parses the digits captured by checkboxLineRe, which are always a
// valid non-negative integer by construction of the regular expression.
func parseInt(digits string) int {
	n := 0
	for _, r := range digits {
		n = n*10 + int(r-'0')
	}
	return n
}

// commentBlockSeparator is the standalone "---" line Backlog.md reserves to
// separate a comment's header from its body, and one comment block from the
// next.
const commentBlockSeparator = "---"

// extractComments parses the comments section's raw text (as returned by
// extractBlock) into individual comments, in the block format
// docs/especificacion.md, "Comentarios", describes:
//
//	author: @ann        (optional)
//	created: 2026-09-20 22:08
//	---
//	body, possibly several lines
//	---
//
// missingCreated lists the index (0-based, in the returned slice) of every
// comment whose block had no "created:" line at all, so the caller can
// raise a finding for each one.
func extractComments(text string) (comments []Comment, missingCreated []int) {
	lines := strings.Split(text, "\n")
	i := 0
	for i < len(lines) {
		if strings.TrimSpace(lines[i]) == "" {
			i++
			continue
		}

		var comment Comment
		hasCreated := false

		// Header lines: zero or one "author:", then a "created:", up to
		// the block separator that opens the body.
		for i < len(lines) && lines[i] != commentBlockSeparator {
			switch {
			case strings.HasPrefix(lines[i], "author:"):
				comment.Author = strings.TrimSpace(strings.TrimPrefix(lines[i], "author:"))
			case strings.HasPrefix(lines[i], "created:"):
				comment.CreatedAt = strings.TrimSpace(strings.TrimPrefix(lines[i], "created:"))
				hasCreated = true
			}
			i++
		}
		if i >= len(lines) {
			// No opening separator: an incomplete trailing block, nothing
			// more to parse.
			break
		}
		i++ // skip the opening "---"

		var bodyLines []string
		for i < len(lines) && lines[i] != commentBlockSeparator {
			bodyLines = append(bodyLines, lines[i])
			i++
		}
		comment.Body = strings.Join(bodyLines, "\n")
		if i < len(lines) {
			i++ // skip the closing "---"
		}

		if !hasCreated {
			missingCreated = append(missingCreated, len(comments))
		}
		comments = append(comments, comment)
	}
	return comments, missingCreated
}
