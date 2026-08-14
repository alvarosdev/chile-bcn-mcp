## Why

El scaffold solo tiene la tool demo `echo`. El dominio real de chile-bcn-mcp es la **Biblioteca del Congreso Nacional de Chile (BCN / LeyChile)**: buscar normas (leyes, decretos, resoluciones) de forma paginada y navegar al contenido de una norma por su `idNorma`. Los endpoints de BCN deben declararse en un archivo `api.resources.yaml` (url, path, method, timeout, retry, circuit breaker **por endpoint**) para no hardcodear URLs en el código, con carga única al arranque.

## What Changes

- Crear `config/api.resources.yaml`: contrato plano y versionado (`version: 1`) con los recursos `search_laws` (JSON paginado en `nuevo.leychile.cl`) y `get_law` (contenido estructurado en `nuevo.leychile.cl`), cada uno con `url`, `path`, `method`, `timeout`, `retry` (attempts/backoff) y `circuit_breaker` count-based.
- Crear `internal/config`: loader de YAML con validación fail-fast al arranque, cargado **una sola vez** e inyectado por constructor (instancia única del proceso, sin estado global). La ruta es **fija** (`config/api.resources.yaml`, sin override por env): el contrato no es secreto y los cambios de configuración se despliegan rebuildando la imagen, no con hot-reload.
- Crear `internal/bcn`: cliente HTTP con `go-resty/resty` v3 — un `*resty.Client` por resource (el circuit breaker de resty es client-level), retry per-request desde el YAML, `SetResult` para el JSON, hooks de logging con `slog`.
- Crear las tools MCP `search_laws` (búsqueda paginada: `query`, `page`, `page_size`; devuelve resultados + total para navegar) y `get_law` (contenido estructurado de una norma por `norm_id` vía `get_norma_json`: metadatos seleccionados, estructura/TOC, proyectos y contenido completo convertido a **Markdown**; con opción `structure_only` para explorar sin el contenido). El campo `RESUMEN` de la búsqueda se devuelve **limpio para el LLM** (`html.UnescapeString` + strip de tags XML). Un `norm_id` inexistente se reporta como error de "norma no encontrada".
- **Idioma de la interfaz (convención del proyecto)**: nombres de tools MCP, argumentos, nombres de archivos, descripciones de tools y comentarios de código en **inglés**; los datos crudos del dominio (texto legal, campos del response de BCN) se entregan tal cual vienen.
- **Norma LLM-first + structured opcional**: toda tool devuelve texto formateado para el modelo en `content[]` **y** `structuredContent` tipado (schema autogenerado por el go-sdk) — el texto es una vista del structured, con los datos completos en el structured cuando el texto trunca para ahorrar tokens.
- **Caché de normas con revalidación ETag** en el cliente (solo `get_law`): `If-None-Match` ante re-consultas del mismo `norm_id`, `304` sirve desde caché sin re-descargar ni re-convertir. Verificado contra la API real (responde `ETag` y `304` con cuerpo vacío).
- Dependencias nuevas: `resty/v3`, `yaml.v3` (ya indirecta → directa), `testify` (ya indirecta → directa), `html-to-markdown/v2` (conversión del contenido de la norma) y `mockery/v3` como tool de Go.
- Tests sin llamadas externas: **suite** con `testify/suite`, mocks generados con **mockery** (template testify, `{{.InterfaceDir}}`, mock al lado del archivo productivo), y `httptest.Server` local para el cliente real. Nombres de archivos en snake_case.

## Capabilities

### New Capabilities

- `leychile-search`: Búsqueda paginada de normas chilenas en LeyChile (BCN) y acceso al contenido estructurado de una norma por su identificador (metadatos, estructura y texto en Markdown), con endpoints declarados en `api.resources.yaml` y transporte resiliente (retry + circuit breaker por endpoint).

### Modified Capabilities

<!-- Ninguna: el spec mcp-server existente no cambia su comportamiento. -->

## Impact

- **Código**: archivos nuevos en `config/`, `internal/config/`, `internal/bcn/`, `internal/tools/`; `main.go` agrega el bootstrap (Load → NewLawClient → RegisterTools). No se toca el transporte del server ni la auth.
- **Dependencias**: `go-resty/resty/v3`, `JohannesKaufmann/html-to-markdown/v2` (nuevas), `gopkg.in/yaml.v3` y `stretchr/testify` (indirectas → directas), `vektra/mockery/v3` (tool). `sony/gobreaker` **no** se agrega (breaker built-in de resty).
- **Sin cambios breaking**: la tool `echo` se conserva; las tools nuevas son aditivas.
