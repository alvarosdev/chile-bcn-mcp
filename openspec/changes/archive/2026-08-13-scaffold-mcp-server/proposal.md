## Why

`chile-bcn-mcp` está vacío: solo tiene `go.mod` (con el go-sdk v1.7.0 declarado como indirecto) y la estructura OpenSpec. Necesitamos un scaffold mínimo de un MCP server en Go compatible con **stdio y streamable HTTP** sobre el que construir las features reales (aún por definir). El patrón base está probado en `godot-mcp-docs` (mismo SDK, transportes conmutables, auth por token) y fue revisado contra los ejemplos oficiales del go-sdk (`examples/server/hello`, `memory`, `everything`).

## What Changes

- Crear la estructura base del proyecto: `cmd/chile-bcn-mcp/main.go`, `internal/server/`, `internal/tools/`.
- Servidor MCP con un único binario y transporte conmutable por env var (`FASTMCP_TRANSPORT` = `stdio` | `http`).
- Transporte HTTP streamable en `/mcp` con auth Bearer opcional por token estático (`MCP_AUTH_TOKEN`), endpoint `/health` y shutdown graceful.
- Config por variables de entorno (`FASTMCP_HOST`, `FASTMCP_PORT`, `FASTMCP_PATH`, `MCP_AUTH_TOKEN`) cargada una vez al arranque.
- Una tool demo (`echo`) como placeholder de dominio, registrada con el patrón `RegisterTools` + handlers tipados.
- `Makefile`, `.env.example` y `README.md` con las configuraciones de cliente MCP para ambos transportes.
- `go.mod`: el go-sdk pasa de dependencia indirecta a directa.
- Test básico del `LoadConfig` y del handler de la tool demo.

## Capabilities

### New Capabilities

- `mcp-server`: Capacidad base del servidor MCP — transporte stdio y streamable HTTP, configuración por entorno, autenticación opcional, health check y registro de tools.

### Modified Capabilities

<!-- Ninguna: el proyecto no tiene specs previas. -->

## Impact

- **Código**: solo archivos nuevos en `chile-bcn-mcp`; `godot-mcp-docs` no se toca (queda como referencia de patrón).
- **Dependencias**: `github.com/modelcontextprotocol/go-sdk` v1.7.0 pasa de indirect a directa; no se agregan dependencias nuevas.
- **Sin cambios breaking**: no existe código previo que migrar.
