package app

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/tools"
)

func TestMain(m *testing.M) {
	lipgloss.SetColorProfile(termenv.Ascii)
	SetTheme(TokyoNight)
	os.Exit(m.Run())
}

// withTheme sets the active theme for the duration of a test.
func withTheme(t *testing.T, theme Theme) {
	t.Helper()
	original := ActiveTheme
	SetTheme(theme)
	t.Cleanup(func() { SetTheme(original) })
}

// sendKey builds a tea.KeyMsg for a rune key press.
func sendKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

// sendSpecialKey builds a tea.KeyMsg for a special key.
func sendSpecialKey(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// makeTestListModel creates a ListModel with standard dimensions.
func makeTestListModel(sessions []*store.Session) ListModel {
	return ListModel{
		Sessions: sessions,
		Width:    80,
		Height:   20,
	}
}

// makeTestSessions creates 3 deterministic test sessions.
func makeTestSessions() []*store.Session {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return []*store.Session{
		{
			ID: "1", Name: "feat-auth", Tool: "claude",
			Directory: "/home/user/repos/myapp/worktrees/feat-auth",
			RepoRoot:  strPtr("/home/user/repos/myapp"),
			Status:    "running", CreatedAt: base,
		},
		{
			ID: "2", Name: "fix-bug", Tool: "copilot",
			Directory: "/home/user/repos/backend/worktrees/fix-bug",
			RepoRoot:  strPtr("/home/user/repos/backend"),
			Status:    "stopped", CreatedAt: base,
		},
		{
			ID: "3", Name: "refactor", Tool: "claude",
			Directory: "/home/user/repos/frontend/worktrees/refactor",
			RepoRoot:  nil,
			Status:    "finished", CreatedAt: base,
		},
	}
}

// makeTestAppModel creates an AppModel with real Store, stub Manager, and pre-loaded sessions.
func makeTestAppModel(t *testing.T, sessions []*store.Session) AppModel {
	t.Helper()
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude"},
		Tools: map[string]config.ToolConfig{
			"claude": {Command: "claude"},
			"codex":  {Command: "codex"},
		},
	}

	mgr := &session.Manager{
		Store: s,
		Adapters: map[string]tools.Adapter{
			"claude": stubAdapter{name: "claude"},
			"codex":  stubAdapter{name: "codex"},
		},
	}

	app := New(s, cfg, mgr)
	app.width = 80
	app.height = 24
	app.list.Width = 80
	app.list.Height = 22
	app.list.Sessions = sessions
	return app
}

// stabilizeView replaces dynamic content (age) with fixed strings for golden file determinism.
var ageRegexp = regexp.MustCompile(`\d+[dhm]\s+\d+[dhm]|\d+[dhm]|<1m`)

func stabilizeView(view string) string {
	return ageRegexp.ReplaceAllString(view, "AGE")
}
