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
	text := ln.Text
	if strings.HasSuffix(text, ":") && !strings.HasSuffix(text, `":`) {
		text = strings.TrimSpace(text[:len(text)-1])
	}

	res, err := p.matcher.Match(ln.Number, text)
	if err != nil {
		hint := FuzzySuggest(ln.Number, ln.Text, p.dict)
		return Node{}, &ParseError{
			Line:    ln.Number,
			Raw:     ln.Text,
			Message: "I don't understand this line.",
			Hint:    hint,
		}
	}
	node, err := p.buildNode(res)
	if err != nil {
		return Node{}, err
	}
	node.Indent = ln.Indent
	return node, nil
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
		node.Type = NodeComment
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

// nestBlocks takes the flat list of nodes and nests them into a proper AST tree
// using indentation to determine parent-child relationships.
func nestBlocks(flat []Node) ([]Node, error) {
	result, _, err := nestLevel(flat, 0, -1)
	return result, err
}

func isBlockOpener(t NodeType) bool {
	switch t {
	case NodeLoop, NodeWhile, NodeFor, NodeForEach, NodeFnDef:
		return true
	}
	return false
}

// nestLevel recursively collects nodes whose Indent is strictly greater than
// parentIndent (pass -1 for the root call to accept everything). It returns
// the collected node slice and the index of the first unconsumed node.
func nestLevel(flat []Node, idx int, parentIndent int) ([]Node, int, error) {
	var result []Node

	for idx < len(flat) {
		n := flat[idx]

		// Stop when we de-indent back to or past the parent's level.
		if parentIndent >= 0 && n.Indent <= parentIndent {
			break
		}
		idx++

		switch n.Type {
		case NodeComment:
			// drop

		case NodeEnd:
			// Silently ignore — indentation now handles block scoping.

		case NodeElse, NodeElseIf:
			// These are only valid as continuations of an if block and should
			// be consumed by the NodeIf case below, never appear standalone.
			return nil, idx, &ParseError{
				Line:    n.Line,
				Raw:     n.Raw,
				Message: fmt.Sprintf("'%s' without a matching 'if'.", n.Type),
				Hint:    "Make sure 'else' / 'else if' follows an indented 'if' block.",
			}

		case NodeIf:
			n.Body = []Node{}
			var err error
			n.Body, idx, err = nestLevel(flat, idx, n.Indent)
			if err != nil {
				return nil, 0, err
			}

			// Consume any else-if / else branches at the same indent level.
			for idx < len(flat) {
				next := flat[idx]
				if next.Indent != n.Indent {
					break
				}
				if next.Type == NodeElseIf {
					idx++
					branch := ElseIfBranch{
						Cmp:      next.Cmp,
						Left:     next.Name,
						Right:    next.Right,
						RightRef: next.RightRef,
						Body:     []Node{},
					}
					branch.Body, idx, err = nestLevel(flat, idx, n.Indent)
					if err != nil {
						return nil, 0, err
					}
					n.ElseIfs = append(n.ElseIfs, branch)
				} else if next.Type == NodeElse {
					idx++
					n.ElseBody = []Node{}
					n.ElseBody, idx, err = nestLevel(flat, idx, n.Indent)
					if err != nil {
						return nil, 0, err
					}
					break
				} else {
					break
				}
			}

			result = append(result, n)

		default:
			if isBlockOpener(n.Type) {
				n.Body = []Node{}
				var err error
				n.Body, idx, err = nestLevel(flat, idx, n.Indent)
				if err != nil {
					return nil, 0, err
				}
			}
			result = append(result, n)
		}
	}

	return result, idx, nil
}
