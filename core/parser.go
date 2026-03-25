package core

import (
	"fmt"
	"strconv"
	"strings"
)

type ParseError struct {
	Line    int
	Raw     string
	Message string
	Hint    string
}

func (e *ParseError) Error() string {
	msg := fmt.Sprintf("Line %d: %s\n  -> %s", e.Line, e.Message, e.Raw)
	if e.Hint != "" {
		msg += "\n  -> " + e.Hint
	}
	return msg
}

type Parser struct {
	matcher *Matcher
	dict    Dictionary
}

func NewParser(c *Compiler) *Parser {
	return &Parser{
		matcher: NewMatcher(c),
		dict:    c.Dict,
	}
}

func (p *Parser) Parse(source string) ([]Node, []error) {
	lines := Tokenize(source)
	flat, errs := p.matchLines(lines)
	if len(errs) > 0 {
		return nil, errs
	}
	nested, err := nestBlocks(flat)
	if err != nil {
		return nil, []error{err}
	}
	return nested, nil
}

func (p *Parser) matchLines(lines []Line) ([]Node, []error) {
	var nodes []Node
	var errs []error

	for _, ln := range lines {
		node, err := p.matchLine(ln)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes, errs
}

func (p *Parser) matchLine(ln Line) (Node, error) {
	res, err := p.matcher.Match(ln.Number, ln.Text)
	if err != nil {
		hint := FuzzySuggest(ln.Number, ln.Text, p.dict)
		return Node{}, &ParseError{
			Line:    ln.Number,
			Raw:     ln.Text,
			Message: "I don't understand this line.",
			Hint:    hint,
		}
	}
	return p.buildNode(res)
}

func (p *Parser) buildNode(res MatchResult) (Node, error) {
	c := res.Captures
	node := Node{
		Line: res.Line,
		Raw:  res.Raw,
	}

	switch res.Op {
	case "comment":
		node.Type = NodeComment

	case "declare":
		node.Type = NodeDeclare
		node.Name = c["name"]

	case "assign":
		node.Type = NodeAssign
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "print":
		node.Type = NodePrint
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "add":
		node.Type = NodeAdd
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "subtract":
		node.Type = NodeSubtract
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "multiply":
		node.Type = NodeMultiply
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "divide":
		node.Type = NodeDivide
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "loop":
		node.Type = NodeLoop
		n, err := strconv.ParseFloat(c["number"], 64)
		if err != nil {
			return node, p.wrapErr(res, fmt.Errorf("expected a number but got %q", c["number"]))
		}
		node.Value = n

	case "while":
		node.Type = NodeWhile
		node.Name = c["name"]
		node.Cmp = c["cmp"]
		if err := p.resolveRight(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "for":
		node.Type = NodeFor
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}
		// store the end value in Extra so the interpreter can read it
		lit, ref, _ := ParseValue(c["tovalue"], p.dict)
		if ref != "" {
			node.Extra = ref
		} else if lit != nil {
			node.Extra = fmt.Sprintf("%v", lit)
		}

	case "foreach":
		node.Type = NodeForEach
		node.Name = c["name"]
		node.Extra = c["listname"]

	case "if":
		node.Type = NodeIf
		node.Name = c["name"]
		node.Cmp = c["cmp"]
		if err := p.resolveRight(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "elseif":
		node.Type = NodeElseIf
		node.Name = c["name"]
		node.Cmp = c["cmp"]
		if err := p.resolveRight(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "else":
		node.Type = NodeElse

	case "end":
		node.Type = NodeEnd

	case "fn_def":
		node.Type = NodeFnDef
		node.Name = c["name"]
		if params, ok := c["params"]; ok && params != "" {
			node.Params = ParseParams(params)
		}

	case "fn_call":
		node.Type = NodeFnCall
		node.Name = c["name"]
		if args, ok := c["args"]; ok && args != "" {
			rawArgs := ParseArgs(args)
			for _, a := range rawArgs {
				lit, ref, _ := ParseValue(a, p.dict)
				if ref != "" {
					node.ArgRefs = append(node.ArgRefs, ref)
					node.Args = append(node.Args, nil)
				} else {
					node.ArgRefs = append(node.ArgRefs, "")
					node.Args = append(node.Args, lit)
				}
			}
		}

	case "return":
		node.Type = NodeReturn
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "delete":
		node.Type = NodeDelete
		node.Name = c["name"]

	case "exit_prog":
		node.Type = NodeBuiltin
		node.Name = "exit"
		code := float64(0)
		if s, ok := c["code"]; ok && s != "" {
			if n, err2 := strconv.ParseFloat(s, 64); err2 == nil {
				code = n
			}
		}
		node.Args = []interface{}{code}
		node.ArgRefs = []string{""}

	case "sleep_stmt":
		node.Type = NodeBuiltin
		node.Name = "sleep"
		secs := float64(1)
		if s, ok := c["seconds"]; ok && s != "" {
			if n, err2 := strconv.ParseFloat(s, 64); err2 == nil {
				secs = n
			}
		}
		node.Args = []interface{}{secs}
		node.ArgRefs = []string{""}

	case "writefile_stmt":
		node.Type = NodeBuiltin
		node.Name = "writefile"
		node.Args = []interface{}{c["content"], c["filename"]}
		node.ArgRefs = []string{"", ""}

	case "appendfile_stmt":
		node.Type = NodeBuiltin
		node.Name = "appendfile"
		node.Args = []interface{}{c["content"], c["filename"]}
		node.ArgRefs = []string{"", ""}

	case "deletefile_stmt":
		node.Type = NodeBuiltin
		node.Name = "deletefile"
		node.Args = []interface{}{c["filename"]}
		node.ArgRefs = []string{""}

	case "runcmd_stmt":
		node.Type = NodeBuiltin
		node.Name = "runcmd"
		node.Args = []interface{}{c["command"]}
		node.ArgRefs = []string{""}

	case "sort_list":
		node.Type = NodeBuiltin
		node.Name = "listsort"
		node.Args = []interface{}{nil}
		node.ArgRefs = []string{c["name"]}
		node.Extra = c["name"]

	case "sort_list_desc":
		node.Type = NodeBuiltin
		node.Name = "listsortdesc"
		node.Args = []interface{}{nil}
		node.ArgRefs = []string{c["name"]}
		node.Extra = c["name"]

	case "shuffle_list":
		node.Type = NodeBuiltin
		node.Name = "listshuffle"
		node.Args = []interface{}{nil}
		node.ArgRefs = []string{c["name"]}
		node.Extra = c["name"]

	case "reverse_list":
		node.Type = NodeBuiltin
		node.Name = "listreverse"
		node.Args = []interface{}{nil}
		node.ArgRefs = []string{c["name"]}
		node.Extra = c["name"]

	case "break":
		node.Type = NodeBreak

	case "continue":
		node.Type = NodeContinue

	case "input":
		node.Type = NodeInput
		node.Name = c["name"]
		if prompt, ok := c["prompt"]; ok {
			node.Extra = prompt
		}

	case "list_def":
		node.Type = NodeListDef
		node.Name = c["name"]

	case "list_add":
		node.Type = NodeListAdd
		node.Name = c["name"]
		if err := p.resolveValue(c["value"], &node); err != nil {
			return node, p.wrapErr(res, err)
		}

	case "list_rem":
		node.Type = NodeListRem
		node.Name = c["name"]
		if v, ok := c["value"]; ok && v != "" {
			if err := p.resolveValue(v, &node); err != nil {
				return node, p.wrapErr(res, err)
			}
		}

	case "list_get":
		node.Type = NodeListGet
		node.Name = c["name"]
		if idx, ok := c["index"]; ok && idx != "" {
			n, err := strconv.Atoi(idx)
			if err == nil {
				node.Index = n
			} else {
				node.IndexRef = idx
			}
		}
		// "first item" and "last item" translate to index 0 and -1
		raw := strings.ToLower(strings.TrimSpace(res.Raw))
		if strings.Contains(raw, "first") {
			node.Index = 0
		} else if strings.Contains(raw, "last") {
			node.Index = -1
		}

	default:
		node.Type = NodeComment // unknown — treat as no-op
	}

	return node, nil
}

func (p *Parser) resolveValue(raw string, node *Node) error {
	if raw == "" {
		return nil
	}
	lit, ref, err := ParseValue(raw, p.dict)
	if err != nil {
		return err
	}
	if ref != "" {
		node.ValueRef = ref
	} else {
		node.Value = lit
	}
	return nil
}

func (p *Parser) resolveRight(raw string, node *Node) error {
	if raw == "" {
		return nil
	}
	lit, ref, err := ParseValue(raw, p.dict)
	if err != nil {
		return err
	}
	if ref != "" {
		node.RightRef = ref
	} else {
		node.Right = lit
	}
	return nil
}

func (p *Parser) wrapErr(res MatchResult, err error) error {
	return &ParseError{
		Line:    res.Line,
		Raw:     res.Raw,
		Message: err.Error(),
		Hint:    "Check the values or variable names on this line.",
	}
}

type stackFrame struct {
	nodeType NodeType
	body     *[]Node
	node     *Node   // pointer to node in parent slice
	parent   *[]Node // slice we're appending to
}

// nestBlocks takes the flat list of nodes and nests them into a proper tree.
// loop/if/function bodies become children of their parent node.
func nestBlocks(flat []Node) ([]Node, error) {
	root := make([]Node, 0, len(flat))
	stack := []stackFrame{}

	current := func() *[]Node {
		if len(stack) == 0 {
			return &root
		}
		return stack[len(stack)-1].body
	}

	for i := range flat {
		n := flat[i]

		switch n.Type {
		case NodeLoop, NodeWhile, NodeFor, NodeForEach, NodeFnDef, NodeIf:
			*current() = append(*current(), n)
			pushed := &(*current())[len(*current())-1]
			pushed.Body = []Node{}
			stack = append(stack, stackFrame{
				nodeType: n.Type,
				body:     &pushed.Body,
				node:     pushed,
				parent:   current(),
			})

		case NodeElseIf:
			if len(stack) == 0 {
				return nil, &ParseError{Line: n.Line, Raw: n.Raw,
					Message: "'else if' without a matching 'if'.",
					Hint:    "Make sure every 'else if' follows an 'if' block."}
			}
			top := &stack[len(stack)-1]
			if top.nodeType != NodeIf && top.nodeType != NodeElseIf {
				return nil, &ParseError{Line: n.Line, Raw: n.Raw,
					Message: "'else if' does not follow an 'if' block.",
					Hint:    "Make sure the structure is: if ... else if ... else ... end"}
			}
			branch := ElseIfBranch{
				Cmp:  n.Cmp,
				Left: n.Name,
			}
			if n.Right != nil {
				branch.Right = n.Right
			}
			if n.RightRef != "" {
				branch.RightRef = n.RightRef
			}
			branch.Body = []Node{}
			top.node.ElseIfs = append(top.node.ElseIfs, branch)
			top.body = &top.node.ElseIfs[len(top.node.ElseIfs)-1].Body
			top.nodeType = NodeElseIf

		case NodeElse:
			if len(stack) == 0 {
				return nil, &ParseError{Line: n.Line, Raw: n.Raw,
					Message: "'else' without a matching 'if'.",
					Hint:    "Make sure every 'else' follows an 'if' block."}
			}
			top := &stack[len(stack)-1]
			if top.nodeType != NodeIf && top.nodeType != NodeElseIf {
				return nil, &ParseError{Line: n.Line, Raw: n.Raw,
					Message: "'else' does not follow an 'if' block.",
					Hint:    "Make sure the structure is: if ... else ... end"}
			}
			top.node.ElseBody = []Node{}
			top.body = &top.node.ElseBody
			top.nodeType = NodeElse

		case NodeEnd:
			if len(stack) == 0 {
				return nil, &ParseError{Line: n.Line, Raw: n.Raw,
					Message: "Extra 'end' — nothing to close here.",
					Hint:    "Remove the extra 'end' or check your block structure."}
			}
			stack = stack[:len(stack)-1]

		case NodeComment:
			// comments are dropped

		default:
			*current() = append(*current(), n)
		}
	}

	if len(stack) > 0 {
		top := stack[len(stack)-1]
		return nil, &ParseError{
			Line:    0,
			Raw:     "",
			Message: fmt.Sprintf("Block of type '%s' was never closed with 'end'.", top.nodeType),
			Hint:    "Add 'end' after the last line of your loop, if, or function block.",
		}
	}

	return root, nil
}
