## Why

Los agentes que consultan leyes largas terminan descargando la norma completa — Ley 21.600 son ~426K caracteres (~150K tokens) — aunque solo necesiten un título o un artículo. Hoy `get_law` solo puede devolver "todo" o "solo estructura" (`structure_only`), no hay forma de pedir una parte puntual, y el agente no conoce el tamaño de lo que va a pedir antes de disparar la llamada. El resultado observado: el agente descarga el texto completo, se ahoga, y tiene que re-fetchear para recuperarse.

## What Changes

- `get_law` acepta un parámetro opcional `section_id` que devuelve solo el subárbol de esa parte (título/capítulo/artículo) en vez de la norma completa.
- `get_law` y `get_law_summary` incluyen `char_count` y `article_count` en su output, que describen el contenido devuelto: la sección cuando se pide `section_id`, la norma completa si no; el summary siempre reporta el total. El summary además incluye la estructura de la norma (para encadenar el drill-down sin otra llamada).
- Las descripciones de las tools documentan el flujo barato (resumen → estructura → sección) para que el agente lo siga por defecto.
- Nuevo prompt curado `law_research_workflow` que codifica ese flujo desde el arranque.
- Robustez y endurecimiento (sin cambio de comportamiento MCP visible): pool de converters HTML→Markdown (hoy el mutex de la librería serializa todas las conversiones), singleflight por norma@versión (requests concurrentes comparten una sola llamada a BCN), evicción LRU del caché ETag, tope de tamaño de respuesta upstream (guard OOM), `ReadHeaderTimeout` en el servidor HTTP, contenedor endurecido (rootfs read-only, `cap_drop: ALL`, `no-new-privileges`), `govulncheck` en CI, `dependabot.yml`, fuzz targets para el sanitizer/converter y test de concurrencia 200/304 del caché.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `leychile-search`: `get_law` gana `section_id` (recuperación parcial por parte de la norma), `get_law`/`get_law_summary` exponen el tamaño (`char_count`, `article_count`) y `get_law_summary` incluye la estructura en su structured output.
- `law-prompts`: nuevo prompt curado `law_research_workflow` que guía el flujo resumen → estructura → sección.
- `container-deployment`: el contenedor se ejecuta endurecido (rootfs read-only, capacidades mínimas, no-new-privileges) en las configuraciones de despliegue provistas.

## Impact

- `internal/tools/get_law.go` — nuevo arg `section_id`, slice del árbol de contenido, campos de tamaño en el output.
- `internal/tools/get_law_summary.go` — campos de tamaño en el output.
- `internal/bcn` — helpers para recorrer el árbol ya parseado (contar artículos/caracteres, extraer un subárbol por `i`); reutilizan el `NormaFull` ya cacheado. También: pool de converters, singleflight, LRU y tope de respuesta.
- `internal/prompts/prompts.go` — séptimo prompt curado.
- `cmd/chile-bcn-mcp/main.go` — `ReadHeaderTimeout`.
- `docker-compose.yml` + Makefile (`podman-run`) — contenedor endurecido.
- `.github/workflows/ci.yml` — paso `govulncheck`; `.github/dependabot.yml` — nuevo.
- **Sin cambios** en `config/api.resources.yaml`: se usa el mismo endpoint (`get_norma_json`). Verificado contra la API real que `idParte` es **ignorado** (devuelve la norma completa), por lo que el filtrado es local sobre el documento ya descargado y cacheado — no un parámetro nuevo al servicio.
- **Sin dependencias nuevas, sin storage nuevo**: no se introduce índice vectorial ni persistencia en disco; `golang.org/x/sync` (singleflight) ya estaba como dependencia indirecta.
