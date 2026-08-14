// Package bcn provides the LeyChile (Biblioteca del Congreso Nacional)
// client: search and retrieval of Chilean legal norms. All HTTP behavior
// (timeout, retry, circuit breaker) is configured per resource from the
// api.resources.yaml contract — no hardcoded URLs or policies in code.
package bcn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"resty.dev/v3"

	"dev.alvaros.chile-bcn-mcp/internal/config"
)

// Resource ids in api.resources.yaml.
const (
	resourceSearchLaws = "search_laws"
	resourceGetLaw     = "get_law"
)

// ErrNormaNotFound is returned when the requested norm id does not exist
// in LeyChile (the API answers HTTP 500 with a "not found" body).
var ErrNormaNotFound = errors.New("norma not found")

// LawClient defines the operations the MCP tools need from the BCN API.
// Implemented by *Client; mocked by MockLawClient (mockery).
type LawClient interface {
	// Search returns paginated search results for a query string.
	Search(ctx context.Context, params SearchParams) (SearchResponse, error)
	// GetNorma returns the full structured content of one norm.
	GetNorma(ctx context.Context, normID int64) (NormaFull, error)
	// GetNormaSummary returns the lightweight metadata view of one norm
	// (no content). Shares the ETag cache with GetNorma.
	GetNormaSummary(ctx context.Context, normID int64) (NormaSummary, error)
}

// Client implements LawClient over resty, with one *resty.Client per
// resource — resty's circuit breaker is client-level. Clients are built
// once in NewClient and reused for every request (no per-call setup).
// The embedded converter is thread-safe (mutex inside the library).
type Client struct {
	logger    *slog.Logger
	resources *config.Resources
	clients   map[string]*resty.Client
	converter *converter.Converter
	cache     *NormaCache
}

// NewClient builds a Client from the resources contract.
func NewClient(resources *config.Resources, logger *slog.Logger) *Client {
	c := &Client{
		logger:    logger,
		resources: resources,
		clients:   make(map[string]*resty.Client, len(resources.Resources)),
		converter: newConverter(),
		cache:     NewNormaCache(),
	}
	for name, res := range resources.Resources {
		c.clients[name] = c.newRestyClient(name, res)
	}
	return c
}

// newRestyClient configures one resty client per resource: base URL,
// timeout, count-based circuit breaker, and error logging per attempt.
func (c *Client) newRestyClient(name string, res config.Resource) *resty.Client {
	return resty.New().
		SetBaseURL(res.URL).
		SetTimeout(time.Duration(res.Timeout)).
		SetCircuitBreaker(resty.NewCircuitBreakerCount(
			uint64(res.CircuitBreaker.FailureThreshold),
			uint64(res.CircuitBreaker.SuccessThreshold),
			time.Duration(res.CircuitBreaker.ResetTimeout),
		)).
		OnError(func(req *resty.Request, err error) {
			c.logger.Debug("bcn request attempt failed",
				"resource", name,
				"error", err,
			)
		})
}

// SearchParams carries the arguments for a LeyChile search.
type SearchParams struct {
	Query    string
	Page     int
	PageSize int
}

// SearchResponse is the paginated result of a LeyChile search.
//
// The wire format is a 3-element array [results, pagination, facets];
// UnmarshalJSON splits it into the typed fields. Facets (element 2) are
// intentionally ignored for now.
type SearchResponse struct {
	Results    []Norma
	Pagination Pagination
}

// UnmarshalJSON decodes the heterogeneous 3-element array of buscarjson.
func (r *SearchResponse) UnmarshalJSON(data []byte) error {
	var parts [3]json.RawMessage
	if err := json.Unmarshal(data, &parts); err != nil {
		return err
	}
	if err := json.Unmarshal(parts[0], &r.Results); err != nil {
		return fmt.Errorf("decode results: %w", err)
	}
	if err := json.Unmarshal(parts[1], &r.Pagination); err != nil {
		return fmt.Errorf("decode pagination: %w", err)
	}
	return nil
}

// Norma is one search result. Field names mirror the raw API response.
type Norma struct {
	IDNorma          int64  `json:"IDNORMA"`
	Norma            string `json:"NORMA"`
	Tipo             string `json:"DESCRIPCION"`
	TituloNorma      string `json:"TITULO_NORMA"`
	Organismo        string `json:"ORGANISMO"`
	FechaPublicacion string `json:"FECHA_PUBLICACION"`
	Resumen          string `json:"RESUMEN"`
}

// Pagination reports the total result count so callers can page through.
type Pagination struct {
	TotalItems int    `json:"totalitems"`
	Page       string `json:"npagina"`
	PageSize   string `json:"itemsporpagina"`
	Query      string `json:"cadena"`
}

// NormaFull is the full structured content of one norm.
type NormaFull struct {
	Html       []HtmlBlock      `json:"html"`
	Proyectos  []Proyecto       `json:"proyectos"`
	Metadatos  Metadatos        `json:"metadatos"`
	Estructura []EstructuraPart `json:"estructura"`
}

