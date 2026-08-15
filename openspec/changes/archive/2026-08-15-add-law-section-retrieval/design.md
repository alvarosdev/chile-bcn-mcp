## Context

El cliente ya descarga y parsea la norma completa (`NormaFull`) en cada `get_law`: convierte los bloques HTML anidados (`HtmlBlock.H`) a Markdown en `ConvertContent`, y cachea el resultado en `etagCache[NormaFull]` con revalidación ETag. La estructura (`EstructuraPart`) y el contenido (`HtmlBlock`) comparten el mismo espacio de identificadores `i` — `ConvertContent` mapea `part.I` a `block.I` para asignar el `SectionName`. La tool `get_law` ya expone `structure_only` (omite el contenido) y `version_date` (versión histórica). Ver motivación en proposal.md — Why.

El spike contra la API real (documentado en proposal.md — Impact) estableció que `get_norma_json` **ignora** `idParte`: la respuesta con y sin `idParte` es byte-idéntica (280930 bytes). Por lo tanto no existe filtrado en la fuente para este endpoint.

## Goals / Non-Goals

**Goals:**
- Permitir leer una sola sección de una norma (por `i`) sin re-emitir el contenido completo al agente.
- Exponer la magnitud del contenido (`char_count`, `article_count`) para que el agente decida antes de pedir el texto.
- Codificar el flujo barato en las descripciones de tools y en un prompt curado.

**Non-Goals:**
- No se elimina la descarga completa de BCN en el primer `get_law` (se sigue trayendo y cacheando; solo se evita re-emitirla al modelo).
- No se introduce índice vectorial, persistencia en disco ni endpoint nuevo en `api.resources.yaml`.
- `get_law_summary` no acepta `section_id` (no tiene contenido que recortar).

## Decisions

### 1. Recorte local sobre el árbol ya parseado (no filtrado en la fuente)

`section_id` mapea a `HtmlBlock.I` / `EstructuraPart.I` (mismo espacio, verificado por `ConvertContent`). El recorte recorre `norma.Html` en profundidad y devuelve el subárbol cuyo bloque raíz tiene `I == sectionID`. Los metadatos y la estructura se devuelven completos (son baratos y le sirven al agente para encadenar la siguiente sección). En la vista de TEXTO de una sección se omiten el resumen y los proyectos (el structured los lleva completos); el TOC se mantiene para encadenar.

- **Alternativa rechazada — `idParte` en la API**: el spike probó que se ignora (respuesta idéntica). Requeriría reverse-engineering de un endpoint BCN distinto; queda como exploración futura, no bloquea este change.

### 2. `char_count` y `article_count` derivados del documento, no de un render aparte

- `char_count` = cantidad de runas del contenido Markdown que emitiría `renderBlocks` (sin construir el string). Se calcula con un paseo que replica `renderBlocks` (suma de headings + `Markdown` de hojas + recursión), de modo que **funciona también con `structure_only`** — que es justo cuando el agente lo necesita.
- `article_count` = cantidad de entradas de `Estructura` con `T == 6` (artículo; la clasificación `1=título, 4=párrafo, 6=artículo` ya está documentada en `law_client.go`).
- Semántica: describen el **contenido devuelto** — con `section_id`, el subárbol de la sección; sin él, la norma completa. `get_law_summary` reporta siempre el total (magnitud del documento).
- Son `int` **sin** `omitempty` (siempre presentes en éxito); no son opcionales.

### 3. Ubicación: operaciones de árbol en `bcn`, presentación en `tools`

El recorte y los conteos son operaciones sobre el dominio (`NormaFull`) → métodos en `internal/bcn` (p. ej. `NormaFull.SectionContent`, `CountArticles`, `ContentCharCount`). Las tools siguen solo validando argumentos y formateando (principio de capas de CLAUDE.md). `section_id` lleva `omitempty` (arg opcional).

### 4. Validación fail-fast de `section_id`

Si `section_id` no aparece en la estructura, la tool devuelve error de argumentos (consistente con la validación estricta de `version_date`). El error sugiere el camino de recuperación: listar los `section_id` válidos con `structure_only=true`. El agente obtiene el `section_id` de la estructura de la misma respuesta, así que un ID inválido indica un error de uso, no un dato del dominio.

### 5. Descripciones que enseñan el flujo

Las descripciones de `get_law` y `get_law_summary` documentan el camino barato (resumen con estructura → `section_id`) y advierten que las normas largas pueden exceder cientos de miles de caracteres. Es el mecanismo de menor fricción para cambiar el comportamiento del agente sin lógica nueva.

### 6. Séptimo prompt `law_research_workflow`

Prompt puro (builder de template sin red, como los seis existentes) con argumentos `norm_id` (requerido) y `question` (opcional), que fija el orden resumen (con estructura) → sección y la regla de no inventar contenido.

### 6b. El summary incluye la estructura (flujo de dos llamadas)

