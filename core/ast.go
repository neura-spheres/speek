package core

type NodeType string

const (
	NodeDeclare  NodeType = "declare"
	NodeAssign   NodeType = "assign"
	NodePrint    NodeType = "print"
	NodeAdd      NodeType = "add"
	NodeSubtract NodeType = "subtract"
	NodeMultiply NodeType = "multiply"
	NodeDivide   NodeType = "divide"
	NodeModulo   NodeType = "modulo"
	NodeLoop     NodeType = "loop"
	NodeWhile    NodeType = "while"
	NodeFor      NodeType = "for"
	NodeForEach  NodeType = "foreach"
	NodeIf       NodeType = "if"
	NodeElseIf   NodeType = "elseif"
	NodeElse     NodeType = "else"
	NodeEnd      NodeType = "end"
	NodeFnDef    NodeType = "fn_def"
	NodeFnCall   NodeType = "fn_call"
	NodeReturn   NodeType = "return"
	NodeDelete   NodeType = "delete"
	NodeBreak    NodeType = "break"
	NodeContinue NodeType = "continue"
	NodeInput    NodeType = "input"
	NodeBuiltin  NodeType = "builtin"
	NodeComment  NodeType = "comment"
	NodeListDef  NodeType = "list_def"
	NodeListAdd  NodeType = "list_add"
	NodeListGet  NodeType = "list_get"
	NodeListRem  NodeType = "list_rem"
)

type Node struct {
	Type     NodeType
	Name     string
	Value    interface{} // literal value (string, float64, bool, nil)
	ValueRef string      // variable name if the value comes from a variable
	Args     []interface{}
	ArgRefs  []string
	Params   []string
	Cmp      string // comparison op: gt, lt, gte, lte, eq, neq, divisible
	Right    interface{}
	RightRef string
	Body     []Node
	ElseBody []Node
	ElseIfs  []ElseIfBranch
	Index    interface{}
	IndexRef string
	Extra    string // catch-all for misc data (loop end value, input prompt, etc.)
	Line     int    // 1-based line number, used in error messages
	Raw      string // original source line, used in error messages
	Indent   int    // leading whitespace depth of the source line
}

type ElseIfBranch struct {
	Cmp      string
	Left     string
	LeftVal  interface{}
	Right    interface{}
	RightRef string
	Body     []Node
}