// HtmlBlock is one content part of the norm (header, article, annex...).
// Blocks NEST via H: the API returns titles and paragraphs with their
// articles as children (e.g. Ley 21.600) — a flat parse loses the articles.
// Markdown and SectionName are filled by ConvertContent, not the API.
type HtmlBlock struct {
	T           string      `json:"t"`
	I           int64       `json:"i"`
	V           []string    `json:"v,omitempty"`
	H           []HtmlBlock `json:"h,omitempty"`
	Markdown    string      `json:"-"`
	SectionName string      `json:"-"`
}

// EstructuraPart is one entry of the norm's table of contents. The TOC
// nests via H too: T (1=título, 4=párrafo, 6=artículo) classifies entries.
type EstructuraPart struct {
	N string           `json:"n"`
	I int64            `json:"i"`
	T int              `json:"t,omitempty"`
	H []EstructuraPart `json:"h,omitempty"`
}

// Proyecto links the norm to its legislative bill (boletín).
type Proyecto struct {
	Categoria string `json:"categoria"`
	Pls       []struct {
		Enlace      string `json:"enlace"`
		Informacion string `json:"informacion"`
		NroBoletin  string `json:"nroBoletin"`
	} `json:"pls"`
}

// Metadatos holds the selected norm metadata.
type Metadatos struct {
	TiposNumeros     []TipoNumero  `json:"tipos_numeros"`
	Organismos       []string      `json:"organismos"`
	TituloNorma      string        `json:"titulo_norma"`
	Fuente           string        `json:"fuente"`
	NumeroFuente     string        `json:"numero_fuente"`
	Materias         []string      `json:"materias"`
	CategoriasNorma  []string      `json:"categorias_norma"`
	Derogado         bool          `json:"derogado"`
	FechaPublicacion string        `json:"fecha_publicacion"`
	Vigencia         Vigencia      `json:"vigencia"`
	Vinculaciones    []Vinculacion `json:"vinculaciones"`
	Resumenes        []string      `json:"resumenes"`
}

// TipoNumero identifies the norm type and number (e.g. Ley 21600).
// All five API fields are kept: the structured content carries complete data.
type TipoNumero struct {
	Tipo        string `json:"tipo"`
	Numero      string `json:"numero"`
	Abreviacion string `json:"abreviacion"`
	Descripcion string `json:"descripcion"`
	Compuesto   string `json:"compuesto"`
}

// Vigencia is the norm's validity window.
type Vigencia struct {
	InicioVigencia string `json:"inicio_vigencia"`
	FinVigencia    string `json:"fin_vigencia"`
}

// Vinculacion is a relation to other norms (modification, concordance...).
type Vinculacion struct {
	Text    string `json:"text"`
	URLPart string `json:"url_part"`
}

// NormaSummary is the lightweight projection of a norm: the metadata fields
// that answer "what is this norm about", without the content.
type NormaSummary struct {
	TituloNorma     string   `json:"titulo_norma"`
	Fuente          string   `json:"fuente"`
	Materias        []string `json:"materias"`
	CategoriasNorma []string `json:"categorias_norma"`
	Resumenes       []string `json:"resumenes"`
}

// newConverter builds the shared HTML→Markdown converter. The standard
// plugins (base + commonmark + table) already render BCN's div.p markup
// as paragraphs — verified with real API HTML in a spike; no custom rules
// are needed.
func newConverter() *converter.Converter {
	return converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)
}

// ConvertContent converts every content block (recursively, blocks nest
// via H) to sanitized Markdown and attaches the section name from the
// structure tree (estructura[i] ↔ html[i]).
func (n *NormaFull) ConvertContent(conv *converter.Converter) {
	names := make(map[int64]string)
	var indexEstructura func(parts []EstructuraPart)
	indexEstructura = func(parts []EstructuraPart) {
		for _, part := range parts {
			names[part.I] = part.N
			indexEstructura(part.H)
		}
	}
	indexEstructura(n.Estructura)

	var convert func(blocks []HtmlBlock)
	convert = func(blocks []HtmlBlock) {
		for i := range blocks {
			md, err := conv.ConvertString(blocks[i].T)
			if err != nil {
				// Conversion errors are not fatal: fall back to sanitized raw text.
				md = blocks[i].T
			}
			blocks[i].Markdown = SanitizeMarkdown(md)
			blocks[i].SectionName = names[blocks[i].I]
			convert(blocks[i].H)
		}
	}
	convert(n.Html)
}

// retryConditions is the set of transient failures that trigger a retry:
// 5xx server errors and timeouts/connection failures (status zero).
// Declared explicitly — resty does not retry anything by default.
var retryConditions = []resty.RetryConditionFunc{
	resty.RetryConditionStatus5XX,
	resty.RetryConditionStatusZero,
}

