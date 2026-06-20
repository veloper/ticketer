package internal

import (
	"fmt"
	"io"
	"testing"
)

// ── Me ──

func TestMe(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", "/api/me", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got User
	mustDecode(t, rr.Body, &got)
	if got.Username != "alice" {
		t.Errorf("expected username alice, got %s", got.Username)
	}
}

// ── Users ──

func TestListUsersHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", "/api/users", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var users []User
	mustDecode(t, rr.Body, &users)
	if len(users) != len(testUsers) {
		t.Errorf("expected %d users, got %d", len(testUsers), len(users))
	}
}

func TestGetUserHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/users/%d", alice.ID), nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got User
	mustDecode(t, rr.Body, &got)
	if got.Username != "alice" {
		t.Errorf("expected username alice, got %s", got.Username)
	}
}

func TestGetUserHandler_notFound(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", "/api/users/nonexistent", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 404)
	assertErrorBody(t, rr.Body, "user not found")
}

// ── Projects ──

func TestCreateProjectHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "POST", "/api/projects", map[string]string{
		"name": "My Project", "description": "A test project",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 201)
	var p Project
	mustDecode(t, rr.Body, &p)
	if p.Name != "My Project" {
		t.Errorf("expected name %q, got %q", "My Project", p.Name)
	}
	if p.CreatedBy != alice.ID {
		t.Errorf("expected created_by %q, got %q", alice.ID, p.CreatedBy)
	}
}

func TestCreateProjectHandler_missingName(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "POST", "/api/projects", map[string]string{
		"name": "",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 400)
	assertErrorBody(t, rr.Body, "name is required")
}

func TestCreateProjectHandler_invalidJSON(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "POST", "/api/projects", "not json", alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 400)
	assertErrorBody(t, rr.Body, "invalid json")
}

func TestListProjectsHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	// Create two projects
	mustCreateProject(t, s, "Project A", alice.ID)
	mustCreateProject(t, s, "Project B", alice.ID)

	req := handlerRequest(t, "GET", "/api/projects", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var projects []Project
	mustDecode(t, rr.Body, &projects)
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestGetProjectHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "My Project", alice.ID)

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/projects/%d", p.ID), nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Project
	mustDecode(t, rr.Body, &got)
	if got.Name != "My Project" {
		t.Errorf("expected name %q, got %q", "My Project", got.Name)
	}
}

func TestGetProjectHandler_notFound(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", "/api/projects/nonexistent", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 404)
	assertErrorBody(t, rr.Body, "project not found")
}

func TestUpdateProjectHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Original", alice.ID)

	req := handlerRequest(t, "PATCH", fmt.Sprintf("/api/projects/%d", p.ID), map[string]string{
		"name": "Updated", "description": "New desc",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Project
	mustDecode(t, rr.Body, &got)
	if got.Name != "Updated" {
		t.Errorf("expected name %q, got %q", "Updated", got.Name)
	}
}

func TestDeleteProjectHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "To Delete", alice.ID)

	req := handlerRequest(t, "DELETE", fmt.Sprintf("/api/projects/%d", p.ID), nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 204)
	if rr.Body.Len() > 0 {
		t.Error("expected empty body for 204")
	}

	// Verify deleted
	_, err := s.GetProject(p.ID)
	if err == nil {
		t.Error("expected project to be deleted")
	}
}

// ── Issues ──

func TestCreateIssueHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Game", alice.ID)

	req := handlerRequest(t, "POST", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues", map[string]any{
		"title":       "Add feature",
		"description": "Do the thing",
		"type":        "feature",
		"state":       "todo",
		"assignee":    alice.ID,
		"priority":    2,
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 201)
	var iss Issue
	mustDecode(t, rr.Body, &iss)
	if iss.Title != "Add feature" {
		t.Errorf("expected title %q, got %q", "Add feature", iss.Title)
	}
	if iss.Type != "feature" {
		t.Errorf("expected type feature, got %s", iss.Type)
	}
	if iss.State != "todo" {
		t.Errorf("expected state todo, got %s", iss.State)
	}
	if iss.Priority != 2 {
		t.Errorf("expected priority 2, got %d", iss.Priority)
	}
	if iss.Slug == "" {
		t.Error("expected non-empty slug")
	}
}

