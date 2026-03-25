package runtime

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// EvalExpr tries to evaluate expr as arithmetic (+ - * / % ** and parens).
// Returns (Null, false) if it doesn't look like an expression.
func EvalExpr(expr string, scope *Scope) (Value, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return Null, false
	}

	if !looksLikeExpr(expr) {
		return Null, false
	}

	tokens, err := tokenizeExpr(expr)
	if err != nil || len(tokens) == 0 {
		return Null, false
	}

	p := &exprParser{tokens: tokens, scope: scope}
	val, err := p.parseExpr()
	if err != nil || p.pos != len(p.tokens) {
		return Null, false
	}
	return val, true
}

func looksLikeExpr(s string) bool {
	depth := 0
	inStr := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			inStr = !inStr
			continue
		}
		if inStr {
			continue
		}
		if c == '(' {
			depth++
		} else if c == ')' {
			depth--
		} else if depth == 0 && (c == '+' || c == '-' || c == '*' || c == '/' || c == '%') {
			// skip leading sign -- it's not a binary operator
			if (c == '-' || c == '+') && i == 0 {
				continue
			}
			return true
		}
	}
	return false
}

const (
	tokNum  = "NUM"
	tokIdent = "IDENT"
	tokOp   = "OP"
	tokLPar = "LPAR"
	tokRPar = "RPAR"
)

type token struct {
	kind  string
	value string
}

func tokenizeExpr(s string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(s) {
		c := rune(s[i])

		if unicode.IsSpace(c) {
			i++
			continue
		}

		if unicode.IsDigit(c) || (c == '.' && i+1 < len(s) && unicode.IsDigit(rune(s[i+1]))) {
			j := i
			if j < len(s) && s[j] == '-' {
				j++
			}
			for j < len(s) && (unicode.IsDigit(rune(s[j])) || s[j] == '.') {
				j++
			}
			tokens = append(tokens, token{tokNum, s[i:j]})
			i = j
			continue
		}

		if unicode.IsLetter(c) || c == '_' {
			j := i
			for j < len(s) && (unicode.IsLetter(rune(s[j])) || unicode.IsDigit(rune(s[j])) || s[j] == '_') {
				j++
			}
			tokens = append(tokens, token{tokIdent, s[i:j]})
			i = j
			continue
		}

		// Double-star exponent
		if c == '*' && i+1 < len(s) && s[i+1] == '*' {
			tokens = append(tokens, token{tokOp, "**"})
			i += 2
			continue
		}

		if c == '+' || c == '-' || c == '*' || c == '/' || c == '%' {
			tokens = append(tokens, token{tokOp, string(c)})
			i++
			continue
		}

		if c == '(' {
			tokens = append(tokens, token{tokLPar, "("})
			i++
			continue
		}
		if c == ')' {
			tokens = append(tokens, token{tokRPar, ")"})
			i++
			continue
		}

		return nil, fmt.Errorf("unexpected character: %c", c)
	}
	return tokens, nil
}

type exprParser struct {
	tokens []token
	pos    int
	scope  *Scope
}

func (p *exprParser) peek() (token, bool) {
	if p.pos >= len(p.tokens) {
		return token{}, false
	}
	return p.tokens[p.pos], true
}

func (p *exprParser) consume() token {
	t := p.tokens[p.pos]
	p.pos++
	return t
}

// parseExpr, parseTerm, parsePower, parseUnary, parsePrimary form a standard recursive descent parser.
// precedence (low to high): +/- -> */% -> ** -> unary -> primary
func (p *exprParser) parseExpr() (Value, error) {
	left, err := p.parseTerm()
	if err != nil {
		return Null, err
	}

	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.value != "+" && t.value != "-") {
			break
		}
		p.consume()
		right, err := p.parseTerm()
		if err != nil {
			return Null, err
		}
		if t.value == "+" {
			if left.Type == TypeString || right.Type == TypeString {
				left = StringVal(left.AsString() + right.AsString())
			} else {
				left = NumberVal(left.AsNumber() + right.AsNumber())
			}
		} else {
			left = NumberVal(left.AsNumber() - right.AsNumber())
		}
	}
	return left, nil
}

func (p *exprParser) parseTerm() (Value, error) {
	left, err := p.parsePower()
	if err != nil {
		return Null, err
	}

	for {
		t, ok := p.peek()
		if !ok || t.kind != tokOp || (t.value != "*" && t.value != "/" && t.value != "%") {
			break
		}
		p.consume()
		right, err := p.parsePower()
		if err != nil {
			return Null, err
		}
		switch t.value {
		case "*":
			left = NumberVal(left.AsNumber() * right.AsNumber())
		case "/":
			if right.AsNumber() == 0 {
				return Null, fmt.Errorf("division by zero")
			}
			left = NumberVal(left.AsNumber() / right.AsNumber())
		case "%":
			if right.AsNumber() == 0 {
				return Null, fmt.Errorf("modulo by zero")
			}
			left = NumberVal(math.Mod(left.AsNumber(), right.AsNumber()))
		}
	}
	return left, nil
}

func (p *exprParser) parsePower() (Value, error) {
	base, err := p.parseUnary()
	if err != nil {
		return Null, err
	}

	t, ok := p.peek()
	if ok && t.kind == tokOp && t.value == "**" {
		p.consume()
		exp, err := p.parsePower() // right-associative: 2**3**2 = 2**(3**2)
		if err != nil {
			return Null, err
		}
		return NumberVal(math.Pow(base.AsNumber(), exp.AsNumber())), nil
	}
	return base, nil
}

func (p *exprParser) parseUnary() (Value, error) {
	t, ok := p.peek()
	if ok && t.kind == tokOp && t.value == "-" {
		p.consume()
		val, err := p.parsePrimary()
		if err != nil {
			return Null, err
		}
		return NumberVal(-val.AsNumber()), nil
	}
	if ok && t.kind == tokOp && t.value == "+" {
		p.consume()
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (Value, error) {
	t, ok := p.peek()
	if !ok {
		return Null, fmt.Errorf("unexpected end of expression")
	}

	switch t.kind {
	case tokNum:
		p.consume()
		f, err := strconv.ParseFloat(t.value, 64)
		if err != nil {
			return Null, err
		}
		return NumberVal(f), nil

	case tokIdent:
		p.consume()
		if val, ok := p.scope.Get(t.value); ok {
			return val, nil
		}
		// also check builtin constants like pi, e, infinity
		if canon, ok := ResolveBuiltinAlias(t.value); ok {
			result, err := CallBuiltin(canon, nil)
			if err == nil {
				return result, nil
			}
		}
		return Null, fmt.Errorf("unknown variable '%s'", t.value)

	case tokLPar:
		p.consume()
		val, err := p.parseExpr()
		if err != nil {
			return Null, err
		}
		t2, ok := p.peek()
		if !ok || t2.kind != tokRPar {
			return Null, fmt.Errorf("missing closing parenthesis")
		}
		p.consume()
		return val, nil
	}

	return Null, fmt.Errorf("unexpected token: %s", t.value)
}
