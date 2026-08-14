package bcn

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"resty.dev/v3"

	"dev.alvaros.chile-bcn-mcp/internal/config"
)

// LawClientSuite exercises the real resty client against a local
// httptest.Server — no external network calls, ever.
type LawClientSuite struct {
	suite.Suite
}

func TestLawClientSuite(t *testing.T) {
	suite.Run(t, new(LawClientSuite))
}

func (s *LawClientSuite) logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testResources builds the contract pointing at the local test server.
// Defaults: 1 retry, tiny backoff, high breaker threshold (no interference).
func (s *LawClientSuite) testResources(serverURL string) *config.Resources {
	return &config.Resources{
		Version: 1,
		Resources: map[string]config.Resource{
			resourceSearchLaws: {
				URL:     serverURL,
				Path:    "/buscarjson",
				Method:  "GET",
				Timeout: config.Duration(2 * time.Second),
				Retry: config.Retry{
					Attempts:   1,
					Backoff:    config.Duration(time.Millisecond),
					MaxBackoff: config.Duration(2 * time.Millisecond),
				},
				CircuitBreaker: config.CircuitBreaker{
					FailureThreshold: 100,
					SuccessThreshold: 1,
					ResetTimeout:     config.Duration(time.Second),
				},
			},
			resourceGetLaw: {
				URL:     serverURL,
				Path:    "/get_norma_json",
				Method:  "GET",
				Timeout: config.Duration(2 * time.Second),
				Retry: config.Retry{
					Attempts:   1,
					Backoff:    config.Duration(time.Millisecond),
					MaxBackoff: config.Duration(2 * time.Millisecond),
				},
				CircuitBreaker: config.CircuitBreaker{
					FailureThreshold: 100,
					SuccessThreshold: 1,
					ResetTimeout:     config.Duration(time.Second),
				},
			},
		},
	}
}

func (s *LawClientSuite) fixture(name string) []byte {
	data, err := os.ReadFile(filepath.Join("testdata", name))
	s.Require().NoError(err)
	return data
}

func (s *LawClientSuite) TestSearchParsesRealResponse() {
	searchJSON := s.fixture("search_response.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/buscarjson", r.URL.Path)
		s.Equal("Ley 21.600", r.URL.Query().Get("cadena"))
		s.Equal("1", r.URL.Query().Get("npagina"))
		s.Equal("3", r.URL.Query().Get("itemsporpagina"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(searchJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	result, err := client.Search(context.Background(), SearchParams{
		Query: "Ley 21.600", Page: 1, PageSize: 3,
	})
	s.Require().NoError(err)
	s.Require().Len(result.Results, 3)
	s.Equal(140, result.Pagination.TotalItems)

	first := result.Results[0]
	s.Equal(int64(1195666), first.IDNorma)
	s.Equal("Ley", first.Tipo)
	s.Contains(first.TituloNorma, "BIODIVERSIDAD")
	// The XML wrapper and its indentation are gone; entities are decoded.
	s.NotContains(first.Resumen, "<RESUMENES>")
	s.Contains(first.Resumen, "Servicio de Biodiversidad")
	s.NotContains(first.Resumen, " ")
}

func (s *LawClientSuite) TestSearchRetriesTransient5xx() {
	searchJSON := s.fixture("search_response.json")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(searchJSON)
	}))
	defer server.Close()

	res := s.testResources(server.URL)
	searchRes := res.Resources[resourceSearchLaws]
	searchRes.Retry.Attempts = 3 // 2 failures + 1 success
	res.Resources[resourceSearchLaws] = searchRes
	client := NewClient(res, s.logger())

	result, err := client.Search(context.Background(), SearchParams{
		Query: "Ley 21.600", Page: 1, PageSize: 3,
	})
	s.Require().NoError(err)
	s.Len(result.Results, 3)
	s.Equal(int32(3), calls.Load())
}

