package core

import (
	"bufio"
	"regexp"
	"strings"
)

// Line holds a single source line with its original number.
type Line struct {
	Number int
	Text   string
}

// Tokenize splits source text into a slice of Line structs.
//
// Comma (,) and period (.) act as statement terminators/separators,
// so "create variable x, put 4 into x." becomes two separate statements.
// Leading filler words like "and" are stripped from the start of each segment.
// A trailing period is always optional.
//
// Blank lines and blank segments are filtered out automatically.
func Tokenize(source string) []Line {
	var lines []Line
	scanner := bufio.NewScanner(strings.NewReader(source))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := strings.TrimRight(scanner.Text(), " \t\r")
		if strings.TrimSpace(raw) == "" {
			continue
		}

		// Split line on commas and periods that act as separators.
		segments := splitStatements(raw)

		for _, seg := range segments {
			seg = strings.TrimSpace(seg)
			if seg == "" {
				continue
			}
			lines = append(lines, Line{Number: lineNum, Text: seg})
		}
	}
	return lines
}

// leadingConnectors strips "and", "then", "also" from the front of a segment
// so "create x, and put 4 into x" works just as well as "create x, put 4 into x".
var leadingConnectors = regexp.MustCompile(`(?i)^(?:and|then|also|but|so|next|now|after that)\s+`)

// splitStatements splits a raw source line on commas and sentence-ending periods,
// treating quoted string contents as opaque.
func splitStatements(line string) []string {
	var segments []string
	var current strings.Builder
	runes := []rune(line)
	inStr := false

	for i := 0; i < len(runes); i++ {
		ch := runes[i]

		if ch == '"' {
			inStr = !inStr
			current.WriteRune(ch)
			continue
		}

		if inStr {
			current.WriteRune(ch)
			continue
		}

		if ch == ',' || ch == '.' {
			seg := strings.TrimSpace(current.String())
			if seg != "" {
				seg = leadingConnectors.ReplaceAllString(seg, "")
				seg = strings.TrimSpace(seg)
				if seg != "" {
					segments = append(segments, seg)
				}
			}
			current.Reset()
			continue
		}

		current.WriteRune(ch)
	}

	// Trailing segment (no trailing punctuation required)
	seg := strings.TrimSpace(current.String())
	if seg != "" {
		seg = leadingConnectors.ReplaceAllString(seg, "")
		seg = strings.TrimSpace(seg)
		if seg != "" {
			segments = append(segments, seg)
		}
	}

	if len(segments) == 0 {
		return nil
	}
	return segments
}

// IndentLevel returns the number of leading spaces/tabs.
func IndentLevel(text string) int {
	count := 0
	for _, ch := range text {
		if ch == ' ' || ch == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}
