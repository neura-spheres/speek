package runtime

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"speek/core"
)

// these three are used as panic values so we can unwind the call stack
// without actually crashing -- the loop/function handlers catch them
type returnSignal struct{ val Value }
type breakSignal struct{}
type continueSignal struct{}

const maxCallDepth = 500

type Interpreter struct {
	Debug     bool
	reader    *bufio.Reader
	callDepth int
}

func NewInterpreter(debug bool) *Interpreter {
	return &Interpreter{
		Debug:  debug,
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run is the external entry point. It catches any unhandled control-flow signals
// (break/continue outside a loop, return outside a function) and turns them into errors.
func (interp *Interpreter) Run(nodes []core.Node, scope *Scope) (rerr error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case breakSignal, continueSignal:
				rerr = fmt.Errorf("'break' or 'continue' can only be used inside a loop")
			case returnSignal:
				rerr = fmt.Errorf("'return' can only be used inside a function")
			default:
				rerr = fmt.Errorf("internal error: %v", r)
			}
		}
	}()
	return interp.runBlock(nodes, scope)
}

// runBlock executes nodes without any recovery. Used for block bodies (if/else, function bodies)
// so that break/continue/return signals can propagate up to their proper handlers.
func (interp *Interpreter) runBlock(nodes []core.Node, scope *Scope) error {
	for i := range nodes {
		if err := interp.exec(nodes[i], scope); err != nil {
			return err
		}
	}
	return nil
}

