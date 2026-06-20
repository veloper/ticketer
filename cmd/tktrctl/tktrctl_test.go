package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/daniel/roomy-ticketer/internal"
)

// setupTestServer creates a real Ticketer server on a random port and
// returns the host URL. The server has an admin user (admin / pat_admin)
// and clean state.
func setupTestServer(t *testing.T) string {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := internal.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { store.Close() })

	if err := store.EnsureAdmin("admin", "pat_admin"); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}

	hub := internal.NewHub()
	go hub.Run()

	handler := internal.NewHandler(store, hub)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", handler.Info)
	mux.HandleFunc("GET /api/me", handler.Me)
	mux.HandleFunc("GET /api/users", handler.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", handler.GetUser)
	mux.HandleFunc("POST /api/users", handler.CreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", handler.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", handler.DeleteUser)
	mux.HandleFunc("GET /api/projects", handler.ListProjects)
	mux.HandleFunc("POST /api/projects", handler.CreateProject)
	mux.HandleFunc("GET /api/projects/{id}", handler.GetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", handler.UpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", handler.DeleteProject)
	mux.HandleFunc("GET /api/projects/{id}/issues", handler.ListIssues)
	mux.HandleFunc("POST /api/projects/{id}/issues", handler.CreateIssue)
	mux.HandleFunc("GET /api/issues/{id}", handler.GetIssue)
	mux.HandleFunc("PATCH /api/issues/{id}", handler.UpdateIssue)
	mux.HandleFunc("DELETE /api/issues/{id}", handler.DeleteIssue)
	mux.HandleFunc("PUT /api/issues/{id}/state", handler.UpdateIssueState)

	server := httptest.NewServer(internal.PATMiddleware(store)(mux))
	t.Cleanup(server.Close)

	return server.URL
}

func TestCLIInfo(t *testing.T) {
	srv := setupTestServer(t)

	// Set globals used by the api() helper
	host = srv
	pat = "pat_admin"

	var info map[string]any
	if err := get("info", &info); err != nil {
		t.Fatalf("get info: %v", err)
	}

	states, ok := info["states"].([]any)
	if !ok || len(states) != 6 {
		t.Errorf("expected 6 states, got %d", len(states))
	}
	types, ok := info["types"].([]any)
	if !ok || len(types) != 4 {
		t.Errorf("expected 4 types, got %d", len(types))
	}
}

func TestCLIUsers(t *testing.T) {
	srv := setupTestServer(t)
	host = srv
	pat = "pat_admin"

	// users list should show the admin user
	var users []map[string]any
	if err := get("users", &users); err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0]["username"] != "admin" {
		t.Errorf("expected username admin, got %v", users[0]["username"])
	}

	// users create a new user
	var created map[string]any
	if err := post("users", map[string]any{
		"username": "bot",
		"admin":    false,
	}, &created); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if created["username"] != "bot" {
		t.Errorf("expected username bot, got %v", created["username"])
	}
	patOut, _ := created["pat"].(string)
	if patOut == "" {
		t.Error("expected generated pat in response, got", created["pat"])
	}

	// users show the new user
	var shown map[string]any
	if err := get("users/2", &shown); err != nil {
		t.Fatalf("show user: %v", err)
	}
	if shown["username"] != "bot" {
		t.Errorf("expected username bot, got %v", shown["username"])
	}
	// PAT should not be exposed on GET
	if _, has := shown["pat"]; has {
		t.Error("pat should not be exposed on user GET")
	}
}

func TestCLIProjects(t *testing.T) {
	srv := setupTestServer(t)
	host = srv
	pat = "pat_admin"

	// projects create
	var p map[string]any
	if err := post("projects", map[string]string{
		"name": "Asteroid Game",
		"slug": "ASTEROID-GAME",
	}, &p); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if p["slug"] != "ASTEROID-GAME" {
		t.Errorf("expected slug ASTEROID-GAME, got %v", p["slug"])
	}
	projectID := fmt.Sprintf("%.0f", p["id"])

	// projects list
	var projects []map[string]any
	if err := get("projects", &projects); err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}

	// projects show
	var shown map[string]any
	if err := get("projects/"+projectID, &shown); err != nil {
		t.Fatalf("show project: %v", err)
	}
	if shown["name"] != "Asteroid Game" {
		t.Errorf("expected name 'Asteroid Game', got %v", shown["name"])
	}

	// projects update
	var updated map[string]any
	if err := api("PATCH", "projects/"+projectID, map[string]string{
		"description": "A space game",
	}, &updated); err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated["description"] != "A space game" {
		t.Errorf("expected description 'A space game', got %v", updated["description"])
	}
}

