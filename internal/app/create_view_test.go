package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/config"
)

func TestCreateView_DirSourceStep(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	teatest.RequireEqualOutput(t, []byte(m.View()))
}

func TestCreateView_ProgressBar_Existing(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirExisting
	m.step = stepConfirm
	m.resolvedDir = "/tmp/test"

	view := m.View()
	for _, step := range []string{"Source", "Directory", "Tool", "Prompt", "Confirm"} {
		if !strings.Contains(view, step) {
			t.Errorf("expected progress step %q in view", step)
		}
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_ProgressBar_Worktree(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/myrepo", WorktreeBase: "/projects/myrepo-wt"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirWorktree
	m.step = stepConfirm
	m.selectedRepo = &repos[0]
	m.worktreeBranch = "feat/test"

	view := m.View()
	for _, step := range []string{"Source", "Repo", "Branch", "Tool", "Prompt", "Confirm"} {
		if !strings.Contains(view, step) {
			t.Errorf("expected progress step %q in view", step)
		}
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_RepoSelectStep(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/fermat"},
		{RepoRoot: "/projects/grove"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.step = stepRepoSelect
	m.dirSource = dirWorktree
	m.repoNames = []string{"fermat", "grove"}
	m.repoConfigs = repos

	view := m.View()
	if !strings.Contains(view, "fermat") || !strings.Contains(view, "grove") {
		t.Error("expected repo names in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_DirInputStep_Existing(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirExisting
	m.step = stepDirInput

	view := m.View()
	if !strings.Contains(view, "Enter directory path") {
		t.Error("expected 'Enter directory path' in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_DirInputStep_Worktree(t *testing.T) {
	repos := []config.RepoConfig{{RepoRoot: "/projects/myrepo"}}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirWorktree
	m.step = stepDirInput
	m.selectedRepo = &repos[0]

	view := m.View()
	if !strings.Contains(view, "Enter branch name") {
		t.Error("expected 'Enter branch name' in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_ToolSelectStep(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.step = stepTool

	view := m.View()
	if !strings.Contains(view, "claude") || !strings.Contains(view, "codex") {
		t.Error("expected tool names in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_PromptStep(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.step = stepPrompt

	view := m.View()
	if !strings.Contains(view, "Prompt or plan file") {
		t.Error("expected 'Prompt or plan file' in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_ConfirmStep_Existing(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirExisting
	m.step = stepConfirm
	m.resolvedDir = "/home/user/projects/myapp"

	view := m.View()
	if !strings.Contains(view, "Directory:") || !strings.Contains(view, "/home/user/projects/myapp") {
		t.Error("expected directory in confirm view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_ConfirmStep_Worktree(t *testing.T) {
	repos := []config.RepoConfig{
		{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
	}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.dirSource = dirWorktree
	m.step = stepConfirm
	m.selectedRepo = &repos[0]
	m.worktreeBranch = "feat/test"

	view := m.View()
	if !strings.Contains(view, "Repo:") || !strings.Contains(view, "Worktree:") {
		t.Error("expected Repo: and Worktree: in confirm view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_ErrorDisplay(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24
	m.err = "something went wrong"

	view := m.View()
	if !strings.Contains(view, "something went wrong") {
		t.Error("expected error message in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestCreateView_StatusBar(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "next") || !strings.Contains(view, "cancel") {
		t.Error("expected 'next' and 'cancel' in status bar")
	}
}
