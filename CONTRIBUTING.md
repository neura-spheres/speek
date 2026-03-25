# Contributing to Speek

Thanks for wanting to contribute. This document explains how the codebase is structured and how to add new things to the language.

---

## Getting started

Clone the repo and build it.

```bash
git clone https://github.com/neura-spheres/speek
cd speek
go build -o speek.exe .
```

Run the tests to make sure everything works before making changes.

```bash
go test ./...
```

---

## Project structure

```
speek/
  core/
    dictionary.go     synonym word lists for every keyword slot
    patterns.go       all statement patterns (most specific first)
    compiler.go       compiles patterns into regex at startup
    tokenizer.go      splits source lines into individual statements
    matcher.go        matches a line against compiled patterns
    parser.go         nests matched nodes into a block tree
    ast.go            node type definitions
    fuzzy.go          suggests corrections for unrecognized lines
  runtime/
    interpreter.go    executes the AST node by node
    builtins.go       all built-in functions (sqrt, length, etc.)
    expr.go           expression evaluator for inline math (a + b * c)
    scope.go          variable and function scoping
    types.go          the Value type and helpers
  cli/
    run.go            speek run command
    check.go          speek check command
    repl.go           interactive shell
    help.go           speek help command
    vscode.go         speek install-vscode command
    update.go         speek update command
  vscode-ext/         VS Code extension source (embedded into binary)
  examples/           example .spk programs
  main.go             CLI entry point
  embed.go            embeds vscode-ext into the binary
```

---

## Adding new synonym words

Every keyword in Speek has a slot in `core/dictionary.go`. A slot is just a list of words that all mean the same thing.

For example, the print slot looks like this:

```go
"VERB_PRINT": {"show", "print", "display", "output", "log", "say", "tell", "echo"},
```

To add a new word, just append it to the list.

```go
"VERB_PRINT": {"show", "print", "display", "output", "log", "say", "tell", "echo", "write"},
```

Then rebuild and test.

```bash
go build -o speek.exe .
go test ./...
```

---

## Adding a new statement type

Adding a new kind of statement requires changes in four files. Here is the full process using a simple example: adding a `wait X seconds` statement.

### Step 1 - Add a node type in `core/ast.go`

```go
NodeWait NodeType = "wait"
```

### Step 2 - Add a pattern in `core/patterns.go`

Patterns are matched top to bottom, so put more specific patterns before more general ones.

```go
{
    Op:       "wait",
    Template: `wait\s+{VALUE}\s+seconds?`,
    Captures: []string{"value"},
},
```

### Step 3 - Build the node in `core/parser.go`

Inside the `buildNode` function, add a case for your new op.

```go
case "wait":
    return Node{Type: NodeWait, Value: caps["value"]}, nil
```

### Step 4 - Execute it in `runtime/interpreter.go`

Inside the `execNode` function, add a case for your new node type.

```go
case core.NodeWait:
    val, err := interp.resolveValue(node, scope)
    if err != nil {
        return err
    }
    seconds := val.AsNumber()
    time.Sleep(time.Duration(seconds * float64(time.Second)))
```

---

## Adding a new built-in function

Built-in functions live in `runtime/builtins.go`. There are two things to add.

First, add an alias if the function has a natural language name.

```go
"square root of": "sqrt",
"length of":      "len",
"my new function": "mynewfn",
```

Second, add the actual implementation inside `CallBuiltin`.

```go
case "mynewfn":
    if len(args) < 1 {
        return Null, fmt.Errorf("mynewfn requires 1 argument")
    }
    // do something with args[0]
    return NumberVal(result), nil
```

---

## Updating the VS Code extension

The extension source is in `vscode-ext/`. After editing any file there, rebuild the binary and reinstall.

```bash
go build -o speek.exe .
./speek.exe install-vscode --force
```

---

## Submitting a pull request

1. Fork the repo on GitHub
2. Create a branch for your change (`git checkout -b add-my-feature`)
3. Make your changes and run `go test ./...`
4. Commit and push to your fork
5. Open a pull request against `main`

Please keep pull requests focused on one thing. If you want to add a new feature and also fix a bug, open two separate PRs. It makes reviewing much easier.

---

## Reporting bugs

Open an issue on GitHub and include the following.

- The `.spk` file or the lines that caused the problem
- The output you got
- The output you expected
- Your OS and Speek version (`speek version`)
