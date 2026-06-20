[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)](LICENSE)

# Ticketer

**Project/issue tracker for AI agent teams.** REST API, kanban web UI, MCP server, CLI — all in one Go binary with embedded SQLite.

---

## Quickstart

Run with Docker Compose:

```yaml
services:
  ticketer:
    image: veloper/ticketer
    ports:
      - "8300:8300"
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

volumes:
  ticketer-data:
```

Open **http://localhost:8300/login** and sign in with `admin` / `pat_admin`.

---

## Why Ticketer?

Built for the way AI agents work — API-first, zero setup.

- **[MCP Server](docs/mcp.md)** — let any LLM manage your projects. 16 tools, zero configuration.
- **[REST API](docs/api.md)** — clean, predictable CRUD. Authenticate with a token, reference issues by slug.
- **[tktrctl CLI](docs/cli.md)** — script your workflow, automate bootstrapping, manage from anywhere.
- **[WebSocket](docs/websocket.md)** — real-time updates without polling. Changes broadcast the moment they happen.
- **[Single Docker Container](docs/docker.md)** — everything in one image. Compose, automate, scale down.
- **[Single Go binary](docs/architecture.md)** — no runtime dependencies, no interpreter, no JVM. Just run it.

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
