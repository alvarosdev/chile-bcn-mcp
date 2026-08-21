package cgr

import (
	"html"
	"slices"
	"strings"
)

// SanitizeDocumento cleans the documento_completo field of dictamen
// responses: decodes HTML entities and normalizes whitespace. Returns
// LLM-readable plain text. Duplicated normalize logic from bcn to keep
// cgr decoupled.
func SanitizeDocumento(raw string) string {
	s := html.UnescapeString(raw)
	s = resumenTagRe.ReplaceAllString(s, "")
	return normalize(s)
}

// SanitizeMateria cleans the materia field of search results.
func SanitizeMateria(raw string) string {
	s := html.UnescapeString(raw)
	return normalize(s)
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
			// Collapse runs of 3+ newlines to at most 2.
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
	return r >= controlMin && r <= controlMax && r != '\n' && r != '\t'
}
