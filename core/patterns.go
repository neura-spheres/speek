package core

// PatternTemplate describes one way to write a Speek statement.
// Template slots like {VERB_DECLARE}, {NAME}, {VALUE} are replaced by
// the compiler with actual regex groups.
type PatternTemplate struct {
	Op       string
	Template string
	Captures []string // named capture group order
}

// AllPatterns is the master list of every supported statement form.
// More specific patterns must come before more general ones.
var AllPatterns = []PatternTemplate{

	// COMMENTS
	{Op: "comment", Template: `#.*`, Captures: nil},
	{Op: "comment", Template: `//.*`, Captures: nil},
	{Op: "comment", Template: `--.*`, Captures: nil},
	{Op: "comment", Template: `note:.*`, Captures: nil},

	// ELSE IF (must come before plain IF and ELSE)
	{
		Op:       "elseif",
		Template: `{VERB_ELSE}\s+if\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},
	{
		Op:       "elseif",
		Template: `otherwise\s+if\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},

	// ELSE
	{Op: "else", Template: `{VERB_ELSE}`, Captures: nil},

	// END
	{Op: "end", Template: `{VERB_END}(?:\s+.*)?`, Captures: nil},

	// EXIT PROGRAM (must be before BREAK so "exit" and "quit" route here, not to loop-break)
	{Op: "exit_prog", Template: `(?:exit|quit|halt|terminate|abort)\s+with\s+(?P<code>\d+)`, Captures: []string{"code"}},
	{Op: "exit_prog", Template: `(?:exit|quit|halt|terminate|abort)(?:\s+.*)?`, Captures: nil},

	// BREAK / CONTINUE
	{Op: "break", Template: `{VERB_STOP}(?:\s+.*)?`, Captures: nil},
	{Op: "continue", Template: `(?:continue|next|skip)(?:\s+.*)?`, Captures: nil},

	// RETURN
	{
		Op:       "return",
		Template: `{VERB_RETURN}\s+{VALUE}`,
		Captures: []string{"value"},
	},

	// FUNCTION DEFINITION
	// The word "function" (or synonym) must appear explicitly to prevent
	// "make a variable called x" from matching as a function definition.
	{
		Op:       "fn_def",
		Template: `{VERB_FUNCTION}\s+{FILLER}\s*(?:function|action|task|method|procedure)\s+{FILLER}\s*{NAME}\s+that\s+takes\s+{PARAMS}`,
		Captures: []string{"name", "params"},
	},
	{
		Op:       "fn_def",
		Template: `{VERB_FUNCTION}\s+{FILLER}\s*(?:function|action|task|method|procedure)\s+{FILLER}\s*{NAME}\s+with\s+parameters?\s+{PARAMS}`,
		Captures: []string{"name", "params"},
	},
	{
		Op:       "fn_def",
		Template: `{VERB_FUNCTION}\s+{FILLER}\s*(?:function|action|task|method|procedure)\s+{FILLER}\s*{NAME}\s+using\s+{PARAMS}`,
		Captures: []string{"name", "params"},
	},
	{
		Op:       "fn_def",
		Template: `{VERB_FUNCTION}\s+{FILLER}\s*(?:function|action|task|method|procedure)\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},

	// FUNCTION CALL
	{
		Op:       "fn_call",
		Template: `{VERB_CALL}\s+{FILLER}\s*{NAME}\s+with\s+{ARGS}`,
		Captures: []string{"name", "args"},
	},
	{
		Op:       "fn_call",
		Template: `{VERB_CALL}\s+{FILLER}\s*{NAME}\s+using\s+{ARGS}`,
		Captures: []string{"name", "args"},
	},
	{
		Op:       "fn_call",
		Template: `{VERB_CALL}\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},

	// LOOP TIMES
	{
		Op:       "loop",
		Template: `{VERB_LOOP}\s+{FILLER}\s*{NUMBER}\s+{LOOP_UNIT}`,
		Captures: []string{"number"},
	},
	{
		Op:       "loop",
		Template: `{NUMBER}\s+{LOOP_UNIT}`,
		Captures: []string{"number"},
	},

	// LOOP WHILE
	{
		Op:       "while",
		Template: `{VERB_LOOP}\s+while\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},
	{
		Op:       "while",
		Template: `while\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},
	{
		Op:       "while",
		Template: `keep\s+going\s+while\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},

	// FOR EACH
	{
		Op:       "foreach",
		Template: `for\s+each\s+{NAME}\s+in\s+{LISTNAME}`,
		Captures: []string{"name", "listname"},
	},
	{
		Op:       "foreach",
		Template: `go\s+through\s+each\s+{NAME}\s+in\s+{LISTNAME}`,
		Captures: []string{"name", "listname"},
	},
	{
		Op:       "foreach",
		Template: `{VERB_LOOP}\s+through\s+each\s+{NAME}\s+in\s+{LISTNAME}`,
		Captures: []string{"name", "listname"},
	},
	{
		Op:       "foreach",
		Template: `{VERB_LOOP}\s+each\s+{NAME}\s+in\s+{LISTNAME}`,
		Captures: []string{"name", "listname"},
	},

	// FOR RANGE
	{
		Op:       "for",
		Template: `{VERB_LOOP}\s+{FILLER}\s*{NAME}\s+from\s+{VALUE}\s+to\s+{TOVALUE}`,
		Captures: []string{"name", "value", "tovalue"},
	},
	{
		Op:       "for",
		Template: `for\s+{NAME}\s+from\s+{VALUE}\s+to\s+{TOVALUE}`,
		Captures: []string{"name", "value", "tovalue"},
	},

	// IF
	{
		Op:       "if",
		Template: `{VERB_IF}\s+{NAME}\s+{CMP}\s+{VALUE}`,
		Captures: []string{"name", "cmp", "value"},
	},

	// SLEEP / WAIT
	{Op: "sleep_stmt", Template: `(?:wait|sleep|pause)\s+(?P<seconds>\d+(?:\.\d+)?)\s+(?:seconds?|secs?)`, Captures: []string{"seconds"}},
	{Op: "sleep_stmt", Template: `(?:wait|sleep|pause)\s+(?P<seconds>\d+(?:\.\d+)?)`, Captures: []string{"seconds"}},

	// IN-PLACE LIST MUTATIONS (sort_list_desc must come before sort_list)
	{Op: "sort_list_desc", Template: `sort\s+{FILLER}\s*{NAME}\s+(?:descending|desc)`, Captures: []string{"name"}},
	{Op: "sort_list", Template: `sort\s+{FILLER}\s*{NAME}`, Captures: []string{"name"}},
	{Op: "shuffle_list", Template: `shuffle\s+{FILLER}\s*{NAME}`, Captures: []string{"name"}},
	{Op: "reverse_list", Template: `reverse\s+{FILLER}\s*{NAME}`, Captures: []string{"name"}},

	// FILE OPERATIONS AS STATEMENTS
	// These must come before LIST_ADD (append) and PRINT (write) which share those verbs.
	{Op: "writefile_stmt", Template: `write\s+"(?P<content>[^"]+)"\s+to\s+file\s+"(?P<filename>[^"]+)"`, Captures: []string{"content", "filename"}},
	{Op: "appendfile_stmt", Template: `append\s+"(?P<content>[^"]+)"\s+to\s+file\s+"(?P<filename>[^"]+)"`, Captures: []string{"content", "filename"}},
	{Op: "deletefile_stmt", Template: `delete\s+file\s+"(?P<filename>[^"]+)"`, Captures: []string{"filename"}},
	{Op: "runcmd_stmt", Template: `run\s+command\s+"(?P<command>[^"]+)"`, Captures: []string{"command"}},

	// LIST OPERATIONS
	{
		Op:       "list_def",
		Template: `{VERB_DECLARE}\s+{FILLER}\s*list\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_def",
		Template: `{VERB_DECLARE}\s+{FILLER}\s*{NAME}\s+as\s+{FILLER}\s*list`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_add",
		Template: `add\s+{VALUE}\s+to\s+{FILLER}\s*list\s+{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "list_add",
		Template: `append\s+{VALUE}\s+to\s+{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "list_add",
		Template: `push\s+{VALUE}\s+into\s+{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "list_add",
		Template: `put\s+{VALUE}\s+in\s+{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "list_rem",
		Template: `pop\s+from\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_rem",
		Template: `remove\s+last\s+from\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_rem",
		Template: `remove\s+first\s+from\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_rem",
		Template: `remove\s+{VALUE}\s+from\s+{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "list_get",
		Template: `get\s+{FILLER}\s*first\s+{FILLER}\s*item\s+{FILLER}\s*in\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_get",
		Template: `get\s+{FILLER}\s*last\s+{FILLER}\s*item\s+{FILLER}\s*in\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_get",
		Template: `first\s+item\s+in\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_get",
		Template: `last\s+item\s+in\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "list_get",
		Template: `get\s+{FILLER}\s*item\s+{INDEX}\s+from\s+{NAME}`,
		Captures: []string{"index", "name"},
	},
	{
		Op:       "list_get",
		Template: `get\s+{FILLER}\s*{INDEX}\s+item\s+in\s+{NAME}`,
		Captures: []string{"index", "name"},
	},
	{
		Op:       "list_get",
		Template: `{NAME}\s+at\s+{INDEX}`,
		Captures: []string{"name", "index"},
	},

	// INPUT
	{
		Op:       "input",
		Template: `{VERB_INPUT}\s+"(?P<prompt>[^"]+)"\s+(?:for|into|to)\s+{NAME}`,
		Captures: []string{"prompt", "name"},
	},
	{
		Op:       "input",
		Template: `{VERB_INPUT}\s+{FILLER}\s*for\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "input",
		Template: `{VERB_INPUT}\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "input",
		Template: `read\s+number\s+for\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "input",
		Template: `read\s+line\s+for\s+{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "delete",
		Template: `{VERB_DELETE}\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "add",
		Template: `{VERB_ADD}\s+{VALUE}\s+to\s+{FILLER}\s*{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "add",
		Template: `{VERB_ADD}\s+{FILLER}\s*{NAME}\s+by\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	{
		Op:       "subtract",
		Template: `{VERB_SUBTRACT}\s+{VALUE}\s+from\s+{FILLER}\s*{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "subtract",
		Template: `{VERB_SUBTRACT}\s+{FILLER}\s*{NAME}\s+by\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	{
		Op:       "multiply",
		Template: `{VERB_MULTIPLY}\s+{FILLER}\s*{NAME}\s+by\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	{
		Op:       "multiply",
		Template: `{NAME}\s+times\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	{
		Op:       "divide",
		Template: `{VERB_DIVIDE}\s+{FILLER}\s*{NAME}\s+by\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},

	// DECLARE
	// Declare must come before assign so "make a variable called x" doesn't
	// match the assign pattern first.
	{
		Op:       "declare",
		Template: `{VERB_DECLARE}\s+{FILLER}\s*{NAME}`,
		Captures: []string{"name"},
	},
	{
		Op:       "declare",
		Template: `{NAME}\s+is\s+{FILLER}\s*(?:a\s+)?(?:variable|val|var|number|thing)`,
		Captures: []string{"name"},
	},

	// ASSIGN
	{
		Op:       "assign",
		Template: `{VERB_ASSIGN}\s+{FILLER}\s*{NAME}\s+{PREP_TO}\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	{
		Op:       "assign",
		Template: `{VERB_ASSIGN}\s+{VALUE}\s+{PREP_TO}\s+{FILLER}\s*{NAME}`,
		Captures: []string{"value", "name"},
	},
	{
		Op:       "assign",
		Template: `{NAME}\s+=\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},
	// "total is a + b + c" or "result is x * 2"
	{
		Op:       "assign",
		Template: `{NAME}\s+is\s+{VALUE}`,
		Captures: []string{"name", "value"},
	},

	// PRINT
	{
		Op:       "print",
		Template: `{VERB_PRINT}\s+{VALUE}`,
		Captures: []string{"value"},
	},
}
