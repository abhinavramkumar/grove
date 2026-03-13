package app

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles used throughout the TUI.
type Styles struct {
	// Status dot styles.
	StatusRunning  lipgloss.Style
	StatusStopped  lipgloss.Style
	StatusFinished lipgloss.Style

	// Row styles.
	SelectedRow lipgloss.Style
	NormalRow   lipgloss.Style
	FinishedRow lipgloss.Style

	// Header bar.
	Header lipgloss.Style

	// Status bar.
	StatusBar lipgloss.Style

	// Help text.
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Empty state.
	Empty lipgloss.Style

	// Flash messages.
	Error lipgloss.Style
	Info  lipgloss.Style

	// Filter bar.
	FilterBar       lipgloss.Style
	FilterLabel     lipgloss.Style
	FilterActive    lipgloss.Style

	// Overlay (help, confirmation dialogs).
	Overlay lipgloss.Style

	// Wizard styles.
	WizardTitle        lipgloss.Style
	WizardLabel        lipgloss.Style
	WizardChoice       lipgloss.Style
	WizardSelectedTool lipgloss.Style
	WizardTool         lipgloss.Style
	Dim                lipgloss.Style

	// Peek styles.
	PeekHeader lipgloss.Style

	// Accent border.
	AccentBorder lipgloss.Style

	// Repo list styles.
	RepoListHeader lipgloss.Style
	RepoListCell   lipgloss.Style
	RepoListDim    lipgloss.Style
	RepoListTitle  lipgloss.Style
	RepoListBorder lipgloss.Style

	// Contextual status label.
	ContextualStatusLabel lipgloss.Style
}

// S is the package-level current styles instance.
var S Styles

// RebuildStyles rebuilds S from ActiveTheme.
func RebuildStyles() {
	S = BuildStyles(ActiveTheme)
}

// BuildStyles constructs all styles from the given theme.
func BuildStyles(t Theme) Styles {
	return Styles{
		// Status dots.
		StatusRunning:  lipgloss.NewStyle().Foreground(t.Green),
		StatusStopped:  lipgloss.NewStyle().Foreground(t.Yellow),
		StatusFinished: lipgloss.NewStyle().Foreground(t.FgDim),

		// Row styles.
		SelectedRow: lipgloss.NewStyle().Background(t.BgHighlight),
		NormalRow:   lipgloss.NewStyle(),
		FinishedRow: lipgloss.NewStyle().Foreground(t.FgDim),

		// Header.
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.FgMuted),

		// Status bar.
		StatusBar: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Background(t.BgSurface).
			Padding(0, 1),

		// Help text.
		HelpKey:  lipgloss.NewStyle().Foreground(t.FgMuted),
		HelpDesc: lipgloss.NewStyle().Foreground(t.FgDim),

		// Empty state.
		Empty: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Italic(true),

		// Flash messages.
		Error: lipgloss.NewStyle().Foreground(t.Red),
		Info:  lipgloss.NewStyle().Foreground(t.Accent),

		// Filter bar.
		FilterBar: lipgloss.NewStyle().
			Foreground(t.Fg).
			Background(t.BgSurface).
			Padding(0, 1),
		FilterLabel: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true),
		FilterActive: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Italic(true),

		// Overlay.
		Overlay: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.FgDim).
			Padding(1, 3),

		// Wizard.
		WizardTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Accent).
			Padding(1, 2),
		WizardLabel: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.FgMuted),
		WizardChoice: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true),
		WizardSelectedTool: lipgloss.NewStyle().
			Background(t.Accent).
			Foreground(t.BgAlt).
			Bold(true),
		WizardTool: lipgloss.NewStyle().
			Foreground(t.FgMuted),
		Dim: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Italic(true),

		// Peek.
		PeekHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Fg).
			Background(t.Accent).
			Padding(0, 1),

		// Accent border.
		AccentBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Accent),

		// Repo list.
		RepoListHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Accent).
			Underline(true),
		RepoListCell: lipgloss.NewStyle().
			Foreground(t.Fg),
		RepoListDim: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Italic(true),
		RepoListTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Accent).
			Padding(0, 0, 1, 0),
		RepoListBorder: lipgloss.NewStyle().
			Foreground(t.FgFaint),

		// Contextual status label.
		ContextualStatusLabel: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Italic(true),
	}
}

// statusDot returns a colored dot for the given session status.
func statusDot(status string) string {
	switch status {
	case "running":
		return S.StatusRunning.Render("●")
	case "stopped":
		return S.StatusStopped.Render("●")
	default:
		return S.StatusFinished.Render("●")
	}
}

func init() {
	RebuildStyles()
}