`get_law_summary` incluye la estructura de la norma — gratis: deriva del `NormaFull` cacheado, y en miss ya pasa por `GetNorma` completo — de modo que el flujo del agente pasa de tres llamadas (resumen → `structure_only` → `section_id`) a dos (resumen → `section_id`), sin subir los tokens totales. La estructura se entrega **APLANADA** (`StructurePartOut` con depth, movido a `bcn`): el tipo recursivo `EstructuraPart` en un Output rompería el schema generator (gotcha 2b del proyecto). `structure_only` se mantiene para refrescar el TOC sin el resumen.

### 7. Pool de converters (paralelizar conversiones)

`sync.Pool` de `*converter.Converter` en `Client`: el mutex interno de la librería serializa todas las conversiones, y una ley de 426K chars bloquea a los demás requests. `ConvertContent` adquiere un converter, convierte y lo libera. Alternativa rechazada: crear un converter por request (construirlo con los plugins en cada llamada es desperdicio; el pool reusa).

### 8. Singleflight por norma@versión

`singleflight.Group` en `Client` alrededor del fetch de `GetNorma`, keyed por `normaCacheKey`: requests concurrentes de la misma norma y versión comparten UNA llamada a BCN (retry y breaker adentro del vuelo). `golang.org/x/sync` ya está en go.mod como indirecta (v0.20.0) — solo se promueve a directa. Trade-off aceptado: si el contexto del vuelo líder se cancela, el vuelo falla para todos los esperantes (mismo comportamiento de timeout que hoy).

### 9. Evicción LRU del caché ETag

`container/list` + map en `etagCache`: cada `get` mueve la entrada al frente y al desbordar se evicta la de atrás (la menos usada recientemente), reemplazando la evicción arbitraria por iteración de map. ~30 líneas; deja el caché predecible ante un futuro cap configurable.

### 10. Tope de tamaño de respuesta upstream

Guard OOM: la norma más grande conocida pesa ~280KB, así que se fija un tope de 5MB. En `GetNorma` se lee el body manualmente con `SetResponseDoNotParse(true)` + `io.LimitReader` (límite + 1 byte centinela) y `json.Unmarshal` — el rechazo ocurre ANTES de tener el body completo en memoria, sin depender de una API de límites de resty. Exceder el tope devuelve error, nunca contenido truncado. (Verificar el nombre exacto de la opción do-not-parse en resty rc.3 durante el apply.)

### 11. Contenedor endurecido

Rootfs read-only + tmpfs en `/tmp` + `cap_drop: ALL` + `no-new-privileges` en `docker-compose.yml` y en `make podman-run`. El servidor no escribe en disco (config embebida, caché en memoria, `GOMEMLIMIT` ya fijado en la imagen), así que read-only es gratis. En apply se verifica que podman-compose respete las claves; si alguna no la soporta, los flags viven documentados en el Makefile.

### 12. govulncheck + dependabot

Paso `govulncheck` en `ci.yml` con versión pineada (no `@latest`, para builds reproducibles) — el job falla ante vulnerabilidades conocidas — y `.github/dependabot.yml` para Go modules. `resty v3.0.0-rc.3` queda pinneado con el upgrade a estable registrado como pendiente; el escaneo cubre el interin.

### 13. Fuzz del sanitizer y converter

Targets `FuzzSanitizeMarkdown` y `FuzzConvertContent` con corpus semilla desde los fixtures reales de `testdata/` (input externo hostil). Se corren localmente durante el apply; si aparece un panic se corrige en este mismo change. Sin job de fuzz en CI (gestión de corpus separada).

### 14. `ReadHeaderTimeout` en el servidor HTTP

Defensa anti-Slowloris de una línea; `ReadTimeout` no cubre la fase de headers.

### Alternativas rechazadas (registro de decisiones)

| Alternativa | Motivo del descarte |
|---|---|
| RAG vectorial | El chunking a ciegas parte definiciones legales; a ~426K chars el vector no aporta sobre el término exacto; añade infra de embeddings. |
| Persistencia en disco (SQLite o carpeta) | Proxy stateless; SQLite exige CGo (`CGO_ENABLED=0` en la matriz de build) o un driver pure-Go pesado; beneficio marginal (~200ms de cold-start); el ETag en memoria ya cubre el working set. |
| `idParte` en la fuente | Verificado ignorado (spike). |

## Risks / Trade-offs

- [Estabilidad de los `i` entre versiones/reindexados de BCN] → `section_id` se valida contra la estructura recién obtenida; un ID desactualizado falla con error en vez de devolver contenido incorrecto. En una misma sesión el agente toma el `i` de la propia respuesta.
- [Posible doble cálculo de `char_count`] → se evita con el paseo que replica `renderBlocks`; no se construye el string para contar.
- [Leyes sin artículos (`T==6`)] → `article_count` puede ser 0; es un conteo honesto y el agente se apoya en `char_count`.
- [La descarga completa de BCN persiste en el primer `get_law`] → costo residual de ser proxy; aceptado (los descartes de arriba no lo eliminan sin otra infraestructura).
- [Converter del pool con estado entre conversiones] → los converters se crean idénticos con `newConverter`; en apply se verifica que la librería no guarde estado entre llamadas (si lo hiciera, se vuelve al converter único o a pool con reset).
- [Rootfs read-only rompe el arranque] → el smoke en contenedor con la configuración endurecida lo detecta (task 7.3).
