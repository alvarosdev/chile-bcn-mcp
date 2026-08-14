## Purpose

Flujo de release del proyecto: el merge de una rama `release/v*` a `main` compila distribuciones binarias cross-platform en un zip autocontenido con checksums, publica la imagen OCI versionada y crea un GitHub Release en draft con el tag de la versión, que el mantenedor publica manualmente.

## Requirements

### Requirement: Gate de release por merge

El flujo de release DEBE ejecutarse únicamente cuando un PR hacia `main` cuya rama origen sea `release/v*` es **mergeado**. Un PR cerrado sin merge NO DEBE generar ninguna imagen ni distribución. Un PR mergeado desde una rama que no sea `release/*` NO DEBE generar ninguna imagen ni distribución.

#### Scenario: Merge de release
- **WHEN** un PR de `release/v1.2.0` a `main` es mergeado
- **THEN** se compilan las distribuciones, se publica la imagen OCI y se crea el release draft con tag `v1.2.0`

#### Scenario: Cerrado sin merge
- **WHEN** un PR de `release/v1.2.0` a `main` es cerrado sin mergear
- **THEN** no se genera imagen, ni distribución, ni release

#### Scenario: Merge desde otra rama
- **WHEN** un PR desde una rama que no empieza con `release/` es mergeado a `main`
- **THEN** no se genera imagen, ni distribución, ni release

### Requirement: Distribuciones cross-platform en zip autocontenido

El flujo DEBE compilar el binario para seis objetivos — `linux/amd64`, `linux/arm64`, `darwin/amd64` (Intel pre-Apple Silicon), `darwin/arm64`, `windows/amd64` y `windows/arm64` — y empaquetarlos en UN zip donde cada carpeta por sistema operativo y arquitectura contenga el binario (con extensión `.exe` en Windows) y el archivo `config/api.resources.yaml` (el contrato de endpoints se carga por ruta relativa fija). El zip DEBE incluir un `SHA256SUMS.txt` con los checksums de todos los binarios.

#### Scenario: Zip completo
- **WHEN** el flujo de release corre para una versión
- **THEN** el zip contiene las seis carpetas os/arch, cada una con su binario y su `config/api.resources.yaml`, y el `SHA256SUMS.txt` cubre los seis binarios

#### Scenario: Distribución ejecutable tras extraer
- **WHEN** un usuario extrae una carpeta (p.ej. `linux/amd64/`) y ejecuta el binario desde esa carpeta
- **THEN** el server arranca usando el `config/api.resources.yaml` incluido en la misma carpeta

### Requirement: Release draft versionado

El flujo DEBE crear un GitHub Release en estado **draft** con el tag `v<version>` (la versión extraída de la rama `release/v<version>`) y el zip como asset, sin publicarlo — la publicación la hace el mantenedor manualmente. La versión DEBE derivarse del nombre de la rama; el dispatch manual DEBE pedir la versión como input.

#### Scenario: Draft creado
- **WHEN** un PR de `release/v1.2.0` es mergeado
- **THEN** existe un release draft con tag `v1.2.0` y el zip adjunto, no publicado

#### Scenario: Tag preexistente
- **WHEN** el tag `v1.2.0` ya tiene un release draft (re-ejecución)
- **THEN** el draft previo se reemplaza con el nuevo (sin fallar el flujo)

#### Scenario: Dispatch manual
- **WHEN** el workflow se dispara manualmente
- **THEN** solicita la versión como input y la usa en lugar del nombre de rama

#### Scenario: Release solo con pipeline verde
- **WHEN** el job de distribuciones o el job de la imagen OCI fallan
- **THEN** NO se crea el release draft (el job de release depende de ambos)
