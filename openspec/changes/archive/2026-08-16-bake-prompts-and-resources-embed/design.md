## Context

Ver `proposal.md` para motivación. Estado actual (validado con `graphify` y lectura directa):
- `internal/prompts/prompts.go` (305 líneas, 9 templates `template func(map[string]string) string`) usa `fmt.Sprintf` con `%s` posicional; `checkNormConstitutionalityTemplate` tiene 13 interpolaciones posicionales. `RegisterPrompts(srv)` registra 9 `mcp.Prompt` puros (sin llamadas BCN). Tests en `prompts_test.go` con `PromptsSuite` + `TestTemplatesReferenceOnlyRegisteredTools`.
- `internal/config/resources.go` expone `Load(path) (*Resources, error)` que hace `os.ReadFile` + `yaml.Unmarshal` + `validate()` fail-fast. `cmd/chile-bcn-mcp/main.go:44-45` usa `const resourcesPath = "config/api.resources.yaml"` y `config.Load(resourcesPath)`. `graphify` confirma `main() --calls--> Load() --calls--> NewClient()`.
- `config/api.resources.yaml` (3 recursos: `search_laws`, `get_law`, `get_law_history`) se distribuye vía `Dockerfile:33 COPY config/ /app/config/` y `scripts/build-dist.sh:53-54 mkdir -p .../config && cp config/api.resources.yaml` en cada `dist/$os/$arch/`. `release-distributions/spec.md` exige ese layout.
- Convención del proyecto (`openspec/config.yaml`): config embebida sin hot-reload ni override por env; rebuild para cambiar contrato. `html-to-markdown`, `resty`, `yaml.v3` ya en `go.mod`.

Restricciones: prompts y recursos deben seguir siendo puros/embebidos; `go build` produce binario portable (`mv` entre carpetas funciona); `go test ./... -race` y `go vet` deben seguir pasando sin setup extra; `//go:embed` debe fallar en compile si el archivo falta.

## Goals / Non-Goals

**Goals:**
- Binario autocontenido: `go build ./cmd/chile-bcn-mcp` embebe `api.resources.yaml` y `prompts.yaml`; mover el binario no requiere `config/`.
- Autoría de prompts desacoplada de Go: `internal/prompts/prompts.yaml` multiline `|` con placeholders nombrados `{{.var}}`, validado en startup.
- Eliminación de artefactos redundantes: `dist.zip` solo binarios + `SHA256SUMS.txt`; imagen OCI sin capa `config/`.
- Validación fail-fast y testeable: `missingkey=error` + whitelist + conteo exacto de 9 prompts.

**Non-Goals:**
- Hot-reload o override de prompts/recursos por env/volumen — descartado (una sola verdad: el binario).
- Nuevo lenguaje de template (Sprig, Mustache) — `text/template` stdlib es suficiente.
- Cambios en `internal/bcn`, tools MCP, caché ETag o breaker — solo cambia la fuente de datos estática.
- i18n de prompts o A/B testing sin rebuild — no hay caso de uso probado.

## Decisions

**Decisión 1: Ubicación + `//go:embed` junto al parser (opción A).**
- `internal/config/api.resources.yaml` embebido en `internal/config/resources.go` via `//go:embed api.resources.yaml` (`rawResources []byte`).
- `internal/prompts/prompts.yaml` embebido en `internal/prompts/prompts.go` via `//go:embed prompts.yaml` (`rawPrompts string`).
- Alternativa `config/` top-level + `//go:embed ../../config/...` descartada: `../` en embed es frágil y mezcla ownership infra (top-level visible) con dominio (paquete). Opción A mantiene cohesión — Community 5 "Law Prompts Registry" (4 nodos) y Community 9 "Config Validation" (8 nodos) cada una owning su asset.
- `config/` top-level se elimina en el mismo change.

**Decisión 2: `text/template` con `missingkey=error` vs `fmt.Sprintf` posicional.**
- Elegido `text/template` stdlib. `fmt.Sprintf` con `%s` es seguro si el formato es literal, pero con 13 `%s` el orden es bug farm (`%!s(MISSING)` silencioso). `text/template` da nombres auto-documentados (`{{.norm_id}}`), `missingkey=error` fail-fast, y condicionales `{{if .aspect}}...{{end}}` sin Go `if`.
- Alternativa `strings.ReplaceAll` reimplementa template mal. `html/template` descartado (auto-escape no deseado). Sin deps nuevas.

**Decisión 3: Contrato de `prompts.yaml` — keys = `mcp.Prompt.Name` snake_case, whitelist cerrada.**
- Estructura:
  ```yaml
  prompts:
    analyze_law: |
      Analyze Chilean norm {{.norm_id}}.
      Step 1: call {{.tool_get_law_summary}}(norm_id={{.norm_id}}) ...
      {{if .aspect}}Focus the analysis on: {{.aspect}}.{{end}}
    check_norm_constitutionality: |
      Assess whether Chilean norm {{.norm_id}} is compatible with the Chilean Constitution (Decreto 100, norm_id=242302).{{.question_arg}}...
  ```
