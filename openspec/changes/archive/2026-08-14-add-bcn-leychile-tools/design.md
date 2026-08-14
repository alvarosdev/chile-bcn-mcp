## Context

El scaffold (`scaffold-mcp-server`) ya está archivado: server MCP con transporte stdio+HTTP, auth, health, y la tool demo `echo` en `internal/tools/tools.go`. El dominio real son los servicios de LeyChile/BCN: `buscarjson` (JSON paginado con un response heterogéneo de 3 elementos: resultados, paginación, facets) y `get_norma_json` (contenido estructurado de una norma: `html[]` con bloques HTML, `estructura[]` TOC, `proyectos[]` y `metadatos` ricos). Verificado con la API real: ambos responden 200 con headers mínimos (sin headers de browser) y `get_norma_json` funciona con solo `idNorma`; un id inexistente responde HTTP 500 con mensaje en el body; detrás hay CloudFront. Ver proposal.md para la motivación; requisitos observables en specs/leychile-search.

## Goals / Non-Goals

**Goals:**
- Endpoints 100% declarados en `api.resources.yaml` (url, path, method, timeout, retry, circuit breaker **por endpoint**), cargados una vez al arranque.
- Transporte resiliente configurable por recurso sin librería de breaker externa (breaker built-in de resty).
- Cero llamadas a servicios externos durante los tests (mocks + httptest).
- Estilo senior Go 1.26: instancia única inyectada (sin estado global), filenames snake_case, tests con suite testify, mocks mockery al lado del archivo productivo.

**Non-Goals:**
- `leyes_por_tema` y los tags de categorías (fuera de scope, declarado por el usuario).
- El HTML crudo de `Navegar?idNorma` (reemplazado por `get_norma` estructurado; el recurso no se declara en el YAML).
- Paginación automática multi-página (el LLM navega con `npagina`).
- Caché de resultados de búsqueda (la búsqueda cambia seguido y tiene claves de texto libre; el caché ETag aplica solo a `get_law`).

## Decisions

**1. `api.resources.yaml` plano, por recurso, versionado**
Contrato con `version: 1` y un mapa `resources:` donde cada recurso es autosuficiente: `url` (fqdn), `path`, `method`, `timeout`, `retry` (`attempts`, `backoff`, `max_backoff`) y `circuit_breaker` count-based (`failure_threshold`, `success_threshold`, `reset_timeout` — el modelo del breaker de resty). La clave del mapa es el id usado por el código (`client.Call("search_laws", ...)`). Recursos: `search_laws` (`/servicios/buscarjson`), `get_law` (`/servicios/Navegar/get_norma_json`) y `navegar_norma` — este último queda **fuera del cambio**: `get_law` reemplaza al HTML crudo de `Navegar` (decisión del usuario; mismo input `norm_id`, salida estructurada superior). Se descartó el anidamiento tipo docker-compose (`services:` → `endpoints:`) por pedido explícito: formato plano, explícito, "senior".
*Alternativa descartada*: schema declarativo completo de query params en YAML — el YAML solo declara el contrato de transporte; los args quedan tipados en Go (structs `json`/`jsonschema`).

**2. resty v3 como cliente HTTP — un `*resty.Client` por recurso**
**Nota de versión (decidida en apply)**: resty v3 solo existe como release candidate (`v3.0.0-rc.3`) y su module path migró a `resty.dev/v3` (no `github.com/go-resty/resty/v3`). Decisión del usuario: usar `resty.dev/v3` rc.3 — es lo que el design aprobó (breaker built-in + retry per-request); ante una release final se actualiza el import y se revisa la migración.
`go-resty/resty/v3` porque: retry **per-request** (`SetRetryCount`/`SetRetryWaitTime`/`SetRetryMaxWaitTime` en `Request`) que mapea 1:1 al YAML por endpoint; circuit breaker **built-in** (`NewCircuitBreakerCount` + `SetCircuitBreaker`, count-based, client-level — por eso un client por recurso, 2-3 clients livianos); `SetResult` para unmarshal automático del JSON; hooks (`OnBeforeRequest`/`OnAfterResponse`) para logging con `slog`. Debajo sigue siendo `http.Client`: `httptest.Server` y `SetTransport` funcionan igual para los tests.
*Alternativa descartada*: `net/http` + retry manual (~30 líneas) + `sony/gobreaker` — válida, pero el usuario eligió resty tras la comparación (retry declarativo que mapea al YAML, breaker sin dep extra).

