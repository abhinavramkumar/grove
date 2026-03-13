package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/store"
)

// ListModel holds the session list state.
type ListModel struct {
	Sessions  []*store.Session
	Cursor    int
	Width     int
	Height    int
	RepoOrder []string // repo basenames in config order, for sorting

	Filtering  bool   // true when filter input is focused
	FilterText string // current filter query

	filterInput textinput.Model
}

// Selected returns the currently selected session from the filtered list, or nil.
func (m *ListModel) Selected() *store.Session {
	filtered := m.FilteredSessions()
	if len(filtered) == 0 {
		return nil
	}
	if m.Cursor >= len(filtered) {
		m.Cursor = len(filtered) - 1
	}
	return filtered[m.Cursor]
}

// MoveUp moves the cursor up one row within the filtered list.
func (m *ListModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves the cursor down one row within the filtered list.
func (m *ListModel) MoveDown() {
	filtered := m.FilteredSessions()
	if m.Cursor < len(filtered)-1 {
		m.Cursor++
	}
}

// ClampCursor ensures the cursor is within bounds of the filtered list.
func (m *ListModel) ClampCursor() {
	filtered := m.FilteredSessions()
	if m.Cursor >= len(filtered) {
		m.Cursor = max(0, len(filtered)-1)
	}
}

// FilteredSessions returns the sessions matching the current filter.
func (m *ListModel) FilteredSessions() []*store.Session {
	return filterSessions(m.Sessions, m.FilterText)
}

// StartFilter initializes and focuses the filter text input.
func (m *ListModel) StartFilter() {
	ti := textinput.New()
	ti.Placeholder = "type to filter…"
	ti.CharLimit = 64
	ti.Width = 30
	ti.SetValue(m.FilterText)
	ti.Focus()
	m.filterInput = ti
	m.Filtering = true
}

// ClearFilter resets the filter and exits filter mode.
func (m *ListModel) ClearFilter() {
	m.FilterText = ""
	m.filterInput.Blur()
	m.Filtering = false
	m.Cursor = 0
}

// CommitFilter keeps the current filter text and exits filter mode.
func (m *ListModel) CommitFilter() {
	m.FilterText = m.filterInput.Value()
	m.filterInput.Blur()
	m.Filtering = false
}

// HandleFilterKey processes a key event while the filter input is focused.
// Returns a tea.Cmd if the input needs to update (e.g., cursor blink).
func (m *ListModel) HandleFilterKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.FilterText = m.filterInput.Value()
	m.Cursor = 0 // reset cursor as filter changes
	return cmd
}

// FilterInputView returns the rendered filter text input.
func (m *ListModel) FilterInputView() string {
	return m.filterInput.View()
}

// statusCounts returns running, stopped, finished counts from all sessions (unfiltered).
func (m *ListModel) statusCounts() (int, int, int) {
	var running, stopped, finished int
	for _, sess := range m.Sessions {
		switch sess.Status {
		case "running":
			running++
		case "stopped":
			stopped++
		default:
			finished++
		}
	}
	return running, stopped, finished
}

