## Context

Hoy `VERSION` contiene `v0.0.6` pero `internal/server/server.go:57` hardcodea `"0.1.0"` en `mcp.NewServer`. No hay inyección de versión en ningún camino de build: `Makefile:19` (`go build -ldflags="-s -w"`), `scripts/build-dist.sh:48` y `Dockerfile:22` compilan sin `-X`. El SSOT real es la rama `release/v*` solo en CI (`publish.yml:39 VERSION="${GITHUB_HEAD_REF#release/}"`), desconectado del código. `go:embed` no sirve para embebedar `../../VERSION` desde `internal/version` (patrón fuera del paquete → error de compilación). Ver constraints del proyecto: `BUILDPLATFORM` para cross-compile nativo, ruta `config/` embebida, `dev` como fallback local aceptado.

## Goals / Non-Goals

**Goals:**
- Un único lugar editable por release: `VERSION` (archivo en raíz, con o sin `v`).
- `mcp.NewServer` refleja exactamente esa versión (normalizada sin `v`), nunca un literal stale.
- `make build`, `make dist`/`build-dist.sh`, `docker build` y `publish.yml` inyectan la misma versión vía un único `ldflags -X`.
- `go run` sin flags sigue funcionando y reporta `dev` (no falla ni miente).

**Non-Goals:**
- Eliminar `VERSION` en favor de `git describe` / `debug.ReadBuildInfo` (acopla a tags git y rompe `release/v*` branch gate).
- Usar `go:embed` con copia/symlink de `VERSION` dentro de `internal/version` (duplica SSOT y añade paso frágil).
- Versionar el contrato `config/api.resources.yaml` (su `version: 1` es independiente).
- Cambiar el flujo de release (sigue siendo `release/v*` merge → draft).

## Decisions

**1. `internal/version.Version var string = "dev"` + `ldflags -X` (elegido) vs `go:embed`**
- Por qué: estándar Go (`goreleaser`, `kubernetes`), sin archivo runtime, un solo `-X` propaga a todo. `embed` obligaría duplicar `VERSION` dentro del paquete o `//go:embed ../../VERSION` que el compilador rechaza. `ReadBuildInfo` depende de `vcs` tag y no cubre `release/v*` branch sin tag previo.
- Alternativas descartadas: `embed` con `go:generate cp ../../VERSION ./VERSION` (drift), `os.ReadFile("VERSION")` en runtime (falla en contenedor/dist donde `VERSION` no se copia).

**2. Normalización en dos capas: Make `sed 's/^v//'` + Go `strings.TrimPrefix(v,"v")`**
- Por qué: `VERSION` hoy es `v0.0.6`; CI ya hace `VERSION="${VERSION#v}"`. Doble trim hace idempotente soportar ambos formatos sin migración forzada. El MCP `Version` se espera semver sin `v` (`0.1.0` actual).
- Alternativa: forzar `VERSION` sin `v` y fallar si viene con `v` — más frágil para contribuidores.

**3. `VERSION` permanece como SSOT en raíz (no mover a `internal/version/`)**
- Por qué: convención visible, usada por scripts/CI/humanos (`cat VERSION`). Moverlo es churn y rompe expectativas. Inyección via Make variable `$(shell cat VERSION)` mantiene lectura en un solo lugar.

**4. `ARG VERSION=dev` en Dockerfile + `build-args` en `publish.yml`**
- Por qué: `docker/build-push-action` soporta `build-args` por plataforma; sin `ARG` el build local quedaría en `dev` silenciosamente. Default `dev` alinea con `go run`.

## Risks / Trade-offs

- **Olvidar `ldflags` en un nuevo camino de build → reporta `dev` en release** → Mitigación: `grep` en CI que falle si algún `go build` no contiene `-X.*version.Version`; documentar en `Makefile` como variable `LDFLAGS` reutilizable.
- **`go run` reporta `dev` y confunde al testear versión** → Mitigación: documentar `make run` vs `go run -ldflags` ; health check no expone versión, solo `initialize` MCP (testeable con `CallTool` + assert).
- **`VERSION` con salto de línea/espacios** → Mitigación: `tr -d ' \n'` y `strings.TrimSpace` antes de `TrimPrefix`.
- **Cache de `go build` no invalida al cambiar `VERSION` si solo cambia `ldflags`** → Mitigación: `ldflags` cambia el `actionID`, Go invalida; verificado con `go build -a` no necesario.

## Migration Plan

1. Crear `internal/version/version.go`.
2. Editar `internal/server/server.go` para importar `version` y usar `strings.TrimPrefix`.
3. Editar `Makefile` (variable `VERSION` + `LDFLAGS`), `scripts/build-dist.sh` (añadir `-X`), `Dockerfile` (`ARG VERSION` + `ldflags`), `.github/workflows/publish.yml` (añadir `build-args: VERSION=...` y asegurar `setup-go` usa `VERSION` sin `v`).
4. `make build && make test && make vet` local; verificar `initialize` retorna `$(cat VERSION | sed 's/^v//')`.
5. Opcional: normalizar `VERSION` a sin `v` en siguiente release (`v0.0.6` → `0.0.6`) — no bloqueante por doble trim.
- Rollback: revertir 5 archivos y volver a literal; ningún dato migra.

## Open Questions

- ¿Exponer versión también en `GET /health`? No incluido: expone superficie y no está en spec; puede añadirse luego sin cambiar `versioning` ni `mcp-server` si se decide (nueva requirement en `mcp-server`).
- ¿Añadir `make version` helper que imprima `version.Version`? Útil pero no necesario para SSOT; deferrable.
