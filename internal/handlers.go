package internal

import (
	"encoding/json"
	"net/http"
	"github.com/mark3labs/mcp-go/server"
	"strconv"
)

type Handler struct {
	store *Store
	hub   *Hub
	mcp   *server.StreamableHTTPServer
	mcpUser *server.StreamableHTTPServer
}

func NewHandler(store *Store, hub *Hub) *Handler {
	return &Handler{store: store, hub: hub}
}

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	jsonResp(w, status, map[string]string{"error": msg})
}

// ── Info ──

type infoUser struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
}

type infoProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	users, _ := h.store.ListUsers()
	projects, _ := h.store.ListProjects()

	compactUsers := make([]infoUser, 0, len(users))
	for _, u := range users {
		compactUsers = append(compactUsers, infoUser{ID: u.ID, DisplayName: u.DisplayName})
	}
	compactProjects := make([]infoProject, 0, len(projects))
	for _, p := range projects {
		compactProjects = append(compactProjects, infoProject{ID: p.ID, Name: p.Name, Slug: p.Slug})
	}

	jsonResp(w, 200, map[string]any{
		"states":          ValidStates,
		"types":           ValidTypes,
		"priority_levels": ValidPriorityLevels,
		"priority_labels": map[int]string{0: "none", 1: "urgent", 2: "high", 3: "medium", 4: "low"},
		"users":           compactUsers,
		"projects":        compactProjects,
	})
}

// parseInt64 parses a string as an int64, returning 0 on failure.
func parseInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// ── Users ──

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	jsonResp(w, 200, UserFromCtx(r.Context()))
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.ListUsers()
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, users)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user, err := h.store.GetUser(id)
	if err != nil {
		jsonErr(w, 404, "user not found")
		return
	}
	jsonResp(w, 200, user)
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	admin := UserFromCtx(r.Context())
	if !admin.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	var body struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		Admin       bool   `json:"admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Username == "" {
		jsonErr(w, 400, "username is required")
		return
	}
	pat := generatePAT()
	user, err := h.store.CreateUser(body.Username, body.DisplayName, pat, body.Admin)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 201, user)
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	var body struct {
		DisplayName string `json:"display_name"`
		PAT         string `json:"pat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	user, err := h.store.UpdateUser(id, body.DisplayName, body.PAT)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, user)
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	user := UserFromCtx(r.Context())
	if !user.IsAdmin {
		jsonErr(w, 403, "admin required")
		return
	}
	id := parseInt64(r.PathValue("id"))
	if err := h.store.DeleteUser(id); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}

// ── Projects ──

func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.store.ListProjects()
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, projects)
}

func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Name == "" {
		jsonErr(w, 400, "name is required")
		return
	}
	user := UserFromCtx(r.Context())
	p, err := h.store.CreateProject(body.Name, body.Slug, body.Description, user.ID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectCreated, Payload: p, By: user.ID})
	jsonResp(w, 201, p)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	p, err := h.store.GetProject(id)
	if err != nil {
		jsonErr(w, 404, "project not found")
		return
	}
	jsonResp(w, 200, p)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user := UserFromCtx(r.Context())
	var body struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	existing, err := h.store.GetProject(id)
	if err != nil {
		jsonErr(w, 404, "project not found")
		return
	}
	before := *existing
	p, err := h.store.UpdateProject(id, body.Name, body.Slug, body.Description)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectUpdated,
		Payload: map[string]any{"id": p.ID, "changed": diffProject(&before, p)},
		By:      user.ID})
	jsonResp(w, 200, p)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := parseInt64(r.PathValue("id"))
	user := UserFromCtx(r.Context())
	if err := h.store.DeleteProject(id); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventProjectDeleted,
		Payload: map[string]int64{"id": id}, By: user.ID})
	w.WriteHeader(204)
}

// resolveIssue tries to find an issue by int64 ID first, then by slug.
func (h *Handler) resolveIssue(id string) (*Issue, error) {
	if n, err := strconv.ParseInt(id, 10, 64); err == nil {
		iss, err := h.store.GetIssue(n)
		if err == nil {
			return iss, nil
		}
	}
	return h.store.GetIssueBySlug(id)
}

// ── Issues ──

func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	projectID := parseInt64(r.PathValue("id"))

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	f := IssueFilter{
		Type:      r.URL.Query().Get("type"),
		State:     r.URL.Query().Get("state"),
		Assignee:  parseInt64(r.URL.Query().Get("assignee")),
		CreatedBy: parseInt64(r.URL.Query().Get("created_by")),
		Query:     r.URL.Query().Get("q"),
		Page:      page,
		PerPage:   perPage,
	}

	// assigned_to_me convenience
	if r.URL.Query().Get("assigned_to_me") == "true" && f.Assignee == 0 {
		user := UserFromCtx(r.Context())
		f.Assignee = user.ID
	}

	issues, err := h.store.ListIssues(projectID, f)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, issues)
}

func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	projectID := parseInt64(r.PathValue("id"))
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
		Assignee    int64  `json:"assignee"`
		Priority    int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Title == "" {
		jsonErr(w, 400, "title is required")
		return
	}
	user := UserFromCtx(r.Context())
	iss, err := h.store.CreateIssue(projectID, body.Title, body.Description, body.Type, body.State, body.Assignee, 0, user.ID, body.Priority)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueCreated, Payload: iss, By: user.ID})
	jsonResp(w, 201, iss)
}

func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	jsonResp(w, 200, iss)
}

func (h *Handler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	user := UserFromCtx(r.Context())
	before := *iss
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
		Assignee    int64  `json:"assignee"`
		Priority    int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	iss, err = h.store.UpdateIssue(iss.ID, body.Title, body.Description, body.Type, body.State, body.Assignee, 0, body.Priority)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueUpdated,
		Payload: map[string]any{"id": iss.ID, "changed": diffIssue(&before, iss)},
		By:      user.ID})
	jsonResp(w, 200, iss)
}

func (h *Handler) UpdateIssueState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	var body struct {
		State string `json:"state"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.State == "" {
		jsonErr(w, 400, "state is required")
		return
	}
	valid := false
	for _, s := range ValidStates {
		if s == body.State {
			valid = true
			break
		}
	}
	if !valid {
		jsonErr(w, 400, "invalid state")
		return
	}
	user := UserFromCtx(r.Context())
	before := *iss
	iss, err = h.store.UpdateIssue(iss.ID, "", "", "", body.State, 0, 0, 0)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueUpdated,
		Payload: map[string]any{"id": iss.ID, "changed": diffIssue(&before, iss)},
		By:      user.ID})
	jsonResp(w, 200, iss)
}

func (h *Handler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	user := UserFromCtx(r.Context())
	if err := h.store.DeleteIssue(iss.ID); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventIssueDeleted,
		Payload: map[string]int64{"id": iss.ID, "project_id": iss.ProjectID},
		By:      user.ID})
	w.WriteHeader(204)
}

// ── Comments ──

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	comments, err := h.store.ListComments(iss.ID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.resolveIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	if body.Body == "" {
		jsonErr(w, 400, "body is required")
		return
	}
	user := UserFromCtx(r.Context())
	c, err := h.store.CreateComment(iss.ID, body.Body, user.ID, user.ID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	h.hub.Broadcast(Event{Type: EventCommentCreated, Payload: c, By: user.ID})
	jsonResp(w, 201, c)
}