func (interp *Interpreter) exec(n core.Node, scope *Scope) (rerr error) {
	// re-panic return/break/continue so loop and fn handlers can catch them
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case returnSignal, breakSignal, continueSignal:
				panic(v) // re-panic so the loop/fn handler catches it
			default:
				rerr = fmt.Errorf("internal error on line %d: %v", n.Line, r)
			}
		}
	}()

	switch n.Type {
	case core.NodeComment:
		// no-op

	case core.NodeDeclare:
		if err := scope.Declare(n.Name, Null); err != nil {
			return ErrAlreadyDeclared(n.Line, n.Raw, n.Name)
		}

	case core.NodeAssign:
		val, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		scope.Set(n.Name, val)

	case core.NodePrint:
		val, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		fmt.Println(val.String())

	case core.NodeAdd:
		cur, err := interp.getVar(n.Name, n, scope)
		if err != nil {
			return err
		}
		operand, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		// lists use "add" to append, not to do math
		if cur.Type == TypeList {
			cur.List = append(cur.List, operand)
			scope.Set(n.Name, cur)
		} else if cur.Type == TypeString || operand.Type == TypeString {
			scope.Set(n.Name, StringVal(cur.AsString()+operand.AsString()))
		} else {
			scope.Set(n.Name, NumberVal(cur.AsNumber()+operand.AsNumber()))
		}

	case core.NodeSubtract:
		cur, err := interp.getVar(n.Name, n, scope)
		if err != nil {
			return err
		}
		operand, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		scope.Set(n.Name, NumberVal(cur.AsNumber()-operand.AsNumber()))

	case core.NodeMultiply:
		cur, err := interp.getVar(n.Name, n, scope)
		if err != nil {
			return err
		}
		operand, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		scope.Set(n.Name, NumberVal(cur.AsNumber()*operand.AsNumber()))

	case core.NodeDivide:
		cur, err := interp.getVar(n.Name, n, scope)
		if err != nil {
			return err
		}
		operand, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		if operand.AsNumber() == 0 {
			return ErrDivZero(n.Line, n.Raw, scope.AllVars())
		}
		scope.Set(n.Name, NumberVal(cur.AsNumber()/operand.AsNumber()))

	case core.NodeLoop:
		count := int(n.Value.(float64))
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(breakSignal); ok {
						return
					}
					panic(r)
				}
			}()
			for i := 0; i < count; i++ {
				child := NewScope(scope)
				// __i__ holds the loop counter, named weird on purpose so it won't clash with user vars
				child.Set("__i__", NumberVal(float64(i)))
				cont := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							if _, ok := r.(continueSignal); ok {
								cont = true
								return
							}
							panic(r)
						}
					}()
					for _, body := range n.Body {
						if err := interp.exec(body, child); err != nil {
							panic(err)
						}
					}
				}()
				if cont {
					continue
				}
			}
		}()

	case core.NodeWhile:
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(breakSignal); ok {
						return
					}
					panic(r)
				}
			}()
			for {
				cond, err := interp.evalCond(n, scope)
				if err != nil {
					panic(err)
				}
				if !cond {
					break
				}
				child := NewScope(scope)
				cont := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							if _, ok := r.(continueSignal); ok {
								cont = true
								return
							}
							panic(r)
						}
					}()
					for _, body := range n.Body {
						if err := interp.exec(body, child); err != nil {
							panic(err)
						}
					}
				}()
				if cont {
					continue
				}
			}
		}()

	case core.NodeFor:
		startVal, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		endVal, err := interp.resolveExtraAsValue(n, scope)
		if err != nil {
			return err
		}
		start := startVal.AsNumber()
		end := endVal.AsNumber()
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(breakSignal); ok {
						return
					}
					panic(r)
				}
			}()
			for i := start; i <= end; i++ {
				child := NewScope(scope)
				child.Set(n.Name, NumberVal(i))
				cont := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							if _, ok := r.(continueSignal); ok {
								cont = true
								return
							}
							panic(r)
						}
					}()
					for _, body := range n.Body {
						if err := interp.exec(body, child); err != nil {
							panic(err)
						}
					}
				}()
				if cont {
					continue
				}
			}
		}()

	case core.NodeForEach:
		listVal, ok := scope.Get(n.Extra)
		if !ok {
			return ErrUndeclared(n.Line, n.Raw, n.Extra)
		}
		if listVal.Type != TypeList {
			return ErrNotAList(n.Line, n.Raw, n.Extra, listVal.Type)
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(breakSignal); ok {
						return
					}
					panic(r)
				}
			}()
			for _, item := range listVal.List {
				child := NewScope(scope)
				child.Set(n.Name, item)
				cont := false
				func() {
					defer func() {
						if r := recover(); r != nil {
							if _, ok := r.(continueSignal); ok {
								cont = true
								return
							}
							panic(r)
						}
					}()
					for _, body := range n.Body {
						if err := interp.exec(body, child); err != nil {
							panic(err)
						}
					}
				}()
				if cont {
					continue
				}
			}
		}()

	case core.NodeIf:
		cond, err := interp.evalCond(n, scope)
		if err != nil {
			return err
		}
		if cond {
			child := NewScope(scope)
			return interp.runBlock(n.Body, child)
		}
		for _, eib := range n.ElseIfs {
			branchCond, err := interp.evalElseIfCond(eib, scope)
			if err != nil {
				return err
			}
			if branchCond {
				child := NewScope(scope)
				return interp.runBlock(eib.Body, child)
			}
		}
		if len(n.ElseBody) > 0 {
			child := NewScope(scope)
			return interp.runBlock(n.ElseBody, child)
		}

	case core.NodeElse, core.NodeElseIf, core.NodeEnd:
		// handled during block nesting, won't reach here

	case core.NodeFnDef:
		scope.RegisterFn(n.Name, FnDef{
			Params: n.Params,
			Body:   n.Body,
		})

	case core.NodeFnCall:
		_, err := interp.callFn(n, scope)
		return err

	case core.NodeReturn:
		val, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		panic(returnSignal{val})

	case core.NodeDelete:
		if err := scope.Delete(n.Name); err != nil {
			return ErrUndeclared(n.Line, n.Raw, n.Name)
		}

	case core.NodeBreak:
		panic(breakSignal{})

	case core.NodeContinue:
		panic(continueSignal{})

	case core.NodeInput:
		fmt.Print(interp.inputPrompt(n))
		line, _ := interp.reader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if f, err := strconv.ParseFloat(strings.TrimSpace(line), 64); err == nil {
			scope.Set(n.Name, NumberVal(f))
		} else {
			scope.Set(n.Name, StringVal(line))
		}

	case core.NodeListDef:
		if err := scope.Declare(n.Name, ListVal([]Value{})); err != nil {
			return ErrAlreadyDeclared(n.Line, n.Raw, n.Name)
		}

	case core.NodeListAdd:
		lst, ok := scope.Get(n.Name)
		if !ok {
			return ErrUndeclared(n.Line, n.Raw, n.Name)
		}
		if lst.Type != TypeList {
			return ErrNotAList(n.Line, n.Raw, n.Name, lst.Type)
		}
		val, err := interp.resolveValue(n, scope)
		if err != nil {
			return err
		}
		lst.List = append(lst.List, val)
		scope.Set(n.Name, lst)

	case core.NodeListGet:
		lst, ok := scope.Get(n.Name)
		if !ok {
			return ErrUndeclared(n.Line, n.Raw, n.Name)
		}
		if lst.Type != TypeList {
			return ErrNotAList(n.Line, n.Raw, n.Name, lst.Type)
		}
		idx, err := interp.resolveIndex(n, scope, lst)
		if err != nil {
			return err
		}
		fmt.Println(lst.List[idx].String())

	case core.NodeListRem:
		lst, ok := scope.Get(n.Name)
		if !ok {
			return ErrUndeclared(n.Line, n.Raw, n.Name)
		}
		if lst.Type != TypeList {
			return ErrNotAList(n.Line, n.Raw, n.Name, lst.Type)
		}
		if n.Value == nil && n.ValueRef == "" {
			// no value specified = pop the last item
			if len(lst.List) > 0 {
				lst.List = lst.List[:len(lst.List)-1]
			}
		} else {
			val, err := interp.resolveValue(n, scope)
			if err != nil {
				return err
			}
			newList := []Value{}
			for _, v := range lst.List {
				if !v.Equals(val) {
					newList = append(newList, v)
				}
			}
			lst.List = newList
		}
		scope.Set(n.Name, lst)

	case core.NodeBuiltin:
		result, err := interp.callBuiltinNode(n, scope)
		if err != nil {
			return err
		}
		// Extra holds the variable name to write the result back to (used for in-place list ops)
		if n.Extra != "" {
			scope.Set(n.Extra, result)
		}
	}

	return nil
}

