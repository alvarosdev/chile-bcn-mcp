## 1. Dependencias y configuración mockery

- [x] 1.1 Agregar dependencias directas: `github.com/go-resty/resty/v3`, `gopkg.in/yaml.v3`, `github.com/stretchr/testify` y `tool github.com/vektra/mockery/v3` en `go.mod` (`go get` + `go mod tidy`)
- [x] 1.2 Crear `.mockery.yml` (patrón godot-mcp-docs): `all: false`, `template: testify`, paquete `internal/bcn`, interfaz `LawClient` con `dir: "{{.InterfaceDir}}"`, `filename: "{{.InterfaceName | snakecase}}_mock.go"`, `structname: "Mock{{.InterfaceName}}"`

## 2. Config YAML

- [x] 2.1 Crear `config/api.resources.yaml` con `version: 1` y los recursos `search_laws` (url `https://nuevo.leychile.cl`, path `/servicios/buscarjson`, GET, timeout 10s, retry 3/500ms/5s, breaker 5/2/30s) y `get_law` (url `https://nuevo.leychile.cl`, path `/servicios/Navegar/get_norma_json`, GET, timeout 15s, retry 2/1s/4s, breaker 5/2/30s)
- [x] 2.2 Implementar `internal/config/resources.go`: structs tipados (`Resources`, `Resource`, `Retry`, `CircuitBreaker`) con tags yaml, `Load(path)` que lee, valida fail-fast (path requerido, method GET/POST válido, timeouts > 0, attempts ≥ 1, backoff ≤ max_backoff, umbrales de breaker > 0) y devuelve `*Resources`
- [x] 2.3 Escribir `internal/config/resources_test.go` (suite testify): carga válida, YAML malformado, recurso sin path, timeout negativo — con fixtures en `internal/config/testdata/`

## 3. Cliente BCN

- [x] 3.1 Implementar `internal/bcn/law_client.go`: interfaz `LawClient` (`Search(ctx, SearchParams) (SearchResponse, error)`, `GetNorma(ctx, normID int64) (NormaFull, error)`), `NewLawClient(resources)` que crea un `*resty.Client` por recurso con breaker `NewCircuitBreakerCount` + `SetCircuitBreaker` y hooks de logging con `slog`
- [x] 3.2 Crear `internal/bcn/garbage.go` con toda la basura definida (sin magic strings/numbers): constantes runa nombradas (`nbspRune`, `enspRune`, `emspRune`, `zeroWidthSpaceRune`, `bomRune`), límites de control nombrados (`controlMin`/`controlMax`), listas declarativas (`spacesToNormalize`, `zeroWidthChars`) y `resumenTagRe` (regexp nombrada del wrapper XML), cada una con comentario de origen en la API de BCN
- [x] 3.3 Implementar `internal/bcn/sanitize.go` (sanitizer central que consume `garbage.go`): `html.UnescapeString` → (strip de tags XML para RESUMEN / conversión Markdown para contenido) → espacios de `spacesToNormalize` → espacio → eliminar control chars (`controlMin`/`controlMax`) y `zeroWidthChars` → trim por línea + colapso de espacios; **sin tocar** comillas ni enlaces
- [x] 3.4 Implementar `Search`: `SetResult` + retry per-request desde el YAML (`SetRetryCount/WaitTime/MaxWaitTime`), query params tipados; deserialización del response heterogéneo `[resultados, paginación, facets]` con `UnmarshalJSON` custom → `SearchResponse{Results []Norma, Pagination, Facets}`; campo `RESUMEN` pasado por el sanitizer
- [x] 3.5 Implementar `GetNorma`: GET al recurso `get_law` con query `idNorma`, retry desde YAML; modelar `NormaFull` (`Html []HtmlBlock`, `Estructura`, `Proyectos`, `Metadatos` con tipos/organismos/titulo/fuente/materias/derogado/fechas/vigencia/vinculaciones); convertir cada bloque HTML a Markdown con `html-to-markdown/v2` (converter base+commonmark+table, regla `div.p` → párrafo) y pasar por el sanitizer; HTTP 500 con id inexistente → error `ErrNormaNotFound`
- [x] 3.6 Escribir `internal/bcn/law_client_test.go` (suite + `httptest.Server` local): search devuelve resultados parseados (fixture JSON real en `testdata/search_response.json`), resumen limpiado, paginación con total, retry ante 5xx transitorios (contador de intentos), breaker abre tras N fallos (consecutivos) y rechaza sin red; get_law parsea el fixture real (`testdata/norma_full.json`), convierte el contenido a Markdown (entidades decodificadas, enlaces conservados), 500 → ErrNormaNotFound
- [x] 3.7 Escribir `internal/bcn/sanitize_test.go` (suite, table-driven): casos con `&nbsp;`/`&ensp;`/`&emsp;` → espacio, XML embebido del RESUMEN con indentación, entidades `&#241;`/`&#xDA;`, control chars/zero-width eliminados, comillas `&quot;` conservadas, enlaces conservados, whitespace colgado recortado
- [x] 3.8 Escribir `internal/bcn/sanitize_bench_test.go`: benchmark sobre la fixture real (`testdata/norma_full.json` ampliada a ~100KB) con guard rails (assert: <2 ms/op y <15 allocs/op) para custodiar la optimización de pasada única

## 4. Tools MCP

