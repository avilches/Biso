package convert

import "backlog.md-migrate/internal/source"

// MergeDefinitionOfDone appends every Definition of Done item after a
// task's own acceptance criteria, in order (docs/especificacion.md,
// "Definición de hecho", and docs/decisiones.md, "La definición de hecho se
// marca con el sufijo #dod"). Each appended item gets:
//
//   - Number: the next free key after the highest acceptance criterion key
//     the task already has, or 1 when it has none at all;
//   - Checked: copied as-is from the Definition of Done item;
//   - Text: the item's own text, followed by a space and "#dod".
//
// A task with no Definition of Done items (len(dod) == 0) returns criteria
// unchanged. This is never a Finding: folding Definition of Done into
// acceptance criteria this way is the normal conversion, not a problem to
// report.
func MergeDefinitionOfDone(criteria, dod []source.Checkbox) []source.Checkbox {
	if len(dod) == 0 {
		return criteria
	}

	nextKey := 1
	for _, c := range criteria {
		if c.Number >= nextKey {
			nextKey = c.Number + 1
		}
	}

	merged := append([]source.Checkbox(nil), criteria...)
	for _, d := range dod {
		merged = append(merged, source.Checkbox{
			Number:  nextKey,
			Checked: d.Checked,
			Text:    d.Text + " #dod",
		})
		nextKey++
	}
	return merged
}
