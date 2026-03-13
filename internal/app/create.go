package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/worktree"
)

// createDoneMsg signals that a new session was created successfully.
type createDoneMsg struct{ session *store.Session }

// createCancelMsg signals that the user cancelled the wizard.
type createCancelMsg struct{}

// createErrMsg carries an error from an async creation command.
type createErrMsg struct{ err error }

// createStep enumerates wizard steps.
type createStep int

const (
	stepDirSource createStep = iota
	stepDirInput
	stepTool
	stepPrompt
	stepConfirm
)

// dirSource tracks whether the user chose existing dir or worktree.
type dirSource int

const (
	dirExisting dirSource = iota
	dirWorktree
)

// CreateModel is the self-contained new-session wizard.
type CreateModel struct {
	step      createStep
	dirSource dirSource

	// Inputs
	dirInput    textinput.Model // path (existing) or branch name (worktree)
	promptInput textinput.Model

	// Tool selection
	toolNames    []string
	toolSelected int

	// Config references needed for creation
	config  *config.Config
	manager *session.Manager

	// Resolved values
	resolvedDir string // set after step validation
	worktreeBranch string

	// Display
	width  int
	height int
	err    string // error to show on current step
}

// NewCreateModel initialises the wizard.
func NewCreateModel(cfg *config.Config, mgr *session.Manager) CreateModel {
	di := textinput.New()
	di.Placeholder = "path or ~/..."
	di.CharLimit = 256
	di.Width = 60

	pi := textinput.New()
	pi.Placeholder = "prompt or plan file path (optional)"
	pi.CharLimit = 1024
	pi.Width = 60

	// Collect tool names sorted, with default first.
	toolSet := make(map[string]struct{})
	for name := range mgr.Adapters {
		toolSet[name] = struct{}{}
	}
	var names []string
	for n := range toolSet {
		names = append(names, n)
	}
	sort.Strings(names)

	// Move the default tool to front if set.
	defaultTool := cfg.Defaults.AITool
	selectedIdx := 0
	if defaultTool != "" {
		for i, n := range names {
			if n == defaultTool {
				selectedIdx = i
				break
			}
		}
	}

	return CreateModel{
		step:         stepDirSource,
		dirSource:    dirExisting,
		dirInput:     di,
		promptInput:  pi,
		toolNames:    names,
		toolSelected: selectedIdx,
		config:       cfg,
		manager:      mgr,
	}
}

// Init returns no command; the wizard is purely interactive.
func (m CreateModel) Init() tea.Cmd {
	return nil
}

// Update handles input for the wizard.
func (m CreateModel) Update(msg tea.Msg) (CreateModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case createErrMsg:
		m.err = msg.err.Error()
		return m, nil
	}

	// Pass through to active text input.
	return m.updateInputs(msg)
}

func (m CreateModel) updateInputs(msg tea.Msg) (CreateModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.step {
	case stepDirInput:
		m.dirInput, cmd = m.dirInput.Update(msg)
	case stepPrompt:
		m.promptInput, cmd = m.promptInput.Update(msg)
	}
	return m, cmd
}

func (m CreateModel) handleKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	// Global: esc cancels.
	if key.Matches(msg, keys.Escape) {
		return m, func() tea.Msg { return createCancelMsg{} }
	}

	switch m.step {
	case stepDirSource:
		return m.handleDirSourceKey(msg)
	case stepDirInput:
		return m.handleDirInputKey(msg)
	case stepTool:
		return m.handleToolKey(msg)
	case stepPrompt:
		return m.handlePromptKey(msg)
	case stepConfirm:
		return m.handleConfirmKey(msg)
	}
	return m, nil
}

// --- Step: directory source ---

func (m CreateModel) handleDirSourceKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	switch msg.String() {
	case "1", "e":
		m.dirSource = dirExisting
		m.dirInput.Placeholder = "directory path (e.g. ~/projects/myapp)"
		m.step = stepDirInput
		m.dirInput.Focus()
		m.err = ""
		return m, textinput.Blink
	case "2", "w":
		m.dirSource = dirWorktree
		m.dirInput.Placeholder = "branch name"
		m.step = stepDirInput
		m.dirInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}
	return m, nil
}

// --- Step: directory/branch input ---

func (m CreateModel) handleDirInputKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	if msg.String() == "enter" {
		val := strings.TrimSpace(m.dirInput.Value())
		if val == "" {
			m.err = "value cannot be empty"
			return m, nil
		}

		if m.dirSource == dirExisting {
			expanded := expandHome(val)
			abs, err := filepath.Abs(expanded)
			if err != nil {
				m.err = fmt.Sprintf("invalid path: %v", err)
				return m, nil
			}
			info, err := os.Stat(abs)
			if err != nil || !info.IsDir() {
				m.err = "directory does not exist"
				return m, nil
			}
			m.resolvedDir = abs
		} else {
			if !worktree.ValidateBranchName(val) {
				m.err = "invalid branch name"
				return m, nil
			}
			m.worktreeBranch = val
		}

		m.step = stepTool
		m.dirInput.Blur()
		m.err = ""
		return m, nil
	}

	// Pass through typing to the input.
	var cmd tea.Cmd
	m.dirInput, cmd = m.dirInput.Update(msg)
	return m, cmd
}

// --- Step: tool selection ---

