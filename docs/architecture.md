# Architecture

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

- Single Go binary, zero runtime dependencies
- All services share the same port
- PAT middleware authenticates `/api/*` and `/mcp` routes
- WebSocket and MCP use `?pat=` query param (browsers can't set custom headers)
- SQLite with WAL mode for concurrent reads during writes
- Pure Go SQLite driver — no CGO needed
