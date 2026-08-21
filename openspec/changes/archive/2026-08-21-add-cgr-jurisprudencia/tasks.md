## 1. Contrato y cliente base

- [x] 1.1 Extender `internal/config/api.resources.yaml` con `cgr_search` (POST /apibusca/search/dictamenes, 10s, retry 3, breaker 5/2/30s) y `cgr_count` (POST /apibusca/count/dictamenes, 10s, retry 2, breaker 5/2/30s); validar con `LoadEmbedded` y `resources_test.go`
- [x] 1.2 Crear `internal/cgr/cgr_client.go` con `CgrClient` interface (`SearchDictamenes`, `GetDictamen`, `CountJurisprudencia`), `Client` con `map[string]*resty.Client` por resource, `singleflight.Group`, `lruCache[T]` (LRU 100, sin ETag) y `newRestyClient` reuse; `internal/cgr/types.go` con `DictamenSummary` (ex Brief, embed en `DictamenFull`), `SearchResponse` sin `counts` inline, `CountResponse`, `Pagination`, `wire` `cgrHit`/`cgrSource` + `FuentesLegales string` csv
- [x] 1.3 Duplicar `internal/cgr/sanitize.go` + `garbage.go` (normalize single-pass, constantes nombradas, sin converter pool) y exponer `SanitizeDocumento`/`SanitizeMateria`

## 2. Handlers MCP (LLM-first dual output)

- [x] 2.1 Implementar `search_cgr_dictamenes` en `internal/tools/search_cgr_dictamenes.go` (o `internal/cgr/tools/`) con args `query`, `exact_search?`, `order?` (date|dateasc|score), `page?` (1-indexed → 0-indexed), validación, `buildOutput` + `formatSearchResults`, registro `RegisterSearchCgrDictamenes`
- [x] 2.2 Implementar `get_cgr_dictamen` con arg `dictamen_id` (regex `^[A-Z]*[0-9]+N[0-9]{2}$`), consulta `cgr_search` exact, sanitización clean directo, `char_count` + `url` canónica, texto `## Documento Completo`
- [x] 2.3 Implementar `count_cgr_jurisprudencia` con args `query`, `exact_search?`, proyección de `buckets` y texto agregado
- [x] 2.4 Registrar las 3 tools en `RegisterTools` y cablear `CgrClient` en `cmd/chile-bcn-mcp/main.go` (segundo singleton junto a `LawClient`)

## 3. Tests

- [x] 3.1 `internal/cgr/cgr_client_test.go`: suite testify con `httptest.Server` para `cgr_search` y `cgr_count` (fixtures `testdata/` con respuestas reales truncadas), casos: paginación 0/1/beyond, order variants, total `gte` 10k, retry 5xx, breaker, singleflight coalescing, LRU eviction
- [x] 3.2 `internal/tools/*_test.go`: mocks `MockCgrClient` con `mock.Anything` para ctx, tests de handlers (dual output text vs structured sin drift, `dictamen_id` validación, `page` mapping, texto truncado vs structured completo)
- [x] 3.3 Regenerar mocks `internal/cgr/cgr_client_mock.go` con `mockery` (template testify, snake_case) y verificar `go vet`

## 4. Documentación y validación

- [x] 4.1 Actualizar `README.md` (tabla de tools), `internal/config/api.resources.yaml` docs y `openspec/specs/cgr-*/spec.md` si aplica tras archive
- [x] 4.2 Ejecutar `make check` (build+vet+test), `make fmt-check` y `openspec validate --strict` sobre el change
- [x] 4.3 Smoke manual con `podman`/`compose`: `search_cgr_dictamenes` (quillota page 1/2), `get_cgr_dictamen` (E179593N25), `count_cgr_jurisprudencia` (quillota, bono) y verificar `documento_completo` sanitizado
