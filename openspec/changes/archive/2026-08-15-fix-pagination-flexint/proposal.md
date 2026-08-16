## Why

`search_laws` falla de forma intermitente según la query. `search_laws(query="Ley 21.600")` funciona, pero `search_laws(query="Ley 21461")` y `search_laws(query="21461")` retornan `search failed: decode pagination: json: cannot unmarshal number into Go struct field Pagination.itemsporpagina of type string`. La API de LeyChile devuelve paginación inconsistente: `itemsporpagina` (y análogamente `npagina`/`totalitems`) a veces como string `"10"` y a veces como número `10`. El struct `Pagination` en `internal/bcn/law_client.go:160` asume un solo tipo y rompe el decode, bloqueando búsquedas válidas como Ley 21461 (`norm_id: 1178004`).

## What Changes

- Crear tipo `FlexInt` en `internal/bcn` que deserialice `string` o `number` (con trim, tolerancia a `float` 10.0, `""` y `null` → 0) e implemente `MarshalJSON` devolviendo número.
- Migrar `Pagination` para usar `FlexInt` en los campos numéricos `npagina`, `itemsporpagina` y `totalitems` (manteniendo `cadena` como `string`), sin cambiar el contrato MCP.
- Actualizar `internal/bcn/law_client.go:130-146` (`SearchResponse.UnmarshalJSON` y `Client.Search`) para que `search_laws` sea resiliente a ambos formatos sin `decode pagination`.
- Actualizar literals y expectativas en `internal/bcn/law_client_test.go` y `internal/tools/search_laws_test.go` (`Pagination{Page:"1"}` → `FlexInt(1)`).
- Añadir tests unitarios de deserialización para ambos JSON (`{"itemsporpagina":10}` y `{"itemsporpagina":"10"}`) incluyendo mixto, trim y float/null.
- Verificar regresión: `Ley 21.600` → 1195666 sigue funcionando; `Ley 21461`/`21461` → 1178004 sin error; `get_law_summary(norm_id=1178004)` sin regresión.

## Capabilities

### New Capabilities

<!-- No se crean capacidades nuevas; el fix es resiliencia dentro de búsqueda existente -->

### Modified Capabilities

- `leychile-search`: Requisito "Búsqueda paginada de normas" — la paginación debe tolerar `string|number` en `npagina`, `itemsporpagina` y `totalitems` sin fallar el decode. El contrato `search_laws` (query/page/page_size → results + total_items/total_pages) no cambia.

## Impact

- **Código**: `internal/bcn/law_client.go` (struct `Pagination` y `SearchResponse`), nuevo `internal/bcn/flexint.go`, `internal/bcn/law_client_test.go`, `internal/tools/search_laws_test.go` y `internal/tools/search_laws.go` (si requiere conversión `int(FlexInt)`).
- **API MCP**: Sin breaking change; `search_laws` mantiene `SearchLawsArgs`/`SearchLawsOutput` (`int` en `total_items`/`total_pages`); `FlexInt` es interno a `internal/bcn`.
- **Tests/Fixtures**: Nuevo `flexint_test.go`; `testdata/search_response.json` permanece como regresión string; nuevos casos httptest con paginación numérica.
- **Dependencias**: Ninguna nueva; usa `encoding/json`, `strconv`, `strings` estándar. Respeto a `resty v3` retry/breaker y a sanitización existente.
