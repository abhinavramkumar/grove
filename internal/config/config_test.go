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
