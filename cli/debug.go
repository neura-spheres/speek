package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"speek/core"
)

// ShowDebugPanel prints the intent panel before execution.
// For each line it shows: raw source → interpreted operation.
func ShowDebugPanel(nodes []core.Node, source string) {
	lines := core.Tokenize(source)

	fmt.Println("=== Speek Debug Panel ===")
	fmt.Printf("%-4s  %-40s  %s\n", "Line", "Source", "Interpretation")
	fmt.Println(strings.Repeat("-", 80))

	// Build a quick map from line number to node
	lineToNode := make(map[int]core.Node)
	flattenNodes(nodes, lineToNode)

	for _, ln := range lines {
		interp := "???"
		if n, ok := lineToNode[ln.Number]; ok {
			interp = describeNode(n)
		}
		src := ln.Text
		if len(src) > 38 {
			src = src[:35] + "..."
		}
		fmt.Printf("%-4d  %-40s  %s\n", ln.Number, src, interp)
	}
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()
}

func flattenNodes(nodes []core.Node, out map[int]core.Node) {
	for _, n := range nodes {
		if n.Line > 0 {
			out[n.Line] = n
		}
		flattenNodes(n.Body, out)
		flattenNodes(n.ElseBody, out)
		for _, eib := range n.ElseIfs {
			flattenNodes(eib.Body, out)
		}
	}
}

func describeNode(n core.Node) string {
	switch n.Type {
	case core.NodeDeclare:
		return fmt.Sprintf("DECLARE %s = null", n.Name)
	case core.NodeAssign:
		if n.ValueRef != "" {
			return fmt.Sprintf("ASSIGN %s <- %s", n.Name, n.ValueRef)
		}
		return fmt.Sprintf("ASSIGN %s <- %v", n.Name, n.Value)
	case core.NodePrint:
		if n.ValueRef != "" {
			return fmt.Sprintf("PRINT %s", n.ValueRef)
		}
		return fmt.Sprintf("PRINT %v", n.Value)
	case core.NodeAdd:
		return fmt.Sprintf("ADD %v to %s", valueDesc(n), n.Name)
	case core.NodeSubtract:
		return fmt.Sprintf("SUBTRACT %v from %s", valueDesc(n), n.Name)
	case core.NodeMultiply:
		return fmt.Sprintf("MULTIPLY %s by %v", n.Name, valueDesc(n))
	case core.NodeDivide:
		return fmt.Sprintf("DIVIDE %s by %v", n.Name, valueDesc(n))
	case core.NodeLoop:
		return fmt.Sprintf("LOOP %v times", n.Value)
	case core.NodeWhile:
		return fmt.Sprintf("WHILE %s %s %v", n.Name, n.Cmp, rightDesc(n))
	case core.NodeFor:
		return fmt.Sprintf("FOR %s from %v to %s", n.Name, valueDesc(n), n.Extra)
	case core.NodeForEach:
		return fmt.Sprintf("FOR EACH %s IN %s", n.Name, n.Extra)
	case core.NodeIf:
		return fmt.Sprintf("IF %s %s %v", n.Name, n.Cmp, rightDesc(n))
	case core.NodeElseIf:
		return fmt.Sprintf("ELSE IF %s %s %v", n.Name, n.Cmp, rightDesc(n))
	case core.NodeElse:
		return "ELSE"
	case core.NodeEnd:
		return "END"
	case core.NodeFnDef:
		return fmt.Sprintf("DEFINE fn %s(%s)", n.Name, strings.Join(n.Params, ", "))
	case core.NodeFnCall:
		return fmt.Sprintf("CALL %s", n.Name)
	case core.NodeReturn:
		return fmt.Sprintf("RETURN %v", valueDesc(n))
	case core.NodeDelete:
		return fmt.Sprintf("DELETE %s", n.Name)
	case core.NodeBreak:
		return "BREAK"
	case core.NodeContinue:
		return "CONTINUE"
	case core.NodeInput:
		return fmt.Sprintf("INPUT -> %s", n.Name)
	case core.NodeListDef:
		return fmt.Sprintf("DECLARE list %s = []", n.Name)
	case core.NodeListAdd:
		return fmt.Sprintf("LIST %s.append(%v)", n.Name, valueDesc(n))
	case core.NodeListGet:
		return fmt.Sprintf("LIST %s[%v]", n.Name, n.Index)
	case core.NodeListRem:
		return fmt.Sprintf("LIST %s.remove(%v)", n.Name, valueDesc(n))
	case core.NodeComment:
		return "# comment"
	default:
		return string(n.Type)
	}
}

func valueDesc(n core.Node) string {
	if n.ValueRef != "" {
		return n.ValueRef
	}
	return fmt.Sprintf("%v", n.Value)
}

func rightDesc(n core.Node) string {
	if n.RightRef != "" {
		return n.RightRef
	}
	return fmt.Sprintf("%v", n.Right)
}

// ConfirmInterpretation asks the user to confirm interpretation of ambiguous lines.
func ConfirmInterpretation(nodes []core.Node) bool {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Confirm this interpretation? [y/N] ")
	scanner.Scan()
	input := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return input == "y" || input == "yes"
}