func TestCLIIssues(t *testing.T) {
	srv := setupTestServer(t)
	host = srv
	pat = "pat_admin"

	// Create a project first
	var p map[string]any
	if err := post("projects", map[string]string{
		"name": "Game",
		"slug": "GAME",
	}, &p); err != nil {
		t.Fatalf("create project: %v", err)
	}
	projectID := fmt.Sprintf("%.0f", p["id"])
	slug := p["slug"].(string)

	// issues create
	var iss map[string]any
	if err := post("projects/"+projectID+"/issues", map[string]any{
		"title": "Add rotation",
		"type":  "feature",
	}, &iss); err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if iss["title"] != "Add rotation" {
		t.Errorf("expected title 'Add rotation', got %v", iss["title"])
	}
	if iss["slug"] != slug+"-1" {
		t.Errorf("expected slug %s-1, got %v", slug, iss["slug"])
	}
	issueID := fmt.Sprintf("%.0f", iss["id"])
	issueSlug := iss["slug"].(string)

	// issues list
	var issues []map[string]any
	if err := get("projects/"+projectID+"/issues", &issues); err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 issue, got %d", len(issues))
	}

	// issues show by id
	var shown map[string]any
	if err := get("issues/"+issueID, &shown); err != nil {
		t.Fatalf("show issue by id: %v", err)
	}
	if shown["title"] != "Add rotation" {
		t.Errorf("expected title 'Add rotation', got %v", shown["title"])
	}

	// issues show by slug
	if err := get("issues/"+issueSlug, &shown); err != nil {
		t.Fatalf("show issue by slug: %v", err)
	}
	if shown["title"] != "Add rotation" {
		t.Errorf("expected title 'Add rotation' via slug, got %v", shown["title"])
	}

	// issues state - show current state
	var stateOnly map[string]any
	if err := get("issues/"+issueID, &stateOnly); err != nil {
		t.Fatalf("get issue state: %v", err)
	}
	if stateOnly["state"] != "todo" {
		t.Errorf("expected initial state todo, got %v", stateOnly["state"])
	}

	// issues state-update
	var updated map[string]any
	if err := put("issues/"+issueID+"/state", map[string]string{
		"state": "qa",
	}, &updated); err != nil {
		t.Fatalf("state update: %v", err)
	}
	if updated["state"] != "qa" {
		t.Errorf("expected state qa, got %v", updated["state"])
	}

	// Verify state via slug
	var verified map[string]any
	if err := get("issues/"+issueSlug, &verified); err != nil {
		t.Fatalf("verify state: %v", err)
	}
	if verified["state"] != "qa" {
		t.Errorf("verified state = %v, want qa", verified["state"])
	}

	// issues update
	var afterUpdate map[string]any
	if err := api("PATCH", "issues/"+issueID, map[string]any{
		"title": "Add ship rotation",
	}, &afterUpdate); err != nil {
		t.Fatalf("update issue: %v", err)
	}
	if afterUpdate["title"] != "Add ship rotation" {
		t.Errorf("expected title 'Add ship rotation', got %v", afterUpdate["title"])
	}
	if afterUpdate["state"] != "qa" {
		t.Errorf("state should remain qa after title update, got %v", afterUpdate["state"])
	}
}

func TestCLIAuth(t *testing.T) {
	srv := setupTestServer(t)
	host = srv
	pat = "pat_admin"

	// Verify that an invalid PAT gets rejected
	pat = "bad_pat"
	var info map[string]any
	err := get("info", &info)
	if err == nil {
		t.Error("expected error with bad PAT")
	}
}

func TestCLIAdminRequired(t *testing.T) {
	srv := setupTestServer(t)
	host = srv

	// Create a non-admin user
	pat = "pat_admin"
	var created map[string]any
	if err := post("users", map[string]any{
		"username": "regular",
		"admin":    false,
	}, &created); err != nil {
		t.Fatalf("create regular user: %v", err)
	}

	// Use the non-admin user's PAT
	pat, _ = created["pat"].(string)

	// Try to create another user — should fail with 403
	err := post("users", map[string]any{
		"username": "another",
		"admin":    false,
	}, nil)
	if err == nil {
		t.Error("expected 403 for non-admin user creation")
	}
}
