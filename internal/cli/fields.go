package cli

// This file is the specification table of docs/spec/familias-de-flags.md:
// every field flag of the program, once. `biso new` and `biso set` take all
// of them with the same meaning, and so will the six verbs of the cycle and
// `biso archive`, which is the promise that page makes and which one shared
// table is the only way of keeping.
//
// The four shapes of a list field, the three of the criteria, the two of a
// prose field, the three of an external field and the pair of a scalar all
// come out of the same four builders below, so a field that broke the shape
// would be visible here as an entry that did not go through one of them.

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
}{
	{suffix: "labels", field: "labels", alphabet: TokenAlphabet, noun: "label"},
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
}{
	{name: "title", field: "title", singleLine: true, text: true},
	{name: "status", field: "status", vocabulary: true},
	{name: "type", field: "type", clear: "clear-type", vocabulary: true},
	{name: "priority", field: "priority", clear: "clear-priority", vocabulary: true},
	{name: "parent", field: "parent", clear: "clear-parent"},
	{name: "due", field: "due", clear: "clear-due"},
	{name: "ordinal", field: "ordinal", clear: "clear-ordinal"},
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
				Alphabet: f.alphabet, Noun: f.noun, Field: f.field,
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
			flags = append(flags, FlagSpec{Name: f.clear, Category: Clear, Field: f.field})
		}
		flags = append(flags, FlagSpec{
			Name: f.name, Value: value, Category: Scalar,
			ClosedVocabulary: f.vocabulary, SingleLine: f.singleLine,
			ClearFlag: f.clear, Field: f.field,
		})
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
