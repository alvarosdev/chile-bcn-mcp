## Context

Ver `proposal.md` para motivación. Estado actual: siete prompts puros en `internal/prompts/prompts.go` que guían `search_laws`/`get_law`/`get_law_summary`/`get_law_history`. Existe `law_research_workflow` que ya resuelve "norma gigante → summary + section_id" pero es genérico (requiere `norm_id`). La CPR (Decreto 100, ID 242302 hardcodeado — estable históricamente, verificado vía `search_laws`) es la norma más grande del sistema (~361K chars, 143 art + 51 transitorias, ~90K tokens) y la más consultada para constitucionalidad, con alta demanda de versionado histórico por reformas frecuentes. No hay prompts que fijen la CPR como fuente canónica ni que formalicen el contraste "ley vs CPR".

Restricciones del proyecto: prompts son templates puros sin llamadas BCN; interfaz en inglés; tools devuelven TextContent + structuredContent; tests sin red con MockLawClient; config embebida.

## Goals / Non-Goals

**Goals:**
- Dos prompts hermanos que cubran Q&A CPR y contraste normativo sin cargar texto completo, con versionado y hedge para veredictos.
- Reusar mecánica existente `get_law_summary → section_id → get_law(section_id)` sin nuevos tools ni cambios en `internal/bcn`.
- Disclaimer pragmático: permitir "constitucional/inconstitucional" con lenguaje condicional y atribución al TC.

**Non-Goals:**
- Búsqueda semántica local dentro de la norma (índice vectorial/RAG local); se pospone — requeriría infra y frescura, no se justifica ahora.
- Nuevo tool `search_within_norm`; no se justifica aún.
- Cambios en cliente BCN, caché o breaker.
- Snapshot local de la Constitución; descartado — toda lectura va contra LeyChile.

## Decisions

**Decisión 1: Dos prompts separados vs uno híbrido con `norm_id` opcional.**
Elegido: dos prompts. Alternativa híbrido (`constitutional_interpretation` con `norm_id?`) ahorra un registro pero obliga a ramas `if norm_id` en el template, más frágil y menos testeable. Dos prompts hacen el contrato explícito: `answer_constitutional_question` (1 norma: CPR) vs `check_norm_constitutionality` (2 normas: CPR + otra). Sigue el precedente `analyze_law` vs `law_research_workflow`.

**Decisión 2: Decreto 100 hardcodeado 242302 interno al template.**
Elegido: template hardcodea `norm_id=242302` como canónico interno (estable por años; verificado vía `search_laws(query='Decreto 100 Constitución Política')`), sin documentarlo como canónico en README. Alternativa resolver siempre vía `search_laws` añade un call y ruido. Se mantiene mención de fallback solo como nota de recuperación, no como flujo principal.

**Decisión 3: Reusar flujo `get_law_summary` + `section_id`, sin tool nueva. Transitorias ya cubiertas.**
Elegido: `EstructuraPart {n,i,t,h}` ya expone `section_id` y `ContentCharCount`/`humanCount` el costo. `get_law_summary` es liviano y cacheable por `(normID, versionDate)` con ETag; `get_law(section_id)` trae 4-8K chars vs ~361K de la norma completa. El código ya cubre transitorias: `DISPOSICIONES TRANSITORIAS` es `t:1` con hijos `t:6` (ej: `Artículo primero Transitorio | id:10455632`), por lo que `SectionContent(id)` y `FlattenStructure` las exponen sin cambio. Sin snapshot local: toda lectura va contra LeyChile para mantener frescura y versionado. El prompt solo debe instruir al LLM a considerar `DISPOSICIONES TRANSITORIAS` como sección elegible cuando `question` mencione reforma/plebiscito/vigencia/transitoria (ej: "plebiscito 2022" → `get_law(242302, section_id=<VIGÉSIMAQUINTA>)`).

**Decisión 4: `version_date` compartido, `article_hint` string libre, `get_law_history` recomendado.**
`version_date` compartido para ambas normas en el contraste para coherencia temporal (comparar Ley 2020 vs CPR 2020); fechas separadas quedan como extensión futura no necesaria para MVP. `article_hint` es string libre con ejemplos "19", "19 Nº24", "93", "transitoria primera" — match laxo contra `Estructura[].n`, no validación estricta. `get_law_history` pasa de opcional silencioso a paso recomendado (SHOULD) en `check_norm_constitutionality` cuando la otra norma pudo tener control TC.

**Decisión 5: Veredicto permitido con hedge obligatorio.**
Elegido: template enseña patrón `"conforme al texto retornado de Art. X CPR (sección Y), ... podría interpretarse como (in)compatible, en la medida que..."` + disclaimer inicio y cierre. Alternativa prohibir veredicto binario es más segura legalmente pero frustra la petición explícita del usuario ("lo permito"). Hedge + atribución al TC es el compromiso defensivo.

## Risks / Trade-offs

- **Elección errónea de sección por LLM** si `question` es vaga ("¿es justa?") → Mitigación: template obliga a justificar elección ("Indica qué entrada del TOC elegiste y por qué antes de cargar"), incluyendo DISPOSICIONES TRANSITORIAS cuando aplique.
- **Falso veredicto tomado como asesoría** aunque haya disclaimer → Mitigación: disclaimer enmarcado (inicio + cierre), lenguaje condicional, mención explícita "calificación vinculante corresponde al TC".
- **Transitorias confundidas con artículos permanentes** (51 transitorias con `t:6`) → Mitigación: ya cubiertas por código; el prompt advierte que son secciones separadas bajo `DISPOSICIONES TRANSITORIAS` y deben pedirse por su `section_id` (ej: `Artículo primero Transitorio | section_id: 10455632`).
- **Carga doble en contraste** (dos normas grandes) → Mitigación: ambas usan `section_id`; si una es corta (`Size` lo indica) se permite full. Caché por `(id, version)` + singleflight evita tormenta.
- **Alucinación de artículos** → Mitigación: regla heredada "NEVER invent articles — verify against returned text" en ambos templates.

## Migration Plan

- Cambio aditivo, sin breaking: `prompts/list` pasa de 7 a 9; clientes existentes no afectados.
- Deploy: build + push imagen (embebe prompts), `make check` local, `make podman-build`/`compose-up`. Rollback: revertir commit de prompts y redeployar.
- Validación: `go test ./internal/prompts -run TestListPrompts` debe reflejar 9 prompts; `TestTemplatesReferenceOnlyRegisteredTools` debe pasar sin mencionar tools no registradas.

## Open Questions

Sin preguntas abiertas. Resueltos: 242302 hardcodeado interno al template (no canónico en README), `article_hint` string libre con ejemplos, `version_date` compartido, `get_law_history` como paso recomendado, transitorias cubiertas por código existente (solo wording del prompt).
