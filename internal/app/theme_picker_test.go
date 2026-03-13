package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/config"
)

func TestThemePicker_View_Default(t *testing.T) {
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)
	teatest.RequireEqualOutput(t, []byte(m.View()))
}

func TestThemePicker_View_SecondSelected(t *testing.T) {
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)
	m.cursor = 1
	SetTheme(Themes[1])
	t.Cleanup(func() { SetTheme(TokyoNight) })
	teatest.RequireEqualOutput(t, []byte(m.View()))
}

func TestThemePicker_NavigateDown(t *testing.T) {
	withTheme(t, TokyoNight)
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)

	m, _ = m.Update(sendKey('j'))
	if m.cursor != 1 {
		t.Fatalf("expected cursor=1 after j, got %d", m.cursor)
	}
	if ActiveTheme.Name != Themes[1].Name {
		t.Fatalf("expected theme %q, got %q", Themes[1].Name, ActiveTheme.Name)
	}
}

func TestThemePicker_NavigateUpClamp(t *testing.T) {
	withTheme(t, TokyoNight)
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)

	m, _ = m.Update(sendKey('k'))
	if m.cursor != 0 {
		t.Fatalf("expected cursor=0 (clamped), got %d", m.cursor)
	}
}

func TestThemePicker_NavigateDownClamp(t *testing.T) {
	withTheme(t, TokyoNight)
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)

	// Move past end.
	for i := 0; i < len(Themes)+2; i++ {
		m, _ = m.Update(sendKey('j'))
	}
	if m.cursor != len(Themes)-1 {
		t.Fatalf("expected cursor=%d (clamped), got %d", len(Themes)-1, m.cursor)
	}
}

func TestThemePicker_LivePreview(t *testing.T) {
	withTheme(t, TokyoNight)
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)

	if ActiveTheme.Name != TokyoNight.Name {
		t.Fatalf("expected initial theme %q", TokyoNight.Name)
	}

	m, _ = m.Update(sendKey('j'))
	if ActiveTheme.Name != Themes[1].Name {
		t.Fatalf("expected preview theme %q, got %q", Themes[1].Name, ActiveTheme.Name)
	}
}

func TestThemePicker_EnterSaves(t *testing.T) {
	withTheme(t, TokyoNight)
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir) // redirect config.Save

	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)
	_, cmd := m.Update(sendSpecialKey(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("expected command from enter")
	}
	msg := cmd()
	if _, ok := msg.(themePickerDoneMsg); !ok {
		t.Fatalf("expected themePickerDoneMsg, got %T", msg)
	}
	if cfg.Defaults.Theme != Themes[0].Name {
		t.Fatalf("expected config theme %q, got %q", Themes[0].Name, cfg.Defaults.Theme)
	}
}

func TestThemePicker_EscReverts(t *testing.T) {
	withTheme(t, TokyoNight)
	cfg := &config.Config{}
	m := NewThemePickerModel(cfg)

	// Move to second theme.
	m, _ = m.Update(sendKey('j'))
	if ActiveTheme.Name != Themes[1].Name {
		t.Fatalf("expected preview to change theme")
	}

	// Esc should revert.
	_, cmd := m.Update(sendSpecialKey(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("expected command from esc")
	}
	msg := cmd()
	if _, ok := msg.(themePickerCancelMsg); !ok {
		t.Fatalf("expected themePickerCancelMsg, got %T", msg)
	}
	if ActiveTheme.Name != TokyoNight.Name {
		t.Fatalf("expected theme reverted to %q, got %q", TokyoNight.Name, ActiveTheme.Name)
	}
}

func TestThemePicker_BothThemesRender(t *testing.T) {
	var outputs []string
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			withTheme(t, theme)
			cfg := &config.Config{}
			m := NewThemePickerModel(cfg)
			view := m.View()
			if !strings.Contains(view, "Theme") {
				t.Error("expected 'Theme' in view")
			}
			outputs = append(outputs, view)
		})
	}
	// With Ascii color profile, outputs should be identical structurally,
	// but we verify both render without panic.
}
