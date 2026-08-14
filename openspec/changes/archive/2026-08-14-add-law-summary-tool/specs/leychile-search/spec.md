## ADDED Requirements

### Requirement: Resumen de norma por identificador

El servidor DEBE exponer una tool `get_law_summary` que consulta el mismo recurso de contenido de LeyChile que `get_law` y devuelve una versión liviana de la norma: `titulo_norma`, `fuente`, `materias`, `categorias_norma` y `resumenes` (sanitizados). La tool DEBE requerir `norm_id` (mapeado al parámetro `idNorma` del servicio) y NO DEBE incluir el contenido de la norma ni convertirlo a Markdown.

#### Scenario: Resumen válido
- **WHEN** un cliente llama a `get_law_summary` con `norm_id: 1142880`
- **THEN** la respuesta incluye el título de la norma, la fuente, las materias, las categorías de norma y los resúmenes sanitizados
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

### Requirement: Categorías de norma en metadatos

Los metadatos de la norma entregados por `get_law` DEBEN incluir el campo `categorias_norma` tal como lo entrega el servicio (etiquetas temáticas de la norma).

#### Scenario: Norma con categorías
- **WHEN** se solicita una norma cuyo response incluye `categorias_norma` (p.ej. `["Ley \"Educación sin Dicom\""]`)
- **THEN** los metadatos de la respuesta incluyen esas categorías en el mismo orden
