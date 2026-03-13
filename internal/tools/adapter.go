package tools

import (
	"fmt"
	"strings"

	"github.com/abhinav/grove/internal/config"
	"github.com/google/uuid"
)

// Adapter defines how to launch and resume a specific AI coding tool.
// Each tool has its own session management semantics.
type Adapter interface {
	// Name returns the tool identifier.
	Name() string
	// NewSessionCmd returns a shell command to start a new session.
	// The returned toolSessionID is the native session ID to store for later resume.
	NewSessionCmd(dir, prompt, planFile string) (cmd string, toolSessionID string)
	// ResumeSessionCmd returns a shell command to resume an existing session.
	// toolSessionID is the value previously returned by NewSessionCmd.
	ResumeSessionCmd(dir, toolSessionID string) string
	// SupportsResume reports whether this tool can resume previous sessions.
	SupportsResume() bool
}

// ClaudeAdapter implements Adapter for Claude Code CLI.
// Creates sessions with --session-id, resumes with --resume.
type ClaudeAdapter struct {
	Command string
	Args    []string // additional args (e.g. from config)
}

func (a *ClaudeAdapter) Name() string { return "claude" }

func (a *ClaudeAdapter) NewSessionCmd(dir, prompt, planFile string) (string, string) {
	sessionID := uuid.New().String()
	parts := []string{fmt.Sprintf("cd %q", dir)}

	cmdParts := []string{a.Command}
	cmdParts = append(cmdParts, a.Args...)
	cmdParts = append(cmdParts, "--session-id", sessionID)

	if planFile != "" {
		cmdParts = append(cmdParts, "--plan-file", planFile)
	} else if prompt != "" {
		cmdParts = append(cmdParts, "-p", fmt.Sprintf("%q", prompt))
	}

	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && "), sessionID
}

func (a *ClaudeAdapter) ResumeSessionCmd(dir, toolSessionID string) string {
	parts := []string{fmt.Sprintf("cd %q", dir)}
	cmdParts := []string{a.Command, "--resume", toolSessionID}
	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && ")
}

func (a *ClaudeAdapter) SupportsResume() bool { return true }

// OpenCodeAdapter implements Adapter for OpenCode.
// Resumes sessions with --session <id> or --continue.
type OpenCodeAdapter struct {
	Command string
	Args    []string
}

func (a *OpenCodeAdapter) Name() string { return "opencode" }

func (a *OpenCodeAdapter) NewSessionCmd(dir, prompt, planFile string) (string, string) {
	// OpenCode generates its own session ID; we can't pre-assign one.
	// We store the directory as the "session ID" and use --continue for resume.
	parts := []string{fmt.Sprintf("cd %q", dir)}

	cmdParts := []string{a.Command}
	cmdParts = append(cmdParts, a.Args...)

	if prompt != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("%q", prompt))
	}

	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && "), "" // no pre-assigned session ID
}

func (a *OpenCodeAdapter) ResumeSessionCmd(dir, toolSessionID string) string {
	parts := []string{fmt.Sprintf("cd %q", dir)}
	cmdParts := []string{a.Command, "--continue"}
	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && ")
}

func (a *OpenCodeAdapter) SupportsResume() bool { return true }

// CodexAdapter implements Adapter for OpenAI Codex.
// Codex uses immutable rollouts — no true session resume.
type CodexAdapter struct {
	Command string
	Args    []string
}

func (a *CodexAdapter) Name() string { return "codex" }

func (a *CodexAdapter) NewSessionCmd(dir, prompt, planFile string) (string, string) {
	parts := []string{fmt.Sprintf("cd %q", dir)}

	cmdParts := []string{a.Command}
	cmdParts = append(cmdParts, a.Args...)

	if prompt != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("%q", prompt))
	}

	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && "), ""
}

func (a *CodexAdapter) ResumeSessionCmd(dir, toolSessionID string) string {
	// Codex doesn't support true resume; start fresh.
	cmd, _ := a.NewSessionCmd(dir, "", "")
	return cmd
}

func (a *CodexAdapter) SupportsResume() bool { return false }

// GenericAdapter is a fallback for unknown/custom tools.
type GenericAdapter struct {
	ToolName string
	Command  string
	Args     []string
}

func (a *GenericAdapter) Name() string { return a.ToolName }

func (a *GenericAdapter) NewSessionCmd(dir, prompt, planFile string) (string, string) {
	parts := []string{fmt.Sprintf("cd %q", dir)}

	cmdParts := []string{a.Command}
	cmdParts = append(cmdParts, a.Args...)

	if prompt != "" {
		cmdParts = append(cmdParts, fmt.Sprintf("%q", prompt))
	}

	parts = append(parts, strings.Join(cmdParts, " "))
	return strings.Join(parts, " && "), ""
}

func (a *GenericAdapter) ResumeSessionCmd(dir, toolSessionID string) string {
	cmd, _ := a.NewSessionCmd(dir, "", "")
	return cmd
}

func (a *GenericAdapter) SupportsResume() bool { return false }

// LoadAdapters builds a map of tool adapters from the config.
// Built-in tools (claude, opencode, codex) get their specific adapters.
// Config entries can override the command/args for built-in tools,
// or define entirely new tools (which get GenericAdapter).
func LoadAdapters(cfg *config.Config) map[string]Adapter {
	adapters := map[string]Adapter{
		"claude":   &ClaudeAdapter{Command: "claude"},
		"opencode": &OpenCodeAdapter{Command: "opencode"},
		"codex":    &CodexAdapter{Command: "codex"},
	}

	if cfg == nil {
		return adapters
	}

	for name, tc := range cfg.Tools {
		switch name {
		case "claude":
			adapters[name] = &ClaudeAdapter{Command: tc.Command, Args: tc.Args}
		case "opencode":
			adapters[name] = &OpenCodeAdapter{Command: tc.Command, Args: tc.Args}
		case "codex":
			adapters[name] = &CodexAdapter{Command: tc.Command, Args: tc.Args}
		default:
			adapters[name] = &GenericAdapter{
				ToolName: name,
				Command:  tc.Command,
				Args:     tc.Args,
			}
		}
	}

	return adapters
}
