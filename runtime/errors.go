package runtime

import (
	"fmt"
	"sort"
	"strings"
)

type RuntimeError struct {
	Line    int
	Raw     string
	Message string
	Hint    string
	Vars    map[string]Value
}

func (e *RuntimeError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Line %d: %s\n", e.Line, e.Message))
	if e.Raw != "" {
		sb.WriteString(fmt.Sprintf("  Code: %s\n", e.Raw))
	}
	if e.Hint != "" {
		sb.WriteString(fmt.Sprintf("  -> %s\n", e.Hint))
	}
	if len(e.Vars) > 0 {
		sb.WriteString("  Variables at this point:\n")
		keys := make([]string, 0, len(e.Vars))
		for k := range e.Vars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("    %s = %s\n", k, e.Vars[k].String()))
		}
	}
	return strings.TrimRight(sb.String(), "\n")
}

func Errorf(line int, raw, hint, format string, args ...interface{}) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf(format, args...),
		Hint:    hint,
	}
}

func ErrUndeclared(line int, raw, name string) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("'%s' hasn't been declared yet.", name),
		Hint:    fmt.Sprintf("Try: make a variable called %s", name),
	}
}

func ErrDivZero(line int, raw string, vars map[string]Value) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: "Can't divide by zero.",
		Hint:    "Make sure the denominator is not zero before dividing.",
		Vars:    vars,
	}
}

func ErrTypeMismatch(line int, raw, op string, a, b Value) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("Can't %s '%s' and '%s' - one is %s, the other is %s.", op, a.String(), b.String(), a.TypeName(), b.TypeName()),
		Hint:    "Convert with: 'x as number' or 'x as string'",
	}
}

func ErrOutOfRange(line int, raw string, idx, length int) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("Index %d is out of range. The list has %d items (indexes 0 to %d).", idx, length, length-1),
		Hint:    "Check that your index is within the list boundaries.",
	}
}

func ErrWrongArgCount(line int, raw, name string, want, got int) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("Function '%s' expects %d argument(s) but got %d.", name, want, got),
		Hint:    fmt.Sprintf("Call it like: call %s with arg1, arg2", name),
	}
}

func ErrNotAList(line int, raw, name string, got ValueType) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("'%s' is a %s, not a list.", name, got),
		Hint:    fmt.Sprintf("Declare a list first: make a list called %s", name),
	}
}

func ErrUnknownFn(line int, raw, name string) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("Function '%s' hasn't been defined.", name),
		Hint:    fmt.Sprintf("Define it first: define a function called %s", name),
	}
}

func ErrAlreadyDeclared(line int, raw, name string) *RuntimeError {
	return &RuntimeError{
		Line:    line,
		Raw:     raw,
		Message: fmt.Sprintf("Variable '%s' is already declared in this scope.", name),
		Hint:    fmt.Sprintf("To update its value, use: set %s to <new value>", name),
	}
}
