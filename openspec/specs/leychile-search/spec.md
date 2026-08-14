## Purpose

Búsqueda de normas jurídicas chilenas en LeyChile (Biblioteca del Congreso Nacional): búsqueda paginada por texto y acceso al contenido de una norma por su identificador, con endpoints declarados externamente y transporte resiliente configurable.

## Requirements

### Requirement: Búsqueda paginada de normas

El servidor DEBE exponer una tool `search_laws` que consulta el endpoint de búsqueda de LeyChile con los parámetros `query` (texto a buscar, requerido), `page` (número de página, por defecto 1) y `page_size` (resultados por página, por defecto 10). La tool DEBE devolver los resultados de la página solicitada junto con el total de resultados, para que el cliente pueda navegar las páginas.

#### Scenario: Resultados en formato LLM-first con estructura opcional
- **WHEN** un cliente llama a `search_laws` con éxito
- **THEN** el resultado incluye contenido de texto formateado para lectura del modelo (lista de resultados con paginación y total)
- **AND** incluye contenido estructurado tipado (`structuredContent`) con los campos completos de los resultados, el total y la paginación

#### Scenario: Búsqueda simple
- **WHEN** un cliente llama a `search_laws` con `query: "Ley 21.827"`
- **THEN** la respuesta incluye los resultados de la primera página (cada uno con tipo de norma, número, título, fecha de publicación e `IDNORMA`)

#### Scenario: Navegación de páginas
- **WHEN** un cliente llama a `search_laws` con `query: "Ley 21.827"`, `page: 2` y `page_size: 5`
- **THEN** la respuesta incluye la página 2 y el total de resultados disponible, permitiendo saber cuántas páginas existen

#### Scenario: Cadena vacía
- **WHEN** un cliente llama a `search_laws` sin `query` o con `query` vacía
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

### Requirement: Contenido de norma por identificador

El servidor DEBE exponer una tool `get_law` que consulta el recurso de contenido de LeyChile con el identificador de norma (requerido, mapeado al parámetro `idNorma` del servicio) y devuelve la norma estructurada: metadatos seleccionados (tipo(s) de norma, organismos, título, fuente, materias, estado de derogación, fecha de publicación y vigencia), la estructura de la norma (título de cada parte con su identificador), los proyectos de ley relacionados y el contenido. La tool DEBE aceptar un parámetro `structure_only` que, cuando es verdadero, omite el contenido y devuelve metadatos, estructura y proyectos.

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

#### Scenario: Identificador inexistente
- **WHEN** un cliente llama a `get_law` con un `norm_id` que no existe en LeyChile
- **THEN** la tool devuelve un error de "norma no encontrada" sin contenido

#### Scenario: Identificador faltante
- **WHEN** un cliente llama a `get_law` sin `norm_id`
- **THEN** la tool devuelve un error de argumentos sin consultar el servicio

### Requirement: Contenido en Markdown

El contenido de la norma (los bloques HTML del servicio) DEBE entregarse convertido a **Markdown**, con las entidades HTML decodificadas y los enlaces conservados, para que el texto sea legible por el cliente.

#### Scenario: Contenido convertido
- **WHEN** la norma tiene bloques de contenido con entidades como `&#xED;` y enlaces a otras normas
- **THEN** la respuesta contiene el texto decodificado (`í`) en formato Markdown y los enlaces conservados como enlaces

### Requirement: Resumen legible para el cliente

El campo de resumen de cada resultado de búsqueda, cuando el servicio lo entrega como XML embebido con entidades HTML, DEBE devolverse decodificado, sin el marcado XML y sin el whitespace de indentación del wrapper, para que el texto sea legible.

#### Scenario: Resumen con entidades y XML
- **WHEN** un resultado de búsqueda trae un `RESUMEN` con `<RESUMENES>`, entidades como `&#241;` e indentación del wrapper XML
- **THEN** la respuesta contiene el texto decodificado (`ñ`) y sin etiquetas XML ni indentación residual

### Requirement: Contenido de la norma sin basura de marcado

El contenido de la norma (tras la conversión a Markdown) DEBE estar normalizado: los espacios no separadores y sus variantes (`&nbsp;`, `&ensp;`, `&emsp;` → U+00A0, U+2002, U+2003) DEBEN convertirse a espacios normales, los espacios al inicio y fin de cada línea DEBEN recortarse, los espacios consecutivos DEBEN colapsarse, y los caracteres de control y de ancho cero DEBEN eliminarse. Las comillas de citas (`&quot;`) y los enlaces a otras normas DEBEN conservarse como parte del contenido.

