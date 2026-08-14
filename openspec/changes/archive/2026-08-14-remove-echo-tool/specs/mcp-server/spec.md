## REMOVED Requirements

### Requirement: Tool demo echo

**Reason**: La tool `echo` fue el placeholder del scaffold para validar transporte, registro de tools y generación de schemas. Esa validación ya está cubierta por las tools reales (`search_laws`, `get_law`, `get_law_summary`) y por las suites de tests con in-memory transports; una tool sin función consume tokens de contexto en cada `tools/list` e invita al modelo a usarla sin razón. El requirement del scaffold la declaró explícitamente "reemplazable por las tools de dominio reales".

**Migration**: Ninguna — la tool no era parte del dominio. Para probar el server de punta a punta usar `scripts/smoke.sh` (`make smoke`), el MCP Inspector o las suites de tests.
