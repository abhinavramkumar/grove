package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/store"
)

// ListModel holds the session list state.
type ListModel struct {
	Sessions []*store.Session
	Cursor   int
	Width    int
	Height   int
}

// Selected returns the currently selected session, or nil if the list is empty.
func (m *ListModel) Selected() *store.Session {
	if len(m.Sessions) == 0 {
		return nil
	}
	if m.Cursor >= len(m.Sessions) {
		m.Cursor = len(m.Sessions) - 1
	}
	return m.Sessions[m.Cursor]
}

// MoveUp moves the cursor up one row.
func (m *ListModel) MoveUp() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

// MoveDown moves the cursor down one row.
func (m *ListModel) MoveDown() {
	if m.Cursor < len(m.Sessions)-1 {
		m.Cursor++
	}
}

// ClampCursor ensures the cursor is within bounds.
func (m *ListModel) ClampCursor() {
	if m.Cursor >= len(m.Sessions) {
		m.Cursor = max(0, len(m.Sessions)-1)
	}
}

// View renders the session list table.
func (m *ListModel) View() string {
	if len(m.Sessions) == 0 {
		msg := "No sessions. Press n to create one."
		if m.Width > 0 {
			return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
				emptyStyle.Render(msg))
		}
		return emptyStyle.Render(msg)
	}

	// Column widths: status+name, tool, directory, duration.
	// Reserve space: 2 for status dot + space, some padding between columns.
	colGap := "  "

	// Determine available width for directory truncation.
	toolW := 10
	durW := 10
	fixedW := 2 + toolW + durW + len(colGap)*3 // dot+space, gaps
	nameW := 20
	dirW := m.Width - fixedW - nameW
	if dirW < 10 {
		dirW = 10
	}

	var b strings.Builder

	// Header.
	hdr := fmt.Sprintf("  %-*s%s%-*s%s%-*s%s%*s",
		nameW, "NAME", colGap,
		toolW, "TOOL", colGap,
		dirW, "DIRECTORY", colGap,
		durW, "AGE")
	b.WriteString(headerStyle.Render(hdr))
	b.WriteByte('\n')

	// Determine how many rows we can show (height minus header line).
	maxRows := m.Height - 1
	if maxRows < 1 {
		maxRows = len(m.Sessions)
	}

	for i, sess := range m.Sessions {
		if i >= maxRows {
			break
		}

		dot := statusDot(sess.Status)
		name := truncate(sess.Name, nameW-1)
		tool := truncate(sess.Tool, toolW)
		dir := truncateLeft(sess.Directory, dirW)
		dur := formatDuration(time.Since(sess.CreatedAt))

		row := fmt.Sprintf("%s %-*s%s%-*s%s%-*s%s%*s",
			dot,
			nameW-1, name, colGap,
			toolW, tool, colGap,
			dirW, dir, colGap,
			durW, dur)

		if i == m.Cursor {
			row = selectedRow.Width(m.Width).Render(row)
		} else {
			row = normalRow.Render(row)
		}

		b.WriteString(row)
		if i < len(m.Sessions)-1 && i < maxRows-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
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
