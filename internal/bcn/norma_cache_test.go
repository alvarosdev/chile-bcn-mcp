package bcn

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// NormaCacheSuite validates the generic in-memory etag cache with LRU
// eviction.
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

func (s *NormaCacheSuite) TestCapEvictsLeastRecentlyUsed() {
	c := newEtagCache[NormaFull]()
	for i := range cacheMax + 10 {
		c.put(fmt.Sprintf("%d", i), etagEntry[NormaFull]{etag: "e"})
	}
	// LRU: the 10 oldest entries (never touched) were evicted on overflow
	// and the most recent one survives.
	for i := range 10 {
		_, ok := c.get(fmt.Sprintf("%d", i))
		s.False(ok, "oldest untouched entry %d must be evicted", i)
	}
	_, ok := c.get(fmt.Sprintf("%d", cacheMax+9))
	s.True(ok, "most recent entry must survive")
}

func (s *NormaCacheSuite) TestTouchMovesEntryToFront() {
	// A get refreshes recency: the touched entry survives the next overflow
	// while its untouched neighbor is evicted.
	c := newEtagCache[string]()
	for i := range cacheMax {
		c.put(fmt.Sprintf("k%d", i), etagEntry[string]{etag: fmt.Sprintf("e%d", i), value: "v"})
	}

	_, ok := c.get("k0")
	s.Require().True(ok)

	c.put("overflow", etagEntry[string]{etag: "e", value: "v"})
	_, ok = c.get("k1")
	s.False(ok, "least recently used entry must be evicted")
	_, ok = c.get("k0")
	s.True(ok, "recently used entry must survive overflow")
}
