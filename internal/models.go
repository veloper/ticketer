package internal

import "time"

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Issue struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Slug  string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	State       string `json:"state"`
	Assignee    string `json:"assignee"`
	Priority    int    `json:"priority"`
	ParentID    string `json:"parent_id,omitempty"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Comment struct {
	ID        string `json:"id"`
	IssueID   string `json:"issue_id"`
	Body      string `json:"body"`
	Author    string `json:"author"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SeedUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	PAT         string `json:"pat"`
}

type Config struct {
	Users      []SeedUser `json:"users"`
	DefaultPAT string     `json:"default_pat"`
	DBPath     string     `json:"db_path"`
	Addr       string     `json:"addr"`
}

var ValidStates = []string{"backlog", "todo", "in_progress", "review", "done", "cancelled"}
var ValidTypes = []string{"epic", "feature", "bug", "chore"}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
