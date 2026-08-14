## 1. Docker

- [x] 1.1 Crear `Dockerfile` multi-stage con patrón `BUILDPLATFORM` (estándar Docker para Go): builder `FROM --platform=$BUILDPLATFORM golang:1.26` con `ARG TARGETOS TARGETARCH` y `GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/chile-bcn-mcp` → runtime `alpine:3.23` con `ca-certificates` + `curl`, usuario no-root `appuser`, `GOMEMLIMIT=256MiB`, `FASTMCP_HOST=0.0.0.0`, `EXPOSE 8000`, `ENTRYPOINT ["chile-bcn-mcp"]`
- [x] 1.2 Crear `docker-compose.yml`: build context `.`, passthrough de `FASTMCP_TRANSPORT`/`FASTMCP_HOST`/`FASTMCP_PORT`/`FASTMCP_PATH`/`MCP_AUTH_TOKEN` con defaults, `restart: unless-stopped`, healthcheck `curl -f http://localhost:8000/health`
- [x] 1.3 Crear `.dockerignore`: `.git/`, `.claude/`, `openspec/`, `bin/`, `.env`, `*.test`, `*.out` (mantener `go.mod`/`go.sum`)

## 2. Makefile

- [x] 2.1 Actualizar `build` a `CGO_ENABLED=0 -ldflags="-s -w"` hacia `bin/chile-bcn-mcp`
- [x] 2.2 Agregar `build-amd64` (`GOOS=linux GOARCH=amd64`) y `build-arm64` (`GOOS=linux GOARCH=arm64`) hacia `bin/`
- [x] 2.3 Agregar targets `fmt` y `clean`; mantener `run-http`, `run-stdio`, `test`, `vet`

## 3. GitHub Actions

- [x] 3.1 Crear `.github/workflows/ci.yml`: push + PR → checkout, setup-go (go 1.26), `go test ./...`, `go vet ./...`
- [x] 3.2 Crear `.github/workflows/publish.yml`: triggers `v*` + `workflow_dispatch`; env `REGISTRY=ghcr.io` e `IMAGE_NAME=${{ github.repository }}`; jobs: checkout → `setup-qemu-action` (necesario para la etapa runtime arm64) → `setup-buildx-action` → login GHCR (`GITHUB_TOKEN`) → `build-push-action` con `platforms: linux/amd64,linux/arm64`, `push: true`, tags `${{ github.ref_name }}` + `latest`, cache de buildx
- [x] 3.3 Validar sintaxis de ambos workflows con `actionlint` si está disponible (o revisión manual YAML)

## 4. Documentación

- [x] 4.1 Actualizar `README.md`: sección Docker (`docker compose up`, build manual, healthcheck) y tabla de env vars con el default de contenedor `FASTMCP_HOST=0.0.0.0`
- [x] 4.2 Agregar sección "Cómo reutilizar este scaffold": los 3 lugares con el nombre del binario (`Makefile BINARY`, `Dockerfile COPY`/`ENTRYPOINT`) y nota del primer push a GHCR (habilitar paquete en settings)

## 5. Verificación

- [x] 5.1 `make build-amd64` y `make build-arm64` producen binarios estáticos en `bin/`
- [x] 5.2 `docker build -t chile-bcn-mcp:test .` construye sin errores y el contenedor arranca; `curl /health` responde healthy dentro del contenedor — **verificado con podman** (equivalente sin daemon): build OK, run OK, `/health` healthy, smoke de tools dentro del contenedor
- [x] 5.3 `docker compose up -d` levanta el servicio y el healthcheck pasa; `docker compose down` limpia — **verificado con podman-compose 1.6.0** (el proyecto usa podman como runtime principal, docker es fallback): `up -d` construye y arranca, `ps` reporta `(healthy)` del healthcheck del compose, `down` elimina
- [x] 5.4 Verificar usuario no-root dentro del contenedor (no es `root`) — **verificado con podman**: `podman exec chile-bcn-mcp id -u` → `1000`
- [x] 5.5 `go test ./...` y `make vet` siguen en verde tras los cambios del Makefile
