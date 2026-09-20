package cli

import (
	"fmt"
	"strings"

	"biso/internal/ops"
)

// This file is `biso doctor` (docs/spec/cmd/doctor.md): the report, whole,
// on stdout, and the exit code that comes out of it.
//
// The report travels on stdout and not on stderr because it is what the
// command produces and not a message beside something else
// (docs/spec/salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno). Its
// two levels are told apart by the headings `Errors:` and `Warnings:` and
// never by the `warning:` prefix, which that same section reserves for
// stderr.

func runDoctor(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	result, err := ops.Doctor(env, ops.DoctorParams{
		Fix: p.Has("fix"), DryRun: p.Has("dry-run"),
	})
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	printWarnings(s, p)
	if asJSON {
		writeEnvelope(s, env, "doctor", doctorData(result))
		return doctorCode(result)
	}
	fmt.Fprint(s.Stdout, renderDoctor(result))
	return doctorCode(result)
}

// doctorCode is the table of codes of that page: 6 while an error remains,
// and 0 when nothing was wrong or every error found was repaired. A
// preview answers the code the real call that follows it would answer, and
// never a 7 of its own, because this command has no 7.
func doctorCode(r *ops.DoctorResult) int {
	if len(r.Problems) > 0 {
		return 6
	}
	return 0
}

// renderDoctor writes the report: the count, the errors, the warnings and
// what was repaired, each group under its own heading and an empty group
// not printed at all.
func renderDoctor(r *ops.DoctorResult) string {
	var b strings.Builder
	found := len(r.Problems) + len(r.Fixed)
	if found == 0 && len(r.Warnings) == 0 {
		return "no problems found\n"
	}
	// The count of the first line is of what the check found, repaired or
	// not, so the same board says the same number with --fix and without
	// it (docs/spec/cmd/doctor.md#salida).
	fmt.Fprintf(&b, "%s found, %s found\n",
		plural(found, "error", "errors"), plural(len(r.Warnings), "warning", "warnings"))
	if len(r.Problems) > 0 {
		b.WriteString("Errors:\n")
		for _, f := range r.Problems {
			b.WriteString(doctorLine(f))
		}
	}
	if len(r.Warnings) > 0 {
		b.WriteString("Warnings:\n")
		for _, f := range r.Warnings {
			b.WriteString(doctorLine(f))
		}
	}
	if len(r.Fixed) > 0 {
		if r.DryRun {
			fmt.Fprintf(&b, "%s would be fixed, nothing was written (--dry-run)\n",
				plural(len(r.Fixed), "error", "errors"))
		} else {
			fmt.Fprintf(&b, "%s fixed\n", plural(len(r.Fixed), "error", "errors"))
		}
		for _, f := range r.Fixed {
			b.WriteString(doctorLine(f))
		}
	}
	return b.String()
}

// doctorLine is one finding: two spaces, the task it is about when it is
// about one, and the message. A finding of the board itself carries no
// identifier, so its message starts right after the two spaces.
func doctorLine(f ops.Finding) string {
	if f.Task == "" {
		return "  " + f.Message + "\n"
	}
	return "  " + f.Task + "  " + f.Message + "\n"
}

func plural(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

// The doctor envelope of docs/spec/cmd/doctor.md#el-esquema-json. There is
// no count key: the three numbers of the text come out of the length of the
// three lists, and `problems` are the errors that remain, so the number of
// the first line is len(problems) + len(fixed).
type doctorEnvelopeData struct {
	Problems []doctorFinding `json:"problems"`
	Warnings []doctorFinding `json:"warnings"`
	Fixed    []doctorRepair  `json:"fixed"`
}

type doctorFinding struct {
	Task    *string `json:"task"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
}

// doctorRepair carries no task: a repair is named by what it repaired, and
// the message already says which task that was where there is one.
type doctorRepair struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func doctorData(r *ops.DoctorResult) doctorEnvelopeData {
	data := doctorEnvelopeData{
		Problems: []doctorFinding{},
		Warnings: []doctorFinding{},
		Fixed:    []doctorRepair{},
	}
	for _, f := range r.Problems {
		data.Problems = append(data.Problems, doctorFindingOf(f))
	}
	for _, f := range r.Warnings {
		data.Warnings = append(data.Warnings, doctorFindingOf(f))
	}
	for _, f := range r.Fixed {
		data.Fixed = append(data.Fixed, doctorRepair{Code: f.Code, Message: f.Message})
	}
	return data
}

// doctorFindingOf writes `task` as null when the finding is about the board
// and not about any one task, which is the same criterion the two board
// level warnings already use.
func doctorFindingOf(f ops.Finding) doctorFinding {
	out := doctorFinding{Code: f.Code, Message: f.Message}
	if f.Task != "" {
		task := f.Task
		out.Task = &task
	}
	return out
}
