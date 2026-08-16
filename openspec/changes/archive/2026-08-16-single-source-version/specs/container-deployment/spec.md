## ADDED Requirements

### Requirement: Versión inyectada en build de imagen

El `Dockerfile` SHALL aceptar `ARG VERSION` y compilar el binario con `-ldflags "-X github.com/alvarosdev/chile-bcn-mcp/internal/version.Version=${VERSION}"`. El workflow de publish SHALL pasar `VERSION=${{ needs.version.outputs.version }}` como `build-args` a `docker/build-push-action`. Cuando `VERSION` no se provee, el binario SHALL reportar `"dev"`.

#### Scenario: Build local de imagen usa VERSION
- **WHEN** se ejecuta `podman build --build-arg VERSION=$(cat VERSION | sed 's/^v//') .`
- **THEN** el contenedor resultante reporta esa versión en `initialize`

#### Scenario: Build CI publica versión correcta
- **WHEN** el workflow publica `linux/amd64,linux/arm64` con `VERSION=1.2.0`
- **THEN** ambas arquitecturas reportan `1.2.0`

#### Scenario: Build sin VERSION
- **WHEN** se ejecuta `podman build .` sin `build-arg`
- **THEN** el binario reporta `dev`
