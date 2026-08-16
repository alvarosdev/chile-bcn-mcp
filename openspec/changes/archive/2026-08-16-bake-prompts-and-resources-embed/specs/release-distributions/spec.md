## MODIFIED Requirements

### Requirement: Distribuciones cross-platform en zip autocontenido
El flujo DEBE compilar el binario para seis objetivos — `linux/amd64`, `linux/arm64`, `darwin/amd64` (Intel pre-Apple Silicon), `darwin/arm64`, `windows/amd64` y `windows/arm64` — y empaquetarlos en UN zip donde cada carpeta por sistema operativo y arquitectura contenga **únicamente el binario** (con extensión `.exe` en Windows) — el contrato de endpoints y los prompts van embebidos en el binario via `//go:embed` (`internal/config/api.resources.yaml` y `internal/prompts/prompts.yaml`), no como archivos separados. El zip DEBE incluir un `SHA256SUMS.txt` con los checksums de todos los binarios.

#### Scenario: Zip completo
- **WHEN** el flujo de release corre para una versión
- **THEN** el zip contiene las seis carpetas os/arch, cada una con solo su binario, y el `SHA256SUMS.txt` cubre los seis binarios; ninguna carpeta contiene `config/api.resources.yaml`

#### Scenario: Distribución ejecutable tras extraer
- **WHEN** un usuario extrae una carpeta (p.ej. `linux/amd64/`) y mueve el binario a otra carpeta sin `config/` y lo ejecuta
- **THEN** el server arranca correctamente usando el contrato y prompts embebidos, sin requerir `config/api.resources.yaml` en el filesystem

#### Scenario: Zip sin config redundante
- **WHEN** se inspecciona `dist.zip` tras `scripts/build-dist.sh`
- **THEN** no existe ningún archivo `config/api.resources.yaml` dentro del zip

### Requirement: Versión de distribuciones derivada del SSOT
El flujo de release SHALL derivar la versión de las distribuciones cross-platform (`dist.zip`) y de la imagen OCI exclusivamente del SSOT (`VERSION` / rama `release/v*` / input `workflow_dispatch`), inyectada por `ldflags` en cada binario compilado. Ningún binario dentro de `dist.zip` SHALL contener una versión distinta al tag del release.

#### Scenario: Distribuciones embeben versión correcta
- **WHEN** el job `dist` compila con `scripts/build-dist.sh 1.2.0`
- **THEN** cada uno de los seis binarios en `dist.zip` reporta `1.2.0` vía `initialize` y `SHA256SUMS.txt` corresponde a esos binarios versionados

#### Scenario: Tag OCI coincide con versión del binario
- **WHEN** el job `docker` publica `ghcr.io/owner/repo:1.2.0`
- **THEN** el binario dentro de esa imagen reporta `1.2.0` (no `0.1.0` stale ni `dev`)
