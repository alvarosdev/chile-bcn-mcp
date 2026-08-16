## 1. Tipo FlexInt

- [x] 1.1 Crear `internal/bcn/flexint.go`: `type FlexInt int` con `UnmarshalJSON` (null → 0; número directo; string con trim, `""` → 0 y `strconv.Atoi`; decimal de parte entera `10.0` → truncado; cualquier otro valor → error explícito que nombre el valor) y `MarshalJSON` devolviendo número
- [x] 1.2 Crear `internal/bcn/flexint_test.go`: suite testify con tabla de wires — `10`, `"10"`, `" 10 "`, `10.0`, `""`, `null` (→ 0), `"abc"` (→ error) — y round-trip de `MarshalJSON` como número

## 2. Migración de Pagination

- [x] 2.1 Migrar `Pagination` en `internal/bcn/law_client.go`: `npagina`, `itemsporpagina` y `totalitems` → `FlexInt`; `cadena` queda `string`; actualizar doc comment del struct con la inconsistencia del servicio
- [x] 2.2 Ajustar consumidores: conversión `int(...)` donde se lee `Pagination.TotalItems` (`internal/tools/search_laws.go` en `buildSearchOutput`/`formatSearchResults`, debug log de `Client.Search`); `SearchLawsOutput` mantiene `int`
- [x] 2.3 Migrar literales de test: `grep -rn "Pagination{"` y actualizar `internal/tools/search_laws_test.go` (`Page: "1"` → `FlexInt(1)`) y cualquier otro literal afectado

## 3. Tests de integración (sin red)

- [x] 3.1 Sub-test httptest en `LawClientSuite` con paginación numérica (wire tipo `Ley 21461`: `[[], {"npagina":1,"itemsporpagina":10,"totalitems":5}, []]`): `Search` sin error, `TotalItems` correcto
- [x] 3.2 Sub-test con formatos mixtos y vacíos (`"npagina": 1` junto a `"itemsporpagina": "10"`, `""` y `null` → 0); el fixture existente `testdata/search_response.json` queda como regresión del formato string

## 4. Verificación

- [x] 4.1 `make check` (build + vet + test) y `make fmt-check` en verde
- [x] 4.2 Smoke contra la API real (server corriendo): `Ley 21.600` → norm_id 1195666 (regresión), `Ley 21461` y `21461` → norm_id 1178004 sin `decode pagination`, y `get_law_summary(norm_id=1178004)` sin regresión
- [x] 4.3 Documentar la inconsistencia string|number del servicio como gotcha en `CLAUDE.md` y correr `graphify update .`
