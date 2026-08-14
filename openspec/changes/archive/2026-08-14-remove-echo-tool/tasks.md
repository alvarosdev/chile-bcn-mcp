## 1. Remoción de echo

- [x] 1.1 Quitar `echo` de `RegisterTools` en `internal/tools/tools.go` y borrar `EchoArgs` + `makeEcho` (conservar `errorResult`, lo usan las tools reales)
- [x] 1.2 Quitar `TestEchoTool` y `TestEchoToolIsListed` de `internal/tools/tools_test.go` (y el import de `slices`/helper si quedan sin uso)
- [x] 1.3 Actualizar `README.md`: quitar la mención de echo de la sección de tools
- [x] 1.4 Verificar que `tools/list` del server real muestra solo `search_laws`, `get_law`, `get_law_summary`

## 2. Smoke test

- [x] 2.1 Crear `scripts/smoke.sh` con `set -euo pipefail`: health → initialize (captura sesión) → notifications/initialized → tools/list (verifica las 3 tools presentes y echo ausente) → tools/call search_laws (espera norm_id 1195666) → tools/call get_law_summary (norm_id 1142880, espera categorias_norma); cada paso falla con exit 1; soporta `SMOKE_TOKEN` para Bearer si `MCP_AUTH_TOKEN` está activo
- [x] 2.2 Agregar target `smoke` al Makefile (`bash scripts/smoke.sh`, con comentario `##` para el help y `.PHONY`)
- [x] 2.3 Probar el smoke contra el server real (sin token y con `MCP_AUTH_TOKEN`+`SMOKE_TOKEN`) y verificar que falla correctamente con un server caído
- [x] 2.4 `make check` en verde tras la remoción
