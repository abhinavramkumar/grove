package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/worktree"
)

// view represents which screen is active.
type view int

const (
	viewList view = iota
	viewPeek
	viewNew
	viewHelp
	viewPruneConfirm
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

// pruneConfirm holds state for the worktree prune confirmation dialog.
type pruneConfirm struct {
	session  *store.Session
	dirty    bool // true if worktree has uncommitted changes
	repoDir  string
}

// AppModel is the root Bubbletea model.
type AppModel struct {
	list    ListModel
	create  CreateModel
	peek    PeekModel
	prune   pruneConfirm
	view    view
	store   *store.Store
	config  *config.Config
	manager *session.Manager
	flash   string // transient message displayed in the status bar
	width   int
	height  int
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
	// Route to active sub-model first for non-list views.
	switch m.view {
	case viewNew:
		return m.updateCreate(msg)
	case viewPeek:
		return m.updatePeek(msg)
	case viewHelp:
		return m.updateHelp(msg)
	case viewPruneConfirm:
		return m.updatePruneConfirm(msg)
	}

	// When filter input is active, intercept key events.
	if m.list.Filtering {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.Type {
			case tea.KeyEsc:
				m.list.ClearFilter()
				return m, nil
			case tea.KeyEnter:
				m.list.CommitFilter()
				return m, nil
			default:
				cmd := m.list.HandleFilterKey(msg)
				return m, cmd
			}
		}
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
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

	case pruneReadyMsg:
		m.prune = pruneConfirm{
			session: msg.session,
			dirty:   msg.dirty,
			repoDir: msg.repoDir,
		}
		m.view = viewPruneConfirm
		return m, nil

	case tea.KeyMsg:
		// Clear flash on any keypress.
		m.flash = ""
		return m.handleKey(msg)
	}

	return m, nil
}

func (m AppModel) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.Width = msg.Width
		m.list.Height = msg.Height - 1
		// Also forward to create model.
		m.create, _ = m.create.Update(msg)
		return m, nil

	case createDoneMsg:
		m.view = viewList
		m.flash = infoStyle.Render("created session: " + msg.session.Name)
		return m, m.reconcileAndLoad()

	case createCancelMsg:
		m.view = viewList
		return m, nil

	case createErrMsg:
		// Forward to create model to display the error.
		var cmd tea.Cmd
		m.create, cmd = m.create.Update(msg)
		return m, cmd

	case tickMsg:
		// Keep background reconciliation running.
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())

	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil
	}

	var cmd tea.Cmd
	m.create, cmd = m.create.Update(msg)
	return m, cmd
}

func (m AppModel) updatePeek(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.Width = msg.Width
		m.list.Height = msg.Height - 1
		m.peek, _ = m.peek.Update(msg)
		return m, nil

	case peekBackMsg:
		m.view = viewList
		return m, nil

	case peekAttachMsg:
		m.view = viewList
		return m, m.attachByID(msg.sessionID)

	case tickMsg:
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())

	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil
	}

	var cmd tea.Cmd
	m.peek, cmd = m.peek.Update(msg)
	return m, cmd
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

	case key.Matches(msg, keys.New):
		m.create = NewCreateModel(m.config, m.manager)
		m.view = viewNew
		return m, m.create.Init()

	case key.Matches(msg, keys.Peek):
		sess := m.list.Selected()
		if sess == nil {
			return m, nil
		}
		if sess.Status != "running" {
			m.flash = infoStyle.Render("can only peek running sessions")
			return m, nil
		}
		m.peek = NewPeekModel(sess, m.manager, m.list.Width, m.list.Height+1)
		m.view = viewPeek
		return m, m.peek.Init()

	case key.Matches(msg, keys.Prune):
		return m, m.startPrune()

	case key.Matches(msg, keys.Filter):
		m.list.StartFilter()
		return m, textinput.Blink

	case key.Matches(msg, keys.Help):
		m.view = viewHelp
		return m, nil

	case key.Matches(msg, keys.Escape):
		// No-op in list view.
		return m, nil
	}

	return m, nil
}