// Search calls buscarjson with the query parameters, decodes the
// heterogeneous response and sanitizes every result summary.
func (c *Client) Search(ctx context.Context, params SearchParams) (SearchResponse, error) {
	res := c.resources.Resources[resourceSearchLaws]
	var result SearchResponse

	resp, err := c.clients[resourceSearchLaws].R().
		SetContext(ctx).
		SetQueryParams(map[string]string{
			"cadena":         params.Query,
			"npagina":        strconv.Itoa(params.Page),
			"itemsporpagina": strconv.Itoa(params.PageSize),
		}).
		SetRetryCount(res.Retry.Attempts).
		SetRetryWaitTime(time.Duration(res.Retry.Backoff)).
		SetRetryMaxWaitTime(time.Duration(res.Retry.MaxBackoff)).
		SetRetryConditions(retryConditions...).
		SetResult(&result).
		Get(res.Path)
	if err != nil {
		c.logger.Error("bcn search failed", "error", err)
		return SearchResponse{}, fmt.Errorf("search laws: %w", err)
	}
	if resp.IsStatusFailure() {
		return SearchResponse{}, fmt.Errorf("search laws: unexpected status %d", resp.StatusCode())
	}

	for i := range result.Results {
		result.Results[i].Resumen = SanitizeSummary(result.Results[i].Resumen)
	}
	c.logger.Debug("bcn search ok",
		"results", len(result.Results),
		"total_items", result.Pagination.TotalItems,
	)
	return result, nil
}

// GetNorma calls get_norma_json for one norm id, converts the content to
// Markdown and maps the HTTP 500 "not found" answer to ErrNormaNotFound.
//
// The response is cached per norm id with ETag revalidation: a repeated
// request sends If-None-Match and a 304 serves the cached (already
// converted) norm without re-downloading or re-converting.
func (c *Client) GetNorma(ctx context.Context, normID int64) (NormaFull, error) {
	res := c.resources.Resources[resourceGetLaw]
	entry, cached := c.cache.get(normID)

	req := c.clients[resourceGetLaw].R().
		SetContext(ctx).
		SetQueryParam("idNorma", strconv.FormatInt(normID, 10)).
		SetRetryCount(res.Retry.Attempts).
		SetRetryWaitTime(time.Duration(res.Retry.Backoff)).
		SetRetryMaxWaitTime(time.Duration(res.Retry.MaxBackoff)).
		SetRetryConditions(retryConditions...)
	if cached {
		req.SetHeader("If-None-Match", entry.etag)
	}

	var result NormaFull
	resp, err := req.SetResult(&result).Get(res.Path)

	// A 304 on a cached norm is a hit: serve the cached copy.
	if resp != nil && resp.StatusCode() == http.StatusNotModified && cached {
		c.logger.Debug("bcn get norm cache hit", "norm_id", normID)
		return entry.norma, nil
	}
	if err != nil {
		c.logger.Error("bcn get norm failed", "norm_id", normID, "error", err)
		return NormaFull{}, fmt.Errorf("get norm: %w", err)
	}
	if resp.StatusCode() == http.StatusInternalServerError {
		// The API answers nonexistent ids with HTTP 500 and a "no se
		// encuentra en nuestra Base de Datos" body.
		return NormaFull{}, ErrNormaNotFound
	}
	if resp.IsStatusFailure() {
		return NormaFull{}, fmt.Errorf("get norm: unexpected status %d", resp.StatusCode())
	}

	result.ConvertContent(c.converter)
	// The API delivers metadatos.resumenes with leading indentation;
	// sanitize it like every other text that reaches the LLM.
	for i := range result.Metadatos.Resumenes {
		result.Metadatos.Resumenes[i] = SanitizeSummary(result.Metadatos.Resumenes[i])
	}

	c.cache.put(normID, cacheEntry{etag: resp.Header().Get("ETag"), norma: result})
	return result, nil
}

// GetNormaSummary returns the lightweight metadata view of a norm. On a
// cache hit it derives the summary from the cached entry without any HTTP
// call; on a miss it goes through GetNorma (full flow with revalidation)
// and projects. The cache entry is shared, so a later get_law benefits.
func (c *Client) GetNormaSummary(ctx context.Context, normID int64) (NormaSummary, error) {
	if entry, ok := c.cache.get(normID); ok {
		c.logger.Debug("bcn norm summary cache hit", "norm_id", normID)
		return projectSummary(entry.norma), nil
	}
	norma, err := c.GetNorma(ctx, normID)
	if err != nil {
		return NormaSummary{}, err
	}
	return projectSummary(norma), nil
}

// projectSummary derives the lightweight view from a full norm. The
// resumenes are already sanitized by GetNorma.
func projectSummary(n NormaFull) NormaSummary {
	return NormaSummary{
		TituloNorma:     n.Metadatos.TituloNorma,
		Fuente:          n.Metadatos.Fuente,
		Materias:        n.Metadatos.Materias,
		CategoriasNorma: n.Metadatos.CategoriasNorma,
		Resumenes:       n.Metadatos.Resumenes,
	}
}
