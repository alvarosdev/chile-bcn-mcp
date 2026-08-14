## Why

El scaffold de `chile-bcn-mcp` cubre el servidor (stdio + HTTP) pero no su despliegue: no hay forma de correrlo en contenedor, no hay imágenes multi-arquitectura (amd64/arm64) ni CI. El patrón de despliegue ya está probado en `godot-mcp-docs` (Dockerfile multi-stage, non-root, healthcheck, build por arquitectura en GitHub Actions) y solo necesita adaptarse y **generalizarse** para que el scaffold sea copiable a cualquier otro proyecto MCP.

## What Changes

- Crear `Dockerfile` multi-stage (Go builder → `alpine` runtime) sin los pasos de dominio de godot-mcp-docs (docs, pandoc, tree): binario `CGO_ENABLED=0`, usuario non-root, `GOMEMLIMIT`, `FASTMCP_HOST=0.0.0.0` como default de contenedor.
- Crear `docker-compose.yml` con passthrough de env vars `FASTMCP_*`/`MCP_AUTH_TOKEN` y healthcheck sobre `/health`.
- Crear `.dockerignore` limpio (sin restos del proyecto anterior de docs).
- Fusionar `Makefile`: build con `-s -w`, targets de cross-compile `build-amd64`/`build-arm64` (heredados de godot-mcp-docs) sobre los targets actuales (`run-http`, `run-stdio`, `test`, `vet`).
- Crear `.github/workflows/ci.yml`: `go test` + `go vet` en cada push y PR.
- Crear `.github/workflows/publish.yml`: buildx multi-plataforma (`linux/amd64`, `linux/arm64`) con push a GHCR en tags `v*` y dispatch manual; tags `latest` + versión.
- Actualizar `README.md` con secciones de Docker y "cómo reutilizar este scaffold" (los 3 lugares donde se hardcodea el nombre del binario).

## Capabilities

### New Capabilities

- `container-deployment`: Despliegue del servidor en contenedor — imagen OCI multi-arquitectura publicada en GHCR, usuario no-root, healthcheck y configuración por variables de entorno.

### Modified Capabilities

<!-- Ninguna: el comportamiento del servidor (spec mcp-server) no cambia. -->

## Impact

- **Código**: solo archivos nuevos (`Dockerfile`, `docker-compose.yml`, `.dockerignore`, `.github/workflows/*`) y modificación de `Makefile` y `README.md`. `internal/` y `cmd/` no se tocan.
- **Dependencias**: ninguna nueva en Go; nuevas dependencias de infraestructura (GitHub Actions, Docker buildx) — todo gestionado por el CI.
- **Sin cambios breaking**: el servidor se comporta igual; los cambios son aditivos.
