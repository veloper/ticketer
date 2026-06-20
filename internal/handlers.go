package internal

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func jsonResp(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	jsonResp(w, status, map[string]string{"error": msg})
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
	id := r.PathValue("id")
	user, err := h.store.GetUser(id)
	if err != nil {
		jsonErr(w, 404, "user not found")
		return
	}
	jsonResp(w, 200, user)
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
	p, err := h.store.CreateProject(body.Name, body.Description, user.ID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 201, p)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.store.GetProject(id)
	if err != nil {
		jsonErr(w, 404, "project not found")
		return
	}
	jsonResp(w, 200, p)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	p, err := h.store.UpdateProject(id, body.Name, body.Description)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, p)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.DeleteProject(id); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 204, nil)
}

// ── Issues ──

func (h *Handler) ListIssues(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	f := IssueFilter{
		Type:      r.URL.Query().Get("type"),
			State:     r.URL.Query().Get("state"),
		Assignee:  r.URL.Query().Get("assignee"),
		CreatedBy: r.URL.Query().Get("created_by"),
		Query:     r.URL.Query().Get("q"),
		Page:      page,
		PerPage:   perPage,
	}

	// assigned_to_me convenience
	if r.URL.Query().Get("assigned_to_me") == "true" && f.Assignee == "" {
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
	projectID := r.PathValue("id")
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
		Assignee    string `json:"assignee"`
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
	iss, err := h.store.CreateIssue(projectID, body.Title, body.Description, body.Type, body.State, body.Assignee, "", user.ID, body.Priority)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 201, iss)
}

func (h *Handler) GetIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iss, err := h.store.GetIssue(id)
	if err != nil {
		jsonErr(w, 404, "issue not found")
		return
	}
	jsonResp(w, 200, iss)
}

func (h *Handler) UpdateIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Type        string `json:"type"`
		State       string `json:"state"`
		Assignee    string `json:"assignee"`
		Priority    int    `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonErr(w, 400, "invalid json")
		return
	}
	iss, err := h.store.UpdateIssue(id, body.Title, body.Description, body.Type, body.State, body.Assignee, "", body.Priority)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, iss)
}

func (h *Handler) DeleteIssue(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.DeleteIssue(id); err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 204, nil)
}

// ── Comments ──

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("id")
	comments, err := h.store.ListComments(issueID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 200, comments)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	issueID := r.PathValue("id")
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
	c, err := h.store.CreateComment(issueID, body.Body, user.ID, user.ID)
	if err != nil {
		jsonErr(w, 500, err.Error())
		return
	}
	jsonResp(w, 201, c)
}
