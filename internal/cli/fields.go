package cli

// This file is the specification table of docs/spec/familias-de-flags.md:
// every field flag of the program, once. `biso new` and `biso set` take all
// of them with the same meaning, and so will the six verbs of the cycle and
// `biso archive`, which is the promise that page makes and which one shared
// table is the only way of keeping.
//
// The four shapes of a list field, the three of the criteria, the two of a
// prose field and the pair of a scalar all come out of the builders below,
// so a field that broke the shape would be visible here as an entry that did
// not go through one of them.

// listFieldFlags are the four fields of
// docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma, each with
// its four flags. The name of the flag is the tail, the field of the JSON
// envelope is the model's own name, and the two token fields close their
// alphabet while the other two are free text, because a reference cannot have
// its alphabet closed without leaving legitimate values out.
var listFieldFlags = []struct {
	suffix   string
	field    string
	alphabet Alphabet
	noun     string
	labels   LabelSyntax
}{
	{suffix: "labels", field: "labels", alphabet: TokenAlphabet, noun: "label", labels: LabelWritten},
	{suffix: "assignees", field: "assignees", alphabet: TokenAlphabet, noun: "assignee"},
	{suffix: "refs", field: "references"},
	{suffix: "deps", field: "dependencies"},
}

// proseFields are the four of
// docs/spec/familias-de-flags.md#campos-de-prosa, each with its two flags.
// There is no flag that replaces a block of prose: replacing one is emptying
// and appending in the same call.
var proseFields = []struct {
	appendName string
	clearName  string
	field      string
}{
	{appendName: "append-desc", clearName: "clear-desc", field: "description"},
	{appendName: "append-plan", clearName: "clear-plan", field: "plan"},
	{appendName: "append-note", clearName: "clear-notes", field: "notes"},
	{appendName: "append-summary", clearName: "clear-summary", field: "summary"},
}

// scalarFields are the ones of
// docs/spec/familias-de-flags.md#campos-escalares. The title and the status
// cannot be emptied, because both are required, so those two carry no
// clearing flag.
var scalarFields = []struct {
	name       string
	field      string
	clear      string
	vocabulary bool
	singleLine bool
	text       bool
	// placement marks the manual order, whose four flags all write the
	// same field and are therefore incompatible with one another
	// (docs/spec/familias-de-flags.md#el-orden-manual).
	placement bool
}{
	{name: "title", field: "title", singleLine: true, text: true},
	{name: "status", field: "status", vocabulary: true},
	{name: "type", field: "type", clear: "clear-type", vocabulary: true},
	{name: "priority", field: "priority", clear: "clear-priority", vocabulary: true},
	{name: "parent", field: "parent", clear: "clear-parent"},
	{name: "due", field: "due", clear: "clear-due"},
	// The manual order is the one scalar whose value is not the value of
	// the field: --ordinal takes first or last and the two flags below take
	// a neighbour, and what gets written is the key the program computes
	// from them (docs/spec/familias-de-flags.md#el-orden-manual).
	{name: "ordinal", field: "ordinal", clear: "clear-ordinal", placement: true},
	// A person field never interprets a leading "@", so --author takes its
	// value exactly as it is typed
	// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).
	{name: "author", field: "author", clear: "clear-author", singleLine: true},
}

