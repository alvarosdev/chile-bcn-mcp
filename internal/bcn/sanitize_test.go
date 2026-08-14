package bcn

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// SanitizeSuite validates the garbage-cleanup pipeline with one case per
// garbage type found in real BCN API responses.
type SanitizeSuite struct {
	suite.Suite
}

func TestSanitizeSuite(t *testing.T) {
	suite.Run(t, new(SanitizeSuite))
}

func (s *SanitizeSuite) TestSummaryStripsXMLWrapperAndEntities() {
	in := "<RESUMENES>\n  <RESUMEN idioma=\"Espa\u00f1ol\">    La ley N&#176; 2 crea el servicio. </RESUMEN>\n</RESUMENES>"
	out := SanitizeSummary(in)
	s.Equal("La ley N\u00b0 2 crea el servicio.", out)
}

func (s *SanitizeSuite) TestSummaryHandlesEmpty() {
	s.Equal("", SanitizeSummary(""))
}

func (s *SanitizeSuite) TestNbspVariantsBecomePlainSpace() {
	in := "\u00a0\u00a0  Art\u00edculo 1.- El \u00a0servicio\u2002y\u2003la ley."
	out := SanitizeMarkdown(in)
	s.Equal("Art\u00edculo 1.- El servicio y la ley.", out)
}

func (s *SanitizeSuite) TestZeroWidthAndBOMRemoved() {
	in := "ley\u200b21.600\ufeff y otras"
	out := SanitizeMarkdown(in)
	s.Equal("ley21.600 y otras", out)
}

func (s *SanitizeSuite) TestControlCharsRemoved() {
	in := "ley\x0121.600\x00 y otras"
	out := SanitizeMarkdown(in)
	s.Equal("ley21.600 y otras", out)
}

func (s *SanitizeSuite) TestQuotesPreserved() {
	// Entity decoding (&quot; \u2192 ") happens inside the markdown converter;
	// the sanitizer's job is to not touch literal quotes. The full pipeline
	// (decode + normalize) is verified in law_client_test.go with the real
	// API fixture.
	in := "\"Art\u00edculo 1.-\""
	out := SanitizeMarkdown(in)
	s.Equal("\"Art\u00edculo 1.-\"", out)
}

func (s *SanitizeSuite) TestTrailingWhitespacePerLineTrimmed() {
	in := "hola \n  mundo \n\t"
	out := SanitizeMarkdown(in)
	s.Equal("hola\nmundo", out)
}

func (s *SanitizeSuite) TestLinksAreContentAndSurvive() {
	in := "ver [ley N\u00b0 21.809](https://www.bcn.cl/leychile/navegar?idNorma=1222799) hoy"
	out := SanitizeMarkdown(in)
	s.Equal("ver [ley N\u00b0 21.809](https://www.bcn.cl/leychile/navegar?idNorma=1222799) hoy", out)
}

func (s *SanitizeSuite) TestConsecutiveSpacesCollapse() {
	in := "la   ley  21.600"
	out := SanitizeMarkdown(in)
	s.Equal("la ley 21.600", out)
}

func (s *SanitizeSuite) TestNewlinesPreserved() {
	in := "Art\u00edculo 1.\n\nInciso primero."
	out := SanitizeMarkdown(in)
	s.Equal("Art\u00edculo 1.\n\nInciso primero.", out)
}

func (s *SanitizeSuite) TestNewlineRunsCollapsed() {
	// BCN empty <div class="p"> paragraphs produce 3+ newlines; they
	// collapse to at most 2 (paragraph break).
	in := "Art\u00edculo 1.\n\n\n\n\nInciso primero."
	out := SanitizeMarkdown(in)
	s.Equal("Art\u00edculo 1.\n\nInciso primero.", out)
}
