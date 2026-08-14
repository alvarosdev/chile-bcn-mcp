## Purpose

Despliegue del servidor chile-bcn-mcp en contenedores: imagen OCI multi-arquitectura publicada en GHCR, ejecutada como usuario no-root, configurable por entorno y con healthcheck, más validación automática de los cambios del repositorio.

## ADDED Requirements

### Requirement: Imagen OCI multi-arquitectura

El repositorio DEBE publicar una imagen OCI del servidor en el registro GHCR bajo el nombre del repositorio (`ghcr.io/<owner>/<repo>`), compilada para `linux/amd64` y `linux/arm64` bajo el mismo tag. El tag `latest` DEBE existir y cada release versionado (`vX.Y.Z`) DEBE publicar además un tag con la versión.

#### Scenario: Publicación por release versionado
- **WHEN** se publica un tag `v1.2.0` en el repositorio
- **THEN** GHCR recibe las imágenes `ghcr.io/<owner>/<repo>:1.2.0` y `ghcr.io/<owner>/<repo>:latest`, ambas con manifest multi-arquitectura que incluye `linux/amd64` y `linux/arm64`

#### Scenario: Publicación manual
- **WHEN** un mantenedor dispara el workflow de publicación manualmente
- **THEN** las imágenes multi-arquitectura se publican con tag `latest`

### Requirement: Ejecución como usuario no-root

El proceso del servidor dentro del contenedor DEBE ejecutarse con un usuario sin privilegios, no con root.

#### Scenario: Verificación de usuario en contenedor
- **WHEN** se ejecuta un comando dentro del contenedor en ejecución
- **THEN** el usuario efectivo no es `root`

### Requirement: Configuración por entorno en el contenedor

El contenedor DEBE respetar las mismas variables de entorno que el binario nativo (`FASTMCP_TRANSPORT`, `FASTMCP_HOST`, `FASTMCP_PORT`, `FASTMCP_PATH`, `MCP_AUTH_TOKEN`). En el contenedor, el host de escucha por defecto DEBE ser `0.0.0.0` para permitir el acceso desde fuera del contenedor, y el puerto expuesto DEBE ser `8000`.

#### Scenario: Defaults de contenedor
- **WHEN** el contenedor se inicia sin variables de entorno
- **THEN** el servidor escucha en `0.0.0.0:8000` y responde en `/mcp`

#### Scenario: Override de entorno
- **WHEN** el contenedor se inicia con `FASTMCP_PORT=9000` y `MCP_AUTH_TOKEN=secret`
- **THEN** el servidor escucha en el puerto 9000 y exige el token en las requests HTTP

### Requirement: Healthcheck del contenedor

El contenedor DEBE poder ser verificado por su endpoint de salud: el proceso DEBE responder `GET /health` con estado healthy y el contenedor DEBE incluir las herramientas necesarias para que un healthcheck por HTTP funcione sin herramientas externas.

#### Scenario: Healthcheck del contenedor
- **WHEN** se consulta `GET /health` en el puerto del contenedor
- **THEN** responde `200 OK` con estado `healthy`

### Requirement: Validación automática en CI

Todo push y pull request al repositorio DEBE ejecutar automáticamente la compilación y los tests del proyecto; si algún test o análisis estático falla, el estado del workflow DEBE ser de fallo y bloquear la integración.

#### Scenario: Cambios válidos
- **WHEN** un push o PR contiene cambios que compilan y pasan todos los tests
- **THEN** el workflow de CI finaliza con éxito

#### Scenario: Cambios que rompen tests
- **WHEN** un push o PR introduce un cambio que hace fallar algún test o el análisis estático
- **THEN** el workflow de CI finaliza en fallo
