package app

import "github.com/charmbracelet/lipgloss"

var (
	// Status dot colors.
	statusRunning  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))  // green
	statusStopped  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // yellow
	statusFinished = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray/dim

	// Row styles.
	selectedRow = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	normalRow   = lipgloss.NewStyle()

	// Header.
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("7"))

	// Status bar at the bottom.
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	// Help text within the status bar.
	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	// Empty state.
	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)

	// Error / info flash messages.
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	// Filter bar styles.
	filterBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	filterLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("6")).
				Bold(true)

	filterActiveIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Italic(true)

	// Overlay (help, confirmation dialogs).
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(1, 3)
)

// statusDot returns a colored dot for the given session status.
func statusDot(status string) string {
	switch status {
	case "running":
		return statusRunning.Render("●")
	case "stopped":
		return statusStopped.Render("●")
	default:
		return statusFinished.Render("●")
	}
}
