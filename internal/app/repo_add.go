package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

// repoAddDoneMsg signals that a repo was added successfully.
type repoAddDoneMsg struct{ repo config.RepoConfig }

// repoAddCancelMsg signals that the user cancelled the wizard.
type repoAddCancelMsg struct{}

// repoAddStep enumerates wizard steps.
type repoAddStep int

const (
	repoStepConfirmRoot repoAddStep = iota
	repoStepWorktreeBase
	repoStepAITool
	repoStepSetupCommands
	repoStepFinalConfirm
)

// repoAddStepNames are the step labels for the progress bar.
var repoAddStepNames = []string{"Root", "Worktree", "AI Tool", "Setup", "Save"}

// repoAddStepIndex maps repoAddStep to progress bar index.
func repoAddStepIndex(step repoAddStep) int {
	switch step {
	case repoStepConfirmRoot:
		return 0
	case repoStepWorktreeBase:
		return 1
	case repoStepAITool:
		return 2
	case repoStepSetupCommands:
		return 3
	case repoStepFinalConfirm:
		return 4
	}
	return 0
}

// RepoAddModel is the interactive wizard for adding a repo to the config.
type RepoAddModel struct {
	step repoAddStep

	// The detected/provided repo root.
	repoRoot string

	// Text inputs for each step.
	worktreeBaseInput  textinput.Model
	aiToolInput        textinput.Model
	setupCommandsInput textinput.Model

	// Resolved values.
	worktreeBase  string
	aiTool        string
	setupCommands []string

	// Config reference for duplicate checking and saving.
	config *config.Config

	// Display.
	width  int
	height int
	err    string
}

// NewRepoAddModel creates a new repo-add wizard.
func NewRepoAddModel(cfg *config.Config, repoRoot string) RepoAddModel {
	defaultWT := filepath.Join(filepath.Dir(repoRoot), filepath.Base(repoRoot)+"-worktrees")

	wti := textinput.New()
	wti.Placeholder = defaultWT
	wti.CharLimit = 256
	wti.Width = 60

	ati := textinput.New()
	ati.Placeholder = "leave empty to inherit global default"
	ati.CharLimit = 64
	ati.Width = 60

	sci := textinput.New()
	sci.Placeholder = "e.g. npm install, make build (comma-separated)"
	sci.CharLimit = 512
	sci.Width = 60

	return RepoAddModel{
		step:               repoStepConfirmRoot,
		repoRoot:           repoRoot,
		worktreeBaseInput:  wti,
		aiToolInput:        ati,
		setupCommandsInput: sci,
		config:             cfg,
	}
}

// Init returns no command.
func (m RepoAddModel) Init() tea.Cmd {
	return nil
}

// Update handles input for the repo-add wizard.
func (m RepoAddModel) Update(msg tea.Msg) (RepoAddModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m.updateInputs(msg)
}

func (m RepoAddModel) updateInputs(msg tea.Msg) (RepoAddModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.step {
	case repoStepWorktreeBase:
		m.worktreeBaseInput, cmd = m.worktreeBaseInput.Update(msg)
	case repoStepAITool:
		m.aiToolInput, cmd = m.aiToolInput.Update(msg)
	case repoStepSetupCommands:
		m.setupCommandsInput, cmd = m.setupCommandsInput.Update(msg)
	}
	return m, cmd
}

func (m RepoAddModel) handleKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if key.Matches(msg, keys.Escape) {
		return m, func() tea.Msg { return repoAddCancelMsg{} }
	}

	switch m.step {
	case repoStepConfirmRoot:
		return m.handleConfirmRootKey(msg)
	case repoStepWorktreeBase:
		return m.handleWorktreeBaseKey(msg)
	case repoStepAITool:
		return m.handleAIToolKey(msg)
	case repoStepSetupCommands:
		return m.handleSetupCommandsKey(msg)
	case repoStepFinalConfirm:
		return m.handleFinalConfirmKey(msg)
	}
	return m, nil
}

// --- Step: confirm repo root ---

func (m RepoAddModel) handleConfirmRootKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		// Check for duplicate.
		if m.config.RepoFor(m.repoRoot) != nil {
			m.err = fmt.Sprintf("repo %q is already registered", m.repoRoot)
			return m, nil
		}
		m.step = repoStepWorktreeBase
		m.worktreeBaseInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}
	return m, nil
}

// --- Step: worktree base ---

func (m RepoAddModel) handleWorktreeBaseKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		val := strings.TrimSpace(m.worktreeBaseInput.Value())
		if val == "" {
			// Use the placeholder default.
			val = m.worktreeBaseInput.Placeholder
		} else {
			val = expandHome(val)
			abs, err := filepath.Abs(val)
			if err != nil {
				m.err = fmt.Sprintf("invalid path: %v", err)
				return m, nil
			}
			val = abs
		}
		m.worktreeBase = val
		m.worktreeBaseInput.Blur()
		m.step = repoStepAITool
		m.aiToolInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.worktreeBaseInput, cmd = m.worktreeBaseInput.Update(msg)
	return m, cmd
}

// --- Step: AI tool override ---

