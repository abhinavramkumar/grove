package session

import (
	"fmt"
	"os/exec"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/tools"
)

// Manager coordinates session lifecycle between the store and tmux.
type Manager struct {
	Store    *store.Store
	Adapters map[string]tools.Adapter
}

// NewManager creates a Manager with adapters loaded from config.
func NewManager(s *store.Store, cfg *config.Config) *Manager {
	return &Manager{
		Store:    s,
		Adapters: tools.LoadAdapters(cfg),
	}
}

// Create builds a new session: inserts it in the store, starts a tmux session,
// and returns the session record.
func (m *Manager) Create(name, tool, dir string, worktree *string, prompt, planFile string) (*store.Session, error) {
	adapter, ok := m.Adapters[tool]
	if !ok {
		return nil, fmt.Errorf("unknown tool %q", tool)
	}

	cmd, toolSessionID := adapter.NewSessionCmd(dir, prompt, planFile)

	var promptPtr, planPtr, toolSessPtr *string
	if prompt != "" {
		promptPtr = &prompt
	}
	if planFile != "" {
		planPtr = &planFile
	}
	if toolSessionID != "" {
		toolSessPtr = &toolSessionID
	}

	sess, err := m.Store.CreateSession(name, tool, dir, worktree, promptPtr, planPtr, toolSessPtr)
	if err != nil {
		return nil, fmt.Errorf("creating session record: %w", err)
	}

	if err := NewSession(sess.TmuxSession, dir, cmd); err != nil {
		_ = m.Store.DeleteSession(sess.ID)
		return nil, fmt.Errorf("creating tmux session: %w", err)
	}

	return sess, nil
}

// Resume restarts a stopped session using the tool's native resume mechanism.
// Creates a new tmux session running the tool's resume command.
func (m *Manager) Resume(sessionID string) (*store.Session, error) {
	sess, err := m.Store.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	if sess.Status == "running" && SessionExists(sess.TmuxSession) {
		return sess, nil // already running
	}

	adapter, ok := m.Adapters[sess.Tool]
	if !ok {
		return nil, fmt.Errorf("unknown tool %q", sess.Tool)
	}

	var cmd string
	toolSessID := ""
	if sess.ToolSessionID != nil {
		toolSessID = *sess.ToolSessionID
	}

	if adapter.SupportsResume() && toolSessID != "" {
		cmd = adapter.ResumeSessionCmd(sess.Directory, toolSessID)
	} else {
		// Fall back to a fresh session (no resume support or no session ID).
		prompt := ""
		if sess.Prompt != nil {
			prompt = *sess.Prompt
		}
		planFile := ""
		if sess.PlanFile != nil {
			planFile = *sess.PlanFile
		}
		var newToolSessID string
		cmd, newToolSessID = adapter.NewSessionCmd(sess.Directory, prompt, planFile)
		if newToolSessID != "" {
			_ = m.Store.UpdateToolSessionID(sess.ID, newToolSessID)
		}
	}

	if err := NewSession(sess.TmuxSession, sess.Directory, cmd); err != nil {
		return nil, fmt.Errorf("creating tmux session: %w", err)
	}

	if err := m.Store.UpdateSessionStatus(sess.ID, "running"); err != nil {
		return nil, fmt.Errorf("updating session status: %w", err)
	}

	return m.Store.GetSession(sess.ID)
}

// Attach returns an exec.Cmd that attaches to the session's tmux session.
func (m *Manager) Attach(sessionID string) (*exec.Cmd, error) {
	sess, err := m.Store.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("getting session: %w", err)
	}

	if !SessionExists(sess.TmuxSession) {
		return nil, fmt.Errorf("tmux session %s is not running", sess.TmuxSession)
	}

	return AttachSession(sess.TmuxSession), nil
}

// Delete kills the tmux session and removes the session from the store.
func (m *Manager) Delete(sessionID string) error {
	sess, err := m.Store.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("getting session: %w", err)
	}

	_ = KillSession(sess.TmuxSession)

	if err := m.Store.DeleteSession(sessionID); err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

// Stop kills the tmux session and marks the session as stopped.
func (m *Manager) Stop(sessionID string) error {
	sess, err := m.Store.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("getting session: %w", err)
	}

	_ = KillSession(sess.TmuxSession)

	if err := m.Store.UpdateSessionStatus(sessionID, "stopped"); err != nil {
		return fmt.Errorf("updating session status: %w", err)
	}
	return nil
}
