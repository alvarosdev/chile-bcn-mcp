## Purpose

Búsqueda paginada de dictámenes de la Contraloría General de la República vía `contraloria.cl/apibusca`, con paginación fija y ordenamiento, expuesta como tool MCP LLM-first con contenido estructurado y texto derivado.

## Requirements

### Requirement: Búsqueda paginada de dictámenes

El servidor DEBE exponer una tool `search_cgr_dictamenes` que consulta `POST /apibusca/search/dictamenes` con `query` (texto a buscar, requerido, vacío permitido para listar recientes), `exact_search` (bool, por defecto false), `order` (enum `date` DESC por defecto, `dateasc` ASC, `score` relevancia DESC) y `page` (número de página 1-indexed para el cliente, por defecto 1). La tool DEBE mapear `page` a `page` 0-indexed del servicio, fijar `source:"dictamenes"`, `date_name:"fecha_documento"` y `options:[]`, y devolver los resultados de la página solicitada junto con `pagination{total, page, page_size:20, total_pages, has_more}`. Cada resultado DEBE incluir `dictamen_id` (_id/doc_id), `n_dictamen`, `fecha_documento`, `materia` y `descriptores` como mínimo, con `materia` truncada en el texto pero completa en el structured, y las URLs `url` (`/buscadorpdf/dictamenes/{id}/html`) y `pdf_url` (`/buscadorpdf/dictamenes/{id}/pdf`) para citación y descarga.
#### Scenario: Búsqueda por texto con paginación
- **WHEN** un cliente llama a `search_cgr_dictamenes` con `query:"quillota"` y `page:1`
- **THEN** la respuesta incluye hasta 20 dictámenes de la primera página, `total:312`, `page:1`, `page_size:20`, `total_pages:16`, `has_more:true`

#### Scenario: Segunda página sin duplicados
- **WHEN** un cliente llama con `query:"quillota"`, `page:2` (mapeado a `page:1` del servicio)
- **THEN** la respuesta incluye los siguientes 20 dictámenes sin duplicar los de la página 1, con `has_more` correcto

#### Scenario: Ordenamiento por fecha ascendente
- **WHEN** un cliente llama con `order:"dateasc"`
- **THEN** los resultados vienen ordenados por `fecha_documento` ASC (más antiguo primero), con el dictamen de 1966 antes que el de 2026

#### Scenario: Ordenamiento por relevancia
- **WHEN** un cliente llama con `order:"score"`
- **THEN** los resultados vienen ordenados por `_score` DESC de Elasticsearch

#### Scenario: Búsqueda exacta
- **WHEN** un cliente llama con `exact_search:true` y `query:"E179593N25"`
- **THEN** la respuesta contiene el dictamen con ese ID si existe, respetando el flag enviado al servicio

#### Scenario: Página más allá del total
- **WHEN** un cliente pide `page:99` con `query:"quillota"` (total 312, 16 páginas)
- **THEN** la respuesta incluye `results:[]`, `total:312`, `has_more:false`, sin error de protocolo

#### Scenario: Query vacía
- **WHEN** un cliente llama con `query:""`
- **THEN** la tool consulta el servicio y retorna los dictámenes más recientes con `total` capped (`relation:"gte"` a 10k), sin error de argumentos

#### Scenario: LLM-first dual output
- **WHEN** la búsqueda tiene éxito
- **THEN** la respuesta incluye `Content` de texto con lista numerada (materia, fecha, dictamen_id para drill-down, y URLs `url`/`pdf_url` para citación) y `structuredContent` tipado con los mismos datos completos

### Requirement: Paginación fija de 20 y metadatos derivados

La tool DEBE tratar `page_size` como fijo 20 (el servicio ignora cualquier `size` enviado). `total_pages` DEBE ser `ceil(total/20)` y `has_more` DEBE ser `page*20 < total`. `total` DEBE leerse de `hits.total.value` y cuando `relation:"gte"` y `value==10000` el texto DEBE advertir "más de 10.000 resultados, refina tu búsqueda".

#### Scenario: Cálculo de total_pages en última página parcial
- **WHEN** `total:312` (16 páginas, última con 12)
- **THEN** `total_pages:16`, `page:16` retorna 12 resultados y `has_more:false`

#### Scenario: Aviso de cap 10k
- **WHEN** `query:""` retorna `total:{value:10000, relation:"gte"}`
- **THEN** el texto incluye aviso de cap y el structured mantiene `total:10000`

### Requirement: Endpoints declarados en YAML y transporte resiliente

La búsqueda DEBE usar el resource `cgr_search` declarado en `internal/config/api.resources.yaml` (url `https://www.contraloria.cl`, path `/apibusca/search/dictamenes`, method `POST`, timeout 10s, retry 3 intentos, breaker 5/2/30s) con headers `Accept: application/json`, `Content-Type: application/json` y `Origin: https://www.contraloria.cl`. Ante fallos transitorios (5xx, timeout, status 0) DEBE reintentar según el resource; ante breaker abierto DEBE fallar rápido.

#### Scenario: Falla transitoria con retry
- **WHEN** el servicio responde 5xx en el primer intento
- **THEN** el cliente reintenta y retorna éxito si el siguiente intento responde 200

#### Scenario: Validación de args
- **WHEN** `order` es distinto de `date|dateasc|score` o `page <=0`
- **THEN** la tool devuelve error de argumentos sin consultar el servicio

