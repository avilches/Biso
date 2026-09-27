package convert

import "strings"

// ConvertDate converts a Backlog.md date already known to have one of the
// two shapes reading the source board validated (source.isValidDateShape:
// "YYYY-MM-DD HH:mm" or "YYYY-MM-DD") into biso's UTC instant shape
// (docs/especificacion.md, "Fechas"):
//
//	"2026-09-20 22:08" -> "2026-09-20T22:08:00Z"
//	"2026-09-20"        -> "2026-09-20T00:00:00Z"
//
// It does not re-validate the shape: reading the source board already
// discarded, with a finding, anything that matched neither form, so an
// unexpected shape reaching this function would be a bug earlier in the
// pipeline, not a new finding for ConvertDate to raise.
//
// ConvertDate("") is "", so a caller does not need to special-case an
// absent date (an absent updated_date, or a comment with no created line)
// before calling it.
//
// due_date is explicitly excluded from this conversion:
// docs/especificacion.md, "El mapeo de campos", says due_date -> due stays
// "YYYY-MM-DD, tal cual", so Task copies t.DueDate into Result.Due directly
// and never passes it through ConvertDate.
func ConvertDate(s string) string {
	if s == "" {
		return ""
	}
	if idx := strings.IndexByte(s, ' '); idx >= 0 {
		return s[:idx] + "T" + s[idx+1:] + ":00Z"
	}
	return s + "T00:00:00Z"
}
