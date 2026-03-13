package config

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents the Grove configuration file.
type Config struct {
	Defaults DefaultsConfig        `toml:"defaults"`
	Worktree WorktreeConfig        `toml:"worktree"`
	Tools    map[string]ToolConfig `toml:"tools"`
	Repos    []RepoConfig          `toml:"repos,omitempty"`
}

// RepoConfig holds per-repository overrides.
type RepoConfig struct {
	RepoRoot      string   `toml:"repo_root"`
	WorktreeBase  string   `toml:"worktree_base,omitempty"`
	AITool        string   `toml:"ai_tool,omitempty"`
	SetupCommands []string `toml:"setup_commands,omitempty"`
}

type DefaultsConfig struct {
	AITool       string `toml:"ai_tool"`
	WorktreeBase string `toml:"worktree_base"`
	Theme        string `toml:"theme,omitempty"`
}

type WorktreeConfig struct {
	SetupCommands []string `toml:"setup_commands"`
}

type ToolConfig struct {
	Command string   `toml:"command"`
	Args    []string `toml:"args,omitempty"`
}

// ConfigDir returns the directory where the config file is stored.
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "grove")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "grove")
}

// ConfigPath returns the full path to the config file.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

// ConfigExists returns true if the config file exists.
func ConfigExists() bool {
	_, err := os.Stat(ConfigPath())
	return err == nil
}

// Load reads and parses the config file.
func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0o644)
}

// RepoFor returns the RepoConfig matching the given repoRoot, or nil if not found.
func (c *Config) RepoFor(repoRoot string) *RepoConfig {
	for i := range c.Repos {
		if c.Repos[i].RepoRoot == repoRoot {
			return &c.Repos[i]
		}
	}
	return nil
}

// EffectiveAITool returns the repo-level AI tool override, falling back to the global default.
func (c *Config) EffectiveAITool(repo *RepoConfig) string {
	if repo != nil && repo.AITool != "" {
		return repo.AITool
	}
	return c.Defaults.AITool
}

// EffectiveWorktreeBase returns the repo-level worktree base override, falling back to the global default.
func (c *Config) EffectiveWorktreeBase(repo *RepoConfig) string {
	if repo != nil && repo.WorktreeBase != "" {
		return repo.WorktreeBase
	}
	return c.Defaults.WorktreeBase
}

// EffectiveSetupCommands returns the repo-level setup commands override, falling back to the global default.
func (c *Config) EffectiveSetupCommands(repo *RepoConfig) []string {
	if repo != nil && len(repo.SetupCommands) > 0 {
		return repo.SetupCommands
	}
	return c.Worktree.SetupCommands
}

// AddRepo appends a new repo to the config and saves. Returns an error if the repo already exists.
func (c *Config) AddRepo(repo RepoConfig) error {
	if c.RepoFor(repo.RepoRoot) != nil {
		return fmt.Errorf("repo %q already registered", repo.RepoRoot)
	}
	c.Repos = append(c.Repos, repo)
	return Save(c)
}
