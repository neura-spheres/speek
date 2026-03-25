package core

// Dictionary maps slot names to lists of synonymous words/phrases.
type Dictionary map[string][]string

// DefaultDictionary is the full vocabulary of the Speek language.
// Every entry here is treated as equivalent for pattern matching purposes.
var DefaultDictionary = Dictionary{
	"VERB_DECLARE": {
		"make", "create", "declare", "define", "initialize", "init",
		"new", "set up", "build", "establish", "introduce", "add",
		"spawn", "allocate", "register", "prepare", "setup",
	},
	"VERB_ASSIGN": {
		"set", "put", "assign", "store", "save", "let", "change",
		"update", "give", "move", "write", "load", "feed", "push",
		"mark", "point", "bind", "fix", "force",
	},
	"VERB_PRINT": {
		"show", "print", "display", "output", "log", "write", "say",
		"tell", "echo", "reveal", "dump", "report", "emit", "post",
		"announce", "present", "render", "spit",
	},
	"VERB_ADD": {
		"add", "increase", "increment", "plus", "bump", "raise", "grow",
		"boost", "expand", "extend", "append", "accumulate", "combine",
		"join", "include",
	},
	"VERB_SUBTRACT": {
		"subtract", "decrease", "decrement", "minus", "reduce", "lower",
		"remove", "drop", "shrink", "cut", "trim", "deduct", "strip",
	},
	"VERB_MULTIPLY": {
		"multiply", "times", "scale", "double", "triple", "quadruple",
		"amplify", "factor",
	},
	"VERB_DIVIDE": {
		"divide", "split", "halve", "share",
	},
	"VERB_LOOP": {
		"loop", "repeat", "do", "run", "cycle", "iterate", "go", "keep",
		"execute", "perform", "redo",
	},
	"VERB_IF": {
		"if", "when", "check", "test", "assuming", "suppose", "given",
		"whenever", "provided", "should",
	},
	"VERB_ELSE": {
		"else", "otherwise",
	},
	"VERB_END": {
		"end", "done", "finish", "close", "stop", "endloop", "endif",
	},
	"VERB_FUNCTION": {
		"define", "teach", "create", "make", "build", "write", "declare",
		"establish",
	},
	"VERB_CALL": {
		"call", "run", "use", "execute", "invoke", "do", "trigger",
		"perform", "activate", "fire", "apply", "start", "launch",
	},
	"VERB_RETURN": {
		"return", "give back", "send back", "yield",
	},
	"VERB_DELETE": {
		"delete", "remove", "clear", "reset", "erase", "destroy", "drop",
		"forget", "wipe", "trash", "kill", "unset", "nullify", "discard",
	},
	"VERB_STOP": {
		"stop", "break", "leave",
	},
	"VERB_INPUT": {
		"ask", "read", "get", "input", "accept", "receive", "take",
		"prompt", "request", "scan",
	},
	"VERB_IMPORT": {
		"import", "include", "use", "load", "require",
	},

	"CMP_GT":  {"is greater than", "is more than", "is bigger than", "is above", "is over", "greater than", "more than", "bigger than", "above", "over", "exceeds", "larger than"},
	"CMP_LT":  {"is less than", "is fewer than", "is smaller than", "is below", "is under", "less than", "fewer than", "smaller than", "below", "under", "lower than"},
	"CMP_GTE": {"is greater than or equal", "is at least", "is no less than", "greater than or equal", "at least", "minimum", "no less than"},
	"CMP_LTE": {"is less than or equal", "is at most", "is no more than", "less than or equal", "at most", "maximum", "no more than"},
	"CMP_EQ":  {"equals", "is equal to", "equal to", "same as", "matches", "is"},
	"CMP_NEQ": {"not equal to", "different from", "isn't", "is not", "doesn't equal", "not equal"},
	"CMP_DIV": {"divisible by", "is divisible by", "divides evenly by", "a multiple of"},

	"PREP_TO":   {"to", "into", "as", "be", "become"},
	"PREP_FROM": {"from", "out of", "off"},
	"PREP_BY":   {"by", "with", "using", "via", "through"},
	"PREP_IN":   {"in", "inside", "within", "onto"},
	"PREP_AND":  {"and", "with", "plus"},

	"FILLER": {
		"a", "an", "the", "my", "some", "new", "me", "please", "now", "just",
		"variable", "var", "val", "value", "number", "thing", "item", "object",
		"data", "field", "called", "named", "labeled", "known as",
		"function", "action", "task", "method", "procedure",
		"list", "array", "collection", "group",
	},

	"LOOP_UNIT": {"times", "time", "iterations", "iteration", "rounds", "round", "cycles", "cycle", "steps", "step", "passes", "pass", "x"},

	"BOOL_TRUE":  {"true", "yes", "on", "enabled", "active", "correct", "right"},
	"BOOL_FALSE": {"false", "no", "off", "disabled", "inactive", "wrong"},
}

func AllVerbs(d Dictionary) []string {
	out := []string{}
	verbSlots := []string{
		"VERB_DECLARE", "VERB_ASSIGN", "VERB_PRINT", "VERB_ADD",
		"VERB_SUBTRACT", "VERB_MULTIPLY", "VERB_DIVIDE", "VERB_LOOP",
		"VERB_IF", "VERB_ELSE", "VERB_END", "VERB_FUNCTION", "VERB_CALL",
		"VERB_RETURN", "VERB_DELETE", "VERB_STOP", "VERB_INPUT",
	}
	for _, slot := range verbSlots {
		out = append(out, d[slot]...)
	}
	return out
}
