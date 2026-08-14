## 1. Paquete de prompts

- [x] 1.1 Crear `internal/prompts/prompts.go`: constantes con los nombres de tools referenciados y `RegisterPrompts(srv)` que registra los 6 prompts con `mcp.AddPrompt` (Name/Title/Description/Arguments con `required` donde corresponde)
- [x] 1.2 Implementar los handlers de template puro (solo inyectan los argumentos en el mensaje, sin clientes externos): `analyze_law`, `search_legal_topic`, `compare_law_versions`, `trace_law_history`, `check_law_validity`, `explain_law_simply` — con los textos de referencia del design (Decisión 2), incluyendo la regla "verify against the text, never invent" y el disclaimer de no-asesoría en `explain_law_simply`
- [x] 1.3 Registrar `prompts.RegisterPrompts(srv)` en `cmd/chile-bcn-mcp/main.go` (junto a `RegisterTools`)

## 2. Tests

- [x] 2.1 Escribir `internal/prompts/prompts_test.go` (suite, server in-memory + RegisterPrompts): `ListPrompts` muestra los 6 con sus argumentos y `required` correctos
- [x] 2.2 Tests de `GetPrompt` por prompt: `analyze_law` inyecta `norm_id` en el mensaje, `compare_law_versions` inyecta ambas fechas y referencia `get_law` + `version_date`, `explain_law_simply` contiene el disclaimer, `trace_law_history` referencia `get_law_history`, arg ausente → template se sirve sin error (comportamiento del SDK documentado)
- [x] 2.3 Verificar que cada template referencia solo tools registradas (assert simple sobre el texto vs. la lista de `RegisterTools`)

## 3. Verificación y documentación

- [x] 3.1 `make check` en verde
- [x] 3.2 Smoke real (no en CI): `prompts/list` del server real muestra los 6; `prompts/get` de `analyze_law` con norm_id inyectado; `prompts/get` funciona con el server sin acceso a BCN (solo plantilla)
- [x] 3.3 Actualizar `README.md`: sección de Prompts (tabla con los 6, argumentos y cuándo usarlos) + nota de que los prompts codifican el flujo recomendado del sistema (complementan el "Recommended System Prompt")
