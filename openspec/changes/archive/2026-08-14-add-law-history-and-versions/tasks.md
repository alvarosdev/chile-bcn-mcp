## 1. Catálogo de tipos

- [x] 1.1 Crear `internal/bcn/norm_types.go`: `map[int]normType{Abbr, Valor}` con los 40 tipos oficiales (cod → abbr/valor, comentario de origen `/Consulta/getTiposNorma`), lookup `canonicalNormType(cod int) (valor, abbr string, ok bool)`
- [x] 1.2 Escribir `internal/bcn/norm_types_test.go` (suite): lookups conocidos (1→Ley/LEY, 2→Decreto/DTO, 3→Resolución/RES), código desconocido → ok=false

## 2. Modelo y cliente

- [x] 2.1 Refactor del caché a `etagCache[T]` genérico (mismo patrón: mutex + map + cap 100) en `internal/bcn/norma_cache.go`; instancias `normas etagCache[NormaFull]` y `historias etagCache[[]HistoriaGrupo]` en Client
- [x] 2.2 Introducir `NormaQuery{NormID int64, VersionDate string}` y cambiar `LawClient`: `GetNorma(ctx, NormaQuery)`, `GetNormaSummary(ctx, NormaQuery)`, nuevo `GetLawHistory(ctx, normID int64) ([]HistoriaGrupo, error)`
- [x] 2.3 Clave compuesta del caché de normas: `fmt.Sprintf("%d@%s", normID, versionDate)`; `GetNorma` agrega `SetQueryParam("idVersion", q.VersionDate)` solo cuando viene; historias con `If-None-Match` + 304 por normID
- [x] 2.4 Agregar `CanonicalType`/`CanonicalAbbr` (`omitempty`) a `TipoNumero` y llenarlos en el pipeline de `GetNorma` vía `canonicalNormType` (append, sin reemplazar raw)
- [x] 2.5 Modelar `HistoriaGrupo{Titulo, TipoDesc, TipoCod, Hls []HistoriaEntrada}` y `HistoriaEntrada{Tipo, IDNorma, Enlace, Bajada, Fecha, Descripcion, IDNormaHL}`; `GetLawHistory` con recurso `get_law_history` del YAML, retry desde YAML, 304→caché, `[]`→slice vacío sin error
- [x] 2.6 Agregar `get_law_history` a `config/api.resources.yaml` (url nuevo.leychile.cl, path `/servicios/Navegar/get_historias_de_ley`, GET, timeout 10s, retry 2/500ms/4s, breaker 5/2/30s)
- [x] 2.7 Extender `internal/bcn/law_client_test.go` (suite + httptest): versiones históricas distintas por `idVersion` (fixture `norma_full_v2010.json` real de la Ley 19.628), clave compuesta no mezcla versiones (2 requests, 2 entradas), 304 de historias sirve caché, `[]` → slice vacío, canonical_type/abbr en tipos

## 3. Tools MCP

- [x] 3.1 `get_law.go`: arg `VersionDate string `json:"version_date,omitempty" jsonschema:"version in force at this date (YYYY-MM-DD, optional)"`; validación estricta `time.Parse("2006-01-02", ...)` → error de argumentos; header del texto con `Version: as of <fecha>` cuando viene; output estructurado con `version_date` solicitado
- [x] 3.2 `get_law_summary.go`: mismo arg `version_date` con validación estricta, passthrough a `NormaQuery`, "Version: as of …" en el texto
- [x] 3.3 Crear `internal/tools/get_law_history.go`: `GetLawHistoryArgs{NormID int64}`, `RegisterGetLawHistory`, handler valida `NormID > 0`, `[]`→mensaje amable ("no legislative history found"), texto LLM-first + structured `[]bcn.HistoriaGrupo`; el texto construye el enlace a LeyChile con `id_norma_hl` (`https://www.leychile.cl/Navegar?idNorma=<id_norma_hl>`) — nunca con `id_norma` ni con el número del `enlace` (Historia ID)
- [x] 3.4 Registrar `get_law_history` en `RegisterTools`
- [x] 3.5 Regenerar el mock (`make mock`) y actualizar los tests de tools existentes (firma nueva de GetNorma/GetNormaSummary)
- [x] 3.6 Escribir `internal/tools/get_law_history_test.go` (suite + MockLawClient): historia válida (3 grupos en texto + structured), vacía → mensaje amable, norm_id inválido → error sin llamar al cliente; extender `get_law_test.go`/`get_law_summary_test.go` con casos de version_date (válida, inválida → error de argumentos sin llamar al cliente, header "as of")

## 4. Verificación y documentación

- [x] 4.1 `make check` en verde
- [x] 4.2 Smoke real (no en CI): `get_law` con `version_date` de la Ley 19.628 (verificar texto distinto + "Version: as of"), `get_law_history` de la 21.600 (3 grupos con 21.770/21.755 modificatorias), `get_law_summary` con version_date
- [x] 4.3 Actualizar `README.md`: tabla de argumentos de `get_law` y `get_law_summary` con la fila `version_date` (opcional, YYYY-MM-DD, validación estricta), nueva sección `get_law_history` con su tabla de argumentos (historia propia/modificatorias/modificadas), párrafo de caching actualizado (clave compuesta `norm_id@version_date` + historias con ETag), ejemplo en Sample Usage de versión histórica ("¿Qué decía la Ley 19.628 en 2010?")
