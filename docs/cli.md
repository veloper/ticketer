# CLI (tktrctl)

Configure via environment variables:

```bash
export TICKETER_HOST=http://localhost:8300
export TICKETER_PAT=pat_admin
```

Projects and issues can be referenced by numeric ID or slug.

## Commands

```
tktrctl info                            Server metadata
tktrctl users list                      List users
tktrctl users show <id>                 Get user
tktrctl users create <username>         Create user (--display-name, --admin)
tktrctl users update <id>               Update user (--display-name, --pat)
tktrctl users delete <id>               Delete user
tktrctl projects list                   List projects
tktrctl projects show <id>              Get project
tktrctl projects create <name> <slug>   Create project (--description)
tktrctl projects update <id>            Update project (--name, --slug, --description)
tktrctl projects delete <id>            Delete project
tktrctl issues list <project>           List issues (--state, --assignee)
tktrctl issues show <id>                Get issue
tktrctl issues create <project> <title> Create issue (--type, --state, --priority, --assignee)
tktrctl issues update <id>              Update issue (--title, --state, --type, --priority, --assignee)
tktrctl issues state <id>               Show current state
tktrctl issues state-update <id> <state> Update state
```
