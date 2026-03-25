package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"speek/core"
	"speek/runtime"
)

// REPL runs the interactive Read-Eval-Print Loop.
func REPL() {
	fmt.Println("Speek v0.1.0 - just type naturally")
	fmt.Println("Type 'exit' or 'quit' to leave. Type 'help' for commands.")
	fmt.Println()

	compiler, err := core.NewCompiler(core.DefaultDictionary, core.AllPatterns)
	if err != nil {
		fmt.Fprintln(os.Stderr, "speek: failed to initialize:", err)
		os.Exit(1)
	}

	parser := core.NewParser(compiler)
	scope := runtime.NewScope(nil)
	interp := runtime.NewInterpreter(false)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">>> ")

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Built-in REPL commands
		switch strings.ToLower(line) {
		case "exit", "quit", "bye", "goodbye":
			fmt.Println("Goodbye!")
			return
		case "help":
			PrintHelp("")
			continue
		case "help math":
			PrintHelp("math")
			continue
		case "help strings", "help string":
			PrintHelp("strings")
			continue
		case "help lists", "help list":
			PrintHelp("lists")
			continue
		case "vars", "variables", "show vars", "list vars":
			vars := scope.AllVars()
			if len(vars) == 0 {
				fmt.Println("  (no variables declared yet)")
			} else {
				for k, v := range vars {
					fmt.Printf("  %s = %s\n", k, v.String())
				}
			}
			continue
		case "clear", "reset":
			scope = runtime.NewScope(nil)
			fmt.Println("  Scope cleared.")
			continue
		}

		// Parse and run the line
		nodes, errs := parser.Parse(line)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintln(os.Stderr, "  Error:", e)
			}
			continue
		}

		if err := interp.Run(nodes, scope); err != nil {
			fmt.Fprintln(os.Stderr, " ", err)
			continue
		}

		// Show declare/assign feedback
		for _, n := range nodes {
			switch n.Type {
			case core.NodeDeclare:
				val, _ := scope.Get(n.Name)
				fmt.Printf("  declared: %s = %s\n", n.Name, val.String())
			case core.NodeAssign:
				val, _ := scope.Get(n.Name)
				fmt.Printf("  %s = %s\n", n.Name, val.String())
			case core.NodeListDef:
				fmt.Printf("  declared list: %s = []\n", n.Name)
			}
		}
	}
}
