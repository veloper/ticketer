# Ticketer

A minimal, API-first project/issue tracker designed for AI agent teams. Built in Go, backed by embedded SQLite, with an embedded kanban web UI.

Agents manage issues through a REST API authenticated by pre-shared Personal Access Tokens (PATs). Humans view progress through a read/write web interface served from the same binary on the same port.

## Quickstart

```bash
go build -o ticketer ./cmd/ticketer
TICKETER_DB_PATH=ticketer.db ./ticketer
# → listening on :8080
```

Open `http://localhost:8080` for the web UI.

## Project Structure

```
ticketer/
├── cmd/ticketer/main.go      # Entry point, config, server, embedded web UI
├── internal/
│   ├── models.go             # Data types: User, Project, Issue, Comment, Config
│   ├── store.go              # SQLite storage — migrations, CRUD, queries
│   ├── handlers.go           # HTTP handlers — all REST endpoints
│   └── middleware.go          # PAT authentication middleware
├── web/
│   ├── index.html            # Projects list
│   ├── project.html          # Kanban board
│   └── issue.html            # Issue detail with comments
├── Dockerfile                # Multi-stage: Go build → alpine image
├── go.mod / go.sum
├── AGENTS.md                 # This file
└── README.md
```

## Data Model

```
User ──┬── Project (created_by)
        ├── Issue (created_by + assignee)
        └── Comment (created_by + author)
```

### User
```json
{"id": "uuid", "username": "sam_builder", "display_name": "Sam Builder"}
```
Seeded on startup from config/env. PAT authenticates the user.

### Project
```json
{"id": "uuid", "name": "Asteroid Game", "description": "...", "created_by": "uuid"}
```

### Issue
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "slug": "ASTEROID-GAME-42",
  "title": "Add ship rotation",
  "description": "Left/right arrows rotate the ship",
  "type": "feature",
  "state": "todo",
  "assignee": "uuid",
  "priority": 2,
  "parent_id": null
}
```

**Types** (standardized, opinionated):
| Type | Meaning |
|------|---------|
| `epic` | Large initiative, groups multiple issues |
| `feature` | New capability |
| `bug` | Something broken |
| `chore` | Maintenance, non-user-facing work |

**States** (fixed pipeline): `backlog` → `todo` → `in_progress` → `review` → `done` → `cancelled`

**Priority**: `0`=none, `1`=urgent, `2`=high, `3`=medium, `4`=low

The `slug` is the human-readable ID (e.g. `ASTEROID-GAME-1`), auto-generated from the project name + incrementing counter. Use this to reference issues in chat instead of UUIDs.

`parent_id` links sub-issues to their parent epic.

### Comment
```json
{"id": "uuid", "issue_id": "uuid", "author": "uuid", "body": "Rotation is off by 90 degrees"}
```
`author` is set from the PAT user automatically.

## API

All requests require `Authorization: Bearer <pat>`. PATs are seeded on startup.

### Users
```
GET /api/users             List all users
GET /api/users/{id}        Get user
GET /api/me                Get current user (from PAT)
```

### Projects
```
POST   /api/projects              Create project
GET    /api/projects              List projects
GET    /api/projects/{id}         Get project
PATCH  /api/projects/{id}         Update project
DELETE /api/projects/{id}         Delete project
```

### Issues
```
GET    /api/projects/{id}/issues   List issues with filters
POST   /api/projects/{id}/issues   Create issue
GET    /api/issues/{id}            Get issue
PATCH  /api/issues/{id}            Update issue
DELETE /api/issues/{id}            Delete issue
```

**Filters for listing issues:** `?type=bug&state=review&assignee=<id>&q=search&page=1&per_page=50`

### Comments
```
GET    /api/issues/{id}/comments   List comments
POST   /api/issues/{id}/comments   Add comment
```

## Web UI

Served at `/`. No auth required — the server injects a default PAT into the page so the browser can call the API transparently.

| Route | View |
|-------|------|
| `/` | Projects list |
| `/projects/{id}` | Kanban board grouped by state |
| `/issues/{id}` | Issue detail with comments and state controls |

Agent-created and agent-updated issues appear in real-time on refresh.

## Configuration

| Env var | Default | Description |
|---------|---------|-------------|
| `TICKETER_ADDR` | `:8080` | Listen address |
| `TICKETER_DB_PATH` | `ticketer.db` | SQLite database path |
| `TICKETER_DEFAULT_PAT` | `""` | PAT injected into web UI |
| `TICKETER_USERS` | (defaults) | JSON array of users with PATs |

Default seed users:
```
alex_planner / pat_alex  → display: Alex Planner
sam_builder  / pat_sam   → display: Sam Builder
tommy_tester / pat_tommy → display: Tommy Tester
```

## Architecture

```
Single binary, zero runtime dependencies
─────────────────────────────────────────

http.ServeMux (:8080)
├── /api/* → REST handlers (PAT auth required)
└── /*     → Embedded web UI (no auth)
               │
               └── SQLite (WAL mode, pure Go, no CGO)
```

- No CGO. Uses `modernc.org/sqlite` — a pure Go SQLite port.
- UUIDs for all IDs. Slugs for human-readable references.
- WAL mode for concurrent reads during writes.
- Multi-stage Dockerfile builds in golang:1.22-alpine, runs on alpine:3.20.

## Typical Workflow (3-Agent Team)

```
Alex (planner) creates project + issues (epics, features, chores)
  → Issues start in backlog/todo
  → Alex assigns issues to Sam

Sam (builder) picks up assigned issues
  → Moves to in_progress
  → Builds the feature
  → Moves to review

Tommy (tester) reads code, tests, finds bugs
  → Comments with repro steps on the issue
  → If bug found, Alex creates a bug issue, assigns to Sam

Sam fixes → moves back to review
Tommy signs off → moves to done
```

## Development

```bash
go build -o ticketer ./cmd/ticketer   # Build
go test ./...                          # Test
docker build -t ticketer .             # Docker image
```
