## Why

Los prompts viven hoy como `fmt.Sprintf` con `%s` posicional dentro de `internal/prompts/prompts.go` — editar copy obliga a tocar Go y contar 13 posiciones en `checkNormConstitutionalityTemplate` es frágil. El contrato de endpoints `config/api.resources.yaml` se carga por `os.ReadFile("config/api.resources.yaml")` y se distribuye copiado en `Dockerfile` y en cada carpeta de `dist.zip`; mover el binario fuera de `config/` lo rompe y obliga a mantener artefactos redundantes. Se necesita bakeado en binario: `go build` produce un binario autocontenido, sin hot-reload ni override por env, portable entre carpetas.

## What Changes

- Mueve `config/api.resources.yaml` → `internal/config/api.resources.yaml` y lo embebe con `//go:embed api.resources.yaml` en `internal/config/resources.go`; expone `LoadEmbedded() (*Resources, error)` que hace `yaml.Unmarshal` sobre `rawResources` + `validate()` fail-fast. `Load(path)` se mantiene solo para tests con `testdata/`.
- Crea `internal/prompts/prompts.yaml` embebido con `//go:embed prompts.yaml` en `internal/prompts/prompts.go`; cada llave es el `mcp.Prompt.Name` en `snake_case` (`analyze_law`, `check_norm_constitutionality`, etc.) con bloque literal `|` multiline. Contenido migra de `fmt.Sprintf` a `text/template` con placeholders nombrados `{{.norm_id}}`, `{{.question}}`, `{{.version_date}}`, etc. y condicionales `{{if .aspect}}...{{end}}`.
- Estandariza placeholders descriptivos: whitelist `norm_id, topic, from_date, to_date, date, aspect, audience, question, article_hint, version_date` + `tool_search_laws, tool_get_law, tool_get_law_summary, tool_get_law_history`. Templates se parsean con `Option("missingkey=error")` y `validate()` que exige exactamente 9 prompts y rechaza placeholders fuera de whitelist. Tool names se inyectan como vars, no hardcodeados en YAML.
- Mantiene CPR `242302` (Decreto 100) hardcodeado y documentado dentro de los dos prompts constitucionales en el YAML — no parametrizable.
- Actualiza `cmd/chile-bcn-mcp/main.go` a `config.LoadEmbedded()` y `prompts.LoadEmbedded()` + `prompts.RegisterPrompts(srv, promptSet)`; elimina `const resourcesPath`.
- Limpia infra: elimina `COPY config/ /app/config/` y `WORKDIR /app` innecesario de `Dockerfile`; elimina `mkdir -p .../config && cp config/api.resources.yaml` de `scripts/build-dist.sh`; `dist.zip` pasa a contener solo binarios + `SHA256SUMS.txt`. `config/` top-level se elimina.
- Sin hot-reload, sin fallback a filesystem: una sola verdad es el binario. Cambiar copy o endpoints requiere rebuild.

## Capabilities

### New Capabilities
- No se crean capabilities nuevas aisladas; el bakeado es un cambio de empaquetado y de autoría de prompts dentro de capabilities existentes.

### Modified Capabilities
- `container-deployment`: la imagen OCI ya no necesita capa `config/`; el binario lleva el contrato embebido y el contenedor sigue siendo read-only sin escritura a disco.
- `release-distributions`: `dist.zip` ya no incluye `config/api.resources.yaml` por `os/arch`; cada carpeta es solo el binario autocontenido. El zip sigue siendo el artefacto versionado con `SHA256SUMS.txt`.

## Impact

- Código: `internal/config/resources.go` (+ embed + `LoadEmbedded`), `internal/config/api.resources.yaml` (movido), `internal/prompts/prompts.go` (+ embed + `text/template` + validate), `internal/prompts/prompts.yaml` (nuevo), `cmd/chile-bcn-mcp/main.go` (usa loaders embebidos), tests de ambos paquetes.
- Infra: `Dockerfile`, `scripts/build-dist.sh`; `config/api.resources.yaml` eliminado de raíz. `publish.yml`/`ci.yml` sin cambios funcionales (el `go build` ya embebe).
- Docs: `README.md` si menciona `config/`; specs delta para `container-deployment` y `release-distributions`.
- **BREAKING** para quien montaba `config/api.resources.yaml` externo por volumen: ya no se lee del filesystem; debe rebuildar imagen/binario para cambiar endpoints. Intencional y documentado.
