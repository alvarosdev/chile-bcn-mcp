## MODIFIED Requirements

### Requirement: Prompts curados expuestos

El servidor DEBE exponer los siete prompts curados vía `prompts/list`: `analyze_law` (argumentos `norm_id` requerido y `aspect` opcional), `search_legal_topic` (`topic` requerido), `compare_law_versions` (`norm_id`, `from_date` y `to_date` requeridos), `trace_law_history` (`norm_id` requerido), `check_law_validity` (`norm_id` requerido y `date` opcional), `explain_law_simply` (`norm_id` requerido y `audience` opcional) y `law_research_workflow` (`norm_id` requerido y `question` opcional). Cada prompt DEBE declarar su descripción y la obligatoriedad de sus argumentos.

#### Scenario: Lista de prompts
- **WHEN** un cliente MCP consulta `prompts/list`
- **THEN** la respuesta incluye los siete prompts con sus descripciones y argumentos (con `required` donde corresponde)

## ADDED Requirements

### Requirement: Prompt de flujo de investigación

El prompt `law_research_workflow` DEBE devolver un template que instruye al cliente a seguir el flujo económico de lectura de una norma: obtener primero el resumen con la estructura (`get_law_summary`), y recién entonces leer las secciones de interés (`get_law` con `section_id`), evitando descargar el contenido completo de normas largas. El template DEBE indicar que el `section_id` se obtiene de la estructura y DEBE reusar la regla de dominio de no inventar números de artículo ni contenido.

#### Scenario: Template del flujo
- **WHEN** un cliente pide `prompts/get` para `law_research_workflow` con `norm_id: 1195666`
- **THEN** el mensaje incluye el `norm_id` y las instrucciones de usar `get_law_summary` y luego `get_law` con `section_id`, en ese orden

#### Scenario: Prompt servido sin red
- **WHEN** un cliente pide `prompts/get` para `law_research_workflow` con la API de BCN inaccesible
- **THEN** el prompt se sirve sin llamadas a la API de BCN, sin error ni latencia de red
