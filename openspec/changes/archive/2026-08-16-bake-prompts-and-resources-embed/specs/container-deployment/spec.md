## MODIFIED Requirements

### Requirement: Imagen OCI multi-arquitectura
El repositorio DEBE publicar una imagen OCI del servidor en el registro GHCR bajo el nombre del repositorio (`ghcr.io/<owner>/<repo>`), compilada para `linux/amd64` y `linux/arm64` bajo el mismo tag. La publicación DEBE gatillarse únicamente por el merge de una rama `release/v<version>` a `main` (o por dispatch manual con versión): el tag de imagen DEBE ser la versión extraída (`<version>`) más `latest`. Un PR cerrado sin merge NO DEBE publicar imagen. El binario dentro de la imagen DEBE llevar embebido el contrato de endpoints (`internal/config/api.resources.yaml` via `//go:embed`) y los prompts (`internal/prompts/prompts.yaml` via `//go:embed`); la imagen NO DEBE incluir una capa `config/` ni requerir archivos externos para arrancar.

#### Scenario: Publicación por release versionado
- **WHEN** un PR de `release/v1.2.0` a `main` es mergeado
- **THEN** GHCR recibe las imágenes `ghcr.io/<owner>/<repo>:1.2.0` y `ghcr.io/<owner>/<repo>:latest`, ambas con manifest multi-arquitectura que incluye `linux/amd64` y `linux/arm64` y el binario arranca sin `config/` en el filesystem

#### Scenario: Sin merge no hay imagen
- **WHEN** un PR de `release/v1.2.0` a `main` es cerrado sin mergear
- **THEN** no se publica ninguna imagen

#### Scenario: Publicación manual
- **WHEN** un mantenedor dispara el workflow manualmente indicando una versión
- **THEN** las imágenes multi-arquitectura se publican con esa versión y `latest`

#### Scenario: Binario autocontenido en la imagen
- **WHEN** el contenedor se inicia sin volumen `config/` montado y sin `config/api.resources.yaml` en el filesystem
- **THEN** el servidor arranca correctamente usando el contrato embebido en el binario

### Requirement: Contenedor endurecido
Las configuraciones de despliegue provistas (`docker-compose.yml` y los targets podman del Makefile) DEBEN ejecutar el contenedor endurecido: sistema de archivos raíz de solo lectura (con un tmpfs para `/tmp`), capacidades Linux mínimas (`cap_drop: ALL`) y `no-new-privileges`. El servidor NO DEBE necesitar escritura en disco ni archivos de configuración externos para operar.

#### Scenario: Rootfs de solo lectura
- **WHEN** el contenedor se ejecuta con la configuración provista y sin `config/` en el filesystem
- **THEN** el sistema de archivos raíz es read-only y el servidor arranca y responde sin escrituras a disco

#### Scenario: Capacidades mínimas
- **WHEN** el contenedor se ejecuta con la configuración provista
- **THEN** el contenedor corre con `cap_drop: ALL` y `no-new-privileges: true`

#### Scenario: Operación sin config externa
- **WHEN** el contenedor se ejecuta con `readOnlyRootfs: true` y sin montar `config/`
- **THEN** el servidor atiende `GET /health` y `prompts/list` correctamente
