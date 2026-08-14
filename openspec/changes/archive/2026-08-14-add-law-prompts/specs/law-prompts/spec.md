## Purpose

Prompts MCP curados que codifican los flujos correctos de uso de las tools de LeyChile (buscar, resumir, leer, verificar, comparar, rastrear) con argumentos tipados, de modo que el cliente descubre y aplica la expertise del dominio sin instrucciones externas.

## ADDED Requirements

### Requirement: Prompts curados expuestos

El servidor DEBE exponer los seis prompts curados vía `prompts/list`: `analyze_law` (argumentos `norm_id` requerido y `aspect` opcional), `search_legal_topic` (`topic` requerido), `compare_law_versions` (`norm_id`, `from_date` y `to_date` requeridos), `trace_law_history` (`norm_id` requerido), `check_law_validity` (`norm_id` requerido y `date` opcional) y `explain_law_simply` (`norm_id` requerido y `audience` opcional). Cada prompt DEBE declarar su descripción y la obligatoriedad de sus argumentos.

#### Scenario: Lista de prompts
- **WHEN** un cliente MCP consulta `prompts/list`
- **THEN** la respuesta incluye los seis prompts con sus descripciones y argumentos (con `required` donde corresponde)

### Requirement: Templates que guían el uso de las tools

Cada prompt DEBE devolver un template de mensaje que (a) inyecta los valores de los argumentos recibidos, (b) referencia las tools reales por su nombre (`search_laws`, `get_law`, `get_law_summary`, `get_law_history`), y (c) embebe la regla de dominio: verificar las afirmaciones contra el texto real de la norma y no inventar números de artículo ni contenido.

#### Scenario: Prompt con argumentos inyectados
- **WHEN** un cliente pide `prompts/get` para `analyze_law` con `norm_id: 1195666`
- **THEN** el mensaje devuelto incluye el `norm_id` en el texto y las instrucciones de usar `get_law_summary` y `get_law` en ese orden

#### Scenario: Prompt con fechas de comparación
- **WHEN** un cliente pide `prompts/get` para `compare_law_versions` con `from_date` y `to_date`
- **THEN** el mensaje instruye usar `get_law` con `version_date` para ambas fechas e incluye ambas fechas en el texto

### Requirement: Disclaimer en explicaciones simples

El prompt `explain_law_simply` DEBE incluir en su template la instrucción de citar el artículo fuente de cada afirmación y de advertir que la explicación no constituye asesoría legal.

#### Scenario: Prompt de explicación con disclaimer
- **WHEN** un cliente pide `prompts/get` para `explain_law_simply`
- **THEN** el mensaje incluye la instrucción de citar fuentes y el disclaimer de no-asesoría legal

### Requirement: Prompts sin llamadas externas

Servir un prompt DEBE ser una operación pura de templates: `prompts/get` NO DEBE realizar llamadas a la API de BCN ni depender del estado del cliente de normas.

#### Scenario: Prompt servido sin red
- **WHEN** un cliente pide `prompts/get` con la API de BCN inaccesible
- **THEN** el prompt se sirve igualmente, sin error ni latencia de red
