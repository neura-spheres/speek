package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ValueType string

const (
	TypeNumber ValueType = "number"
	TypeString ValueType = "string"
	TypeBool   ValueType = "bool"
	TypeList   ValueType = "list"
	TypeNull   ValueType = "null"
)

// Value is the one type used for everything at runtime: numbers, strings, bools, lists, null
type Value struct {
	Type   ValueType
	Number float64
	Str    string
	Bool   bool
	List   []Value
}

var Null = Value{Type: TypeNull}

func NumberVal(n float64) Value { return Value{Type: TypeNumber, Number: n} }
func StringVal(s string) Value  { return Value{Type: TypeString, Str: s} }
func BoolVal(b bool) Value      { return Value{Type: TypeBool, Bool: b} }
func ListVal(items []Value) Value { return Value{Type: TypeList, List: items} }

func (v Value) AsNumber() float64 {
	switch v.Type {
	case TypeNumber:
		return v.Number
	case TypeString:
		f, err := strconv.ParseFloat(strings.TrimSpace(v.Str), 64)
		if err == nil {
			return f
		}
		return 0
	case TypeBool:
		if v.Bool {
			return 1
		}
		return 0
	case TypeList:
		return float64(len(v.List))
	default:
		return 0
	}
}

func (v Value) AsString() string { return v.String() }
func (v Value) AsBool() bool     { return v.IsTruthy() }

func (v Value) IsTruthy() bool {
	switch v.Type {
	case TypeBool:
		return v.Bool
	case TypeNumber:
		return v.Number != 0 && !math.IsNaN(v.Number)
	case TypeString:
		return v.Str != "" && v.Str != "false" && v.Str != "0"
	case TypeList:
		return len(v.List) > 0
	case TypeNull:
		return false
	default:
		return false
	}
}

func (v Value) Equals(other Value) bool {
	if v.Type != other.Type {
		// numbers and strings can cross-compare ("5" == 5)
		if (v.Type == TypeNumber || v.Type == TypeString) &&
			(other.Type == TypeNumber || other.Type == TypeString) {
			return v.AsNumber() == other.AsNumber()
		}
		return false
	}
	switch v.Type {
	case TypeNumber:
		return v.Number == other.Number
	case TypeString:
		return v.Str == other.Str
	case TypeBool:
		return v.Bool == other.Bool
	case TypeNull:
		return true
	case TypeList:
		if len(v.List) != len(other.List) {
			return false
		}
		for i := range v.List {
			if !v.List[i].Equals(other.List[i]) {
				return false
			}
		}
		return true
	}
	return false
}

func (v Value) String() string {
	switch v.Type {
	case TypeNumber:
		if v.Number == math.Trunc(v.Number) && !math.IsInf(v.Number, 0) {
			return strconv.FormatInt(int64(v.Number), 10)
		}
		return strconv.FormatFloat(v.Number, 'f', -1, 64)
	case TypeString:
		return v.Str
	case TypeBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case TypeNull:
		return "null"
	case TypeList:
		parts := make([]string, len(v.List))
		for i, item := range v.List {
			parts[i] = fmt.Sprintf("%v", item.String())
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return ""
	}
}

func (v Value) TypeName() string {
	return string(v.Type)
}