// resolveValue figures out what value a node is pointing to.
// lookup order: scope variable -> arithmetic expr -> builtin call -> user fn -> error
func (interp *Interpreter) resolveValue(n core.Node, scope *Scope) (Value, error) {
	if n.ValueRef != "" {
		if val, ok := scope.Get(n.ValueRef); ok {
			return val, nil
		}
		if v, ok := EvalExpr(n.ValueRef, scope); ok {
			return v, nil
		}
		if v, ok := interp.tryBuiltinExpression(n.ValueRef, scope); ok {
			return v, nil
		}
		if v, found, err := interp.tryUserFnExpression(n.ValueRef, n, scope); found {
			return v, err
		}
		return Null, ErrUndeclared(n.Line, n.Raw, n.ValueRef)
	}
	if n.Value == nil {
		return Null, nil
	}
	return literalToValue(n.Value), nil
}

func (interp *Interpreter) resolveRight(n core.Node, scope *Scope) (Value, error) {
	if n.RightRef != "" {
		val, ok := scope.Get(n.RightRef)
		if !ok {
			return Null, ErrUndeclared(n.Line, n.Raw, n.RightRef)
		}
		return val, nil
	}
	if n.Right == nil {
		return Null, nil
	}
	return literalToValue(n.Right), nil
}

func (interp *Interpreter) resolveExtraAsValue(n core.Node, scope *Scope) (Value, error) {
	if n.Extra == "" {
		return NumberVal(0), nil
	}
	if val, ok := scope.Get(n.Extra); ok {
		return val, nil
	}
	if f, err := strconv.ParseFloat(n.Extra, 64); err == nil {
		return NumberVal(f), nil
	}
	return StringVal(n.Extra), nil
}

func (interp *Interpreter) getVar(name string, n core.Node, scope *Scope) (Value, error) {
	val, ok := scope.Get(name)
	if !ok {
		return Null, ErrUndeclared(n.Line, n.Raw, name)
	}
	return val, nil
}

func literalToValue(v interface{}) Value {
	switch x := v.(type) {
	case float64:
		return NumberVal(x)
	case string:
		return StringVal(x)
	case bool:
		return BoolVal(x)
	case int:
		return NumberVal(float64(x))
	}
	return Null
}

func (interp *Interpreter) evalCond(n core.Node, scope *Scope) (bool, error) {
	left, ok := scope.Get(n.Name)
	if !ok {
		return false, ErrUndeclared(n.Line, n.Raw, n.Name)
	}
	right, err := interp.resolveRight(n, scope)
	if err != nil {
		return false, err
	}
	return compare(left, n.Cmp, right, n.Line, n.Raw)
}

func (interp *Interpreter) evalElseIfCond(branch core.ElseIfBranch, scope *Scope) (bool, error) {
	left, ok := scope.Get(branch.Left)
	if !ok {
		return false, &RuntimeError{Message: fmt.Sprintf("'%s' hasn't been declared.", branch.Left)}
	}
	var right Value
	if branch.RightRef != "" {
		var ok2 bool
		right, ok2 = scope.Get(branch.RightRef)
		if !ok2 {
			return false, &RuntimeError{Message: fmt.Sprintf("'%s' hasn't been declared.", branch.RightRef)}
		}
	} else if branch.Right != nil {
		right = literalToValue(branch.Right)
	}
	return compare(left, branch.Cmp, right, 0, "")
}

