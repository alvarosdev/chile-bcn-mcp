## Context

El scaffold (`scaffold-mcp-server`, archivado) dejó la tool `echo` como placeholder con el requirement correspondiente en `openspec/specs/mcp-server/spec.md`. El proyecto ya tiene tres tools de dominio (`search_laws`, `get_law`, `get_law_summary`) y 4 suites de tests que validan el registro y los schemas con in-memory transports — echo ya no valida nada que no esté cubierto. Ver proposal.md para la motivación.

## Goals / Non-Goals

**Goals:**
- Quitar toda la superficie de `echo`: tool, args, handler, tests y mención en README.
- Dejar un smoke test formalizado (`scripts/smoke.sh`) como la forma canónica de probar el server real de punta a punta.

**Non-Goals:**
- Tocar las tools de dominio (el cambio es solo remoción + smoke).
- Un integration test con httptest del transporte HTTP (requeriría mover `runHTTP` a `internal/server` — queda para un change futuro).
- MCP conformance oficial.

## Decisions

**1. Remoción completa, no deprecación**
No tiene sentido deprecar una tool placeholder: se quita de `RegisterTools`, se borran `EchoArgs`/`makeEcho` y los dos tests de echo. `errorResult` se conserva (lo usan las tres tools reales). El spec de `mcp-server` lleva un delta REMOVED con razón y migración — primer REMOVED del proyecto, sirve de plantilla para futuros retiros.

**2. `scripts/smoke.sh` — el smoke test canónico**

**Versión del protocolo (explorado con el SDK)**: go-sdk v1.7.0 soporta las 5 versiones del spec y negocia por sesión; el `initialize` legacy está deprecado en 2026-07-28 y el SDK capa esa vía en 2025-11-25 (la negociación nueva es SEP-2575 via `_meta`). El smoke inicializa con `"protocolVersion": "2026-07-28"` y ACEPTA como correcta la respuesta `2025-11-25` (cap legacy esperado, verificado empíricamente) — así el smoke ejercita la negociación más reciente sin falsos fallos.

Script bash con `set -euo pipefail` que asume el server corriendo en `http://127.0.0.1:8000` (sin auth — si `MCP_AUTH_TOKEN` está seteado, el script lo toma de `SMOKE_TOKEN` env para pasarlo como Bearer). Flujo: `GET /health` → `initialize` (captura `Mcp-Session-Id`) → `notifications/initialized` → `tools/list` (verifica que `search_laws`/`get_law`/`get_law_summary` están y `echo` NO) → `tools/call search_laws` (query "Ley 21.600", espera `1195666`) → `tools/call get_law_summary` (norm_id 1142880, espera `categorias_norma`). Cada paso con `|| { echo "✗ paso"; exit 1; }` — falla rápido. El `tools/list` sin `echo` es además la verificación de ESTE change.

**3. Makefile: target `smoke`**
`smoke: ## Run the smoke test (requires the server running: make run-http)` → `bash scripts/smoke.sh`. No levanta el server automáticamente (mantener el script simple y sin responsabilidades de orquestación).

**4. Fixture viva vs valores fijos**
El smoke usa valores reales de BCN (Ley 21.600 → norm_id 1195666; Ley 21.214 → 1142880 con categorías). Si BCN cambiara estos datos, el smoke fallaría — aceptado: es la API pública estable de BCN y un fallo aquí es exactamente lo que el smoke debe detectar (integración rota).

## Risks / Trade-offs

- [El smoke depende de la API real de BCN (red externa)] → Por diseño: no corre en CI, es un target manual (`make smoke`); el fallo es la señal que buscamos.
- [Remover echo es breaking para clientes que la llamaran] → Ningún cliente real la usa (placeholder del scaffold); documentado en el proposal como breaking deliberado.
- [El smoke sin auth] → Soporta `SMOKE_TOKEN` env para servers con `MCP_AUTH_TOKEN`; si no, se documenta correrlo sin token.

## Migration Plan

Sin migración: remover la tool y actualizar la documentación. Rollback trivial (revert). El smoke script es aditivo.

## Open Questions

<!-- Ninguna. -->
