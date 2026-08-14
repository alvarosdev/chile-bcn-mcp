## Why

Hoy `publish.yml` publica la imagen OCI con el trigger de tags `v*` + dispatch, y no hay distribuciones binarias: un usuario que no quiere Docker no tiene cómo descargar el server. Se redefine el flujo de release completo: **el merge de `release/v*` a `main` es la única vía** — en ese momento se compilan las distribuciones cross-platform (windows/darwin/linux × amd64/arm64), se publica la imagen OCI con el tag de la versión, y se crea un **GitHub Release en draft** con el zip de binarios (el usuario lo publica manualmente). Un PR cerrado **sin merge** no genera nada.

## What Changes

- **Reescribir `.github/workflows/publish.yml`**: trigger `pull_request: closed → main` con gate por job (`merged == true` + `head_ref` = `release/*`); `workflow_dispatch` pide la **versión como input manual**. Se elimina el trigger de tags `v*` (el merge es la única vía — evita el doble disparo al publicar el draft).
- **Nuevo `scripts/build-dist.sh <version>`** (+ target `make dist`): compila la matrix de 6 targets (`linux/amd64`, `linux/arm64`, `darwin/amd64` — Intel pre-Apple Silicon, `darwin/arm64`, `windows/amd64`, `windows/arm64`) con `CGO_ENABLED=0`, arma el **zip único autocontenido** (cada carpeta os/arch lleva su binario + `config/api.resources.yaml`, por la ruta relativa fija del contrato) y genera `SHA256SUMS.txt`. El workflow llama al mismo script (local = CI, testeable).
- **Job docker**: tags de imagen computados de la versión extraída (`1.2.0` + `latest` — el `metadata-action type=version` ya no sirve porque el ref es `main`).
- **Job release**: `gh release create v<version> dist.zip --draft --generate-notes` con idempotencia (tag existente → borrar draft previo y recrear); el usuario convierte el draft en release publicado.
- **Permisos**: `contents: write` para el release.

## Capabilities

### New Capabilities

- `release-distributions`: Distribuciones binarias cross-platform empaquetadas en un zip autocontenido con checksums, y creación del GitHub Release en draft versionado con el tag del release, gatillado exclusivamente por el merge de `release/v*` a `main`.

### Modified Capabilities

- `container-deployment`: MODIFIED del requirement "Imagen OCI multi-arquitectura" — el disparador de publicación cambia de tag `v*` a **merge de release a main** (sin merge no hay imagen); el dispatch manual pide versión.

## Impact

- **Código**: `.github/workflows/publish.yml` (reescrito), `scripts/build-dist.sh` (nuevo), `Makefile` (target `dist`), `README.md` (sección Release).
- **Flujo de operación**: el merge de `release/v*` a main pasa a ser el único disparador de release (cambio de proceso, no de código del server).
- **Sin cambios breaking** en el servidor ni en las tools.
