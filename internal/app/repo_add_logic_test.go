package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRepoAddModel_WindowSizeMsg(t *testing.T) {
	m := makeTestRepoAddModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	if m.width != 100 || m.height != 50 {
		t.Fatalf("expected 100x50, got %dx%d", m.width, m.height)
	}
}

func TestRepoAddModel_AIToolStep_Enter(t *testing.T) {
	m := makeTestRepoAddModel()
	m.step = repoStepAITool
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepSetupCommands {
		t.Fatalf("expected repoStepSetupCommands, got %d", m.step)
	}
}

func TestRepoAddModel_AIToolStep_Typing(t *testing.T) {
	m := makeTestRepoAddModel()
	m.step = repoStepAITool
	m.aiToolInput.Focus()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if m.step != repoStepAITool {
		t.Fatalf("expected to stay on repoStepAITool, got %d", m.step)
	}
}

func TestRepoAddModel_SetupCommandsStep_Enter(t *testing.T) {
	m := makeTestRepoAddModel()
	m.step = repoStepSetupCommands
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepFinalConfirm {
		t.Fatalf("expected repoStepFinalConfirm, got %d", m.step)
	}
}

func TestRepoAddModel_SetupCommandsStep_Typing(t *testing.T) {
	m := makeTestRepoAddModel()
	m.step = repoStepSetupCommands
	m.setupCommandsInput.Focus()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if m.step != repoStepSetupCommands {
		t.Fatalf("expected to stay on repoStepSetupCommands, got %d", m.step)
	}
}

func TestRepoAddModel_SetupCommandsWithValue(t *testing.T) {
	m := makeTestRepoAddModel()
	m.step = repoStepSetupCommands
	m.setupCommandsInput.SetValue("npm install, npm run build")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if len(m.setupCommands) != 2 {
		t.Fatalf("expected 2 setup commands, got %d", len(m.setupCommands))
	}
}
