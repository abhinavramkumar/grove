package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/store"
)

// --- handleKey coverage ---

func TestHandleKey_Attach_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil) // no sessions
	_, cmd := app.Update(sendSpecialKey(tea.KeyEnter))
	if cmd != nil {
		t.Fatal("expected nil cmd when no session selected")
	}
}

func TestHandleKey_Delete_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil)
	_, cmd := app.Update(sendKey('d'))
	if cmd != nil {
		t.Fatal("expected nil cmd when no session selected")
	}
}

func TestHandleKey_Stop_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil)
	_, cmd := app.Update(sendKey('s'))
	if cmd != nil {
		t.Fatal("expected nil cmd when no session selected")
	}
}

func TestHandleKey_Resume_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil)
	_, cmd := app.Update(sendKey('r'))
	if cmd != nil {
		t.Fatal("expected nil cmd when no session selected")
	}
}

func TestHandleKey_Peek_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil)
	m, _ := app.Update(sendKey('p'))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatal("expected to stay on viewList")
	}
}

func TestHandleKey_Peek_NotRunning(t *testing.T) {
	sessions := []*store.Session{
		{ID: "1", Name: "stopped", Status: "stopped", Tool: "claude", Directory: "/tmp"},
	}
	app := makeTestAppModel(t, sessions)
	m, _ := app.Update(sendKey('p'))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatal("expected to stay on viewList for non-running session")
	}
	if !strings.Contains(result.flash, "running") {
		t.Fatalf("expected flash about running, got %q", result.flash)
	}
}

func TestHandleKey_Prune_NoSession(t *testing.T) {
	app := makeTestAppModel(t, nil)
	_, cmd := app.Update(sendKey('x'))
	if cmd != nil {
		t.Fatal("expected nil cmd when no session")
	}
}

func TestHandleKey_Prune_NoWorktree(t *testing.T) {
	sessions := []*store.Session{
		{ID: "1", Name: "test", Status: "finished", Tool: "claude", Directory: "/tmp"},
	}
	app := makeTestAppModel(t, sessions)
	_, cmd := app.Update(sendKey('x'))
	if cmd == nil {
		t.Fatal("expected info command")
	}
	msg := cmd()
	if info, ok := msg.(infoMsg); !ok || !strings.Contains(string(info), "worktree") {
		t.Fatalf("expected info about no worktree, got %T: %v", msg, msg)
	}
}

func TestHandleKey_New(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(sendKey('n'))
	result := m.(AppModel)
	if result.view != viewNew {
		t.Fatalf("expected viewNew, got %d", result.view)
	}
}

func TestHandleKey_Filter(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	result := m.(AppModel)
	if !result.list.Filtering {
		t.Fatal("expected Filtering=true")
	}
}

func TestHandleKey_Escape(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(sendSpecialKey(tea.KeyEscape))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatal("escape in list should be no-op")
	}
}

// --- Filter key interception ---

func TestFilterKey_EscClears(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.StartFilter()
	app.list.FilterText = "test"

	m, _ := app.Update(sendSpecialKey(tea.KeyEscape))
	result := m.(AppModel)
	if result.list.Filtering {
		t.Fatal("expected Filtering=false after esc")
	}
	if result.list.FilterText != "" {
		t.Fatalf("expected empty FilterText, got %q", result.list.FilterText)
	}
}

func TestFilterKey_EnterCommits(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.StartFilter()
	app.list.filterInput.SetValue("auth")

	m, _ := app.Update(sendSpecialKey(tea.KeyEnter))
	result := m.(AppModel)
	if result.list.Filtering {
		t.Fatal("expected Filtering=false after enter")
	}
	if result.list.FilterText != "auth" {
		t.Fatalf("expected FilterText='auth', got %q", result.list.FilterText)
	}
}

func TestFilterKey_Typing(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.list.StartFilter()

	m, _ := app.Update(sendKey('a'))
	result := m.(AppModel)
	if !result.list.Filtering {
		t.Fatal("expected still Filtering")
	}
}

// --- WindowSizeMsg ---

func TestWindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := m.(AppModel)
	if result.width != 120 || result.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", result.width, result.height)
	}
	if result.list.Width != 120 || result.list.Height != 38 {
		t.Fatalf("expected list 120x38, got %dx%d", result.list.Width, result.list.Height)
	}
}

// --- sessionsMsg ---

func TestSessionsMsg(t *testing.T) {
	app := makeTestAppModel(t, nil)
	newSessions := makeTestSessions()
	m, _ := app.Update(sessionsMsg(newSessions))
	result := m.(AppModel)
	if len(result.list.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result.list.Sessions))
	}
}

// --- errMsg / infoMsg ---

func TestErrMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(errMsg{err: &testError{"something failed"}})
	result := m.(AppModel)
	if !strings.Contains(result.flash, "something failed") {
		t.Fatalf("expected flash with error, got %q", result.flash)
	}
}

func TestInfoMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	m, _ := app.Update(infoMsg("done"))
	result := m.(AppModel)
	if !strings.Contains(result.flash, "done") {
		t.Fatalf("expected flash with info, got %q", result.flash)
	}
}

// --- KeyMsg clears flash ---

func TestKeyMsg_ClearsFlash(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.flash = "old message"
	m, _ := app.Update(sendKey('j'))
	result := m.(AppModel)
	if result.flash != "" {
		t.Fatalf("expected flash cleared, got %q", result.flash)
	}
}

// --- updateHelp message routing ---

func TestUpdateHelp_WindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewHelp
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	result := m.(AppModel)
	if result.width != 100 {
		t.Fatalf("expected width=100, got %d", result.width)
	}
}

