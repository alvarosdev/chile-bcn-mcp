## MODIFIED Requirements

### Requirement: Imagen OCI multi-arquitectura

El repositorio DEBE publicar una imagen OCI del servidor en el registro GHCR bajo el nombre del repositorio (`ghcr.io/<owner>/<repo>`), compilada para `linux/amd64` y `linux/arm64` bajo el mismo tag. La publicación DEBE gatillarse únicamente por el merge de una rama `release/v<version>` a `main` (o por dispatch manual con versión): el tag de imagen DEBE ser la versión extraída (`<version>`) más `latest`. Un PR cerrado sin merge NO DEBE publicar imagen. (Cambio: antes el disparador era el push de un tag `v*` — se elimina para que el merge sea la única vía y no haya doble disparo al publicar el release.)

#### Scenario: Publicación por release versionado
- **WHEN** un PR de `release/v1.2.0` a `main` es mergeado
- **THEN** GHCR recibe las imágenes `ghcr.io/<owner>/<repo>:1.2.0` y `ghcr.io/<owner>/<repo>:latest`, ambas con manifest multi-arquitectura que incluye `linux/amd64` y `linux/arm64`

#### Scenario: Sin merge no hay imagen
- **WHEN** un PR de `release/v1.2.0` a `main` es cerrado sin mergear
- **THEN** no se publica ninguna imagen

#### Scenario: Publicación manual
- **WHEN** un mantenedor dispara el workflow manualmente indicando una versión
- **THEN** las imágenes multi-arquitectura se publican con esa versión y `latest`
