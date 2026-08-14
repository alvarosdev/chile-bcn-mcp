// This file centralizes every piece of "garbage" that LeyChile responses
// carry, so the sanitizer never deals with magic strings or raw rune
// literals. Each entry documents where it comes from in the BCN API.
//
// NOTE: every rune here is written as an escape — literal control/invisible
// characters (e.g. a BOM) are illegal in Go source files.
package bcn

import "regexp"

const (
	// nbspRune is the non-breaking space (&nbsp;, U+00A0). BCN uses it as
	// visual indentation at the start of every paragraph in get_norma_json.
	nbspRune = '\u00a0'
	// enspRune is the en-space (&ensp;, U+2002) variant.
	enspRune = '\u2002'
	// emspRune is the em-space (&emsp;, U+2003) variant.
	emspRune = '\u2003'
	// zeroWidthSpaceRune is U+200B, occasionally embedded in BCN text.
	zeroWidthSpaceRune = '\u200b'
	// bomRune is the byte-order mark (U+FEFF) sometimes embedded in responses.
	bomRune = '\ufeff'
	// controlMin and controlMax delimit the C0 control block. The sanitizer
	// removes it, keeping only newline, tab and carriage return.
	controlMin = '\u0000'
	controlMax = '\u001f'
)

// spacesToNormalize lists every whitespace rune that must become a plain
// space before the text reaches the LLM.
var spacesToNormalize = []rune{nbspRune, enspRune, emspRune}

// zeroWidthChars lists invisible runes that are dropped entirely.
var zeroWidthChars = []rune{zeroWidthSpaceRune, bomRune}

// resumenTagRe matches the XML wrapper that surrounds the RESUMEN field of
// search results: <RESUMENES>...<RESUMEN idioma="Español">...</RESUMEN>...</RESUMENES>.
var resumenTagRe = regexp.MustCompile(`(?i)</?resumen[^>]*>`)
