package internal

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mcpCtxKey string

const mcpUserKey mcpCtxKey = "mcp_user"

// mcpUserFromCtx extracts the authenticated user from the MCP request context.
func mcpUserFromCtx(ctx context.Context) *User {
	u, _ := ctx.Value(mcpUserKey).(*User)
	return u
}

// NewMCPServer creates an MCP server with all Ticketer tools backed by the Store.
func NewMCPServer(store *Store) *server.MCPServer {
	s := server.NewMCPServer(
		"ticketer",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(true, true),
	)

	// ── Info ──
	s.AddTool(mcp.NewTool("get_info",
		mcp.WithDescription("Discover the full Ticketer surface — valid states (backlog→done), issue types, priority levels with labels, all registered users, and all projects. Call this first to understand what values are accepted by other tools."),
	), handleGetInfo(store))

	// ── Users ──
	s.AddTool(mcp.NewTool("list_users",
		mcp.WithDescription("List all users"),
	), handleListUsers(store))

	s.AddTool(mcp.NewTool("get_user",
		mcp.WithDescription("Get a user by ID"),
		mcp.WithNumber("id", mcp.Description("User ID"), mcp.Required()),
	), handleGetUser(store))

	// ── Projects ──
	s.AddTool(mcp.NewTool("list_projects",
		mcp.WithDescription("List all projects"),
	), handleListProjects(store))

	s.AddTool(mcp.NewTool("get_project",
		mcp.WithDescription("Get a project by ID or slug"),
		mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
	), handleGetProject(store))

	s.AddTool(mcp.NewTool("create_project",
		mcp.WithDescription("Create a new project — both name and slug are required"),
		mcp.WithString("name", mcp.Description("Project name"), mcp.Required()),
		mcp.WithString("slug", mcp.Description("Project slug"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Project description")),
	), handleCreateProject(store))

	s.AddTool(mcp.NewTool("update_project",
		mcp.WithDescription("Update an existing project — only provided fields will change"),
		mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
		mcp.WithString("name", mcp.Description("New name")),
		mcp.WithString("slug", mcp.Description("New slug")),
		mcp.WithString("description", mcp.Description("New description")),
	), handleUpdateProject(store))

	s.AddTool(mcp.NewTool("delete_project",
		mcp.WithDescription("Delete a project and all its issues permanently"),
		mcp.WithString("id", mcp.Description("Project ID or slug"), mcp.Required()),
	), handleDeleteProject(store))

	// ── Issues ──
	s.AddTool(mcp.NewTool("list_issues",
		mcp.WithDescription("List all issues in a project, optionally filtered by state or assignee"),
		mcp.WithString("project_id", mcp.Description("Project ID or slug"), mcp.Required()),
		mcp.WithString("state", mcp.Description("Filter by state")),
		mcp.WithNumber("assignee", mcp.Description("Filter by assignee user ID")),
	), handleListIssues(store))

	s.AddTool(mcp.NewTool("get_issue",
		mcp.WithDescription("Get an issue by ID or slug"),
		mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
	), handleGetIssue(store))

	s.AddTool(mcp.NewTool("create_issue",
		mcp.WithDescription("Create a new issue in a project — only title and project_id are required"),
		mcp.WithString("project_id", mcp.Description("Project ID or slug"), mcp.Required()),
		mcp.WithString("title", mcp.Description("Issue title"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Issue description")),
		mcp.WithString("type", mcp.Description("Issue type: epic, feature, bug, chore")),
		mcp.WithString("state", mcp.Description("Initial state")),
		mcp.WithNumber("priority", mcp.Description("Priority: 0=none, 1=urgent, 2=high, 3=medium, 4=low")),
	), handleCreateIssue(store))

	s.AddTool(mcp.NewTool("update_issue",
		mcp.WithDescription("Update an existing issue — only provided fields will change (pass 0 to leave priority/assignee unchanged)"),
		mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("description", mcp.Description("New description")),
		mcp.WithString("type", mcp.Description("New type")),
		mcp.WithString("state", mcp.Description("New state")),
		mcp.WithNumber("priority", mcp.Description("New priority")),
		mcp.WithNumber("assignee", mcp.Description("New assignee user ID")),
	), handleUpdateIssue(store))

	s.AddTool(mcp.NewTool("update_issue_state",
		mcp.WithDescription("Move an issue to a new state in the pipeline — valid states: backlog, todo, in_progress, qa, done, cancelled"),
		mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
		mcp.WithString("state", mcp.Description("Target state: backlog, todo, in_progress, qa, done, or cancelled"), mcp.Required()),
	), handleUpdateIssueState(store))

	s.AddTool(mcp.NewTool("delete_issue",
		mcp.WithDescription("Delete an issue permanently"),
		mcp.WithString("id", mcp.Description("Issue ID or slug"), mcp.Required()),
	), handleDeleteIssue(store))

	// ── Comments ──
	s.AddTool(mcp.NewTool("list_comments",
		mcp.WithDescription("List all comments on an issue, oldest first"),
		mcp.WithString("issue_id", mcp.Description("Issue ID or slug"), mcp.Required()),
	), handleListComments(store))

	s.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to an issue — body is required"),
		mcp.WithString("issue_id", mcp.Description("Issue ID or slug"), mcp.Required()),
		mcp.WithString("body", mcp.Description("Comment text"), mcp.Required()),
	), handleAddComment(store))

	return s
}

// ── Tool handlers ──

var _ = fmt.Sprint // keep import

func textResult(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func jsonResult(v any) *mcp.CallToolResult {
	r, err := mcp.NewToolResultJSON(v)
	if err != nil {
		return mcp.NewToolResultError("marshal: " + err.Error())
	}
	return r
}

func resolveProjectID(store *Store, id string) (int64, string) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		return n, ""
	}
	p, err := store.GetProjectBySlug(id)
	if err != nil {
		return 0, err.Error()
	}
	return p.ID, ""
}

func resolveIssueID(store *Store, id string) (int64, string) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		return n, ""
	}
	iss, err := store.GetIssueBySlug(id)
	if err != nil {
		return 0, err.Error()
	}
	return iss.ID, ""
}

