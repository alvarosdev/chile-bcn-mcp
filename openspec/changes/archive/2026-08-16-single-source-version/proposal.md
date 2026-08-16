## Why

`VERSION` (`v0.0.6`) y `internal/server/server.go:57` (`"0.1.0"` hardcodeado) divergen y obligan a editar dos lugares por release. `mcp.NewServer` reporta una versión desactualizada y cada artefacto (binario local, `dist.zip`, imagen OCI) puede construirse con una versión distinta, rompiendo trazabilidad y `initialize` del protocolo MCP.

## What Changes

- Crear `internal/version/version.go` con `var Version = "dev"` como único símbolo de versión en código, inyectado por `-ldflags -X` en todos los builds.
- Cambiar `internal/server/server.go:New()` para usar `version.Version` (con `strings.TrimPrefix(v, "v")`) en `mcp.Implementation{Version}` en lugar del literal `"0.1.0"`.
- Hacer que `Makefile:build`, `scripts/build-dist.sh` y `Dockerfile` lean la versión desde el archivo `VERSION` (SSOT) y la inyecten vía `ldflags`; `make build` local usa `VERSION` por defecto, CI usa el output `needs.version` (ya normalizado sin `v`).
- Normalizar `VERSION` sin prefijo `v` o tolerarlo con `TrimPrefix` en código y `sed 's/^v//'` en Make; `go run` sin flags reporta `dev` (no miente versión).
- Eliminar drift: `grep -r "0.1.0"` deja de existir en el repo.

## Capabilities

### New Capabilities
- `versioning`: infraestructura de versionado de fuente única (archivo `VERSION` como SSOT, variable `internal/version.Version` inyectada por build, normalización y fallback).

### Modified Capabilities
- `mcp-server`: el servidor DEBE reportar su versión real (derivada de `VERSION`) en el handshake MCP `Implementation.Version`.
- `release-distributions`: las distribuciones y la imagen OCI DEBEN derivar su tag/versión del mismo SSOT (`VERSION` / `release/v*`).
- `container-deployment`: el build de la imagen DEBE inyectar la versión vía `ARG VERSION` + `ldflags`.

## Impact

- Código: `internal/version/version.go` (nuevo), `internal/server/server.go`, `Makefile`, `scripts/build-dist.sh`, `Dockerfile`, `.github/workflows/publish.yml` (build-args).
- APIs/MCP: `initialize` ahora retorna la versión correcta; clientes MCP ven `0.0.6` en lugar de `0.1.0` stale.
- Infra/Release: `make dist`, `make podman-build`, `docker/build-push-action` y `build-dist.sh` comparten la misma versión; `VERSION` es la única edición manual por release.
- Riesgo bajo: cambio build-time only; `dev` local es observable y testeable sin romper compatibilidad.
