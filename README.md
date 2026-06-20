# Ticketer

A minimal, API-first project/issue tracker for AI agent teams. Built in Go, backed by SQLite, with an embedded kanban web UI and an MCP server for LLM-driven management.

## Quickstart

```bash
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  veloper/ticketer
```

Open http://localhost:8300/login and sign in with `admin` / `pat_admin`.

## Features

- **REST API** — full CRUD for projects, issues, comments, users. PAT-based auth.
- **Web UI** — kanban board, issue detail, comments. Login with your PAT.
- **MCP Server** — 16 tools for LLMs to manage projects directly (Streamable HTTP).
- **WebSocket** — real-time change broadcasting with self-event suppression.
- **CLI** — `tktrctl` for scripting, bootstrapping, and automation.
- **Single binary** — Go + embedded SQLite (WAL mode, zero CGO). No runtime deps.

## Web UI

| Route | View |
|-------|------|
| `/login` | Sign in with username + PAT |
| `/` | Projects list |
| `/projects/{id}` | Kanban board grouped by state |
| `/issues/{id}` | Issue detail with comments and state controls |

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `TICKETER_ADMIN_USERNAME` | — | Admin username **(required)** |
| `TICKETER_ADMIN_PAT` | — | Admin personal access token **(required)** |
| `TICKETER_HOST` | `""` | Listen host (all interfaces) |
| `TICKETER_PORT` | `"8300"` | Listen port |
| `TICKETER_DB_PATH` | `"ticketer.db"` | SQLite database path (`/data/ticketer.db` in Docker) |

## Docs

| Doc | Contents |
|-----|----------|
| [`docs/api.md`](docs/api.md) | REST API reference |
| [`docs/cli.md`](docs/cli.md) | tktrctl commands |
| [`docs/mcp.md`](docs/mcp.md) | MCP server tools and config |
| [`docs/websocket.md`](docs/websocket.md) | WebSocket events |
| [`docs/data-model.md`](docs/data-model.md) | States, types, priorities, slugs |
| [`docs/docker.md`](docs/docker.md) | Docker Compose and automation |
| [`docs/architecture.md`](docs/architecture.md) | System architecture diagram |
| [`AGENTS.md`](AGENTS.md) | Agent guide — how to use Ticketer programmatically |
