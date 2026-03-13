package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/store"
)

func makeTestPeekModel() PeekModel {
	sess := &store.Session{
		ID:           "1",
		Name:         "test-session",
		TmuxSession:  "grove_test",
		Status:       "running",
	}
	vp := viewport.New(80, 22)
	vp.SetContent("hello world\nline 2\nline 3")
	return PeekModel{
		session:  sess,
		viewport: vp,
		width:    80,
		height:   24,
		ready:    true,
	}
}

func TestPeekView_Layout(t *testing.T) {
	m := makeTestPeekModel()
	view := m.View()
	if !strings.Contains(view, "Peek: test-session (grove_test)") {
		t.Error("expected header with session name and tmux session")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestPeekView_HelpText(t *testing.T) {
	m := makeTestPeekModel()
	view := m.View()
	for _, substr := range []string{"esc", "back", "enter", "attach", "scroll"} {
		if !strings.Contains(view, substr) {
			t.Errorf("expected %q in peek help text", substr)
		}
	}
}

func TestPeekView_Content(t *testing.T) {
	m := makeTestPeekModel()
	view := m.View()
	if !strings.Contains(view, "hello world") {
		t.Error("expected viewport content in view")
	}
}

func TestPeekView_EscReturns(t *testing.T) {
	m := makeTestPeekModel()
	_, cmd := m.Update(sendSpecialKey(tea.KeyEscape))
	if cmd == nil {
		t.Fatal("expected command from esc")
	}
	msg := cmd()
	if _, ok := msg.(peekBackMsg); !ok {
		t.Fatalf("expected peekBackMsg, got %T", msg)
	}
}

func TestPeekView_EnterAttaches(t *testing.T) {
	m := makeTestPeekModel()
	_, cmd := m.Update(sendSpecialKey(tea.KeyEnter))
	if cmd == nil {
		t.Fatal("expected command from enter")
	}
	msg := cmd()
	attach, ok := msg.(peekAttachMsg)
	if !ok {
		t.Fatalf("expected peekAttachMsg, got %T", msg)
	}
	if attach.sessionID != "1" {
		t.Fatalf("expected sessionID '1', got %q", attach.sessionID)
	}
}
