## 1. Operaciones de árbol en internal/bcn

- [x] 1.1 Agregar `NormaFull.SectionContent(sectionID int64) ([]HtmlBlock, bool)` que recorre `Html` en profundidad y devuelve el subárbol cuyo bloque raíz tiene `I == sectionID` (el `bool` indica si se encontró).
- [x] 1.2 Agregar `NormaFull.CountArticles() int` que cuenta las entradas de `Estructura` con `T == 6` (artículo).
- [x] 1.3 Agregar `NormaFull.ContentCharCount() int` que replica el paseo de `renderBlocks` (headings + `Markdown` de hojas + recursión) sumando runas sin construir el string.
- [x] 1.4 Mover `StructurePartOut` a `bcn` (o alias en tools) para que `projectSummary` pueda incluir la estructura aplanada.

## 2. Tools get_law y get_law_summary

- [x] 2.1 En `GetLawArgs` agregar `SectionID int64` con tag `json:"section_id,omitempty"` y `jsonschema` descriptivo.
- [x] 2.2 En `GetLawOutput` y `bcn.NormaSummary` agregar `CharCount int` (`json:"char_count"`) y `ArticleCount int` (`json:"article_count"`) — sin `omitempty`.
- [x] 2.3 En `makeGetLaw` validar `section_id`: si es > 0 y `SectionContent` no lo encuentra, devolver error de argumentos sin consultar el servicio, sugiriendo `structure_only=true` para listar los `section_id` válidos.
- [x] 2.4 En `buildGetLawOutput` recortar el contenido con `SectionContent` cuando `section_id` > 0, y poblar `CharCount`/`ArticleCount`.
- [x] 2.5 En `formatNorma` renderizar el subárbol recortado (no la norma completa) cuando `section_id` está presente.
- [x] 2.6 En `makeGetLawSummary` poblar `CharCount`/`ArticleCount` y la estructura aplanada (`projectSummary` incluye `Estructura`), derivados del `NormaFull` (cacheado o recién obtenido).
- [x] 2.7 Reescribir las descripciones de `get_law` y `get_law_summary` para documentar el flujo barato (resumen → `structure_only` → `section_id`) y advertir sobre normas largas.
- [x] 2.8 Mostrar el tamaño en el texto: `formatNorma` con "Size: <chars> chars · <articles> articles" en el encabezado y "Section: <nombre>" cuando hay `section_id`; `formatNormaSummary` con el tamaño y la estructura ("## Structure").
- [x] 2.9 Eco de `section_id` en `GetLawOutput` (simetría con `VersionDate`).

## 3. Prompt law_research_workflow

- [x] 3.1 Agregar el builder puro del prompt `law_research_workflow` con argumentos `norm_id` (requerido) y `question` (opcional), que fija el orden `get_law_summary` (resumen con estructura) → `get_law` (`section_id`) y la regla de no inventar contenido.
- [x] 3.2 Registrar el prompt en `RegisterPrompts` y actualizar los tests de `prompts/list` (siete prompts).

## 4. Tests

- [x] 4.1 Tests unitarios de `SectionContent`, `CountArticles` y `ContentCharCount` contra un fixture real (`norma_full.json` o `norma_2010.json`).
- [x] 4.2 Tests de `makeGetLaw`: `section_id` válido devuelve el subárbol; `section_id` inexistente devuelve error; `char_count`/`article_count` presentes; interacción `section_id` + `structure_only`.
- [x] 4.3 Tests de `makeGetLawSummary`: `char_count`/`article_count` presentes en el output.
- [x] 4.4 Tests de prompts: `law_research_workflow` aparece en `prompts/list` y su template referencia el flujo correcto.
- [x] 4.5 Evicción LRU: llenar el caché por sobre el cap y verificar que se evicta la menos usada recientemente y sobrevive la recién accedida.
- [x] 4.6 Tope de respuesta: httptest con body mayor al tope → error sin resultado; el fixture normal sigue pasando.
- [x] 4.7 Pool de converters: N goroutines convirtiendo normas distintas → markdown correcto y sin deadlock (correr con `-race`).
- [x] 4.8 Actualizar aserciones existentes de outputs por los campos nuevos (conteos en structured de `get_law`/`get_law_summary`, eco de `section_id`).
- [x] 4.9 Texto: "Size:" presente en `get_law` y `get_law_summary`; "Section:" presente con `section_id`; el error de sección inexistente menciona `structure_only`.
- [x] 4.10 Actualizar `get_law_summary_test.go` (permite `## Structure`, sigue prohibiendo `## Content`) y `smoke.sh` (misma regla).
- [x] 4.11 Conteos: con `section_id` corresponden al slice; sin `section_id`, al total; `get_law_summary` siempre reporta el total.

## 5. Verificación

- [x] 5.1 `make check` (build + vet + test) verde.
- [x] 5.2 `make fmt-check` limpio.
- [x] 5.3 Smoke manual opcional: `get_law(norm_id=1195666, structure_only=true)` → tomar un `section_id` de la estructura → `get_law(norm_id=1195666, section_id=...)` y verificar que devuelve solo esa sección.
- [x] 5.4 `go test ./... -race -count=1` verde (cubre singleflight, pool y LRU).

## 6. Robustez del cliente BCN

- [x] 6.1 Pool de converters (`sync.Pool`) en `Client`; `ConvertContent` adquiere y libera.
- [x] 6.2 Singleflight keyed por `normaCacheKey` alrededor del fetch de `GetNorma`; promover `golang.org/x/sync` a dependencia directa.
- [x] 6.3 Evicción LRU (`container/list`) en `etagCache`.
- [x] 6.4 Tope de respuesta upstream (5MB): `SetResponseDoNotParse(true)` + `io.LimitReader` (límite + 1 byte centinela) + `json.Unmarshal` manual en `GetNorma`; error claro si excede.
- [x] 6.5 Test de concurrencia: N `GetNorma` concurrentes de la misma clave → una sola llamada real y el resto servidas con el mismo contenido.

## 7. Endurecimiento

- [x] 7.1 `docker-compose.yml`: `read_only: true`, `tmpfs` en `/tmp`, `cap_drop: [ALL]`, `security_opt: ["no-new-privileges:true"]`.
- [x] 7.2 `make podman-run`: `--read-only --tmpfs /tmp --cap-drop ALL --security-opt no-new-privileges`.
- [x] 7.3 Verificar arranque con rootfs read-only (`make podman-run` + smoke); si podman-compose no soporta alguna clave, dejarla como flags documentados en el Makefile.
- [x] 7.4 `ReadHeaderTimeout` en el `http.Server` (main.go).

## 8. Cadena de suministro y CI

- [x] 8.1 Paso `govulncheck` en `ci.yml` con versión pineada (no `@latest`; falla el job ante vulnerabilidades) y target `vuln` en el Makefile.
- [x] 8.2 `.github/dependabot.yml` para Go modules.
- [x] 8.3 Correr los tests de `ci.yml` con `-race` (este change introduce código concurrente).

## 9. Fuzzing

- [x] 9.1 Targets `FuzzSanitizeMarkdown` y `FuzzConvertContent` con corpus semilla desde fixtures reales.
- [x] 9.2 Correr fuzz local (30–60s por target); corregir y testear cualquier panic encontrado.