// View renders the current view.
func (m AppModel) View() string {
	switch m.view {
	case viewNew:
		return m.create.View()
	case viewPeek:
		return m.peek.View()
	case viewHelp:
		return m.viewHelpOverlay()
	case viewPruneConfirm:
		return m.viewPruneConfirmOverlay()
	}

	body := m.list.View()

	// Status bar — changes depending on filter state.
	var bar string
	if m.list.Filtering {
		bar = filterBarStyle.Width(m.list.Width).Render(
			filterLabelStyle.Render("filter: ") + m.list.FilterInputView())
	} else if m.list.FilterText != "" {
		bar = statusBarStyle.Width(m.list.Width).Render(
			filterActiveIndicator.Render("filter: "+m.list.FilterText) + "  " + statusBarHelp())
	} else if m.flash != "" {
		bar = statusBarStyle.Width(m.list.Width).Render(m.flash)
	} else {
		bar = statusBarStyle.Width(m.list.Width).Render(statusBarHelp())
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

func (m AppModel) attachByID(sessionID string) tea.Cmd {
	cmd, err := m.manager.Attach(sessionID)
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

// --- Prune worktree ---

// pruneReadyMsg is sent after checking the worktree state.
type pruneReadyMsg struct {
	session *store.Session
	dirty   bool
	repoDir string
}

func (m AppModel) startPrune() tea.Cmd {
	sess := m.list.Selected()
	if sess == nil {
		return nil
	}
	if sess.Worktree == nil {
		return func() tea.Msg { return infoMsg("session has no worktree") }
	}
	wtPath := *sess.Worktree
	return func() tea.Msg {
		repoDir, err := worktree.GetMainRepoPath(wtPath)
		if err != nil {
			return errMsg{fmt.Errorf("getting repo path: %w", err)}
		}
		clean, err := worktree.IsWorktreeClean(wtPath)
		if err != nil {
			// If we can't check, treat as dirty to be safe.
			return pruneReadyMsg{session: sess, dirty: true, repoDir: repoDir}
		}
		return pruneReadyMsg{session: sess, dirty: !clean, repoDir: repoDir}
	}
}

func (m AppModel) updatePruneConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.Width = msg.Width
		m.list.Height = msg.Height - 1
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.view = viewList
			return m, m.executePrune()
		case "n", "N", "esc":
			m.view = viewList
			m.flash = infoStyle.Render("prune cancelled")
			return m, nil
		}
	case tickMsg:
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())
	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil
	}
	return m, nil
}

func (m AppModel) executePrune() tea.Cmd {
	sess := m.prune.session
	repoDir := m.prune.repoDir
	force := m.prune.dirty
	wtPath := *sess.Worktree
	sessionID := sess.ID
	stopped := sess.Status == "stopped" || sess.Status == "finished"

	return func() tea.Msg {
		if err := worktree.Remove(repoDir, wtPath, force); err != nil {
			return errMsg{fmt.Errorf("removing worktree: %w", err)}
		}
		// If session is stopped/finished, also delete it.
		if stopped {
			_ = m.manager.Delete(sessionID)
		}
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}

func (m AppModel) viewPruneConfirmOverlay() string {
	sess := m.prune.session
	wtPath := ""
	if sess.Worktree != nil {
		wtPath = *sess.Worktree
	}

	var msg strings.Builder
	msg.WriteString("Prune worktree for session: " + sess.Name + "\n")
	msg.WriteString("Path: " + wtPath + "\n\n")

	if m.prune.dirty {
		msg.WriteString(errorStyle.Render("WARNING: worktree has uncommitted changes!") + "\n\n")
	}

	if sess.Status == "stopped" || sess.Status == "finished" {
		msg.WriteString("Session is " + sess.Status + " and will also be deleted.\n\n")
	}

	msg.WriteString("Are you sure? (y/n)")

	overlay := overlayStyle.Render(msg.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}

// --- Help overlay ---

func (m AppModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.Width = msg.Width
		m.list.Height = msg.Height - 1
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Help), key.Matches(msg, keys.Escape):
			m.view = viewList
			return m, nil
		}
	case tickMsg:
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())
	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil
	}
	return m, nil
}

func (m AppModel) viewHelpOverlay() string {
	bindings := []key.Binding{
		keys.Up, keys.Down, keys.Attach, keys.Peek,
		keys.New, keys.Delete, keys.Stop, keys.Resume,
		keys.Prune, keys.Filter, keys.Help, keys.Quit, keys.Escape,
	}

	var b strings.Builder
	b.WriteString("Keybindings\n")
	b.WriteString(strings.Repeat("─", 28) + "\n")

	for _, binding := range bindings {
		h := binding.Help()
		b.WriteString(fmt.Sprintf("  %-12s %s\n", h.Key, h.Desc))
	}

	b.WriteString("\nPress ? or esc to close")

	overlay := overlayStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}
