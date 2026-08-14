# Chile BCN MCP Server

[![Podman](https://img.shields.io/badge/Podman-First-purple?logo=podman)](https://podman.io/)
[![Docker](https://img.shields.io/badge/Docker-Fallback-blue?logo=docker)](https://www.docker.com/)
[![MCP](https://img.shields.io/badge/MCP-Compatible-green)](https://modelcontextprotocol.io/)
[![Go](https://img.shields.io/badge/Go-blue?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Model Context Protocol (MCP) server that gives AI assistants direct access to Chilean laws, decrees and resolutions from **LeyChile**, the legal database of the Biblioteca del Congreso Nacional de Chile (BCN) — in a format designed for how LLMs consume text.

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
- [Environment Variables](#environment-variables)
- [Available Tools](#available-tools)
- [Sample Usage](#sample-usage)
- [Image Tags & Updates](#image-tags--updates)
- [Repository Layout](#repository-layout)
- [FAQ](#faq)
- [Recommended System Prompt](#recommended-system-prompt)
- [Disclaimer](#disclaimer)
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
      "args": ["run", "--rm", "-i", "-e", "FASTMCP_TRANSPORT=stdio",
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

For development or when you want a single static binary with no container runtime.

```bash
# Build a static binary (outputs to bin/)
make build

# Or run directly for development
make run-http        # HTTP mode
make run-stdio       # stdio mode

# Cross-compile for other architectures
make build-arm64
```

The binary is fully static — it needs no Go toolchain at runtime. The API endpoints contract is baked next to the binary at `config/api.resources.yaml` (fixed path, no hot reload — configuration changes are deployed by rebuilding).

---

## Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `FASTMCP_TRANSPORT` | `http` | Transport mode: `http` (self-hosted) or `stdio` (agent-launched) |
| `FASTMCP_HOST` | `127.0.0.1` (`0.0.0.0` in the container) | Bind address (HTTP only) |
| `FASTMCP_PORT` | `8000` | Port to listen on (HTTP only) |
| `FASTMCP_PATH` | `/mcp` | HTTP endpoint path (HTTP only) |
| `MCP_AUTH_TOKEN` | *(empty)* | Bearer token required for HTTP auth; ignored in stdio |

The LeyChile endpoints (URLs, timeouts, retry policy, circuit breaker) are declared in `config/api.resources.yaml`, loaded once at startup with fail-fast validation, and baked into the container image at build time.

---

## Available Tools

The server exposes three MCP tools:

| Tool | Purpose |
|------|---------|
| **`search_laws(query, page?, page_size?)`** | Paginated search across Chilean laws, decrees and resolutions. Returns each result with its `norm_id`, ready to fetch. |
| **`get_law(norm_id, structure_only?)`** | Full content of a norm by `norm_id`: metadata, nested table of contents, related bills, and the complete text in Markdown. |
| **`get_law_summary(norm_id)`** | Lightweight overview of a norm — title, source, matters, categories and the official BCN summary — without the content. Use before reading the full text. |

Every tool returns both readable text (`content[]`) and typed `structuredContent` with a generated JSON schema.

**Caching**: norm content is cached in memory with **ETag revalidation** — re-requesting the same `norm_id` sends `If-None-Match` and a `304` from LeyChile serves the cached copy without re-downloading or re-converting. `get_law_summary` derives from the same cache without any network call on a hit.

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

**Explore a long norm cheaply:**
```
¿Cuántos títulos tiene la Ley 21.600?
```
→ The LLM calls `get_law(norm_id=1195666, structure_only=true)` — metadata + table of contents only

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

## Repository Layout

| Path | What it is |
|------|-----------|
| `cmd/chile-bcn-mcp/` | The MCP server binary entry point |
| `internal/bcn/` | LeyChile domain client: resty per-endpoint, retry/circuit breaker, nested-content parsing, Markdown conversion, sanitizer, ETag cache |
| `internal/config/` | The `api.resources.yaml` contract loader with fail-fast validation |
| `internal/server/` | MCP server setup and configuration |
| `internal/tools/` | The three MCP tools (presentation layer) |
| `config/api.resources.yaml` | LeyChile endpoints contract (baked into the image) |
| `scripts/smoke.sh` | End-to-end smoke test against the real BCN API (`make smoke`) |
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

**What happens if LeyChile is down or rate-limits?**
Each endpoint has its own timeout, retry (transient 5xx/timeouts) and circuit breaker, configured in `config/api.resources.yaml`. When the breaker opens, calls fail fast without hammering the API until it recovers.

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

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

Third-party components (the go-sdk, resty, html-to-markdown, and their dependency tree — MIT, Apache-2.0, BSD-3-Clause, ISC) are attributed with their full license texts in [THIRD_PARTY_NOTICES](THIRD_PARTY_NOTICES).

Legal content is served verbatim from the public LeyChile API of the Biblioteca del Congreso Nacional de Chile (BCN) and belongs to its original source.
