## Why

LeyChile permite recuperar **versiones históricas** de una norma (`idVersion=YYYY-MM-DD` selecciona el texto vigente a esa fecha — verificado con la API real: la Ley 19.628 cambia de 41.4K a 36.7K chars según la fecha) y consultar la **historia legislativa** (`get_historias_de_ley`: la historia propia, las leyes que la modificaron y las que ella modificó). Hoy `get_law` y `get_law_summary` solo devuelven la última versión, y no hay tool para ver quién modificó qué. Además, los códigos numéricos de tipo de norma (p.ej. `tipos_numeros[].tipo = "1"`) no se decodifican — el catálogo oficial de BCN (cod → abbr → valor, 40 tipos) no está mapeado en el código.

## What Changes

- **`version_date` en `get_law` y `get_law_summary`** (parámetro opcional, YYYY-MM-DD, validación estricta): selecciona la versión vigente a esa fecha; sin valor → última versión. El texto indica la versión mostrada ("Version: as of 2010-01-01").
- **Nueva tool `get_law_history(norm_id)`**: los 3 grupos de historia legislativa (propia, modificatorias, modificadas) con sus entradas tipadas; un `norm_id` inexistente devuelve un mensaje amable (la API responde `[]`, no 500). Caché ETag **desde el día 1** (verificado: el endpoint envía ETag y responde 304).
- **Catálogo de tipos de norma** en `internal/bcn/norm_types.go` (hardcodeado, patrón `garbage.go`): cod → abbr/valor para los 40 tipos oficiales. Se usa para **decodificar sin reemplazar**: `TipoNumero` gana los campos append `canonical_type` y `canonical_abbr` (omitempty) — los valores crudos de la API quedan intactos.
- **Caché generalizado con clave compuesta**: el caché ETag de normas pasa a clave `(norm_id, version_date)` — hoy la clave es solo `norm_id` y una versión histórica recibiría la respuesta cacheada de otra versión (los ETags difieren por versión, verificado). Refactor a un `etagCache[T]` genérico que sirve tanto normas como historias.
- **Nuevo recurso en `api.resources.yaml`**: `get_law_history` (path `/servicios/Navegar/get_historias_de_ley`) con su propio timeout/retry/breaker.

## Capabilities

### New Capabilities

<!-- Ninguna: la capability leychile-search ya existe en las specs principales. -->

### Modified Capabilities

- `leychile-search`: requirements ADDED — versión histórica por fecha (get_law + get_law_summary) e historia legislativa (get_law_history), más los campos canonical de tipo de norma.

## Impact

- **Código**: `internal/bcn` (NormaQuery, caché genérico con clave compuesta, GetLawHistory + modelo, catálogo norm_types.go, campos canonical en TipoNumero), `internal/tools` (args con `version_date`, tool nueva, registro), `config/api.resources.yaml` (recurso nuevo), mock regenerado (`make mock`).
- **Compatibilidad**: `LawClient` cambia su firma (`GetNorma`/`GetNormaSummary` pasan a `NormaQuery`) — cambio interno, los consumidores externos (tools MCP) no se afectan; `get_law`/`get_law_summary` ganan un parámetro opcional (aditivo).
- **Dependencias**: ninguna nueva.
