package app

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/store"
)

func TestListView_HeaderBar(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	teatest.RequireEqualOutput(t, []byte(stabilizeView(m.View())))
}

func TestListView_TableHeader(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	view := m.View()
	for _, col := range []string{"NAME", "REPO", "TOOL", "DIRECTORY", "AGE"} {
		if !strings.Contains(view, col) {
			t.Errorf("expected column header %q in view", col)
		}
	}
}

func TestListView_SessionRows(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	view := m.View()
	for _, substr := range []string{"feat-auth", "myapp", "claude", "fix-bug", "backend", "copilot", "refactor"} {
		if !strings.Contains(view, substr) {
			t.Errorf("expected %q in view", substr)
		}
	}
}

func TestListView_SelectedRow(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.Cursor = 0
	teatest.RequireEqualOutput(t, []byte(stabilizeView(m.View())))
}

func TestListView_CursorMoved(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.Cursor = 1
	teatest.RequireEqualOutput(t, []byte(stabilizeView(m.View())))
}

func TestListView_EmptyState_NoSessions(t *testing.T) {
	m := makeTestListModel(nil)
	teatest.RequireEqualOutput(t, []byte(m.View()))
}

func TestListView_EmptyState_FilterNoMatch(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.FilterText = "zzz"
	teatest.RequireEqualOutput(t, []byte(m.View()))
}

func TestListView_TruncateLongName(t *testing.T) {
	sessions := []*store.Session{
		{
			ID: "1", Name: "this-is-a-very-long-session-name-that-exceeds-width",
			Tool: "claude", Directory: "/tmp/test",
			Status: "running", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	m := ListModel{Sessions: sessions, Width: 60, Height: 20}
	teatest.RequireEqualOutput(t, []byte(stabilizeView(m.View())))
}

func TestListView_TruncateLeftDir(t *testing.T) {
	sessions := []*store.Session{
		{
			ID: "1", Name: "test",
			Tool:      "claude",
			Directory: "/very/long/directory/path/that/should/be/truncated/from/the/left/side/to/fit/in/column",
			Status:    "running", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	m := ListModel{Sessions: sessions, Width: 80, Height: 20}
	teatest.RequireEqualOutput(t, []byte(stabilizeView(m.View())))
}

func TestListView_BothThemes(t *testing.T) {
	for _, theme := range Themes {
		t.Run(theme.Name, func(t *testing.T) {
			withTheme(t, theme)
			m := makeTestListModel(makeTestSessions())
			view := stabilizeView(m.View())
			if !strings.Contains(view, "grove") {
				t.Error("expected 'grove' in view")
			}
			teatest.RequireEqualOutput(t, []byte(view))
		})
	}
}

func TestListView_MoveDown(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.MoveDown()
	if m.Cursor != 1 {
		t.Fatalf("expected cursor=1, got %d", m.Cursor)
	}
	m.MoveDown()
	if m.Cursor != 2 {
		t.Fatalf("expected cursor=2, got %d", m.Cursor)
	}
	// Should clamp at end.
	m.MoveDown()
	if m.Cursor != 2 {
		t.Fatalf("expected cursor=2 (clamped), got %d", m.Cursor)
	}
}

func TestListView_MoveUpClamp(t *testing.T) {
	m := makeTestListModel(makeTestSessions())
	m.MoveUp()
	if m.Cursor != 0 {
		t.Fatalf("expected cursor=0 (clamped), got %d", m.Cursor)
	}
}
