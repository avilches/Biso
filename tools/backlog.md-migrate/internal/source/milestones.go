package source

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// milestoneFrontmatter is the small slice of a milestone file's frontmatter
// this package needs: docs/especificacion.md, "Qué lee del origen", says
// only the title is used, keyed by the milestone's own id.
type milestoneFrontmatter struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
}

// readMilestones reads every ".md" file directly inside dir as a milestone
// and returns a map from its id to its title. A missing dir yields an empty
// map and no findings.
//
// A milestone file with a broken frontmatter is skipped with a Finding, the
// same way a broken task frontmatter is: docs/especificacion.md does not
// single out milestone files as an exception, and the rest of the layout
// keeps this rule for anything with YAML frontmatter.
func readMilestones(dir string) (map[string]string, []Finding, error) {
	titles := make(map[string]string)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return titles, nil, nil
		}
		return nil, nil, fmt.Errorf("source: %w", err)
	}

	var findings []Finding
	for _, entry := range entries {
		if !entry.Type().IsRegular() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("source: %w", err)
		}

		frontmatter, _, ok := splitFrontmatter(string(content))
		if !ok {
			findings = append(findings, Finding{
				File:    entry.Name(),
				Field:   "frontmatter",
				Message: "missing closing frontmatter delimiter",
			})
			continue
		}

		var mf milestoneFrontmatter
		if err := yaml.Unmarshal([]byte(frontmatter), &mf); err != nil {
			findings = append(findings, Finding{
				File:    entry.Name(),
				Field:   "frontmatter",
				Message: err.Error(),
			})
			continue
		}
		if mf.ID == "" {
			continue
		}
		titles[mf.ID] = mf.Title
	}

	return titles, findings, nil
}