// fieldFlags builds the whole table, fresh on every call so that no command
// can change the one another command is reading.
func fieldFlags() []FlagSpec {
	var flags []FlagSpec
	for _, f := range listFieldFlags {
		list := func(prefix string, category Category) FlagSpec {
			return FlagSpec{
				Name: prefix + f.suffix, Value: PlainValue,
				Repeatable: true, Comma: true, Category: category,
				Alphabet: f.alphabet, Noun: f.noun, Labels: f.labels, Field: f.field,
			}
		}
		flags = append(flags,
			FlagSpec{Name: "clear-" + f.suffix, Category: Clear, Field: f.field},
			list("replace-", Replace),
			list("rm-", Remove),
			list("add-", Add),
		)
	}

	// The criteria have three shapes and not four: a criterion's text can
	// contain a comma, so --add-ac never splits on one and there is no
	// whole-list replace (docs/spec/familias-de-flags.md#campos-de-lista-sin-coma-criterios).
	flags = append(flags,
		FlagSpec{Name: "clear-acs", Category: Clear, Field: "acceptanceCriteria"},
		FlagSpec{
			Name: "rm-ac", Value: PlainValue, Repeatable: true,
			Category: Remove, Field: "acceptanceCriteria",
		},
		FlagSpec{
			Name: "add-ac", Value: TextValue, Repeatable: true,
			Category: Add, Field: "acceptanceCriteria",
		},
		FlagSpec{Name: "check-ac", Value: PlainValue, Repeatable: true, Category: CheckAC},
		FlagSpec{Name: "uncheck-ac", Value: PlainValue, Repeatable: true, Category: CheckAC},
	)

	for _, f := range proseFields {
		flags = append(flags,
			FlagSpec{Name: f.clearName, Category: Clear, Field: f.field},
			FlagSpec{
				Name: f.appendName, Value: TextValue,
				Repeatable: true, Category: Add, Field: f.field,
			},
		)
	}

	for _, f := range scalarFields {
		value := PlainValue
		if f.text {
			value = TextValue
		}
		if f.clear != "" {
			flags = append(flags, FlagSpec{
				Name: f.clear, Category: Clear, Field: f.field,
				Conflicts: placementConflicts(f.placement, f.clear),
			})
		}
		scalar := FlagSpec{
			Name: f.name, Value: value, Category: Scalar,
			ClosedVocabulary: f.vocabulary, SingleLine: f.singleLine,
			ClearFlag: f.clear, Field: f.field,
			Conflicts: placementConflicts(f.placement, f.name),
		}
		if f.placement {
			// --ordinal is the one scalar with a domain the program owns
			// and the board does not, so an unknown value is exit code 2
			// and never the 3 of a value the board does not have
			// (docs/spec/familias-de-flags.md#el-orden-manual).
			scalar.Domain = []string{"first", "last"}
			scalar.DomainCode = "invalid_ordinal_value"
			scalar.DomainHints = []string{
				"--ordinal takes first or last; " +
					"to place a task next to another one, use --above or --below",
			}
		}
		flags = append(flags, scalar)
		if !f.placement {
			continue
		}
		for _, name := range []string{"above", "below"} {
			flags = append(flags, FlagSpec{
				Name: name, Value: PlainValue, Category: Scalar,
				Field: f.field, Conflicts: placementConflicts(f.placement, name),
			})
		}
	}

	// The comments, which are the one list of objects that is never edited
	// in place: a body and an author are written once and for all
	// (docs/spec/familias-de-flags.md#comentarios).
	flags = append(flags,
		FlagSpec{
			Name: "rm-comment", Value: PlainValue, Repeatable: true,
			Category: Remove, Field: "comments",
		},
		FlagSpec{
			Name: "set-comment-date", Value: PlainValue, Repeatable: true,
			Category: CommentDate, Pair: PairAtLastEquals, PairSyntax: "<sel>=<instant>",
			Field: "comments",
		},
		FlagSpec{
			Name: "comment", Value: TextValue, Repeatable: true,
			Category: AddComment, Field: "comments",
		},
		FlagSpec{
			Name: "comment-author", Value: PlainValue, Category: NotAChange,
			Requires: []string{"comment"}, Field: "comment-author",
		},
	)
	return flags
}

// placementFlagNames are the four flags of the manual order, in the order
// of the table above, which is the order a message that names two of them
// names them in.
var placementFlagNames = []string{"clear-ordinal", "ordinal", "above", "below"}

// placementConflicts answers what one flag of the manual order cannot share
// a call with: the other three, because all four write the same field
// (docs/spec/familias-de-flags.md#el-orden-manual). For every other scalar
// it answers nothing, since the pair of a scalar and its --clear-<field> is
// not a conflict anywhere else: --clear-due --due 2026-09-20 is the ordinary
// way of replacing a value, and only here is it a contradiction, because
// what the caller would be asking for is a place and no place at once.
func placementConflicts(placement bool, self string) []string {
	if !placement {
		return nil
	}
	others := make([]string, 0, len(placementFlagNames)-1)
	for _, name := range placementFlagNames {
		if name != self {
			others = append(others, name)
		}
	}
	return others
}