**3. Singleton = instancia única inyectada, no estado global**
`main` → `config.Load("config/api.resources.yaml")` → `bcn.NewLawClient(resources)` → `RegisterTools(srv, lawClient)`. Una sola instancia creada en el bootstrap, reutilizada en cada request; los handlers reciben la dependencia por closure (patrón `make*` de godot-mcp-docs). Se descartó `var global + sync.Once` (estado global mutable: tests frágiles, inicialización implícita).
**Ruta fija (decisión del usuario, revisada)**: el YAML se carga SIEMPRE desde `config/api.resources.yaml`, sin override por variable de entorno. Racional: el contrato no es secreto, y un cambio de configuración se despliega rebuildando la imagen (el Dockerfile copia `config/` — verificado con podman), no con hot-reload — se descarta explícitamente cualquier mecanismo de hot swap. Cargado y validado al arranque.

**4. Interfaz `LawClient` para mockery, mocks al lado del archivo**
`internal/bcn/law_client.go` define `LawClient` con `Search(ctx, SearchParams) (SearchResponse, error)` y `GetNorma(ctx, idNorma int64) (NormaFull, error)`. Mock generado con mockery v3 (`template: testify`, `dir: "{{.InterfaceDir}}"`, `filename: "{{.InterfaceName | snakecase}}_mock.go"`) → `law_client_mock.go` en el mismo paquete, siguiendo `.mockery.yml` de godot-mcp-docs. Las tools dependen de la interfaz; el cliente real la implementa.

**5. Modelo de los responses de LeyChile**
`buscarjson` (`[resultados, paginación, facets]`) se deserializa con `UnmarshalJSON` custom: `[3]json.RawMessage` → structs tipados `Norma`, `Pagination` (con `totalitems`), `Facets`. El campo `RESUMEN` de cada `Norma` se limpia al devolverlo: `html.UnescapeString` + strip de tags XML (decisión del usuario: limpio para el LLM).
`get_norma_json` se modela con `NormaFull`: `Html []HtmlBlock` (`t` HTML crudo, `i` id de parte, `v` etiquetas), `Estructura []EstructuraPart` (TOC: `n` nombre, `i` id), `Proyectos []Proyecto` (boletín, enlace), `Metadatos` (tipos_norma, organismos, titulo, fuente, materias, derogado bool, fechas, vigencia, vinculaciones, resumenes — este último ya viene limpio en la API) y los campos secundarios (`jurisprudencia`, `doctrina`, `alertas`) solo si no están vacíos. El HTTP 500 con mensaje en el body para id inexistente se traduce a un error `ErrNormaNotFound` reutilizable.

**Bug encontrado en smoke (leyes largas)**: `HtmlBlock` y `EstructuraPart` son RECURSIVOS — la API anida los artículos bajo títulos/párrafos en el campo `h` (p.ej. Ley 21.600: TÍTULO I → Artículo 1°..3°; TÍTULO II → Párrafo → Artículo). El parseo plano inicial perdía TODO el contenido de las leyes largas (21.600 salía con solo títulos, 7KB vs 205KB reales). Fix: campos `H` recursivos, `ConvertContent` y el render caminan el árbol (headings por profundidad, los bloques-título muestran solo su heading — su body repite el título), el TOC se renderiza como lista anidada, y el output tipado de `get_law` aplana la estructura con `depth` (un tipo recursivo hace ciclar al generador de JSON Schema del go-sdk). El sanitizer además colapsa runs de 3+ newlines (los `div.p` vacíos de BCN).

**10. Interfaz en inglés (convención del proyecto)**
Todo lo que el LLM ve y todo el código fuente está en **inglés**: nombres de tools MCP (`search_laws`, `get_law`), argumentos (`query`, `page`, `page_size`, `norm_id`, `structure_only`), filenames (`search_laws.go`, `get_law.go`), descripciones de tools y comentarios de código. Los **datos crudos del dominio se mantienen tal cual** vienen de la API: el texto legal de las normas, el campo `RESUMEN`, los títulos y materias (español, contenido del servicio) y los nombres de parámetros de la API (`idNorma`). La traducción termina en la interfaz: `norm_id` (tool) → `idNorma` (query param del servicio). El server instructions de `server.New` ya está en inglés.

