package app

import (
	"testing"
	"time"

	"github.com/abhinav/grove/internal/store"
)

func strPtr(s string) *string { return &s }

func makeSessions() []*store.Session {
	return []*store.Session{
		{
			ID: "1", Name: "feat-auth", Tool: "claude",
			Directory: "/home/user/repos/myapp/worktrees/feat-auth",
			RepoRoot:  strPtr("/home/user/repos/myapp"),
			Status: "running", CreatedAt: time.Now(),
		},
		{
			ID: "2", Name: "fix-bug", Tool: "copilot",
			Directory: "/home/user/repos/backend/worktrees/fix-bug",
			RepoRoot:  strPtr("/home/user/repos/backend"),
			Status: "stopped", CreatedAt: time.Now(),
		},
		{
			ID: "3", Name: "refactor", Tool: "claude",
			Directory: "/home/user/repos/frontend/worktrees/refactor",
			RepoRoot:  nil,
			Status: "finished", CreatedAt: time.Now(),
		},
	}
}

func TestFilterSessions_NoFilter(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "")
	if len(got) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(got))
	}
}

func TestFilterSessions_ByRepoName(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "myapp")
	if len(got) != 1 {
		t.Fatalf("expected 1 session matching 'myapp', got %d", len(got))
	}
	if got[0].ID != "1" {
		t.Fatalf("expected session '1', got '%s'", got[0].ID)
	}
}

func TestFilterSessions_ByToolName(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "copilot")
	if len(got) != 1 {
		t.Fatalf("expected 1 session, got %d", len(got))
	}
	if got[0].ID != "2" {
		t.Fatalf("expected session '2', got '%s'", got[0].ID)
	}
}

func TestFilterSessions_BySessionName(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "refactor")
	if len(got) != 1 {
		t.Fatalf("expected 1 session, got %d", len(got))
	}
	if got[0].ID != "3" {
		t.Fatalf("expected session '3', got '%s'", got[0].ID)
	}
}

func TestFilterSessions_CaseInsensitive(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "CLAUDE")
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions matching 'CLAUDE', got %d", len(got))
	}
}

func TestFilterSessions_EmptyList(t *testing.T) {
	got := filterSessions(nil, "anything")
	if len(got) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(got))
	}
}

func TestFilterSessions_NilRepoRoot(t *testing.T) {
	sessions := []*store.Session{
		{
			ID: "1", Name: "test", Tool: "claude",
			Directory: "/tmp/test",
			RepoRoot:  nil,
			Status: "running", CreatedAt: time.Now(),
		},
	}
	// Should not panic.
	got := filterSessions(sessions, "nonexistent")
	if len(got) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(got))
	}
}

func TestFilterSessions_NoMatches(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "zzzznotfound")
	if len(got) != 0 {
		t.Fatalf("expected 0 sessions, got %d", len(got))
	}
}

func TestFilterSessions_MatchesDirectory(t *testing.T) {
	sessions := makeSessions()
	got := filterSessions(sessions, "backend")
	if len(got) != 1 {
		t.Fatalf("expected 1 session matching directory 'backend', got %d", len(got))
	}
	if got[0].ID != "2" {
		t.Fatalf("expected session '2', got '%s'", got[0].ID)
	}
}

func TestRepoDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		repoRoot *string
		want     string
	}{
		{"with repo", strPtr("/home/user/repos/myapp"), "myapp"},
		{"nil repo", nil, "—"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sess := &store.Session{RepoRoot: tt.repoRoot}
			got := repoDisplayName(sess)
			if got != tt.want {
				t.Fatalf("repoDisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