- [x] 4.1b Implementar output tipado `SearchLawsOutput` (query, page, page_size, total_items, total_pages, results[] con norm_id/type/title/published/organism/summary completos) y cambiar el handler a `mcp.ToolHandlerFor[SearchLawsArgs, SearchLawsOutput]`: texto como vista del output, structuredContent con los datos completos (resúmenes sin truncar)
- [x] 4.2b Implementar output tipado `GetLawOutput` (metadatos, estructura, proyectos, content con `omitempty`) — `TipoNumero` ya lleva los 5 campos completos de la API (incluido `tipo`, agregado tras revisión del usuario) y cambiar el handler a `mcp.ToolHandlerFor[GetLawArgs, GetLawOutput]`: structuredContent con `content` omitido cuando `StructureOnly`
- [x] 4.3 Actualizar `RegisterTools` en `internal/tools/tools.go` para recibir `client bcn.LawClient` y registrar las dos tools nuevas (manteniendo `echo`)
- [x] 4.5 Extender `search_laws_test.go` y `get_law_test.go`: verificar `res.StructuredContent` tipado (datos completos en search, `content` omitido con `StructureOnly` en get_law) y que el texto sigue presente como `TextContent`

## 5. Bootstrap y verificación

- [x] 5.1 Actualizar `cmd/chile-bcn-mcp/main.go`: leer `API_RESOURCES` (default `config/api.resources.yaml`), `config.Load`, `bcn.NewLawClient`, pasar el cliente a `RegisterTools`; log del número de recursos cargados
- [x] 5.1b Quitar la variable `API_RESOURCES` de `cmd/chile-bcn-mcp/main.go`: ruta fija `config/api.resources.yaml` (sin override por env, decidido con el usuario); mantener `config.Load` + `bcn.NewLawClient` + `RegisterTools` y el log de recursos cargados
- [x] 5.2 Generar el mock: `go tool mockery` (o `make mock` si se agrega target) → verificar `law_client_mock.go` en `internal/bcn/`
- [x] 5.3 `go build ./...`, `go vet ./...` y `go test ./...` en verde (sin acceso a red en tests)
- [x] 5.4 Prueba manual con la API real (solo smoke, no en CI): `curl` del server con `tools/call search_laws` (query "Ley 21.827") y `tools/call get_law` (norm_id 1226950, con y sin `structure_only`) contra `localhost:8000/mcp`
- [x] 5.5 Verificar que el server no arranca con un `api.resources.yaml` inválido (fail-fast) y que arranca con el válido

## 6. Caché ETag

- [x] 6.1 Implementar `internal/bcn/norma_cache.go`: caché en memoria (`sync.Mutex` + `map[int64]cacheEntry{etag string, norma NormaFull}`), tope de 100 entradas con evicción arbitraria, métodos `get(normID)` y `put(normID, etag, norma)`
- [x] 6.2 Integrar el caché en `GetNorma` (`law_client.go`): hit → `SetHeader("If-None-Match", etag)`; respuesta `304` → servir la entrada cacheada sin re-descargar ni re-convertir; `200` → reemplazar entrada con el nuevo ETag
- [x] 6.3 Escribir `internal/bcn/norma_cache_test.go` (suite): put/get round-trip, update reemplaza etag y contenido, cap evicta; extender `law_client_test.go`: 304 → segundo resultado idéntico desde caché (server recibe 2 requests, una sola conversión), 200 con ETag nuevo → reemplaza

## 7. Documentación y verificación final

- [x] 7.1 Actualizar `README.md`: sección de tools (`search_laws`, `get_law` con argumentos y ejemplos de respuesta), mención del caché ETag de `get_law`, y la sección Docker indicando que `config/` va embebida en la imagen (sin hot-reload — los cambios se despliegan con rebuild)
- [x] 7.2 Verificar el Dockerfile: `COPY config/ /app/config/` + `WORKDIR /app` presentes (la imagen arranca con la ruta fija; ya validado con podman — tarea de confirmación)
- [x] 7.3 `make check` completo en verde tras todos los cambios

## 8. Fix de contenido anidado (bug encontrado en smoke con la Ley 21.600)

- [x] 8.1 Modelar el anidamiento: `HtmlBlock.H []HtmlBlock` y `EstructuraPart.H []EstructuraPart` (recursivos) — la API anida artículos bajo títulos/párrafos en el campo `h`
- [x] 8.2 `ConvertContent` recorre el árbol completo (names index del árbol de estructura) — sin esto las leyes largas pierden los artículos (21.600: 7KB vs 205KB)
- [x] 8.3 Render jerárquico en `get_law`: headings por profundidad (`###`/`####`...), bloques con hijos muestran solo el heading (el body duplica el título), TOC como lista anidada
- [x] 8.4 Output tipado sin ciclos: `StructurePartOut` plano con `depth` (un tipo recursivo en el output hace ciclar el generador de JSON Schema del go-sdk)
- [x] 8.5 Sanitizer: colapsar runs de 3+ newlines (los `<div class="p">` vacíos de BCN) a máximo 2
- [x] 8.6 Fixture actualizada con la respuesta anidada real + tests: `TestGetNormaParsesNestedArticles` (TÍTULO I → 3 artículos, Párrafo → Artículo 4°), test de newlines colapsados, smoke real 21.600 → 205K chars con artículos
