## Context

El server tiene 3 tools (search_laws, get_law, get_law_summary) y el change activo `add-law-history-and-versions` agregará `version_date` y `get_law_history`. La expertise de uso vive hoy solo en el "Recommended System Prompt" del README — no está dentro del protocolo. Verificado en el go-sdk v1.7.0: `AddPrompt(&Prompt{Name, Title, Description, Arguments}, PromptHandler)` con `PromptArgument{Name, Title, Description, Required}` y `GetPromptResult.Messages`. Ver proposal.md para la motivación.

## Goals / Non-Goals

**Goals:**
- 6 prompts curados que codifican los flujos verificados del dominio, con la regla "verificar contra el texto real, nunca inventar" embebida en cada template.
- Surface MCP estándar: `prompts/list` + `prompts/get` sin dependencias nuevas.

**Non-Goals:**
- Prompts dinámicos que consulten BCN (templates puros — spec lo exige).
- Reescribir el "Recommended System Prompt" del README (sigue existiendo para clientes que no usan prompts).
- Prompts para búsqueda avanzada (filtros por tipo, facetas) — futuro.

## Decisions

**1. Templates puros, sin LawClient**
`internal/prompts/prompts.go` define handlers que solo inyectan los args recibidos en el template (`req.Params.Arguments["norm_id"]`). Cero dependencias del cliente de normas: `prompts/get` funciona aunque BCN esté caída (requisito del spec) y los tests no necesitan mocks del LawClient — solo el server in-memory.

**2. Los 6 prompts con sus templates (referencia de implementación)**
- `analyze_law(norm_id, aspect?)`: "Analyze Chilean norm <norm_id>... Step 1: call get_law_summary(norm_id) to understand scope. Step 2: call get_law(norm_id, structure_only=true) for the table of contents. Step 3: call get_law(norm_id) to read the full text. Structure the analysis: purpose, scope, obligations, sanctions, entry into force. Always cite the article number for every claim; never invent articles."
- `search_legal_topic(topic)`: "Find Chilean norms about <topic>... Step 1: search_laws(query=topic). Step 2: read the summaries in the results — do NOT open norms before reading their summaries. Step 3: pick the most relevant norm_id, call get_law_summary to confirm. Step 4: verify with get_law before stating anything as fact."
- `compare_law_versions(norm_id, from_date, to_date)`: "Compare norm <norm_id> between <from_date> and <to_date>. Call get_law(norm_id, version_date=<from_date>) and get_law(norm_id, version_date=<to_date>). Report what changed, which articles were modified, and when each change took effect. Verify both versions against the actual returned text."
- `trace_law_history(norm_id)`: "Trace the legislative history of norm <norm_id>. Call get_law_history(norm_id). Identify the 'modificatorias' group (laws that modified this norm). For each: date, law number, summary. To READ any of these norms, use get_law with the id_norma_hl value (the LeyChile id of the record's norm) — never the id_norma field nor the number in the history URL. Present a chronological timeline."
- `check_law_validity(norm_id, date?)`: "Check whether norm <norm_id> is in force<, as of date>". Call get_law_summary(norm_id<, version_date=date>). Report: in force / derogated / in force at the given date, based on the derogated flag and validity window. Distinguish 'vigente', 'derogada', 'vigente a la fecha'."
- `explain_law_simply(norm_id, audience?)`: "Explain norm <norm_id> in plain language< for audience>. Call get_law_summary(norm_id), then get_law(norm_id) to read the actual text. Explain without legal jargon. For every claim, cite the source article. End with: this explanation is not legal advice; consult the official text at bcn.cl."

**3. Idioma y convenciones heredadas**
Nombres/títulos/descripciones/argumentos en inglés; el template de mensaje en inglés (el LLM responde al usuario final en su idioma). Filenames snake_case, suite testify. Los prompts se registran desde `main.go` (`prompts.RegisterPrompts(srv)`) — sin tocar `RegisterTools`.

**4. Dependencia de orden**
`compare_law_versions`, `trace_law_history` y `check_law_validity` referencian `version_date`/`get_law_history`. Este change se aplica y archiva DESPUÉS de `add-law-history-and-versions` (documentado en el proposal). Los templates son texto: si se aplicara antes, no rompería el build — pero los prompts referenciarían tools inexistentes, por eso el orden.

**5. Tests**
Suite `internal/prompts/prompts_test.go` con el patrón `newTestClient` de tools (server in-memory + `prompts.RegisterPrompts`): `ListPrompts` muestra los 6 con argumentos y `required` correctos; `GetPrompt` de cada uno inyecta los args en el mensaje; `GetPrompt` sin `norm_id` requerido → comportamiento del SDK (el handler recibe el mapa, la validación de `required` es del cliente — se documenta en el test qué hace el SDK). Sin MockLawClient (no hay cliente involucrado).

## Risks / Trade-offs

- [Los templates son texto libre: si una tool cambia de nombre, el prompt queda desincronizado] → Mitigación: los nombres de tools referenciados se declaran como constantes en el paquete y los tests verifican que cada template referencia solo tools registradas (assert simple sobre el texto).
- [El SDK no valida argumentos requeridos en prompts/get] → Mitigación: el handler tolera arg ausente (lo deja vacío en el template); el `required` declarado en `prompts/list` es la guía para el cliente. Documentado en tests.
- [6 prompts = más superficie en prompts/list] → Aceptado: es el valor del feature (expertise curada); cada prompt tiene Title descriptivo para UIs.

## Migration Plan

Aditivo: prompts nuevos sin tocar tools. Rollback trivial (borrar el paquete y el registro).

## Open Questions

<!-- Ninguna: la lista de 6, los argumentos, la pureza de templates y el orden de dependencia fueron decididos con el usuario en la exploración. -->
