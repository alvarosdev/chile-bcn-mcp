## 1. Mover y embebeder contrato de endpoints

- [x] 1.1 Mover `config/api.resources.yaml` → `internal/config/api.resources.yaml` y añadir `//go:embed api.resources.yaml` (`rawResources []byte`) en `internal/config/resources.go`
- [x] 1.2 Implementar `LoadEmbedded() (*Resources, error)` en `internal/config/resources.go` ( `yaml.Unmarshal(rawResources)` + `validate()` fail-fast); mantener `Load(path)` solo para `testdata/` y documentar en comentario
- [x] 1.3 Verificar `internal/config/resources_test.go` sigue pasando (`TestLoadValid` via `Load` con fixture, nuevo test `TestLoadEmbedded` que valida el YAML embebido)
- [x] 1.4 Eliminar `config/` top-level y actualizar referencias en `README.md` si lo menciona

## 2. Externalizar prompts a YAML embebido con nombres descriptivos

- [x] 2.1 Crear `internal/prompts/prompts.yaml` con llave `prompts:` y 9 entradas `snake_case` (`analyze_law`, ..., `check_norm_constitutionality`) en bloque literal `|`; migrar contenido de los 9 `fmt.Sprintf` actuales, reemplazando `%s` posicional por `{{.norm_id}}`, `{{.question}}`, `{{.version_date}}`, `{{.tool_get_law}}`, etc. y condicionales `{{if .aspect}}...{{end}}`; CPR `242302` queda literal hardcodeado en los dos prompts constitucionales
- [x] 2.2 Añadir `//go:embed prompts.yaml` (`rawPrompts string`) en `internal/prompts/prompts.go` e implementar parser: `yaml.Unmarshal` → `map[string]string` → `template.New(name).Option("missingkey=error").Parse()` por prompt; definir `PromptSet` y whitelist `norm_id, topic, from_date, to_date, date, aspect, audience, question, article_hint, version_date` + `tool_*`
- [x] 2.3 Implementar `LoadEmbedded() (*PromptSet, error)` con validación: exactamente 9 prompts, sin keys extra/faltantes, sin placeholders fuera de whitelist, `template.Parse` sin error; exponer `LoadEmbedded` fail-fast y mantener `RegisterPrompts` puro (inyecta `PromptSet`, no lee filesystem)
- [x] 2.4 Reescribir `RegisterPrompts(srv, promptSet)` para renderizar `template.Execute(map[string]string)` con `args` de `prompts/get` + `tool_*` inyectados; eliminar `fmt.Sprintf` y ramas `if args["x"] != ""` — pasan a `{{if .x}}` en YAML

## 3. Wiring del servidor

- [x] 3.1 Actualizar `cmd/chile-bcn-mcp/main.go`: reemplazar `const resourcesPath` + `config.Load(resourcesPath)` por `config.LoadEmbedded()` y `prompts.LoadEmbedded()`; inyectar `promptSet` en `prompts.RegisterPrompts(srv, promptSet)`; manejar errores fail-fast con `logger.Error` + `os.Exit(1)` consistente con `resources` existente
- [x] 3.2 Verificar `go vet` no reporta `missingkey` y `go test ./... -race` pasa sin `config/` en disco; smoke `go build -o /tmp/chile-bcn-mcp ./cmd/chile-bcn-mcp && /tmp/chile-bcn-mcp --help` y `mv /tmp/chile-bcn-mcp /tmp/otra/ && /tmp/otra/chile-bcn-mcp` arranca sin `config/`

## 4. Infra: imagen y distribuciones sin config externa

- [x] 4.1 `Dockerfile`: eliminar `COPY config/ /app/config/` y `WORKDIR /app` innecesario; verificar `podman build -t test . && podman run --rm test /usr/local/bin/chile-bcn-mcp --help` arranca sin `config/`; healthcheck sigue en `curl`
- [x] 4.2 `scripts/build-dist.sh`: eliminar `mkdir -p "$DIST/$os/$arch/config" && cp config/api.resources.yaml ...`; actualizar comentario de cabecera a "binaries are self-contained via go:embed, no external config"; verificar `bash scripts/build-dist.sh 0.0.0-test && unzip -l dist.zip | grep -v config` no lista `config/` y cada `dist/$os/$arch/` solo contiene binario
- [x] 4.3 Verificar `SHA256SUMS.txt` sigue generándose solo sobre binarios y `dist.zip` es instalable tras `unzip`

## 5. Tests y validación

- [x] 5.1 Ampliar `internal/prompts/prompts_test.go`: validar que `LoadEmbedded` parsea 9 prompts, que `TestTemplatesReferenceOnlyRegisteredTools` sigue pasando con YAML, que placeholders fuera de whitelist fallan, y que `missingkey` provoca error; mantener `TestListPrompts` (9 prompts) y `TestMissingArgServesWithoutError` (ahora via template)
- [x] 5.2 Ejecutar `make check` (build+vet+test), `make fmt-check` y `openspec validate --strict` sobre el change
- [x] 5.3 Smoke manual: `prompts/list` y `prompts/get` para los 9 prompts (con y sin `version_date`/`article_hint`/`aspect`) contra servidor local `podman`/`go run`, verificando `NEVER invent` y `section_id` en prompts de CPR