**8. Conversión del contenido a Markdown con `html-to-markdown/v2`**
Cada bloque `HtmlBlock.t` se convierte con `converter.NewConverter` (plugins `base` + `commonmark` + `table`) y reglas custom: `div.p` → párrafo (`RendererFor` TagTypeBlock), para que los artículos de BCN salgan como texto legible; las entidades HTML se decodifican automáticamente y los enlaces a otras normas se conservan. La conversión vive en el cliente (los tests de la tool reciben el mock ya convertido); se testea con fixtures reales de normas variadas (tablas, anexos).
*Alternativa descartada*: walker manual con `x/net/html` (~100-150 líneas) — cero deps pero se rompe con la variedad de markup de las normas (tablas, listas, anexos); `x/net` tampoco está en el grafo de dependencias actual.
*Alternativa descartada*: `jaytaylor/html2text` — texto plano, no Markdown.
**Concurrencia (decidido)**: el `Converter` de v2 es thread-safe (mutex interno, confirmado en la doc oficial) y se comparte como instancia única del proceso. El mutex serializa conversiones concurrentes en milisegundos — imperceptible frente a la latencia del cliente; se acepta el mutex tal cual. Se descarta `sync.Pool` (paralelismo real) hasta que un benchmark bajo carga real mida contención: decisión diferible, punto de cambio identificado (5 líneas en el cliente).

**12. Caché de normas con revalidación ETag (solo `get_law`)**
Verificado contra la API real: `get_norma_json` responde `ETag: W/"..."` y ante un `If-None-Match` coincidente responde `304` con 0 bytes (además envía `cache-control: max-age=600`). Diseño: `internal/bcn/norma_cache.go` — `sync.Mutex` + `map[int64]cacheEntry{etag string, norma NormaFull}`, con tope de 100 entradas y evicción arbitraria (iteración de mapa) para acotar memoria; vive en el proceso (se pierde al reiniciar — aceptado en el spec). Flujo de `GetNorma`: hit → `SetHeader("If-None-Match", etag)`; `304` → servir la entrada cacheada (sin re-descargar ni re-convertir — el punto caro es el parseo del DOM); `200` → reemplazar entrada. La búsqueda NO se cachea (claves de texto libre, baja tasa de re-consulta, resultados que cambian seguido). Correctitud por revalidación, sin TTL — el ETag invalida solo ante cambios reales. Riesgo de implementación: el manejo del 304 con `SetResult` en resty se valida con tests (el body viene vacío).

**11. Norma del proyecto: LLM-first con structuredContent opcional**
Toda tool devuelve **ambos**: `content[]` con `TextContent` formateado para lectura del modelo (el canal primario, renderizado por cualquier cliente) y `structuredContent` tipado como segundo valor del handler (`mcp.ToolHandlerFor[Args, Output]`) — el go-sdk genera el `outputSchema` automáticamente del tipo de Output. Reglas de la norma: (a) el texto es una **vista** del structured, ambos derivan del mismo struct de respuesta (sin drift, sin copias manuales); (b) jamás JSON embebido como string dentro del texto (anti-patrón: duplica tokens, para eso existe structuredContent); (c) el texto puede resumir/truncar para el modelo mientras el structured lleva los datos completos (caso: resúmenes truncados a 600 chars en texto vs. completos en structured); (d) outputs con campos opcionales llevan `omitempty` (caso: `get_law` con `structure_only` omite el contenido). La tool demo `echo` queda como excepción intencional (sin structured, `struct{}{}`).

**9. Pipeline de normalización de texto (sanitizer central, sin magic strings, una pasada)**
Todo texto que llegue al LLM (RESUMEN de la búsqueda y contenido de la norma) pasa por un sanitizer único en `internal/bcn/sanitize.go`, aplicado en este orden: (1) decodificar entidades (`html.UnescapeString`), (2) conversión a Markdown para el contenido (o strip de tags XML para el RESUMEN con `resumenTagRe`), (3) espacios no separadores → espacio normal, (4) eliminar caracteres de control y de ancho cero, (5) colapso de espacios consecutivos y trim (conservando saltos de línea). **No se toca**: comillas `&quot;` → `"` (contenido legal citado) ni enlaces (referencia a otras normas).
Los pasos (3)-(5) se implementan como **una sola pasada** (state machine sobre runas + `strings.Builder` con `Grow`), no como pasos encadenados: validado con benchmark sobre HTML real de BCN (~108KB): pasada única = 0.63 ms/op y 8 allocs vs. multi-pasada = 1.19 ms/op y 47 allocs (2x tiempo, 6x alocaciones; ambos correctos). El benchmark queda en el repo (`sanitize_bench_test.go`) con guard rails (<2 ms por 100KB, <15 allocs/op) para custodiar la optimización. Nota: el bottleneck real del pipeline es `html-to-markdown` (parseo DOM); el sanitizer es barato al lado, no se justifica micro-optimizar más.
Toda la "basura" está **definida en `internal/bcn/garbage.go`**, no como literales en la lógica: constantes runa nombradas (`nbspRune = ' '`, `enspRune = ' '`, `emspRune = ' '`, `zeroWidthSpaceRune = '​'`, `bomRune = '﻿'` — siempre como escapes, un literal BOM no compila en Go), límites de control nombrados (`controlMin/controlMax = ' '/''`), listas declarativas que el sanitizer itera (`spacesToNormalize`, `zeroWidthChars`) y la expresión del wrapper XML del RESUMEN como `resumenTagRe` nombrada. El inventario completo de basura (XML embebido, sangrías `&nbsp;`, whitespace del wrapper, control chars) queda documentado en ese archivo con un comentario por entrada y su origen en la API de BCN.

