## Why

La tool `echo` fue el placeholder del scaffold: su propósito era validar que el transporte, el registro de tools y la generación de schemas funcionaran. Eso ya está cubierto por las tools reales (`search_laws`, `get_law`, `get_law_summary`) y por las suites de tests con in-memory transports. Una tool sin función es ruido para el LLM: consume tokens de contexto en cada `tools/list` e invita al modelo a llamarla sin razón. Además, el proyecto no tiene una forma formalizada de probar el server MCP de punta a punta — el smoke que se corre a mano cada vez debería ser un script reproducible.

## What Changes

- Eliminar la tool `echo`: handler, args, registro en `RegisterTools` y sus tests.
- Actualizar el README (quitar `echo` de la sección de tools).
- **REMOVED** del requirement "Tool demo echo" en la capability `mcp-server` (con razón y migración — es el primer delta REMOVED del proyecto, sigue el flujo completo del spec-driven).
- Crear `scripts/smoke.sh`: smoke test JSON-RPC contra el server real (health → initialize → tools/list → search_laws → get_law_summary → get_law), falla con exit 1 ante cualquier respuesta inesperada.
- Agregar el target `smoke` al Makefile (depende de que el server esté corriendo — documentado en el help).

## Capabilities

### New Capabilities

<!-- Ninguna. -->

### Modified Capabilities

- `mcp-server`: REMOVED del requirement "Tool demo echo" — la tool demo se elimina; la validación del scaffold queda cubierta por las tools reales y las suites.

## Impact

- **Código**: `internal/tools/tools.go` (quitar echo + `EchoArgs` + `makeEcho` + `errorResult` si queda sin uso), `internal/tools/tools_test.go` (quitar los 2 tests de echo), `README.md`, `scripts/smoke.sh` (nuevo), `Makefile` (target `smoke`).
- **Compatibilidad**: los clientes MCP verán `echo` desaparecer de `tools/list` — cambio **BREAKING** menor y deliberado (la tool era placeholder, documentado como reemplazable desde el scaffold).
- **Dependencias**: ninguna nueva.
