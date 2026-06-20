# MCP Server

**Endpoint:** `POST /mcp?pat=pat_admin` — Streamable HTTP transport.

## Tools

| Tool | Args |
|------|------|
| `get_info` | _(none)_ |
| `list_users` | _(none)_ |
| `get_user` | `id` (number) |
| `list_projects` | _(none)_ |
| `get_project` | `id` (string) |
| `create_project` | `name`, `slug` |
| `update_project` | `id`, optional name/slug/description |
| `delete_project` | `id` |
| `list_issues` | `project_id`, optional `state`, `assignee` |
| `get_issue` | `id` (string) |
| `create_issue` | `project_id`, `title` |
| `update_issue` | `id`, optional fields |
| `update_issue_state` | `id`, `state` |
| `delete_issue` | `id` |
| `list_comments` | `issue_id` |
| `add_comment` | `issue_id`, `body` |

All `id`/`project_id` args accept numeric IDs or slugs.

## Client Configuration

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
