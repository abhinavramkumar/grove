package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/teatest"

	"github.com/abhinav/grove/internal/config"
)

func TestRepoAddView_ConfirmRootStep(t *testing.T) {
	m := makeTestRepoAddModel()
	m.width = 80
	m.height = 24

	view := m.View()
	if !strings.Contains(view, "Step 1") || !strings.Contains(view, "/tmp/test-repo") {
		t.Error("expected Step 1 and repo root in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestRepoAddView_WorktreeBaseStep(t *testing.T) {
	m := makeTestRepoAddModel()
	m.width = 80
	m.height = 24
	m.step = repoStepWorktreeBase

	view := m.View()
	if !strings.Contains(view, "Step 2") {
		t.Error("expected Step 2 in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestRepoAddView_AIToolStep(t *testing.T) {
	m := makeTestRepoAddModel()
	m.width = 80
	m.height = 24
	m.step = repoStepAITool

	view := m.View()
	if !strings.Contains(view, "Step 3") {
		t.Error("expected Step 3 in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestRepoAddView_SetupCommandsStep(t *testing.T) {
	m := makeTestRepoAddModel()
	m.width = 80
	m.height = 24
	m.step = repoStepSetupCommands

	view := m.View()
	if !strings.Contains(view, "Step 4") {
		t.Error("expected Step 4 in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestRepoAddView_FinalConfirmStep(t *testing.T) {
	m := makeTestRepoAddModel()
	m.width = 80
	m.height = 24
	m.step = repoStepFinalConfirm

	view := m.View()
	if !strings.Contains(view, "Step 5") || !strings.Contains(view, "/tmp/test-repo") {
		t.Error("expected Step 5 and repo root in summary")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}

func TestRepoAddView_ProgressBarAdvances(t *testing.T) {
	steps := []struct {
		step repoAddStep
		name string
	}{
		{repoStepConfirmRoot, "Root"},
		{repoStepWorktreeBase, "Worktree"},
		{repoStepAITool, "AI_Tool"},
		{repoStepSetupCommands, "Setup"},
		{repoStepFinalConfirm, "Save"},
	}

	for _, tc := range steps {
		t.Run(tc.name, func(t *testing.T) {
			m := makeTestRepoAddModel()
			m.width = 80
			m.height = 24
			m.step = tc.step
			teatest.RequireEqualOutput(t, []byte(m.View()))
		})
	}
}

func TestRepoAddView_ErrorDisplay(t *testing.T) {
	existing := config.RepoConfig{RepoRoot: "/tmp/test-repo"}
	m := makeTestRepoAddModel(existing)
	m.width = 80
	m.height = 24
	m.err = "already registered"

	view := m.View()
	if !strings.Contains(view, "already registered") {
		t.Error("expected error message in view")
	}
	teatest.RequireEqualOutput(t, []byte(view))
}
