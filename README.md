# Ticketer

**Project/issue tracker for AI agent teams.** REST API, kanban web UI, MCP server, CLI — all in one Go binary with embedded SQLite.

[![Docker Hub](https://img.shields.io/docker/pulls/veloper/ticketer?color=2563eb&label=Docker%20Pulls)](https://hub.docker.com/r/veloper/ticketer)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

## Quickstart

```bash
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  veloper/ticketer
```

Open **http://localhost:8300/login** and sign in with `admin` / `pat_admin`.

---

## Why Ticketer?

Built for the way AI agents work — stateless, API-first, zero setup.

- **[MCP Server](docs/mcp.md)** at `/mcp` — 16 tools for LLMs to manage projects directly. Streamable HTTP transport.
- **[REST API](docs/api.md)** with PAT auth — full CRUD for projects, issues, comments, users. Slug-based references.
- **[tktrctl CLI](docs/cli.md)** — script bootstrapping, automate workflows, manage from the terminal.
- **[WebSocket](docs/websocket.md)** — real-time change broadcasting with self-event suppression.
- **[Single Docker Container](docs/docker.md)** — one image, both `ticketer` and `tktrctl` binaries. Compose, automation, setup service.
- **[Single Go binary](docs/architecture.md)** — Go + SQLite (WAL, no CGO). No runtime deps. ~20 MB image.

## Configuration

| Variable | Default | Required |
|----------|---------|----------|
| `TICKETER_ADMIN_USERNAME` | — | Yes |
| `TICKETER_ADMIN_PAT` | — | Yes |
| `TICKETER_PORT` | `8300` | |
| `TICKETER_HOST` | `""` (all) | |
| `TICKETER_DB_PATH` | `ticketer.db` | |

## Docs

| | |
|---|---|
| **API** | [`docs/api.md`](docs/api.md) — endpoints, examples, errors |
| **CLI** | [`docs/cli.md`](docs/cli.md) — tktrctl commands and usage |
| **MCP** | [`docs/mcp.md`](docs/mcp.md) — LLM tools and client config |
| **WebSocket** | [`docs/websocket.md`](docs/websocket.md) — real-time events |
| **Data Model** | [`docs/data-model.md`](docs/data-model.md) — states, types, priorities |
| **Docker** | [`docs/docker.md`](docs/docker.md) — Compose, automation, setup |
| **Architecture** | [`docs/architecture.md`](docs/architecture.md) — system design |
| **Agent Guide** | [`AGENTS.md`](AGENTS.md) — for AI agents using Ticketer |
