# Chile BCN MCP Server

[![Podman](https://img.shields.io/badge/Podman-First-purple?logo=podman)](https://podman.io/)
[![Docker](https://img.shields.io/badge/Docker-Fallback-blue?logo=docker)](https://www.docker.com/)
[![MCP](https://img.shields.io/badge/MCP-Compatible-green)](https://modelcontextprotocol.io/)
[![Go](https://img.shields.io/badge/Go-blue?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server that gives AI assistants direct access to Chilean laws, decrees and resolutions from **LeyChile** (Biblioteca del Congreso Nacional) and **Contraloría dictámenes** (jurisprudencia administrativa via `contraloria.cl/apibusca`) — in a format designed for how LLMs consume text.
> **Built with Go** using the official [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk). Legal content is served by the public LeyChile API of the Biblioteca del Congreso Nacional.

> ⚠️ **Disclaimer:** this is an independent community project — not affiliated with BCN/LeyChile or any Chilean government institution — and is not intended for production use. See the [Disclaimer](#disclaimer) section for the full terms.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Why Markdown?](#why-markdown)
- [Usage](#usage)
  - [Container, agent-launched (stdio)](#container-agent-launched-stdio)
  - [Container, self-hosted (HTTP)](#container-self-hosted-http)
  - [Native (no container)](#native-no-container)
    - [Running the binary (stdio vs HTTP)](#running-the-binary-stdio-vs-http)
    - [Agent examples: Claude Code, Codex and Hermes](#agent-examples-claude-code-codex-and-hermes)
- [Environment Variables](#environment-variables)
- [Available Tools](#available-tools)
- [Sample Usage](#sample-usage)
- [Prompts](#prompts)
- [Image Tags & Updates](#image-tags--updates)
- [Repository Layout](#repository-layout)
- [FAQ](#faq)
- [Recommended System Prompt](#recommended-system-prompt)
- [Disclaimer](#disclaimer)
- [Release Process](#release-process)
- [License](#license)

---

## Quick Start

Try it in 30 seconds with podman (the preferred runtime — docker works too):

```bash
podman build -t chile-bcn-mcp:local .
podman run -d -p 8000:8000 \
  -e MCP_AUTH_TOKEN=your-token \
  chile-bcn-mcp:local
```

Then point your MCP client at `http://localhost:8000/mcp`:

```json
{
  "mcpServers": {
    "chile-bcn": {
      "type": "http",
      "url": "http://localhost:8000/mcp",
      "headers": { "Authorization": "Bearer your-token" }
    }
  }
}
```

If you didn't set `MCP_AUTH_TOKEN`, omit the `headers` block.

> Prebuilt images on GHCR are **coming soon** — until the first release, build locally as shown above.

---

## Why Markdown?

LeyChile delivers norm content as HTML fragments wrapped in JSON — paragraphs are `<div class="p">` blocks full of `&nbsp;` indentation, HTML entities, and nested article trees. This server converts that markup to clean **Markdown** at request time: entities decoded, indentation normalized, the nested structure (títulos → párrafos → artículos) rendered with real headings, and links to related norms preserved. Formatting garbage is stripped; legal content (quoted inserted text, article wording) is kept verbatim — it carries legal meaning.

Every tool also answers **LLM-first with structured data**: human-readable text in `content[]` for the model, plus a typed `structuredContent` (with an auto-generated JSON schema) for programmatic consumers. The text is a compact view; the structured output carries the complete data.

---

## Usage

### Container, agent-launched (stdio)

The simplest experience: your MCP client launches the container itself and talks to it over stdin/stdout. No port mapping, no daemon to manage — the agent handles the container lifecycle.

Add this to your MCP client config (podman shown; swap `podman` → `docker` if you use Docker):

```json
{
  "mcpServers": {
    "chile-bcn": {
      "command": "podman",
      "args": ["run", "--rm", "-i", "-e", "MCP_TRANSPORT=stdio",
               "chile-bcn-mcp:local"]
    }
  }
}
```

The three flags matter:

- `-i` keeps stdin open so the agent can send messages (required)
- **no** `-t` — a terminal (TTY) would corrupt the protocol (do not add it)
- `--rm` removes the container when the agent closes it

Auth (`MCP_AUTH_TOKEN`) is not used in stdio mode — the container runs locally as a child of your agent, so there's no network boundary to protect.

### Container, self-hosted (HTTP)

Run your own always-on server and connect multiple clients to it. Recommended when you want it hosted separately from a single agent.

**With podman or docker:**

```bash
podman run -d -p 8000:8000 \
  -e MCP_AUTH_TOKEN=your-token \
  chile-bcn-mcp:local
```

**Recommended: Compose.** The repo ships a [`docker-compose.yml`](docker-compose.yml) that wires up the env vars and a healthcheck — podman-compose is the main driver, docker compose is the fallback:

```bash
make compose-up      # podman-compose up -d (docker compose as fallback)
make compose-down
```

The compose file exposes the env vars listed in [Environment Variables](#environment-variables) — edit the `environment:` block or pass them via a `.env` file.

### Native (no container)

For development or when you want a single static binary with no container runtime. Every binary is fully static (`CGO_ENABLED=0`) and self-contained — the LeyChile endpoints contract (`internal/config/api.resources.yaml`) and curated prompts (`internal/prompts/prompts.yaml`) are baked in via `go:embed`. No external files, no Go toolchain at runtime; `cd` into any extracted folder and run the binary directly.

#### Option A — download prebuilt binaries (recommended)

1. Download `dist.zip` from the latest [GitHub Release](../../releases) (attached as a release asset — see [Release Process](#release-process)).
2. Unzip and pick the binary for your platform:

```
dist.zip
├── windows/{amd64,arm64}/chile-bcn-mcp.exe
├── linux/{amd64,arm64}/chile-bcn-mcp
├── darwin/{amd64,arm64}/chile-bcn-mcp     (amd64 = Intel, arm64 = Apple Silicon)
└── SHA256SUMS.txt
```

```bash
unzip dist.zip
# optional: verify checksums
sha256sum -c dist/SHA256SUMS.txt
```

3. Run it — see [Running the binary](#running-the-binary-stdio-vs-http) below.

#### Option B — build from source

```bash
# Single binary for your host (outputs to bin/)
make build
./bin/chile-bcn-mcp            # defaults to HTTP on 127.0.0.1:8000/mcp

# Or cross-compile all 6 targets exactly as CI does (outputs to dist/ + dist.zip)
make dist                      # same script CI runs: scripts/build-dist.sh
```

Dev shortcuts that run without building `bin/`:

```bash
make run-http        # HTTP mode (go run, no auth)
make run-http-auth   # HTTP mode with MCP_AUTH_TOKEN=devtoken (override: DEV_AUTH_TOKEN=my-token make run-http-auth)
make run-stdio       # stdio mode
```

#### Running the binary (stdio vs HTTP)

The same binary speaks both transports — selected by `MCP_TRANSPORT` (see [Environment Variables](#environment-variables)). `MCP_AUTH_TOKEN` is only enforced in HTTP mode and ignored in stdio.

**stdio — agent-launched (no network):**

```bash
# direct
MCP_TRANSPORT=stdio ./dist/linux/amd64/chile-bcn-mcp
# Windows (PowerShell)
$env:MCP_TRANSPORT="stdio"; .\dist\windows\amd64\chile-bcn-mcp.exe
```

MCP client config (generic — Claude Code / Desktop, Cursor, etc.) — point `command` at the absolute path:

```json
{
  "mcpServers": {
    "chile-bcn": {
      "command": "/absolute/path/to/chile-bcn-mcp",
      "env": { "MCP_TRANSPORT": "stdio" }
    }
  }
}
```

> `command` must be an absolute path; `MCP_TRANSPORT=stdio` is required. No port, no token, no `-i`/`-t` flags — those are container-only.

**HTTP — self-hosted (always-on):**

```bash
# default: http://127.0.0.1:8000/mcp  (+ health at /health)
./dist/linux/amd64/chile-bcn-mcp

# custom host/port/path + auth
MCP_HOST=0.0.0.0 MCP_PORT=9000 MCP_PATH=/mcp MCP_AUTH_TOKEN=your-token \
  ./dist/linux/amd64/chile-bcn-mcp
# health check
curl http://localhost:8000/health
# -> {"status":"healthy"}
```

MCP client config for HTTP:

```json
{
  "mcpServers": {
    "chile-bcn": {
      "type": "http",
      "url": "http://localhost:8000/mcp",
      "headers": { "Authorization": "Bearer your-token" }
    }
  }
}
```

Omit `headers` if you didn't set `MCP_AUTH_TOKEN`. Adjust `url` if you changed `MCP_PORT` or `MCP_PATH`.

#### Agent examples: Claude Code, Codex and Hermes

All three agents can use the same binary. Pick **stdio** (agent spawns the binary) or **HTTP** (binary runs as a daemon). `command`/`args` below must be absolute paths — `~` is not expanded by the agents.

##### Claude Code

Claude Code reads `.mcp.json` (project scope, committed) or `~/.claude.json` (user scope). `type: "http"` is required for HTTP; an entry with `url` but no `type` is treated as stdio and skipped.

**stdio via CLI:**

```bash
# add (user scope is default; add --scope project to write .mcp.json)
claude mcp add --env MCP_TRANSPORT=stdio --transport stdio chile-bcn -- /absolute/path/to/chile-bcn-mcp

# verify
claude mcp list
claude mcp get chile-bcn
```

**stdio via JSON** (`.mcp.json` or `~/.claude.json`):

```json
{
  "mcpServers": {
    "chile-bcn": {
      "command": "/absolute/path/to/chile-bcn-mcp",
      "env": { "MCP_TRANSPORT": "stdio" }
    }
  }
}
```

Windows example: `"command": "C:\\tools\\chile-bcn-mcp.exe"` (same `env`).

**HTTP via CLI:**

```bash
claude mcp add --transport http chile-bcn http://localhost:8000/mcp --header "Authorization: Bearer your-token"
# without auth:
claude mcp add --transport http chile-bcn http://localhost:8000/mcp
```

**HTTP via JSON:**

```json
{
  "mcpServers": {
    "chile-bcn": {
      "type": "http",
      "url": "http://localhost:8000/mcp",
      "headers": { "Authorization": "Bearer your-token" }
    }
  }
}
```

Omit `headers` when `MCP_AUTH_TOKEN` is empty. `type` also accepts `streamable-http` as an alias.

##### Codex CLI

Codex reads `~/.codex/config.toml` (user) or `.codex/config.toml` (project). Restart Codex after editing the file. TOML keys are `mcp_servers` (with underscore).

**stdio via `config.toml`:**

```toml
[mcp_servers.chile-bcn]
command = "/absolute/path/to/chile-bcn-mcp"
args = []

[mcp_servers.chile-bcn.env]
MCP_TRANSPORT = "stdio"
```

**HTTP via `config.toml`:**

```toml
[mcp_servers.chile-bcn]
url = "http://localhost:8000/mcp"
# if you set MCP_AUTH_TOKEN on the server, add the header:
# headers = { Authorization = "Bearer your-token" }
```

> Codex also supports per-project overrides and `codex mcp` subcommands in newer releases — the TOML above is the stable shared format for CLI and VS Code.

##### Hermes Agent

Hermes reads `~/.hermes/config.yaml` (or `~/.config/hermes/config.yaml` depending on install) under the `mcp_servers` key, and can also migrate Claude Code's `mcpServers` via `hermes import-agent claude-code`. After changing the YAML, run `/reload-mcp` or restart Hermes.

**stdio via YAML:**

```yaml
mcp_servers:
  chile-bcn:
    command: "/absolute/path/to/chile-bcn-mcp"
    args: []
    env:
      MCP_TRANSPORT: "stdio"
    # optional: limit the tool surface
    # tools:
    #   include: [search_laws, get_law, get_law_summary, get_law_history]
```

**stdio via CLI:**

```bash
hermes mcp add chile-bcn --command /absolute/path/to/chile-bcn-mcp --env MCP_TRANSPORT=stdio
hermes mcp test chile-bcn
```

**HTTP via YAML:**

```yaml
mcp_servers:
  chile-bcn:
    url: "http://localhost:8000/mcp"
    headers:
      Authorization: "Bearer your-token"
    # optional filtering — same as stdio
    # tools:
    #   exclude: [get_law_history]
```

Omit `headers` when auth is disabled. Hermes also supports `tools.prompts` / `tools.resources` toggles — leave them at defaults unless you need to hide wrapper tools.

**HTTP via CLI:**

```bash
hermes mcp add chile-bcn --url http://localhost:8000/mcp --header "Authorization: Bearer your-token"
```

---

## Environment Variables
| Variable | Default | Purpose |
|----------|---------|---------|
| `MCP_TRANSPORT` | `http` | Transport mode: `http` (self-hosted) or `stdio` (agent-launched) |
| `MCP_HOST` | `127.0.0.1` (`0.0.0.0` in the container) | Bind address (HTTP only) |
| `MCP_PORT` | `8000` | Port to listen on (HTTP only) |
| `MCP_PATH` | `/mcp` | HTTP endpoint path (HTTP only) |
| `MCP_AUTH_TOKEN` | *(empty)* | Bearer token required for HTTP auth; ignored in stdio |
The LeyChile endpoints (URLs, timeouts, retry policy, circuit breaker) are declared in `internal/config/api.resources.yaml`, loaded once at startup from the embedded contract with fail-fast validation, and baked into the binary at build time.

---

## Available Tools

The server exposes three MCP tools:

| Tool | Purpose |
|------|---------|
| **`search_laws(query, page?, page_size?)`** | Paginated search across Chilean laws, decrees and resolutions. Returns each result with its `norm_id`, ready to fetch. |
| **`get_law(norm_id, version_date?, structure_only?)`** | Full content of a norm by `norm_id`: metadata, nested table of contents, related bills, and the complete text in Markdown. `version_date` (`YYYY-MM-DD`, strict) returns the version in force at that date. |
| **`get_law_history(norm_id)`** | Legislative history of a norm: its own history, the laws that modified it and the laws it modified — with dates, descriptions and LeyChile ficha links. |
| **`get_law_summary(norm_id, version_date?)`** | Lightweight overview of a norm — title, source, matters, categories and the official BCN summary — without the content. Use before reading the full text. |
| **`search_cgr_dictamenes(query, exact_search?, order?, page?)`** | Paginated search of Contraloría dictámenes (20 per page, `order` `date`/`dateasc`/`score`). Returns `dictamen_id`, `materia`, `descriptores` and HTML/PDF URLs for citation. |
| **`get_cgr_dictamen(dictamen_id)`** | Full dictamen by `dictamen_id` (e.g. `E179593N25`): metadata + sanitized `documento_completo` with `char_count` and `url`/`pdf_url` for citation and PDF download. |
| **`count_cgr_jurisprudencia(query, exact_search?)`** | Cross-type count for a query (dictamenes, auditoria, legislacion, etc.) without fetching documents — buckets with counts per type. |

**Caching**: norm content is cached in memory with **ETag revalidation**, keyed per version (`norm_id@version_date` — historical versions never share cache entries). Re-requesting the same norm+version sends `If-None-Match` and a `304` from LeyChile serves the cached copy without re-downloading or re-converting. `get_law_summary` derives from the same cache without any network call on a hit; `get_law_history` has its own ETag cache. Contraloría dictámenes use a simple **LRU 100 + singleflight** cache (no ETag — CGR does not send it) with keys `search:{query}|{exact}|{order}|{page}`, `dictamen:{id}` and `count:{query}|{exact}`.
---

## Sample Usage

**Find a law:**
```
Busca la Ley 21.600
```
→ The LLM calls `search_laws("Ley 21.600")`, reads the summaries, picks `norm_id: 1195666`

**Get an overview before the full text:**
```
¿De qué trata la ley 21.214?
```
→ The LLM calls `get_law_summary(norm_id=1142880)` — title, categories and the official BCN summary, no content

**Read the full law:**
```
¿Qué dice el Artículo 1 de la Ley 21.600?
```
→ The LLM calls `get_law(norm_id=1195666)` and navigates the Markdown (títulos → artículos)

**Read a historical version:**
```
¿Qué decía la Ley 19.628 en 2010?
```
→ The LLM calls `get_law(norm_id=141599, version_date="2010-01-01")` — the header shows "Version: as of 2010-01-01"

**Trace what modified a law:**
```
¿Qué leyes han modificado la Ley 21.600?
```
→ The LLM calls `get_law_history(norm_id=1195666)` and reads the modificatorias group

**Explore a long norm cheaply:**
```
¿Cuántos títulos tiene la Ley 21.600?
```
→ The LLM calls `get_law(norm_id=1195666, structure_only=true)` — metadata + table of contents only

---

## Prompts

The server also exposes nine **curated prompts** — server-side templates that guide the model through the correct workflow for each task. They encode the domain rules (read summaries before opening norms, verify against the actual text, never invent articles) so clients don't need a custom system prompt.

| Prompt | Arguments | When to use |
|--------|-----------|-------------|
| **`analyze_law`** | `norm_id`*, `aspect` | Structured legal analysis of a norm (purpose, scope, obligations, sanctions) with citations |
| **`search_legal_topic`** | `topic`* | Guided search: pick a query, read summaries first, verify with the full text |
| **`compare_law_versions`** | `norm_id`*, `from_date`*, `to_date`* | Compare a norm between two dates using historical `version_date` |
| **`trace_law_history`** | `norm_id`* | Trace which laws modified a norm, with the correct LeyChile ids |
| **`check_law_validity`** | `norm_id`*, `date` | In force / derogated / in force at a given date |
| **`explain_law_simply`** | `norm_id`*, `audience` | Plain-language explanation with citations and a no-legal-advice disclaimer |
| **`law_research_workflow`** | `norm_id`*, `question` | Efficient section-by-section reading via `get_law_summary` → `section_id` |
| **`answer_constitutional_question`** | `question`*, `article_hint`, `version_date` | Q&A about the Chilean Constitution (Decreto 100, 242302) via TOC + `section_id`, with hedge + disclaimer |
| **`check_norm_constitutionality`** | `norm_id`*, `question`, `version_date` | Contrast a norm vs the Constitution side-by-side, citing both, with version support |

(* = required)

Prompts complement the [Recommended System Prompt](#recommended-system-prompt): that one covers the general stance; the prompts encode the step-by-step workflow per task. Serving a prompt is a pure template operation — it never calls the BCN API.

---

## Image Tags & Updates

Images are tagged by **semver**:

| Tag | Meaning |
|-----|---------|
| `latest` | Most recent release |
| `1.2.0` | Release built from the `v1.2.0` tag |

**Supported architectures:** `linux/amd64`, `linux/arm64` (multi-arch manifest, built with the standard Docker `BUILDPLATFORM` cross-compile pattern — Go compiles natively, no QEMU for the build stage).

### How is the image published?

Publishing is deliberate, not automatic:

1. A tag `v*` (or a manual `workflow_dispatch`) triggers the publish workflow.
2. `docker/build-push-action` builds both architectures and pushes to `ghcr.io/<owner>/<repo>` with the version tag + `latest`.
3. CI (tests + vet) runs on every push and PR, independently of publishing.

---

| Path | What it is |
|------|-----------|
| `cmd/chile-bcn-mcp/` | The MCP server binary entry point |
| `internal/bcn/` | LeyChile domain client: resty per-endpoint, retry/circuit breaker, nested-content parsing, Markdown conversion, sanitizer, ETag cache |
| `internal/cgr/` | Contraloría domain client: resty per-endpoint, retry/circuit breaker, sanitizer (clean-directo), LRU 100 + singleflight |
| `internal/config/` | The `api.resources.yaml` contract loader with fail-fast validation (embedded via `go:embed`) |
| `internal/config/api.resources.yaml` | LeyChile + Contraloría endpoints contract (baked into the binary) |
| `internal/prompts/prompts.yaml` | Curated MCP prompts (baked into the binary) |
| `.github/workflows/` | CI (test + vet) and publish (multi-arch → GHCR) |
| `Makefile` | Build helpers: `make check`, `make smoke`, `make mock`, `make podman-*`, etc. |
| `.mockery.yml` | Mock generation config (mocks live next to the production file) |

---

## FAQ

**How do I update the image?**
Until GHCR is live: `podman build -t chile-bcn-mcp:local .` (or `docker build`). Once published: `podman pull ghcr.io/<owner>/<repo>:latest`.

**Can I use my own MCP client?**
Yes. The server speaks standard MCP over stdio or HTTP — Claude Code, Claude Desktop, Cursor, and other MCP-compatible agents work.

**Is the legal data complete?**
The server serves exactly what the public LeyChile API provides — full norm content (nested titles, paragraphs and articles), official BCN summaries, related bills and metadata. Nothing is filtered or summarized by an LLM.

**How is auth handled?**
HTTP mode uses an optional bearer token (`MCP_AUTH_TOKEN`). Stdio mode needs no auth — the container runs locally as a child of your agent.

Each endpoint has its own timeout, retry (transient 5xx/timeouts) and circuit breaker, configured in `internal/config/api.resources.yaml` (embedded). When the breaker opens, calls fail fast without hammering the API until it recovers.

---

## Recommended System Prompt

For optimal results, pair the server with this system prompt:

> "When working with Chilean legislation questions, use the chile-bcn MCP tools. Start with `search_laws(query)` to find the norm and get its `norm_id`. Use `get_law_summary(norm_id)` to understand what a norm is about before reading the full text, and `get_law(norm_id)` to read the complete content in Markdown — with `structure_only=true` when you only need the table of contents. Always ground your answers in the actual norm text; never invent article numbers or content. When the user asks about a specific law (e.g. 'Ley 21.600'), search for it and read it before answering."

---

## Disclaimer

This project is provided **"as is"**, for educational and informational purposes only. By using it, you accept the following:

- **No responsibility for misuse.** The author assumes no responsibility for any misuse of this software, or for any consequence — legal, technical or otherwise — arising from its use. Information returned through these tools should always be verified against the official source.
- **No data stored.** This project does not store, persist or share any data — neither from its users nor from BCN/LeyChile. Content is fetched at request time and cached in memory only, for the lifetime of the running process.
- **No affiliation.** This project is **not** affiliated with, endorsed by, or connected in any way to the Biblioteca del Congreso Nacional (BCN), LeyChile, or any Chilean government or public institution. It is an independent, community-built tool.
- **No availability guarantee.** Given the nature of BCN's public API, the author is not responsible if it stops working, changes its responses, or is interrupted at any time.
- **Not for production.** This project was not created with the intention of running in production or critical environments. Use it at your own discretion.
- **Be considerate of a public service.** Please be responsible with the number of requests made to each tool, and avoid saturating the servers of a public service such as BCN.

---

## Release Process

Releases happen **only by merging a `release/v*` branch into `main`** — there is no tag-push trigger. The flow:

1. Create a PR from `release/v<version>` (e.g. `release/v0.1.0`) to `main` and **merge it**.
2. On merge, the workflow:
   - builds the **cross-platform distributions** into `dist.zip` (see below), and attaches them to the release
   - publishes the OCI image to GHCR with tags `<version>` and `latest`
   - creates a **draft** GitHub Release with tag `v<version>` — you publish it manually
3. A PR closed **without** merging generates nothing.

For a manual run, use the workflow dispatch and enter the version.

Self-contained builds for six targets — each folder carries only its binary (the endpoints contract and prompts are embedded via `go:embed`, so `cd linux/amd64 && ./chile-bcn-mcp` just works from any folder):
```
dist.zip
├── windows/{amd64,arm64}/chile-bcn-mcp.exe
├── linux/{amd64,arm64}/chile-bcn-mcp
├── darwin/{amd64,arm64}/chile-bcn-mcp     (amd64 = Intel, arm64 = Apple Silicon)
└── SHA256SUMS.txt
```

Build locally with `make dist` (runs the same script as CI).

---

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

Third-party components (the go-sdk, resty, html-to-markdown, and their dependency tree — MIT, Apache-2.0, BSD-3-Clause, ISC) are attributed with their full license texts in [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).

Legal content is served verbatim from the public LeyChile API of the Biblioteca del Congreso Nacional de Chile (BCN) and belongs to its original source.
