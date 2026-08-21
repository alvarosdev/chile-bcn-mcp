## Context

Estado actual: MCP con dominio BCN (`internal/bcn`, `internal/config/api.resources.yaml` embed, `internal/tools` con 4 tools, `internal/prompts`). Transporte `resty` por recurso (breaker count-based client-level, retry por request), `etagCache` LRU 100 + `singleflight`, `converter` pool para HTML→Markdown, `sanitize.go` single-pass. Todo LLM-first (TextContent + structuredContent tipado, sin drift).

Contraloría expone `POST /apibusca/search/dictamenes` (7 campos: search, exact_search, options, order, date_name, source, page; 20 fijos por página, 0-indexed, total en `hits.total.value` con `relation:"gte"` a 10k) y `POST /apibusca/count/dictamenes` (2 campos: search, exact_search; `hits:[]` + `aggregations.count_by_type.buckets`). `GET /buscadorpdf/dictamenes/{id}/html` existe pero es SPA con `__APP_INITIAL_STATE__` que embebe el mismo `_source` — frágil vs `POST exact_search:true` que retorna idéntico `_source` con `total:1`. Ver proposal.md — Why.

## Goals / Non-Goals

**Goals:**
- Tres tools `search_cgr_dictamenes`, `get_cgr_dictamen`, `count_cgr_jurisprudencia` con el mismo nivel de optimización que BCN (YAML declarativo, resty por recurso, cache LRU, singleflight, sanitize, tests sin red).
- Clean directo: `documento_completo` ya es plain text; solo `normalize()` (sin converter pool), con fallback a `""` si vacío.
- Paginación fija 20 expuesta como `page` 1-indexed para LLM → `page-1` interno; `pagination{total,page,page_size:20,total_pages,has_more}`.
- Separación de dominios: `internal/cgr` exclusivo, sin mezclar tipos `Norma*` vs `Dictamen*`, sin dependencia `cgr → bcn`.

**Non-Goals:**
- Filtros `options` (siempre `[]` en MVP); `get_dictamen_summary` (p95 <10KB, no hace falta); `html-to-markdown` para CGR; ETag/304 para CGR (no lo envía); exponer `page_size`; mapear `source` distinto de `dictamenes`.

**1. Package `internal/cgr` separado vs extender `LawClient`:**
Elegido separado. `LawClient` con `Search/GetNorma/GetLawHistory` ya cohesiona BCN (int64 IDs, FlexInt, ETag, converter). Mezclar `Dictamen` (string IDs `E179593N25`, sin FlexInt, sin ETag, sin converter) rompe cohesión y acopla breaker por recurso. `main.go` crea `cgrClient := cgr.NewClient(resources, logger)` paralelo a `lawClient`.

**2. Fuente de `get_cgr_dictamen`: `POST /search exact_search:true` (no `GET /html`):**
Verificado: `POST {search:"E179593N25", exact_search:true}` retorna mismo `_source` que el JSON embebido en `/html`, con `total:1`. Evita brace-counting, scraping y fragilidad de SPA. Un solo resource `cgr_search` sirve a `search` y `get` (distinto body).

**3. Recursos YAML: `cgr_search` + `cgr_count` (POST ambos):**
```yaml
cgr_search:
  url: https://www.contraloria.cl
  path: /apibusca/search/dictamenes
  method: POST
  timeout: 10s
  retry: {attempts: 3, backoff: 500ms, max_backoff: 5s}
  circuit_breaker: {failure_threshold: 5, success_threshold: 2, reset_timeout: 30s}
cgr_count:
  url: https://www.contraloria.cl
  path: /apibusca/count/dictamenes
  method: POST
  timeout: 10s
  retry: {attempts: 2, backoff: 500ms, max_backoff: 4s}
  circuit_breaker: {failure_threshold: 5, success_threshold: 2, reset_timeout: 30s}
```
Copia thresholds de BCN (`search_laws`/`get_law`). `cgr_count` más barato, `attempts:2` basta. Headers mínimos: `Accept`, `Content-Type: application/json`, `Origin: https://www.contraloria.cl` (cookie `JSESSIONID` no requerida, verificado).

**4. Paginación CGR vs BCN:**
BCN: `GET ?cadena=&npagina=1&itemsporpagina=10` (1-indexed, configurable). CGR: `POST {page:0}` (0-indexed, 20 fijo, `size` ignorado). Tool expone `page` 1-indexed; mapeo `apiPage = max(0, page-1)`. `total_pages = ceil(total/20)`, `has_more = page*20 < total`. Beyond → `hits:[]` sin error.

**5. Orden y exact_search expuestos:**
`order: "date"|"dateasc"|"score"` (date DESC default, dateasc ASC, score relevance DESC) + `exact_search?: bool` default false. `date_name:"fecha_documento"` y `source:"dictamenes"` fijos internos; `options:[]` fijo (sin filtros MVP).