func (s *LawClientSuite) TestCircuitBreakerOpensAfterFailures() {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	res := s.testResources(server.URL)
	searchRes := res.Resources[resourceSearchLaws]
	searchRes.Retry.Attempts = 0 // no retries
	searchRes.CircuitBreaker.FailureThreshold = 2
	res.Resources[resourceSearchLaws] = searchRes
	client := NewClient(res, s.logger())

	params := SearchParams{Query: "x", Page: 1, PageSize: 1}
	_, err1 := client.Search(context.Background(), params)
	s.Require().Error(err1)
	_, err2 := client.Search(context.Background(), params)
	s.Require().Error(err2)
	s.Equal(int32(2), calls.Load(), "two requests reached the server")

	// The breaker is open: the third call fails without touching the server.
	_, err3 := client.Search(context.Background(), params)
	s.Require().Error(err3)
	s.True(errors.Is(err3, resty.ErrCircuitBreakerOpen), "expected breaker-open error, got: %v", err3)
	s.Equal(int32(2), calls.Load(), "no third request reached the server")
}

func (s *LawClientSuite) TestGetNormaParsesRealResponse() {
	normaJSON := s.fixture("norma_full.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("/get_norma_json", r.URL.Path)
		s.Equal("1195666", r.URL.Query().Get("idNorma"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(normaJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	norma, err := client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)

	// Metadata.
	s.Equal("CREA EL SERVICIO DE BIODIVERSIDAD Y ÁREAS PROTEGIDAS Y EL SISTEMA NACIONAL DE ÁREAS PROTEGIDAS", norma.Metadatos.TituloNorma)
	s.Equal("Ley", norma.Metadatos.TiposNumeros[0].Descripcion)
	s.Equal("21600", norma.Metadatos.TiposNumeros[0].Numero)
	s.Contains(norma.Metadatos.Organismos, "MINISTERIO DEL MEDIO AMBIENTE")
	s.False(norma.Metadatos.Derogado)
	s.Equal("2023-09-06", norma.Metadatos.FechaPublicacion)
	s.Equal("2025-09-29", norma.Metadatos.Vigencia.InicioVigencia)
	s.NotEmpty(norma.Metadatos.Materias)

	// Structure and content: 10 top-level blocks, all converted to markdown.
	s.Require().Len(norma.Html, 10)
	s.Require().Len(norma.Estructura, 10)

	header := norma.Html[0]
	s.Equal("Encabezado", header.SectionName)
	// Entities decoded, nbsp indentation gone, no raw HTML left.
	s.Contains(header.Markdown, "LEY NÚM. 21.600")
	s.NotContains(header.Markdown, "&#xDA;")
	s.NotContains(header.Markdown, " ")
	s.NotContains(header.Markdown, "<div")

	// A titled section pairs structure[i] ↔ html[i].
	idx := slices.IndexFunc(norma.Html, func(block HtmlBlock) bool {
		return strings.Contains(block.SectionName, "TÍTULO I")
	})
	s.Require().NotEqual(-1, idx, "TÍTULO I section with content expected")
	s.NotEmpty(norma.Html[idx].Markdown)
}

func (s *LawClientSuite) TestGetNormaParsesNestedArticles() {
	normaJSON := s.fixture("norma_full.json")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(normaJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	norma, err := client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)

	// The API nests articles under titles (field "h"): TÍTULO I has 3
	// article children, and the article text must be converted.
	titulo := norma.Html[1]
	s.Require().Len(titulo.H, 3, "TÍTULO I must carry its nested articles")
	s.Equal("Artículo 1°", titulo.H[0].SectionName)
	s.Contains(titulo.H[0].Markdown, "Artículo 1°.- Objeto")
	s.NotContains(titulo.H[0].Markdown, "<div", "nested article must be converted to markdown")

	// The structure tree mirrors the nesting.
	s.Equal("Artículo 1°", norma.Estructura[1].H[0].N)

	// Deeper nesting: TÍTULO II → Párrafo 1° → Artículo 4°.
	tituloII := norma.Html[2]
	s.Require().NotEmpty(tituloII.H, "TÍTULO II must have paragraphs")
	parrafo := tituloII.H[0]
	s.Require().NotEmpty(parrafo.H, "Párrafo must have articles")
	s.Equal("Artículo 4°", parrafo.H[0].SectionName)
}

func (s *LawClientSuite) TestGetNormaServes304FromCache() {
	normaJSON := s.fixture("norma_full.json")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", `W/"abc123"`)
			_, _ = w.Write(normaJSON)
			return
		}
		s.Equal(`W/"abc123"`, r.Header.Get("If-None-Match"))
		w.WriteHeader(http.StatusNotModified)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	first, err := client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)
	second, err := client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)

	// The second call was served from cache: only ONE 200 download happened
	// and the content is identical (already converted, no re-fetch).
	s.Equal(int32(2), calls.Load(), "one download + one revalidation")
	s.Equal(first.Metadatos.TituloNorma, second.Metadatos.TituloNorma)
	s.Equal(first.Html[0].Markdown, second.Html[0].Markdown)
}

