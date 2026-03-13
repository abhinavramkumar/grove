package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/store"
)

func TestAppView_DefaultList(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	teatest.RequireEqualOutput(t, []byte(stabilizeView(app.View())))
}

func TestAppView_HelpOverlay(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewHelp
	teatest.RequireEqualOutput(t, []byte(app.View()))
}

func TestAppView_PruneConfirm_Clean(t *testing.T) {
	sessions := makeTestSessions()
	app := makeTestAppModel(t, sessions)
	app.view = viewPruneConfirm
	wtPath := "/home/user/repos/myapp/worktrees/feat-auth"
	app.prune = pruneConfirm{
		session: &store.Session{
			ID: "1", Name: "feat-auth", Status: "stopped",
			Worktree: &wtPath,
		},
		dirty:   false,
		repoDir: "/home/user/repos/myapp",
	}
	teatest.RequireEqualOutput(t, []byte(app.View()))
}

func TestAppView_PruneConfirm_Dirty(t *testing.T) {
	sessions := makeTestSessions()
	app := makeTestAppModel(t, sessions)
	app.view = viewPruneConfirm
	wtPath := "/home/user/repos/myapp/worktrees/feat-auth"
	app.prune = pruneConfirm{
		session: &store.Session{
			ID: "1", Name: "feat-auth", Status: "finished",
			Worktree: &wtPath,
		},
		dirty:   true,
		repoDir: "/home/user/repos/myapp",
	}
	teatest.RequireEqualOutput(t, []byte(app.View()))
}

func TestAppView_ThemePickerOverlay(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewThemePicker
	app.themePicker = NewThemePickerModel(app.config)
	teatest.RequireEqualOutput(t, []byte(app.View()))
}

// --- Status bar tests ---

func TestAppView_StatusBar_Running(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	// Cursor on first session (running).
	app.list.Cursor = 0
	view := stabilizeView(app.View())
	for _, substr := range []string{"running", "attach", "peek", "stop", "new", "help", "quit"} {
		if !strings.Contains(view, substr) {
			t.Errorf("expected %q in status bar", substr)
		}
	}
}

func TestAppView_StatusBar_Stopped(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.Cursor = 1 // stopped session
	view := stabilizeView(app.View())
	for _, substr := range []string{"stopped", "resume", "delete", "new", "help", "quit"} {
		if !strings.Contains(view, substr) {
			t.Errorf("expected %q in status bar", substr)
		}
	}
}

func TestAppView_StatusBar_Finished(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.Cursor = 2 // finished session
	view := stabilizeView(app.View())
	for _, substr := range []string{"delete", "prune", "new", "help", "quit"} {
		if !strings.Contains(view, substr) {
			t.Errorf("expected %q in status bar", substr)
		}
	}
}

func TestAppView_StatusBar_Flash(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.flash = "session created"
	view := stabilizeView(app.View())
	if !strings.Contains(view, "session created") {
		t.Error("expected flash message in status bar")
	}
}

func TestAppView_StatusBar_FilterActive(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.FilterText = "auth"
	view := stabilizeView(app.View())
	if !strings.Contains(view, "filter:") || !strings.Contains(view, "auth") {
		t.Error("expected filter text in status bar")
	}
}

// --- Behavior tests ---

func TestAppView_HelpOverlay_CloseOnEsc(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewHelp
	m, _ := app.Update(sendSpecialKey(tea.KeyEscape))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList, got %d", result.view)
	}
}

func TestAppView_HelpOverlay_CloseOnQuestion(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewHelp
	m, _ := app.Update(sendKey('?'))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList, got %d", result.view)
	}
}

func TestAppView_PruneConfirm_YConfirms(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPruneConfirm
	wtPath := "/tmp/wt"
	app.prune = pruneConfirm{
		session: &store.Session{ID: "1", Name: "test", Worktree: &wtPath, Status: "stopped"},
		dirty:   false,
		repoDir: "/tmp/repo",
	}
	m, _ := app.Update(sendKey('y'))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after y, got %d", result.view)
	}
}

func TestAppView_PruneConfirm_NReturns(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPruneConfirm
	wtPath := "/tmp/wt"
	app.prune = pruneConfirm{
		session: &store.Session{ID: "1", Name: "test", Worktree: &wtPath, Status: "stopped"},
		dirty:   false,
		repoDir: "/tmp/repo",
	}
	m, _ := app.Update(sendKey('n'))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after n, got %d", result.view)
	}
	if !strings.Contains(result.flash, "cancelled") {
		t.Fatalf("expected 'cancelled' in flash, got %q", result.flash)
	}
}

func TestAppView_NavigateToHelp(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(sendKey('?'))
	result := m.(AppModel)
	if result.view != viewHelp {
		t.Fatalf("expected viewHelp, got %d", result.view)
	}
}

func TestAppView_NavigateToThemePicker(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(sendKey('t'))
	result := m.(AppModel)
	if result.view != viewThemePicker {
		t.Fatalf("expected viewThemePicker, got %d", result.view)
	}
}

func TestAppView_QuitReturns(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	_, cmd := app.Update(sendKey('q'))
	if cmd == nil {
		t.Fatal("expected quit command, got nil")
	}
	// tea.Quit returns a special quit message.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestAppView_CursorJK(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())

	// j moves down
	m, _ := app.Update(sendKey('j'))
	result := m.(AppModel)
	if result.list.Cursor != 1 {
		t.Fatalf("expected cursor=1 after j, got %d", result.list.Cursor)
	}

	// j again
	m, _ = result.Update(sendKey('j'))
	result = m.(AppModel)
	if result.list.Cursor != 2 {
		t.Fatalf("expected cursor=2 after j, got %d", result.list.Cursor)
	}

	// k moves up
	m, _ = result.Update(sendKey('k'))
	result = m.(AppModel)
	if result.list.Cursor != 1 {
		t.Fatalf("expected cursor=1 after k, got %d", result.list.Cursor)
	}
}
