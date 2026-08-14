package bcn

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// NormaCacheSuite validates the in-memory norm cache.
type NormaCacheSuite struct {
	suite.Suite
}

func TestNormaCacheSuite(t *testing.T) {
	suite.Run(t, new(NormaCacheSuite))
}

func (s *NormaCacheSuite) TestPutGetRoundTrip() {
	c := NewNormaCache()
	c.put(1195666, cacheEntry{etag: "abc", norma: NormaFull{Metadatos: Metadatos{TituloNorma: "X"}}})

	entry, ok := c.get(1195666)
	s.Require().True(ok)
	s.Equal("abc", entry.etag)
	s.Equal("X", entry.norma.Metadatos.TituloNorma)
}

func (s *NormaCacheSuite) TestGetMissing() {
	c := NewNormaCache()
	_, ok := c.get(42)
	s.False(ok)
}

func (s *NormaCacheSuite) TestPutReplacesEntry() {
	c := NewNormaCache()
	c.put(1, cacheEntry{etag: "old", norma: NormaFull{Metadatos: Metadatos{TituloNorma: "old"}}})
	c.put(1, cacheEntry{etag: "new", norma: NormaFull{Metadatos: Metadatos{TituloNorma: "new"}}})

	entry, ok := c.get(1)
	s.Require().True(ok)
	s.Equal("new", entry.etag)
	s.Equal("new", entry.norma.Metadatos.TituloNorma)
}

func (s *NormaCacheSuite) TestCapEvictsArbitraryEntry() {
	c := NewNormaCache()
	for i := 0; i < normaCacheMax+10; i++ {
		c.put(int64(i), cacheEntry{etag: "e"})
	}
	// The cache never exceeds the cap: eviction happened on overflow.
	hits := 0
	for i := 0; i < normaCacheMax+10; i++ {
		if _, ok := c.get(int64(i)); ok {
			hits++
		}
	}
	s.LessOrEqual(hits, normaCacheMax, "cache exceeded the cap")
	s.Greater(hits, 0, "cache must retain entries after eviction")
}
