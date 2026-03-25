package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type MatchResult struct {
	Op       string
	Captures map[string]string
	Line     int
	Raw      string
}

type Matcher struct {
	compiler *Compiler
}

func NewMatcher(c *Compiler) *Matcher {
	return &Matcher{compiler: c}
}

func (m *Matcher) Match(lineNum int, text string) (MatchResult, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return MatchResult{}, fmt.Errorf("empty line")
	}

	for _, cp := range m.compiler.Patterns {
		if !cp.Re.MatchString(trimmed) {
			continue
		}
		caps := extractCaptures(cp.Re, trimmed)
		// Merge CMP sub-groups into single "cmp" key
		caps = normalizeCmp(caps)
		return MatchResult{
			Op:       cp.Op,
			Captures: caps,
			Line:     lineNum,
			Raw:      text,
		}, nil
	}
	return MatchResult{}, fmt.Errorf("no pattern matched")
}

func extractCaptures(re *regexp.Regexp, s string) map[string]string {
	caps := make(map[string]string)
	match := re.FindStringSubmatch(s)
	if match == nil {
		return caps
	}
	for i, name := range re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		if i < len(match) && match[i] != "" {
			caps[name] = strings.TrimSpace(match[i])
		}
	}
	return caps
}

func normalizeCmp(caps map[string]string) map[string]string {
	cmpMap := map[string]string{
		"cmp_gt":       "gt",
		"cmp_lt":       "lt",
		"cmp_gte":      "gte",
		"cmp_lte":      "lte",
		"cmp_eq":       "eq",
		"cmp_neq":      "neq",
		"cmp_divisible": "divisible",
	}
	for k, v := range cmpMap {
		if val, ok := caps[k]; ok && val != "" {
			caps["cmp"] = v
			delete(caps, k)
		}
	}
	if raw, ok := caps["cmp"]; ok && caps["cmp"] == raw {
		_ = raw // already normalized above
	}
	return caps
}

// ParseValue figures out what kind of thing a captured string is.
// checks in order: quoted string, number, bool keyword, identifier (var ref), expression, raw string
func ParseValue(raw string, dict Dictionary) (interface{}, string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, "", fmt.Errorf("empty value")
	}

	if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
		(strings.HasPrefix(s, `'`) && strings.HasSuffix(s, `'`)) {
		return s[1 : len(s)-1], "", nil
	}

	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, "", nil
	}

	sLow := strings.ToLower(s)
	for _, w := range dict["BOOL_TRUE"] {
		if sLow == strings.ToLower(w) {
			return true, "", nil
		}
	}
	for _, w := range dict["BOOL_FALSE"] {
		if sLow == strings.ToLower(w) {
			return false, "", nil
		}
	}

	identRe := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if identRe.MatchString(s) {
		return nil, s, nil
	}

	// multi-word things like "square root of x" get passed to the interpreter as-is
	if strings.Contains(s, " ") {
		return nil, s, nil
	}

	return s, "", nil
}

func ParseArgs(raw string) []string {
	re := regexp.MustCompile(`,\s*|\s+and\s+`)
	parts := re.Split(strings.TrimSpace(raw), -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func ParseParams(raw string) []string {
	return ParseArgs(raw)
}
