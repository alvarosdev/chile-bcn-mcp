## Context

`publish.yml` hoy dispara con tags `v*` + dispatch y publica la imagen OCI con `metadata-action type=version` (que depende del ref de tag). No existen distribuciones binarias. El usuario definió el flujo completo: merge de `release/v*` a `main` como ÚNICA vía (cerrado sin merge = nada; sin trigger de tags), zip autocontenido con `config/api.resources.yaml` por carpeta os/arch, SHA256SUMS, release draft con tag `v<version>` que el mantenedor publica manualmente, y dispatch manual que pide versión. Ver proposal.md para la motivación; requisitos en specs/release-distributions y specs/container-deployment.

## Goals / Non-Goals

**Goals:**
- Un solo flujo de release, determinista y localmente testeable (el script de dist corre igual en local y en CI).
- Distribuciones ejecutables tras extraer (config embebida por carpeta, ruta relativa fija del server intacta).

**Non-Goals:**
- Firma de binarios / notarización de macOS (requieren certificados — futuro).
- Auto-publicar el release (draft siempre; lo publica el mantenedor).
- Versionado automático de go.mod o changelog.

## Decisions

**1. Trigger con gate por job (no hay evento nativo de "merge desde X")**
`on: pull_request: types: [closed], branches: [main]` + en cada job `if: github.event.pull_request.merged == true && startsWith(github.head_ref, 'release/')`. El dispatch (`workflow_dispatch`) lleva input `version` requerido y corre el flujo completo con esa versión. Se ELIMINA el trigger `push: tags: v*`: publicar el draft crea el tag → re-dispararía el flujo (duplicado). `head_ref` solo existe en eventos de PR — el input del dispatch cubre el caso manual.

**2. Extracción de versión**
Merge: `VERSION=${GITHUB_HEAD_REF#release/}` (normalizar: quitar `v` si viene `release/v1.2.0`), `TAG="v$VERSION"`. Dispatch: `VERSION=${{ inputs.version }}`. Expresión única al inicio del job, pasada como output/env entre jobs.

**3. `scripts/build-dist.sh` — el script es la fuente de verdad (local = CI)**
`scripts/build-dist.sh <version>`: compila los 6 targets (`CGO_ENABLED=0 GOOS/GOARCH`, matrix con `darwin/amd64` para Intel pre-Apple Silicon — `x64` es el mismo ISA con otro nombre, Go usa `amd64`), copia `config/api.resources.yaml` a cada `dist/<os>/<arch>/config/`, genera `SHA256SUMS.txt` (`sha256sum`), y empaqueta `dist.zip` con `zip -r` (incluyendo los checksums). El workflow llama el script; local `make dist` lo corre igual — testeable hoy. Sufijo `.exe` solo en Windows.

**4. Docker tags computados (no metadata-action type=version)**
Con ref `main`, `type=version` no resuelve — tags raw: `${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:$VERSION` + `:latest`. El resto del job (QEMU + buildx + BUILDPLATFORM) no cambia.

**5. Release draft con idempotencia**
`gh release create "v$VERSION" dist.zip --draft --generate-notes --repo ${{ github.repository }}`. Idempotencia: `gh release view "v$VERSION"` → si existe un draft, `gh release delete` (solo borra si es draft) y recrear; si existe publicado → fallar con mensaje claro (no sobreescribir un release publicado). Permisos: `contents: write`.

**6. Estructura del zip**
`windows/{amd64,arm64}/chile-bcn-mcp.exe`, `linux/{amd64,arm64}/chile-bcn-mcp`, `darwin/{amd64,arm64}/chile-bcn-mcp`, cada carpeta con `config/api.resources.yaml`, + `SHA256SUMS.txt` en la raíz. El yaml duplicado ×6 es ~700 bytes — el costo de que cada dist sea ejecutable tras extraer sin cambiar la ruta fija del server.

**7. Release solo con pipeline verde (decisión del usuario)**
`release` pasa a `needs: [version, dist, docker]` — cualquier fallo en dist o docker bloquea la creación del draft. El invariante "imagen solo si el build es correcto" ya lo garantiza `build-push-action` (el push ocurre dentro del build; build fallido = sin push) y quedó completo al quitar `cache-to` (única fase post-push que podía fallar).