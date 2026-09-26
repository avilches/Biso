package source

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// parseTask reads a single task file's content (file is its base name, used
// only for findings) into a Task. skipped is true when the frontmatter
// could not be parsed as YAML at all, in which case the returned Task is
// the zero value and findings holds exactly one Finding with field
// "frontmatter", per docs/especificacion.md, "Qué lee del origen": the rest
// of the batch keeps being read regardless.
func parseTask(file, content string, archived bool) (task Task, findings []Finding, skipped bool) {
	frontmatter, body, ok := splitFrontmatter(content)
	if !ok {
		return Task{}, []Finding{{
			File:    file,
			Field:   "frontmatter",
			Message: "missing closing frontmatter delimiter",
		}}, true
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatter), &raw); err != nil {
		return Task{}, []Finding{{
			File:    file,
			Field:   "frontmatter",
			Message: err.Error(),
		}}, true
	}

	task = Task{File: file, Archived: archived}

	for key := range raw {
		if !knownFrontmatterKeys[key] {
			findings = append(findings, Finding{
				File:    file,
				Field:   key,
				Message: "unrecognized frontmatter key",
			})
		}
	}

	task.ID = getString(raw, "id")
	task.Title = getString(raw, "title")
	task.Status = getString(raw, "status")
	task.Priority = getString(raw, "priority")
	task.Type = getString(raw, "type")
	task.Project = getString(raw, "project")
	task.Milestone = getString(raw, "milestone")
	task.ParentTaskID = getString(raw, "parent_task_id")

	task.Assignees = getStringSlice(raw, "assignee")
	task.Labels = getStringSlice(raw, "labels")
	task.Dependencies = getStringSlice(raw, "dependencies")
	task.References = getStringSlice(raw, "references")
	task.Documentation = getStringSlice(raw, "documentation")
	task.ModifiedFiles = getStringSlice(raw, "modified_files")

	task.Ordinal = getOrdinal(raw)

	extractDate := func(key string) string {
		v, present := raw[key]
		if !present || v == nil {
			return ""
		}
		s, isStr := v.(string)
		if !isStr {
			s = fmt.Sprint(v)
		}
		if s == "" {
			return ""
		}
		if !isValidDateShape(s) {
			findings = append(findings, Finding{
				File:    file,
				Field:   key,
				Message: fmt.Sprintf("date %q does not match either Backlog.md date shape", s),
			})
			return ""
		}
		return s
	}
	task.CreatedDate = extractDate("created_date")
	task.UpdatedDate = extractDate("updated_date")
	task.DueDate = extractDate("due_date")

	for _, name := range findUnknownSections(body) {
		findings = append(findings, Finding{
			File:    file,
			Field:   name,
			Message: "unrecognized body section",
		})
	}

	task.Description, _ = extractSection(body, "DESCRIPTION")
	task.Plan, _ = extractSection(body, "PLAN")
	task.Notes, _ = extractSection(body, "NOTES")
	task.Summary, _ = extractSection(body, "FINAL_SUMMARY")

	if acText, present := extractBlock(body, "AC"); present {
		boxes, joined := extractCheckboxes(acText)
		task.AcceptanceCriteria = boxes
		if joined {
			findings = append(findings, Finding{
				File:    file,
				Field:   "acceptanceCriteria",
				Message: "a multi-line criterion was joined into a single line",
			})
		}
	}

	if dodText, present := extractBlock(body, "DOD"); present {
		boxes, joined := extractCheckboxes(dodText)
		task.DefinitionOfDone = boxes
		if joined {
			findings = append(findings, Finding{
				File:    file,
				Field:   "definitionOfDone",
				Message: "a multi-line item was joined into a single line",
			})
		}
	}

	if commentsText, present := extractBlock(body, "COMMENTS"); present {
		comments, missingCreated := extractComments(commentsText)
		missing := make(map[int]bool, len(missingCreated))
		for _, idx := range missingCreated {
			missing[idx] = true
		}
		for i := range comments {
			if missing[i] {
				findings = append(findings, Finding{
					File:    file,
					Field:   "created",
					Message: fmt.Sprintf("comment %d has no created date", i+1),
				})
				continue
			}
			if comments[i].CreatedAt != "" && !isValidDateShape(comments[i].CreatedAt) {
				findings = append(findings, Finding{
					File:    file,
					Field:   "created",
					Message: fmt.Sprintf("comment %d date %q does not match either Backlog.md date shape", i+1, comments[i].CreatedAt),
				})
				comments[i].CreatedAt = ""
			}
		}
		task.Comments = comments
	}

	return task, findings, false
}

// getString returns raw[key] as a string, or "" when the key is absent.
// Backlog.md always writes these fields as YAML strings; a non-string
// value is rendered with fmt.Sprint rather than dropped, since it still
// carries information a finding elsewhere did not already report.
func getString(raw map[string]interface{}, key string) string {
	v, ok := raw[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

// getStringSlice returns raw[key] as a slice of strings, tal cual: no
// filtering, no deduplication, no space conversion. Backlog.md always
// writes these fields as a YAML list, but a bare scalar is also accepted
// and treated as a single-element list, defensively.
func getStringSlice(raw map[string]interface{}, key string) []string {
	v, ok := raw[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}:
		var out []string
		for _, item := range t {
			if s, ok := item.(string); ok {
				out = append(out, s)
			} else {
				out = append(out, fmt.Sprint(item))
			}
		}
		return out
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	default:
		return nil
	}
}

// getOrdinal returns raw["ordinal"] as a float64 pointer, or nil when the
// task has no ordinal field. It keeps whatever numeric precision YAML
// decoded, so a fractional ordinal (never seen so far, but not forbidden by
// the format) is not truncated.
func getOrdinal(raw map[string]interface{}) *float64 {
	v, ok := raw["ordinal"]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case int:
		f := float64(t)
		return &f
	case int64:
		f := float64(t)
		return &f
	case uint64:
		f := float64(t)
		return &f
	case float64:
		f := t
		return &f
	default:
		return nil
	}
}
