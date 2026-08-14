## ADDED Requirements

### Requirement: Versión histórica de norma por fecha

Las tools `get_law` y `get_law_summary` DEBEN aceptar un parámetro opcional `version_date` en formato `YYYY-MM-DD` que selecciona la versión de la norma vigente a esa fecha; sin el parámetro DEBEN devolver la última versión. El formato DEBE validarse estrictamente: un valor que no sea una fecha válida DEBE devolver un error de argumentos sin consultar el servicio. El contenido de la respuesta DEBE corresponder a la versión solicitada (verificado con la API real: el texto cambia según la fecha).

#### Scenario: Versión histórica
- **WHEN** un cliente llama a `get_law` con `norm_id` de una norma modificada y `version_date: "2010-01-01"`
- **THEN** la respuesta contiene el texto de la norma vigente a esa fecha, distinto del texto vigente actual

#### Scenario: Sin fecha = última versión
- **WHEN** un cliente llama a `get_law` sin `version_date`
- **THEN** la respuesta contiene la última versión de la norma

#### Scenario: Fecha inválida
- **WHEN** un cliente llama a `get_law` con `version_date: "2010-13-45"` o `"basura"`
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

#### Scenario: Versión indicada en la respuesta
- **WHEN** un cliente llama a `get_law` con `version_date`
- **THEN** el texto de la respuesta indica la versión mostrada (p.ej. "Version: as of 2010-01-01")

### Requirement: Historia legislativa de una norma

El servidor DEBE exponer una tool `get_law_history` que consulta el endpoint de historias legislativas con `norm_id` (requerido) y devuelve los tres grupos oficiales — historia de la ley, historias de las leyes modificatorias e historias de las leyes modificadas — cada uno con sus entradas (fecha, descripción, bajada, enlace e identificadores). Un `norm_id` inexistente DEBE devolver un mensaje amable de que no hay historia (la API responde una lista vacía).

#### Scenario: Historia con modificaciones
- **WHEN** un cliente llama a `get_law_history` con `norm_id: 1195666`
- **THEN** la respuesta incluye los tres grupos con sus entradas, incluyendo las leyes que modificaron a la norma con su fecha y enlace

#### Scenario: Norma sin historia
- **WHEN** un cliente llama a `get_law_history` con un `norm_id` que no existe
- **THEN** la tool devuelve un mensaje amable indicando que no hay historia, sin error de protocolo

#### Scenario: Identificador faltante
- **WHEN** un cliente llama a `get_law_history` sin `norm_id`
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

#### Scenario: Enlaces construidos con el id correcto
- **WHEN** la respuesta de `get_law_history` incluye entradas con `id_norma_hl` e `id_norma`
- **THEN** los enlaces a la ficha de LeyChile se construyen con `id_norma_hl` (el idNorma de la norma dueña del registro), nunca con `id_norma` ni con el número del enlace de historia

### Requirement: Caché con revalidación por versión e historia

El caché ETag DEBE distinguir versiones: la clave de las normas DEBE componer `norm_id` y `version_date` (sin fecha = última), de modo que una versión histórica nunca reciba la respuesta cacheada de otra versión. La historia legislativa DEBE cachearse con revalidación ETag por `norm_id` (el endpoint envía ETag y responde 304, verificado).

#### Scenario: Versiones no se mezclan en caché
- **WHEN** un cliente solicita la misma norma con y sin `version_date`
- **THEN** cada versión se descarga y cachea por separado, sin servirse contenido de la otra

#### Scenario: Historia revalidada
- **WHEN** la misma historia se solicita dos veces y el servicio responde 304
- **THEN** la segunda respuesta se sirve del caché sin re-descargar

### Requirement: Tipo de norma decodificado como dato anexo

Los metadatos de la norma DEBEN anexar a cada tipo de norma los campos `canonical_type` y `canonical_abbr` decodificados del catálogo oficial (p.ej. `tipo: "1"` → `canonical_type: "Ley"`, `canonical_abbr: "LEY"`), SIN reemplazar los valores crudos de la API.

#### Scenario: Código decodificado
- **WHEN** una norma trae `tipos_numeros` con `tipo: "1"`
- **THEN** la respuesta incluye `canonical_type: "Ley"` y `canonical_abbr: "LEY"` junto al valor crudo

#### Scenario: Código desconocido
- **WHEN** un tipo trae un código fuera del catálogo
- **THEN** los campos canonical se omiten (omitempty) y el valor crudo permanece
