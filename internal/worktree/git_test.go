package worktree

import (
	"testing"
)

func TestSanitizeBranchName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feature/foo", "feature-foo"},
		{"simple", "simple"},
		{"a/b/c", "a-b-c"},
		{"hello world", "hello-world"},
		{"--leading", "leading"},
		{"trailing..", "trailing"},
		{"ok-name", "ok-name"},
		{"", "branch"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeBranchName(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeBranchName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseWorktreeList(t *testing.T) {
	input := `worktree /Users/test/repo
HEAD abc123def456
branch refs/heads/main

worktree /Users/test/repo-feature
HEAD def456abc789
branch refs/heads/feature

`
	result := parseWorktreeList(input)
	if len(result) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(result))
	}

	if result[0].Path != "/Users/test/repo" {
		t.Fatalf("unexpected path: %s", result[0].Path)
	}
	if result[0].Branch != "refs/heads/main" {
		t.Fatalf("unexpected branch: %s", result[0].Branch)
	}
	if result[1].Path != "/Users/test/repo-feature" {
		t.Fatalf("unexpected path: %s", result[1].Path)
	}
}

func TestParseWorktreeListBare(t *testing.T) {
	input := `worktree /Users/test/repo
HEAD abc123
bare

`
	result := parseWorktreeList(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(result))
	}
	if !result[0].Bare {
		t.Fatal("expected bare worktree")
	}
}
