## MODIFIED Requirements

### Requirement: Contenido de norma por identificador

El servidor DEBE exponer una tool `get_law` que consulta el recurso de contenido de LeyChile con el identificador de norma (requerido, mapeado al parámetro `idNorma` del servicio) y devuelve la norma estructurada: metadatos seleccionados (tipo(s) de norma, organismos, título, fuente, materias, estado de derogación, fecha de publicación y vigencia), la estructura de la norma (título de cada parte con su identificador), los proyectos de ley relacionados y el contenido. La tool DEBE aceptar un parámetro `structure_only` que, cuando es verdadero, omite el contenido y devuelve metadatos, estructura y proyectos. La tool DEBE aceptar un parámetro opcional `section_id` (identificador de una parte de la estructura, el mismo `i` que la API asigna a cada bloque de la norma) que, cuando se indica, limita el contenido devuelto al subárbol de esa parte — sin omitir metadatos ni estructura.

#### Scenario: Resultados en formato LLM-first con estructura opcional
- **WHEN** un cliente llama a `get_law` con éxito
- **THEN** el resultado incluye contenido de texto formateado para lectura del modelo (metadatos, estructura y contenido en Markdown)
- **AND** incluye contenido estructurado tipado (`structuredContent`) con los mismos campos; el campo de contenido se omite cuando `structure_only` es verdadero

#### Scenario: Norma válida con contenido
- **WHEN** un cliente llama a `get_law` con `norm_id: 1226950`
- **THEN** la respuesta incluye los metadatos de la norma, su estructura, los proyectos relacionados y el contenido completo de la norma

#### Scenario: Solo estructura
- **WHEN** un cliente llama a `get_law` con `norm_id: 1226950` y `structure_only: true`
- **THEN** la respuesta incluye metadatos, estructura y proyectos pero omite el contenido

#### Scenario: Sección específica
- **WHEN** un cliente llama a `get_law` con `norm_id` de una norma larga y un `section_id` que corresponde a un título de su estructura
- **THEN** la respuesta incluye los metadatos y la estructura completos, pero el contenido se limita a los bloques de esa sección y sus descendientes, omitiendo el resto de la norma
- **AND** el texto de la respuesta indica qué sección se muestra (p.ej. "Section: Título III")

#### Scenario: Sección inexistente
- **WHEN** un cliente llama a `get_law` con un `section_id` que no corresponde a ninguna parte de la estructura de la norma
- **THEN** la tool devuelve un error de argumentos indicando que la sección no existe, sin devolver contenido

#### Scenario: Identificador inexistente
- **WHEN** un cliente llama a `get_law` con un `norm_id` que no existe en LeyChile
- **THEN** la tool devuelve un error de "norma no encontrada" sin contenido

#### Scenario: Identificador faltante
- **WHEN** un cliente llama a `get_law` sin `norm_id`
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

### Requirement: Resumen de norma por identificador

El servidor DEBE exponer una tool `get_law_summary` que consulta el mismo recurso de contenido de LeyChile que `get_law` y devuelve una versión liviana de la norma: `titulo_norma`, `fuente`, `materias`, `categorias_norma`, `resumenes` (sanitizados) y `estructura` (título de cada parte con su identificador, para poder solicitar secciones con `get_law`). La tool DEBE requerir `norm_id` (mapeado al parámetro `idNorma` del servicio) y NO DEBE incluir el contenido de la norma ni convertirlo a Markdown.

#### Scenario: Resumen válido
- **WHEN** un cliente llama a `get_law_summary` con `norm_id: 1142880`
- **THEN** la respuesta incluye el título de la norma, la fuente, las materias, las categorías de norma, los resúmenes sanitizados y la estructura con los identificadores de cada parte
- **AND** no incluye el contenido de la norma

#### Scenario: Identificador inexistente
- **WHEN** un cliente llama a `get_law_summary` con un `norm_id` que no existe en LeyChile
- **THEN** la tool devuelve un error de "norma no encontrada" sin contenido

#### Scenario: Identificador faltante
- **WHEN** un cliente llama a `get_law_summary` sin `norm_id`
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

#### Scenario: Resultados en formato LLM-first con estructura opcional
- **WHEN** un cliente llama a `get_law_summary` con éxito
- **THEN** el resultado incluye contenido de texto formateado para lectura del modelo
- **AND** incluye contenido estructurado tipado (`structuredContent`) con los mismos campos

## ADDED Requirements

### Requirement: Tamaño de la norma en el output

El `structuredContent` de `get_law` DEBE incluir `char_count` (cantidad de caracteres del contenido devuelto, convertido a Markdown) y `article_count` (cantidad de artículos del contenido devuelto): sin `section_id` corresponden a la norma completa; con `section_id`, a esa sección. El `structuredContent` de `get_law_summary` DEBE incluir `char_count` y `article_count` de la norma completa, para que el cliente conozca la magnitud del documento antes de solicitar contenido.

#### Scenario: Conteo en get_law
- **WHEN** un cliente llama a `get_law` con éxito sobre una norma con contenido
- **THEN** el `structuredContent` incluye `char_count` y `article_count` con valores mayores a cero

#### Scenario: Conteo en get_law_summary
- **WHEN** un cliente llama a `get_law_summary` con éxito
- **THEN** el `structuredContent` incluye `char_count` y `article_count`

#### Scenario: Conteo corresponde a la versión solicitada
- **WHEN** un cliente llama a `get_law` con `version_date` de una versión histórica
- **THEN** `char_count` y `article_count` corresponden a la versión solicitada

#### Scenario: Conteo de la sección solicitada
- **WHEN** un cliente llama a `get_law` con `section_id`
- **THEN** `char_count` y `article_count` corresponden al contenido de esa sección, no al total de la norma

#### Scenario: Tamaño visible en el texto
- **WHEN** un cliente llama a `get_law` o `get_law_summary` con éxito
- **THEN** el texto de la respuesta incluye el tamaño (p.ej. "Size: 426K chars · 154 articles")
