package bcn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// NormaCacheSuite validates the generic in-memory etag cache.
type NormaCacheSuite struct {
	suite.Suite
}

func TestNormaCacheSuite(t *testing.T) {
	suite.Run(t, new(NormaCacheSuite))
}

func (s *NormaCacheSuite) TestPutGetRoundTrip() {
	c := newEtagCache[NormaFull]()
	c.put("1@", etagEntry[NormaFull]{etag: "abc", value: NormaFull{Metadatos: Metadatos{TituloNorma: "X"}}})

	entry, ok := c.get("1@")
	s.Require().True(ok)
	s.Equal("abc", entry.etag)
	s.Equal("X", entry.value.Metadatos.TituloNorma)
}

func (s *NormaCacheSuite) TestGetMissing() {
	c := newEtagCache[NormaFull]()
	_, ok := c.get("42@")
	s.False(ok)
}

func (s *NormaCacheSuite) TestPutReplacesEntry() {
	c := newEtagCache[NormaFull]()
	c.put("1@", etagEntry[NormaFull]{etag: "old", value: NormaFull{Metadatos: Metadatos{TituloNorma: "old"}}})
	c.put("1@", etagEntry[NormaFull]{etag: "new", value: NormaFull{Metadatos: Metadatos{TituloNorma: "new"}}})

	entry, ok := c.get("1@")
	s.Require().True(ok)
	s.Equal("new", entry.etag)
	s.Equal("new", entry.value.Metadatos.TituloNorma)
}

func (s *NormaCacheSuite) TestCompositeKeysDoNotMix() {
	// The composite key (normID@versionDate) must keep versions apart.
	c := newEtagCache[NormaFull]()
	c.put("1@", etagEntry[NormaFull]{value: NormaFull{Metadatos: Metadatos{TituloNorma: "latest"}}})
	c.put("1@2010-01-01", etagEntry[NormaFull]{value: NormaFull{Metadatos: Metadatos{TituloNorma: "v2010"}}})

	latest, ok := c.get("1@")
	s.Require().True(ok)
	s.Equal("latest", latest.value.Metadatos.TituloNorma)
	v2010, ok := c.get("1@2010-01-01")
	s.Require().True(ok)
	s.Equal("v2010", v2010.value.Metadatos.TituloNorma)
}

func (s *NormaCacheSuite) TestCapEvictsArbitraryEntry() {
	c := newEtagCache[NormaFull]()
	for i := 0; i < cacheMax+10; i++ {
		c.put(fmt.Sprintf("%d", i), etagEntry[NormaFull]{etag: "e"})
	}
	// The cache never exceeds the cap: eviction happened on overflow.
	hits := 0
	for i := 0; i < cacheMax+10; i++ {
		if _, ok := c.get(fmt.Sprintf("%d", i)); ok {
			hits++
		}
	}
	s.LessOrEqual(hits, cacheMax, "cache exceeded the cap")
	s.Greater(hits, 0, "cache must retain entries after eviction")
}