func (m RepoAddModel) handleAIToolKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.aiTool = strings.TrimSpace(m.aiToolInput.Value())
		m.aiToolInput.Blur()
		m.step = repoStepSetupCommands
		m.setupCommandsInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	m.aiToolInput, cmd = m.aiToolInput.Update(msg)
	return m, cmd
}

// --- Step: setup commands ---

func (m RepoAddModel) handleSetupCommandsKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		val := strings.TrimSpace(m.setupCommandsInput.Value())
		if val != "" {
			parts := strings.Split(val, ",")
			for _, p := range parts {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					m.setupCommands = append(m.setupCommands, trimmed)
				}
			}
		}
		m.setupCommandsInput.Blur()
		m.step = repoStepFinalConfirm
		m.err = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.setupCommandsInput, cmd = m.setupCommandsInput.Update(msg)
	return m, cmd
}

// --- Step: final confirm ---

func (m RepoAddModel) handleFinalConfirmKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		repo := config.RepoConfig{
			RepoRoot:      m.repoRoot,
			WorktreeBase:  m.worktreeBase,
			AITool:        m.aiTool,
			SetupCommands: m.setupCommands,
		}
		return m, func() tea.Msg { return repoAddDoneMsg{repo: repo} }
	}
	return m, nil
}

// View renders the wizard.
func (m RepoAddModel) View() string {
	var b strings.Builder

	title := S.WizardTitle.Render("Add Repository")
	b.WriteString(title)
	b.WriteString("\n")

	// Step progress bar.
	b.WriteString("  " + renderStepProgress(repoAddStepIndex(m.step), repoAddStepNames))
	b.WriteString("\n\n")

	switch m.step {
	case repoStepConfirmRoot:
		b.WriteString(S.WizardLabel.Render("Step 1: Confirm repository root"))
		b.WriteString("\n\n")
		b.WriteString("  " + S.WizardChoice.Render(m.repoRoot))
		b.WriteString("\n\n  " + S.Dim.Render("enter to confirm, esc to cancel"))

	case repoStepWorktreeBase:
		b.WriteString(S.WizardLabel.Render("Step 2: Worktree base directory"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.worktreeBaseInput.View())
		b.WriteString("\n\n  " + S.Dim.Render("enter to accept (empty = default shown above), esc to cancel"))

	case repoStepAITool:
		b.WriteString(S.WizardLabel.Render("Step 3: AI tool override (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.aiToolInput.View())
		b.WriteString("\n\n  " + S.Dim.Render("enter to continue (empty = inherit global default)"))

	case repoStepSetupCommands:
		b.WriteString(S.WizardLabel.Render("Step 4: Setup commands (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.setupCommandsInput.View())
		b.WriteString("\n\n  " + S.Dim.Render("comma-separated, enter to continue (empty = inherit global)"))

	case repoStepFinalConfirm:
		b.WriteString(S.WizardLabel.Render("Step 5: Confirm & Save"))
		b.WriteString("\n\n")
		b.WriteString("  " + S.WizardLabel.Render("Repo root:     ") + m.repoRoot + "\n")
		b.WriteString("  " + S.WizardLabel.Render("Worktree base: ") + m.worktreeBase + "\n")
		aiTool := m.aiTool
		if aiTool == "" {
			aiTool = S.Dim.Render("(global default)")
		}
		b.WriteString("  " + S.WizardLabel.Render("AI tool:       ") + aiTool + "\n")
		setupCmds := strings.Join(m.setupCommands, ", ")
		if setupCmds == "" {
			setupCmds = S.Dim.Render("(global default)")
		}
		b.WriteString("  " + S.WizardLabel.Render("Setup cmds:    ") + setupCmds + "\n")
		b.WriteString("\n  " + S.Dim.Render("enter to save, esc to cancel"))
	}

	if m.err != "" {
		b.WriteString("\n\n  " + S.Error.Render(m.err))
	}

	return b.String()
}

// --- Standalone TUI runner ---

// repoAddRunner wraps RepoAddModel as a top-level tea.Model so it can be
// run in its own tea.Program.
type repoAddRunner struct {
	inner  RepoAddModel
	result *config.RepoConfig
}

func (r repoAddRunner) Init() tea.Cmd {
	return r.inner.Init()
}

func (r repoAddRunner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case repoAddDoneMsg:
		repo := msg.(repoAddDoneMsg).repo
		r.result = &repo
		return r, tea.Quit
	case repoAddCancelMsg:
		r.result = nil
		return r, tea.Quit
	}

	var cmd tea.Cmd
	r.inner, cmd = r.inner.Update(msg)
	return r, cmd
}

func (r repoAddRunner) View() string {
	return r.inner.View()
}

// RunRepoAddTUI launches an interactive TUI wizard for adding a repo.
// Returns the created RepoConfig, or nil if the user cancelled.
func RunRepoAddTUI(cfg *config.Config, repoRoot string) (*config.RepoConfig, error) {
	model := repoAddRunner{
		inner: NewRepoAddModel(cfg, repoRoot),
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running repo add wizard: %w", err)
	}

	result := finalModel.(repoAddRunner).result
	if result == nil {
		// User cancelled.
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil, nil
	}

	return result, nil
}
