package convert

import (
	"fmt"
	"sort"
	"strings"

	"backlog.md-migrate/internal/source"
)

// MilestoneSlugs computes the milestone slug docs task point 4 describes
// for every milestone in titles (a source.Board.Milestones map, milestone
// id to title), plus:
//
//   - a Finding for a title that slugifies to the empty string (docs task
//     point 4, step 3), falling back to using the milestone id itself as
//     the slug;
//   - a Finding for two or more DISTINCT milestone ids that end up with
//     the same slug (docs task point 4, step 4), because they would
//     collapse into a single milestone::<slug> label.
//
// This must run once over the whole map, not once per task: the collision
// check needs every milestone's slug at once to compare them against each
// other, which a single task's conversion cannot see on its own. A
// milestone id that no task in the batch actually references is still
// included and checked exactly like one that is; docs/especificacion.md
// never makes that distinction.
//
// Task looks up its own t.Milestone in the returned map; a milestone id
// that is not a key of titles at all (the milestone file the task's
// milestone field named was never found while reading the board) is a
// separate, per-task Finding that Task/ScopedLabels raises, not this
// function, since it depends on which tasks reference which id, not on the
// milestones map alone.
func MilestoneSlugs(titles map[string]string) (slugs map[string]string, findings []source.Finding) {
	ids := make([]string, 0, len(titles))
	for id := range titles {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	slugs = make(map[string]string, len(titles))
	for _, id := range ids {
		slug := slugify(titles[id])
		if slug == "" {
			findings = append(findings, source.Finding{
				File:  "-",
				Field: "milestone",
				Message: fmt.Sprintf(
					"milestone %s has a title that slugifies to nothing, using the milestone id as the slug instead",
					id,
				),
			})
			slug = id
		}
		slugs[id] = slug
	}

	idsBySlug := make(map[string][]string)
	for _, id := range ids {
		slug := slugs[id]
		idsBySlug[slug] = append(idsBySlug[slug], id)
	}
	collidingSlugs := make([]string, 0, len(idsBySlug))
	for slug, colliding := range idsBySlug {
		if len(colliding) > 1 {
			collidingSlugs = append(collidingSlugs, slug)
		}
	}
	sort.Strings(collidingSlugs)
	for _, slug := range collidingSlugs {
		findings = append(findings, source.Finding{
			File:  "-",
			Field: "milestone",
			Message: fmt.Sprintf(
				"milestones %s all slugify to %q, they would collapse into the same milestone::%s label",
				strings.Join(idsBySlug[slug], ", "), slug, slug,
			),
		})
	}

	return slugs, findings
}

// scopedLabelKey returns a scoped label's key: the text before its first
// colon (docs/spec/valores-de-entrada.md, "Las etiquetas con ámbito": the
// key is what precedes the separator, whether the separator turns out to
// be one colon or two, since a second colon only extends the separator, it
// never moves where the key starts). A label with no colon at all has no
// key and hasKey is false; docs task point 4 says such a label "nunca
// choca" with a derived milestone:: or project:: label.
func scopedLabelKey(label string) (key string, hasKey bool) {
	idx := strings.IndexByte(label, ':')
	if idx < 0 {
		return "", false
	}
	return label[:idx], true
}

// RemoveCollidingScopedLabel removes every label in labels whose scoped
// key case-folds (Unicode case folding, no diacritic or space stripping,
// per docs/spec/valores-de-entrada.md, "La clave se compara plegada, y el
// separador no cuenta al comparar valores") to targetKey, returning the
// filtered slice and a Finding for each one removed
// (docs/decisiones.md, "Una etiqueta con ámbito derivada del dato real gana
// a la etiqueta de origen que choca con ella"). A label with no colon, or
// whose key folds to something else, is kept untouched.
//
// This key comparison is deliberately its own function and not
// normalizeVocabulary: the two rules differ (this one folds case only, it
// does not strip diacritics, spaces, hyphens, or underscores), and reusing
// the vocabulary function here would silently change which labels collide.
//
// Exported because phase 4b (identifiers.go) reuses it verbatim for the
// backlog.id:: label, per docs/especificacion.md, "Identificadores", point
// 9: a source label whose key collides with backlog.id is dropped the same
// way one colliding with milestone or project already is here.
func RemoveCollidingScopedLabel(file string, labels []string, targetKey string) (kept []string, findings []source.Finding) {
	foldedTarget := foldCase(targetKey)
	for _, label := range labels {
		key, hasKey := scopedLabelKey(label)
		if hasKey && foldCase(key) == foldedTarget {
			findings = append(findings, source.Finding{
				File:  file,
				Field: "labels",
				Message: fmt.Sprintf(
					"source label %q collides with the derived %s:: label, dropped",
					label, targetKey,
				),
			})
			continue
		}
		kept = append(kept, label)
	}
	return kept, findings
}

// ScopedLabels computes the milestone:: and project:: labels of docs task
// point 4 for a single task, on top of labels: the task's origin labels
// list, already cleaned through CleanTokenList. It:
//
//  1. Looks up milestoneID (t.Milestone, a milestone id such as "m-4") in
//     milestoneSlugs, the map MilestoneSlugs already computed for the
//     whole batch. When milestoneID names a milestone that map does not
//     have (the milestone file the id pointed at was never found while
//     reading the board), it falls back to using the id itself as the
//     slug, with a Finding.
//  2. Slugifies projectValue (t.Project, already a plain string with no
//     file to look up) directly with slugify. When that yields the empty
//     string, docs/especificacion.md does not define a fallback the way it
//     does for milestone's id; this implementation falls back to the raw
//     projectValue itself as the slug, with a Finding (a decision made by
//     this phase's implementer, documented in its final report, since the
//     specification leaves it open).
//  3. Removes from labels any label whose key collides with "milestone" or
//     "project" (RemoveCollidingScopedLabel), then appends
//     "milestone::<slug>" and "project::<slug>", in that order, after
//     whatever origin labels remain: the exact final ordering docs task
//     point 4 requires.
//
// A task with no milestone (milestoneID == "") gets no milestone:: label
// and no related Finding; the same holds for project. The backlog.id::
// label for a subtask's original id is a later phase's job, not this
// one's.
func ScopedLabels(file string, labels []string, milestoneID, projectValue string, milestoneSlugs map[string]string) ([]string, []source.Finding) {
	result := append([]string(nil), labels...)
	var findings []source.Finding

	if milestoneID != "" {
		slug, ok := milestoneSlugs[milestoneID]
		if !ok {
			findings = append(findings, source.Finding{
				File:  file,
				Field: "milestone",
				Message: fmt.Sprintf(
					"milestone file for %q not found, using the id as the slug instead",
					milestoneID,
				),
			})
			slug = milestoneID
		}

		kept, collisionFindings := RemoveCollidingScopedLabel(file, result, "milestone")
		result = kept
		findings = append(findings, collisionFindings...)
		result = append(result, "milestone::"+slug)
	}

	if projectValue != "" {
		slug := slugify(projectValue)
		if slug == "" {
			findings = append(findings, source.Finding{
				File:  file,
				Field: "project",
				Message: fmt.Sprintf(
					"project %q slugifies to nothing, using the raw value as the slug instead",
					projectValue,
				),
			})
			slug = projectValue
		}

		kept, collisionFindings := RemoveCollidingScopedLabel(file, result, "project")
		result = kept
		findings = append(findings, collisionFindings...)
		result = append(result, "project::"+slug)
	}

	return result, findings
}