- Llaves = `analyze_law`, `search_legal_topic`, `compare_law_versions`, `trace_law_history`, `check_law_validity`, `explain_law_simply`, `law_research_workflow`, `answer_constitutional_question`, `check_norm_constitutionality` — el contrato público ya validado por `TestListPrompts` (9 prompts). No `CamelCase` Go.
- Whitelist de vars: `norm_id, topic, from_date, to_date, date, aspect, audience, question, article_hint, version_date` + `tool_search_laws, tool_get_law, tool_get_law_summary, tool_get_law_history`. Cualquier `{{.foo}}` fuera de ahí → `validate()` falla. Tool names se inyectan como vars para que `TestTemplatesReferenceOnlyRegisteredTools` siga siendo fuente de verdad.
- CPR `242302` queda literal hardcodeado y documentado dentro de los dos prompts constitucionales en el YAML — no parametrizable (canónico Decreto 100).
- Formato: bloque literal `|` para todos (preserva `\n` y `"` sin escapar `Answer "{{.question}}"`).

**Decisión 4: Loaders embebidos sin fallback a filesystem.**
- `config.LoadEmbedded() (*Resources, error)` hace `yaml.Unmarshal(rawResources, &r)` + `r.validate()`. `prompts.LoadEmbedded() (*PromptSet, error)` hace `yaml.Unmarshal(rawPrompts)` + `template.Parse` por prompt + `validate` (9 keys, placeholders whitelist, `missingkey=error`).
- `config.Load(path)` se conserva solo como helper para tests con `testdata/*.yaml` (`ResourcesSuite.fixture`). `main.go` deja de usarlo:
  ```go
  resources, err := config.LoadEmbedded()
  promptSet, err := prompts.LoadEmbedded()
  prompts.RegisterPrompts(srv, promptSet)
  ```
- Alternativa fallback (`if file exists on disk use it else embed`) descartada: introduce ambigüedad "¿qué config estoy corriendo?" y reabre drift `VERSION` vs `0.1.0` ya resuelto. Una sola verdad: el binario.

**Decisión 5: Limpieza de infra.**
- `Dockerfile`: eliminar `COPY config/ /app/config/` y `WORKDIR /app` innecesario; runtime sigue `alpine` non-root, sin capa config.
- `scripts/build-dist.sh`: eliminar `mkdir -p "$DIST/$os/$arch/config" && cp config/api.resources.yaml ...`; `dist.zip` pasa a `binary (+ .exe en windows)` + `SHA256SUMS.txt` por `os/arch`. Comentario de cabecera actualizado ("self-contained binary, no external config").
- `publish.yml`/`ci.yml`: sin cambios funcionales; `go build` y `docker/build-push-action` ya embeben via `//go:embed`.

## Risks / Trade-offs

- **YAML multiline sensible a indent:** `|` + indent extra rompe prompt. Mitigación: `validate()` que hace `template.Parse` + test snapshot que compara `strings.Contains` clave (ya existe `TestTemplatesReferenceOnlyRegisteredTools` ampliado para YAML).
- **Pérdida de `go vet` para `%s`:** `vet` ya no detecta mismatch posicional. Mitigación: `missingkey=error` + whitelist + test que renderiza los 9 prompts con args completos y verifica `get_law_summary` etc.
- **Edición de prompts requiere Go toolchain mentalmente:** aunque el copy está en YAML, el PR sigue siendo Go (embed). Mitigación: `internal/prompts/prompts.yaml` es editable sin tocar `.go`; reviewer no-Go puede aprobar solo el YAML.
- **Volumen externo ya no funciona (BREAKING):** quien montaba `config/api.resources.yaml` por volumen verá que se ignora. Mitigación: documentado en `proposal.md` Impact y en delta specs; migración = rebuild imagen/binario.
- **Tamaño binario:** `rawResources` (~1KB) + `rawPrompts` (~12KB) < 15KB overhead, despreciable.

## Migration Plan

- Cambio atómico en un PR: mover YAML, añadir embeds, migrar templates a `{{.var}}`, actualizar `main.go`, limpiar `Dockerfile`/`build-dist.sh`, borrar `config/`.
- Deploy: `make check` (build+vet+test) local, `make podman-build`/`compose-up` smoke `prompts/list` + `prompts/get` para los 9 prompts, `scripts/build-dist.sh 0.0.0-test` y verificar `dist.zip` sin `config/`.
- Rollback: revert commit y redeploy; no hay migración de datos.

## Open Questions

Ninguna — las 4 preguntas abiertas (ubicación A, whitelist cerrada, sin fallback, validación estricta 9 keys + `|` literal) fueron confirmadas por el usuario el 2026-08-16.
