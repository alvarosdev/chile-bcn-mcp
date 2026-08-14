## Context

`chile-bcn-mcp` es un proyecto nuevo: solo `go.mod` (con `modelcontextprotocol/go-sdk` v1.7.0 declarado como indirecto) y la estructura OpenSpec. Como plantilla se usa `godot-mcp-docs`, un MCP server Go funcional (stdio + streamable HTTP, auth Bearer, health check, graceful shutdown) del que se hereda el patrón de configuración por entorno y registro de tools. Los ejemplos oficiales del SDK (`examples/server/hello`, `memory`, `everything`, `examples/http`) fueron revisados para validar y refinar ese patrón. Ver proposal.md para la motivación.

## Goals / Non-Goals

**Goals:**
- Estructura de carpetas estándar Go (`cmd/` + `internal/`) que escale cuando llegue el dominio real.
- Un solo binario con transporte conmutable — misma instancia de `mcp.Server` bajo stdio o streamable HTTP.
- Config 100% por variables de entorno, cargada una vez al arranque.
- Código mínimo: cero features de dominio hasta que el usuario las defina.

**Non-Goals:**
- Funcionalidad de dominio (docs store, search, resources, prompts).
- Docker, CI, despliegue.
- Tests de integración con cliente MCP real (se prueban con `curl`/cliente manual).
- Autenticación OAuth (solo token estático).
- Binario dual client|server como `examples/http`.

## Decisions

**1. Layout `cmd/` + `internal/` en vez de todo en `package main`**
Los ejemplos del SDK son demos de un solo paquete (hasta 524 líneas en un `main.go`). Para un proyecto que va a crecer, el layout estándar Go separa entrypoint (`cmd/chile-bcn-mcp/main.go`) de lógica (`internal/`), y coincide con la estructura interna del propio go-sdk y con godot-mcp-docs.
*Alternativa descartada*: replicar los ejemplos tal cual (un `main.go` gigante).

**2. Un solo `internal/tools/tools.go` con `RegisterTools` + closures `make*`**
Se hereda el patrón de godot-mcp-docs (`RegisterTools(srv)` + `make*Handler(deps)` que devuelve `mcp.ToolHandlerFor[Args, Result]`), que es exactamente el mecanismo que el SDK usa para inyectar dependencias a los handlers. Pero en lugar de un archivo por feature (como `navigation.go`/`overview.go`), un único `tools.go` con la tool demo — la división por archivos vendrá sola cuando exista dominio con 3+ tools relacionadas, igual que `memory` separó `kb.go` por dominio.
*Alternativa descartada*: registración inline en `main()` (patrón `hello`), que acopla el setup del server con las tools.

**3. Config por env vars (`FASTMCP_*`, `MCP_AUTH_TOKEN`) en vez de flags**
Los ejemplos del SDK usan `flag` (`-http=:8080`). Para MCP, las env vars son superiores: los clientes MCP lanzan el proceso y le inyectan el entorno (no pueden pasar flags fácilmente), y es el patrón que ya funciona en godot-mcp-docs con Docker. El `Makefile` compensa la ergonomía en dev local (`make run-http`, `make run-stdio`).

**4. `mcp.AddTool` genérico con structs de args tipados (`json` + `jsonschema` tags)**
El SDK genera los esquemas JSON automáticamente desde los structs. Se evita el `struct{}` anónimo: la tool demo usa un struct nombrado (`EchoArgs`), que documenta mejor y sirve de plantilla para las tools futuras.

**5. Auth Bearer estático con `AllowMissingExpiration: true`**
Heredado de godot-mcp-docs, donde quedó depurado (issue real: el SDK rechaza tokens sin expiración a menos que se habilite el flag). OAuth se agrega después sin tocar el spec — solo cambia el verifier.

**6. `KeepAlive` de 5 min en `ServerOptions`**
Mitiga el leak de goroutines conocido del streamable HTTP (issue #499 del go-sdk) con sesiones abandonadas. Ya validado en godot-mcp-docs.

**7. Tool demo: `echo`**
Placeholder genérico hasta conocer el dominio de chile-bcn. Reemplazable por tools reales sin tocar nada más que `RegisterTools` (el spec lo declara explícitamente como reemplazable).

**8. go-sdk v1.7.0 como dependencia directa**
`go.mod` ya lo declara (indirecto); al importar `mcp` y `auth` pasa a directa. Sin dependencias nuevas.

## Risks / Trade-offs

- [Un solo `tools.go` puede crecer con las tools del dominio] → Mitigación: dividir por dominio cuando haya 3+ tools, siguiendo el patrón ya probado en godot-mcp-docs (`navigation.go`, `overview.go`).
- [Env vars menos visibles que flags para desarrollo] → Mitigación: `Makefile` con targets `run-http`/`run-stdio` y `.env.example` documentado.
- [Token estático no sirve para multi-usuario] → Fuera de alcance; el spec solo exige auth opcional, y el punto de extensión (verifier de `RequireBearerToken`) ya está aislado.
- [La tool `echo` será código muerto cuando llegue el dominio] → Costo de reemplazo de ~10 líneas en un solo archivo; aceptado como costo del scaffold.

## Migration Plan

Proyecto nuevo sin código previo: no hay migración ni rollback. `godot-mcp-docs` permanece intacto como referencia.

## Open Questions

- **Dominio de "chile-bcn"**: el nombre del `Implementation` y las tools reales quedan como `chile-bcn-mcp-server` y `echo` hasta que el usuario defina el dominio. No afecta specs, diseño ni tasks (el spec declara `echo` como reemplazable).
- **LoggingTransport para debug en stdio**: envolver el transporte stdio para volcar tráfico MCP a stderr es 3 líneas y se decide en implementación, no cambia el comportamiento observable.
