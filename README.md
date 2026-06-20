# Ticketer

A minimal, API-first project/issue tracker for AI agent teams. Built in Go, backed by SQLite, with an embedded kanban web UI and an MCP server for LLM-driven management.

## Features

- **REST API** — full CRUD for projects, issues, comments, users. PAT-based auth.
- **Web UI** — kanban board, issue detail, comments. Login with your PAT.
- **MCP Server** — 16 tools for LLMs to manage projects directly (Streamable HTTP).
- **WebSocket** — real-time change broadcasting with self-event suppression.
- **CLI** — `tktrctl` for scripting, bootstrapping, and automation.
- **Single binary** — Go + embedded SQLite (WAL mode, zero CGO). No runtime deps.

## Quickstart

```bash
git clone https://github.com/veloper/ticketer.git
cd ticketer
docker compose up
```

Open http://localhost:8300/login and sign in with `admin` / `pat_admin`.

The default `docker compose up` uses the admin credentials above. See [Docker Setup](#docker-setup) for customization.

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `TICKETER_ADMIN_USERNAME` | — | Admin username **(required)** |
| `TICKETER_ADMIN_PAT` | — | Admin personal access token **(required)** |
| `TICKETER_HOST` | `""` | Listen host (all interfaces) |
| `TICKETER_PORT` | `"8300"` | Listen port |
| `TICKETER_DB_PATH` | `"ticketer.db"` | SQLite database path (`/data/ticketer.db` in Docker) |

The admin user is created on startup. Additional users can be created via the API or CLI by the admin.

## Docker Setup

### Single container

```bash
docker build -t ticketer .
docker run -p 8300:8300 \
  -e TICKETER_ADMIN_USERNAME=admin \
  -e TICKETER_ADMIN_PAT=pat_admin \
  ticketer
```

### Docker Compose

```yaml
services:
  ticketer:
    build: .
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

The database persists at `/data/ticketer.db`. The Docker image includes both `ticketer` and `tktrctl` binaries.

### Using tktrctl in Compose

Run one-off commands:

```bash
docker compose exec ticketer \
  tktrctl projects create "Asteroid Game" ASTEROID-GAME
```

Or automate setup with a separate service:

```yaml
services:
  ticketer:
    build: .
    ports: ["8300:8300"]
    environment:
      TICKETER_ADMIN_USERNAME: admin
      TICKETER_ADMIN_PAT: pat_admin
    volumes:
      - ticketer-data:/data

  setup:
    build: .
    profiles: ["setup"]
    environment:
      TICKETER_HOST: http://ticketer:8300
      TICKETER_PAT: pat_admin
    depends_on:
      ticketer:
        condition: service_started
    command: >
      tktrctl projects create "Asteroid Game" ASTEROID-GAME &&
      tktrctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
```

Run with: `docker compose --profile setup run setup`

## CLI (tktrctl)

A companion CLI for bootstrapping and managing the server. Configured via environment variables — no flags needed.

```bash
# Build (or use the Docker image which includes both binaries)
go build -o tktrctl ./cmd/tktrctl

# Configure
export TICKETER_HOST=http://localhost:8300
export TICKETER_PAT=pat_admin

tktrctl info
tktrctl users list
tktrctl projects list
tktrctl projects create "Asteroid Game" ASTEROID-GAME
tktrctl issues create ASTEROID-GAME "Fix login" --type bug --priority 1
tktrctl issues state ASTEROID-GAME-42 qa
```

### Commands

| Command | Description |
|---------|-------------|
| `info` | Show server metadata, valid values, users, projects |
| `users` | `list`, `show <id>`, `create <username>`, `update <id>`, `delete <id>` |
| `projects` | `list`, `show <id>`, `create <name> <slug>`, `update <id>`, `delete <id>` |
| `issues` | `list <project>`, `show <id>`, `create <project> <title>`, `update <id>`, `state <id>`, `state-update <id> <state>` |

Projects and issues can be referenced by numeric ID or slug. Set `TICKETER_PAT` to your admin PAT.

## Web UI

The kanban board is served on the same port as the API. Sign in at `/login` with your admin credentials.

| Route | View |
|-------|------|
| `/login` | Sign in with username + PAT |
| `/` | Projects list |
| `/projects/{id}` | Kanban board grouped by state |
| `/issues/{id}` | Issue detail with comments and state controls |

## API

All API requests require `Authorization: Bearer <pat>`.

### Projects

```
POST   /api/projects                  Create project
GET    /api/projects                  List projects
GET    /api/projects/{id}             Get project
PATCH  /api/projects/{id}             Update project
DELETE /api/projects/{id}             Delete project
```

### Issues

```
GET    /api/projects/{id}/issues      List issues (filterable)
POST   /api/projects/{id}/issues      Create issue
GET    /api/issues/{id}               Get issue (by ID or slug)
PATCH  /api/issues/{id}               Update issue fields
PUT    /api/issues/{id}/state         Update issue state only
DELETE /api/issues/{id}               Delete issue
```

**Filters:** `?state=qa&assignee=<id>&q=search&page=1&per_page=50`

### Comments

```
GET    /api/issues/{id}/comments      List comments
POST   /api/issues/{id}/comments      Add comment
```

### Users

```
GET    /api/users                     List users
GET    /api/users/{id}                Get user
POST   /api/users                     Create user (admin only)
PATCH  /api/users/{id}                Update user (admin only)
DELETE /api/users/{id}                Delete user (admin only)
GET    /api/me                        Get current user
```

### Info

```
GET    /api/info                      Server metadata (states, types, priorities, users, projects)
```

## MCP Server

A Model Context Protocol server is available for LLM-driven project management.

```
Endpoint:  POST /mcp?pat=pat_admin
Transport: Streamable HTTP
SDK:       github.com/mark3labs/mcp-go
```

### Tools

| Tool | Description |
|------|-------------|
| `get_info` | Discover the server — valid states, types, priorities, users, projects, authenticated user |
| `list_users` / `get_user` | List all users or get one by ID |
| `list_projects` / `get_project` | List projects or get one by ID/slug |
| `create_project` | Create a project (name + slug required) |
| `update_project` | Update project name, slug, or description |
| `delete_project` | Delete a project permanently |
| `list_issues` | List issues in a project, filtered by state or assignee |
| `get_issue` | Get an issue by ID or slug (e.g. `GAME-42`) |
| `create_issue` | Create an issue (project + title required) |
| `update_issue` | Update issue fields — only provided fields change |
| `update_issue_state` | Move issue through the state pipeline |
| `delete_issue` | Delete an issue permanently |
| `list_comments` / `add_comment` | List or add comments on an issue |

### Client Configuration

```json
{
  "mcpServers": {
    "ticketer": {
      "type": "http",
      "url": "http://localhost:8300/mcp?pat=pat_admin"
    }
  }
}
```

## Data Model

```
User ──┬── Project   (created_by)
       ├── Issue     (created_by + assignee)
       └── Comment   (author + created_by)
```

**States** (fixed pipeline): `backlog` → `todo` → `in_progress` → `qa` → `done` → `cancelled`

**Types**: `epic`, `feature`, `bug`, `chore`

**Priority**: `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low

Issues get auto-generated slugs like `ASTEROID-GAME-42` using the project slug and the issue's auto-increment ID.

## Architecture

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ REST API    │  │ Web UI      │  │ MCP Server  │  │ WebSocket   │
│ /api/*      │  │ /           │  │ /mcp        │  │ /api/ws     │
└──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘
       │                │                │                │
       └────────────────┴────────────────┴────────────────┘
                              net/http
       ┌─────────────────────────────────────────────────────┐
       │                  SQLite (WAL mode)                  │
       │             modernc.org/sqlite (no CGO)             │
       └─────────────────────────────────────────────────────┘
```

Single Go binary, zero runtime dependencies. All services share the same port.

## Development

```bash
# Run tests
go test ./...

# Build for current platform
go build -o ticketer ./cmd/ticketer
go build -o tktrctl ./cmd/tktrctl

# Build Docker image
docker build -t ticketer .
```
