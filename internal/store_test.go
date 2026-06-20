package internal

import (
	"testing"
)

// ── Store lifecycle ──

func TestNewStore(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	s, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer s.Close()
}

func TestSeedUsers_idempotent(t *testing.T) {
	s := newTestStore(t)
	// Second seed should be a no-op (INSERT OR IGNORE)
	if err := s.SeedUsers(testUsers); err != nil {
		t.Fatalf("SeedUsers (second): %v", err)
	}
	users, _ := s.ListUsers()
	if len(users) != len(testUsers) {
		t.Errorf("expected %d users after re-seed, got %d", len(testUsers), len(users))
	}
}

// ── Users ──

func TestGetUserByPAT(t *testing.T) {
	s := newTestStore(t)
	u, err := s.GetUserByPAT("pat_alice")
	if err != nil {
		t.Fatalf("GetUserByPAT: %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("expected username alice, got %s", u.Username)
	}
}

func TestGetUserByPAT_invalid(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetUserByPAT("invalid_pat")
	if err == nil {
		t.Error("expected error for invalid PAT")
	}
}

func TestListUsers(t *testing.T) {
	s := newTestStore(t)
	users, err := s.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != len(testUsers) {
		t.Errorf("expected %d users, got %d", len(testUsers), len(users))
	}
	// Ordered by username
	for i := 1; i < len(users); i++ {
		if users[i].Username < users[i-1].Username {
			t.Error("users not ordered by username")
		}
	}
}

func TestGetUser(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	got, err := s.GetUser(alice.ID)
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if got.Username != "alice" {
		t.Errorf("expected username alice, got %s", got.Username)
	}
}

func TestGetUser_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetUser(999)
	if err == nil {
		t.Error("expected error for nonexistent user")
	}
}

// ── Projects ──

func TestCreateProject(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")

	p, err := s.CreateProject("Test Project", "", "A test project", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Name != "Test Project" {
		t.Errorf("expected name %q, got %q", "Test Project", p.Name)
	}
	if p.CreatedBy != alice.ID {
		t.Errorf("expected created_by %q, got %q", alice.ID, p.CreatedBy)
	}
	if p.ID == 0 {
		t.Error("expected non-empty id")
	}
	if p.CreatedAt == "" {
		t.Error("expected non-empty created_at")
	}
}

func TestCreateProject_autoSlug(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")

	p, err := s.CreateProject("Slug Test", "", "", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Slug != "SLUG-TEST" {
		t.Errorf("expected slug %q, got %q", "SLUG-TEST", p.Slug)
	}
}

func TestCreateProject_customSlug(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")

	p, err := s.CreateProject("Some Name", "CUSTOM-SLUG", "", alice.ID)
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Slug != "CUSTOM-SLUG" {
		t.Errorf("expected slug %q, got %q", "CUSTOM-SLUG", p.Slug)
	}
}

func TestGetProjectBySlug(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Lookup By Slug", "", "", alice.ID)

	got, err := s.GetProjectBySlug("LOOKUP-BY-SLUG")
	if err != nil {
		t.Fatalf("GetProjectBySlug: %v", err)
	}
	if got.ID != p.ID {
		t.Errorf("expected project id %q, got %q", p.ID, got.ID)
	}
}

func TestGetProjectBySlug_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetProjectBySlug("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent slug")
	}
}

func TestUpdateProject_slug(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Original", "ORIG", "", alice.ID)

	updated, err := s.UpdateProject(p.ID, "", "NEW-SLUG", "")
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Slug != "NEW-SLUG" {
		t.Errorf("expected slug %q, got %q", "NEW-SLUG", updated.Slug)
	}
	if updated.Name != "Original" {
		t.Errorf("name should remain %q, got %q", "Original", updated.Name)
	}

	// Verify persisted
	got, _ := s.GetProject(p.ID)
	if got.Slug != "NEW-SLUG" {
		t.Errorf("persisted slug %q, want %q", got.Slug, "NEW-SLUG")
	}
}

