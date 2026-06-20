package internal

import (
	"database/sql"
	"fmt"
	"strings"
)

// CreateIssue creates a new issue in a project. The slug is auto-generated
// from the project slug and the issue's auto-increment ID (e.g. "ASTEROID-GAME-42").
func (s *Store) CreateIssue(projectID int64, title, description, typ, state string, assignee int64, parentID int64, createdBy int64, priority int) (*Issue, error) {
	// Load project for slug prefix
	p, err := s.GetProject(projectID)
	if err != nil {
		return nil, err
	}
	if state == "" {
		state = "todo"
	}
	if typ == "" {
		typ = "feature"
	}
	iss := &Issue{
		ProjectID: projectID, Slug: "_", // placeholder — updated below
		Title: title, Description: description, Type: typ,
		State: state, Assignee: assignee, Priority: priority,
		ParentID: parentID, CreatedBy: createdBy,
		CreatedAt: now(), UpdatedAt: now(),
	}
	// Insert first to get the auto-increment ID, then set the real slug.
	res, err := s.db.Exec(
		`INSERT INTO issues (project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		iss.ProjectID, iss.Slug, iss.Title, iss.Description,
		iss.Type, iss.State, nullableInt(iss.Assignee), iss.Priority, nullableInt(iss.ParentID),
		iss.CreatedBy, iss.CreatedAt, iss.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	iss.ID, _ = res.LastInsertId()
	iss.Slug = fmt.Sprintf("%s-%d", p.Slug, iss.ID)
	// Update the slug with the real value
	_, err = s.db.Exec(`UPDATE issues SET slug = ? WHERE id = ?`, iss.Slug, iss.ID)
	if err != nil {
		return nil, err
	}
	return iss, nil
}

// IssueFilter specifies optional filters for listing issues.
type IssueFilter struct {
	Type      string
	State     string
	Assignee  int64
	CreatedBy int64
	Query     string
	Page      int
	PerPage   int
}

// ListIssues returns issues for a project, optionally filtered.
func (s *Store) ListIssues(projectID int64, f IssueFilter) ([]Issue, error) {
	where := []string{"project_id = ?"}
	args := []any{projectID}

	if f.Type != "" {
		where = append(where, "type = ?")
		args = append(args, f.Type)
	}
	if f.State != "" {
		where = append(where, "state = ?")
		args = append(args, f.State)
	}
	if f.Assignee != 0 {
		where = append(where, "assignee = ?")
		args = append(args, f.Assignee)
	}
	if f.CreatedBy != 0 {
		where = append(where, "created_by = ?")
		args = append(args, f.CreatedBy)
	}
	if f.Query != "" {
		where = append(where, "(title LIKE ? OR description LIKE ?)")
		q := "%" + f.Query + "%"
		args = append(args, q, q)
	}

	if f.PerPage <= 0 {
		f.PerPage = 50
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.PerPage

	q := fmt.Sprintf(
		`SELECT id, project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at
		 FROM issues WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		strings.Join(where, " AND "),
	)
	args = append(args, f.PerPage, offset)

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Issue
	for rows.Next() {
		var i Issue
		var assignee, parentID sql.NullInt64
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
			&i.Type, &i.State, &assignee, &i.Priority, &parentID,
			&i.CreatedBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		i.Assignee = assignee.Int64
		i.ParentID = parentID.Int64
		out = append(out, i)
	}
	return out, rows.Err()
}

// GetIssue returns a single issue by ID.
func (s *Store) GetIssue(id int64) (*Issue, error) {
	i := &Issue{}
	var assignee, parentID sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at
		 FROM issues WHERE id = ?`, id,
	).Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
		&i.Type, &i.State, &assignee, &i.Priority, &parentID,
		&i.CreatedBy, &i.CreatedAt, &i.UpdatedAt)
	i.Assignee = assignee.Int64
	i.ParentID = parentID.Int64
	return i, err
}

// GetIssueBySlug returns a single issue by its slug.
func (s *Store) GetIssueBySlug(slug string) (*Issue, error) {
	i := &Issue{}
	var assignee, parentID sql.NullInt64
	err := s.db.QueryRow(
		`SELECT id, project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at
		 FROM issues WHERE slug = ?`, slug,
	).Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
		&i.Type, &i.State, &assignee, &i.Priority, &parentID,
		&i.CreatedBy, &i.CreatedAt, &i.UpdatedAt)
	i.Assignee = assignee.Int64
	i.ParentID = parentID.Int64
	return i, err
}

// UpdateIssue updates fields on an issue. Empty strings/0 values are left unchanged.
func (s *Store) UpdateIssue(id int64, title, description, typ, state string, assignee int64, parentID int64, priority int) (*Issue, error) {
	i, err := s.GetIssue(id)
	if err != nil {
		return nil, err
	}
	if title != "" {
		i.Title = title
	}
	if description != "" {
		i.Description = description
	}
	if typ != "" {
		i.Type = typ
	}
	if state != "" {
		i.State = state
	}
	if assignee != 0 {
		i.Assignee = assignee
	}
	if priority != 0 {
		i.Priority = priority
	}
	if parentID != 0 {
		i.ParentID = parentID
	}
	i.UpdatedAt = now()
	_, err = s.db.Exec(
		`UPDATE issues SET title=?, description=?, type=?, state=?, assignee=?, priority=?, parent_id=?, updated_at=? WHERE id=?`,
		i.Title, i.Description, i.Type, i.State, nullableInt(i.Assignee), i.Priority, nullableInt(i.ParentID), i.UpdatedAt, id,
	)
	return i, err
}

// DeleteIssue deletes an issue by ID.
func (s *Store) DeleteIssue(id int64) error {
	_, err := s.db.Exec(`DELETE FROM issues WHERE id = ?`, id)
	return err
}
