# Graph Report - chile-bcn-mcp  (2026-08-15)

## Corpus Check
- 65 files · ~86,796 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 594 nodes · 1301 edges · 40 communities (32 shown, 8 thin omitted)
- Extraction: 68% EXTRACTED · 32% INFERRED · 0% AMBIGUOUS · INFERRED: 416 edges (avg confidence: 0.83)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `41a43d34`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- openspec-apply-change/SKILL.md
- openspec-explore/SKILL.md
- NewClient
- law-prompts/spec.md
- SanitizeSuite
- leychile-search/spec.md
- release-distributions/spec.md
- Project context — shown to the AI when creating artifacts.
- <capability> Specification
- NormaFull
- PromptsSuite
- context.Context
- Load
- container-deployment/spec.md
- openspec/specs/mcp-server/spec.md
- get_law.go
- Chile BCN MCP Server
- GetLawSuite
- LeyChile API resources contract.
- etagCache
- The ONLY release path: a PR from release/v* merged into main.
- CLAUDE.md
- testing.T
- github.com/stretchr/testify/suite.Suite
- resources
- NormaFull
- Mockery v3 configuration.
- smoke.sh
- negative_timeout.yaml
- valid.yaml
- docker-compose.yml
- build-dist.sh
- github.com/alvarosdev/chile-bcn-mcp

## God Nodes (most connected - your core abstractions)
1. `NewClient()` - 30 edges
2. `LawClientSuite` - 27 edges
3. `GetLawSuite` - 21 edges
4. `Client` - 16 edges
5. `Chile BCN MCP Server` - 16 edges
6. `PromptsSuite` - 15 edges
7. `<capability> Specification` - 15 edges
8. `<capability> Specification` - 14 edges
9. `SanitizeSuite` - 13 edges
10. `SanitizeMarkdown()` - 13 edges

## Surprising Connections (you probably didn't know these)
- `version` --semantically_similar_to--> `version`  [INFERRED] [semantically similar]
  .github/dependabot.yml → config/api.resources.yaml
- `main()` --calls--> `RegisterPrompts()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/prompts/prompts.go
- `main()` --calls--> `Load()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/config/resources.go
- `main()` --calls--> `RegisterTools()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/tools/tools.go
- `main()` --calls--> `NewClient()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/bcn/law_client.go

## Import Cycles
- None detected.

## Communities (40 total, 8 thin omitted)

### Community 0 - "openspec-apply-change/SKILL.md"
Cohesion: 0.25
Nodes (14): Completed This Session, Implementation Complete, Implementation Paused, Implementing: <change-name> (schema: <schema-name>), Issue Encountered, Archive Complete, Archive Complete (with warnings), Archive Failed (+6 more)

### Community 1 - "openspec-explore/SKILL.md"
Cohesion: 0.27
Nodes (20): Check for context, Ending Discovery, Guardrails, OpenSpec Awareness, The Stance, What You Don't Have To Do, What You Might Do, When a change exists (+12 more)

### Community 6 - "NewClient"
Cohesion: 0.20
Nodes (5): FlexInt, LawClientSuite, Pagination, log/slog.Logger, NewClient()

### Community 8 - "law-prompts/spec.md"
Cohesion: 0.31
Nodes (17): analyze_law, compare_law_versions, prompts/list, Purpose, Requirement: Disclaimer en explicaciones simples, Requirement: Prompt de flujo de investigación, Requirement: Prompts curados expuestos, Requirement: Prompts sin llamadas externas (+9 more)

### Community 10 - "SanitizeSuite"
Cohesion: 0.09
Nodes (18): SanitizeSuite, github.com/JohannesKaufmann/html-to-markdown/v2/converter.Converter, testing.B, testing.F, testing.TB, FuzzConvertContent(), FuzzSanitizeMarkdown(), newConverter() (+10 more)

### Community 11 - "leychile-search/spec.md"
Cohesion: 0.33
Nodes (16): IDNORMA, Purpose, Requirement: Búsqueda paginada de normas, Requirement: Contenido de norma por identificador, Requirements, Scenario: Búsqueda simple, Scenario: Cadena vacía, Scenario: Identificador faltante (+8 more)

### Community 12 - "release-distributions/spec.md"
Cohesion: 0.37
Nodes (14): Purpose, Requirement: Distribuciones cross-platform en zip autocontenido, Requirement: Gate de release por merge, Requirement: Release draft versionado, Requirements, Scenario: Cerrado sin merge, Scenario: Dispatch manual, Scenario: Distribución ejecutable tras extraer (+6 more)

### Community 13 - "Project context — shown to the AI when creating artifacts."
Cohesion: 0.67
Nodes (3): context, Project context — shown to the AI when creating artifacts., schema

### Community 14 - "<capability> Specification"
Cohesion: 0.23
Nodes (28): ADDED Requirements, <capability> Specification, <capability> Specification, MODIFIED Requirements, Purpose, REMOVED Requirements, RENAMED Requirements, Requirement: Deprecated Feature (+20 more)

### Community 17 - "PromptsSuite"
Cohesion: 0.11
Nodes (5): RegisterPrompts(), TestPromptsSuite(), ToolNames(), PromptsSuite, template

