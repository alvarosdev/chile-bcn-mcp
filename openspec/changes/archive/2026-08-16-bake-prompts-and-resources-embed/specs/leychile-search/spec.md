## MODIFIED Requirements

### Requirement: Endpoints declarados en archivo YAML

El servidor DEBE cargar la definición de sus recursos de LeyChile desde el contrato embebido en el binario via `go:embed` (`internal/config/api.resources.yaml`), sin ruta fija en filesystem ni override por variable de entorno, que declara para cada recurso: `url`, `path`, `method`, `timeout`, `retry` (intentos y backoff) y `circuit_breaker` (umbrales de fallo, éxito y cooldown). El contenido DEBE validarse al cargar (en startup via `LoadEmbedded`); una configuración inválida DEBE impedir el arranque del servidor. El binario es autocontenido y no requiere `config/` en el filesystem ni en la imagen de contenedor.

#### Scenario: Carga exitosa con ruta fija
- **WHEN** el servidor arranca (el contrato está embebido via `go:embed` en `internal/config/api.resources.yaml`, sin `config/api.resources.yaml` en el filesystem)
- **THEN** los recursos quedan disponibles y el servidor continúa el arranque
#### Scenario: Configuración inválida
- **WHEN** el YAML embebido tiene un recurso sin `path`, con timeout negativo o con retry/breaker incoherentes
- **THEN** el servidor no arranca e informa el error de validación

#### Scenario: Configuración en la imagen de contenedor
- **WHEN** se construye la imagen de contenedor desde el Dockerfile
- **THEN** la imagen no necesita incluir la carpeta `config/` y el servidor arranca sin configuración adicional (contrato embebido)
