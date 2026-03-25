package runtime

import (
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BuiltinFn func(args []Value) (Value, error)
var builtinRegistry map[string]BuiltinFn

func init() {
	builtinRegistry = buildRegistry()
}

// CallBuiltin looks up and executes a built-in by its canonical name.
func CallBuiltin(name string, args []Value) (Value, error) {
	fn, ok := builtinRegistry[name]
	if !ok {
		return Null, fmt.Errorf("unknown built-in '%s'", name)
	}
	return fn(args)
}

// IsBuiltin returns true if the name resolves to a built-in function.
func IsBuiltin(name string) bool {
	_, ok := builtinRegistry[name]
	return ok
}

// AllBuiltinNames returns the list of all registered built-in names.
func AllBuiltinNames() []string {
	names := make([]string, 0, len(builtinRegistry))
	for k := range builtinRegistry {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// builtinAliases maps natural language phrases to canonical built-in names.
// The interpreter resolves calls through this map before hitting the registry.
var builtinAliases = map[string]string{
	// Math
	"abs":                     "abs",
	"absolute value":          "abs",
	"absolute value of":       "abs",
	"square root":             "sqrt",
	"square root of":          "sqrt",
	"sqrt":                    "sqrt",
	"cube root":               "cbrt",
	"cube root of":            "cbrt",
	"cbrt":                    "cbrt",
	"power":                   "pow",
	"power of":                "pow",
	"pow":                     "pow",
	"ceiling":                 "ceil",
	"ceiling of":              "ceil",
	"ceil":                    "ceil",
	"floor":                   "floor",
	"floor of":                "floor",
	"round":                   "round",
	"round to":                "round",
	"truncate":                "trunc",
	"truncate to":             "trunc",
	"trunc":                   "trunc",
	"log":                     "log",
	"log of":                  "log",
	"natural log":             "log",
	"natural log of":          "log",
	"log base 10":             "log10",
	"log base 10 of":          "log10",
	"log10":                   "log10",
	"log base 2":              "log2",
	"log base 2 of":           "log2",
	"log2":                    "log2",
	"sin":                     "sin",
	"sin of":                  "sin",
	"sine":                    "sin",
	"cos":                     "cos",
	"cos of":                  "cos",
	"cosine":                  "cos",
	"tan":                     "tan",
	"tan of":                  "tan",
	"tangent":                 "tan",
	"arcsin":                  "asin",
	"arcsin of":               "asin",
	"asin":                    "asin",
	"arccos":                  "acos",
	"arccos of":               "acos",
	"acos":                    "acos",
	"arctan":                  "atan",
	"arctan of":               "atan",
	"atan":                    "atan",
	"arctan2":                 "atan2",
	"atan2":                   "atan2",
	"exp":                     "exp",
	"exp of":                  "exp",
	"exp2":                    "exp2",
	"hypot":                   "hypot",
	"hypotenuse":              "hypot",
	"max":                     "max",
	"maximum":                 "max",
	"max of":                  "max",
	"min":                     "min",
	"minimum":                 "min",
	"min of":                  "min",
	"sign":                    "sign",
	"sign of":                 "sign",
	"signum":                  "sign",
	"pi":                      "pi",
	"e":                       "euler",
	"euler":                   "euler",
	"euler number":            "euler",
	"infinity":                "infinity",
	"inf":                     "infinity",
	"is nan":                  "isnan",
	"isnan":                   "isnan",
	"is infinite":             "isinf",
	"isinf":                   "isinf",
	"clamp":                   "clamp",
	"length":           "length",
	"length of":        "length",
	"len":              "length",
	"len of":           "length",
	"size":             "length",
	"size of":          "length",
	"uppercase":        "uppercase",
	"upper":            "uppercase",
	"to upper":         "uppercase",
	"to uppercase":     "uppercase",
	"lowercase":        "lowercase",
	"lower":            "lowercase",
	"to lower":         "lowercase",
	"to lowercase":     "lowercase",
	"reverse":          "reverse",
	"reverse of":       "reverse",
	"trim":             "trim",
	"strip":            "trim",
	"trim left":        "trimleft",
	"left trim":        "trimleft",
	"trim right":       "trimright",
	"right trim":       "trimright",
	"contains":         "contains",
	"starts with":      "startswith",
	"ends with":        "endswith",
	"replace":          "replace",
	"split":            "split",
	"split by":         "split",
	"join":             "join",
	"join with":        "join",
	"index of":         "indexof",
	"find":             "indexof",
	"find index":       "indexof",
	"substring":        "substring",
	"substr":           "substring",
	"slice":            "substring",
	"repeat":           "repeat",
	"repeat string":    "repeat",
	"count":            "count",
	"count of":         "count",
	"compare":          "compare",
	"compare strings":  "compare",
	"is empty":         "isempty",
	"empty":            "isempty",
	"pad left":         "padleft",
	"pad right":        "padright",
	"first char":       "firstchar",
	"first character":  "firstchar",
	"last char":        "lastchar",
	"last character":   "lastchar",
	"char at":          "charat",
	"character at":     "charat",
	"remove from":      "removefrom",
	"insert into":      "insertinto",

	// Type conversion
	"as number":  "tonumber",
	"to number":  "tonumber",
	"parse":      "tonumber",
	"as string":  "tostring",
	"to string":  "tostring",
	"as bool":    "tobool",
	"to bool":    "tobool",
	"as list":    "tolist",
	"to list":    "tolist",
	"type of":    "typeof",
	"type":       "typeof",
	"is number":  "isnumber",
	"is string":  "isstring",
	"is bool":    "isbool",
	"is list":    "islist",
	"is null":    "isnull",

	// List
	"length of list":    "listlength",
	"list length":       "listlength",
	"count items":       "listlength",
	"count items in":    "listlength",
	"first item":        "listfirst",
	"first item in":     "listfirst",
	"last item":         "listlast",
	"last item in":      "listlast",
	"item":              "listitem",
	"item at":           "listitem",
	"add item":          "listadd",
	"add to list":       "listadd",
	"remove last":       "listpop",
	"pop":               "listpop",
	"remove first":      "listshift",
	"shift":             "listshift",
	"insert at":         "listinsert",
	"remove at":         "listremoveat",
	"contains item":     "listcontains",
	"list contains":     "listcontains",
	"index of item":     "listindexof",
	"find in list":      "listindexof",
	"reverse list":      "listreverse",
	"sort list":         "listsort",
	"sort ascending":    "listsort",
	"sort descending":   "listsortdesc",
	"shuffle":           "listshuffle",
	"shuffle list":      "listshuffle",
	"slice list":        "listslice",
	"flatten":           "listflatten",
	"unique":            "listunique",
	"deduplicate":       "listunique",
	"sum":               "listsum",
	"sum of":            "listsum",
	"total":             "listsum",
	"average":           "listavg",
	"mean":              "listavg",
	"average of":        "listavg",

	// Random
	"random number":   "random",
	"random":          "random",
	"random between":  "randrange",
	"random item":     "randitem",
	"random item from": "randitem",

	// Time
	"current time":    "unixtime",
	"unix time":       "unixtime",
	"timestamp":       "unixtime",
	"current year":    "year",
	"this year":       "year",
	"current month":   "month",
	"this month":      "month",
	"current day":     "day",
	"today":           "day",
	"current hour":    "hour",
	"current minute":  "minute",
	"current second":  "second",
	"format time":     "formattime",
	"wait":            "sleep",
	"sleep":           "sleep",
	"pause":           "sleep",

	// System
	"exit":             "exit",
	"exit with":        "exit",
	"quit":             "exit",
	"env":              "getenv",
	"environment":      "getenv",
	"environment variable": "getenv",
	"run command":      "runcmd",
	"execute command":  "runcmd",
	"shell":            "runcmd",

	// File
	"read file":        "readfile",
	"write to file":    "writefile",
	"append to file":   "appendfile",
	"file exists":      "fileexists",
	"exists":           "fileexists",
	"delete file":      "deletefile",
	"list files":       "listfiles",
	"list files in":    "listfiles",
	"list directory":   "listfiles",

	// Utility
	"format":           "format",
	"sprintf":          "format",
	"hash":             "hash",
	"hash of":          "hash",
	"encode base64":    "base64enc",
	"base64 encode":    "base64enc",
	"encode as base64": "base64enc",
	"decode base64":    "base64dec",
	"base64 decode":    "base64dec",
	"decode from base64": "base64dec",
}

// ResolveBuiltinAlias maps a natural language phrase to a canonical name.
func ResolveBuiltinAlias(phrase string) (string, bool) {
	phrase = strings.ToLower(strings.TrimSpace(phrase))
	if canon, ok := builtinAliases[phrase]; ok {
		return canon, true
	}
	return "", false
}

func buildRegistry() map[string]BuiltinFn {
	return map[string]BuiltinFn{
		// MATH
		"abs": func(args []Value) (Value, error) {
			n := requireNum(args, 0)
			return NumberVal(math.Abs(n)), nil
		},
		"sqrt": func(args []Value) (Value, error) {
			return NumberVal(math.Sqrt(requireNum(args, 0))), nil
		},
		"cbrt": func(args []Value) (Value, error) {
			return NumberVal(math.Cbrt(requireNum(args, 0))), nil
		},
		"pow": func(args []Value) (Value, error) {
			return NumberVal(math.Pow(requireNum(args, 0), requireNum(args, 1))), nil
		},
		"ceil": func(args []Value) (Value, error) {
			return NumberVal(math.Ceil(requireNum(args, 0))), nil
		},
		"floor": func(args []Value) (Value, error) {
			return NumberVal(math.Floor(requireNum(args, 0))), nil
		},
		"round": func(args []Value) (Value, error) {
			return NumberVal(math.Round(requireNum(args, 0))), nil
		},
		"trunc": func(args []Value) (Value, error) {
			return NumberVal(math.Trunc(requireNum(args, 0))), nil
		},
		"log": func(args []Value) (Value, error) {
			return NumberVal(math.Log(requireNum(args, 0))), nil
		},
		"log10": func(args []Value) (Value, error) {
			return NumberVal(math.Log10(requireNum(args, 0))), nil
		},
		"log2": func(args []Value) (Value, error) {
			return NumberVal(math.Log2(requireNum(args, 0))), nil
		},
		"sin": func(args []Value) (Value, error) {
			return NumberVal(math.Sin(requireNum(args, 0))), nil
		},
		"cos": func(args []Value) (Value, error) {
			return NumberVal(math.Cos(requireNum(args, 0))), nil
		},
		"tan": func(args []Value) (Value, error) {
			return NumberVal(math.Tan(requireNum(args, 0))), nil
		},
		"asin": func(args []Value) (Value, error) {
			return NumberVal(math.Asin(requireNum(args, 0))), nil
		},
		"acos": func(args []Value) (Value, error) {
			return NumberVal(math.Acos(requireNum(args, 0))), nil
		},
		"atan": func(args []Value) (Value, error) {
			return NumberVal(math.Atan(requireNum(args, 0))), nil
		},
		"atan2": func(args []Value) (Value, error) {
			return NumberVal(math.Atan2(requireNum(args, 0), requireNum(args, 1))), nil
		},
		"exp": func(args []Value) (Value, error) {
			return NumberVal(math.Exp(requireNum(args, 0))), nil
		},
		"exp2": func(args []Value) (Value, error) {
			return NumberVal(math.Exp2(requireNum(args, 0))), nil
		},
		"hypot": func(args []Value) (Value, error) {
			return NumberVal(math.Hypot(requireNum(args, 0), requireNum(args, 1))), nil
		},
		"max": func(args []Value) (Value, error) {
			return NumberVal(math.Max(requireNum(args, 0), requireNum(args, 1))), nil
		},
		"min": func(args []Value) (Value, error) {
			return NumberVal(math.Min(requireNum(args, 0), requireNum(args, 1))), nil
		},
		"sign": func(args []Value) (Value, error) {
			n := requireNum(args, 0)
			if n > 0 {
				return NumberVal(1), nil
			} else if n < 0 {
				return NumberVal(-1), nil
			}
			return NumberVal(0), nil
		},
		"pi": func(args []Value) (Value, error) {
			return NumberVal(math.Pi), nil
		},
		"euler": func(args []Value) (Value, error) {
			return NumberVal(math.E), nil
		},
		"infinity": func(args []Value) (Value, error) {
			return NumberVal(math.Inf(1)), nil
		},
		"isnan": func(args []Value) (Value, error) {
			return BoolVal(math.IsNaN(requireNum(args, 0))), nil
		},
		"isinf": func(args []Value) (Value, error) {
			return BoolVal(math.IsInf(requireNum(args, 0), 0)), nil
		},
		"clamp": func(args []Value) (Value, error) {
			x := requireNum(args, 0)
			lo := requireNum(args, 1)
			hi := requireNum(args, 2)
			return NumberVal(math.Min(math.Max(x, lo), hi)), nil
		},

		// STRING
		"length": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return NumberVal(0), nil
			}
			if args[0].Type == TypeList {
				return NumberVal(float64(len(args[0].List))), nil
			}
			return NumberVal(float64(len([]rune(args[0].AsString())))), nil
		},
		"uppercase": func(args []Value) (Value, error) {
			return StringVal(strings.ToUpper(requireStr(args, 0))), nil
		},
		"lowercase": func(args []Value) (Value, error) {
			return StringVal(strings.ToLower(requireStr(args, 0))), nil
		},
		"reverse": func(args []Value) (Value, error) {
			if len(args) > 0 && args[0].Type == TypeList {
				lst := make([]Value, len(args[0].List))
				copy(lst, args[0].List)
				for i, j := 0, len(lst)-1; i < j; i, j = i+1, j-1 {
					lst[i], lst[j] = lst[j], lst[i]
				}
				return ListVal(lst), nil
			}
			r := []rune(requireStr(args, 0))
			for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
				r[i], r[j] = r[j], r[i]
			}
			return StringVal(string(r)), nil
		},
		"trim": func(args []Value) (Value, error) {
			return StringVal(strings.TrimSpace(requireStr(args, 0))), nil
		},
		"trimleft": func(args []Value) (Value, error) {
			return StringVal(strings.TrimLeft(requireStr(args, 0), " \t\n\r")), nil
		},
		"trimright": func(args []Value) (Value, error) {
			return StringVal(strings.TrimRight(requireStr(args, 0), " \t\n\r")), nil
		},
		"contains": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return BoolVal(false), nil
			}
			return BoolVal(strings.Contains(args[1].AsString(), args[0].AsString())), nil
		},
		"startswith": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return BoolVal(false), nil
			}
			return BoolVal(strings.HasPrefix(args[1].AsString(), args[0].AsString())), nil
		},
		"endswith": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return BoolVal(false), nil
			}
			return BoolVal(strings.HasSuffix(args[1].AsString(), args[0].AsString())), nil
		},
		"replace": func(args []Value) (Value, error) {
			if len(args) < 3 {
				return Null, fmt.Errorf("replace needs 3 args: old, new, source")
			}
			return StringVal(strings.ReplaceAll(args[2].AsString(), args[0].AsString(), args[1].AsString())), nil
		},
		"split": func(args []Value) (Value, error) {
			if len(args) < 2 {
				parts := strings.Fields(args[0].AsString())
				vals := make([]Value, len(parts))
				for i, p := range parts {
					vals[i] = StringVal(p)
				}
				return ListVal(vals), nil
			}
			parts := strings.Split(args[0].AsString(), args[1].AsString())
			vals := make([]Value, len(parts))
			for i, p := range parts {
				vals[i] = StringVal(p)
			}
			return ListVal(vals), nil
		},
		"join": func(args []Value) (Value, error) {
			if len(args) < 1 || args[0].Type != TypeList {
				return StringVal(""), nil
			}
			sep := ""
			if len(args) >= 2 {
				sep = args[1].AsString()
			}
			parts := make([]string, len(args[0].List))
			for i, v := range args[0].List {
				parts[i] = v.AsString()
			}
			return StringVal(strings.Join(parts, sep)), nil
		},
		"indexof": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return NumberVal(-1), nil
			}
			idx := strings.Index(args[1].AsString(), args[0].AsString())
			return NumberVal(float64(idx)), nil
		},
		"substring": func(args []Value) (Value, error) {
			if len(args) < 3 {
				return StringVal(requireStr(args, 0)), nil
			}
			s := []rune(args[0].AsString())
			a := int(args[1].AsNumber())
			b := int(args[2].AsNumber())
			if a < 0 {
				a = 0
			}
			if b > len(s) {
				b = len(s)
			}
			if a > b {
				a, b = b, a
			}
			return StringVal(string(s[a:b])), nil
		},
		"repeat": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return StringVal(requireStr(args, 0)), nil
			}
			return StringVal(strings.Repeat(args[0].AsString(), int(args[1].AsNumber()))), nil
		},
		"count": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return NumberVal(0), nil
			}
			return NumberVal(float64(strings.Count(args[1].AsString(), args[0].AsString()))), nil
		},
		"compare": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return NumberVal(0), nil
			}
			return NumberVal(float64(strings.Compare(args[0].AsString(), args[1].AsString()))), nil
		},
		"isempty": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return BoolVal(true), nil
			}
			return BoolVal(strings.TrimSpace(args[0].AsString()) == ""), nil
		},
		"padleft": func(args []Value) (Value, error) {
			if len(args) < 3 {
				return StringVal(requireStr(args, 0)), nil
			}
			s := args[0].AsString()
			n := int(args[1].AsNumber())
			pad := args[2].AsString()
			if pad == "" || n <= 0 {
				return StringVal(s), nil
			}
			for len([]rune(s)) < n {
				s = pad + s
			}
			if r := []rune(s); len(r) > n {
				s = string(r[len(r)-n:])
			}
			return StringVal(s), nil
		},
		"padright": func(args []Value) (Value, error) {
			if len(args) < 3 {
				return StringVal(requireStr(args, 0)), nil
			}
			s := args[0].AsString()
			n := int(args[1].AsNumber())
			pad := args[2].AsString()
			if pad == "" || n <= 0 {
				return StringVal(s), nil
			}
			for len([]rune(s)) < n {
				s = s + pad
			}
			if r := []rune(s); len(r) > n {
				s = string(r[:n])
			}
			return StringVal(s), nil
		},
		"firstchar": func(args []Value) (Value, error) {
			s := requireStr(args, 0)
			if s == "" {
				return StringVal(""), nil
			}
			return StringVal(string([]rune(s)[0])), nil
		},
		"lastchar": func(args []Value) (Value, error) {
			r := []rune(requireStr(args, 0))
			if len(r) == 0 {
				return StringVal(""), nil
			}
			return StringVal(string(r[len(r)-1])), nil
		},
		"charat": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return StringVal(""), nil
			}
			r := []rune(args[0].AsString())
			idx := int(args[1].AsNumber())
			if idx < 0 || idx >= len(r) {
				return StringVal(""), nil
			}
			return StringVal(string(r[idx])), nil
		},
		"removefrom": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return StringVal(requireStr(args, 0)), nil
			}
			return StringVal(strings.ReplaceAll(args[1].AsString(), args[0].AsString(), "")), nil
		},
		"insertinto": func(args []Value) (Value, error) {
			if len(args) < 3 {
				return StringVal(requireStr(args, 0)), nil
			}
			s := []rune(args[1].AsString())
			idx := int(args[2].AsNumber())
			ins := []rune(args[0].AsString())
			if idx < 0 {
				idx = 0
			}
			if idx > len(s) {
				idx = len(s)
			}
			out := make([]rune, 0, len(s)+len(ins))
			out = append(out, s[:idx]...)
			out = append(out, ins...)
			out = append(out, s[idx:]...)
			return StringVal(string(out)), nil
		},

		// TYPE CONVERSION
		"tonumber": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return NumberVal(0), nil
			}
			return NumberVal(args[0].AsNumber()), nil
		},
		"tostring": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal(""), nil
			}
			return StringVal(args[0].AsString()), nil
		},
		"tobool": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return BoolVal(false), nil
			}
			return BoolVal(args[0].IsTruthy()), nil
		},
		"tolist": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return ListVal([]Value{}), nil
			}
			if args[0].Type == TypeList {
				return args[0], nil
			}
			return ListVal([]Value{args[0]}), nil
		},
		"typeof": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal("null"), nil
			}
			return StringVal(args[0].TypeName()), nil
		},
		"isnumber": func(args []Value) (Value, error) {
			return BoolVal(len(args) > 0 && args[0].Type == TypeNumber), nil
		},
		"isstring": func(args []Value) (Value, error) {
			return BoolVal(len(args) > 0 && args[0].Type == TypeString), nil
		},
		"isbool": func(args []Value) (Value, error) {
			return BoolVal(len(args) > 0 && args[0].Type == TypeBool), nil
		},
		"islist": func(args []Value) (Value, error) {
			return BoolVal(len(args) > 0 && args[0].Type == TypeList), nil
		},
		"isnull": func(args []Value) (Value, error) {
			return BoolVal(len(args) == 0 || args[0].Type == TypeNull), nil
		},

		// LIST
		"listlength": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return NumberVal(0), nil
			}
			return NumberVal(float64(len(args[0].List))), nil
		},
		"listfirst": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return Null, nil
			}
			return args[0].List[0], nil
		},
		"listlast": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return Null, nil
			}
			lst := args[0].List
			return lst[len(lst)-1], nil
		},
		"listitem": func(args []Value) (Value, error) {
			if len(args) < 2 || args[0].Type != TypeList {
				return Null, nil
			}
			idx := int(args[1].AsNumber())
			lst := args[0].List
			if idx < 0 {
				idx = len(lst) + idx
			}
			if idx < 0 || idx >= len(lst) {
				return Null, nil
			}
			return lst[idx], nil
		},
		"listadd": func(args []Value) (Value, error) {
			if len(args) < 2 || args[0].Type != TypeList {
				return Null, nil
			}
			args[0].List = append(args[0].List, args[1])
			return args[0], nil
		},
		"listpop": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return Null, nil
			}
			lst := args[0].List
			last := lst[len(lst)-1]
			args[0].List = lst[:len(lst)-1]
			return last, nil
		},
		"listshift": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return Null, nil
			}
			first := args[0].List[0]
			args[0].List = args[0].List[1:]
			return first, nil
		},
		"listinsert": func(args []Value) (Value, error) {
			if len(args) < 3 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := args[0].List
			idx := int(args[2].AsNumber())
			item := args[1]
			if idx < 0 {
				idx = 0
			}
			if idx > len(lst) {
				idx = len(lst)
			}
			out := make([]Value, 0, len(lst)+1)
			out = append(out, lst[:idx]...)
			out = append(out, item)
			out = append(out, lst[idx:]...)
			return ListVal(out), nil
		},
		"listremoveat": func(args []Value) (Value, error) {
			if len(args) < 2 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := args[0].List
			idx := int(args[1].AsNumber())
			if idx < 0 || idx >= len(lst) {
				return Null, nil
			}
			out := make([]Value, 0, len(lst)-1)
			out = append(out, lst[:idx]...)
			out = append(out, lst[idx+1:]...)
			return ListVal(out), nil
		},
		"listcontains": func(args []Value) (Value, error) {
			if len(args) < 2 || args[0].Type != TypeList {
				return BoolVal(false), nil
			}
			for _, v := range args[0].List {
				if v.Equals(args[1]) {
					return BoolVal(true), nil
				}
			}
			return BoolVal(false), nil
		},
		"listindexof": func(args []Value) (Value, error) {
			if len(args) < 2 || args[0].Type != TypeList {
				return NumberVal(-1), nil
			}
			for i, v := range args[0].List {
				if v.Equals(args[1]) {
					return NumberVal(float64(i)), nil
				}
			}
			return NumberVal(-1), nil
		},
		"listreverse": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := make([]Value, len(args[0].List))
			copy(lst, args[0].List)
			for i, j := 0, len(lst)-1; i < j; i, j = i+1, j-1 {
				lst[i], lst[j] = lst[j], lst[i]
			}
			return ListVal(lst), nil
		},
		"listsort": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := make([]Value, len(args[0].List))
			copy(lst, args[0].List)
			sort.Slice(lst, func(i, j int) bool {
				if lst[i].Type == TypeNumber && lst[j].Type == TypeNumber {
					return lst[i].Number < lst[j].Number
				}
				return lst[i].AsString() < lst[j].AsString()
			})
			return ListVal(lst), nil
		},
		"listsortdesc": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := make([]Value, len(args[0].List))
			copy(lst, args[0].List)
			sort.Slice(lst, func(i, j int) bool {
				if lst[i].Type == TypeNumber && lst[j].Type == TypeNumber {
					return lst[i].Number > lst[j].Number
				}
				return lst[i].AsString() > lst[j].AsString()
			})
			return ListVal(lst), nil
		},
		"listshuffle": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := make([]Value, len(args[0].List))
			copy(lst, args[0].List)
			rand.Shuffle(len(lst), func(i, j int) { lst[i], lst[j] = lst[j], lst[i] })
			return ListVal(lst), nil
		},
		"listslice": func(args []Value) (Value, error) {
			if len(args) < 3 || args[0].Type != TypeList {
				return Null, nil
			}
			lst := args[0].List
			a := int(args[1].AsNumber())
			b := int(args[2].AsNumber())
			if a < 0 {
				a = 0
			}
			if b > len(lst) {
				b = len(lst)
			}
			if a > b {
				a, b = b, a
			}
			out := make([]Value, b-a)
			copy(out, lst[a:b])
			return ListVal(out), nil
		},
		"listflatten": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			var flat []Value
			for _, v := range args[0].List {
				if v.Type == TypeList {
					flat = append(flat, v.List...)
				} else {
					flat = append(flat, v)
				}
			}
			return ListVal(flat), nil
		},
		"listunique": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return Null, nil
			}
			seen := []Value{}
			for _, v := range args[0].List {
				found := false
				for _, s := range seen {
					if s.Equals(v) {
						found = true
						break
					}
				}
				if !found {
					seen = append(seen, v)
				}
			}
			return ListVal(seen), nil
		},
		"listsum": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList {
				return NumberVal(0), nil
			}
			var sum float64
			for _, v := range args[0].List {
				sum += v.AsNumber()
			}
			return NumberVal(sum), nil
		},
		"listavg": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return NumberVal(0), nil
			}
			var sum float64
			for _, v := range args[0].List {
				sum += v.AsNumber()
			}
			return NumberVal(sum / float64(len(args[0].List))), nil
		},

		// RANDOM
		"random": func(args []Value) (Value, error) {
			return NumberVal(rand.Float64()), nil
		},
		"randrange": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return NumberVal(0), nil
			}
			lo := int(args[0].AsNumber())
			hi := int(args[1].AsNumber())
			if hi <= lo {
				return NumberVal(float64(lo)), nil
			}
			return NumberVal(float64(lo + rand.Intn(hi-lo+1))), nil
		},
		"randitem": func(args []Value) (Value, error) {
			if len(args) == 0 || args[0].Type != TypeList || len(args[0].List) == 0 {
				return Null, nil
			}
			return args[0].List[rand.Intn(len(args[0].List))], nil
		},

		// TIME
		"unixtime": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Unix())), nil
		},
		"year": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Year())), nil
		},
		"month": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Month())), nil
		},
		"day": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Day())), nil
		},
		"hour": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Hour())), nil
		},
		"minute": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Minute())), nil
		},
		"second": func(args []Value) (Value, error) {
			return NumberVal(float64(time.Now().Second())), nil
		},
		"formattime": func(args []Value) (Value, error) {
			t := time.Now()
			if len(args) > 0 {
				ts := int64(args[0].AsNumber())
				t = time.Unix(ts, 0)
			}
			return StringVal(t.Format("2006-01-02 15:04:05")), nil
		},
		"sleep": func(args []Value) (Value, error) {
			secs := 1.0
			if len(args) > 0 {
				secs = args[0].AsNumber()
			}
			time.Sleep(time.Duration(secs * float64(time.Second)))
			return Null, nil
		},

		// SYSTEM
		"exit": func(args []Value) (Value, error) {
			code := 0
			if len(args) > 0 {
				code = int(args[0].AsNumber())
			}
			os.Exit(code)
			return Null, nil
		},
		"getenv": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal(""), nil
			}
			return StringVal(os.Getenv(args[0].AsString())), nil
		},
		"runcmd": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal(""), nil
			}
			parts := strings.Fields(args[0].AsString())
			if len(parts) == 0 {
				return StringVal(""), nil
			}
			out, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
			if err != nil {
				return StringVal(string(out)), fmt.Errorf("command failed: %w", err)
			}
			return StringVal(strings.TrimRight(string(out), "\n")), nil
		},

		// FILE
		"readfile": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal(""), nil
			}
			data, err := os.ReadFile(args[0].AsString())
			if err != nil {
				return StringVal(""), err
			}
			return StringVal(string(data)), nil
		},
		"writefile": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return Null, fmt.Errorf("writefile needs content and filename")
			}
			err := os.WriteFile(args[1].AsString(), []byte(args[0].AsString()), 0644)
			return Null, err
		},
		"appendfile": func(args []Value) (Value, error) {
			if len(args) < 2 {
				return Null, fmt.Errorf("appendfile needs content and filename")
			}
			f, err := os.OpenFile(args[1].AsString(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return Null, err
			}
			defer f.Close()
			_, err = f.WriteString(args[0].AsString())
			return Null, err
		},
		"fileexists": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return BoolVal(false), nil
			}
			_, err := os.Stat(args[0].AsString())
			return BoolVal(!os.IsNotExist(err)), nil
		},
		"deletefile": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return Null, nil
			}
			return Null, os.Remove(args[0].AsString())
		},
		"listfiles": func(args []Value) (Value, error) {
			dir := "."
			if len(args) > 0 {
				dir = args[0].AsString()
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				return ListVal([]Value{}), err
			}
			vals := make([]Value, 0, len(entries))
			for _, e := range entries {
				vals = append(vals, StringVal(filepath.Join(dir, e.Name())))
			}
			return ListVal(vals), nil
		},

		// UTILITY
		"format": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return StringVal(""), nil
			}
			fmtStr := args[0].AsString()
			iargs := make([]interface{}, len(args)-1)
			for i, a := range args[1:] {
				iargs[i] = a.AsString()
			}
			return StringVal(fmt.Sprintf(fmtStr, iargs...)), nil
		},
		"hash": func(args []Value) (Value, error) {
			h := fnv.New64a()
			h.Write([]byte(requireStr(args, 0)))
			return NumberVal(float64(h.Sum64())), nil
		},
		"base64enc": func(args []Value) (Value, error) {
			return StringVal(base64.StdEncoding.EncodeToString([]byte(requireStr(args, 0)))), nil
		},
		"base64dec": func(args []Value) (Value, error) {
			data, err := base64.StdEncoding.DecodeString(requireStr(args, 0))
			if err != nil {
				return StringVal(""), err
			}
			return StringVal(string(data)), nil
		},
		"tonum": func(args []Value) (Value, error) {
			if len(args) == 0 {
				return NumberVal(0), nil
			}
			f, err := strconv.ParseFloat(strings.TrimSpace(args[0].AsString()), 64)
			if err != nil {
				return NumberVal(0), nil
			}
			return NumberVal(f), nil
		},
	}
}

// helpers

func requireNum(args []Value, idx int) float64 {
	if idx < len(args) {
		return args[idx].AsNumber()
	}
	return 0
}

func requireStr(args []Value, idx int) string {
	if idx < len(args) {
		return args[idx].AsString()
	}
	return ""
}
