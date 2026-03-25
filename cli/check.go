package cli

import (
	"fmt"
	"os"

	"speek/core"
)

func Check(filename string) int {
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
	_, errs := parser.Parse(string(data))

	if len(errs) == 0 {
		fmt.Printf("speek: '%s' looks good. No errors found.\n", filename)
		return 0
	}

	fmt.Printf("speek: found %d error(s) in '%s':\n\n", len(errs), filename)
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, e)
		fmt.Fprintln(os.Stderr)
	}
	return 1
}
