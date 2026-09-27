package source

import "regexp"

// dateWithTime and dateOnly are the two shapes Backlog.md writes a date in:
// "YYYY-MM-DD HH:mm" for created_date, updated_date, and a comment's
// created, and "YYYY-MM-DD" for due_date. The check is shape only (four
// digits, two digits, two digits, and so on): it does not reject a
// calendar-invalid date such as 2026-02-30, because the specification only
// asks that the text match one of the two shapes, and leaves any further
// validation to convert.ConvertDate, which converts it to UTC.
var (
	dateWithTimeShape = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$`)
	dateOnlyShape     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// isValidDateShape reports whether s has one of the two shapes Backlog.md
// uses for a date, as described in docs/especificacion.md, "Fechas".
func isValidDateShape(s string) bool {
	return dateWithTimeShape.MatchString(s) || dateOnlyShape.MatchString(s)
}
