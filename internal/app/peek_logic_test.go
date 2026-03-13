package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPeekModel_WindowSizeMsg(t *testing.T) {
	m := makeTestPeekModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Fatalf("expected 100x50, got %dx%d", m.width, m.height)
	}
}

func TestPeekModel_ContentMsg(t *testing.T) {
	m := makeTestPeekModel()
	m, _ = m.Update(peekContentMsg("new content here"))
	view := m.View()
	if !strings.Contains(view, "new content here") {
		t.Error("expected updated content in view")
	}
}

func TestPeekModel_ErrMsg(t *testing.T) {
	m := makeTestPeekModel()
	m, _ = m.Update(peekErrMsg{err: &testError{"capture failed"}})
	view := m.View()
	if !strings.Contains(view, "capture failed") {
		t.Error("expected error message in viewport")
	}
}

func TestPeekModel_ScrollUpDown(t *testing.T) {
	m := makeTestPeekModel()
	// Just verify no panic on scroll keys.
	m, _ = m.Update(sendKey('k'))
	m, _ = m.Update(sendKey('j'))
}
