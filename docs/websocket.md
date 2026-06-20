# WebSocket

**Endpoint:** `ws://<host>:<port>/api/ws?pat=pat_admin`

Events are JSON with `type` and `payload`. Self-events (caused by the connected user) are suppressed.

## Events

| Type | Payload |
|------|---------|
| `project_created` | Full project |
| `project_updated` | `{"id": ..., "changed": {...}}` |
| `project_deleted` | `{"id": ...}` |
| `issue_created` | Full issue |
| `issue_updated` | `{"id": ..., "changed": {...}}` |
| `issue_deleted` | `{"id": ..., "project_id": ...}` |
| `comment_created` | Full comment |

## Update Format

Update events send only the fields that changed:

```json
{
  "type": "issue_updated",
  "payload": {
    "id": 1,
    "changed": {
      "state": {"before": "todo", "after": "qa"},
      "assignee": {"before": null, "after": 2}
    }
  }
}
```
