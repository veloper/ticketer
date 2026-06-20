# API

Base URL: `http://<host>:<port>/api`. All requests require `Authorization: Bearer <pat>`.

## Users

```
GET    /api/users                  List users
GET    /api/users/{id}             Get user
POST   /api/users                  Create user (admin only)
PATCH  /api/users/{id}             Update user (admin only)
DELETE /api/users/{id}             Delete user (admin only)
GET    /api/me                     Get current user
```

## Projects

```
POST   /api/projects               Create project
GET    /api/projects               List projects
GET    /api/projects/{id}          Get project (by ID or slug)
PATCH  /api/projects/{id}          Update project
DELETE /api/projects/{id}          Delete project
```

## Issues

```
GET    /api/projects/{id}/issues   List issues (filterable)
POST   /api/projects/{id}/issues   Create issue
GET    /api/issues/{id}            Get issue (by ID or slug)
PATCH  /api/issues/{id}            Update issue fields
PUT    /api/issues/{id}/state      Update issue state only
DELETE /api/issues/{id}            Delete issue
```

Filters: `?state=qa&assignee=<id>&q=search&page=1&per_page=50`

## Comments

```
GET    /api/issues/{id}/comments   List comments
POST   /api/issues/{id}/comments   Add comment
```

## Info

```
GET    /api/info                   Server metadata (states, types, priorities, users, projects)
```
