package main

import (
	"fmt"
	"os"
	"strings"

	"speek/cli"
)

const version = "v0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	cmd := strings.ToLower(args[0])

	switch cmd {
	case "version", "--version", "-v":
		fmt.Printf("Speek %s\n", version)
		fmt.Println("Natural language programming, no AI, no internet, forever free.")

	case "help", "--help", "-h":
		topic := ""
		if len(args) > 1 {
			topic = strings.ToLower(args[1])
		}
		cli.PrintHelp(topic)

	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "speek run: missing filename")
			fmt.Fprintln(os.Stderr, "Usage: speek run <file.spk> [--debug] [--confirm]")
			os.Exit(1)
		}
		filename := args[1]
		debug := hasFlag(args, "--debug")
		confirm := hasFlag(args, "--confirm")
		os.Exit(cli.Run(filename, debug, confirm))

	case "check":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "speek check: missing filename")
			fmt.Fprintln(os.Stderr, "Usage: speek check <file.spk>")
			os.Exit(1)
		}
		os.Exit(cli.Check(args[1]))

	case "repl":
		cli.REPL()

	case "install-vscode", "install-vs-code", "vscode":
		force := hasFlag(args, "--force")
		os.Exit(cli.InstallVSCode(VSCodeExtFS, force))

	case "register-filetype", "register-file-type":
		os.Exit(cli.RegisterFileType(SpeekLogoICO))

	case "update", "upgrade":
		os.Exit(cli.Update(version))

	default:
		// allow running "speek file.spk" without the "run" keyword
		if strings.HasSuffix(cmd, ".spk") || strings.HasSuffix(cmd, ".speek") {
			debug := hasFlag(args, "--debug")
			confirm := hasFlag(args, "--confirm")
			os.Exit(cli.Run(args[0], debug, confirm))
		}
		fmt.Fprintf(os.Stderr, "speek: unknown command '%s'\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if strings.ToLower(a) == flag {
			return true
		}
	}
	return false
}

func printUsage() {
	fmt.Print(`Speek - Natural Language Programming Language

Usage:
  speek run <file.spk>           Run a Speek program
  speek run <file.spk> --debug   Show intent panel before running
  speek run <file.spk> --confirm Ask for confirmation before running
  speek check <file.spk>         Validate without running
  speek repl                    Interactive shell
  speek install-vscode          Install VS Code syntax highlighting
  speek install-vscode --force  Reinstall / overwrite existing extension
  speek update                  Update Speek to the latest version
  speek help                    Show all commands
  speek help math               Show math built-ins
  speek help strings            Show string built-ins
  speek help lists              Show list operations
  speek version                 Show version info

Examples:
  speek run hello.spk
  speek run fizzbuzz.spk --debug
  speek repl

Statements can be written across lines OR separated by commas and periods:
  create variable x, put 5 into x. show x.

`)
}
