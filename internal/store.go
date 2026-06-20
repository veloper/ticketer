package internal

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"

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

func generatePAT() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return fmt.Sprintf("pat_%x", b)
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// nullable returns nil for empty strings so SQL stores NULL instead of "".
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullableInt returns nil for zero values so SQL stores NULL instead of 0.
func nullableInt(n int64) any {
	if n == 0 {
		return nil
	}
	return n
}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL,
		display_name TEXT NOT NULL, pat TEXT UNIQUE NOT NULL,
		is_admin INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, slug TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL,
		created_by INTEGER NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY AUTOINCREMENT, project_id INTEGER NOT NULL REFERENCES projects(id),
		slug TEXT NOT NULL, title TEXT NOT NULL, description TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'feature', state TEXT NOT NULL DEFAULT 'todo',
		assignee INTEGER REFERENCES users(id),
		priority INTEGER NOT NULL DEFAULT 3, parent_id INTEGER REFERENCES issues(id),
		created_by INTEGER NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
		UNIQUE(project_id, slug)
	);
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT, issue_id INTEGER NOT NULL REFERENCES issues(id),
		body TEXT NOT NULL, author INTEGER NOT NULL REFERENCES users(id),
		created_by INTEGER NOT NULL REFERENCES users(id),
		created_at TEXT NOT NULL, updated_at TEXT NOT NULL
	);
`)
	return err
}

// EnsureAdmin creates or updates the admin user with the given username and PAT.
func (s *Store) EnsureAdmin(username, pat string) error {
	now := now()
	_, err := s.db.Exec(
		`INSERT INTO users (username, display_name, pat, is_admin, created_at, updated_at)
		 VALUES (?, ?, ?, 1, ?, ?)
		 ON CONFLICT(username) DO UPDATE SET pat = ?, is_admin = 1, updated_at = ?`,
		username, username, pat, now, now, pat, now,
	)
	return err
}

// SeedUsers inserts seed users, skipping any that already exist (by username).
func (s *Store) SeedUsers(users []SeedUser) error {
	for _, u := range users {
		isAdmin := 0
		if u.Admin {
			isAdmin = 1
		}
		_, err := s.db.Exec(
			`INSERT OR IGNORE INTO users (username, display_name, pat, is_admin, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			u.Username, u.DisplayName, u.PAT, isAdmin, now(), now(),
		)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", u.Username, err)
		}
	}
	return nil
}
