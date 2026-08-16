## ADDED Requirements

### Requirement: Paginación tolerante a tipos inconsistentes del servicio

La búsqueda DEBE decodificar el bloque de paginación que entrega el servicio aceptando indistintamente string (`"10"`) o número (`10`) en cada uno de sus campos numéricos (`npagina`, `itemsporpagina`, `totalitems`), incluyendo combinaciones de ambos formatos en una misma respuesta. La decodificación DEBE además tolerar números con espacios (`" 10 "`), decimales de parte entera (`10.0`), cadena vacía (`""`) y `null` — estos dos últimos interpretados como 0 — sin fallar la búsqueda. Un valor no numérico (p.ej. `"abc"`) DEBE fallar la búsqueda con un error de decodificación explícito, nunca silenciarse como 0. El contrato de la tool hacia el cliente NO cambia: `total_items`, `page`, `page_size` y `total_pages` del contenido estructurado SIGUEN siendo números.

#### Scenario: Paginación numérica
- **WHEN** el servicio responde la búsqueda con `npagina`, `itemsporpagina` y `totalitems` como números (p.ej. la búsqueda `Ley 21461`)
- **THEN** la búsqueda retorna los resultados y el total sin error de decodificación, con el `norm_id` esperado entre los resultados

#### Scenario: Paginación como string (regresión)
- **WHEN** el servicio responde la búsqueda con `npagina`, `itemsporpagina` y `totalitems` como strings (p.ej. la búsqueda `Ley 21.600`)
- **THEN** la búsqueda sigue retornando los resultados y el total sin error de decodificación

#### Scenario: Formatos mixtos y valores vacíos
- **WHEN** la respuesta combina formatos en el mismo bloque (p.ej. `"npagina": 1` junto a `"itemsporpagina": "10"`) o trae `""` o `null` en un campo numérico
- **THEN** la búsqueda no falla por decodificación y los campos vacíos o nulos se interpretan como 0

#### Scenario: Variantes de formato numérico
- **WHEN** un campo numérico llega como string con espacios (`" 10 "`) o como decimal de parte entera (`10.0`)
- **THEN** la búsqueda lo interpreta como 10 sin error

#### Scenario: Valor no numérico falla explícito
- **WHEN** un campo numérico llega como un string sin contenido numérico (p.ej. `"abc"`)
- **THEN** la búsqueda falla con un error de decodificación explícito en lugar de interpretarlo como 0

#### Scenario: Contrato de la tool sin cambios
- **WHEN** la búsqueda tiene éxito con cualquiera de los formatos anteriores
- **THEN** el contenido estructurado entrega `total_items`, `page`, `page_size` y `total_pages` como números
