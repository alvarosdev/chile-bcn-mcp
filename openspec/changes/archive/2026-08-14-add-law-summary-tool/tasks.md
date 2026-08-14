## 1. Cliente BCN

- [x] 1.1 Agregar `CategoriasNorma []string `json:"categorias_norma"`` al struct `Metadatos` en `internal/bcn/law_client.go` (campo aditivo, tal cual viene de la API)
- [x] 1.2 Definir `NormaSummary{TituloNorma, Fuente, Materias, CategoriasNorma, Resumenes}` y el método `GetNormaSummary(ctx, normID) (NormaSummary, error)` en `LawClient`: reutiliza el resty client del recurso `get_law`, el parseo de `NormaFull`, el sanitizer de `resumenes` y el manejo de `ErrNormaNotFound`; consulta el caché ETag por `norm_id` (si existe) y proyecta sin re-descargar
- [x] 1.3 Extender `internal/bcn/law_client_test.go` (suite + httptest): summary parsea título/fuente/materias/categorías/resúmenes sanitizados del fixture real, 500 → `ErrNormaNotFound`, y derivación desde caché sin nueva request HTTP (server recibe 1 request en 2 llamadas summary consecutivas — si el caché ya está implementado; si no, marcar como pendiente de la integración del caché)

## 2. Tool MCP

- [x] 2.1 Implementar `internal/tools/get_law_summary.go`: `GetLawSummaryArgs{NormID int64}` con tags, `RegisterGetLawSummary(srv, client)` con `mcp.AddTool` y descripción en inglés, handler que valida `NormID > 0` y mapea `ErrNormaNotFound` a mensaje claro
- [x] 2.2 Output LLM-first: `GetLawSummaryOutput` tipado (mismos campos de `NormaSummary`) como structuredContent + `TextContent` formateado (título, fuente, materias y categorías como listas, resumen oficial)
- [x] 2.3 Actualizar `RegisterTools` en `internal/tools/tools.go` para registrar `get_law_summary` junto a las existentes
- [x] 2.4 Escribir `internal/tools/get_law_summary_test.go` (suite + `MockLawClient` en `SetupTest`, ctx con `mock.Anything`): resumen válido (texto + structuredContent tipado), `norm_id` inválido → error sin llamar al cliente, `ErrNormaNotFound` → mensaje claro

## 3. Mock y verificación

- [x] 3.1 Regenerar el mock: `make mock` → `law_client_mock.go` incluye `GetNormaSummary` (verificar que los tests existentes siguen en verde)
- [x] 3.2 `make check` completo en verde
- [x] 3.3 Smoke manual contra la API real (no en CI): `tools/call get_law_summary` con `norm_id: 1142880` (Ley 21.214, tiene `categorias_norma`) y con `norm_id: 1195666` — verificar payload liviano (sin contenido) y categorías presentes
- [x] 3.4 Actualizar `README.md`: agregar `get_law_summary` a la sección de tools
