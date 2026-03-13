package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Session represents a Grove AI coding session.
type Session struct {
	ID            string
	Name          string
	Tool          string
	Worktree      *string
	Directory     string
	RepoRoot      *string
	Prompt        *string
	PlanFile      *string
	TmuxSession   string
	ToolSessionID *string // native session ID from the AI tool
	Status        string
	CreatedAt     time.Time
	StoppedAt     *time.Time
}

// Store wraps a SQLite database for session persistence.
type Store struct {
	db *sql.DB
}

// DefaultDBPath returns the default database path using XDG conventions.
func DefaultDBPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "grove", "grove.db")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "grove", "grove.db")
}

// Open opens (or creates) the SQLite database and runs migrations.
func Open(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// generateID returns an 8-character hex string from crypto/rand.
func generateID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// CreateSession inserts a new session and returns it with its generated ID.
func (s *Store) CreateSession(name, tool, directory string, worktree, prompt, planFile, toolSessionID, repoRoot *string) (*Session, error) {
	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("generating id: %w", err)
	}

	tmuxSession := "grove-" + id

	_, err = s.db.Exec(`
		INSERT INTO sessions (id, name, tool, worktree, directory, repo_root, prompt, plan_file, tmux_session, tool_session_id, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'running')
	`, id, name, tool, worktree, directory, repoRoot, prompt, planFile, tmuxSession, toolSessionID)
	if err != nil {
		return nil, fmt.Errorf("inserting session: %w", err)
	}

	return s.GetSession(id)
}

const sessionColumns = `id, name, tool, worktree, directory, repo_root, prompt, plan_file, tmux_session, tool_session_id, status, created_at, stopped_at`

// GetSession retrieves a session by ID.
func (s *Store) GetSession(id string) (*Session, error) {
	row := s.db.QueryRow(`SELECT `+sessionColumns+` FROM sessions WHERE id = ?`, id)
	return scanFrom(row)
}

// ListSessions returns all sessions, most recent first.
func (s *Store) ListSessions() ([]*Session, error) {
	rows, err := s.db.Query(`SELECT ` + sessionColumns + ` FROM sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("querying sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		sess, err := scanFrom(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

// UpdateSessionStatus sets the status and optionally the stopped_at time.
func (s *Store) UpdateSessionStatus(id, status string) error {
	var err error
	if status == "stopped" || status == "finished" {
		_, err = s.db.Exec(`
			UPDATE sessions SET status = ?, stopped_at = CURRENT_TIMESTAMP WHERE id = ?
		`, status, id)
	} else {
		_, err = s.db.Exec(`
			UPDATE sessions SET status = ?, stopped_at = NULL WHERE id = ?
		`, status, id)
	}
	if err != nil {
		return fmt.Errorf("updating session status: %w", err)
	}
	return nil
}

// UpdateToolSessionID sets the native tool session ID for a session.
func (s *Store) UpdateToolSessionID(id, toolSessionID string) error {
	_, err := s.db.Exec(`UPDATE sessions SET tool_session_id = ? WHERE id = ?`, toolSessionID, id)
	if err != nil {
		return fmt.Errorf("updating tool session id: %w", err)
	}
	return nil
}

// DeleteSession removes a session by ID.
func (s *Store) DeleteSession(id string) error {
	result, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session %s not found", id)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFrom(s scanner) (*Session, error) {
	var sess Session
	var createdAt string
	var stoppedAt *string
	err := s.Scan(
		&sess.ID, &sess.Name, &sess.Tool, &sess.Worktree,
		&sess.Directory, &sess.RepoRoot, &sess.Prompt, &sess.PlanFile,
		&sess.TmuxSession, &sess.ToolSessionID, &sess.Status,
		&createdAt, &stoppedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning session: %w", err)
	}
	sess.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	if stoppedAt != nil {
		t, _ := time.Parse("2006-01-02 15:04:05", *stoppedAt)
		sess.StoppedAt = &t
	}
	return &sess, nil
}
