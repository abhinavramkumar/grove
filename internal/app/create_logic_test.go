package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	got := expandHome("~/projects/test")
	want := filepath.Join(home, "projects/test")
	if got != want {
		t.Fatalf("expandHome(~/projects/test) = %q, want %q", got, want)
	}
}

func TestExpandHome_NoTilde(t *testing.T) {
	got := expandHome("/absolute/path")
	if got != "/absolute/path" {
		t.Fatalf("expected unchanged path, got %q", got)
	}
}

func TestCreateModel_DirSourceExisting(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)

	// Press "1" for existing directory.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.step != stepDirInput {
		t.Fatalf("expected stepDirInput, got %d", m.step)
	}
	if m.dirSource != dirExisting {
		t.Fatalf("expected dirExisting, got %d", m.dirSource)
	}
}

func TestCreateModel_DirInputExisting_ValidDir(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepDirInput
	m.dirSource = dirExisting

	// Set a valid directory path.
	dir := t.TempDir()
	m.dirInput.SetValue(dir)

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepTool {
		t.Fatalf("expected stepTool after valid dir, got %d", m.step)
	}
	if m.resolvedDir != dir {
		t.Fatalf("expected resolvedDir=%q, got %q", dir, m.resolvedDir)
	}
}

func TestCreateModel_DirInputExisting_EmptyValue(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepDirInput
	m.dirSource = dirExisting
	m.dirInput.SetValue("")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepDirInput {
		t.Fatalf("expected to stay on stepDirInput, got %d", m.step)
	}
	if m.err == "" {
		t.Fatal("expected error for empty value")
	}
}

func TestCreateModel_DirInputExisting_NonexistentDir(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepDirInput
	m.dirSource = dirExisting
	m.dirInput.SetValue("/nonexistent/path/that/does/not/exist")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepDirInput {
		t.Fatalf("expected to stay on stepDirInput, got %d", m.step)
	}
	if !strings.Contains(m.err, "does not exist") {
		t.Fatalf("expected 'does not exist' error, got %q", m.err)
	}
}

func TestCreateModel_ToolSelectNavigation(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepTool

	initial := m.toolSelected
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.toolSelected != initial+1 {
		t.Fatalf("expected toolSelected=%d after right, got %d", initial+1, m.toolSelected)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if m.toolSelected != initial {
		t.Fatalf("expected toolSelected=%d after left, got %d", initial, m.toolSelected)
	}
}

func TestCreateModel_ToolSelectConfirm(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepTool

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepPrompt {
		t.Fatalf("expected stepPrompt after tool confirm, got %d", m.step)
	}
}

func TestCreateModel_PromptStep_Enter(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepPrompt

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepConfirm {
		t.Fatalf("expected stepConfirm after prompt enter, got %d", m.step)
	}
}

func TestCreateModel_PromptStep_Typing(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepPrompt
	m.promptInput.Focus()

	// Type a character — should stay on prompt step.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if m.step != stepPrompt {
		t.Fatalf("expected to stay on stepPrompt, got %d", m.step)
	}
}

func TestCreateModel_EscCancels(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)

	for _, step := range []createStep{stepDirSource, stepDirInput, stepTool, stepPrompt, stepConfirm} {
		m.step = step
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
		if cmd == nil {
			t.Fatalf("step %d: expected command from esc", step)
		}
		msg := cmd()
		if _, ok := msg.(createCancelMsg); !ok {
			t.Fatalf("step %d: expected createCancelMsg, got %T", step, msg)
		}
	}
}

func TestCreateModel_DirInputWorktree_InvalidBranch(t *testing.T) {
	repos := []config.RepoConfig{{RepoRoot: "/projects/myrepo"}}
	cfg, mgr := newTestCreateDeps(t, repos)
	m := NewCreateModel(cfg, mgr)
	m.step = stepDirInput
	m.dirSource = dirWorktree
	m.selectedRepo = &repos[0]
	m.dirInput.SetValue("..invalid..branch")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != stepDirInput {
		t.Fatalf("expected to stay on stepDirInput, got %d", m.step)
	}
	if !strings.Contains(m.err, "invalid branch") {
		t.Fatalf("expected 'invalid branch' error, got %q", m.err)
	}
}

func TestCreateModel_WindowSizeMsg(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if m.width != 100 || m.height != 30 {
		t.Fatalf("expected 100x30, got %dx%d", m.width, m.height)
	}
}

func TestCreateModel_CreateErrMsg(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m, _ = m.Update(createErrMsg{err: os.ErrNotExist})
	if m.err == "" {
		t.Fatal("expected error to be set")
	}
}

func TestCreateModel_ConfirmStep_Enter(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepConfirm
	m.dirSource = dirExisting
	m.resolvedDir = t.TempDir()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	// Enter on confirm triggers createSession command.
	if cmd == nil {
		t.Fatal("expected command from confirm enter")
	}
}

func TestCreateModel_UpdateInputs_DirStep(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepDirInput
	m.dirSource = dirExisting
	m.dirInput.Focus()
	// Send a blink message (non-key) to trigger updateInputs.
	m, _ = m.Update(m.dirInput.Cursor.BlinkCmd()())
	// Should not crash or change step.
	if m.step != stepDirInput {
		t.Fatalf("expected stepDirInput, got %d", m.step)
	}
}

func TestCreateModel_UpdateInputs_PromptStep(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t, nil)
	m := NewCreateModel(cfg, mgr)
	m.step = stepPrompt
	m.promptInput.Focus()
	m, _ = m.Update(m.promptInput.Cursor.BlinkCmd()())
	if m.step != stepPrompt {
		t.Fatalf("expected stepPrompt, got %d", m.step)
	}
}
