package source

// Task is a single Backlog.md task, read from tasks/, completed/, or
// archive/tasks/, with every field docs/especificacion.md, "El mapeo de
// campos", knows about kept verbatim: no id rewriting, no date conversion
// to UTC, no destination vocabulary, no milestone::/project:: labels. Those
// belong to the conversion engine that runs on top of this package.
//
// A Task in a Board.Tasks slice always came from a file whose frontmatter
// parsed as valid YAML; a file whose frontmatter did not parse never
// produces a Task, only a Finding with field "frontmatter".
//
// A string field is empty, and a slice or pointer field is nil, when
// Backlog.md omitted it or when it was present but invalid (an unparseable
// date), in which case a Finding was also raised for it.
type Task struct {
	// File is the source file's base name (for example "task-12 -
	// Title.md"), kept for findings and diagnostics.
	File string
	// Archived is true for a task read from archive/tasks/.
	Archived bool

	ID     string
	Title  string
	Status string
	// Assignees is the assignee list exactly as Backlog.md stored it:
	// with or without a leading "@", with internal spaces untouched.
	// Converting spaces to hyphens is the conversion engine's job.
	Assignees    []string
	CreatedDate  string
	UpdatedDate  string
	DueDate      string
	Labels       []string
	Milestone    string
	Dependencies []string
	References   []string
	// Documentation and ModifiedFiles are the two compatibility lists
	// biso's own "new --from" merges into references on its own; this
	// package keeps them separate, exactly as read.
	Documentation []string
	ModifiedFiles []string
	Priority      string
	Type          string
	Project       string
	// Ordinal is nil when the task has no ordinal field at all. It is a
	// float64 rather than an int or a string because no field in
	// Backlog.md has ever been seen with a fractional ordinal, but the
	// format does not forbid one, so this package keeps whatever
	// precision the source has instead of assuming an integer.
	Ordinal      *float64
	ParentTaskID string

	// Description, Plan, Notes, and Summary are the four free text
	// sections (DESCRIPTION, PLAN, NOTES, FINAL_SUMMARY), each with its
	// single leading and trailing newline removed. An absent section is
	// the empty string.
	Description string
	Plan        string
	Notes       string
	Summary     string

	// AcceptanceCriteria and DefinitionOfDone are kept as two separate
	// lists: reading the board does not merge Definition of Done into
	// acceptance criteria with the "#dod" suffix, because that needs to
	// know the destination's next free acceptance criteria key first,
	// which convert.MergeDefinitionOfDone does instead.
	AcceptanceCriteria []Checkbox
	DefinitionOfDone   []Checkbox

	Comments []Comment
}

// Checkbox is one acceptance criterion or Definition of Done item: a line
// "- [ ] #n text" or "- [x] #n text", with Number kept as written in the
// source (not renumbered) and Text holding every continuation line already
// joined with a single space.
type Checkbox struct {
	Number  int
	Checked bool
	Text    string
}

// Comment is one block from a task's Comments section. Author is empty
// when the block had no "author:" line, which is valid and never raises a
// Finding. CreatedAt is empty when the block had no "created:" line, or
// when it had one whose value did not match either date shape; both cases
// raise a Finding with field "created".
type Comment struct {
	Author    string
	CreatedAt string
	Body      string
}

// knownFrontmatterKeys are the frontmatter keys docs/especificacion.md, "El
// mapeo de campos", recognizes, measured against tasks created by the
// Backlog.md CLI. Any other key in a task's frontmatter is a Finding, and
// its value is not kept.
var knownFrontmatterKeys = map[string]bool{
	"id":             true,
	"title":          true,
	"status":         true,
	"assignee":       true,
	"created_date":   true,
	"updated_date":   true,
	"due_date":       true,
	"labels":         true,
	"milestone":      true,
	"dependencies":   true,
	"references":     true,
	"documentation":  true,
	"modified_files": true,
	"priority":       true,
	"type":           true,
	"project":        true,
	"ordinal":        true,
	"parent_task_id": true,
}
