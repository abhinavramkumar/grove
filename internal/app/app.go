package app

import (
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
)

// view represents which screen is active.
type view int

const (
	viewList view = iota
	viewPeek // placeholder for Phase 4
	viewNew  // placeholder for Phase 4
)

// tickMsg triggers periodic reconciliation and list refresh.
type tickMsg time.Time

// sessionsMsg carries a refreshed session list.
type sessionsMsg []*store.Session

// errMsg carries an error to display briefly.
type errMsg struct{ err error }

// infoMsg carries an info message to display briefly.
type infoMsg string

// attachDoneMsg is sent when the user detaches from a tmux session.
type attachDoneMsg struct{ err error }

// AppModel is the root Bubbletea model.
type AppModel struct {
	list    ListModel
	view    view
	store   *store.Store
	config  *config.Config
	manager *session.Manager
	flash   string // transient message displayed in the status bar
}

// New creates the root model.
func New(s *store.Store, cfg *config.Config, mgr *session.Manager) AppModel {
	return AppModel{
		view:    viewList,
		store:   s,
		config:  cfg,
		manager: mgr,
	}
}

// Init runs reconciliation and loads sessions.
func (m AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.reconcileAndLoad(),
		tickCmd(),
	)
}

// Update handles messages.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.list.Width = msg.Width
		// Reserve 1 line for the status bar.
		m.list.Height = msg.Height - 1
		return m, nil

	case tickMsg:
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())

	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil

	case errMsg:
		m.flash = errorStyle.Render(msg.err.Error())
		return m, nil

	case infoMsg:
		m.flash = infoStyle.Render(string(msg))
		return m, nil

	case attachDoneMsg:
		// After detaching from tmux, refresh the list.
		return m, m.reconcileAndLoad()

	case tea.KeyMsg:
		// Clear flash on any keypress.
		m.flash = ""
		return m.handleKey(msg)
	}

	return m, nil
}

func (m AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, keys.Up):
		m.list.MoveUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.list.MoveDown()
		return m, nil

	case key.Matches(msg, keys.Attach):
		return m, m.attachSelected()

	case key.Matches(msg, keys.Delete):
		return m, m.deleteSelected()

	case key.Matches(msg, keys.Stop):
		return m, m.stopSelected()

	case key.Matches(msg, keys.Resume):
		return m, m.resumeSelected()

	case key.Matches(msg, keys.Escape):
		// No-op for now; only list view exists.
		return m, nil
	}

	return m, nil
}

// View renders the current view.
func (m AppModel) View() string {
	body := m.list.View()

	// Status bar.
	bar := statusBarStyle.Width(m.list.Width).Render(statusBarHelp())
	if m.flash != "" {
		bar = statusBarStyle.Width(m.list.Width).Render(m.flash)
	}

	return body + "\n" + bar
}

// --- Commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m AppModel) reconcileAndLoad() tea.Cmd {
	return func() tea.Msg {
		_ = m.manager.Reconcile()
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}

func (m AppModel) attachSelected() tea.Cmd {
	sess := m.list.Selected()
	if sess == nil {
		return nil
	}
	if sess.Status != "running" {
		return func() tea.Msg { return infoMsg("session is not running") }
	}

	cmd, err := m.manager.Attach(sess.ID)
	if err != nil {
		return func() tea.Msg { return errMsg{err} }
	}

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return attachDoneMsg{err}
	})
}

func (m AppModel) deleteSelected() tea.Cmd {
	sess := m.list.Selected()
	if sess == nil {
		return nil
	}
	id := sess.ID
	return func() tea.Msg {
		if err := m.manager.Delete(id); err != nil {
			return errMsg{err}
		}
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}

func (m AppModel) stopSelected() tea.Cmd {
	sess := m.list.Selected()
	if sess == nil {
		return nil
	}
	if sess.Status != "running" {
		return func() tea.Msg { return infoMsg("session is not running") }
	}
	id := sess.ID
	return func() tea.Msg {
		if err := m.manager.Stop(id); err != nil {
			return errMsg{err}
		}
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}

func (m AppModel) resumeSelected() tea.Cmd {
	sess := m.list.Selected()
	if sess == nil {
		return nil
	}
	if sess.Status == "running" {
		return func() tea.Msg { return infoMsg("session is already running") }
	}
	id := sess.ID
	return func() tea.Msg {
		if _, err := m.manager.Resume(id); err != nil {
			return errMsg{err}
		}
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}
