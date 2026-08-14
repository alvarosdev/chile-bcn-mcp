## Context

Las tools BCN (`search_laws`, `get_law`, `get_law_summary`) están archivadas y en producción local. Exploración nueva verificada contra la API real: `idVersion=YYYY-MM-DD` selecciona texto histórico (Ley 19.628: 36.7K chars a 2010 vs 41.4K última; ETags difieren por versión); `get_historias_de_ley` funciona con solo `idNorma`, envía ETag con 304, y ante norm_id inexistente responde `[]` (no 500). El catálogo de tipos viene de `/Consulta/getTiposNorma` (cod → abbr/valor, 40 tipos). Ver proposal.md para la motivación; requisitos en specs/leychile-search.

## Goals / Non-Goals

**Goals:**
- Versiones históricas en `get_law` y `get_law_summary` con validación estricta y caché por versión.
- Tool `get_law_history` con caché ETag desde el día 1.
- Catálogo de tipos decodificado como dato ANEXO (append, no replace).

**Non-Goals:**
- Enriquecer `search_laws` con el catálogo (postergado — el usuario lo scopeó a get_norma).
- Comparar versiones entre sí (tool de diff) — futuro.
- Cargar el catálogo desde la API (hardcodeado por decisión del usuario).

## Decisions

**1. `version_date` con validación estricta (decisión del usuario)**
Parse con `time.Parse("2006-01-02", versionDate)` en el handler de la tool: formato inválido → error de argumentos sin tocar la API. La API ignora fechas malas silenciosamente (devuelve la última versión) — sin validación, el LLM creería que leyó la versión pedida. Fechas válidas pero anteriores a la norma devuelven el texto original (comportamiento de la API, documentado, no validado).

**2. `NormaQuery` en la interfaz `LawClient`**
`GetNorma(ctx, NormaQuery{NormID, VersionDate})` y `GetNormaSummary(ctx, NormaQuery)` reemplazan las firmas actuales — un struct de query evita el crecimiento de parámetros y hace explícito el concepto "qué norma y qué versión". Cambio interno: las tools MCP no cambian su contrato externo (el parámetro nuevo es opcional/aditivo). Regenerar mock (`make mock`).

**3. Caché generalizado `etagCache[T]` con clave compuesta**
Refactor del `NormaCache` actual a un genérico `etagCache[T]` (mutex + map + cap 100 con evicción arbitraria, mismo patrón). Clave de normas: string compuesto `fmt.Sprintf("%d@%s", normID, versionDate)` (versionDate vacío = última) — sin esto, una versión histórica recibiría la entrada cacheada de otra versión (ETags difieren por versión, verificado). Clave de historias: `normID`. Dos instancias en el Client: `normas etagCache[NormaFull]`, `historias etagCache[[]HistoriaGrupo]`.

**4. `GetLawHistory` + modelo tipado**
`HistoriaGrupo{Titulo, TipoDesc, TipoCod, Hls []HistoriaEntrada}` y `HistoriaEntrada{Tipo, IDNorma, Enlace, Bajada, Fecha, Descripcion, IDNormaHL}` — los campos que el usuario listó. OJO con el dominio: `hls[].tipo` (1/3/4) es el tipo de GRUPO de historia, NO el tipo de norma del catálogo — no aplicar canonical aquí. Norm_id inexistente → `[]` → el cliente devuelve slice vacío y la tool responde mensaje amable (sin ErrNormaNotFound). Nuevo recurso `get_law_history` en el YAML (timeout 10s, retry 2/500ms/4s, breaker 5/2/30s).

**Semántica de los IDs (información verificada en otra sesión — crítica)**: tres identificadores conviven en cada entrada y es fácil confundirlos: (1) `id_norma_hl` es el **idNorma de LeyChile de la norma dueña del registro** (la que nombran `descripcion`/`bajada` — p.ej. en el grupo modificatorias de la 21.600, la entrada de la 21.770 trae `id_norma_hl: 1216930`); (2) `id_norma` (sin sufijo) apunta a la **norma relacionada** (la modificatoria/modificada en ese contexto — no es lo mismo); (3) el número del `enlace` (/historia-de-la-ley/6910/) es un tercer id — el Historia ID del documento de tramitación, que NO sirve para LeyChile. **Regla de implementación**: los enlaces a la ficha de LeyChile se construyen SIEMPRE con `id_norma_hl` (`https://www.leychile.cl/Navegar?idNorma=<id_norma_hl>`); el texto formateado muestra el enlace construido (no el campo crudo `enlace` como sustituto); nunca derivar ids de la URL.

**5. Catálogo en `internal/bcn/norm_types.go` (patrón garbage.go)**
`map[int]normType{Abbr, Valor}` con los 40 tipos oficiales, comentario por bloque documentando el origen (`/Consulta/getTiposNorma`). Lookups: `canonicalNormType(cod int) (valor, abbr string, ok bool)`. Solo cod/abbr/valor — `orden`/`otro` del JSON no aportan (decisión del usuario: append, no replace). Los campos canonical se agregan en el pipeline de `GetNorma` (el cliente hace el trabajo feo), con `omitempty`: código desconocido → campos ausentes, raw intacto.

**6. "Version: as of …" en el texto**
Cuando `version_date` está presente, el header del texto de `get_law` y `get_law_summary` incluye la línea `Version: as of <fecha>` — el LLM debe saber qué versión está leyendo (decisión del usuario). El structured lleva el `version_date` solicitado como campo.

**7. LLM-first + structured en `get_law_history`**
Texto: los 3 grupos como secciones con sus entradas (fecha, ley, bajada, enlace). Structured: `HistoriaGrupo[]` tipado (no recursivo — sin riesgo de ciclo de schema). Convención heredada del config.yaml.

## Risks / Trade-offs

- [El refactor del caché toca el comportamiento verificado del ETag de normas] → Mitigación: los tests existentes (304, reemplazo por ETag nuevo, derivación de summary sin HTTP) se conservan y pasan con la clave compuesta.
- [La API podría cambiar el formato de fecha de `idVersion`] → Mitigación: el parse estricto falla rápido y el error es claro; el formato YYYY-MM-DD es el documentado por el propio frontend (fechas de vigencia).
- [`get_law_history` con caché desde el día 1 agrega superficie de tests] → Mitigación: mismo patrón que normas (suite + httptest con contador de requests); el 304 del endpoint ya fue verificado.
- [El catálogo hardcodeado puede desactualizarse si BCN agrega tipos] → Aceptado por el usuario: actualizar el archivo es trivial y el append con omitempty degrada con gracia (raw intacto).

## Migration Plan

Aditivo a nivel de MCP (parámetro opcional + tool nueva). Interno: `LawClient` cambia firma — los call sites se actualizan en el mismo change y el mock se regenera; no hay consumidores externos de la interfaz. El caché pasa a clave compuesta: las entradas previas se pierden al desplegar (caché en memoria, aceptado).

## Open Questions

<!-- Ninguna: validación estricta, version_date en summary, indicador de versión en el texto y caché de history desde el día 1 fueron decididos con el usuario; el gap de norm_id inexistente en history quedó verificado ([] → mensaje amable). -->