**6. Tests: suite testify + httptest + mocks, cero red real**
- Tools: `suite.Suite` con `SetupTest()` creando `MockLawClient` fresco (expecter API) — sin red.
- Cliente real: `httptest.Server` local con fixtures JSON del response real en `testdata/`; los tests del retry simulan fallos transitorios (5xx) y del breaker, fallos consecutivos.
- Filenames snake_case (`law_client_test.go`, `search_laws.go`, `get_law.go`); identifiers en CamelCase (convención Go).
- **Convenciones descubiertas en la implementación** (capturadas para futuros tools): (a) en las expectativas de `MockLawClient` usadas por handlers MCP, matchear el `ctx` con `mock.Anything`, nunca por identidad — el go-sdk pasa un contexto derivado al handler, y un mismatch hace que testify ejecute `FailNow`/`Goexit` en la goroutine del handler (que no es la del test), matándola sin responder y colgando `CallTool` para siempre; (b) los args de tools con defaults llevan `,omitempty` en el tag `json` (un `int` sin omitempty hace que el schema generado los marque `required`).

**7. Retry solo para GET idempotentes**
`buscar_normas` y `get_norma` son GET — el retry es seguro. Los hooks de resty loguean cada intento con `slog` (debug) y el fallo final (error).

## Risks / Trade-offs

- [resty v3 es una dependencia grande con API fluida propia] → Mitigación: uso mínimo y deliberado (retry per-request, breaker, SetResult); toda la complejidad queda encapsulada en `internal/bcn`.
- [html-to-markdown v2 trae goquery/x/net transitivas] → Mitigación: conversión aislada en un único sitio (`law_client.go`) y probada con fixtures; la dep es el estándar de la comunidad para esto.
- [El markup de las normas varía (tablas, anexos, listas) y puede romper la conversión] → Mitigación: reglas custom (`div.p` → párrafo) + plugin de tablas; fixtures de normas variadas en los tests; ante una variante nueva se agrega una regla, no un rework.
- [Breaker count-based de resty es menos refinado que ratio/window] → Aceptado: para 2 endpoints es suficiente y predecible; migrar a gobreaker es local a `internal/bcn` si algún día se necesita.
- [La API de BCN puede cambiar el shape del JSON o exigir headers] → Mitigación: fixtures de testdata replican el response real; el parseo vive en un solo lugar (`law_client.go`).
- [El campo `RESUMEN` y el contenido pueden traer variantes del marcado (entidades, `&nbsp;`, XML, control chars)] → Mitigación: sanitizer central documentado (decisión 9) que ignora errores y devuelve texto decodificado; tests con fixtures de cada caso de basura (incluido `&nbsp;` real de BCN).
- [La normalización puede dañar contenido legítimo (comillas, enlaces, párrafos)] → Mitigación: lista explícita de lo que NO se toca en la decisión 9 + tests que la verifican (citas con `&quot;`, enlaces a `idNorma`).
- [`mockery` v3 puede no estar instalado en el entorno de CI] → Mitigación: se declara como tool de Go en `go.mod` (`go tool mockery`), mocks generados y commiteados.

## Migration Plan

Cambios aditivos sobre el scaffold: `RegisterTools` gana dos tools; `main.go` gana el bootstrap del cliente. La tool `echo` se conserva. No hay migración de datos ni rollback especial (borrar archivos revierte).

## Open Questions

<!-- Ninguna: nombre del YAML, configuración por endpoint, cliente resty, singleton inyectado, retry manual vs librería, suite/mockery y scope de tools fueron decididos con el usuario en la exploración. -->
