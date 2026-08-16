## Context

Ver `proposal.md — Why`. La API `buscarjson` alterna string/number en el bloque de paginación según la query (verificado: `Ley 21.600` → `"3"`, `Ley 21461` → `10`). El struct `Pagination` (`internal/bcn/law_client.go:160`) fija `npagina`/`itemsporpagina` como `string` y `totalitems` como `int`; cualquier combinación distinta aborta `SearchResponse.UnmarshalJSON` con `decode pagination: json: cannot unmarshal number...`. La falla es de decodificación, no de transporte: el retry/breaker de resty no intervienen y no hay nada que reintentar (la respuesta es 200 válida con tipos inesperados).

Consumo actual de `Pagination`:
- `internal/tools/search_laws.go` usa solo `TotalItems` y `Query` (`buildSearchOutput`, `formatSearchResults`). `Page` y `PageSize` del struct no se exponen — la tool usa `args.Page`/`args.PageSize`.
- `Client.Search` loguea `total_items` en debug.

Es decir: los dos campos string hoy rotos (`Page`, `PageSize`) son casi decorativos, pero su decode falla tumba la búsqueda completa.

## Goals / Non-Goals

**Goals:**
- Decodificar los 3 campos numéricos de paginación (`npagina`, `itemsporpagina`, `totalitems`) en cualquier combinación string/number, incluidos trim (`" 10 "`), float de parte entera (`10.0`), `""` y `null` → 0.
- Fallar ruidoso (error de decode explícito) ante valores no numéricos como `"abc"` — nunca silenciar como 0 un bug de la API.
- Mantener el contrato MCP intacto: `SearchLawsOutput` sigue con `int`; el tipo flexible queda encapsulado en `internal/bcn`.

**Non-Goals:**
- No se auditan `get_norma_json` ni `get_historias_de_ley`: sus fixtures reales no muestran el patrón string|number. Si aparece, se reutiliza el mismo tipo (por eso vive en archivo propio, no inline en `law_client.go`).
- No se tocan las facetas (elemento 2 del array, ignoradas) ni se agregan campos de paginación nuevos (`tipoviene`, `orden`, `fc_*` siguen sin mapear).
- No se cambia el flujo de transporte: timeout/retry/breaker siguen declarados en `config/api.resources.yaml`.

## Decisions

1. **Tipo custom `FlexInt` (`type FlexInt int`) en `internal/bcn/flexint.go` con `UnmarshalJSON` + `MarshalJSON`.**
   - Orden de intentos en Unmarshal: `null` → 0; número → directo; string → trim, `""` → 0, `strconv.Atoi`; número con decimales (`10.0`) → truncado a int; cualquier otra cosa → error explícito con el campo.
   - `MarshalJSON` devuelve número (no string): cualquier serialización futura (logs, snapshots de tests) es determinística y no revive el wire inconsistente.
   - Alternativas descartadas:
     - `json.Number`: es string por dentro; cada lectura exige `.Int64()` + manejo de error disperso por el código, y no cubre `10.0` ni `null` sin helpers propios.
     - `UnmarshalJSON` manual en `Pagination` entero (map de `json.RawMessage`): centraliza pero es verboso, repite lógica por campo y es frágil ante campos nuevos.
     - `interface{}` + post-proceso: pierde tipado, `go vet` deja de ayudar.

2. **`Pagination` migra sus 3 campos numéricos a `FlexInt`; `cadena` queda `string`.**
   - El cambio de tipo es breaking solo dentro del repo: `internal/tools/search_laws_test.go:56` y demás literales `Pagination{Page: "1", ...}` pasan a `FlexInt(1)`. Migración mecánica detectable con `grep -rn "Pagination{"`.
   - `TotalItems` también migra aunque hoy llega como número: si BCN lo manda mañana como `"140"` (misma fuente inconsistente), el bug reaparecería en la dirección opuesta.

3. **`FlexInt` no cruza la frontera MCP.**
   - `SearchLawsOutput` mantiene `int`; la conversión (`int(result.Pagination.TotalItems)`) ocurre donde ya se consume, en `internal/tools`. Esto evita que el generador de schema del go-sdk describa un tipo custom en `structuredContent` y respeta la convención "la traducción termina en la frontera".

4. **Tests en dos niveles, sin red.**
   - Nivel 1: suite de `FlexInt` pura (tabla con todos los wires: number, string, mixto, trim, float, `""`, `null`, `"abc"` → error).
   - Nivel 2: sub-test httptest en `LawClientSuite` con servidor que devuelve paginación numérica (el fixture `testdata/search_response.json` existente queda como regresión del formato string).
   - La verificación con las 3 queries reales (`Ley 21.600`, `Ley 21461`, `21461`) es smoke manual (`make smoke` con el server corriendo), no test de CI — coherente con "los tests nunca tocan la red".

## Risks / Trade-offs

- [BCN agrega una variante no observada (bool, `"10,0"`, notación científica)] → `FlexInt` falla ruidoso con mensaje claro que nombra el valor; el fix puntual es una rama más en `UnmarshalJSON`. No se especulan variantes que el wire nunca mostró.
- [Truncar `10.9` → `10`] → deliberado: paginación no lleva fracción. Cubierto por test para dejar la intención explícita.
- [Churn de literales en tests al cambiar el tipo] → mecánico y acotado (2 archivos); verificado con grep antes de compilar.
- [Marshallar número rompe a alguien que esperaba string] → nadie serializa `Pagination` hoy (solo debug log de un int); riesgo inexistente en la práctica.

## Migration Plan

Cambio interno sin estado persistido ni cambio de contrato: despliegue normal (rebuild de imagen). Rollback = revert del commit; sin migración de datos.
