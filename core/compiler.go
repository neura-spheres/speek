package core

import (
	"fmt"
	"regexp"
	"strings"
)

type CompiledPattern struct {
	Op       string
	Re       *regexp.Regexp
	Captures []string
	Raw      string // original template string, useful for debugging
}

type Compiler struct {
	Patterns []CompiledPattern
	Dict     Dictionary
}

// NewCompiler compiles all patterns once at startup so we don't pay regex compilation cost at runtime
func NewCompiler(dict Dictionary, templates []PatternTemplate) (*Compiler, error) {
	c := &Compiler{Dict: dict}
	for _, tmpl := range templates {
		cp, err := compileTemplate(dict, tmpl)
		if err != nil {
			return nil, fmt.Errorf("compiling pattern %q: %w", tmpl.Template, err)
		}
		c.Patterns = append(c.Patterns, cp)
	}
	return c, nil
}

func compileTemplate(dict Dictionary, tmpl PatternTemplate) (CompiledPattern, error) {
	src := tmpl.Template

	if src == `#.*` || src == `//.*` || src == `--.*` || src == `note:.*` {
		re, err := regexp.Compile(`(?i)^` + src + `$`)
		if err != nil {
			return CompiledPattern{}, err
		}
		return CompiledPattern{Op: tmpl.Op, Re: re, Captures: tmpl.Captures, Raw: tmpl.Template}, nil
	}

	src = replaceCMP(src, dict)

	// expand {VERB_*} and {PREP_*} slots
	for slot, words := range dict {
		if !strings.HasPrefix(slot, "VERB_") {
			continue
		}
		ph := `{` + slot + `}`
		if !strings.Contains(src, ph) {
			continue
		}
		alt := wordAlternation(words)
		src = strings.ReplaceAll(src, ph, alt)
	}

	for slot, words := range dict {
		if !strings.HasPrefix(slot, "PREP_") {
			continue
		}
		ph := `{` + slot + `}`
		if !strings.Contains(src, ph) {
			continue
		}
		alt := wordAlternation(words)
		src = strings.ReplaceAll(src, ph, alt)
	}

	if strings.Contains(src, `{LOOP_UNIT}`) {
		alt := wordAlternation(dict["LOOP_UNIT"])
		src = strings.ReplaceAll(src, `{LOOP_UNIT}`, alt)
	}

	if strings.Contains(src, `{FILLER}`) {
		fillerAlt := wordAlternation(dict["FILLER"])
		fillerGroup := `(?:` + fillerAlt + `\s+)*`
		src = strings.ReplaceAll(src, `{FILLER}`, fillerGroup)
	}

	src = strings.ReplaceAll(src, `{NAME}`, `(?P<name>[a-zA-Z_][a-zA-Z0-9_]*)`)
	src = strings.ReplaceAll(src, `{LISTNAME}`, `(?P<listname>[a-zA-Z_][a-zA-Z0-9_]*)`)
	src = strings.ReplaceAll(src, `{PARAMS}`, `(?P<params>.+)`)
	src = strings.ReplaceAll(src, `{ARGS}`, `(?P<args>.+)`)
	src = strings.ReplaceAll(src, `{NUMBER}`, `(?P<number>\d+(?:\.\d+)?)`)
	src = strings.ReplaceAll(src, `{INDEX}`, `(?P<index>\d+)`)
	src = strings.ReplaceAll(src, `{TOVALUE}`, `(?P<tovalue>[^\s].*)`)

	// {VALUE} must come last -- it's greedy and would eat the other named groups
	src = strings.ReplaceAll(src, `{VALUE}`, `(?P<value>.+?)`)

	src = regexp.MustCompile(`\s+`).ReplaceAllString(src, `\s+`)

	fullPat := `(?i)^\s*` + src + `\s*$`

	re, err := regexp.Compile(fullPat)
	if err != nil {
		return CompiledPattern{}, fmt.Errorf("invalid regex %q: %w", fullPat, err)
	}

	return CompiledPattern{Op: tmpl.Op, Re: re, Captures: tmpl.Captures, Raw: tmpl.Template}, nil
}

func replaceCMP(src string, dict Dictionary) string {
	if !strings.Contains(src, `{CMP}`) {
		return src
	}

	// longer phrases listed first so they match before shorter prefixes
	type cmpEntry struct {
		name  string
		words []string
	}
	entries := []cmpEntry{
		{"divisible", dict["CMP_DIV"]},
		{"gte", dict["CMP_GTE"]},
		{"lte", dict["CMP_LTE"]},
		{"neq", dict["CMP_NEQ"]},
		{"gt", dict["CMP_GT"]},
		{"lt", dict["CMP_LT"]},
		{"eq", dict["CMP_EQ"]},
	}

	var parts []string
	for _, e := range entries {
		if len(e.words) == 0 {
			continue
		}
		alt := wordAlternation(e.words)
		parts = append(parts, `(?P<cmp_`+e.name+`>`+alt+`)`)
	}

	group := `(?P<cmp>` + strings.Join(parts, `|`) + `)`
	return strings.ReplaceAll(src, `{CMP}`, group)
}

// wordAlternation builds a regex alternation, sorting longer phrases first to avoid prefix conflicts
func wordAlternation(words []string) string {
	sorted := make([]string, len(words))
	copy(sorted, words)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if len(sorted[j]) > len(sorted[i]) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	escaped := make([]string, len(sorted))
	for i, w := range sorted {
		escaped[i] = regexp.QuoteMeta(w)
	}
	return `(?:` + strings.Join(escaped, `|`) + `)`
}