func TestUpdateHelp_SessionsMsg(t *testing.T) {
	app := makeTestAppModel(t, nil)
	app.view = viewHelp
	m, _ := app.Update(sessionsMsg(makeTestSessions()))
	result := m.(AppModel)
	if len(result.list.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result.list.Sessions))
	}
}

// --- updatePruneConfirm message routing ---

func TestUpdatePruneConfirm_WindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPruneConfirm
	wtPath := "/tmp/wt"
	app.prune = pruneConfirm{
		session: &store.Session{ID: "1", Name: "test", Worktree: &wtPath, Status: "stopped"},
	}
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	result := m.(AppModel)
	if result.width != 100 {
		t.Fatalf("expected width=100, got %d", result.width)
	}
}

func TestUpdatePruneConfirm_EscCancels(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPruneConfirm
	wtPath := "/tmp/wt"
	app.prune = pruneConfirm{
		session: &store.Session{ID: "1", Name: "test", Worktree: &wtPath, Status: "stopped"},
	}
	m, _ := app.Update(sendSpecialKey(tea.KeyEscape))
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after esc, got %d", result.view)
	}
}

// --- updateThemePicker message routing ---

func TestUpdateThemePicker_WindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewThemePicker
	app.themePicker = NewThemePickerModel(app.config)
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	result := m.(AppModel)
	if result.width != 100 {
		t.Fatalf("expected width=100, got %d", result.width)
	}
}

func TestUpdateThemePicker_DoneMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewThemePicker
	m, _ := app.Update(themePickerDoneMsg{})
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after done, got %d", result.view)
	}
}

func TestUpdateThemePicker_CancelMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewThemePicker
	m, _ := app.Update(themePickerCancelMsg{})
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after cancel, got %d", result.view)
	}
}

func TestUpdateThemePicker_SessionsMsg(t *testing.T) {
	app := makeTestAppModel(t, nil)
	app.view = viewThemePicker
	app.themePicker = NewThemePickerModel(app.config)
	m, _ := app.Update(sessionsMsg(makeTestSessions()))
	result := m.(AppModel)
	if len(result.list.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result.list.Sessions))
	}
}

// --- updateCreate message routing ---

func TestUpdateCreate_WindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewNew
	app.create = NewCreateModel(app.config, app.manager)
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	result := m.(AppModel)
	if result.width != 100 {
		t.Fatalf("expected width=100, got %d", result.width)
	}
}

func TestUpdateCreate_CancelMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewNew
	m, _ := app.Update(createCancelMsg{})
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after cancel, got %d", result.view)
	}
}

func TestUpdateCreate_DoneMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewNew
	m, _ := app.Update(createDoneMsg{session: &store.Session{Name: "new-session"}})
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after done, got %d", result.view)
	}
	if !strings.Contains(result.flash, "new-session") {
		t.Fatalf("expected flash with session name, got %q", result.flash)
	}
}

func TestUpdateCreate_ErrMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewNew
	app.create = NewCreateModel(app.config, app.manager)
	m, _ := app.Update(createErrMsg{err: &testError{"creation failed"}})
	result := m.(AppModel)
	if result.create.err == "" {
		t.Fatal("expected create error to be set")
	}
}

func TestUpdateCreate_SessionsMsg(t *testing.T) {
	app := makeTestAppModel(t, nil)
	app.view = viewNew
	app.create = NewCreateModel(app.config, app.manager)
	m, _ := app.Update(sessionsMsg(makeTestSessions()))
	result := m.(AppModel)
	if len(result.list.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result.list.Sessions))
	}
}

// --- updatePeek message routing ---

func TestUpdatePeek_BackMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPeek
	m, _ := app.Update(peekBackMsg{})
	result := m.(AppModel)
	if result.view != viewList {
		t.Fatalf("expected viewList after back, got %d", result.view)
	}
}

func TestUpdatePeek_WindowSizeMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	app.view = viewPeek
	app.peek = makeTestPeekModel()
	m, _ := app.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	result := m.(AppModel)
	if result.width != 100 {
		t.Fatalf("expected width=100, got %d", result.width)
	}
}

func TestUpdatePeek_SessionsMsg(t *testing.T) {
	app := makeTestAppModel(t, nil)
	app.view = viewPeek
	app.peek = makeTestPeekModel()
	m, _ := app.Update(sessionsMsg(makeTestSessions()))
	result := m.(AppModel)
	if len(result.list.Sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(result.list.Sessions))
	}
}

// --- attachDoneMsg ---

func TestAttachDoneMsg(t *testing.T) {
	app := makeTestAppModel(t, makeTestSessions())
	_, cmd := app.Update(attachDoneMsg{})
	// Should trigger reconcileAndLoad.
	if cmd == nil {
		t.Fatal("expected command after attach done")
	}
}

// --- pruneReadyMsg ---

func TestPruneReadyMsg(t *testing.T) {
	sessions := makeTestSessions()
	app := makeTestAppModel(t, sessions)
	wtPath := "/tmp/wt"
	m, _ := app.Update(pruneReadyMsg{
		session: &store.Session{ID: "1", Name: "test", Worktree: &wtPath, Status: "stopped"},
		dirty:   true,
		repoDir: "/tmp/repo",
	})
	result := m.(AppModel)
	if result.view != viewPruneConfirm {
		t.Fatalf("expected viewPruneConfirm, got %d", result.view)
	}
	if !result.prune.dirty {
		t.Fatal("expected prune.dirty=true")
	}
}

// testError is a simple error type for tests.
type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
