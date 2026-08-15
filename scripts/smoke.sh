#!/usr/bin/env bash
# smoke.sh — smoke test del server MCP contra la API real de BCN.
#
# Uso:
#   make run-http   (en otra terminal)
#   make smoke
#
# Verifica de punta a punta: health, negociación MCP, tools/list (echo
# ausente) y llamadas reales a search_laws y get_law_summary.
# Si el server corre con MCP_AUTH_TOKEN, exportar SMOKE_TOKEN=<token>.
set -euo pipefail

BASE="${SMOKE_BASE:-http://127.0.0.1:8000}"
CT="Content-Type: application/json"
ACCEPT="Accept: application/json, text/event-stream"
AUTH=()
if [[ -n "${SMOKE_TOKEN:-}" ]]; then
  AUTH=(-H "Authorization: Bearer $SMOKE_TOKEN")
fi

step() { echo "· $1"; }
fail() { echo "✗ $1"; exit 1; }

step "health"
curl -sf "$BASE/health" >/dev/null || fail "health check"

step "initialize (protocol 2026-07-28 — el SDK capa el initialize legacy en 2025-11-25)"
# -i: el Mcp-Session-Id viene en los headers de la respuesta
INIT=$(curl -s -i "${AUTH[@]}" -H "$CT" -H "$ACCEPT" -X POST "$BASE/mcp" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2026-07-28","capabilities":{},"clientInfo":{"name":"smoke","version":"0.1"}}}')
grep -q '"result"' <<<"$INIT" || fail "initialize sin result"
SID=$(grep -io 'Mcp-Session-Id: [A-Z0-9]*' <<<"$INIT" | awk '{print $2}')
[[ -n "$SID" ]] || fail "no se obtuvo Mcp-Session-Id"

curl -s "${AUTH[@]}" -H "$CT" -H "$ACCEPT" -H "Mcp-Session-Id: $SID" \
  -X POST "$BASE/mcp" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' >/dev/null

step "tools/list — las 3 tools reales, echo ausente"
TOOLS=$(curl -s "${AUTH[@]}" -H "$CT" -H "$ACCEPT" -H "Mcp-Session-Id: $SID" \
  -X POST "$BASE/mcp" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}')
for t in search_laws get_law get_law_summary; do
  grep -q "\"name\":\"$t\"" <<<"$TOOLS" || fail "tool $t ausente de tools/list"
done
grep -q '"name":"echo"' <<<"$TOOLS" && fail "echo sigue registrada" || true

step "search_laws — Ley 21.600 (norm_id 1195666)"
SEARCH=$(curl -s "${AUTH[@]}" -H "$CT" -H "$ACCEPT" -H "Mcp-Session-Id: $SID" \
  -X POST "$BASE/mcp" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_laws","arguments":{"query":"Ley 21.600","page_size":3}}}')
grep -q '1195666' <<<"$SEARCH" || fail "search_laws no devolvió la Ley 21.600"

step "get_law_summary — Ley 21.214 (norm_id 1142880, con categorias_norma)"
SUMMARY=$(curl -s "${AUTH[@]}" -H "$CT" -H "$ACCEPT" -H "Mcp-Session-Id: $SID" \
  -X POST "$BASE/mcp" \
  -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_law_summary","arguments":{"norm_id":1142880}}}')
grep -q 'categorias_norma' <<<"$SUMMARY" || fail "get_law_summary sin categorias_norma"
# El summary trae la estructura (mapa con section ids) y el tamaño, pero
# nunca el contenido de la norma (## Content solo aparece en get_law).
grep -q '## Structure' <<<"$SUMMARY" || fail "get_law_summary sin estructura"
grep -q 'Size:' <<<"$SUMMARY" || fail "get_law_summary sin tamaño"
grep -q '## Content' <<<"$SUMMARY" && fail "get_law_summary no debe traer contenido" || true

echo "✓ smoke ok — el server habla MCP de verdad"