func TestCreateIssueHandler_slugFromProjectSlug(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	// Create project with custom slug via API
	req := handlerRequest(t, "POST", "/api/projects", map[string]string{
		"name": "My Project", "slug": "MY-PROJ",
	}, alice)
	rr := serveHandler(h, req)
	assertStatus(t, rr.Code, 201)
	var p Project
	mustDecode(t, rr.Body, &p)
	if p.Slug != "MY-PROJ" {
		t.Fatalf("expected slug MY-PROJ, got %s", p.Slug)
	}

	// Create first issue
	req2 := handlerRequest(t, "POST", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues", map[string]any{
		"title": "First issue",
	}, alice)
	rr2 := serveHandler(h, req2)
	assertStatus(t, rr2.Code, 201)
	var iss1 Issue
	mustDecode(t, rr2.Body, &iss1)
	if iss1.Slug != "MY-PROJ-1" {
		t.Errorf("expected slug MY-PROJ-1, got %s", iss1.Slug)
	}

	// Create second issue
	req3 := handlerRequest(t, "POST", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues", map[string]any{
		"title": "Second issue",
	}, alice)
	rr3 := serveHandler(h, req3)
	assertStatus(t, rr3.Code, 201)
	var iss2 Issue
	mustDecode(t, rr3.Body, &iss2)
	if iss2.Slug != "MY-PROJ-2" {
		t.Errorf("expected slug MY-PROJ-2, got %s", iss2.Slug)
	}
}

func TestCreateIssueHandler_missingTitle(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Game", alice.ID)

	req := handlerRequest(t, "POST", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues", map[string]string{
		"title": "",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 400)
	assertErrorBody(t, rr.Body, "title is required")
}

func TestListIssuesHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Game", alice.ID)
	mustCreateIssue(t, s, p.ID, "One", alice.ID)
	mustCreateIssue(t, s, p.ID, "Two", alice.ID)

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var issues []Issue
	mustDecode(t, rr.Body, &issues)
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestListIssuesHandler_filterByState(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Filter", alice.ID)
	s.CreateIssue(p.ID, "One", "", "", "todo", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Two", "", "", "qa", 0, 0, alice.ID, 0)

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues?state=todo", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var issues []Issue
	mustDecode(t, rr.Body, &issues)
	if len(issues) != 1 {
		t.Errorf("expected 1 todo issue, got %d", len(issues))
	}
}

func TestListIssuesHandler_assignedToMe(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	bob := getUserByPAT(t, s, "pat_bob")
	p := mustCreateProject(t, s, "Assign", alice.ID)
	s.CreateIssue(p.ID, "Alice's", "", "", "", alice.ID, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Bob's", "", "", "", bob.ID, 0, bob.ID, 0)

	// alice requests assigned_to_me
	req := handlerRequest(t, "GET", fmt.Sprintf("/api/projects/%d", p.ID)+"/issues?assigned_to_me=true", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var issues []Issue
	mustDecode(t, rr.Body, &issues)
	if len(issues) != 1 || issues[0].Title != "Alice's" {
		t.Errorf("expected 1 issue assigned to alice, got %d", len(issues))
	}
}

func TestGetIssueHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Game", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "My Issue", alice.ID)

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/issues/%d", iss.ID), nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.Title != "My Issue" {
		t.Errorf("expected title %q, got %q", "My Issue", got.Title)
	}
}

func TestGetIssueHandler_notFound(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	req := handlerRequest(t, "GET", "/api/issues/nonexistent", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 404)
	assertErrorBody(t, rr.Body, "issue not found")
}

func TestGetIssueHandler_bySlug(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "SlugLookup", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "Find by slug", alice.ID)

	// Fetch using the issue slug instead of UUID
	req := handlerRequest(t, "GET", "/api/issues/"+iss.Slug, nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.Title != "Find by slug" {
		t.Errorf("expected title %q, got %q", "Find by slug", got.Title)
	}
	if got.ID != iss.ID {
		t.Errorf("expected id %q, got %q", iss.ID, got.ID)
	}
}

func TestUpdateIssueHandler_bySlug(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "SlugUpdate", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "Old title", alice.ID)

	// PATCH using slug instead of UUID
	req := handlerRequest(t, "PATCH", "/api/issues/"+iss.Slug, map[string]any{
		"title": "Updated via slug",
		"state": "qa",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.Title != "Updated via slug" {
		t.Errorf("expected title %q, got %q", "Updated via slug", got.Title)
	}
	if got.State != "qa" {
		t.Errorf("expected state review, got %s", got.State)
	}
}

func TestCreateCommentHandler_byIssueSlug(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "SlugComment", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "Issue for slug comment", alice.ID)

	// POST comment using slug instead of UUID
	req := handlerRequest(t, "POST", "/api/issues/"+iss.Slug+"/comments", map[string]string{
		"body": "Comment via slug",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 201)
	var c Comment
	mustDecode(t, rr.Body, &c)
	if c.Body != "Comment via slug" {
		t.Errorf("expected body %q, got %q", "Comment via slug", c.Body)
	}

	// Verify the comment is attached to the right issue
	comments, _ := s.ListComments(iss.ID)
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

func TestUpdateIssueHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	bob := getUserByPAT(t, s, "pat_bob")
	p := mustCreateProject(t, s, "Update", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "Old title", alice.ID)

	req := handlerRequest(t, "PATCH", fmt.Sprintf("/api/issues/%d", iss.ID), map[string]any{
		"title":    "New title",
		"state":    "qa",
		"assignee": bob.ID,
		"priority": 1,
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.Title != "New title" {
		t.Errorf("expected title %q, got %q", "New title", got.Title)
	}
	if got.State != "qa" {
		t.Errorf("expected state review, got %s", got.State)
	}
	if got.Assignee != bob.ID {
		t.Errorf("expected assignee %q, got %q", bob.ID, got.Assignee)
	}
	if got.Priority != 1 {
		t.Errorf("expected priority 1, got %d", got.Priority)
	}
}

func TestDeleteIssueHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Delete", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "To Delete", alice.ID)

	req := handlerRequest(t, "DELETE", fmt.Sprintf("/api/issues/%d", iss.ID), nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 204)

	// Verify deleted
	_, err := s.GetIssue(iss.ID)
	if err == nil {
		t.Error("expected issue to be deleted")
	}
}

// ── Comments ──

func TestInfo(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")

	// Create a project so it shows up in info
	mustCreateProject(t, s, "Info Project", alice.ID)
	// alice, bob, carol are seeded from testUsers

	req := handlerRequest(t, "GET", "/api/info", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var info struct {
		States          []string          `json:"states"`
		Types           []string          `json:"types"`
		PriorityLevels  []int             `json:"priority_levels"`
		PriorityLabels  map[string]string `json:"priority_labels"`
		Users           []struct {
			ID          int64  `json:"id"`
			DisplayName string `json:"display_name"`
		} `json:"users"`
		Projects []struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		} `json:"projects"`
	}
	mustDecode(t, rr.Body, &info)
	if len(info.States) != 6 {
		t.Errorf("expected 6 states, got %d", len(info.States))
	}
	if len(info.Types) != 4 {
		t.Errorf("expected 4 types, got %d", len(info.Types))
	}
	if info.PriorityLabels["1"] != "urgent" {
		t.Errorf("expected priority 1 = urgent, got %s", info.PriorityLabels["1"])
	}
	if len(info.Users) != 3 {
		t.Errorf("expected 3 users, got %d", len(info.Users))
	}
	if len(info.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(info.Projects))
	}
	if info.Projects[0].Name != "Info Project" {
		t.Errorf("expected project name 'Info Project', got %s", info.Projects[0].Name)
	}
}

func TestUpdateIssueStateHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "StateTest", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "State change", alice.ID)

	// PUT state to qa
	req := handlerRequest(t, "PUT", fmt.Sprintf("/api/issues/%d/state", iss.ID), map[string]string{
		"state": "qa",
	}, alice)
	rr := serveHandler(h, req)
	assertStatus(t, rr.Code, 200)

	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.State != "qa" {
		t.Errorf("expected state qa, got %s", got.State)
	}

	// Verify via store
	stored, _ := s.GetIssue(iss.ID)
	if stored.State != "qa" {
		t.Errorf("persisted state = %s, want qa", stored.State)
	}
}

