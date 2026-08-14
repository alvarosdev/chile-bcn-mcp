## Why

`get_law` devuelve la norma completa (contenido HTML convertido a Markdown + metadatos), un payload pesado para cuando el modelo solo necesita saber **de qué trata una norma** antes de decidir si leerla completa. Se necesita una tool ligera que use el mismo endpoint (`get_norma_json` — no existe un endpoint de resumen en la API) y devuelva únicamente los metadatos clave: `resumenes`, `titulo_norma`, `fuente`, `materias` y `categorias_norma`. Además, el usuario detectó que `get_law` no entrega `categorias_norma` (etiquetas temáticas de la norma, p.ej. Ley "Educación sin Dicom") — debe agregarse a sus metadatos.

## What Changes

- Crear la tool MCP **`get_law_summary`** (args: `norm_id`, requerido) que consulta el mismo recurso `get_law` de `api.resources.yaml` y devuelve una versión liviana: `titulo_norma`, `fuente`, `materias`, `categorias_norma` y `resumenes` (sanitizados). Sin contenido ni conversión a Markdown — el ahorro es para el cliente: payload mínimo y cero costo de conversión.
- Agregar `categorias_norma` al struct `Metadatos` de `get_law` (campo nuevo, aditivo — los consumidores actuales no se afectan).
- La tool nueva sigue la norma del proyecto: texto formateado para el LLM (`content[]`) + `structuredContent` tipado (`NormaSummary` con schema autogenerado), mismo manejo de errores (`ErrNormaNotFound` → "norma not found", `norm_id` inválido → error de argumentos).
- El summary **comparte el caché ETag** de `get_law` (misma entrada por `norm_id`): si la norma está cacheada, el summary se deriva de la entrada sin tocar la red.

## Capabilities

### New Capabilities

- `leychile-search`: dos requirements ADDED — la tool de resumen por identificador y el campo `categorias_norma` en los metadatos de `get_law`.

### Modified Capabilities

<!-- Ninguna. Nota de secuencia: la capability leychile-search aún no está en las specs principales (el change add-bcn-leychile-tools está activo sin archivar). Este change declara requirements ADDED para no modificar texto de un requirement que solo existe en el delta del otro change. Orden de archivo: add-bcn-leychile-tools PRIMERO, luego este. -->

## Impact

- **Código**: `internal/bcn` (campo `CategoriasNorma` en `Metadatos`, método `GetNormaSummary` en `LawClient` — regenerar mock), `internal/tools/get_law_summary.go` (nueva tool + registro en `RegisterTools`), tests con suite + mock.
- **Dependencias**: ninguna nueva.
- **Sin cambios breaking**: `get_law` sigue igual; el campo nuevo de metadatos es aditivo; la tool nueva no altera las existentes.