#### Scenario: Sangría visual normalizada
- **WHEN** un párrafo de la norma comienza con `&nbsp; &nbsp;  ` (sangría visual de la API)
- **THEN** la respuesta contiene el párrafo sin los espacios no separadores al inicio

#### Scenario: Comillas de citas conservadas
- **WHEN** el contenido incluye `&quot;Art&#xED;culo 1.- ...&quot;`
- **THEN** la respuesta conserva las comillas como caracteres normales alrededor del texto citado

#### Scenario: Enlaces conservados
- **WHEN** el contenido incluye un enlace a otra norma (`<a href="...idNorma=...">`)
- **THEN** la respuesta lo conserva como enlace en el texto Markdown

#### Scenario: Caracteres de control eliminados
- **WHEN** el contenido contiene caracteres de control (U+0000–U+001F fuera de salto de línea y tabulación) o de ancho cero (U+200B, U+FEFF)
- **THEN** la respuesta no los incluye

### Requirement: Endpoints declarados en archivo YAML

El servidor DEBE cargar la definición de sus recursos de LeyChile desde el archivo `config/api.resources.yaml` con **ruta fija** (sin override por variable de entorno), que declara para cada recurso: `url`, `path`, `method`, `timeout`, `retry` (intentos y backoff) y `circuit_breaker` (umbrales de fallo, éxito y cooldown). El contenido DEBE validarse al cargar; una configuración inválida DEBE impedir el arranque del servidor. La carpeta `config/` DEBE incluirse en la imagen de contenedor al momento del build.

#### Scenario: Carga exitosa con ruta fija
- **WHEN** el servidor arranca con un `config/api.resources.yaml` válido en su directorio de trabajo
- **THEN** los recursos quedan disponibles y el servidor continúa el arranque

#### Scenario: Configuración inválida
- **WHEN** el archivo YAML tiene un recurso sin `path`, con timeout negativo o con retry/breaker incoherentes
- **THEN** el servidor no arranca e informa el error de validación

#### Scenario: Configuración en la imagen de contenedor
- **WHEN** se construye la imagen de contenedor desde el Dockerfile
- **THEN** la imagen incluye la carpeta `config/` y el servidor arranca con la ruta fija sin configuración adicional

### Requirement: Transporte resiliente por recurso

Cada recurso DEBE aplicar su propio `timeout`, reintentos y circuit breaker declarados en el YAML: ante fallos transitorios (timeouts, respuestas 5xx) el cliente DEBE reintentar según la configuración del recurso, y ante fallos repetidos el circuit breaker DEBE abrirse y rechazar llamadas sin llegar al servicio hasta que se recupere.

#### Scenario: Falla transitoria
- **WHEN** el servicio responde con error 5xx o timeout en la primera llamada y responde bien en las siguientes
- **THEN** el cliente reintenta según la configuración del recurso y devuelve la respuesta exitosa

#### Scenario: Circuit breaker abierto
- **WHEN** el recurso acumula fallos por encima del umbral declarado
- **THEN** las llamadas al recurso se rechazan inmediatamente con un error de circuito abierto, sin intentar contactar el servicio

### Requirement: Caché de normas con revalidación ETag

El cliente DEBE cachear en memoria cada norma recuperada junto con su ETag (validado con la API real: `get_norma_json` responde `ETag` y `304 Not Modified` con cuerpo vacío ante un `If-None-Match` coincidente). En una llamada posterior al mismo `norm_id`, el cliente DEBE enviar el `If-None-Match` guardado; ante un `304` DEBE servir la copia cacheada sin re-descargar ni re-convertir el contenido; ante un `200` DEBE reemplazar la entrada cacheada. El caché vive en el proceso (se pierde al reiniciar, lo cual es aceptable) y NO aplica a la búsqueda.

#### Scenario: Hit de caché
- **WHEN** la misma norma se solicita dos veces y el servicio responde `304` en la segunda llamada
- **THEN** la segunda respuesta se sirve desde el caché con el mismo contenido, sin re-descargar el cuerpo ni re-convertirlo

#### Scenario: Norma actualizada
- **WHEN** el servicio responde `200` con un ETag distinto al guardado
- **THEN** la entrada cacheada se reemplaza con el contenido nuevo

#### Scenario: Reinicio sin caché
- **WHEN** el servidor se reinicia
- **THEN** el caché comienza vacío y la primera llamada a cada norma vuelve a descargarse

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
