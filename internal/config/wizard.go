package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// RunWizard prompts the user for initial configuration and returns
// a Config. It reads from r and writes prompts to w.
func RunWizard(r io.Reader, w io.Writer) (*Config, error) {
	scanner := bufio.NewScanner(r)

	aiTool := prompt(scanner, w, "Default AI tool (claude/codex/opencode)", "claude")
	worktreeBase := prompt(scanner, w, "Worktree base directory", "~/Projects/Work")
	setupStr := prompt(scanner, w, "Setup commands (comma-separated, or empty)", "")

	var setupCommands []string
	if setupStr != "" {
		for _, cmd := range strings.Split(setupStr, ",") {
			if trimmed := strings.TrimSpace(cmd); trimmed != "" {
				setupCommands = append(setupCommands, trimmed)
			}
		}
	}

	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       aiTool,
			WorktreeBase: worktreeBase,
		},
		Worktree: WorktreeConfig{
			SetupCommands: setupCommands,
		},
		Tools: map[string]ToolConfig{
			"claude":   {Command: "claude", Args: []string{"-p"}},
			"codex":    {Command: "codex"},
			"opencode": {Command: "opencode"},
		},
	}

	if err := Save(cfg); err != nil {
		return nil, fmt.Errorf("saving config: %w", err)
	}

	return cfg, nil
}

func prompt(scanner *bufio.Scanner, w io.Writer, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", label, defaultVal)
	} else {
		fmt.Fprintf(w, "%s: ", label)
	}
	if scanner.Scan() {
		if val := strings.TrimSpace(scanner.Text()); val != "" {
			return val
		}
	}
	return defaultVal
}
