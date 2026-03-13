package app

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
)

// peekTickMsg triggers a pane capture refresh.
type peekTickMsg time.Time

// peekContentMsg carries captured pane content.
type peekContentMsg string

// peekErrMsg carries a capture error.
type peekErrMsg struct{ err error }

// peekBackMsg signals return to the list view.
type peekBackMsg struct{}

// peekAttachMsg signals the user wants to attach to the session.
type peekAttachMsg struct{ sessionID string }

// PeekModel displays a live preview of a session's tmux pane.
type PeekModel struct {
	session  *store.Session
	viewport viewport.Model
	manager  *session.Manager
	width    int
	height   int
	ready    bool
}

// NewPeekModel creates a peek view for the given session.
func NewPeekModel(sess *store.Session, mgr *session.Manager, width, height int) PeekModel {
	// Reserve 2 lines: 1 for header, 1 for status bar.
	vp := viewport.New(width, max(1, height-2))
	vp.SetContent("Loading...")

	return PeekModel{
		session:  sess,
		viewport: vp,
		manager:  mgr,
		width:    width,
		height:   height,
		ready:    true,
	}
}

// Init starts the capture tick.
func (m PeekModel) Init() tea.Cmd {
	return tea.Batch(
		m.captureCmd(),
		peekTickCmd(),
	)
}

// Update handles messages for the peek view.
func (m PeekModel) Update(msg tea.Msg) (PeekModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = max(1, msg.Height-2)
		return m, nil

	case peekTickMsg:
		return m, tea.Batch(m.captureCmd(), peekTickCmd())

	case peekContentMsg:
		m.viewport.SetContent(string(msg))
		return m, nil

	case peekErrMsg:
		m.viewport.SetContent(fmt.Sprintf("Error capturing pane: %v", msg.err))
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass scroll events to viewport.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m PeekModel) handleKey(msg tea.KeyMsg) (PeekModel, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		return m, func() tea.Msg { return peekBackMsg{} }

	case key.Matches(msg, keys.Attach):
		return m, func() tea.Msg { return peekAttachMsg{sessionID: m.session.ID} }

	case key.Matches(msg, keys.Up):
		m.viewport.ScrollUp(1)
		return m, nil

	case key.Matches(msg, keys.Down):
		m.viewport.ScrollDown(1)
		return m, nil
	}

	// Also let viewport handle its own keys for pgup/pgdown etc.
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the peek panel.
func (m PeekModel) View() string {
	header := S.PeekHeader.Width(m.width).Render(
		fmt.Sprintf(" Peek: %s (%s)", m.session.Name, m.session.TmuxSession),
	)

	bar := S.StatusBar.Width(m.width).Render(
		peekHelpText(),
	)

	return header + "\n" + m.viewport.View() + "\n" + bar
}

func peekHelpText() string {
	return S.HelpKey.Render("esc") + S.HelpDesc.Render(":back") + "  " +
		S.HelpKey.Render("enter") + S.HelpDesc.Render(":attach") + "  " +
		S.HelpKey.Render("↑↓") + S.HelpDesc.Render(":scroll")
}

// captureCmd returns a command that captures the tmux pane content.
func (m PeekModel) captureCmd() tea.Cmd {
	name := m.session.TmuxSession
	return func() tea.Msg {
		content, err := session.CapturePane(name)
		if err != nil {
			return peekErrMsg{err}
		}
		return peekContentMsg(content)
	}
}

// peekTickCmd returns a tick that fires every 500ms.
func peekTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return peekTickMsg(t)
	})
}
