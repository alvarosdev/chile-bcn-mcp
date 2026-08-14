## Context

`get_law` está implementada sobre `get_norma_json` (change add-bcn-leychile-tools, activo con tasks pendientes de structured-output y caché ETag). La API **no tiene un endpoint de resumen**: devuelve siempre el response completo. El valor de `get_law_summary` es del lado del cliente MCP: payload mínimo para el LLM y cero costo de conversión (el parseo del DOM es el paso caro). Ver proposal.md para la motivación; requisitos en specs/leychile-search.

## Goals / Non-Goals

**Goals:**
- Tool ligera `get_law_summary` sobre el MISMO recurso `get_law` del YAML (mismo transporte: timeout/retry/breaker) — el YAML declara contratos de transporte, no vistas de datos.
- Reutilizar el parseo existente (`NormaFull`) y proyectar a `NormaSummary` — sin duplicar lógica de unmarshal ni de sanitización.
- Compartir el caché ETag de `get_law` (misma entrada por `norm_id`): un summary tras un `get_law` (o viceversa) no toca la red.

**Non-Goals:**
- Un endpoint nuevo en el YAML (el recurso es el mismo).
- Resumir el texto con un modelo de IA (el `resumenes` oficial de BCN es la fuente).
- Alterar la firma o el comportamiento de `get_law` (solo se agrega el campo de metadatos).

## Decisions

**1. `get_law_summary` es una vista, no un recurso nuevo**
Método `GetNormaSummary(ctx, normID)` en `LawClient` que reutiliza el resty client del recurso `get_law`, el parseo de `NormaFull`, el manejo de `ErrNormaNotFound` y el sanitizer de `resumenes`. Proyección `NormaSummary{TituloNorma, Fuente, Materias, CategoriasNorma, Resumenes}` — exactamente los campos pedidos por el usuario, sin `NumeroFuente` ni otros (el usuario definió la lista; agregar campos después es aditivo).
*Alternativa descartada*: resource `get_law_summary` en el YAML — duplica la declaración de transporte para el mismo endpoint y rompe la semántica "el YAML declara dónde y cómo, el código declara qué".

**2. Caché compartido con `get_law`**
El summary consulta la misma entrada del caché ETag por `norm_id` (change add-bcn-leychile-tools, tasks 6.x pendientes). Flujo: si hay entrada cacheada → derivar el summary sin HTTP; si no → GET completo (con `If-None-Match` cuando corresponda) → guardar `NormaFull` → proyectar. **Dependencia de implementación**: el caché debe aterrizar primero (o en el mismo apply) — esta tool no implementa el caché, solo lo consume.

**3. Campo `CategoriasNorma` en `Metadatos`**
`[]string` con tag `json:"categorias_norma"` — aditivo, sin breaking. Se entrega tal cual viene de la API (sin sanitizar: son etiquetas cortas, y el structured lleva datos completos por la norma del proyecto; el texto formateado puede limpiar si hiciera falta).

**4. LLM-first + structured (norma heredada del config.yaml)**
`get_law_summary` devuelve `TextContent` (resumen formateado: título, fuente, materias/categorías como listas, resumen oficial) + `structuredContent` tipado `NormaSummary` (schema autogenerado). Texto truncado a lo útil para el modelo (p.ej. resumen oficial completo — es corto por naturaleza); el structured lleva los datos íntegros.

**5. Interfaz y mocks**
`LawClient` gana `GetNormaSummary` → regenerar `law_client_mock.go` con `make mock`. Tests de la tool con suite + `MockLawClient` (convención `mock.Anything` para ctx); tests del cliente con httptest + fixture real (`testdata/norma_full.json` ya existe; el response de la Ley 21.214 del ejemplo sirve como segundo fixture opcional para `categorias_norma`).

## Risks / Trade-offs

- [El summary descarga el response completo igual que `get_law` (la API no tiene endpoint liviano)] → Aceptado y explícito: el ahorro real es para el cliente (payload pequeño + cero conversión). Con el caché compartido, el costo de red desaparece tras la primera llamada.
- [Dependencia de orden con el change add-bcn-leychile-tools (caché ETag aún no implementado)] → Mitigación: la tool consume el caché si existe; el apply de este change asume que el caché aterrizó (o aterriza en el mismo ciclo). Si se aplicara antes, el summary funciona sin caché (correcto, solo menos óptimo).
- [Dos changes con deltas sobre la misma capability no archivada] → Mitigación: este change usa solo requirements ADDED (no MODIFIED) y se archiva DESPUÉS de add-bcn-leychile-tools — documentado en el proposal.

## Migration Plan

Aditivo: nueva tool + campo de metadatos. El mock se regenera (`make mock`) y los tests existentes no cambian salvo los que verifiquen el mock regenerado. Sin rollback especial.

## Open Questions

<!-- Ninguna: campos del summary, reutilización del recurso, caché compartido y convenciones heredadas quedaron definidos con el usuario y el config.yaml del proyecto. -->
