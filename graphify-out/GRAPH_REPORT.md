# Graph Report - chile-bcn-mcp  (2026-08-15)

## Corpus Check
- 108 files · ~111,615 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1090 nodes · 2650 edges · 65 communities (52 shown, 13 thin omitted)
- Extraction: 50% EXTRACTED · 50% INFERRED · 0% AMBIGUOUS · INFERRED: 1321 edges (avg confidence: 0.83)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `375797d5`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- 2026-08-14-add-law-history-and-versions/specs/leychile-search/spec.md
- openspec-explore/SKILL.md
- 2026-08-14-remove-echo-tool/design.md
- 2026-08-13-scaffold-mcp-server/specs/mcp-server/spec.md
- 2026-08-14-add-law-prompts/design.md
- 2026-08-14-add-law-prompts/proposal.md
- NewClient
- 2026-08-14-add-law-prompts/tasks.md
- openspec/specs/law-prompts/spec.md
- 2026-08-14-add-container-and-ci/proposal.md
- SanitizeSuite
- openspec/specs/leychile-search/spec.md
- 2026-08-14-add-release-workflow/specs/release-distributions/spec.md
- 2026-08-14-add-bcn-leychile-tools/design.md
- <capability> Specification
- NormaFull
- 2026-08-14-add-bcn-leychile-tools/tasks.md
- PromptsSuite
- context.Context
- Load
- openspec/specs/container-deployment/spec.md
- 2026-08-15-add-law-section-retrieval/design.md
- openspec/specs/mcp-server/spec.md
- get_law.go
- 2026-08-14-add-law-summary-tool/specs/leychile-search/spec.md
- 2026-08-15-add-law-section-retrieval/tasks.md
- Chile BCN MCP Server
- GetLawSuite
- Requirement: Paginación tolerante a tipos inconsistentes del servicio
- testing.T
- 2026-08-15-fix-pagination-flexint/proposal.md
- LeyChile API resources contract.
- etagCache
- The ONLY release path: a PR from release/v* merged into main.
- CLAUDE.md
- 2026-08-13-scaffold-mcp-server/.openspec.yaml
- NormaCacheSuite
- github.com/stretchr/testify/suite.Suite
- newTestClient
- resources
- GetLawSummarySuite
- SearchLawsSuite
- 2026-08-14-add-law-history-and-versions/.openspec.yaml
- 2026-08-15-fix-pagination-flexint/design.md
- GetLawHistorySuite
- 2026-08-15-fix-pagination-flexint/tasks.md
- 2026-08-14-add-release-workflow/proposal.md
- 2026-08-14-remove-echo-tool/.openspec.yaml
- NormaFull
- Decisions
- Mockery v3 configuration.
- smoke.sh
- negative_timeout.yaml
- valid.yaml
- 2026-08-14-add-release-workflow/.openspec.yaml
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
- `main()` --calls--> `RegisterTools()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/tools/tools.go
- `main()` --calls--> `Load()`  [EXTRACTED]
  cmd/chile-bcn-mcp/main.go → internal/config/resources.go
- `get_law` --semantically_similar_to--> `get_law`  [INFERRED] [semantically similar]
  openspec/changes/archive/2026-08-14-add-law-history-and-versions/design.md → openspec/changes/archive/2026-08-14-add-law-history-and-versions/specs/leychile-search/spec.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Change 2026-08-13-scaffold-mcp-server** — openspec_changes_archive_2026_08_13_scaffold_mcp_server_specs_mcp_server_spec_file, openspec_changes_archive_2026_08_13_scaffold_mcp_server_proposal_file, openspec_changes_archive_2026_08_13_scaffold_mcp_server_specs_mcp_server_spec_fastmcp_host_fastmcp_port, openspec_changes_archive_2026_08_13_scaffold_mcp_server_specs_mcp_server_spec_scenario_mismas_tools_en_ambos_transportes, openspec_changes_archive_2026_08_13_scaffold_mcp_server_design_decisions, openspec_changes_archive_2026_08_13_scaffold_mcp_server_design_context, openspec_changes_archive_2026_08_13_scaffold_mcp_server_specs_mcp_server_spec_scenario_consulta_de_salud, openspec_changes_archive_2026_08_13_scaffold_mcp_server_specs_mcp_server_spec_scenario_arranque_con_transporte_http [EXTRACTED 1.00]
- **Change 2026-08-14-add-bcn-leychile-tools** — openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_openspec_schema, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_design_get_norma_json, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_design_context, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_tasks_2_config_yaml, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_specs_leychile_search_spec_scenario_b_squeda_simple, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_tasks_5_bootstrap_y_verificaci_n, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_design_file, openspec_changes_archive_2026_08_14_add_bcn_leychile_tools_openspec_created [EXTRACTED 1.00]
- **Change 2026-08-14-add-container-and-ci** — openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_scenario_cambios_v_lidos, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_scenario_healthcheck_del_contenedor, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_purpose, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_requirement_validaci_n_autom_tica_en_ci, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_requirement_imagen_oci_multi_arquitectura, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_added_requirements, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_requirement_ejecuci_n_como_usuario_no_root, openspec_changes_archive_2026_08_14_add_container_and_ci_specs_container_deployment_spec_scenario_cambios_que_rompen_tests [EXTRACTED 1.00]
- **Change 2026-08-14-add-container-and-ci** — openspec_changes_archive_2026_08_14_add_container_and_ci_tasks_1_docker, openspec_changes_archive_2026_08_14_add_container_and_ci_proposal_file, openspec_changes_archive_2026_08_14_add_container_and_ci_proposal_capabilities, openspec_changes_archive_2026_08_14_add_container_and_ci_proposal_chile_bcn_mcp, openspec_changes_archive_2026_08_14_add_container_and_ci_proposal_fastmcp, openspec_changes_archive_2026_08_14_add_container_and_ci_design_context, openspec_changes_archive_2026_08_14_add_container_and_ci_design_godot_mcp_docs, openspec_changes_archive_2026_08_14_add_container_and_ci_design_goals_non_goals [EXTRACTED 1.00]
- **Change 2026-08-14-add-law-history-and-versions** — openspec_changes_archive_2026_08_14_add_law_history_and_versions_design_decisions, openspec_changes_archive_2026_08_14_add_law_history_and_versions_design_idnorma, openspec_changes_archive_2026_08_14_add_law_history_and_versions_tasks_file, openspec_changes_archive_2026_08_14_add_law_history_and_versions_specs_leychile_search_spec_requirement_cach_con_revalidaci_n_por_versi_n_e_historia, openspec_changes_archive_2026_08_14_add_law_history_and_versions_design_search_laws, openspec_changes_archive_2026_08_14_add_law_history_and_versions_specs_leychile_search_spec_scenario_norma_sin_historia, openspec_changes_archive_2026_08_14_add_law_history_and_versions_specs_leychile_search_spec_scenario_identificador_faltante, openspec_changes_archive_2026_08_14_add_law_history_and_versions_tasks_internal_bcn_norm_types_test_go [EXTRACTED 1.00]
- **Change 2026-08-14-add-law-prompts** — openspec_changes_archive_2026_08_14_add_law_prompts_design_internal_prompts_prompts_go, openspec_changes_archive_2026_08_14_add_law_prompts_tasks_registerprompts_srv, openspec_changes_archive_2026_08_14_add_law_prompts_tasks_compare_law_versions, openspec_changes_archive_2026_08_14_add_law_prompts_proposal_capabilities, openspec_changes_archive_2026_08_14_add_law_prompts_proposal_trace_law_history_norm_id, openspec_changes_archive_2026_08_14_add_law_prompts_tasks_trace_law_history, openspec_changes_archive_2026_08_14_add_law_prompts_design_prompts_get, openspec_changes_archive_2026_08_14_add_law_prompts_proposal_prompts_get [EXTRACTED 1.00]
- **Change 2026-08-14-add-release-workflow** — openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_scenario_distribuci_n_ejecutable_tras_extraer, openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_purpose, openspec_changes_archive_2026_08_14_add_release_workflow_specs_container_deployment_spec_scenario_sin_merge_no_hay_imagen, openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_scenario_cerrado_sin_merge, openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_requirement_distribuciones_cross_platform_en_zip_autoconteni, openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_scenario_zip_completo, openspec_changes_archive_2026_08_14_add_release_workflow_specs_release_distributions_spec_scenario_release_solo_con_pipeline_verde, openspec_changes_archive_2026_08_14_add_release_workflow_specs_container_deployment_spec_file [EXTRACTED 1.00]
- **Change 2026-08-14-remove-echo-tool** — openspec_changes_archive_2026_08_14_remove_echo_tool_design_file, openspec_changes_archive_2026_08_14_remove_echo_tool_design_get_law_summary, openspec_changes_archive_2026_08_14_remove_echo_tool_proposal_capabilities, openspec_changes_archive_2026_08_14_remove_echo_tool_specs_mcp_server_spec_file, openspec_changes_archive_2026_08_14_remove_echo_tool_openspec_created, openspec_changes_archive_2026_08_14_remove_echo_tool_proposal_modified_capabilities, openspec_changes_archive_2026_08_14_remove_echo_tool_design_context, openspec_changes_archive_2026_08_14_remove_echo_tool_specs_mcp_server_spec_get_law [EXTRACTED 1.00]
- **Change 2026-08-15-add-law-section-retrieval** — openspec_changes_archive_2026_08_15_add_law_section_retrieval_specs_law_prompts_spec_search_legal_topic, openspec_changes_archive_2026_08_15_add_law_section_retrieval_openspec_schema, openspec_changes_archive_2026_08_15_add_law_section_retrieval_design_5_descripciones_que_ense_an_el_flujo, openspec_changes_archive_2026_08_15_add_law_section_retrieval_design_4_validaci_n_fail_fast_de_section_id, openspec_changes_archive_2026_08_15_add_law_section_retrieval_specs_leychile_search_spec_scenario_secci_n_espec_fica, openspec_changes_archive_2026_08_15_add_law_section_retrieval_design_decisions, openspec_changes_archive_2026_08_15_add_law_section_retrieval_design_7_pool_de_converters_paralelizar_conversiones, openspec_changes_archive_2026_08_15_add_law_section_retrieval_specs_law_prompts_spec_requirement_prompts_curados_expuestos [EXTRACTED 1.00]

## Communities (65 total, 13 thin omitted)

### Community 0 - "2026-08-14-add-law-history-and-versions/specs/leychile-search/spec.md"
Cohesion: 0.11
Nodes (47): ADDED Requirements, Purpose, Requirement: Configuración por entorno en el contenedor, Requirement: Ejecución como usuario no-root, Requirement: Healthcheck del contenedor, Requirement: Imagen OCI multi-arquitectura, Requirement: Validación automática en CI, Scenario: Cambios que rompen tests (+39 more)

### Community 1 - "openspec-explore/SKILL.md"
Cohesion: 0.13
Nodes (34): Completed This Session, Implementation Complete, Implementation Paused, Implementing: <change-name> (schema: <schema-name>), Issue Encountered, Archive Complete, Archive Complete (with warnings), Archive Failed (+26 more)

### Community 2 - "2026-08-14-remove-echo-tool/design.md"
Cohesion: 0.14
Nodes (36): Context, Decisions, get_law, get_law_summary, Goals / Non-Goals, Migration Plan, Open Questions, openspec/specs/mcp-server/spec.md (+28 more)

### Community 3 - "2026-08-13-scaffold-mcp-server/specs/mcp-server/spec.md"
Cohesion: 0.15
Nodes (36): ADDED Requirements, FASTMCP_HOST:FASTMCP_PORT, FASTMCP_TRANSPORT, FASTMCP_TRANSPORT=http, FASTMCP_TRANSPORT=stdio, /mcp, Purpose, Requirement: Autenticación opcional por token Bearer (+28 more)

### Community 4 - "2026-08-14-add-law-prompts/design.md"
Cohesion: 0.17
Nodes (34): /Consulta/getTiposNorma, Context, Decisions, get_law, get_law_summary, Goals / Non-Goals, idNorma, Migration Plan (+26 more)

### Community 5 - "2026-08-14-add-law-prompts/proposal.md"
Cohesion: 0.17
Nodes (34): Capabilities, get_law, get_law_history(norm_id), get_law_summary, Impact, Modified Capabilities, New Capabilities, What Changes (+26 more)

### Community 6 - "NewClient"
Cohesion: 0.14
Nodes (12): FlexInt, LawClientSuite, main(), runHTTP(), staticTokenAuthMiddleware(), log/slog.Logger, net/http.Handler, NewClient() (+4 more)

### Community 7 - "2026-08-14-add-law-prompts/tasks.md"
Cohesion: 0.14
Nodes (32): 1. Catálogo de tipos, 2. Modelo y cliente, 3. Tools MCP, 4. Verificación y documentación, /Consulta/getTiposNorma, etagCache[T], historias etagCache[[]HistoriaGrupo], internal/bcn/norm_types.go (+24 more)

### Community 8 - "openspec/specs/law-prompts/spec.md"
Cohesion: 0.16
Nodes (32): ADDED Requirements, Requirement: Contenedor endurecido, Scenario: Capacidades mínimas, Scenario: Rootfs de solo lectura, ADDED Requirements, analyze_law, compare_law_versions, MODIFIED Requirements (+24 more)

### Community 9 - "2026-08-14-add-container-and-ci/proposal.md"
Cohesion: 0.18
Nodes (31): Capabilities, chile-bcn-mcp, cmd/chile-bcn-mcp/main.go, FASTMCP_TRANSPORT, godot-mcp-docs, Impact, Modified Capabilities, New Capabilities (+23 more)

### Community 10 - "SanitizeSuite"
Cohesion: 0.12
Nodes (15): SanitizeSuite, testing.B, testing.F, testing.TB, FuzzConvertContent(), FuzzSanitizeMarkdown(), benchInput(), BenchmarkNormalize() (+7 more)

### Community 11 - "openspec/specs/leychile-search/spec.md"
Cohesion: 0.18
Nodes (31): ADDED Requirements, get_law, idNorma, MODIFIED Requirements, Requirement: Contenido de norma por identificador, Requirement: Resumen de norma por identificador, Scenario: Identificador faltante, Scenario: Identificador inexistente (+23 more)

### Community 12 - "2026-08-14-add-release-workflow/specs/release-distributions/spec.md"
Cohesion: 0.21
Nodes (28): ADDED Requirements, Purpose, Requirement: Distribuciones cross-platform en zip autocontenido, Requirement: Gate de release por merge, Requirement: Release draft versionado, Scenario: Cerrado sin merge, Scenario: Dispatch manual, Scenario: Distribución ejecutable tras extraer (+20 more)

### Community 13 - "2026-08-14-add-bcn-leychile-tools/design.md"
Cohesion: 0.21
Nodes (26): chile-bcn-mcp, Context, Decisions, Goals / Non-Goals, godot-mcp-docs, Migration Plan, Open Questions, Risks / Trade-offs (+18 more)

### Community 14 - "<capability> Specification"
Cohesion: 0.23
Nodes (28): ADDED Requirements, <capability> Specification, <capability> Specification, MODIFIED Requirements, Purpose, REMOVED Requirements, RENAMED Requirements, Requirement: Deprecated Feature (+20 more)

### Community 16 - "2026-08-14-add-bcn-leychile-tools/tasks.md"
Cohesion: 0.20
Nodes (22): 1. Setup del módulo, 2. Server core, 3. Tools, 4. Entrypoint, 5. Tooling y documentación, 6. Verificación final, cmd/chile-bcn-mcp/, Config (+14 more)

### Community 17 - "PromptsSuite"
Cohesion: 0.12
Nodes (4): RegisterPrompts(), ToolNames(), PromptsSuite, template

### Community 18 - "context.Context"
Cohesion: 0.05
Nodes (43): Client, EstructuraPart, HistoriaEntrada, HistoriaGrupo, HtmlBlock, Metadatos, MockLawClient_Expecter, MockLawClient_GetLawHistory_Call (+35 more)

### Community 19 - "Load"
Cohesion: 0.22
Nodes (8): CircuitBreaker, Duration, ResourcesSuite, Retry, Resource, Resources, Load(), yaml.Node

### Community 20 - "openspec/specs/container-deployment/spec.md"
Cohesion: 0.25
Nodes (20): MODIFIED Requirements, Requirement: Imagen OCI multi-arquitectura, Scenario: Publicación manual, Scenario: Publicación por release versionado, Scenario: Sin merge no hay imagen, Purpose, Requirement: Configuración por entorno en el contenedor, Requirement: Ejecución como usuario no-root (+12 more)

### Community 21 - "2026-08-15-add-law-section-retrieval/design.md"
Cohesion: 0.27
Nodes (20): 10. Tope de tamaño de respuesta upstream, 11. Contenedor endurecido, 1. Recorte local sobre el árbol ya parseado (no filtrado en la fuente), 2. `char_count` y `article_count` derivados del documento, no de un render aparte, 3. Ubicación: operaciones de árbol en `bcn`, presentación en `tools`, 4. Validación fail-fast de `section_id`, 5. Descripciones que enseñan el flujo, 6. Séptimo prompt `law_research_workflow` (+12 more)

### Community 22 - "openspec/specs/mcp-server/spec.md"
Cohesion: 0.27
Nodes (20): FASTMCP_HOST:FASTMCP_PORT, FASTMCP_TRANSPORT, FASTMCP_TRANSPORT=http, FASTMCP_TRANSPORT=stdio, /mcp, Purpose, Requirement: Autenticación opcional por token Bearer, Requirement: Configuración por variables de entorno (+12 more)

### Community 23 - "get_law.go"
Cohesion: 0.08
Nodes (44): NormTreeSuite, github.com/modelcontextprotocol/go-sdk/mcp.CallToolResult, github.com/modelcontextprotocol/go-sdk/mcp.Server, github.com/modelcontextprotocol/go-sdk/mcp.ToolHandlerFor, strings.Builder, LawClient, ContentCharCount(), FlattenStructure() (+36 more)

### Community 24 - "2026-08-14-add-law-summary-tool/specs/leychile-search/spec.md"
Cohesion: 0.40
Nodes (13): ADDED Requirements, categorias_norma, get_law, get_law_summary, idNorma, Requirement: Categorías de norma en metadatos, Requirement: Resumen de norma por identificador, Scenario: Identificador faltante (+5 more)

### Community 25 - "2026-08-15-add-law-section-retrieval/tasks.md"
Cohesion: 0.29
Nodes (15): 1. Remoción de echo, 2. Smoke test, 1. Operaciones de árbol en internal/bcn, 2. Tools get_law y get_law_summary, 3. Prompt law_research_workflow, 4. Tests, 5. Verificación, 6. Robustez del cliente BCN (+7 more)

### Community 26 - "Chile BCN MCP Server"
Cohesion: 0.33
Nodes (17): Available Tools, Build a static binary (outputs to bin/), Chile BCN MCP Server, Container, agent-launched (stdio), Container, self-hosted (HTTP), Cross-platform distributions (6 targets, self-contained with config), Environment Variables, Chile BCN MCP Server (+9 more)

### Community 28 - "Requirement: Paginación tolerante a tipos inconsistentes del servicio"
Cohesion: 0.22
Nodes (8): ADDED Requirements, Requirement: Paginación tolerante a tipos inconsistentes del servicio, Scenario: Contrato de la tool sin cambios, Scenario: Formatos mixtos y valores vacíos, Scenario: Paginación como string (regresión), Scenario: Paginación numérica, Scenario: Valor no numérico falla explícito, Scenario: Variantes de formato numérico

### Community 29 - "testing.T"
Cohesion: 0.15
Nodes (10): testing.T, TestLawClientSuite(), TestNormTreeSuite(), TestSanitizeSuite(), TestResourcesSuite(), TestPromptsSuite(), TestLoadConfigDefaults(), TestLoadConfigOverrides() (+2 more)

### Community 30 - "2026-08-15-fix-pagination-flexint/proposal.md"
Cohesion: 0.29
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

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

### Community 35 - "2026-08-13-scaffold-mcp-server/.openspec.yaml"
Cohesion: 0.33
Nodes (7): created, schema, created, schema, created, schema, schema

### Community 37 - "github.com/stretchr/testify/suite.Suite"
Cohesion: 0.29
Nodes (3): FlexIntSuite, github.com/stretchr/testify/suite.Suite, TestFlexIntSuite()

### Community 38 - "newTestClient"
Cohesion: 0.31
Nodes (4): github.com/modelcontextprotocol/go-sdk/mcp.ClientSession, github.com/stretchr/testify/mock.TestingT, NewMockLawClient(), newTestClient()

### Community 39 - "resources"
Cohesion: 0.33
Nodes (6): resources, version, resources, version, resources, version

### Community 42 - "2026-08-14-add-law-history-and-versions/.openspec.yaml"
Cohesion: 0.42
Nodes (6): created, schema, created, schema, created, schema

### Community 43 - "2026-08-15-fix-pagination-flexint/design.md"
Cohesion: 0.33
Nodes (5): Context, Decisions, Goals / Non-Goals, Migration Plan, Risks / Trade-offs

### Community 45 - "2026-08-15-fix-pagination-flexint/tasks.md"
Cohesion: 0.40
Nodes (4): 1. Tipo FlexInt, 2. Migración de Pagination, 3. Tests de integración (sin red), 4. Verificación

### Community 46 - "2026-08-14-add-release-workflow/proposal.md"
Cohesion: 0.71
Nodes (6): Capabilities, Impact, Modified Capabilities, New Capabilities, What Changes, Why

### Community 47 - "2026-08-14-remove-echo-tool/.openspec.yaml"
Cohesion: 0.47
Nodes (4): created, schema, created, schema

### Community 49 - "Decisions"
Cohesion: 0.90
Nodes (4): config/api.resources.yaml, Context, Decisions, Goals / Non-Goals

### Community 50 - "Mockery v3 configuration."
Cohesion: 0.50
Nodes (4): all, Mockery v3 configuration., packages, template

### Community 51 - "smoke.sh"
Cohesion: 0.83
Nodes (3): fail(), smoke.sh script, step()

## Knowledge Gaps
- **47 isolated node(s):** `Context`, `Goals / Non-Goals`, `Decisions`, `Risks / Trade-offs`, `Migration Plan` (+42 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **13 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `NewClient()` connect `NewClient` to `etagCache`, `context.Context`, `Load`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Why does `newTestClient()` connect `newTestClient` to `context.Context`, `testing.T`, `get_law.go`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `LawClient` connect `get_law.go` to `context.Context`, `newTestClient`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `NewClient()` (e.g. with `.TestCircuitBreakerOpensAfterFailures()` and `.TestGetLawHistoryEmptyResult()`) actually correct?**
  _`NewClient()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `Context`, `Goals / Non-Goals`, `Decisions` to the rest of the system?**
  _47 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `2026-08-14-add-law-history-and-versions/specs/leychile-search/spec.md` be split into smaller, more focused modules?**
  _Cohesion score 0.11346938775510204 - nodes in this community are weakly interconnected._
- **Should `openspec-explore/SKILL.md` be split into smaller, more focused modules?**
  _Cohesion score 0.13205128205128205 - nodes in this community are weakly interconnected._