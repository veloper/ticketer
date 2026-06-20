package internal

type User struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
	PAT         string `json:"pat,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Project struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Issue struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"project_id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	State       string `json:"state"`
	Assignee    int64  `json:"assignee,omitempty"`
	Priority    int    `json:"priority"`
	ParentID    int64  `json:"parent_id,omitempty"`
	CreatedBy   int64  `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Comment struct {
	ID        int64  `json:"id"`
	IssueID   int64  `json:"issue_id"`
	Body      string `json:"body"`
	Author    int64  `json:"author"`
	CreatedBy int64  `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type SeedUser struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	PAT         string `json:"pat"`
	Admin       bool   `json:"admin"`
}

type Config struct {
	AdminUsername string `json:"admin_username"`
	AdminPAT      string `json:"admin_pat"`
	DBPath        string `json:"db_path"`
	Host          string `json:"host"`
	Port          string `json:"port"`
}

// Addr returns the listen address from Host and Port.
func (c *Config) Addr() string {
	return c.Host + ":" + c.Port
}

var ValidStates = []string{"backlog", "todo", "in_progress", "qa", "done", "cancelled"}
var ValidTypes = []string{"epic", "feature", "bug", "chore"}
var ValidPriorityLevels = []int{0, 1, 2, 3, 4}