// View renders the session list table.
func (m *ListModel) View() string {
	filtered := m.FilteredSessions()

	// Header bar.
	running, stopped, finished := m.statusCounts()
	headerBar := lipgloss.NewStyle().Bold(true).Foreground(ActiveTheme.Accent).Render("grove") +
		S.HelpDesc.Render(" │ ") +
		lipgloss.NewStyle().Foreground(ActiveTheme.Green).Render(fmt.Sprintf("%d running", running)) +
		S.HelpDesc.Render(" / ") +
		lipgloss.NewStyle().Foreground(ActiveTheme.Yellow).Render(fmt.Sprintf("%d stopped", stopped)) +
		S.HelpDesc.Render(" / ") +
		lipgloss.NewStyle().Foreground(ActiveTheme.FgDim).Render(fmt.Sprintf("%d finished", finished))

	if len(filtered) == 0 {
		msg := emptyStateView(m.FilterText != "", m.Width, m.Height-1)
		return headerBar + "\n" + msg
	}

	colGap := "  "

	// Column widths: status dot(2) + name + repo + tool + directory + age.
	toolW := 10
	repoW := 14
	durW := 10
	nameW := 20
	fixedW := 2 + nameW + repoW + toolW + durW + len(colGap)*4
	dirW := m.Width - fixedW
	if dirW < 10 {
		dirW = 10
	}

	var b strings.Builder

	// Header bar.
	b.WriteString(headerBar)
	b.WriteByte('\n')

	// Table header.
	hdr := fmt.Sprintf("  %-*s%s%-*s%s%-*s%s%-*s%s%*s",
		nameW, "NAME", colGap,
		repoW, "REPO", colGap,
		toolW, "TOOL", colGap,
		dirW, "DIRECTORY", colGap,
		durW, "AGE")
	b.WriteString(S.Header.Render(hdr))
	b.WriteByte('\n')

	// Separator line.
	b.WriteString(lipgloss.NewStyle().Foreground(ActiveTheme.FgFaint).Render(strings.Repeat("─", m.Width)))
	b.WriteByte('\n')

	// Determine how many rows we can show (height minus header bar, table header, separator).
	maxRows := m.Height - 3
	if maxRows < 1 {
		maxRows = len(filtered)
	}

	zebraRow := lipgloss.NewStyle().Background(ActiveTheme.BgAlt)

	for i, sess := range filtered {
		if i >= maxRows {
			break
		}

		dot := statusDot(sess.Status)
		name := truncate(sess.Name, nameW-1)
		repo := truncate(repoDisplayName(sess), repoW)
		tool := truncate(sess.Tool, toolW)
		dir := truncateLeft(sess.Directory, dirW)
		dur := formatDuration(time.Since(sess.CreatedAt))

		row := fmt.Sprintf("%s %-*s%s%-*s%s%-*s%s%-*s%s%*s",
			dot,
			nameW-1, name, colGap,
			repoW, repo, colGap,
			toolW, tool, colGap,
			dirW, dir, colGap,
			durW, dur)

		if i == m.Cursor {
			row = "▎" + S.SelectedRow.Width(m.Width-1).Render(row)
		} else if sess.Status == "finished" {
			row = S.FinishedRow.Render(row)
		} else if i%2 == 1 {
			row = zebraRow.Render(row)
		} else {
			row = S.NormalRow.Render(row)
		}

		b.WriteString(row)
		if i < len(filtered)-1 && i < maxRows-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// emptyStateView renders a centered empty state.
func emptyStateView(hasFilter bool, width, height int) string {
	if hasFilter {
		msg := S.Empty.Render("No matching sessions.")
		if width > 0 {
			return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
		}
		return msg
	}

	diamond := lipgloss.NewStyle().Foreground(ActiveTheme.Accent).Bold(true).Render("◇")
	title := S.Empty.Render("No sessions yet")
	subtitle := lipgloss.NewStyle().Foreground(ActiveTheme.FgDim).Render("Start your first AI coding session")
	pill := S.HelpKey.Render("n") + " " + S.HelpDesc.Render("new session")

	content := diamond + "\n" + title + "\n" + subtitle + "\n\n" + pill

	if width > 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	}
	return content
}

// repoDisplayName returns the base directory name of the session's repo root,
// or "—" if no repo root is set.
func repoDisplayName(sess *store.Session) string {
	if sess.RepoRoot == nil || *sess.RepoRoot == "" {
		return "—"
	}
	return filepath.Base(*sess.RepoRoot)
}

// filterSessions returns sessions matching query across name, repo, tool, and directory.
// If query is empty, all sessions are returned.
func filterSessions(sessions []*store.Session, query string) []*store.Session {
	if query == "" {
		return sessions
	}
	q := strings.ToLower(query)
	var result []*store.Session
	for _, sess := range sessions {
		if strings.Contains(strings.ToLower(sess.Name), q) ||
			strings.Contains(strings.ToLower(repoDisplayName(sess)), q) ||
			strings.Contains(strings.ToLower(sess.Tool), q) ||
			strings.Contains(strings.ToLower(sess.Directory), q) {
			result = append(result, sess)
		}
	}
	return result
}

// formatDuration returns a human-friendly short duration string.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "<1m"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// truncate trims s to maxLen, adding an ellipsis if needed.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "."
	}
	return s[:maxLen-1] + "~"
}

// truncateLeft trims the left side of s (for paths), showing the tail.
func truncateLeft(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 2 {
		return ".."
	}
	return "~" + s[len(s)-maxLen+1:]
}
