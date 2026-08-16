## 1. Infra de versión (SSOT)

- [x] 1.1 Crear `internal/version/version.go` con `package version` y `var Version = "dev"` (sin lógica extra, solo símbolo para `-X`).
- [x] 1.2 Verificar que `go vet ./internal/version` pasa y que `go run` sin flags compila.

## 2. Servidor MCP reporta versión real

- [x] 2.1 Editar `internal/server/server.go`: importar `strings` y `github.com/alvarosdev/chile-bcn-mcp/internal/version`, cambiar `Version: "0.1.0"` por `Version: strings.TrimSpace(strings.TrimPrefix(version.Version, "v"))` (soporta `VERSION` con/sin `v` y `dev`).
- [x] 2.2 Añadir test en `internal/server/server_test.go` (o `version_test.go`): `TestServerReportsVersion` que fija `version.Version = "v9.9.9"` y aserta `server.New(logger)` expone `9.9.9` vía `initialize` / `Implementation.Version` (usar `mock.Anything` para ctx si usa `CallTool`).
- [x] 2.3 `grep -r '"0\.1\.0"' --include="*.go"` debe dar 0 resultados.
## 3. Builds locales inyectan VERSION

- [x] 3.1 Editar `Makefile`: definir `VERSION ?= $(shell cat VERSION | tr -d ' \n' | sed 's/^v//')` y `LDFLAGS := -s -w -X github.com/alvarosdev/chile-bcn-mcp/internal/version.Version=$(VERSION)`; cambiar `build:` a `go build -ldflags="$(LDFLAGS)"`.
- [x] 3.2 Verificar `make build && bin/chile-bcn-mcp` y `make build` tras `echo "v9.9.9" > VERSION` reporta `9.9.9` (restaurar `VERSION` después).
- [x] 3.3 Asegurar `make run-http` / `make run-stdio` opcionalmente usan `LDFLAGS` o documentar que `go run` sin flags reporta `dev` (no bloqueante).

## 4. Distribuciones y Docker

- [x] 4.1 Editar `scripts/build-dist.sh:47-49`: `go build -trimpath -ldflags="-s -w -X github.com/alvarosdev/chile-bcn-mcp/internal/version.Version=${VERSION}"` (usar `${VERSION#v}` si se quiere normalizar en bash).
- [x] 4.2 Editar `Dockerfile`: añadir `ARG VERSION=dev` antes de `RUN go build` y cambiar `ldflags` a `-s -w -X ...Version=${VERSION}`.
- [x] 4.3 Editar `.github/workflows/publish.yml:docker:Build and push`: añadir `build-args: VERSION=${{ needs.version.outputs.version }}` a `docker/build-push-action`.
- [x] 4.4 Smoke: `bash scripts/build-dist.sh 0.0.0-test && unzip -l dist.zip && strings dist/linux/amd64/chile-bcn-mcp | grep 0.0.0-test` (o test de `initialize` contra binario extraído).

## 5. Validación y limpieza

- [x] 5.1 Ejecutar `make check` (build+vet+test) y `make fmt-check` — todo verde.
- [x] 5.2 `openspec validate --change single-source-version --strict` sin errores.
- [x] 5.3 Actualizar `README.md` si menciona `0.1.0` hardcodeado (opcional).
- [x] 5.4 Confirmar `VERSION` queda como único editable manual; opcional normalizar `v0.0.6` → `0.0.6` en próximo `release/v*`.