func TestCreateIssue_slugFromProjectSlug(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("My Cool Project", "MY-COOL", "", alice.ID)

	iss, err := s.CreateIssue(p.ID, "An issue", "", "", "", 0, 0, alice.ID, 0)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if iss.Slug != "MY-COOL-1" {
		t.Errorf("expected slug %q, got %q", "MY-COOL-1", iss.Slug)
	}
}

func TestListProjects_empty(t *testing.T) {
	s := newTestStore(t)
	projects, err := s.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

func TestListProjects_order(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")

	s.CreateProject("First", "", "", alice.ID)
	s.CreateProject("Second", "", "", alice.ID)

	projects, _ := s.ListProjects()
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestGetProject(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("My Project", "", "desc", alice.ID)

	got, err := s.GetProject(p.ID)
	if err != nil {
		t.Fatalf("GetProject: %v", err)
	}
	if got.Name != "My Project" {
		t.Errorf("expected name %q, got %q", "My Project", got.Name)
	}
}

func TestGetProject_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetProject(999)
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestUpdateProject(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Original", "", "orig desc", alice.ID)

	updated, err := s.UpdateProject(p.ID, "Updated", "", "new desc")
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != "Updated" {
		t.Errorf("expected name %q, got %q", "Updated", updated.Name)
	}
	if updated.Description != "new desc" {
		t.Errorf("expected desc %q, got %q", "new desc", updated.Description)
	}

	// Verify persisted
	got, _ := s.GetProject(p.ID)
	if got.Name != "Updated" {
		t.Errorf("persisted name %q, want %q", got.Name, "Updated")
	}
}

func TestUpdateProject_partial(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Original", "", "orig desc", alice.ID)

	// Only update description
	updated, err := s.UpdateProject(p.ID, "", "", "only desc changed")
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if updated.Name != "Original" {
		t.Errorf("expected name unchanged %q, got %q", "Original", updated.Name)
	}
	if updated.Description != "only desc changed" {
		t.Errorf("expected desc %q, got %q", "only desc changed", updated.Description)
	}
}

func TestUpdateProject_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.UpdateProject(999, "name", "", "desc")
	if err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestDeleteProject(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("To Delete", "", "", alice.ID)

	if err := s.DeleteProject(p.ID); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	_, err := s.GetProject(p.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ── Issues ──

func TestCreateIssue(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Asteroid Game", "", "", alice.ID)

	iss, err := s.CreateIssue(p.ID, "Add ship rotation", "Left/right arrows rotate the ship",
		"feature", "todo", alice.ID, 0, alice.ID, 2)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if iss.Title != "Add ship rotation" {
		t.Errorf("expected title %q, got %q", "Add ship rotation", iss.Title)
	}
	if iss.Slug != "ASTEROID-GAME-1" {
		t.Errorf("expected slug %q, got %q", "ASTEROID-GAME-1", iss.Slug)
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
}

func TestCreateIssue_defaults(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Test", "", "", alice.ID)

	// Only title is required
	iss, err := s.CreateIssue(p.ID, "My Issue", "", "", "", 0, 0, alice.ID, 0)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if iss.Type != "feature" {
		t.Errorf("expected default type feature, got %s", iss.Type)
	}
	if iss.State != "todo" {
		t.Errorf("expected default state todo, got %s", iss.State)
	}
	if iss.Priority != 0 {
		t.Errorf("expected default priority 0, got %d", iss.Priority)
	}
	if iss.Assignee != 0 {
		t.Errorf("expected empty assignee, got %d", iss.Assignee)
	}
}

func TestCreateIssue_epic(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Epic Project", "", "", alice.ID)

	iss, err := s.CreateIssue(p.ID, "Big Initiative", "", "epic", "backlog", 0, 0, alice.ID, 1)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if iss.Type != "epic" {
		t.Errorf("expected type epic, got %s", iss.Type)
	}
}

func TestCreateIssue_parent(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Hierarchy", "", "", alice.ID)

	parent, _ := s.CreateIssue(p.ID, "Parent Epic", "", "epic", "backlog", 0, 0, alice.ID, 0)
	child, err := s.CreateIssue(p.ID, "Child Task", "", "feature", "todo", alice.ID, parent.ID, alice.ID, 0)
	if err != nil {
		t.Fatalf("CreateIssue: %v", err)
	}
	if child.ParentID != parent.ID {
		t.Errorf("expected parent_id %q, got %q", parent.ID, child.ParentID)
	}
}

func TestListIssues_empty(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Empty", "", "", alice.ID)

	issues, err := s.ListIssues(p.ID, IssueFilter{})
	if err != nil {
		t.Fatalf("ListIssues: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}

func TestListIssues_filterByState(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Filter", "", "", alice.ID)

	s.CreateIssue(p.ID, "One", "", "", "todo", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Two", "", "", "qa", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Three", "", "", "todo", 0, 0, alice.ID, 0)

	issues, _ := s.ListIssues(p.ID, IssueFilter{State: "todo"})
	if len(issues) != 2 {
		t.Errorf("expected 2 todo issues, got %d", len(issues))
	}

	issues, _ = s.ListIssues(p.ID, IssueFilter{State: "qa"})
	if len(issues) != 1 {
		t.Errorf("expected 1 review issue, got %d", len(issues))
	}

	issues, _ = s.ListIssues(p.ID, IssueFilter{State: "done"})
	if len(issues) != 0 {
		t.Errorf("expected 0 done issues, got %d", len(issues))
	}
}

func TestListIssues_filterByAssignee(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	bob := getUserByPAT(t, s, "pat_bob")
	p, _ := s.CreateProject("Assign", "", "", alice.ID)

	s.CreateIssue(p.ID, "Alice's", "", "", "", alice.ID, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Bob's", "", "", "", bob.ID, 0, bob.ID, 0)

	issues, _ := s.ListIssues(p.ID, IssueFilter{Assignee: alice.ID})
	if len(issues) != 1 || issues[0].Title != "Alice's" {
		t.Errorf("expected 1 issue assigned to alice, got %d", len(issues))
	}

	issues, _ = s.ListIssues(p.ID, IssueFilter{Assignee: bob.ID})
	if len(issues) != 1 || issues[0].Title != "Bob's" {
		t.Errorf("expected 1 issue assigned to bob, got %d", len(issues))
	}
}

func TestListIssues_filterByType(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Types", "", "", alice.ID)

	s.CreateIssue(p.ID, "Bug report", "", "bug", "", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Feature request", "", "feature", "", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Chore task", "", "chore", "", 0, 0, alice.ID, 0)

	issues, _ := s.ListIssues(p.ID, IssueFilter{Type: "bug"})
	if len(issues) != 1 {
		t.Errorf("expected 1 bug, got %d", len(issues))
	}
	issues, _ = s.ListIssues(p.ID, IssueFilter{Type: "feature"})
	if len(issues) != 1 {
		t.Errorf("expected 1 feature, got %d", len(issues))
	}
	issues, _ = s.ListIssues(p.ID, IssueFilter{Type: "epic"})
	if len(issues) != 0 {
		t.Errorf("expected 0 epics, got %d", len(issues))
	}
}

func TestListIssues_searchQuery(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Search", "", "", alice.ID)

	s.CreateIssue(p.ID, "Login page", "User authentication flow", "", "", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Dashboard", "Main overview page", "", "", 0, 0, alice.ID, 0)
	s.CreateIssue(p.ID, "Logout", "", "", "", 0, 0, alice.ID, 0)

	// Match in title
	issues, _ := s.ListIssues(p.ID, IssueFilter{Query: "login"})
	if len(issues) != 1 {
		t.Errorf("expected 1 for 'login', got %d", len(issues))
	}

	// Match in description
	issues, _ = s.ListIssues(p.ID, IssueFilter{Query: "authentication"})
	if len(issues) != 1 {
		t.Errorf("expected 1 for 'authentication', got %d", len(issues))
	}

	// Matches "Login page" (title) and "Main overview page" (description)
	issues, _ = s.ListIssues(p.ID, IssueFilter{Query: "PAGE"})
	if len(issues) != 2 {
		t.Errorf("expected 2 for 'PAGE' (title + desc match), got %d", len(issues))
	}
}

func TestListIssues_pagination(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Pages", "", "", alice.ID)

	for i := range 7 {
		s.CreateIssue(p.ID, "", "Issue "+string(rune('A'+i)), "", "", 0, 0, alice.ID, 0)
	}
	_ = alice

	// Page 1 with 3 per page
	page1, _ := s.ListIssues(p.ID, IssueFilter{Page: 1, PerPage: 3})
	if len(page1) != 3 {
		t.Errorf("expected 3 on page 1, got %d", len(page1))
	}

	page2, _ := s.ListIssues(p.ID, IssueFilter{Page: 2, PerPage: 3})
	if len(page2) != 3 {
		t.Errorf("expected 3 on page 2, got %d", len(page2))
	}

	page3, _ := s.ListIssues(p.ID, IssueFilter{Page: 3, PerPage: 3})
	if len(page3) != 1 {
		t.Errorf("expected 1 on page 3, got %d", len(page3))
	}

	// Page beyond results
	page4, _ := s.ListIssues(p.ID, IssueFilter{Page: 4, PerPage: 3})
	if len(page4) != 0 {
		t.Errorf("expected 0 on page 4, got %d", len(page4))
	}
}

func TestListIssues_defaultPerPage(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Large", "", "", alice.ID)

	for range 60 {
		now()
		s.CreateIssue(p.ID, "Issue", "", "", "", 0, 0, alice.ID, 0)
	}

	issues, _ := s.ListIssues(p.ID, IssueFilter{})
	if len(issues) != 50 {
		t.Errorf("expected default per_page 50, got %d", len(issues))
	}
}

func TestGetIssue(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Get", "", "", alice.ID)
	created, _ := s.CreateIssue(p.ID, "My Issue", "desc", "bug", "qa", alice.ID, 0, alice.ID, 1)

	got, err := s.GetIssue(created.ID)
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if got.Title != "My Issue" {
		t.Errorf("expected title %q, got %q", "My Issue", got.Title)
	}
	if got.Type != "bug" {
		t.Errorf("expected type bug, got %s", got.Type)
	}
	if got.State != "qa" {
		t.Errorf("expected state review, got %s", got.State)
	}
	if got.Priority != 1 {
		t.Errorf("expected priority 1, got %d", got.Priority)
	}
}

func TestGetIssue_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetIssue(999)
	if err == nil {
		t.Error("expected error for nonexistent issue")
	}
}

func TestUpdateIssue_state(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Flow", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "Task", "", "", "todo", 0, 0, alice.ID, 0)

	updated, err := s.UpdateIssue(iss.ID, "", "", "", "done", 0, 0, 0)
	if err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}
	if updated.State != "done" {
		t.Errorf("expected state done, got %s", updated.State)
	}

	// Verify persisted
	got, _ := s.GetIssue(iss.ID)
	if got.State != "done" {
		t.Errorf("persisted state %s, want done", got.State)
	}
}

func TestUpdateIssue_multipleFields(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	bob := getUserByPAT(t, s, "pat_bob")
	p, _ := s.CreateProject("Update", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "Old title", "old desc", "bug", "backlog", alice.ID, 0, alice.ID, 4)

	updated, err := s.UpdateIssue(iss.ID, "New title", "new desc", "feature", "in_progress", bob.ID, 0, 2)
	if err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("title = %q, want %q", updated.Title, "New title")
	}
	if updated.Description != "new desc" {
		t.Errorf("description = %q, want %q", updated.Description, "new desc")
	}
	if updated.Type != "feature" {
		t.Errorf("type = %s, want feature", updated.Type)
	}
	if updated.State != "in_progress" {
		t.Errorf("state = %s, want in_progress", updated.State)
	}
	if updated.Assignee != bob.ID {
		t.Errorf("assignee = %d, want %d", updated.Assignee, bob.ID)
	}
	if updated.Priority != 2 {
		t.Errorf("priority = %d, want 2", updated.Priority)
	}
}

func TestUpdateIssue_partial(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Partial", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "Orig", "orig desc", "bug", "backlog", alice.ID, 0, alice.ID, 4)

	// Only change title and state
	updated, err := s.UpdateIssue(iss.ID, "New title", "", "", "qa", 0, 0, 0)
	if err != nil {
		t.Fatalf("UpdateIssue: %v", err)
	}
	if updated.Title != "New title" {
		t.Errorf("title = %q, want %q", updated.Title, "New title")
	}
	if updated.Description != "orig desc" {
		t.Errorf("description should remain 'orig desc', got %q", updated.Description)
	}
	if updated.State != "qa" {
		t.Errorf("state = %s, want review", updated.State)
	}
	if updated.Type != "bug" {
		t.Errorf("type should remain bug, got %s", updated.Type)
	}
	if updated.Priority != 4 {
		t.Errorf("priority should remain 4, got %d", updated.Priority)
	}
}

func TestUpdateIssue_notFound(t *testing.T) {
	s := newTestStore(t)
	_, err := s.UpdateIssue(999, "title", "", "", "", 0, 0, 0)
	if err == nil {
		t.Error("expected error for nonexistent issue")
	}
}

func TestDeleteIssue(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Delete", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "To Delete", "", "", "", 0, 0, alice.ID, 0)

	if err := s.DeleteIssue(iss.ID); err != nil {
		t.Fatalf("DeleteIssue: %v", err)
	}
	_, err := s.GetIssue(iss.ID)
	if err == nil {
		t.Error("expected error after delete")
	}
}

// ── Comments ──

func TestCreateComment(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Comments", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "The issue", "", "", "", 0, 0, alice.ID, 0)

	c, err := s.CreateComment(iss.ID, "This is a comment", alice.ID, alice.ID)
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if c.Body != "This is a comment" {
		t.Errorf("expected body %q, got %q", "This is a comment", c.Body)
	}
	if c.IssueID != iss.ID {
		t.Errorf("expected issue_id %q, got %q", iss.ID, c.IssueID)
	}
	if c.Author != alice.ID {
		t.Errorf("expected author %q, got %q", alice.ID, c.Author)
	}
}

func TestListComments_empty(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Empty", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "No comments", "", "", "", 0, 0, alice.ID, 0)

	comments, err := s.ListComments(iss.ID)
	if err != nil {
		t.Fatalf("ListComments: %v", err)
	}
	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestListComments_order(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Order", "", "", alice.ID)
	iss, _ := s.CreateIssue(p.ID, "Thread", "", "", "", 0, 0, alice.ID, 0)

	s.CreateComment(iss.ID, "First", alice.ID, alice.ID)
	s.CreateComment(iss.ID, "Second", alice.ID, alice.ID)
	s.CreateComment(iss.ID, "Third", alice.ID, alice.ID)

	comments, _ := s.ListComments(iss.ID)
	if len(comments) != 3 {
		t.Fatalf("expected 3 comments, got %d", len(comments))
	}
	if comments[0].Body != "First" || comments[2].Body != "Third" {
		t.Errorf("expected chronological order, got %q then %q", comments[0].Body, comments[2].Body)
	}
}

func TestListComments_otherIssue(t *testing.T) {
	s := newTestStore(t)
	alice := getUserByPAT(t, s, "pat_alice")
	p, _ := s.CreateProject("Separation", "", "", alice.ID)
	iss1, _ := s.CreateIssue(p.ID, "One", "", "", "", 0, 0, alice.ID, 0)
	iss2, _ := s.CreateIssue(p.ID, "Two", "", "", "", 0, 0, alice.ID, 0)

	s.CreateComment(iss1.ID, "On issue one", alice.ID, alice.ID)
	s.CreateComment(iss2.ID, "On issue two", alice.ID, alice.ID)

	c1, _ := s.ListComments(iss1.ID)
	if len(c1) != 1 || c1[0].Body != "On issue one" {
		t.Errorf("expected 1 comment on issue 1, got %d", len(c1))
	}
}

// ── Slugify ──

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "HELLO-WORLD"},
		{"Asteroid Game", "ASTEROID-GAME"},
		{"Test-Project", "TEST-PROJECT"},
		{"Multi   Spaces", "MULTI---SPACES"},
		{"Special!@#Chars", "SPECIALCHARS"},
		{"already_slug", "ALREADYSLUG"},
		{"", ""},
		{"123", "123"},
		{"hello-world", "HELLO-WORLD"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
