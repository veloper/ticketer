package internal

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("wal: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("fk: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func newID() string { return uuid.New().String() }

func slugify(name string) string {
	var out strings.Builder
	for _, r := range strings.ToUpper(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out.WriteRune(r)
		} else if r == ' ' || r == '-' {
			out.WriteRune('-')
		}
	}
	return out.String()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY, username TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL, pat TEXT UNIQUE NOT NULL,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS projects (
		id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT NOT NULL,
		created_by TEXT NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS issues (
		id TEXT PRIMARY KEY, project_id TEXT NOT NULL REFERENCES projects(id),
		slug TEXT NOT NULL, title TEXT NOT NULL, description TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'feature', state TEXT NOT NULL DEFAULT 'todo',
		assignee TEXT REFERENCES users(id),
		priority INTEGER NOT NULL DEFAULT 3, parent_id TEXT REFERENCES issues(id),
		created_by TEXT NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
		UNIQUE(project_id, slug)
	);
	CREATE TABLE IF NOT EXISTS comments (
		id TEXT PRIMARY KEY, issue_id TEXT NOT NULL REFERENCES issues(id),
		body TEXT NOT NULL, author TEXT NOT NULL REFERENCES users(id),
		created_by TEXT NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS sequences (
		project_id TEXT PRIMARY KEY, counter INTEGER NOT NULL DEFAULT 0
	);`)
	return err
}

// ── Seed ──

func (s *Store) SeedUsers(users []SeedUser) error {
	for _, u := range users {
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO users (id, username, display_name, pat, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			newID(), u.Username, u.DisplayName, u.PAT, now(), now(),
		)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.Username, err)
		}
	}
	return nil
}

// ── Users ──

func (s *Store) GetUserByPAT(pat string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, display_name, created_at, updated_at FROM users WHERE pat = ?`, pat,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *Store) ListUsers() ([]User, error) {
	rows, err := s.db.Query(`SELECT id, username, display_name, created_at, updated_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) GetUser(id string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, display_name, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// ── Projects ──

func (s *Store) CreateProject(name, description, createdBy string) (*Project, error) {
	p := &Project{
		ID: newID(), Name: name, Description: description,
		CreatedBy: createdBy, CreatedAt: now(), UpdatedAt: now(),
	}
	_, err := s.db.Exec(
		`INSERT INTO projects (id, name, description, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.Description, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	return p, err
}

func (s *Store) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(`SELECT id, name, description, created_by, created_at, updated_at FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetProject(id string) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(
		`SELECT id, name, description, created_by, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (s *Store) UpdateProject(id, name, description string) (*Project, error) {
	p, err := s.GetProject(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if description != "" {
		p.Description = description
	}
	p.UpdatedAt = now()
	_, err = s.db.Exec(`UPDATE projects SET name=?, description=?, updated_at=? WHERE id=?`,
		p.Name, p.Description, p.UpdatedAt, id)
	return p, err
}

func (s *Store) DeleteProject(id string) error {
	_, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

// ── Issues ──

func (s *Store) NextSlug(projectID string) (string, error) {
	p, err := s.GetProject(projectID)
	if err != nil {
		return "", err
	}
	prefix := slugify(p.Name)

	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO sequences (project_id, counter) VALUES (?, 1)
		ON CONFLICT(project_id) DO UPDATE SET counter = counter + 1`, projectID)
	if err != nil {
		return "", err
	}

	var counter int64
	err = tx.QueryRow(`SELECT counter FROM sequences WHERE project_id = ?`, projectID).Scan(&counter)
	if err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d", prefix, counter), nil
}

func (s *Store) CreateIssue(projectID, title, description, typ, state, assignee, parentID, createdBy string, priority int) (*Issue, error) {
	seq, err := s.NextSlug(projectID)
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
		ID: newID(), ProjectID: projectID, Slug: seq,
		Title: title, Description: description, Type: typ,
		State: state, Assignee: assignee, Priority: priority,
		ParentID: parentID, CreatedBy: createdBy,
		CreatedAt: now(), UpdatedAt: now(),
	}
	_, err = s.db.Exec(
		`INSERT INTO issues (id, project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		iss.ID, iss.ProjectID, iss.Slug, iss.Title, iss.Description,
		iss.Type, iss.State, iss.Assignee, iss.Priority, iss.ParentID,
		iss.CreatedBy, iss.CreatedAt, iss.UpdatedAt,
	)
	return iss, err
}

type IssueFilter struct {
	Type      string
	State     string
	Assignee  string
	CreatedBy string
	Query     string
	Page      int
	PerPage   int
}

func (s *Store) ListIssues(projectID string, f IssueFilter) ([]Issue, error) {
	where := []string{"project_id = ?"}
	args := []any{projectID}

	if f.State != "" {
		where = append(where, "state = ?")
		args = append(args, f.State)
	}
	if f.Assignee != "" {
		where = append(where, "assignee = ?")
		args = append(args, f.Assignee)
	}
	if f.CreatedBy != "" {
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
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
			&i.State, &i.Assignee, &i.Priority, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}

func (s *Store) GetIssue(id string) (*Issue, error) {
	i := &Issue{}
	err := s.db.QueryRow(
		`SELECT id, project_id, slug, title, description, type, state, assignee, priority, parent_id, created_by, created_at, updated_at
		 FROM issues WHERE id = ?`, id,
	).Scan(&i.ID, &i.ProjectID, &i.Slug, &i.Title, &i.Description,
		&i.State, &i.Assignee, &i.Priority, &i.CreatedBy, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (s *Store) UpdateIssue(id, title, description, typ, state, assignee, parentID string, priority int) (*Issue, error) {
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
	if state != "" {
		i.State = state
	}
	if assignee != "" {
		i.Assignee = assignee
	}
	if priority != 0 {
		i.Priority = priority
	}
	i.UpdatedAt = now()
	_, err = s.db.Exec(
		`UPDATE issues SET title=?, description=?, state=?, assignee=?, priority=?, updated_at=? WHERE id=?`,
		i.Title, i.Description, i.State, i.Assignee, i.Priority, i.UpdatedAt, id,
	)
	return i, err
}

func (s *Store) DeleteIssue(id string) error {
	_, err := s.db.Exec(`DELETE FROM issues WHERE id = ?`, id)
	return err
}

// ── Comments ──

func (s *Store) CreateComment(issueID, body, author, createdBy string) (*Comment, error) {
	c := &Comment{
		ID: newID(), IssueID: issueID, Body: body,
		Author: author, CreatedBy: createdBy,
		CreatedAt: now(), UpdatedAt: now(),
	}
	_, err := s.db.Exec(
		`INSERT INTO comments (id, issue_id, body, author, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.IssueID, c.Body, c.Author, c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	)
	return c, err
}

func (s *Store) ListComments(issueID string) ([]Comment, error) {
	rows, err := s.db.Query(
		`SELECT id, issue_id, body, author, created_by, created_at, updated_at
		 FROM comments WHERE issue_id = ? ORDER BY created_at ASC`, issueID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.Body, &c.Author, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
