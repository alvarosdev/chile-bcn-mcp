## 1. Verificación previa

- [x] 1.1 Verificar `norm_id` del Decreto 100 contra BCN con `search_laws(query="Decreto 100 Constitución Política")` y confirmar 242302; documentar fallback en templates
- [x] 1.2 Revisar `get_law_summary`/`get_law(section_id)` sobre Decreto 100 en ambiente real para confirmar TOC (Capítulos I-XV + transitorias) y tamaños por sección

## 2. Implementación de prompts

- [x] 2.1 Agregar constantes y templates `answerConstitutionalQuestionTemplate` y `checkNormConstitutionalityTemplate` en `internal/prompts/prompts.go` siguiendo patrón `lawResearchWorkflowTemplate` (pure template, inyección de args, sin llamadas BCN)
- [x] 2.2 Registrar ambos prompts en `RegisterPrompts` con `mcp.Prompt` (nombres en inglés, descripciones, args: `answer_constitutional_question` con `question` required + `article_hint`/`version_date` optional; `check_norm_constitutionality` con `norm_id` required + `question`/`version_date` optional)
- [x] 2.3 Actualizar `ToolNames()` y constantes `toolSearchLaws`/`toolGetLaw`/etc. si aplica (reuso, sin nuevos tools) y asegurar que templates solo referencian tools registradas

## 3. Tests

- [x] 3.1 Ampliar `internal/prompts/prompts_test.go`: `TestListPrompts` espera 9 prompts y valida required/optional de los dos nuevos; `TestTemplatesReferenceOnlyRegisteredTools` incluye ambos templates
- [x] 3.2 Agregar tests específicos: `answer_constitutional_question` inyecta `question`/`article_hint`/`version_date` y contiene `get_law_summary`, `section_id`, hedge y disclaimer; `check_norm_constitutionality` inyecta `norm_id`/`question`/`version_date` y menciona ambas normas + paralelismo
- [x] 3.3 Verificar caso sin `version_date` no menciona `version_date` y caso con `version_date` sí lo incluye (paridad con `check_law_validity`)

## 4. Documentación y validación

- [x] 4.1 Actualizar `README.md` (lista de prompts) y `openspec/specs/law-prompts/spec.md` en `main` tras archive si aplica
- [x] 4.2 Ejecutar `make check` (build+vet+test), `make fmt-check` y `openspec validate --strict` sobre el change
- [x] 4.3 Smoke manual: `prompts/list` y `prompts/get` para ambos prompts con y sin `version_date` en servidor local (podman)
