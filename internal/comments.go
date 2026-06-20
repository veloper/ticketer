package internal

// CreateComment adds a comment to an issue.
func (s *Store) CreateComment(issueID int64, body string, author, createdBy int64) (*Comment, error) {
	c := &Comment{
		IssueID: issueID, Body: body,
		Author: author, CreatedBy: createdBy,
		CreatedAt: now(), UpdatedAt: now(),
	}
	res, err := s.db.Exec(
		`INSERT INTO comments (issue_id, body, author, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		c.IssueID, c.Body, c.Author, c.CreatedBy, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	c.ID, _ = res.LastInsertId()
	return c, nil
}

// ListComments returns all comments for an issue, oldest first.
func (s *Store) ListComments(issueID int64) ([]Comment, error) {
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
