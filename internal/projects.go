package internal

import "fmt"

// CreateProject creates a new project for the given user.
func (s *Store) CreateProject(name, slug, description string, createdBy int64) (*Project, error) {
	if slug == "" {
		slug = slugify(name)
	}
	p := &Project{
		Name: name, Slug: slug, Description: description,
		CreatedBy: createdBy, CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO projects (name, slug, description, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		p.Name, p.Slug, p.Description, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}
	p.ID, _ = res.LastInsertId()
	return p, nil
}

// ListProjects returns all projects, newest first.
func (s *Store) ListProjects() ([]Project, error) {
	rows, err := s.db.Query(`SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetProject returns a single project by ID.
func (s *Store) GetProject(id int64) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// GetProjectBySlug returns a single project by slug.
func (s *Store) GetProjectBySlug(slug string) (*Project, error) {
	p := &Project{}
	err := s.db.QueryRow(
		`SELECT id, name, slug, description, created_by, created_at, updated_at FROM projects WHERE slug = ?`, slug,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

// UpdateProject updates a project's name, slug, and/or description.
func (s *Store) UpdateProject(id int64, name, slug, description string) (*Project, error) {
	p, err := s.GetProject(id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		p.Name = name
	}
	if slug != "" {
		p.Slug = slug
	}
	if description != "" {
		p.Description = description
	}
	p.UpdatedAt = now()
	_, err = s.db.Exec(`UPDATE projects SET name=?, slug=?, description=?, updated_at=? WHERE id=?`,
		p.Name, p.Slug, p.Description, p.UpdatedAt, id)
	return p, err
}

// DeleteProject deletes a project by ID.
func (s *Store) DeleteProject(id int64) error {
	_, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}