func TestUpdateIssueStateHandler_invalid(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "StateTest", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "Invalid state", alice.ID)

	// PUT with invalid state
	req := handlerRequest(t, "PUT", fmt.Sprintf("/api/issues/%d/state", iss.ID), map[string]string{
		"state": "nonexistent",
	}, alice)
	rr := serveHandler(h, req)
	assertStatus(t, rr.Code, 400)
	assertErrorBody(t, rr.Body, "invalid state")

	// PUT with missing state
	req2 := handlerRequest(t, "PUT", fmt.Sprintf("/api/issues/%d/state", iss.ID), map[string]string{
		"state": "",
	}, alice)
	rr2 := serveHandler(h, req2)
	assertStatus(t, rr2.Code, 400)
	assertErrorBody(t, rr2.Body, "state is required")
}

func TestUpdateIssueStateHandler_bySlug(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "SlugState", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "State by slug", alice.ID)

	// PUT using slug instead of ID
	req := handlerRequest(t, "PUT", "/api/issues/"+iss.Slug+"/state", map[string]string{
		"state": "done",
	}, alice)
	rr := serveHandler(h, req)
	assertStatus(t, rr.Code, 200)
	var got Issue
	mustDecode(t, rr.Body, &got)
	if got.State != "done" {
		t.Errorf("expected state done, got %s", got.State)
	}
}

func TestCreateCommentHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Comments", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "The issue", alice.ID)

	req := handlerRequest(t, "POST", fmt.Sprintf("/api/issues/%d", iss.ID)+"/comments", map[string]string{
		"body": "A comment",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 201)
	var c Comment
	mustDecode(t, rr.Body, &c)
	if c.Body != "A comment" {
		t.Errorf("expected body %q, got %q", "A comment", c.Body)
	}
	if c.Author != alice.ID {
		t.Errorf("expected author %q, got %q", alice.ID, c.Author)
	}
}

func TestCreateCommentHandler_missingBody(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Comments", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "The issue", alice.ID)

	req := handlerRequest(t, "POST", fmt.Sprintf("/api/issues/%d", iss.ID)+"/comments", map[string]string{
		"body": "",
	}, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 400)
	assertErrorBody(t, rr.Body, "body is required")
}

func TestListCommentsHandler(t *testing.T) {
	s, h := newTestHandler(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p := mustCreateProject(t, s, "Comments", alice.ID)
	iss := mustCreateIssue(t, s, p.ID, "The issue", alice.ID)
	s.CreateComment(iss.ID, "First", alice.ID, alice.ID)
	s.CreateComment(iss.ID, "Second", alice.ID, alice.ID)

	req := handlerRequest(t, "GET", fmt.Sprintf("/api/issues/%d", iss.ID)+"/comments", nil, alice)
	rr := serveHandler(h, req)

	assertStatus(t, rr.Code, 200)
	var comments []Comment
	mustDecode(t, rr.Body, &comments)
	if len(comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(comments))
	}
}

// ── Helpers ──

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("status = %d, want %d", got, want)
	}
}

func assertErrorBody(t *testing.T, body io.Reader, msg string) {
	t.Helper()
	var errResp struct {
		Error string `json:"error"`
	}
	mustDecode(t, body, &errResp)
	if errResp.Error != msg {
		t.Errorf("error body = %q, want %q", errResp.Error, msg)
	}
}

// mustCreateProject creates a project and fatals on error.
func mustCreateProject(t *testing.T, s *Store, name string, createdBy int64) *Project {
	t.Helper()
	p, err := s.CreateProject(name, "", "", createdBy)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	return p
}

// mustCreateIssue creates an issue and fatals on error.
func mustCreateIssue(t *testing.T, s *Store, projectID int64, title string, createdBy int64) *Issue {
	t.Helper()
	iss, err := s.CreateIssue(projectID, title, "", "", "", 0, 0, createdBy, 0)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	return iss
}
