## ADDED Requirements

### Requirement: Prompt de pregunta constitucional

El servidor DEBE (SHALL) exponer el prompt `answer_constitutional_question` vía `prompts/list` con argumentos `question` (requerido), `article_hint` (opcional, string libre con ejemplos "19", "19 Nº24", "19 inc 2", "93", "transitoria primera") y `version_date` (opcional, formato YYYY-MM-DD). El template DEBE (SHALL) instruir el flujo económico sobre el Decreto 100 (Constitución Política de la República, norm_id hardcodeado 242302): primero `get_law_summary` con `version_date` si se entrega para obtener Size y TOC con `section_id`, luego `get_law` con `section_id` solo para las 1-3 secciones relevantes identificadas desde el TOC (por `article_hint` si existe — match laxo contra `Estructura[].n` — o mapeando `question` a nombres de capítulos/artículos y a DISPOSICIONES TRANSITORIAS cuando la pregunta mencione reforma/plebiscito/vigencia/transitoria), NUNCA `get_law` sin `section_id` salvo que Size indique norma corta. El template DEBE (SHALL) exigir citar el artículo fuente de cada afirmación, verificar contra el texto retornado sin inventar artículos, y usar lenguaje condicional para interpretaciones. DEBE (SHALL) incluir disclaimer que la respuesta es información general, no asesoría legal, que la calificación vinculante corresponde al Tribunal Constitucional y que se debe verificar en bcn.cl y consultar profesional habilitado. El template DEBE (SHALL) permitir veredicto textual "constitucional/inconstitucional" solo bajo hedge ("conforme al texto retornado... podría interpretarse como...").
#### Scenario: Pregunta general sobre la CPR vigente
- **WHEN** un cliente pide `prompts/get` para `answer_constitutional_question` con `question: "¿qué dice sobre el derecho de propiedad?"`
- **THEN** el mensaje incluye la pregunta, instruye `get_law_summary(norm_id=<Decreto100>)` y luego `get_law(norm_id=<Decreto100>, section_id=<id>)` para secciones del TOC, exige citas por artículo y disclaimer de no-asesoría

#### Scenario: Pregunta con pista de artículo y fecha histórica
- **WHEN** un cliente pide `prompts/get` para `answer_constitutional_question` con `question: "¿qué decía?"`, `article_hint: "19 Nº24"` y `version_date: "2019-01-01"`
- **THEN** el mensaje incluye `article_hint` y `version_date`, instruye `get_law_summary` y `get_law` con `version_date=2019-01-01`, e indica que el header "Version: as of" confirma la versión leída

#### Scenario: Sin llamadas externas
- **WHEN** un cliente pide `prompts/get` para `answer_constitutional_question` con la API BCN inaccesible
- **THEN** el prompt se sirve igualmente sin error ni latencia de red (template puro)

### Requirement: Prompt de contraste de constitucionalidad

El servidor DEBE (SHALL) exponer el prompt `check_norm_constitutionality` vía `prompts/list` con argumentos `norm_id` (requerido, norma a contrastar), `question` (opcional, foco del contraste) y `version_date` (opcional, compartido para ambas normas para coherencia temporal). El template DEBE (SHALL) instruir: (1) `get_law_summary` para la norma indicada y para el Decreto 100 (242302) — ambas con `version_date` si se entrega, (2) selección de secciones relevantes de ambas normas desde sus TOC usando `question` o el resumen de la norma (incluyendo DISPOSICIONES TRANSITORIAS cuando aplique), (3) `get_law` con `section_id` para cada sección relevante de ambas normas (evitando descarga completa salvo Size corto), (4) presentación en paralelismo textual "Art. X Ley Y dice ... | Art. Z CPR dice ..." con análisis de compatibilidad textual, citando artículos de ambas. DEBE (SHALL) aplicar la misma regla de no inventar contenido, hedge para veredictos y disclaimer reforzado de no-asesoría y competencia del TC. DEBE (SHALL) recomendar como paso recomendado (SHOULD) `get_law_history(norm_id)` cuando la norma pudo ser objeto de control del TC — si existe grupo modificatorias, resumirlo.
#### Scenario: Contraste vigente sin pregunta enfocada
- **WHEN** un cliente pide `prompts/get` para `check_norm_constitutionality` con `norm_id: 242302` y `question` ausente
- **THEN** el mensaje incluye `norm_id`, instruye `get_law_summary` para ambas normas, selección de secciones y `get_law` con `section_id` en ambas, y exige paralelismo con citas y disclaimer

#### Scenario: Contraste histórico enfocado
- **WHEN** un cliente pide `prompts/get` para `check_norm_constitutionality` con `norm_id: 1195666`, `question: "¿vulnera igualdad ante la ley?"` y `version_date: "2020-06-01"`
- **THEN** el mensaje incluye los tres argumentos, instruye ambas lecturas con `version_date=2020-06-01` y análisis focalizado en Art. 19 Nº2 CPR, con hedge y disclaimer

### Requirement: Decreto 100 como fuente canónica y flujo section_id

Ambos prompts DEBEN (SHALL) referenciar el Decreto 100 como fuente canónica de la CPR vía `norm_id` hardcodeado 242302 (estable históricamente; verificado contra `search_laws` con query "Decreto 100 Constitución Política") y DEBEN (SHALL) documentar el flujo `get_law_summary → section_id → get_law(section_id)` como patrón obligatorio para normas extensas. El template DEBE (SHALL) advertir que `get_law` sin `section_id` solo se permite si `Size` indica norma corta, y DEBE (SHALL) indicar que `section_id` proviene de `Estructura[].id` del summary. El código ya cubre transitorias como `EstructuraPart` con `t:1` (contenedor) y `t:6` (artículos) — el prompt DEBE (SHALL) instruir al LLM a considerar DISPOSICIONES TRANSITORIAS como sección elegible cuando la pregunta mencione reforma/plebiscito/vigencia/transitoria. No se requiere snapshot local ni nuevo tool.
#### Scenario: Prompt menciona Decreto 100 y section_id
- **WHEN** un cliente pide `prompts/get` para cualquiera de los dos prompts constitucionales
- **THEN** el mensaje menciona el norm_id del Decreto 100, instruye `get_law_summary` primero y `get_law` con `section_id`, y prohíbe `get_law` sin `section_id` para la CPR

### Requirement: Templates solo referencian tools registradas

Los templates de `answer_constitutional_question` y `check_norm_constitutionality` DEBEN (SHALL) referenciar únicamente tools registradas (`search_laws`, `get_law`, `get_law_summary`, `get_law_history`) y DEBEN (SHALL) mantener la invariante pura: servir el prompt NO realiza llamadas BCN.

#### Scenario: Validación de tools
- **WHEN** se listan todos los templates de prompts
- **THEN** cada nombre de tool mencionado en los dos nuevos prompts pertenece al conjunto retornado por `ToolNames()` y ningún template menciona tools no registradas

### Requirement: Conteo de prompts curados

El servidor DEBE (SHALL) exponer nueve prompts curados vía `prompts/list` tras el cambio (los siete existentes más los dos constitucionales), cada uno con su descripción y obligatoriedad de argumentos declarada.

#### Scenario: Lista ampliada
- **WHEN** un cliente consulta `prompts/list`
- **THEN** la respuesta incluye `answer_constitutional_question` y `check_norm_constitutionality` además de `analyze_law`, `search_legal_topic`, `compare_law_versions`, `trace_law_history`, `check_law_validity`, `explain_law_simply` y `law_research_workflow`
