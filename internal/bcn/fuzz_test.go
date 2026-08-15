package bcn

import (
	"slices"
	"testing"
)

// Fuzz targets for the hostile-input surface of the client: the sanitizer
// and the HTML→Markdown conversion feed on external BCN markup. Run with
// `go test ./internal/bcn -fuzz=FuzzSanitizeMarkdown -fuzztime=30s`
// (same for FuzzConvertContent). No network involved.

// FuzzSanitizeMarkdown ensures the single-pass sanitizer never panics and
// produces no control characters, whatever the input.
func FuzzSanitizeMarkdown(f *testing.F) {
	f.Add(`<div><p>&nbsp; &nbsp; Art&iacute;culo 1&deg;.- texto</p><br></div>`)
	f.Add(" ​\x00texto\r\n\t con   espacios")
	f.Fuzz(func(t *testing.T, input string) {
		out := SanitizeMarkdown(input)
		// The contract: no zero-width or C0 control runes in the output.
		for _, r := range out {
			if slices.Contains(zeroWidthChars, r) || (r >= controlMin && r <= controlMax && r != '\n' && r != '\t') {
				t.Fatalf("unsanitized rune %q in output", r)
			}
		}
	})
}

// FuzzConvertContent ensures the HTML→Markdown conversion of a norm block
// tree never panics, whatever markup BCN-shaped input arrives.
func FuzzConvertContent(f *testing.F) {
	f.Add(`<div><div class="p">T&iacute;tulo I</div><br></div>`)
	f.Add(`<div class="p">&nbsp; &nbsp; <a href="https://www.bcn.cl/leychile/navegar?idNorma=1222799">ley N° 21.809</a></div>`)
	f.Add(`<RESUMENES><![CDATA[resumen]]></RESUMENES>`)
	f.Fuzz(func(t *testing.T, input string) {
		n := NormaFull{Html: []HtmlBlock{{T: input, I: 1, H: []HtmlBlock{{T: input, I: 2}}}}}
		conv := newConverter()
		n.ConvertContent(conv)
	})
}
