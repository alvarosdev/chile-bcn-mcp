## ADDED Requirements

### Requirement: Versión de distribuciones derivada del SSOT

El flujo de release SHALL derivar la versión de las distribuciones cross-platform (`dist.zip`) y de la imagen OCI exclusivamente del SSOT (`VERSION` / rama `release/v*` / input `workflow_dispatch`), inyectada por `ldflags` en cada binario compilado. Ningún binario dentro de `dist.zip` SHALL contener una versión distinta al tag del release.

#### Scenario: Distribuciones embeben versión correcta
- **WHEN** el job `dist` compila con `scripts/build-dist.sh 1.2.0`
- **THEN** cada uno de los seis binarios en `dist.zip` reporta `1.2.0` vía `initialize` y `SHA256SUMS.txt` corresponde a esos binarios versionados

#### Scenario: Tag OCI coincide con versión del binario
- **WHEN** el job `docker` publica `ghcr.io/owner/repo:1.2.0`
- **THEN** el binario dentro de esa imagen reporta `1.2.0` (no `0.1.0` stale ni `dev`)