func (s *LawClientSuite) TestGetNormaReplacesCacheOnNewETag() {
	normaJSON := s.fixture("norma_full.json")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			w.Header().Set("ETag", `W/"old"`)
		} else {
			w.Header().Set("ETag", `W/"new"`)
		}
		_, _ = w.Write(normaJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	_, err := client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)
	// Server always answers 200: the cache entry must be replaced, not served.
	_, err = client.GetNorma(context.Background(), 1195666)
	s.Require().NoError(err)
	s.Equal(int32(2), calls.Load())
	entry, ok := client.cache.get(1195666)
	s.Require().True(ok)
	s.Equal(`W/"new"`, entry.etag)
}

func (s *LawClientSuite) TestGetNormaSummaryParsesRealResponse() {
	normaJSON := s.fixture("norma_full.json")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(normaJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	summary, err := client.GetNormaSummary(context.Background(), 1195666)
	s.Require().NoError(err)
	s.Equal("CREA EL SERVICIO DE BIODIVERSIDAD Y ÁREAS PROTEGIDAS Y EL SISTEMA NACIONAL DE ÁREAS PROTEGIDAS", summary.TituloNorma)
	s.Equal("Diario Oficial", summary.Fuente)
	s.NotEmpty(summary.Materias)
	s.NotEmpty(summary.Resumenes)
	// Sanitized: no leading indentation from the API.
	s.False(strings.HasPrefix(summary.Resumenes[0], " "), "resumen must not start with spaces")
}

func (s *LawClientSuite) TestGetNormaSummaryServesFromCacheWithoutHTTP() {
	normaJSON := s.fixture("norma_full.json")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(normaJSON)
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	first, err := client.GetNormaSummary(context.Background(), 1195666)
	s.Require().NoError(err)
	second, err := client.GetNormaSummary(context.Background(), 1195666)
	s.Require().NoError(err)

	// The second summary was derived from the cache: ONE request total.
	s.Equal(int32(1), calls.Load(), "second summary must not touch the network")
	s.Equal(first.TituloNorma, second.TituloNorma)
}

func (s *LawClientSuite) TestGetNormaSummaryNotFound() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("El id de la Norma proporcionado (999999999) no se encuentra en nuestra Base de Datos."))
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	_, err := client.GetNormaSummary(context.Background(), 999999999)
	s.Require().Error(err)
	s.True(errors.Is(err, ErrNormaNotFound), "expected ErrNormaNotFound, got: %v", err)
}

func (s *LawClientSuite) TestGetNormaNotFound() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("El id de la Norma proporcionado (999999999) no se encuentra en nuestra Base de Datos."))
	}))
	defer server.Close()

	client := NewClient(s.testResources(server.URL), s.logger())
	_, err := client.GetNorma(context.Background(), 999999999)
	s.Require().Error(err)
	s.True(errors.Is(err, ErrNormaNotFound), "expected ErrNormaNotFound, got: %v", err)
}
