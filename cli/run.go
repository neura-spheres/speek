package cli

import (
	"fmt"
	"os"

	"speek/core"
	"speek/runtime"
)

func Run(filename string, debug bool, confirm bool) int {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "speek: cannot open '%s': %v\n", filename, err)
		return 1
	}

	compiler, err := core.NewCompiler(core.DefaultDictionary, core.AllPatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "speek: internal compiler error: %v\n", err)
		return 1
	}

	parser := core.NewParser(compiler)
	nodes, errs := parser.Parse(string(data))
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		return 1
	}

	if debug {
		ShowDebugPanel(nodes, string(data))
	}

	if confirm {
		if !ConfirmInterpretation(nodes) {
			fmt.Println("Execution cancelled.")
			return 0
		}
	}

	interp := runtime.NewInterpreter(debug)
	scope := runtime.NewScope(nil)

	if err := interp.Run(nodes, scope); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