### Community 18 - "context.Context"
Cohesion: 0.06
Nodes (40): Client, EstructuraPart, HistoriaEntrada, HistoriaGrupo, HtmlBlock, Metadatos, MockLawClient_Expecter, MockLawClient_GetLawHistory_Call (+32 more)

### Community 19 - "Load"
Cohesion: 0.19
Nodes (9): CircuitBreaker, Duration, ResourcesSuite, Retry, Resource, Resources, Load(), TestResourcesSuite() (+1 more)

### Community 20 - "container-deployment/spec.md"
Cohesion: 0.35
Nodes (15): Purpose, Requirement: Configuración por entorno en el contenedor, Requirement: Ejecución como usuario no-root, Requirement: Healthcheck del contenedor, Requirement: Imagen OCI multi-arquitectura, Requirement: Validación automática en CI, Requirements, Scenario: Cambios válidos (+7 more)

### Community 22 - "openspec/specs/mcp-server/spec.md"
Cohesion: 0.27
Nodes (20): FASTMCP_HOST:FASTMCP_PORT, FASTMCP_TRANSPORT, FASTMCP_TRANSPORT=http, FASTMCP_TRANSPORT=stdio, /mcp, Purpose, Requirement: Autenticación opcional por token Bearer, Requirement: Configuración por variables de entorno (+12 more)

### Community 23 - "get_law.go"
Cohesion: 0.08
Nodes (45): NormTreeSuite, github.com/modelcontextprotocol/go-sdk/mcp.CallToolResult, github.com/modelcontextprotocol/go-sdk/mcp.Server, github.com/modelcontextprotocol/go-sdk/mcp.ToolHandlerFor, strings.Builder, LawClient, ContentCharCount(), FlattenStructure() (+37 more)

### Community 26 - "Chile BCN MCP Server"
Cohesion: 0.33
Nodes (17): Available Tools, Build a static binary (outputs to bin/), Chile BCN MCP Server, Container, agent-launched (stdio), Container, self-hosted (HTTP), Cross-platform distributions (6 targets, self-contained with config), Environment Variables, Chile BCN MCP Server (+9 more)

### Community 31 - "LeyChile API resources contract."
Cohesion: 0.15
Nodes (13): breaker, buscadorNormas, ETag, LeyChile API resources contract., getNorma, getNormaHistory, resources, retry (+5 more)

### Community 32 - "etagCache"
Cohesion: 0.27
Nodes (9): cacheItem, etagCache, etagCache[T], etagEntry, container/list.Element, container/list.List, sync.Mutex, newEtagCache() (+1 more)

### Community 33 - "The ONLY release path: a PR from release/v* merged into main."
Cohesion: 0.22
Nodes (10): jobs, name, on, permissions, env, The ONLY release path: a PR from release/v* merged into main., jobs, name (+2 more)

### Community 34 - "CLAUDE.md"
Cohesion: 0.53
Nodes (8): Architecture, Commands, get_law, get_law_summary, Gotchas (aprendidas con dolor — no repetir), graphify, search_laws, What this is

### Community 36 - "testing.T"
Cohesion: 0.08
Nodes (19): NormaCacheSuite, main(), runHTTP(), staticTokenAuthMiddleware(), net/http.Handler, testing.T, TestFlexIntSuite(), TestLawClientSuite() (+11 more)

### Community 38 - "github.com/stretchr/testify/suite.Suite"
Cohesion: 0.10
Nodes (9): FlexIntSuite, github.com/modelcontextprotocol/go-sdk/mcp.ClientSession, github.com/stretchr/testify/mock.TestingT, github.com/stretchr/testify/suite.Suite, NewMockLawClient(), newTestClient(), GetLawHistorySuite, GetLawSummarySuite (+1 more)

### Community 39 - "resources"
Cohesion: 0.33
Nodes (6): resources, version, resources, version, resources, version

### Community 50 - "Mockery v3 configuration."
Cohesion: 0.50
Nodes (4): all, Mockery v3 configuration., packages, template

### Community 51 - "smoke.sh"
Cohesion: 0.83
Nodes (3): fail(), smoke.sh script, step()

## Knowledge Gaps
- **26 isolated node(s):** `template`, `normType`, `build-dist.sh script`, `github.com/alvarosdev/chile-bcn-mcp`, `context` (+21 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **8 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewClient()` connect `NewClient` to `etagCache`, `testing.T`, `SanitizeSuite`, `context.Context`, `Load`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **Why does `LawClient` connect `get_law.go` to `context.Context`, `github.com/stretchr/testify/suite.Suite`?**
  _High betweenness centrality (0.038) - this node is a cross-community bridge._
- **Why does `newTestClient()` connect `github.com/stretchr/testify/suite.Suite` to `context.Context`, `testing.T`, `get_law.go`?**
  _High betweenness centrality (0.037) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `NewClient()` (e.g. with `.TestCircuitBreakerOpensAfterFailures()` and `.TestGetLawHistoryEmptyResult()`) actually correct?**
  _`NewClient()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `template`, `normType`, `build-dist.sh script` to the rest of the system?**
  _26 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `SanitizeSuite` be split into smaller, more focused modules?**
  _Cohesion score 0.0907563025210084 - nodes in this community are weakly interconnected._
- **Should `PromptsSuite` be split into smaller, more focused modules?**
  _Cohesion score 0.11 - nodes in this community are weakly interconnected._