## Why

Hoy no existe un flujo curado para preguntas de constitucionalidad sobre la Constitución Política de la República (Decreto 100). La CPR es la norma más consultada y la más grande (~361K chars, 143 artículos + transitorias, ~90K tokens): cargarla completa como hace `analyze_law` rompe contexto y costo. Se necesita interpretar el texto vigente o histórico y contrastar otras normas contra la CPR, distinguiendo interpretación informativa de asesoría legal personalizada, sin que el LLM invente artículos ni cargue texto innecesario.

## What Changes

- Agrega prompt `answer_constitutional_question`: Q&A informativo sobre la CPR. Responde "¿qué dice la Constitución sobre X?" citando artículos, con soporte de `version_date` para consultar texto histórico.
- Agrega prompt `check_norm_constitutionality`: contraste normativo "¿es la Ley Y compatible con la CPR?". Compara una norma (por `norm_id`) contra la CPR artículo por artículo, también versionado.
- Ambos prompts fijan flujo económico `get_law_summary(Decreto 100) → TOC con section_id → get_law(section_id)` — NUNCA `get_law` sin `section_id` salvo norma corta. Reusan verificación y citación obligatoria.
- Permiten veredicto textual "constitucional/inconstitucional" pero obligan a hedge condicional ("conforme al texto retornado... podría interpretarse como...") y disclaimer fuerte "información general, no asesoría legal; la calificación vinculante corresponde al TC; verificar en bcn.cl y consultar profesional".
- Sin cambios de tools ni de API BCN; prompts puros (sin llamadas externas), consistentes con `law_research_workflow`.

## Capabilities

### New Capabilities
- No se crean capabilities nuevas aisladas; la funcionalidad vive como extensión del dominio de prompts.

### Modified Capabilities
- `law-prompts`: se amplía de siete a nueve prompts curados. Nuevos requisitos: `answer_constitutional_question` (args `question` requerido, `article_hint` y `version_date` opcionales) y `check_norm_constitutionality` (args `norm_id` requerido, `question` y `version_date` opcionales), flujo section_id sobre Decreto 100, regla de hedge para veredictos y disclaimer reforzado.

## Impact

- Código: `internal/prompts/prompts.go` (dos templates + constantes `toolSearchLaws`/`toolGet*` ya existentes, sin nuevos tools), `internal/prompts/prompts_test.go` (nuevos casos + ampliación de `TestTemplatesReferenceOnlyRegisteredTools`).
- Docs: `README.md` lista de prompts, `openspec/specs/law-prompts/spec.md`.
- Sin impacto en `internal/bcn`, `config/api.resources.yaml`, transporte MCP ni infra. Sin breaking changes; prompts nuevos son aditivos.
