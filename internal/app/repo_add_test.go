package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

func makeTestRepoAddModel(repos ...config.RepoConfig) RepoAddModel {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "/tmp/worktrees",
		},
		Repos: repos,
	}
	return NewRepoAddModel(cfg, "/tmp/test-repo")
}

func TestRepoAddModel_StepProgression(t *testing.T) {
	m := makeTestRepoAddModel()

	// Step 0: confirm root — view should mention the repo root.
	view := m.View()
	if !strings.Contains(view, "/tmp/test-repo") {
		t.Fatalf("step 0: expected repo root in view, got: %s", view)
	}
	if !strings.Contains(view, "Step 1") {
		t.Fatalf("step 0: expected Step 1 label, got: %s", view)
	}

	// Press enter to confirm root.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepWorktreeBase {
		t.Fatalf("expected step %d after confirm root, got %d", repoStepWorktreeBase, m.step)
	}
	view = m.View()
	if !strings.Contains(view, "Step 2") {
		t.Fatalf("step 1: expected Step 2 label, got: %s", view)
	}

	// Press enter to accept default worktree base.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepAITool {
		t.Fatalf("expected step %d after worktree base, got %d", repoStepAITool, m.step)
	}
	view = m.View()
	if !strings.Contains(view, "Step 3") {
		t.Fatalf("step 2: expected Step 3 label, got: %s", view)
	}

	// Press enter to skip AI tool.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepSetupCommands {
		t.Fatalf("expected step %d after ai tool, got %d", repoStepSetupCommands, m.step)
	}
	view = m.View()
	if !strings.Contains(view, "Step 4") {
		t.Fatalf("step 3: expected Step 4 label, got: %s", view)
	}

	// Press enter to skip setup commands.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepFinalConfirm {
		t.Fatalf("expected step %d after setup commands, got %d", repoStepFinalConfirm, m.step)
	}
	view = m.View()
	if !strings.Contains(view, "Step 5") {
		t.Fatalf("step 4: expected Step 5 label, got: %s", view)
	}
	if !strings.Contains(view, "/tmp/test-repo") {
		t.Fatalf("step 4: expected repo root in summary, got: %s", view)
	}
}

func TestRepoAddModel_EscCancels(t *testing.T) {
	steps := []repoAddStep{
		repoStepConfirmRoot,
		repoStepWorktreeBase,
		repoStepAITool,
		repoStepSetupCommands,
		repoStepFinalConfirm,
	}

	for _, step := range steps {
		m := makeTestRepoAddModel()
		m.step = step

		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
		if cmd == nil {
			t.Fatalf("step %d: expected a command from esc, got nil", step)
		}

		msg := cmd()
		if _, ok := msg.(repoAddCancelMsg); !ok {
			t.Fatalf("step %d: expected repoAddCancelMsg, got %T", step, msg)
		}
	}
}

func TestRepoAddModel_DuplicateRepoErrors(t *testing.T) {
	existing := config.RepoConfig{RepoRoot: "/tmp/test-repo"}
	m := makeTestRepoAddModel(existing)

	// Press enter to confirm root — should show error.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.step != repoStepConfirmRoot {
		t.Fatalf("expected to stay on step %d due to duplicate, got %d", repoStepConfirmRoot, m.step)
	}
	if !strings.Contains(m.err, "already registered") {
		t.Fatalf("expected duplicate error, got: %q", m.err)
	}
	view := m.View()
	if !strings.Contains(view, "already registered") {
		t.Fatalf("expected error in view, got: %s", view)
	}
}

func TestRepoAddModel_FinalConfirmProducesDoneMsg(t *testing.T) {
	m := makeTestRepoAddModel()

	// Advance through all steps.
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // confirm root
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // worktree base (default)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // ai tool (empty)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // setup commands (empty)

	if m.step != repoStepFinalConfirm {
		t.Fatalf("expected final confirm step, got %d", m.step)
	}

	// Press enter to save.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected a command from final confirm, got nil")
	}

	msg := cmd()
	done, ok := msg.(repoAddDoneMsg)
	if !ok {
		t.Fatalf("expected repoAddDoneMsg, got %T", msg)
	}
	if done.repo.RepoRoot != "/tmp/test-repo" {
		t.Fatalf("expected repo root /tmp/test-repo, got %s", done.repo.RepoRoot)
	}
}
