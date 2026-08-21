## Why

Los agentes de IA que usan este MCP hoy solo acceden a normas de LeyChile (BCN). La jurisprudencia administrativa de la Contraloría General de la República (dictámenes) es la segunda fuente más consultada para interpretar y aplicar esas normas, pero no es accesible desde el MCP. Contraloría expone búsqueda y fichas vía `contraloria.cl/apibusca` con un contrato distinto a BCN (POST JSON, paginación fija, agregaciones). Integrarla con las mismas optimizaciones que BCN (contrato YAML, resty por recurso, cache LRU, singleflight, sanitize, LLM-first dual output) evita que los agentes inventen dictámenes o hagan scraping frágil.

## What Changes

- Nuevo package `internal/cgr` con cliente exclusivo para Contraloría (separado de `internal/bcn`), con 1 `*resty.Client` por recurso, retry/breaker declarativos, LRU 100 + singleflight sin ETag, y sanitización duplicada (clean directo sobre `documento_completo`).
- Dos recursos en `internal/config/api.resources.yaml` (embed): `cgr_search` (`POST /apibusca/search/dictamenes`) y `cgr_count` (`POST /apibusca/count/dictamenes`), con `timeout`, `retry` y `circuit_breaker` copiados de BCN.
- Tres tools MCP nuevas con prefijo `cgr_` y nombres en inglés, dato en español (patrón `norm_id` → `idNorma`):
  - `search_cgr_dictamenes` — búsqueda paginada (query, exact_search, order, page; 20 fijos por página) sobre `cgr_search`.
  - `get_cgr_dictamen` — ficha por `dictamen_id` (string `E179593N25`) vía `POST /search` con `exact_search:true` (fuente `documento_completo` clean, sin HTML→Markdown).
  - `count_cgr_jurisprudencia` — conteo cross-tipo por query (agregaciones `count_by_type`) sobre `cgr_count`.
- `internal/cgr/sanitize.go` + `garbage.go` duplicados (sin dependencia `cgr → bcn`), `dictamen` types, y cache LRU. Sin pool de `html-to-markdown` para CGR (clean directo).
- Registro en `internal/tools` (o `internal/cgr/tools`) e inyección de `CgrClient` en `main.go` junto al `LawClient` existente. Tests con `httptest.Server` + fixtures y mocks `mock.Anything` para ctx.

## Capabilities

### New Capabilities

- `cgr-search`: búsqueda paginada de dictámenes de Contraloría con paginación fija y ordenamiento.
- `cgr-dictamen`: obtención de ficha completa de un dictamen por su ID opaco, con metadatos y documento completo sanitizado.
- `cgr-count`: conteo agregado cross-tipo (dictamenes, auditoria, legislacion, etc.) por query para exploración previa.

### Modified Capabilities

- Ninguna existente cambia requisitos; `mcp-server` amplía el registro de tools (aditivo, sin breaking changes). `container-deployment` y `release-distributions` no cambian contrato, solo embeben nuevo YAML.

## Impact

- Código: `internal/config/api.resources.yaml`, `internal/cgr/*` (cliente, types, cache, sanitize), `internal/tools/*` o `internal/cgr/tools/*` (3 handlers), `cmd/chile-bcn-mcp/main.go` (segundo singleton), `internal/bcn` sin cambios.
- Tests: `internal/cgr/*_test.go`, `internal/tools/*_test.go`, fixtures `testdata/`, `law_client_mock.go` regenerado o `cgr_client_mock.go` nuevo.
- Docs: `README.md` (nuevas tools), `openspec/specs/cgr-*/spec.md`.
- Sin breaking changes; sin nuevos deps (reusa `resty`, `yaml`, `testify`); sin html-to-markdown para CGR.