func (m CreateModel) handleToolKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		if m.toolSelected > 0 {
			m.toolSelected--
		}
	case "right", "l":
		if m.toolSelected < len(m.toolNames)-1 {
			m.toolSelected++
		}
	case "enter":
		m.step = stepPrompt
		m.promptInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}
	return m, nil
}

// --- Step: prompt input ---

func (m CreateModel) handlePromptKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.step = stepConfirm
		m.promptInput.Blur()
		m.err = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.promptInput, cmd = m.promptInput.Update(msg)
	return m, cmd
}

// --- Step: confirm ---

func (m CreateModel) handleConfirmKey(msg tea.KeyMsg) (CreateModel, tea.Cmd) {
	if msg.String() == "enter" {
		return m, m.createSession()
	}
	return m, nil
}

// createSession performs async session creation.
func (m CreateModel) createSession() tea.Cmd {
	return func() tea.Msg {
		dir := m.resolvedDir
		var wtPtr *string

		// If worktree mode, create the worktree first.
		if m.dirSource == dirWorktree {
			repoDir, err := worktree.GetMainRepoPath(".")
			if err != nil {
				return createErrMsg{fmt.Errorf("finding repo: %w", err)}
			}
			basePath := m.config.Defaults.WorktreeBase
			if basePath == "" {
				basePath = filepath.Join(filepath.Dir(repoDir), "worktrees")
			}

			wtPath, err := worktree.Create(repoDir, basePath, m.worktreeBranch, "")
			if err != nil {
				return createErrMsg{fmt.Errorf("creating worktree: %w", err)}
			}

			if len(m.config.Worktree.SetupCommands) > 0 {
				if err := worktree.RunSetupCommands(wtPath, m.config.Worktree.SetupCommands); err != nil {
					return createErrMsg{fmt.Errorf("setup commands: %w", err)}
				}
			}

			dir = wtPath
			branch := m.worktreeBranch
			wtPtr = &branch
		}

		toolName := m.toolNames[m.toolSelected]
		prompt := strings.TrimSpace(m.promptInput.Value())
		planFile := ""

		// If the prompt looks like a file path, treat it as a plan file.
		if prompt != "" {
			expanded := expandHome(prompt)
			if info, err := os.Stat(expanded); err == nil && !info.IsDir() {
				planFile = expanded
				prompt = ""
			}
		}

		// Generate a session name from branch or directory basename.
		name := filepath.Base(dir)
		if m.dirSource == dirWorktree {
			name = worktree.SanitizeBranchName(m.worktreeBranch)
		}

		sess, err := m.manager.Create(name, toolName, dir, wtPtr, prompt, planFile)
		if err != nil {
			return createErrMsg{err}
		}
		return createDoneMsg{session: sess}
	}
}

// View renders the wizard.
func (m CreateModel) View() string {
	var b strings.Builder

	title := wizardTitleStyle.Render("New Session")
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.step {
	case stepDirSource:
		b.WriteString(wizardLabelStyle.Render("Step 1: Directory source"))
		b.WriteString("\n\n")
		b.WriteString("  " + wizardChoiceStyle.Render("[1]") + " Use existing directory\n")
		b.WriteString("  " + wizardChoiceStyle.Render("[2]") + " Create worktree\n")

	case stepDirInput:
		if m.dirSource == dirExisting {
			b.WriteString(wizardLabelStyle.Render("Step 1: Enter directory path"))
		} else {
			b.WriteString(wizardLabelStyle.Render("Step 1: Enter branch name"))
		}
		b.WriteString("\n\n")
		b.WriteString("  " + m.dirInput.View())

	case stepTool:
		b.WriteString(wizardLabelStyle.Render("Step 2: Select AI tool"))
		b.WriteString("\n\n  ")
		for i, name := range m.toolNames {
			if i == m.toolSelected {
				b.WriteString(wizardSelectedToolStyle.Render(" " + name + " "))
			} else {
				b.WriteString(wizardToolStyle.Render(" " + name + " "))
			}
			if i < len(m.toolNames)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n\n  " + dimStyle.Render("← → to select, enter to confirm"))

	case stepPrompt:
		b.WriteString(wizardLabelStyle.Render("Step 3: Prompt or plan file (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.promptInput.View())
		b.WriteString("\n\n  " + dimStyle.Render("enter to continue (leave empty for interactive)"))

	case stepConfirm:
		b.WriteString(wizardLabelStyle.Render("Step 4: Confirm"))
		b.WriteString("\n\n")
		if m.dirSource == dirExisting {
			b.WriteString("  Directory:  " + m.resolvedDir + "\n")
		} else {
			b.WriteString("  Worktree:   " + m.worktreeBranch + "\n")
		}
		b.WriteString("  Tool:       " + m.toolNames[m.toolSelected] + "\n")
		prompt := strings.TrimSpace(m.promptInput.Value())
		if prompt == "" {
			prompt = "(interactive)"
		}
		b.WriteString("  Prompt:     " + prompt + "\n")
		b.WriteString("\n  " + dimStyle.Render("enter to create, esc to cancel"))
	}

	if m.err != "" {
		b.WriteString("\n\n  " + errorStyle.Render(m.err))
	}

	return b.String()
}

// expandHome expands a leading ~/ to the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// Additional styles for the wizard.
var (
	wizardTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("5")).
				Padding(1, 2)

	wizardLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("7"))

	wizardChoiceStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	wizardSelectedToolStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("5")).
				Foreground(lipgloss.Color("15")).
				Bold(true)

	wizardToolStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
)
