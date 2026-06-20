package internal

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

// ── Fixtures ──

var testUsers = []SeedUser{
	{Username: "alice", DisplayName: "Alice", PAT: "pat_alice"},
	{Username: "bob", DisplayName: "Bob Builder", PAT: "pat_bob"},
	{Username: "carol", DisplayName: "Carol Tester", PAT: "pat_carol"},
}

// ── Store helpers ──

// newTestStore creates an in-memory store with seed users loaded.
// It registers cleanup via t.Cleanup.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.SeedUsers(testUsers); err != nil {
		t.Fatalf("SeedUsers: %v", err)
	}
	return s
}

// getUserByPAT is a test helper to look up a user from fixtures.
func getUserByPAT(t *testing.T, s *Store, pat string) *User {
	t.Helper()
	u, err := s.GetUserByPAT(pat)
	if err != nil {
		t.Fatalf("GetUserByPAT(%q): %v", pat, err)
	}
	return u
}

// ── Handler helpers ──

// newTestHandler creates a Handler backed by a test store.
func newTestHandler(t *testing.T) (*Store, *Handler) {
	t.Helper()
	s := newTestStore(t)
	hub := NewHub()
	h := NewHandler(s, hub)
	return s, h
}

// handlerRequest builds an http.Request suitable for testing handlers directly
// (without the PAT middleware). The user is injected into the request context.
func handlerRequest(t *testing.T, method, target string, body any, user *User) *http.Request {
	t.Helper()
	var req *http.Request
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		req = httptest.NewRequest(method, target, io.NopCloser(strings.NewReader(string(b))))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	if user != nil {
		ctx := context.WithValue(req.Context(), userKey, user)
		req = req.WithContext(ctx)
	}
	return req
}

// serveHandler executes a handler and returns the response recorder.
func serveHandler(h *Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	// Route based on method + path pattern.
	// We match a subset of routes manually — add more as tests grow.
	var pattern string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/info", h.Info)
	mux.HandleFunc("GET /api/me", h.Me)
	mux.HandleFunc("GET /api/users", h.ListUsers)
	mux.HandleFunc("GET /api/users/{id}", h.GetUser)
	mux.HandleFunc("POST /api/users", h.CreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", h.DeleteUser)
	mux.HandleFunc("GET /api/projects", h.ListProjects)
	mux.HandleFunc("POST /api/projects", h.CreateProject)
	mux.HandleFunc("GET /api/projects/{id}", h.GetProject)
	mux.HandleFunc("PATCH /api/projects/{id}", h.UpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", h.DeleteProject)
	mux.HandleFunc("GET /api/projects/{id}/issues", h.ListIssues)
	mux.HandleFunc("POST /api/projects/{id}/issues", h.CreateIssue)
	mux.HandleFunc("GET /api/issues/{id}", h.GetIssue)
	mux.HandleFunc("PATCH /api/issues/{id}", h.UpdateIssue)
	mux.HandleFunc("PUT /api/issues/{id}/state", h.UpdateIssueState)
	mux.HandleFunc("DELETE /api/issues/{id}", h.DeleteIssue)
	mux.HandleFunc("GET /api/issues/{id}/comments", h.ListComments)
	mux.HandleFunc("POST /api/issues/{id}/comments", h.CreateComment)
	_ = pattern
	mux.ServeHTTP(rr, req)
	return rr
}

// mustDecode decodes a JSON response body into v. Fatal on error.
func mustDecode(t *testing.T, body io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("json.Decode: %v", err)
	}
}
