package session

import (
	"os/exec"
	"strings"
)

// SessionExists checks whether a tmux session with the given name exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// NewSession creates a new detached tmux session.
func NewSession(name, dir, command string) error {
	args := []string{"new-session", "-d", "-s", name, "-c", dir}
	if command != "" {
		args = append(args, command)
	}
	cmd := exec.Command("tmux", args...)
	return cmd.Run()
}

// KillSession destroys a tmux session by name.
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}

// AttachSession returns an exec.Cmd that will attach to the named tmux session.
// The caller is responsible for running it (e.g. via tea.Exec in the TUI).
func AttachSession(name string) *exec.Cmd {
	return exec.Command("tmux", "attach-session", "-t", name)
}

// CapturePane captures the visible content of a tmux session's pane.
func CapturePane(name string) (string, error) {
	cmd := exec.Command("tmux", "capture-pane", "-p", "-e", "-t", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ListSessions returns the names of all tmux sessions that start with "grove-".
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	out, err := cmd.Output()
	if err != nil {
		// tmux exits non-zero when the server isn't running (no sessions).
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return nil, nil
		}
		return nil, err
	}

	var sessions []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "grove-") {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}
