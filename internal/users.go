package internal

// GetUserByPAT looks up a user by their personal access token.
func (s *Store) GetUserByPAT(pat string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, display_name, is_admin, created_at, updated_at FROM users WHERE pat = ?`, pat,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// ListUsers returns all users ordered by username.
func (s *Store) ListUsers() ([]User, error) {
	rows, err := s.db.Query(`SELECT id, username, display_name, is_admin, created_at, updated_at FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetUser returns a single user by ID.
func (s *Store) GetUser(id int64) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, display_name, is_admin, created_at, updated_at FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// CreateUser creates a new user and returns it.
func (s *Store) CreateUser(username, displayName, pat string, isAdmin bool) (*User, error) {
	if displayName == "" {
		displayName = username
	}
	adminInt := 0
	if isAdmin {
		adminInt = 1
	}
	u := &User{
		Username: username, DisplayName: displayName,
		IsAdmin: isAdmin, PAT: pat, CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO users (username, display_name, pat, is_admin, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		u.Username, u.DisplayName, pat, adminInt, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	u.ID, _ = res.LastInsertId()
	return u, nil
}

// UpdateUser updates a user's display name and/or PAT.
func (s *Store) UpdateUser(id int64, displayName, pat string) (*User, error) {
	u, err := s.GetUser(id)
	if err != nil {
		return nil, err
	}
	if displayName != "" {
		u.DisplayName = displayName
	}
	if pat != "" {
		// Update the PAT in the DB
		_, err = s.db.Exec(`UPDATE users SET pat = ? WHERE id = ?`, pat, id)
		if err != nil {
			return nil, err
		}
	}
	u.UpdatedAt = now()
	_, err = s.db.Exec(`UPDATE users SET display_name = ?, updated_at = ? WHERE id = ?`,
		u.DisplayName, u.UpdatedAt, id)
	return u, err
}

// DeleteUser deletes a user by ID. Fails if the user has created projects or issues.
func (s *Store) DeleteUser(id int64) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

// GetUserByUsername returns a single user by username.
func (s *Store) GetUserByUsername(username string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow(
		`SELECT id, username, display_name, is_admin, created_at, updated_at FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.DisplayName, &u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
