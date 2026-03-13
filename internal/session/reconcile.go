package session

import (
	"fmt"
)

// Reconcile synchronises the store with actual tmux state:
//   - Running sessions whose tmux session is gone are marked "finished".
//   - Orphan grove-* tmux sessions not tracked in the DB are killed.
func (m *Manager) Reconcile() error {
	// Get all sessions from the store.
	sessions, err := m.Store.ListSessions()
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	// Build a set of tmux session names we expect to exist.
	knownTmux := make(map[string]string) // tmux name -> session ID
	for _, sess := range sessions {
		if sess.Status == "running" {
			if !SessionExists(sess.TmuxSession) {
				if err := m.Store.UpdateSessionStatus(sess.ID, "finished"); err != nil {
					return fmt.Errorf("marking session %s finished: %w", sess.ID, err)
				}
			} else {
				knownTmux[sess.TmuxSession] = sess.ID
			}
		}
	}

	// Find orphan grove-* tmux sessions.
	tmuxSessions, err := ListSessions()
	if err != nil {
		return fmt.Errorf("listing tmux sessions: %w", err)
	}
	for _, name := range tmuxSessions {
		if _, ok := knownTmux[name]; !ok {
			_ = KillSession(name)
		}
	}

	return nil
}