func compare(left Value, cmp string, right Value, line int, raw string) (bool, error) {
	switch cmp {
	case "gt":
		return left.AsNumber() > right.AsNumber(), nil
	case "lt":
		return left.AsNumber() < right.AsNumber(), nil
	case "gte":
		return left.AsNumber() >= right.AsNumber(), nil
	case "lte":
		return left.AsNumber() <= right.AsNumber(), nil
	case "eq":
		return left.Equals(right), nil
	case "neq":
		return !left.Equals(right), nil
	case "divisible":
		divisor := right.AsNumber()
		if divisor == 0 {
			return false, nil
		}
		return math.Mod(left.AsNumber(), divisor) == 0, nil
	default:
		return left.IsTruthy(), nil
	}
}

func (interp *Interpreter) callFn(n core.Node, scope *Scope) (Value, error) {
	if canon, ok := ResolveBuiltinAlias(n.Name); ok {
		args, err := interp.resolveArgs(n, scope)
		if err != nil {
			return Null, err
		}
		result, err := CallBuiltin(canon, args)
		if err != nil {
			return Null, &RuntimeError{Line: n.Line, Raw: n.Raw, Message: err.Error()}
		}
		return result, nil
	}

	fn, ok := scope.GetFn(n.Name)
	if !ok {
		return Null, ErrUnknownFn(n.Line, n.Raw, n.Name)
	}

	if interp.callDepth >= maxCallDepth {
		return Null, &RuntimeError{Line: n.Line, Raw: n.Raw, Message: fmt.Sprintf("max call depth (%d) exceeded — possible infinite recursion in '%s'", maxCallDepth, n.Name)}
	}
	interp.callDepth++
	defer func() { interp.callDepth-- }()

	args, err := interp.resolveArgs(n, scope)
	if err != nil {
		return Null, err
	}

	params := fn.Params
	if len(args) != len(params) && len(params) > 0 {
		return Null, ErrWrongArgCount(n.Line, n.Raw, n.Name, len(params), len(args))
	}

	child := NewScope(scope)
	for i, p := range params {
		if i < len(args) {
			child.Set(p, args[i])
		}
	}

	body, ok2 := fn.Body.([]core.Node)
	if !ok2 {
		return Null, &RuntimeError{Line: n.Line, Raw: n.Raw, Message: "corrupted function body"}
	}

	var retVal Value
	var fnErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				if sig, ok := r.(returnSignal); ok {
					retVal = sig.val
				} else {
					panic(r)
				}
			}
		}()
		fnErr = interp.runBlock(body, child)
	}()
	if fnErr != nil {
		return Null, fnErr
	}
	return retVal, nil
}

func (interp *Interpreter) callBuiltinNode(n core.Node, scope *Scope) (Value, error) {
	args, err := interp.resolveArgs(n, scope)
	if err != nil {
		return Null, err
	}
	result, err := CallBuiltin(n.Name, args)
	if err != nil {
		return Null, &RuntimeError{Line: n.Line, Raw: n.Raw, Message: err.Error()}
	}
	return result, nil
}

func (interp *Interpreter) resolveArgs(n core.Node, scope *Scope) ([]Value, error) {
	args := make([]Value, len(n.Args))
	for i := range n.Args {
		if n.ArgRefs[i] != "" {
			if v, ok := interp.tryBuiltinExpression(n.ArgRefs[i], scope); ok {
				args[i] = v
				continue
			}
			val, ok := scope.Get(n.ArgRefs[i])
			if !ok {
				return nil, ErrUndeclared(n.Line, n.Raw, n.ArgRefs[i])
			}
			args[i] = val
		} else {
			args[i] = literalToValue(n.Args[i])
		}
	}
	return args, nil
}

// tryBuiltinExpression checks if expr is something like "square root of x"
// and evaluates it. tries progressively longer word prefixes until one matches.
func (interp *Interpreter) tryBuiltinExpression(expr string, scope *Scope) (Value, bool) {
	expr = strings.TrimSpace(expr)
	words := strings.Fields(expr)
	for prefixLen := len(words) - 1; prefixLen >= 1; prefixLen-- {
		prefix := strings.Join(words[:prefixLen], " ")
		suffix := strings.Join(words[prefixLen:], " ")
		if canon, ok := ResolveBuiltinAlias(prefix); ok {
			// Resolve suffix as args
			args := parseBuiltinArgs(suffix, scope)
			result, err := CallBuiltin(canon, args)
			if err == nil {
				return result, true
			}
		}
	}

	// also try the whole thing as a constant like "pi" or "e"
	if canon, ok := ResolveBuiltinAlias(expr); ok {
		result, err := CallBuiltin(canon, nil)
		if err == nil {
			return result, true
		}
	}

	return Null, false
}

