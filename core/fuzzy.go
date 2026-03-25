package core

import (
	"fmt"
	"strings"
)

// FuzzySuggest returns a human-friendly error message when no pattern matches.
// It uses Levenshtein distance on the first word to guess what the user meant.
func FuzzySuggest(lineNum int, raw string, dict Dictionary) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Sprintf("Line %d: empty line.", lineNum)
	}

	first := firstWord(trimmed)
	allVerbs := AllVerbs(dict)

	best := ""
	bestDist := 999
	for _, verb := range allVerbs {
		d := levenshtein(strings.ToLower(first), strings.ToLower(verb))
		if d < bestDist {
			bestDist = d
			best = verb
		}
	}

	if bestDist <= 2 && best != "" {
		corrected := best + trimmed[len(first):]
		return fmt.Sprintf(
			"Line %d: I don't recognize: %q\n"+
				"  -> Did you mean: %q?\n"+
				"  -> Type 'speek help' to see all commands.",
			lineNum, trimmed, corrected,
		)
	}

	return fmt.Sprintf(
		"Line %d: I don't understand: %q\n"+
			"  -> This might be a typo or unsupported command.\n"+
			"  -> Type 'speek help' to see what I can do.",
		lineNum, trimmed,
	)
}

// firstWord returns the first space-delimited word of a string.
func firstWord(s string) string {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return s
	}
	return parts[0]
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)
	la, lb := len(ra), len(rb)

	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	// Two-row DP
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min3(del, ins, sub)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
