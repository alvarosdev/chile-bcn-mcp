package bcn

import (
	"html"
	"slices"
	"strings"
)

// SanitizeSummary cleans the RESUMEN field of search results: decodes HTML
// entities and strips the XML wrapper. Returns LLM-readable plain text.
func SanitizeSummary(raw string) string {
	s := html.UnescapeString(raw)
	s = resumenTagRe.ReplaceAllString(s, "")
	return normalize(s)
}

// SanitizeMarkdown cleans converted Markdown content: decodes any remaining
// entities and normalizes whitespace.
func SanitizeMarkdown(md string) string {
	return normalize(md)
}

// normalize is the single-pass state machine that strips the garbage defined
// in garbage.go:
//
//   - spacesToNormalize runes become a plain space
//   - zeroWidthChars and C0 control runes are dropped
//   - consecutive spaces collapse (pending-space technique)
//   - trailing whitespace per line is dropped (spaces before \n)
//   - leading whitespace at line start is dropped
//
// Quotes and links are preserved — they are content, not garbage.
//
// Validated by benchmark: ~0.6ms per 100KB, 8 allocs (see sanitize_bench_test.go).
func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	pendingSpace := false
	atLineStart := true
	newlineRun := 0
	for _, r := range s {
		switch {
		case isSpaceToNormalize(r):
			r = ' '
		case isZeroWidth(r):
			continue
		case isControl(r):
			continue
		}
		if r == ' ' {
			pendingSpace = true
			continue
		}
		if r == '\n' {
			pendingSpace = false // trim trailing whitespace per line
			atLineStart = true
			// Collapse runs of 3+ newlines (BCN empty <div class="p">
			// paragraphs) to at most 2.
			if newlineRun < 2 {
				b.WriteRune(r)
			}
			newlineRun++
			continue
		}
		if pendingSpace && !atLineStart {
			b.WriteByte(' ')
		}
		pendingSpace = false
		atLineStart = false
		newlineRun = 0
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func isSpaceToNormalize(r rune) bool {
	return slices.Contains(spacesToNormalize, r)
}

func isZeroWidth(r rune) bool {
	return slices.Contains(zeroWidthChars, r)
}

func isControl(r rune) bool {
	// \r is a control char too: the spec contract eliminates every C0
	// control rune outside \n and \t (caught by FuzzSanitizeMarkdown —
	// a literal \r in the middle of LLM-bound markdown is pure garbage).
	return r >= controlMin && r <= controlMax && r != '\n' && r != '\t'
}
