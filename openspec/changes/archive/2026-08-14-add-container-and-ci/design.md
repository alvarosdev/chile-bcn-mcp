## Context

El servidor ya está implementado y probado (`cmd/`, `internal/`, Makefile básico). La referencia de despliegue es `godot-mcp-docs`, que tiene un Dockerfile multi-stage probado (builder → alpine, non-root, `GOMEMLIMIT`), compose con healthcheck y un workflow de build multi-arch nativo. Ver proposal.md para la motivación; los requisitos observables están en specs/container-deployment.

## Goals / Non-Goals

**Goals:**
- Despliegue en contenedor con las mismas prácticas que godot-mcp-docs (multi-stage, no-root, healthcheck).
- Imágenes multi-arquitectura (amd64 + arm64) publicadas en GHCR con el mínimo de complejidad.
- CI que valide el código en cada push/PR, separado de la publicación.
- Archivos lo más "genéricos" posible: que funcionen copiados a otro proyecto MCP Go con solo renombrar el binario.

**Non-Goals:**
- Automatización de releases (changelog, crear tags desde CI).
- Publicación a Docker Hub.
- Despliegue a un entorno (k8s, VPS).
- Build nativo por runner (matrix con `ubuntu-24.04-arm`) — descartado en favor de buildx simple.

## Decisions

**1. Dockerfile: estándar Docker para Go multi-platform (patrón `BUILDPLATFORM`), sin dominio**
Etapa builder con `FROM --platform=$BUILDPLATFORM golang:1.26` + `ARG TARGETOS TARGETARCH` + `GOOS=${TARGETOS} GOARCH=${TARGETARCH}`: la compilación corre siempre en la plataforma nativa del runner (sin emulación), porque Go es cross-compilador nativo. Es el patrón canónico de la doc oficial de Docker (`docs.docker.com/build/guide/multi-platform/`), que existe "specifically to prevent emulation from kicking in" y advierte que QEMU "can be much slower than native builds, especially for compute-heavy tasks like compilation". Runtime `alpine:3.23` con `ca-certificates` + `curl`, usuario no-root (`adduser -D appuser`), `GOMEMLIMIT=256MiB`, `FASTMCP_HOST=0.0.0.0` (default de contenedor — ver spec), `EXPOSE 8000`. Se omiten los pasos de docs de godot-mcp-docs (COPY docs/, tree, pandoc). El `curl` se conserva porque lo usa el healthcheck del compose; `ca-certificates` queda para futuras llamadas HTTPS salientes (el go-sdk la soporta con OAuth). **Añadido en verificación (podman)**: la imagen DEBE incluir `COPY config/ /app/config/` + `WORKDIR /app` — el servidor carga su contrato de endpoints desde `config/api.resources.yaml` por ruta fija (sin env override) y sin esa copia la imagen arranca y muere con "no such file or directory".
*Alternativa descartada*: `scratch`/`distroless` — sin shell ni curl, el healthcheck del compose (y el debugging) se complican sin ganancia de seguridad relevante para este binario estático.
*Alternativa descartada*: emulación pura (Dockerfile sin `BUILDPLATFORM`) — funciona sin cambios de Dockerfile, pero emula la compilación completa del arm64; el patrón oficial lo evita explícitamente.

**2. buildx simple multi-platform en vez de matrix nativo**
Un solo job con la secuencia canónica de GitHub Actions: `docker/setup-qemu-action` (necesario igualmente para la etapa runtime arm64) → `docker/setup-buildx-action` → `docker/build-push-action` con `platforms: linux/amd64,linux/arm64`; BuildKit construye ambas y publica el manifest multi-arch automáticamente. Con el patrón `BUILDPLATFORM` (decisión 1), la emulación QEMU queda limitada a la etapa runtime (el `apk add` en arm64): segundos, no minutos. godot-mcp-docs usa matrix de runners nativos (ubuntu-latest + ubuntu-24.04-arm) + merge manual de manifest; se migra a ese patrón solo si el build crece (CGO, assets).
*Alternativa descartada*: matrix nativo — más rápido, pero ~80 líneas extra de workflow; la evidencia empírica de la comunidad (builds Go multi-arch de 6 min → 1:30 con `BUILDPLATFORM`) muestra que no compensa para Go sin CGO.

**3. Workflows separados: `ci.yml` + `publish.yml`**
`ci.yml` corre en push/PR (setup Go → `go test ./...` → `go vet ./...`), sin tocar imágenes. `publish.yml` corre solo con tags `v*` + `workflow_dispatch` y publica con `push: true`. Separación: el CI es rápido y no consume builds de imagen en cada push; el publish es una acción deliberada.
*Alternativa descartada*: un solo workflow como godot-mcp-docs — cada push buildearía la imagen innecesariamente.

**4. Genericidad vía convención, no templating**
El workflow usa `IMAGE_NAME: ${{ github.repository }}` — se adapta solo al repo donde viva. El único nombre hardcodeado es el binario (`chile-bcn-mcp`) en 3 lugares: `Makefile BINARY`, `Dockerfile COPY`/`ENTRYPOINT`. Se documentan en el README (sección "Cómo reutilizar este scaffold"). Placeholders tipo `{{BINARY}}` se descartan: exigen una herramienta de templating y no aportan nada que un renombrado de 3 líneas no resuelva.

**5. Makefile fusionado**
Sobre los targets actuales (`run-http`, `run-stdio`, `test`, `vet`): `build` pasa a `CGO_ENABLED=0 -ldflags="-s -w"` y se agregan `build-amd64`/`build-arm64` (cross-compile nativo de Go, como godot-mcp-docs), `fmt` y `clean`. El binario sigue en `bin/`.

**6. Tags de publicación**
Tags del workflow: `${{ github.ref_name }}` (p.ej. `1.2.0` de `v1.2.0`) + `latest`. Sin tags por fecha (godot los usa por su ciclo de docs continuo; aquí el release es semver).

**7. docker-compose.yml**
Passthrough de `FASTMCP_TRANSPORT`/`FASTMCP_HOST`/`FASTMCP_PORT`/`FASTMCP_PATH`/`MCP_AUTH_TOKEN` desde env del host (con defaults), `restart: unless-stopped`, healthcheck `curl -f http://localhost:8000/health`. Idéntico al de godot-mcp-docs sin `DOCS_DIR`.

## Risks / Trade-offs

- [QEMU residual en la etapa runtime (`apk add` en arm64)] → Mitigación: el patrón `BUILDPLATFORM` elimina la emulación de la compilación (decisión 1); el `RUN` del runtime es trivial (descarga + unpack de 2 paquetes). Si algún día hay CGO o builds pesados, migrar al matrix nativo de godot (decisión 2).
- [`alpine` + musl vs binario glibc] → Mitigación: `CGO_ENABLED=0` produce binario estático, independiente de la libc — ya validado en godot-mcp-docs.
- [El publish espera un tag `v*`; sin tag no hay imagen] → Mitigación: `workflow_dispatch` permite publicar manualmente (requisito del spec).
- [Primer push a GHCR exige habilitar el paquete en settings del repo] → Mitigación: acción manual única documentada en el README.

## Migration Plan

Cambios aditivos: no afectan al binario ni a los clientes existentes. Rollback trivial (borrar archivos). El `Makefile` cambia su target `build` (flags nuevos) — verificación: `make build && ./bin/chile-bcn-mcp` sigue funcionando.

## Open Questions

<!-- Ninguna: decisiones de registro, estrategia multi-arch y separación de workflows ya resueltas con el usuario. -->