func handleGetInfo(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		users, _ := store.ListUsers()
		projects, _ := store.ListProjects()
		me := mcpUserFromCtx(ctx)
		return jsonResult(map[string]any{
			"states":          ValidStates,
			"types":           ValidTypes,
			"priority_levels": ValidPriorityLevels,
			"priority_labels": map[int]string{0: "none", 1: "urgent", 2: "high", 3: "medium", 4: "low"},
			"users":           users,
			"projects":        projects,
			"me":              me,
		}), nil
	}
}

func handleListUsers(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		users, err := store.ListUsers()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(users), nil
	}
}

func handleGetUser(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := int64(req.GetFloat("id", 0))
		user, err := store.GetUser(id)
		if err != nil {
			return mcp.NewToolResultError("user not found"), nil
		}
		return jsonResult(user), nil
	}
}

func handleListProjects(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projects, err := store.ListProjects()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(projects), nil
	}
}

func handleGetProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		p, err := store.GetProject(id)
		if err != nil {
			return mcp.NewToolResultError("project not found"), nil
		}
		return jsonResult(p), nil
	}
}

func handleCreateProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		name := args["name"].(string)
		slug := args["slug"].(string)
		desc, _ := args["description"].(string)
		// Need a user context — for MCP we use user ID 1 (admin)
		p, err := store.CreateProject(name, slug, desc, 1)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(p), nil
	}
}

func handleUpdateProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		name, _ := args["name"].(string)
		slug, _ := args["slug"].(string)
		desc, _ := args["description"].(string)
		p, err := store.UpdateProject(id, name, slug, desc)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(p), nil
	}
}

func handleDeleteProject(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveProjectID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		if err := store.DeleteProject(id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return textResult("deleted"), nil
	}
}

func handleListIssues(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		projectStr := args["project_id"].(string)
		pid, errStr := resolveProjectID(store, projectStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		f := IssueFilter{}
		if s, ok := args["state"].(string); ok {
			f.State = s
		}
		if a, ok := args["assignee"].(float64); ok {
			f.Assignee = int64(a)
		}
		issues, err := store.ListIssues(pid, f)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(issues), nil
	}
}

func handleGetIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		iss, err := store.GetIssue(id)
		if err != nil {
			return mcp.NewToolResultError("issue not found"), nil
		}
		return jsonResult(iss), nil
	}
}

func handleCreateIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		projectStr := args["project_id"].(string)
		pid, errStr := resolveProjectID(store, projectStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		title := args["title"].(string)
		desc, _ := args["description"].(string)
		typ, _ := args["type"].(string)
		state, _ := args["state"].(string)
		priority := 3
		if p, ok := args["priority"].(float64); ok {
			priority = int(p)
		}
		iss, err := store.CreateIssue(pid, title, desc, typ, state, 0, 0, 1, priority)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(iss), nil
	}
}

func handleUpdateIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		title, _ := args["title"].(string)
		desc, _ := args["description"].(string)
		typ, _ := args["type"].(string)
		state, _ := args["state"].(string)
		var priority int
		if p, ok := args["priority"].(float64); ok {
			priority = int(p)
		}
		var assignee int64
		if a, ok := args["assignee"].(float64); ok {
			assignee = int64(a)
		}
		iss, err := store.UpdateIssue(id, title, desc, typ, state, assignee, 0, priority)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(iss), nil
	}
}

func handleUpdateIssueState(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["id"].(string)
		newState := args["state"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		iss, err := store.UpdateIssue(id, "", "", "", newState, 0, 0, 0)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(iss), nil
	}
}

func handleDeleteIssue(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		if err := store.DeleteIssue(id); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return textResult("deleted"), nil
	}
}

func handleListComments(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		idStr := req.GetString("issue_id", "")
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		comments, err := store.ListComments(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(comments), nil
	}
}

func handleAddComment(store *Store) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		idStr := args["issue_id"].(string)
		body := args["body"].(string)
		id, errStr := resolveIssueID(store, idStr)
		if errStr != "" {
			return mcp.NewToolResultError(errStr), nil
		}
		c, err := store.CreateComment(id, body, 1, 1)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(c), nil
	}
}

// ── HTTP handler ──

// ServeMCP handles the POST /mcp endpoint with PAT auth via ?pat= query param.
func (h *Handler) ServeMCP(w http.ResponseWriter, r *http.Request) {
	pat := r.URL.Query().Get("pat")
	if pat == "" {
		http.Error(w, `{"error":"missing pat"}`, http.StatusUnauthorized)
		return
	}
	user, err := h.store.GetUserByPAT(pat)
	if err != nil {
		http.Error(w, `{"error":"invalid pat"}`, http.StatusUnauthorized)
		return
	}
	ctx := context.WithValue(r.Context(), mcpUserKey, user)
	h.mcp.ServeHTTP(w, r.WithContext(ctx))
}

// InitMCP sets up the MCP server and stores it on the Handler.
func (h *Handler) InitMCP(store *Store) {
	mcpServer := NewMCPServer(store)
	h.mcp = server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStateful(true),
	)
}
