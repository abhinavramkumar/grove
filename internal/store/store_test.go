package store

import (
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("opening store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGetSession(t *testing.T) {
	s := newTestStore(t)

	wt := "/tmp/wt"
	prompt := "do stuff"
	sess, err := s.CreateSession("test-session", "claude", "/tmp/dir", &wt, &prompt, nil, nil)
	if err != nil {
		t.Fatalf("creating session: %v", err)
	}

	if sess.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if len(sess.ID) != 8 {
		t.Fatalf("expected 8-char ID, got %q", sess.ID)
	}
	if sess.Name != "test-session" {
		t.Fatalf("expected name 'test-session', got %q", sess.Name)
	}
	if sess.Tool != "claude" {
		t.Fatalf("expected tool 'claude', got %q", sess.Tool)
	}
	if sess.Status != "running" {
		t.Fatalf("expected status 'running', got %q", sess.Status)
	}
	if sess.TmuxSession != "grove-"+sess.ID {
		t.Fatalf("expected tmux session 'grove-%s', got %q", sess.ID, sess.TmuxSession)
	}

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatalf("getting session: %v", err)
	}
	if got.Name != sess.Name {
		t.Fatalf("name mismatch: %q vs %q", got.Name, sess.Name)
	}
}

func TestListSessions(t *testing.T) {
	s := newTestStore(t)

	_, err := s.CreateSession("s1", "claude", "/dir1", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("creating s1: %v", err)
	}
	_, err = s.CreateSession("s2", "codex", "/dir2", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("creating s2: %v", err)
	}

	sessions, err := s.ListSessions()
	if err != nil {
		t.Fatalf("listing sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
}

func TestUpdateSessionStatus(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.CreateSession("test", "claude", "/dir", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.UpdateSessionStatus(sess.ID, "stopped"); err != nil {
		t.Fatalf("updating status: %v", err)
	}

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "stopped" {
		t.Fatalf("expected status 'stopped', got %q", got.Status)
	}
	if got.StoppedAt == nil {
		t.Fatal("expected stopped_at to be set")
	}
}

func TestDeleteSession(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.CreateSession("test", "claude", "/dir", nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := s.DeleteSession(sess.ID); err != nil {
		t.Fatalf("deleting: %v", err)
	}

	_, err = s.GetSession(sess.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestDeleteNonexistentSession(t *testing.T) {
	s := newTestStore(t)
	err := s.DeleteSession("nonexist")
	if err == nil {
		t.Fatal("expected error for nonexistent session")
	}
}

func TestDefaultDBPath(t *testing.T) {
	// With XDG_DATA_HOME set
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	got := DefaultDBPath()
	if got != "/custom/data/grove/grove.db" {
		t.Fatalf("unexpected path: %s", got)
	}

	// Without XDG_DATA_HOME
	t.Setenv("XDG_DATA_HOME", "")
	got = DefaultDBPath()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".local", "share", "grove", "grove.db")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
