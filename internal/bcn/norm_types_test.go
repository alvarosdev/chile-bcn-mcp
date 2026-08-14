package bcn

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// NormTypesSuite validates the norm-type catalog lookups.
type NormTypesSuite struct {
	suite.Suite
}

func TestNormTypesSuite(t *testing.T) {
	suite.Run(t, new(NormTypesSuite))
}

func (s *NormTypesSuite) TestKnownCodes() {
	cases := []struct {
		cod   int
		valor string
		abbr  string
	}{
		{1, "Ley", "LEY"},
		{2, "Decreto", "DTO"},
		{3, "Resolución", "RES"},
		{13, "Decreto con Fuerza de Ley", "DFL"},
		{15, "Decreto Ley", "DL"},
		{39, "Constitución Política de la República", "CPR"},
	}
	for _, c := range cases {
		valor, abbr, ok := canonicalNormType(c.cod)
		s.Require().True(ok, "cod %d should be in the catalog", c.cod)
		s.Equal(c.valor, valor, "cod %d valor", c.cod)
		s.Equal(c.abbr, abbr, "cod %d abbr", c.cod)
	}
}

func (s *NormTypesSuite) TestUnknownCode() {
	valor, abbr, ok := canonicalNormType(999)
	s.False(ok)
	s.Empty(valor)
	s.Empty(abbr)
}
