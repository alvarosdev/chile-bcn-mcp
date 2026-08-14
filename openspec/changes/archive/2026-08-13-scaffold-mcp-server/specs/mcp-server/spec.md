## Purpose

Capacidad base del servidor MCP de chile-bcn-mcp: expone las mismas tools a través de transporte stdio o streamable HTTP, configurable por entorno, con autenticación opcional, health check y apagado limpio.

## ADDED Requirements

### Requirement: Transportes conmutables

El servidor DEBE exponer el mismo conjunto de tools registradas a través de transporte **stdio** (stdin/stdout) o **streamable HTTP** (endpoint `/mcp`), según el valor de la variable de entorno `FASTMCP_TRANSPORT`. El valor `stdio` activa el transporte stdio; cualquier otro valor (por defecto `http`) activa streamable HTTP. El cambio de transporte no DEBE alterar las tools disponibles.

#### Scenario: Arranque con transporte stdio
- **WHEN** el servidor se inicia con `FASTMCP_TRANSPORT=stdio`
- **THEN** el servidor atiende el protocolo MCP por stdin/stdout y no abre puertos de red

#### Scenario: Arranque con transporte HTTP
- **WHEN** el servidor se inicia con `FASTMCP_TRANSPORT=http`
- **THEN** el servidor escucha en `FASTMCP_HOST:FASTMCP_PORT` (por defecto `127.0.0.1:8000`) y atiende MCP en la ruta configurada (por defecto `/mcp`)

#### Scenario: Mismas tools en ambos transportes
- **WHEN** un cliente MCP consulta `tools/list` por stdio y otro por HTTP
- **THEN** ambos obtienen exactamente la misma lista de tools registradas

### Requirement: Configuración por variables de entorno

El servidor DEBE leer toda su configuración de variables de entorno al arrancar: `FASTMCP_TRANSPORT` (transporte), `FASTMCP_HOST` (host de escucha), `FASTMCP_PORT` (puerto), `FASTMCP_PATH` (ruta HTTP de MCP, por defecto `/mcp`) y `MCP_AUTH_TOKEN` (token de autenticación). Cada variable DEBE tener un valor por defecto documentado y no DEBE requerir archivos de configuración.

#### Scenario: Valores por defecto
- **WHEN** el servidor se inicia sin variables de entorno definidas
- **THEN** usa `http` como transporte, `127.0.0.1:8000` como dirección, `/mcp` como ruta y no requiere autenticación

#### Scenario: Override explícito
- **WHEN** el operador define `FASTMCP_TRANSPORT=stdio` y `FASTMCP_PORT=9000`
- **THEN** el servidor usa esos valores en lugar de los por defecto

### Requirement: Autenticación opcional por token Bearer

Cuando `MCP_AUTH_TOKEN` está definida, el servidor DEBE rechazar las requests HTTP al endpoint MCP que no incluyan un header `Authorization: Bearer <token>` con el token exacto, y DEBE aceptar las que sí lo incluyan. Cuando `MCP_AUTH_TOKEN` no está definida, el servidor DEBE aceptar requests sin autenticación. La autenticación NO aplica al transporte stdio.

#### Scenario: Token configurado y request válida
- **WHEN** `MCP_AUTH_TOKEN=secret` está definida y un cliente envía una request MCP con `Authorization: Bearer secret`
- **THEN** la request se procesa normalmente

#### Scenario: Token configurado y request sin token
- **WHEN** `MCP_AUTH_TOKEN=secret` está definida y un cliente envía una request MCP sin header de autorización o con token incorrecto
- **THEN** el servidor responde con error de autenticación (HTTP 401) y no procesa la request

#### Scenario: Sin token configurado
- **WHEN** `MCP_AUTH_TOKEN` no está definida
- **THEN** el servidor procesa requests MCP HTTP sin exigir autorización

### Requirement: Health check

El servidor DEBE exponer un endpoint HTTP `GET /health` (en modo HTTP) que responda `200 OK` con un cuerpo JSON que incluya un estado saludable, sin requerir autenticación.

#### Scenario: Consulta de salud
- **WHEN** un cliente hace `GET /health`
- **THEN** recibe `200 OK` con cuerpo JSON que indica estado `healthy`, aun cuando `MCP_AUTH_TOKEN` esté definida

### Requirement: Tool demo echo

El servidor DEBE registrar una tool llamada `echo` que recibe un argumento de texto y devuelve exactamente ese texto como contenido. Esta tool sirve de demostración del registro de tools y es reemplazable por las tools de dominio reales.

#### Scenario: Llamada a echo
- **WHEN** un cliente MCP llama a la tool `echo` con `{"text": "hola"}`
- **THEN** la respuesta contiene el texto `hola`

#### Scenario: echo aparece en tools/list
- **WHEN** un cliente MCP consulta `tools/list`
- **THEN** la lista incluye la tool `echo` con su descripción

### Requirement: Apagado limpio

El servidor DEBE terminar limpiamente al recibir las señales `SIGINT` o `SIGTERM`: en modo HTTP, DEJA de aceptar conexiones nuevas, cierra las activas con un tiempo límite y sale con código de salida 0.

#### Scenario: Señal de terminación
- **WHEN** el proceso recibe `SIGINT` o `SIGTERM` mientras atiende HTTP
- **THEN** el servidor cierra el listener y finaliza sin errores

#### Scenario: Terminación en modo stdio
- **WHEN** el proceso recibe `SIGINT` o `SIGTERM` en modo stdio
- **THEN** el servidor cierra la sesión y finaliza sin errores
