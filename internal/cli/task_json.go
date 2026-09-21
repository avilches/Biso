package cli

import (
	"sort"
	"time"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the one place a task becomes JSON. `task.list` and
// `task.get` are the same object, the second with the body of the task
// added (docs/spec/cmd/get.md#el-esquema-json), so they are built by one
// function and an addition and never by two lists of keys that could drift
// apart.
//
// The object is a map and not a struct because `biso get --section` prints
// a subset of its keys, and the subset is chosen at run time. The order of
// the keys of a JSON object means nothing, so nothing is lost by it.

// listKeys are the keys of a task in `task.list`, which are also the keys
// the `meta` section of `biso get` prints.
func taskObject(v ops.TaskView) map[string]any {
	t := v.Task
	return map[string]any{
		"id":             t.ID,
		"title":          t.Title,
		"status":         t.Status,
		"type":           orNull(t.Type),
		"priority":       orNull(t.Priority),
		"assignees":      list(t.Assignees),
		"author":         orNull(t.Author),
		"labels":         list(t.Labels),
		"parent":         orNull(t.Parent),
		"dependencies":   list(t.Dependencies),
		"references":     list(t.References),
		"modifiedFiles":  list(t.ModifiedFiles),
		"due":            day(t.Due),
		"ordinal":        ordinal(t),
		"createdAt":      instantOrNull(t.CreatedAt),
		"updatedAt":      instantOrNull(t.UpdatedAt),
		"leaseExpiresAt": instantOrNull(t.LeaseExpiresAt),
		"leaseHolder":    orNull(t.LeaseHolder),
		"acDone":         t.AcDone(),
		"acTotal":        t.AcTotal(),
		"commentCount":   t.CommentCount(),
		"urgency":        urgency(v.Urgency),
		"blocks":         list(v.Blocks),
		"blocked":        v.Blocked,
		"waiting":        v.Waiting,
		"leaseExpired":   v.LeaseExpired,
		"archived":       t.Archived,
		"ext":            extObject(t.Ext),
	}
}

// sectionKeys says which keys of `data.task` each section of
// docs/spec/cmd/get.md governs. `meta` is every key of `task.list`, which
// is why it is not listed here: it is the whole object minus the body.
var sectionKeys = map[string][]string{
	ops.SectionDesc:     {"description"},
	ops.SectionAC:       {"acceptanceCriteria"},
	ops.SectionPlan:     {"plan"},
	ops.SectionNotes:    {"notes"},
	ops.SectionSummary:  {"summary"},
	ops.SectionComments: {"comments"},
	ops.SectionQuestion: {"question"},
}

// getObject is the object of `task.get`: the one above plus the body, cut
// down to the sections the call asked for.
//
// With --section it carries `id` and the keys of those sections and no
// other, which is the exception to "no key comes and goes with the data"
// that docs/spec/contrato-json.md#números-fechas-y-ausencias declares: what
// governs it is a flag, and whoever wrote the flag knows what they asked
// for. A section that was asked for and happens to be empty is still
// there, as null: only the text output leaves it out.
func getObject(r *ops.GetResult) map[string]any {
	object := taskObject(r.Task)
	t := r.Task.Task
	object["description"] = orNull(t.Description)
	object["acceptanceCriteria"] = criteriaObject(t.AcceptanceCriteria)
	object["plan"] = orNull(t.Plan)
	object["notes"] = orNull(t.Notes)
	object["summary"] = orNull(t.Summary)
	object["comments"] = commentsObject(t.Comments)
	object["question"] = questionObject(t.Question)
	if r.Task.Breakdown != nil {
		object["urgencyBreakdown"] = breakdownObject(r.Task.Breakdown)
	}
	if r.WholeCard {
		return object
	}

	wanted := map[string]bool{"id": true}
	for _, section := range r.Sections {
		if section == ops.SectionMeta {
			for key := range taskObject(r.Task) {
				wanted[key] = true
			}
			continue
		}
		for _, key := range sectionKeys[section] {
			wanted[key] = true
		}
	}
	// urgencyBreakdown is governed by its own flag and not by a section,
	// so --explain-urgency keeps it whatever was asked for.
	wanted["urgencyBreakdown"] = true
	for key := range object {
		if !wanted[key] {
			delete(object, key)
		}
	}
	return object
}

// breakdownObject is the `urgencyBreakdown` of
// docs/spec/cmd/get.md#el-esquema-json: one number per term, which is the
// product that enters the sum, and `active` the one term that is an object
// because it also says why it contributed nothing.
//
// Every term goes through the same serializer as the urgency itself, so it
// writes one digit after the point: they are the addends of that very
// number, and the schema writes all seven of them as decimals, the seven
// zeros of a terminal task included
// (docs/spec/contrato-json.md#números-fechas-y-ausencias).
func breakdownObject(b *model.UrgencyBreakdown) map[string]any {
	active := map[string]any{"value": urgencyTerm(b.Active.Value), "reason": nil}
	if b.ActiveReason != model.ActiveContributes {
		active["reason"] = b.ActiveReason
	}
	return map[string]any{
		"priority": urgencyTerm(b.Priority.Value),
		"active":   active,
		"blocking": urgencyTerm(b.Blocking.Value),
		"blocked":  urgencyTerm(b.Blocked.Value),
		"due":      urgencyTerm(b.Due.Value),
		"criteria": urgencyTerm(b.Criteria.Value),
		"age":      urgencyTerm(b.Age.Value),
	}
}

// urgencyTerm is one addend of the urgency, ready to be marshalled: a minus
// zero turned into a zero and the one decimal the contract asks for.
func urgencyTerm(value float64) urgency {
	return urgency(noNegativeZero(value))
}

func criteriaObject(criteria []model.Criterion) []map[string]any {
	out := make([]map[string]any, 0, len(criteria))
	for _, c := range criteria {
		out = append(out, map[string]any{
			"key": c.Key, "text": c.Text, "checked": c.Checked,
		})
	}
	return out
}

func commentsObject(comments []model.Comment) []map[string]any {
	out := make([]map[string]any, 0, len(comments))
	for _, c := range comments {
		out = append(out, map[string]any{
			"key":       c.Key,
			"author":    orNull(c.Author),
			"createdAt": instantOrNull(c.CreatedAt),
			"body":      c.Body,
		})
	}
	return out
}

func questionObject(q *model.Question) any {
	if q == nil {
		return nil
	}
	return map[string]any{
		"author":  orNull(q.Author),
		"askedAt": instantOrNull(q.AskedAt),
		"body":    q.Body,
	}
}

// extObject is the external fields, which are a map and therefore `{}` when
// there are none, never null
// (docs/spec/contrato-json.md#números-fechas-y-ausencias).
func extObject(ext map[string]string) map[string]string {
	out := make(map[string]string, len(ext))
	for k, v := range ext {
		out[k] = v
	}
	return out
}

// orNull is the convention of the contract for every string field: an empty
// string is the absence of a value, and the absence of a value is null.
func orNull(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// list is a list<string> as the contract carries it: `[]` when it has
// nothing in it, never null.
func list(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func ordinal(t *model.Task) any {
	if t.Ordinal == nil {
		return nil
	}
	return *t.Ordinal
}

// instantOrNull is a date field that names an instant: ISO 8601 in UTC, to
// the second, ending in Z.
func instantOrNull(at time.Time) any {
	if at.IsZero() {
		return nil
	}
	return at.UTC().Format(model.InstantLayout)
}

// day is `due`, the one date field that names a calendar day and therefore
// travels as YYYY-MM-DD.
func day(at time.Time) any {
	if at.IsZero() {
		return nil
	}
	return at.UTC().Format(model.DateLayout)
}

// sortedKeys is how a map of external fields is written wherever the output
// has to be the same twice: a map has no order of its own.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
