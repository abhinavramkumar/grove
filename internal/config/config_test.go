package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDirXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	got := ConfigDir()
	if got != "/custom/config/grove" {
		t.Fatalf("expected /custom/config/grove, got %s", got)
	}
}

func TestConfigDirDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	got := ConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "grove")
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/Projects/Work",
		},
		Worktree: WorktreeConfig{
			SetupCommands: []string{"npm install"},
		},
		Tools: map[string]ToolConfig{
			"claude": {Command: "claude", Args: []string{"-p"}},
			"codex":  {Command: "codex"},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	if !ConfigExists() {
		t.Fatal("config should exist after save")
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if loaded.Defaults.AITool != "claude" {
		t.Fatalf("expected ai_tool 'claude', got %q", loaded.Defaults.AITool)
	}
	if loaded.Defaults.WorktreeBase != "~/Projects/Work" {
		t.Fatalf("expected worktree_base '~/Projects/Work', got %q", loaded.Defaults.WorktreeBase)
	}
	if len(loaded.Worktree.SetupCommands) != 1 || loaded.Worktree.SetupCommands[0] != "npm install" {
		t.Fatalf("unexpected setup_commands: %v", loaded.Worktree.SetupCommands)
	}
	if loaded.Tools["claude"].Command != "claude" {
		t.Fatal("missing claude tool")
	}
	if len(loaded.Tools["claude"].Args) != 1 || loaded.Tools["claude"].Args[0] != "-p" {
		t.Fatalf("unexpected claude args: %v", loaded.Tools["claude"].Args)
	}
}

func TestConfigExistsReturnsFalse(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if ConfigExists() {
		t.Fatal("config should not exist")
	}
}

func TestRunWizardDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// All empty input → defaults
	input := strings.NewReader("\n\n\n")
	var output strings.Builder

	cfg, err := RunWizard(input, &output)
	if err != nil {
		t.Fatalf("wizard failed: %v", err)
	}

	if cfg.Defaults.AITool != "claude" {
		t.Fatalf("expected default ai_tool 'claude', got %q", cfg.Defaults.AITool)
	}
	if cfg.Defaults.WorktreeBase != "~/Projects/Work" {
		t.Fatalf("expected default worktree_base, got %q", cfg.Defaults.WorktreeBase)
	}
	if len(cfg.Worktree.SetupCommands) != 0 {
		t.Fatalf("expected no setup commands, got %v", cfg.Worktree.SetupCommands)
	}

	// Verify file was written
	if !ConfigExists() {
		t.Fatal("config should exist after wizard")
	}
}

func TestRunWizardCustomValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	input := strings.NewReader("codex\n/my/projects\nnpm install, make build\n")
	var output strings.Builder

	cfg, err := RunWizard(input, &output)
	if err != nil {
		t.Fatalf("wizard failed: %v", err)
	}

	if cfg.Defaults.AITool != "codex" {
		t.Fatalf("expected ai_tool 'codex', got %q", cfg.Defaults.AITool)
	}
	if cfg.Defaults.WorktreeBase != "/my/projects" {
		t.Fatalf("expected worktree_base '/my/projects', got %q", cfg.Defaults.WorktreeBase)
	}
	if len(cfg.Worktree.SetupCommands) != 2 {
		t.Fatalf("expected 2 setup commands, got %v", cfg.Worktree.SetupCommands)
	}
	if cfg.Worktree.SetupCommands[0] != "npm install" {
		t.Fatalf("unexpected first command: %q", cfg.Worktree.SetupCommands[0])
	}
	if cfg.Worktree.SetupCommands[1] != "make build" {
		t.Fatalf("unexpected second command: %q", cfg.Worktree.SetupCommands[1])
	}
}

