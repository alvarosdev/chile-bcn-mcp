## 1. Script de distribuciones (local = CI)

- [x] 1.1 Crear `scripts/build-dist.sh <version>`: matrix de 6 targets (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, `windows/arm64`) con `CGO_ENABLED=0`, salida `dist/<os>/<arch>/` (sufijo `.exe` en Windows), validación temprana de `zip`/`sha256sum`
- [x] 1.2 Copiar `config/api.resources.yaml` a cada `dist/<os>/<arch>/config/` (dist autocontenida — ruta relativa fija del server)
- [x] 1.3 Generar `dist/SHA256SUMS.txt` (checksums de los 6 binarios) y empaquetar `dist.zip` con `zip -r` incluyendo los checksums
- [x] 1.4 Agregar target `dist` al Makefile (`.PHONY` + help) que llame `bash scripts/build-dist.sh 0.0.0-local`
- [x] 1.5 Probar localmente: `make dist` → verificar estructura del zip (6 carpetas con binario + config, SHA256SUMS) y ejecutar `dist/linux/amd64/chile-bcn-mcp` desde su carpeta (health OK con su config embebida)

## 2. Workflow publish.yml

- [x] 2.1 Reescribir triggers: `pull_request: types: [closed], branches: [main]` + `workflow_dispatch` con input `version` requerido; ELIMINAR `push: tags: v*`
- [x] 2.2 Extraer versión: `VERSION=${GITHUB_HEAD_REF#release/}` normalizado (quitar `v` si viene) → `TAG=v$VERSION`; en dispatch usar `inputs.version`; pasar VERSION entre jobs como output/env
- [x] 2.3 Job `dist`: setup-go (1.26) → `bash scripts/build-dist.sh $VERSION` → `upload-artifact` del zip
- [x] 2.4 Job `docker`: mismo flujo actual (QEMU + buildx + login GHCR) pero con tags raw `$VERSION` + `latest` (reemplazar `metadata-action type=version`)
- [x] 2.5 Job `release`: descargar el artifact → `gh release create v$VERSION dist.zip --draft --generate-notes`; idempotencia: `gh release view` → si existe draft, borrarlo y recrear; si existe publicado → fallar con mensaje
- [x] 2.6 Gate en los 3 jobs: `if: github.event.pull_request.merged == true && startsWith(github.head_ref, 'release/')` (dispatch corre siempre); permisos `contents: write`

## 3. Verificación y documentación

- [x] 3.1 `make check` en verde (sin cambios de código del server)
- [x] 3.2 Validar sintaxis YAML del workflow (revisión manual + estructura de jobs/ifs)
- [x] 3.3 Actualizar `README.md`: sección Release (flujo release/v* → main, draft manual, zip con estructura os/arch, SHA256SUMS, dispatch con versión) y nota de que el trigger de tags `v*` ya no publica
- [ ] 3.4 Primer release real: mergear PR `release/v0.1.0` → verificar draft con zip + imagen GHCR `0.1.0`/`latest` → publicar manualmente
- [x] 1.6 Retirar los targets redundantes build-amd64/build-arm64 del Makefile (superados por make dist: mismos binarios + config embebida + checksums + 4 targets más) y actualizar README/CLAUDE.md
- [x] 2.7 Job release: agregar `docker` a `needs` (release solo con pipeline verde — decisión del usuario)
- [x] 2.8 Job backport: PR main → develop tras cada merge de release, con idempotencia (`gh pr list` abierto) y tolerancia a "no commits between"; permisos `pull-requests: write`
