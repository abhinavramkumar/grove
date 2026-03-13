package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/config"
)

// themePickerDoneMsg signals that the user selected a theme.
type themePickerDoneMsg struct{}

// themePickerCancelMsg signals that the user cancelled the theme picker.
type themePickerCancelMsg struct{}

// ThemePickerModel is the theme selection overlay.
type ThemePickerModel struct {
	cursor        int
	originalTheme Theme
	config        *config.Config
}

// NewThemePickerModel creates a new theme picker.
func NewThemePickerModel(cfg *config.Config) ThemePickerModel {
	cursor := 0
	for i, t := range Themes {
		if t.Name == ActiveTheme.Name {
			cursor = i
			break
		}
	}
	return ThemePickerModel{
		cursor:        cursor,
		originalTheme: ActiveTheme,
		config:        cfg,
	}
}

// Update handles messages for the theme picker.
func (m ThemePickerModel) Update(msg tea.Msg) (ThemePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				SetTheme(Themes[m.cursor])
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(Themes)-1 {
				m.cursor++
				SetTheme(Themes[m.cursor])
			}
		case key.Matches(msg, keys.Attach): // enter
			m.config.Defaults.Theme = Themes[m.cursor].Name
			if err := config.Save(m.config); err != nil {
				return m, func() tea.Msg { return errMsg{err} }
			}
			return m, func() tea.Msg { return themePickerDoneMsg{} }
		case key.Matches(msg, keys.Escape):
			SetTheme(m.originalTheme)
			return m, func() tea.Msg { return themePickerCancelMsg{} }
		}
	}
	return m, nil
}

// View renders the theme picker.
func (m ThemePickerModel) View() string {
	var b strings.Builder

	// Title.
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(ActiveTheme.Accent).
		Render("Theme")
	b.WriteString(title)
	b.WriteString("\n")

	// Separator.
	b.WriteString(lipgloss.NewStyle().
		Foreground(ActiveTheme.FgFaint).
		Render("───────────────────"))
	b.WriteString("\n")

	// Theme list.
	for i, t := range Themes {
		if i == m.cursor {
			// Selected row: accent border + highlighted bg + bold name.
			row := lipgloss.NewStyle().
				Bold(true).
				Foreground(ActiveTheme.Accent).
				Background(ActiveTheme.BgHighlight).
				Render(fmt.Sprintf(" %s ", t.Name))
			b.WriteString("▎" + row)
		} else {
			// Normal row: indented + muted name.
			row := lipgloss.NewStyle().
				Foreground(ActiveTheme.FgDim).
				Render(fmt.Sprintf("  %s", t.Name))
			b.WriteString(row)
		}
		b.WriteString("\n")
	}

	// Footer.
	b.WriteString("\n")
	footer := S.HelpKey.Render("j/k") + S.HelpDesc.Render(":navigate  ") +
		S.HelpKey.Render("enter") + S.HelpDesc.Render(":apply  ") +
		S.HelpKey.Render("esc") + S.HelpDesc.Render(":cancel")
	b.WriteString(footer)

	return b.String()
}
