package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/tools"
)

// stubAdapter satisfies tools.Adapter for testing without real tool binaries.
type stubAdapter struct{ name string }

func (s stubAdapter) Name() string                              { return s.name }
func (s stubAdapter) NewSessionCmd(_, _, _ string) (string, string) { return "echo test", "" }
func (s stubAdapter) ResumeSessionCmd(_, _ string) string        { return "echo resume" }
func (s stubAdapter) SupportsResume() bool                       { return false }

func newTestCreateDeps(t *testing.T, repos []config.RepoConfig) (*config.Config, *session.Manager) {
	t.Helper()
	dir := t.TempDir()

	dbPath := dir + "/test.db"
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude", WorktreeBase: "~/wt"},
		Tools: map[string]config.ToolConfig{
			"claude": {Command: "claude"},
			"codex":  {Command: "codex"},
		},
		Repos: repos,
	}

	mgr := &session.Manager{
		Store: s,
		Adapters: map[string]tools.Adapter{
			"claude": stubAdapter{name: "claude"},
			"codex":  stubAdapter{name: "codex"},
		},
	}

	return cfg, mgr
}

func TestCreateModel_WorktreeShowsRepoSelector(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
		{RepoRoot: "/projects/grove", WorktreeBase: "/projects/grove-wt", AITool: "codex"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)

	// Press "2" for worktree mode.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	if m.step != stepRepoSelect {
		t.Fatalf("expected step %d (stepRepoSelect), got %d", stepRepoSelect, m.step)
	}

	view := m.View()
	if !strings.Contains(view, "fermat") || !strings.Contains(view, "grove") {
		t.Fatalf("expected repo names in view, got:\n%s", view)
	}
}

func TestCreateModel_WorktreeNoReposShowsError(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil) // no repos

	m := NewCreateModel(cfg, mgr)

	// Press "2" for worktree mode.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	// Should stay on dirSource step and show error.
	if m.step != stepDirSource {
		t.Fatalf("expected to stay on stepDirSource, got step %d", m.step)
	}
	if m.err == "" || !strings.Contains(m.err, "repo") {
		t.Fatalf("expected error about no repos, got err=%q", m.err)
	}
}

func TestCreateModel_RepoSelectNavigation(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/alpha"},
		{RepoRoot: "/projects/beta"},
		{RepoRoot: "/projects/gamma"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)

	// Enter worktree mode.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.repoSelected != 0 {
		t.Fatalf("expected initial selection 0, got %d", m.repoSelected)
	}

	// Navigate right twice.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.repoSelected != 1 {
		t.Fatalf("expected selection 1 after right, got %d", m.repoSelected)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.repoSelected != 2 {
		t.Fatalf("expected selection 2 after second right, got %d", m.repoSelected)
	}

	// Right at end stays at end.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.repoSelected != 2 {
		t.Fatalf("expected selection to stay at 2, got %d", m.repoSelected)
	}

	// Navigate left.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.repoSelected != 1 {
		t.Fatalf("expected selection 1 after left, got %d", m.repoSelected)
	}
}

func TestCreateModel_RepoSelectPreSelectsAITool(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
		{RepoRoot: "/projects/grove", WorktreeBase: "/projects/grove-wt", AITool: "codex"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)

	// Press "2" for worktree -> goes to repo select.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	// Navigate right to select "grove" repo (which has ai_tool=codex).
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})

	// Confirm repo selection.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepDirInput {
		t.Fatalf("expected step %d (stepDirInput), got %d", stepDirInput, m.step)
	}
	if m.selectedRepo == nil || m.selectedRepo.RepoRoot != "/projects/grove" {
		t.Fatalf("expected selectedRepo to be /projects/grove, got %+v", m.selectedRepo)
	}
	if m.toolNames[m.toolSelected] != "codex" {
		t.Fatalf("expected tool 'codex' pre-selected, got %q", m.toolNames[m.toolSelected])
	}
}

func TestCreateModel_RepoSelectConfirmMovesBranchInput(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/myrepo"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)

	// Worktree mode -> repo select -> confirm.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.step != stepDirInput {
		t.Fatalf("expected stepDirInput after repo confirm, got step %d", m.step)
	}
	if m.selectedRepo == nil {
		t.Fatal("expected selectedRepo to be set")
	}
}

func TestCreateModel_ConfirmViewShowsRepo(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)

	// Walk through: worktree -> repo select -> confirm repo -> branch input.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm repo

	// Manually set worktree branch and advance to confirm step for view test.
	m.worktreeBranch = "feat/test"
	m.step = stepConfirm

	view := m.View()
	if !strings.Contains(view, "/projects/fermat") {
		t.Fatalf("expected confirm view to show repo root, got:\n%s", view)
	}
	if !strings.Contains(view, "feat/test") {
		t.Fatalf("expected confirm view to show branch, got:\n%s", view)
	}
}
