# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

MCP server en Go 1.26 (`modelcontextprotocol/go-sdk` v1.7.0) que expone leyes chilenas de LeyChile (BCN) a agentes de IA: `search_laws`, `get_law` y `get_law_summary`, sobre transporte stdio + streamable HTTP. `make help` es el índice de todo.

## Commands

```bash
make check        # build + vet + test (lo que corre CI — usar siempre antes de terminar)
make test         # go test ./... -count=1
make run-http     # server en :8000 (default)
make run-stdio    # FASTMCP_TRANSPORT=stdio
make smoke        # smoke test contra la API REAL de BCN (requiere el server corriendo;
                  #   si hay MCP_AUTH_TOKEN, exportar SMOKE_TOKEN=<token>)
make mock         # regenera internal/bcn/law_client_mock.go (mockery v3, config en .mockery.yml)
make fmt-check    # falla si hay archivos sin gofmt
make podman-build / podman-run / podman-stop / podman-logs   # podman es el runtime principal
make compose-up / compose-down    # podman-compose con fallback a docker compose
make dist                     # distribuciones cross-platform (6 targets, zip autocontenido)
```

Test individual: `go test ./internal/bcn/ -run TestLawClientSuite/TestGetNormaServes304FromCache -v`. Los tests **nunca** tocan la red: httptest.Server + fixtures reales de BCN en `internal/bcn/testdata/`.

## Architecture

```
cmd/chile-bcn-mcp/main.go     ← bootstrap: config.Load("config/api.resources.yaml") — ruta
                                FIJA (sin env override, sin hot-reload; la imagen embebe
                                config/) → bcn.NewClient (instancia única inyectada, nunca
                                estado global) → RegisterTools(srv, client) → transporte
                                (stdio | HTTP con auth Bearer + /health + shutdown)
internal/config/              ← contrato de endpoints (api.resources.yaml): url/path/method/
                                timeout/retry/breaker POR RECURSO, validación fail-fast
                                (Duration custom parsea "10s")
internal/bcn/                 ← el dominio: LawClient (interfaz mockeada) → resty v3 con UN
                                client por recurso (el breaker de resty es client-level);
                                NormaFull con bloques HTML ANIDADOS (ver gotchas); caché
                                ETag compartido; sanitize.go + garbage.go
internal/tools/               ← las 3 tools MCP: capa de presentación (validar args, formatear)
```

**Separación de capas**: el cliente parsea/limpia/convierte (todo lo feo de BCN vive en `internal/bcn`); las tools solo validan argumentos y formatean lo que ve el LLM.

**Norma LLM-first + structured** (todo el proyecto): cada tool devuelve `TextContent` formateado para el modelo **y** `structuredContent` tipado como segundo valor del handler (`mcp.ToolHandlerFor[Args, Output]` — el go-sdk autogenera el outputSchema). El texto es una **vista** del structured (puede truncar; el structured lleva datos completos). Nunca JSON embebido como string en el texto.

**Flujo OpenSpec**: el proyecto es spec-driven (`openspec/`). Un feature nuevo NO se implementa directo: `openspec new change "<name>"` → proposal/specs/design/tasks → apply → archive con sync a `openspec/specs/`. El `openspec/config.yaml` contiene las convenciones del proyecto (se inyectan al crear artifacts) — léelo antes de escribir artifacts.

## Gotchas (aprendidas con dolor — no repetir)

1. **Bloques HTML anidados**: `get_norma_json` anida artículos bajo títulos/párrafos en el campo `h` (`HtmlBlock.H` recursivo). Un parseo plano pierde TODO el contenido de las leyes largas (Ley 21.600: 7KB vs 205KB). `ConvertContent` y los renders DEBEN caminar el árbol.
2. **Tipos recursivos en outputs rompen el schema**: el generador de JSON Schema del go-sdk entra en ciclo (`panic: cycle detected`) si el Output contiene un tipo recursivo (ej. `EstructuraPart`). Aplanar con `depth` (ver `StructurePartOut`).
2b. **El Output DEBE ser un tipo objeto**: MCP (SEP-2106) exige que `outputSchema` describa un objeto — un slice top-level (ej. `[]HistoriaGrupo`) genera `"type": "array"` y los clientes estrictos (zod) rechazan `tools/list` entero. Envolver en un struct (`GetLawHistoryOutput{Groups []...}`).
3. **Mocks en handlers MCP**: matchear el `ctx` con `mock.Anything` SIEMPRE — el go-sdk pasa un contexto derivado; un mismatch hace `FailNow`/`Goexit` en la goroutine del handler (que no es la del test) y `CallTool` **cuelga para siempre** (sin mensaje).
4. **Args con defaults**: llevan `,omitempty` en el tag `json` — sin omitempty el schema generado los marca `required`.
5. **resty v3**: el módulo es `resty.dev/v3` (v3.0.0-rc.3, NO `github.com/go-resty/resty/v3`). No reintenta nada por defecto — las condiciones (`5xx` + status-zero) se declaran explícitas en `retryConditions`. El 304 con `SetResult` (body vacío) se maneja antes del chequeo de error.
6. **Caracteres especiales en Go source**: los literales invisibles (BOM, nbsp, zero-width) son ILEGALES en archivos .go — siempre como escapes (`'\\u00a0'`, `'\\ufeff'`); toda la "basura" de BCN vive nombrada en `internal/bcn/garbage.go`.
7. **Sanitizer**: limpia basura de FORMATO (nbsp, `div.p` vacíos → colapsar 3+ newlines, XML wrapper, control chars); NUNCA toca contenido legal (comillas de texto insertado, enlaces a normas) — es semántica jurídica.
8. **Protocolo MCP**: go-sdk v1.7.0 soporta las 5 versiones del spec y negocia por sesión. El `initialize` legacy está deprecado en 2026-07-28: el SDK capa esa vía en `2025-11-25` (verificado empíricamente) — un cliente que pide 2026-07-28 recibe 2025-11-25, es correcto. El smoke lo acepta explícitamente.
9. **Etiquetas de tools/interface en INGLÉS**; los datos crudos del dominio (texto legal español, campos como `idNorma`) se mantienen tal cual. La traducción termina en la frontera: `norm_id` (tool) → `idNorma` (query param).
