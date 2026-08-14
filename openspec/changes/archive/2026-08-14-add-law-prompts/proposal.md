## Why

El server expone 3 tools, pero la **expertise de cómo usarlas bien** vive fuera del sistema: en el "Recommended System Prompt" del README, que el usuario debe copiar a mano. MCP tiene una surface nativa para esto — **Prompts**: templates curados por el server con argumentos, que el cliente descubre (`prompts/list`) e invoca (`prompts/get`). Cada prompt codifica un flujo verificado (buscar → resumir → leer → verificar), los nombres reales de las tools, y las reglas de oro del dominio legal (verificar contra el texto real, nunca inventar artículos, disclaimer de no-asesoría). Se agregan 6 prompts curados cubriendo los flujos de valor del dominio.

## What Changes

- Crear `internal/prompts` con 6 prompts curados, cada uno con Name/Title/Description/Arguments y template de mensaje que inyecta los argumentos recibidos:
  - **`analyze_law(norm_id, aspect?)`** — análisis jurídico estructurado de una norma (summary → estructura → texto completo → dimensiones).
  - **`search_legal_topic(topic)`** — búsqueda guiada: elegir query, leer summaries antes de abrir, verificar con el texto.
  - **`compare_law_versions(norm_id, from_date, to_date)`** — comparar dos versiones históricas con `version_date` (requiere el change add-law-history-and-versions).
  - **`trace_law_history(norm_id)`** — rastrear qué leyes modificaron a la norma vía `get_law_history` (requiere el change add-law-history-and-versions).
  - **`check_law_validity(norm_id, date?)`** — verificar vigencia actual o a una fecha (derogado/vigente/vigente-a-la-fecha).
  - **`explain_law_simply(norm_id, audience?)`** — explicación en lenguaje simple con citas al texto y disclaimer de no-asesoría legal.
- Registrar los prompts en el server (`RegisterPrompts(srv)`) — los prompts son **templates puros**: no llaman a la API de BCN al servirse.
- Los prompts quedan visibles en `prompts/list` y servidos por `prompts/get` (cliente estándar MCP).

## Capabilities

### New Capabilities

- `law-prompts`: Prompts MCP curados que guían al cliente en el uso correcto de las tools del dominio LeyChile, con argumentos tipados y reglas de dominio embebidas.

### Modified Capabilities

<!-- Ninguna. -->

## Impact

- **Código**: `internal/prompts/prompts.go` (6 templates + registro), `cmd/chile-bcn-mcp/main.go` (llamar `RegisterPrompts`), tests de `prompts/list` y `prompts/get`.
- **Compatibilidad**: aditivo — las tools no cambian; los clientes ven una surface nueva.
- **Dependencias**: ninguna nueva. **Dependencia de orden**: este change se aplica y archiva DESPUÉS de `add-law-history-and-versions` (3 de los 6 prompts referencian `version_date`/`get_law_history`).
