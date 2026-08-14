## 1. Setup del módulo

- [x] 1.1 Promover `github.com/modelcontextprotocol/go-sdk` v1.7.0 a dependencia directa en `go.mod` (`go mod tidy` tras el primer import)
- [x] 1.2 Crear la estructura de carpetas: `cmd/chile-bcn-mcp/`, `internal/server/`, `internal/tools/`

## 2. Server core

- [x] 2.1 Implementar `internal/server/server.go`: struct `Config` con `Transport`, `Host`, `Port`, `Path`, `AuthToken`, `DocsRoot`-less, y `LoadConfig()` leyendo `FASTMCP_TRANSPORT` (default `http`), `FASTMCP_HOST` (default `127.0.0.1`), `FASTMCP_PORT` (default `8000`), `FASTMCP_PATH` (default vacío → `/mcp`), `MCP_AUTH_TOKEN` (vacío = sin auth)
- [x] 2.2 Implementar `New(logger)`: `mcp.NewServer` con `Implementation{Name: "chile-bcn-mcp-server", Version: "0.1.0"}` y `ServerOptions` con instructions y `KeepAlive` de 5 minutos (`KeepAliveFailureThreshold: 3`)
- [x] 2.3 Escribir `internal/server/server_test.go` cubriendo `LoadConfig` con defaults y con overrides de env

## 3. Tools

- [x] 3.1 Implementar `internal/tools/tools.go`: struct `EchoArgs{Text string \`json:"text" jsonschema:"text to echo"\`}` y handler tipado `makeEcho()` devolviendo `mcp.ToolHandlerFor[EchoArgs, struct{}]`
- [x] 3.2 Implementar `RegisterTools(srv)` registrando la tool `echo` con `mcp.AddTool`
- [x] 3.3 Escribir `internal/tools/tools_test.go` cubriendo la llamada a `echo` (respuesta devuelve el texto exacto) y su presencia en `tools/list` via in-memory transports

## 4. Entrypoint

- [x] 4.1 Implementar `cmd/chile-bcn-mcp/main.go`: logger `slog` a stderr, `server.LoadConfig()`, log de arranque con transporte
- [x] 4.2 Implementar rama `stdio`: `srv.Run(ctx, &mcp.StdioTransport{})` con `signal.NotifyContext(SIGINT, SIGTERM)`
- [x] 4.3 Implementar `runHTTP`: `mcp.NewStreamableHTTPHandler` + middleware `auth.RequireBearerToken` con verifier estático y `AllowMissingExpiration: true` (solo si `AuthToken != ""`) + `http.ServeMux` con MCP en `cfg.Path` y `GET /health` + timeouts HTTP (Read 10s, Write 30s, Idle 120s)
- [x] 4.4 Implementar shutdown graceful del servidor HTTP: goroutine que espera `ctx.Done()`, `Shutdown` con timeout de 10s
- [x] 4.5 Verificar `go build ./...` y `go vet ./...` sin errores, y `go.mod` con go-sdk como dependencia directa

## 5. Tooling y documentación

- [x] 5.1 Crear `Makefile` con targets `run-http`, `run-stdio`, `build`, `test`
- [x] 5.2 Crear `.env.example` con todas las variables `FASTMCP_*` y `MCP_AUTH_TOKEN` comentadas
- [x] 5.3 Crear `README.md`: arranque en ambos transportes y configuración de cliente MCP (stdio: `command: "chile-bcn-mcp"`; HTTP: `url: "http://127.0.0.1:8000/mcp"` + `headers` si hay token)

## 6. Verificación final

- [x] 6.1 Prueba manual HTTP: `curl /health` responde healthy; `tools/list` responde la tool `echo` con y sin `MCP_AUTH_TOKEN` (verificando el 401 sin token)
- [x] 6.2 Prueba manual stdio: arrancar el binario con `FASTMCP_TRANSPORT=stdio` y confirmar sesión MCP limpia con un cliente
- [x] 6.3 `go test ./...` con todos los tests en verde