**6. Clean directo + sanitize duplicado — shape fino A:**
`documento_completo` es plain con `\u00A0`/`&nbsp;` igual que BCN. Solo `normalize()` (espacios, zero-width, C0). `garbage.go` duplicado en `cgr` para evitar `cgr → bcn`. Sin `converter` pool (ahorra mutex y allocs). Fallback: si `documento_completo == ""`, retornar `""` con aviso en texto (no convertir `raw` en MVP). Tipos: `DictamenSummary` (paridad con `NormaSummary`, ex `Brief`) embebido en `DictamenFull`; `FuentesLegales` como `string` csv (no `[]string`, preserva `art/` sin split); `wire` separado `cgrHit`/`cgrSource` para aislar sobre ES (`_id`, `_score`, `materia_raw` ignorado).

**7. Cache CGR: LRU 100 + singleflight, sin ETag:**
CGR no envía `ETag`/`304`. `etagCache[T]` se simplifica a `lruCache[T]` (misma LRU, sin `If-None-Match`). Claves: `search:"search:{query}|{exact}|{order}|{page}"`, `dictamen:"dictamen:{id}"`, `count:"count:{query}|{exact}"`. `singleflight` coalescea concurrentes al mismo dictamen/query. Sin TTL (datos inmutables por ID; LRU eviction basta).

**8. Nombres (interfaz inglés, dato español):**
`search_cgr_dictamenes`, `get_cgr_dictamen`, `count_cgr_jurisprudencia` (prefijo `cgr_`, verbo inglés, sustantivo dominio en español, como `search_laws` + `norm_id`→`idNorma`). Args: `query`, `exact_search`, `order`, `page`; `dictamen_id` (string `^[A-Z]*[0-9]+N[0-9]{2}$` laxa, ej `E179593N25`, `OF80660N26`). Output mínimo + `carácter` (sin `boletin`/`year_doc_id`, redundante con `fecha_documento`).

**9. LLM-first dual output — separado puro para counts:**
Cada handler retorna `*mcp.CallToolResult` (texto) + `Output` struct tipado (structuredContent). Texto es vista del structured (sin drift), con `dictamen_id` para drill-down, `pagination` y `char_count`. `search_cgr_dictamenes` retorna `{results:[]DictamenSummary, pagination}` sin `counts` inline en MVP (separado puro); `counts` queda en `count_cgr_jurisprudencia`. Migración futura a `counts,omitempty` en search es aditiva sin breaking.
**10. Validación, errores y URLs de citación:**
Vacío/ausente `dictamen_id` → `errorResult("dictamen_id is required")` sin HTTP. Formato válido pero `total==0` → `ErrDictamenNotFound` → `errorResult("dictamen E999N99 not found")`. `order` fuera de `date|dateasc|score` o `page<=0` → error de argumentos sin HTTP. `get_cgr_dictamen` y cada `DictamenSummary` en `search` incluyen `url` (`https://www.contraloria.cl/buscadorpdf/dictamenes/{id}/html`) y `pdf_url` (`https://www.contraloria.cl/buscadorpdf/dictamenes/{id}/pdf`) derivados por interpolación (no fetch); el texto sugiere ambas para citación (HTML) y descarga (PDF). `char_count` + `url`/`pdf_url` siempre presentes.

## Risks / Trade-offs

- **Sin ETag:** cache puede servir stale si CGR edita un dictamen (raro, dictámenes son inmutables). Mitiga LRU pequeño y sin TTL largo.
- **20 fijos:** LLM no puede pedir 5 o 50; debe paginar. Documentar en description.
- **`GET /html` descartado:** si `POST exact_search` dejara de indexar por `doc_id`, fallback sería revivir `GET /html` con brace-counting. Riesgo bajo (verificado con 3 IDs distintos).
- **Duplicar sanitize:** DRY violado a cambio de desacoplo. Si `garbage.go` cambia, hay que cambiar 2 lugares. Alternativa `internal/sanitize` compartido se evaluó y se descartó por simplicidad MVP.
- **2 recursos POST:** `resty` con `SetBody` para JSON; breaker por resource implica `cgr_search` y `cgr_count` no comparten ventana de fallos (deseable).
- **`total relation gte` a 10k:** para `search:""` vacío, avisar en texto "más de 10.000, refina query".

## Migration Plan

Aditivo, sin breaking. Orden de implementación: `api.resources.yaml` → `internal/cgr/types + cache + sanitize` → `internal/cgr/cgr_client.go` (Search/Count/Get) → `internal/tools` (3 handlers) → `main.go` wiring → tests + fixtures → docs. Rollback: quitar registro de tools y YAML, resto queda inerte.
