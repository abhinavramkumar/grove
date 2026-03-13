package app

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

// themePickerDoneMsg signals that the user selected a theme.
type themePickerDoneMsg struct{}

// themePickerCancelMsg signals that the user cancelled the theme picker.
type themePickerCancelMsg struct{}

// ThemePickerModel is the theme selection overlay.
type ThemePickerModel struct {
	config *config.Config
}

// NewThemePickerModel creates a new theme picker.
func NewThemePickerModel(cfg *config.Config) ThemePickerModel {
	return ThemePickerModel{config: cfg}
}

// Update handles messages for the theme picker.
func (m ThemePickerModel) Update(msg tea.Msg) (ThemePickerModel, tea.Cmd) {
	return m, nil
}

// View renders the theme picker.
func (m ThemePickerModel) View() string {
	return "theme picker stub"
}
