# Data Model

```
User ──┬── Project   (created_by)
       ├── Issue     (created_by + assignee)
       └── Comment   (author + created_by)
```

## States (fixed pipeline)

`backlog` → `todo` → `in_progress` → `qa` → `done` → `cancelled`

## Types

`epic`, `feature`, `bug`, `chore`

## Priority

| Level | Label |
|-------|-------|
| 0 | none |
| 1 | urgent |
| 2 | high |
| 3 | medium |
| 4 | low |

## Slugs

Issues get auto-generated slugs: `<project-slug>-<auto-increment-id>`.
Example: `ASTEROID-GAME-42`. Use slugs in place of numeric IDs anywhere.