func TestSaveAndLoadWithRepos(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/Projects/Work",
		},
		Repos: []RepoConfig{
			{
				RepoRoot:      "/home/user/myrepo",
				WorktreeBase:  "/tmp/worktrees",
				AITool:        "codex",
				SetupCommands: []string{"make build"},
			},
			{
				RepoRoot: "/home/user/other",
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if len(loaded.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loaded.Repos))
	}

	r0 := loaded.Repos[0]
	if r0.RepoRoot != "/home/user/myrepo" {
		t.Fatalf("expected repo_root '/home/user/myrepo', got %q", r0.RepoRoot)
	}
	if r0.WorktreeBase != "/tmp/worktrees" {
		t.Fatalf("expected worktree_base '/tmp/worktrees', got %q", r0.WorktreeBase)
	}
	if r0.AITool != "codex" {
		t.Fatalf("expected ai_tool 'codex', got %q", r0.AITool)
	}
	if len(r0.SetupCommands) != 1 || r0.SetupCommands[0] != "make build" {
		t.Fatalf("unexpected setup_commands: %v", r0.SetupCommands)
	}

	r1 := loaded.Repos[1]
	if r1.RepoRoot != "/home/user/other" {
		t.Fatalf("expected repo_root '/home/user/other', got %q", r1.RepoRoot)
	}
	if r1.AITool != "" {
		t.Fatalf("expected empty ai_tool, got %q", r1.AITool)
	}
	if r1.WorktreeBase != "" {
		t.Fatalf("expected empty worktree_base, got %q", r1.WorktreeBase)
	}
}

func TestRepoFor(t *testing.T) {
	cfg := &Config{
		Repos: []RepoConfig{
			{RepoRoot: "/home/user/myrepo"},
			{RepoRoot: "/home/user/other"},
		},
	}

	// Found
	r := cfg.RepoFor("/home/user/myrepo")
	if r == nil {
		t.Fatal("expected to find repo")
	}
	if r.RepoRoot != "/home/user/myrepo" {
		t.Fatalf("expected '/home/user/myrepo', got %q", r.RepoRoot)
	}

	// Not found
	r = cfg.RepoFor("/nonexistent")
	if r != nil {
		t.Fatal("expected nil for missing repo")
	}
}

func TestEffectiveSettings(t *testing.T) {
	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/Projects/Work",
		},
		Worktree: WorktreeConfig{
			SetupCommands: []string{"npm install"},
		},
	}

	// nil repo -> global defaults
	if got := cfg.EffectiveAITool(nil); got != "claude" {
		t.Fatalf("expected 'claude', got %q", got)
	}
	if got := cfg.EffectiveWorktreeBase(nil); got != "~/Projects/Work" {
		t.Fatalf("expected '~/Projects/Work', got %q", got)
	}
	if got := cfg.EffectiveSetupCommands(nil); len(got) != 1 || got[0] != "npm install" {
		t.Fatalf("expected ['npm install'], got %v", got)
	}

	// Repo with overrides
	repo := &RepoConfig{
		RepoRoot:      "/home/user/myrepo",
		AITool:        "codex",
		WorktreeBase:  "/custom/base",
		SetupCommands: []string{"make build", "make test"},
	}
	if got := cfg.EffectiveAITool(repo); got != "codex" {
		t.Fatalf("expected 'codex', got %q", got)
	}
	if got := cfg.EffectiveWorktreeBase(repo); got != "/custom/base" {
		t.Fatalf("expected '/custom/base', got %q", got)
	}
	if got := cfg.EffectiveSetupCommands(repo); len(got) != 2 || got[0] != "make build" {
		t.Fatalf("expected ['make build', 'make test'], got %v", got)
	}

	// Repo with empty overrides -> falls back to global
	emptyRepo := &RepoConfig{RepoRoot: "/home/user/other"}
	if got := cfg.EffectiveAITool(emptyRepo); got != "claude" {
		t.Fatalf("expected 'claude', got %q", got)
	}
	if got := cfg.EffectiveWorktreeBase(emptyRepo); got != "~/Projects/Work" {
		t.Fatalf("expected '~/Projects/Work', got %q", got)
	}
	if got := cfg.EffectiveSetupCommands(emptyRepo); len(got) != 1 || got[0] != "npm install" {
		t.Fatalf("expected ['npm install'], got %v", got)
	}
}

func TestAddRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/Projects/Work",
		},
	}
	if err := Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	// Add new repo
	err := cfg.AddRepo(RepoConfig{RepoRoot: "/home/user/myrepo", AITool: "codex"})
	if err != nil {
		t.Fatalf("AddRepo failed: %v", err)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}

	// Verify persisted
	loaded, err := Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}
	if len(loaded.Repos) != 1 || loaded.Repos[0].RepoRoot != "/home/user/myrepo" {
		t.Fatalf("repo not persisted correctly: %v", loaded.Repos)
	}

	// Duplicate error
	err = cfg.AddRepo(RepoConfig{RepoRoot: "/home/user/myrepo"})
	if err == nil {
		t.Fatal("expected error for duplicate repo")
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected still 1 repo after duplicate, got %d", len(cfg.Repos))
	}
}