// tryUserFnExpression handles inline calls like "double with 5" in value position.
// Returns (value, found, error). If found=false the caller should try other strategies.
// If found=true but error!=nil, the function was matched but failed at runtime.
func (interp *Interpreter) tryUserFnExpression(expr string, n core.Node, scope *Scope) (Value, bool, error) {
	expr = strings.TrimSpace(expr)

	var fnName, argsStr string
	lower := strings.ToLower(expr)

	for _, sep := range []string{" with ", " using "} {
		if idx := strings.Index(lower, sep); idx != -1 {
			fnName = strings.TrimSpace(expr[:idx])
			argsStr = strings.TrimSpace(expr[idx+len(sep):])
			break
		}
	}
	if fnName == "" {
		fnName = expr
	}

	// function names are single words
	if strings.ContainsAny(fnName, " \t") {
		return Null, false, nil
	}

	if _, ok := scope.GetFn(fnName); !ok {
		return Null, false, nil
	}

	callNode := core.Node{Line: n.Line, Raw: n.Raw, Name: fnName}

	if argsStr != "" {
		parts := andSplitter.Split(strings.TrimSpace(argsStr), -1)
		callNode.Args = make([]interface{}, len(parts))
		callNode.ArgRefs = make([]string, len(parts))
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if f, err := strconv.ParseFloat(p, 64); err == nil {
				callNode.Args[i] = f
			} else if (strings.HasPrefix(p, `"`) && strings.HasSuffix(p, `"`)) ||
				(strings.HasPrefix(p, `'`) && strings.HasSuffix(p, `'`)) {
				callNode.Args[i] = p[1 : len(p)-1]
			} else {
				callNode.ArgRefs[i] = p
				callNode.Args[i] = nil
			}
		}
	} else {
		callNode.Args = []interface{}{}
		callNode.ArgRefs = []string{}
	}

	result, err := interp.callFn(callNode, scope)
	if err != nil {
		return Null, true, err
	}
	return result, true, nil
}

// andSplitter splits argument strings like "x and y" or "x, y"
var andSplitter = regexp.MustCompile(`\s+and\s+|\s*,\s*`)

func parseBuiltinArgs(suffix string, scope *Scope) []Value {
	if suffix == "" {
		return nil
	}
	parts := andSplitter.Split(strings.TrimSpace(suffix), -1)
	vals := make([]Value, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, ok := scope.Get(p); ok {
			vals = append(vals, v)
			continue
		}
		if f, err := strconv.ParseFloat(p, 64); err == nil {
			vals = append(vals, NumberVal(f))
			continue
		}
		if (strings.HasPrefix(p, `"`) && strings.HasSuffix(p, `"`)) ||
			(strings.HasPrefix(p, `'`) && strings.HasSuffix(p, `'`)) {
			vals = append(vals, StringVal(p[1:len(p)-1]))
			continue
		}
		vals = append(vals, StringVal(p))
	}
	return vals
}

func (interp *Interpreter) resolveIndex(n core.Node, scope *Scope, lst Value) (int, error) {
	length := len(lst.List)
	if length == 0 {
		return 0, &RuntimeError{Line: n.Line, Raw: n.Raw, Message: "The list is empty."}
	}

	if n.IndexRef != "" {
		val, ok := scope.Get(n.IndexRef)
		if !ok {
			return 0, ErrUndeclared(n.Line, n.Raw, n.IndexRef)
		}
		idx := int(val.AsNumber())
		if idx < 0 {
			idx = length + idx
		}
		if idx < 0 || idx >= length {
			return 0, ErrOutOfRange(n.Line, n.Raw, idx, length)
		}
		return idx, nil
	}

	if n.Index == nil {
		return 0, nil
	}

	idx := 0
	switch v := n.Index.(type) {
	case int:
		idx = v
	case float64:
		idx = int(v)
	}
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return 0, ErrOutOfRange(n.Line, n.Raw, idx, length)
	}
	return idx, nil
}

func (interp *Interpreter) inputPrompt(n core.Node) string {
	if n.Extra != "" {
		return n.Extra + " "
	}
	return fmt.Sprintf("Enter %s: ", n.Name)
}
