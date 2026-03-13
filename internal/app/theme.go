package app

import "github.com/charmbracelet/lipgloss"

// Theme defines a complete color palette for the TUI.
type Theme struct {
	Name        string
	Accent      lipgloss.CompleteColor
	Green       lipgloss.CompleteColor
	Yellow      lipgloss.CompleteColor
	Red         lipgloss.CompleteColor
	Fg          lipgloss.CompleteColor
	FgMuted     lipgloss.CompleteColor
	FgDim       lipgloss.CompleteColor
	FgFaint     lipgloss.CompleteColor
	BgHighlight lipgloss.CompleteColor
	BgSurface   lipgloss.CompleteColor
	BgAlt       lipgloss.CompleteColor
}

// TokyoNight is the default theme.
var TokyoNight = Theme{
	Name:        "tokyonight",
	Accent:      lipgloss.CompleteColor{TrueColor: "#7aa2f7", ANSI256: "111", ANSI: "12"},
	Green:       lipgloss.CompleteColor{TrueColor: "#9ece6a", ANSI256: "150", ANSI: "10"},
	Yellow:      lipgloss.CompleteColor{TrueColor: "#e0af68", ANSI256: "180", ANSI: "11"},
	Red:         lipgloss.CompleteColor{TrueColor: "#f7768e", ANSI256: "210", ANSI: "9"},
	Fg:          lipgloss.CompleteColor{TrueColor: "#c0caf5", ANSI256: "189", ANSI: "15"},
	FgMuted:     lipgloss.CompleteColor{TrueColor: "#a9b1d6", ANSI256: "146", ANSI: "7"},
	FgDim:       lipgloss.CompleteColor{TrueColor: "#565f89", ANSI256: "60", ANSI: "8"},
	FgFaint:     lipgloss.CompleteColor{TrueColor: "#414868", ANSI256: "59", ANSI: "8"},
	BgHighlight: lipgloss.CompleteColor{TrueColor: "#292e42", ANSI256: "236", ANSI: "0"},
	BgSurface:   lipgloss.CompleteColor{TrueColor: "#24283b", ANSI256: "235", ANSI: "0"},
	BgAlt:       lipgloss.CompleteColor{TrueColor: "#16161e", ANSI256: "233", ANSI: "0"},
}

// CatppuccinMacchiato is an alternative theme.
var CatppuccinMacchiato = Theme{
	Name:        "catppuccin-macchiato",
	Accent:      lipgloss.CompleteColor{TrueColor: "#8aadf4", ANSI256: "111", ANSI: "12"},
	Green:       lipgloss.CompleteColor{TrueColor: "#a6da95", ANSI256: "150", ANSI: "10"},
	Yellow:      lipgloss.CompleteColor{TrueColor: "#eed49f", ANSI256: "186", ANSI: "11"},
	Red:         lipgloss.CompleteColor{TrueColor: "#ed8796", ANSI256: "211", ANSI: "9"},
	Fg:          lipgloss.CompleteColor{TrueColor: "#cad3f5", ANSI256: "189", ANSI: "15"},
	FgMuted:     lipgloss.CompleteColor{TrueColor: "#b8c0e0", ANSI256: "146", ANSI: "7"},
	FgDim:       lipgloss.CompleteColor{TrueColor: "#6e738d", ANSI256: "60", ANSI: "8"},
	FgFaint:     lipgloss.CompleteColor{TrueColor: "#494d64", ANSI256: "59", ANSI: "8"},
	BgHighlight: lipgloss.CompleteColor{TrueColor: "#363a4f", ANSI256: "237", ANSI: "0"},
	BgSurface:   lipgloss.CompleteColor{TrueColor: "#1e2030", ANSI256: "234", ANSI: "0"},
	BgAlt:       lipgloss.CompleteColor{TrueColor: "#181926", ANSI256: "233", ANSI: "0"},
}

// Themes is the list of all available themes.
var Themes = []Theme{TokyoNight, CatppuccinMacchiato}

// ActiveTheme is the currently active theme.
var ActiveTheme = TokyoNight

// ThemeByName returns the theme with the given name, defaulting to TokyoNight.
func ThemeByName(name string) Theme {
	for _, t := range Themes {
		if t.Name == name {
			return t
		}
	}
	return TokyoNight
}

// SetTheme sets the active theme and rebuilds all styles.
func SetTheme(t Theme) {
	ActiveTheme = t
	RebuildStyles()
}
