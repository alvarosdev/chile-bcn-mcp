## Purpose

Conteo agregado cross-tipo de jurisprudencia de Contraloría por query, para exploración previa sin traer documentos, expuesto como tool MCP LLM-first.

## Requirements

### Requirement: Conteo agregado por query

El servidor DEBE exponer una tool `count_cgr_jurisprudencia` que consulta `POST /apibusca/count/dictamenes` con `query` (texto a buscar, requerido, vacío permitido) y `exact_search` (bool, por defecto false) y devuelve `total` (suma de todos los tipos, `hits.total.value`) y `buckets` (`aggregations.count_by_type.buckets` con `key` y `doc_count` por tipo: `dictamenes`, `auditoria`, `legislacion`, `web`, `cuentas`, etc.). `hits.hits` del servicio es siempre vacío y DEBE ignorarse.

#### Scenario: Conteo de query con resultados en varios tipos
- **WHEN** un cliente llama a `count_cgr_jurisprudencia` con `query:"quillota"`
- **THEN** la respuesta incluye `total:1255` y `buckets:[{key:"auditoria",doc_count:690},{key:"dictamenes",doc_count:312},{key:"legislacion",doc_count:206},...]`

#### Scenario: Conteo de query vacía
- **WHEN** un cliente llama con `query:""`
- **THEN** la respuesta incluye el conteo global (cap `10000 gte` si aplica) sin error

#### Scenario: Conteo sin resultados
- **WHEN** un cliente llama con `query:"nonexistentxyz123"`
- **THEN** la respuesta incluye `total:0` y `buckets:[]`, sin error de protocolo

#### Scenario: Conteo exacto por ID
- **WHEN** un cliente llama con `query:"E179593N25"` y `exact_search:true`
- **THEN** la respuesta incluye `total:1` y `buckets:[{key:"dictamenes",doc_count:1}]`

#### Scenario: LLM-first dual output
- **WHEN** el conteo tiene éxito
- **THEN** la respuesta incluye texto "quillota: 1255 resultados — auditoria 690, dictamenes 312, ..." y `structuredContent` tipado con los mismos campos

### Requirement: Transporte declarado

El conteo DEBE usar el resource `cgr_count` declarado en `api.resources.yaml` (url `https://www.contraloria.cl`, path `/apibusca/count/dictamenes`, method `POST`, timeout 10s, retry 2 intentos, breaker 5/2/30s) con los mismos headers que `cgr_search`. Ante query vacía o faltante DEBE decidir entre error de argumentos o búsqueda global según el requisito de `cgr-search` (consistencia).

#### Scenario: Falla transitoria en conteo
- **WHEN** el servicio responde 5xx en el primer intento
- **THEN** el cliente reintenta según `cgr_count` y retorna éxito si el siguiente intento responde 200

