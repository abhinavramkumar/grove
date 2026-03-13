package config

import (
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents the Grove configuration file.
type Config struct {
	Defaults DefaultsConfig        `toml:"defaults"`
	Worktree WorktreeConfig        `toml:"worktree"`
	Tools    map[string]ToolConfig `toml:"tools"`
}

type DefaultsConfig struct {
	AITool       string `toml:"ai_tool"`
	WorktreeBase string `toml:"worktree_base"`
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
