# TUI Visual Overhaul Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign Grove's TUI with a Tokyo Night / Catppuccin Macchiato theme system, contextual status bar, numbered wizard progress, grouped help overlay, and live theme picker.

**Architecture:** Introduce a `Theme` struct and `Styles` struct that decouples color definitions from style construction. All UI components reference `Styles` built from the active theme. A theme picker overlay lets users switch and persist themes via config.

**Tech Stack:** Go, charmbracelet/bubbletea, charmbracelet/bubbles, charmbracelet/lipgloss v1.1.0 (CompleteColor, layout functions)

**Spec:** `docs/superpowers/specs/2026-03-14-tui-visual-overhaul-design.md`

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/app/theme.go` | **New.** Theme struct, TokyoNight + CatppuccinMacchiato vars, Themes registry, ActiveTheme, SetTheme helper |
| `internal/app/styles.go` | **Rewrite.** Styles struct with all lipgloss styles, BuildStyles(Theme) constructor, package-level `S` var |
| `internal/app/theme_picker.go` | **New.** ThemePickerModel — overlay with cursor navigation, live preview, confirm/cancel |
| `internal/app/list.go` | **Modify.** Header bar, zebra striping, left accent border, status summary, empty state |
| `internal/app/app.go` | **Modify.** Contextual status bar, help overlay grouping, viewThemePicker routing, theme state |
| `internal/app/create.go` | **Modify.** Numbered step progress bar, themed inputs/pills, confirm summary |
| `internal/app/peek.go` | **Minor.** Update colors to theme references |
| `internal/app/repo_add.go` | **Minor.** Match create wizard visual pattern |
| `internal/app/repo_list.go` | **Minor.** Migrate hardcoded styles to theme references |
| `internal/app/keys.go` | **Modify.** Add `t` keybinding, update help descriptions |
| `internal/config/config.go` | **Modify.** Add `Theme` field to DefaultsConfig |

---

## Prerequisite: Create Feature Branch

- [ ] **Create the feature branch before any work**

```bash
git checkout -b feature/tui-visual-overhaul
```

All commits in this plan go on this branch. Do NOT commit to `main`.

---

## Chunk 1: Theme System Foundation

### Task 1: Create theme.go with Theme struct and built-in themes

**Files:**
- Create: `internal/app/theme.go`

- [ ] **Step 1: Create theme.go with Theme struct and two built-in themes**

```go
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

// Built-in themes.
var TokyoNight = Theme{
	Name:        "tokyo-night",
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

// Themes is the ordered list of available themes.
var Themes = []Theme{
	TokyoNight,
	CatppuccinMacchiato,
}

// ActiveTheme is the currently active theme.
var ActiveTheme = TokyoNight

// ThemeByName returns the theme with the given name, or TokyoNight if not found.
func ThemeByName(name string) Theme {
	for _, t := range Themes {
		if t.Name == name {
			return t
		}
	}
	return TokyoNight
}

// SetTheme switches the active theme and rebuilds all styles.
func SetTheme(t Theme) {
	ActiveTheme = t
	RebuildStyles()
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build ./internal/app/`
Expected: Compile error about `RebuildStyles` not defined — that's expected, it comes in Task 2.

- [ ] **Step 3: Commit**

```bash
git add internal/app/theme.go
git commit -m "feat(ui): add Theme struct with Tokyo Night and Catppuccin Macchiato"
```

### Task 2: Rewrite styles.go with Styles struct and BuildStyles

**Files:**
- Rewrite: `internal/app/styles.go`

- [ ] **Step 1: Rewrite styles.go with Styles struct built from active theme**

Replace the entire file. The `Styles` struct contains every lipgloss style used across the app. `BuildStyles` constructs them from a `Theme`. A package-level `S` var holds the current styles, and `RebuildStyles()` refreshes it.

```go
package app

import "github.com/charmbracelet/lipgloss"

// Styles holds all lipgloss styles derived from the active theme.
type Styles struct {
	// Status dots.
	StatusRunning  lipgloss.Style
	StatusStopped  lipgloss.Style
	StatusFinished lipgloss.Style

	// Row styles.
	SelectedRow lipgloss.Style
	NormalRow   lipgloss.Style
	ZebraRow    lipgloss.Style
	FinishedRow lipgloss.Style

	// Header.
	HeaderStyle lipgloss.Style

	// Header bar (top of screen).
	HeaderBarStyle    lipgloss.Style
	HeaderAccent      lipgloss.Style
	HeaderPipe        lipgloss.Style
	HeaderDim         lipgloss.Style
	HeaderGreen       lipgloss.Style
	HeaderYellow      lipgloss.Style
	HeaderVersion     lipgloss.Style

	// Status bar.
	StatusBarStyle lipgloss.Style

	// Help text within status bar.
	HelpKeyStyle  lipgloss.Style
	HelpDescStyle lipgloss.Style
	HelpPipe      lipgloss.Style

	// Empty state.
	EmptyStyle    lipgloss.Style
	EmptyDim      lipgloss.Style
	EmptyPill     lipgloss.Style

	// Error / info flash.
	ErrorStyle lipgloss.Style
	InfoStyle  lipgloss.Style

	// Filter bar.
	FilterBarStyle        lipgloss.Style
	FilterLabelStyle      lipgloss.Style
	FilterActiveIndicator lipgloss.Style

	// Overlay (help, confirmation, theme picker).
	OverlayStyle      lipgloss.Style
	OverlayTitle      lipgloss.Style
	OverlaySeparator  lipgloss.Style
	OverlayCategory   lipgloss.Style
	OverlayKey        lipgloss.Style
	OverlayDesc       lipgloss.Style
	OverlayFooter     lipgloss.Style

	// Wizard styles.
	WizardTitleStyle        lipgloss.Style
	WizardLabelStyle        lipgloss.Style
	WizardChoiceStyle       lipgloss.Style
	WizardSelectedPill      lipgloss.Style
	WizardUnselectedPill    lipgloss.Style
	WizardDimStyle          lipgloss.Style
	WizardStepActive        lipgloss.Style
	WizardStepActiveLabel   lipgloss.Style
	WizardStepInactive      lipgloss.Style
	WizardStepInactiveLabel lipgloss.Style
	WizardStepConnector     lipgloss.Style

	// Peek header.
	PeekHeaderStyle lipgloss.Style

	// Contextual status bar.
	StatusLabel lipgloss.Style

	// Selected row accent border.
	AccentBorder lipgloss.Style

	// Repo list styles.
	RepoListHeaderStyle lipgloss.Style
	RepoListCellStyle   lipgloss.Style
	RepoListDimStyle    lipgloss.Style
	RepoListTitleStyle  lipgloss.Style
	RepoListBorderStyle lipgloss.Style
}

// S is the current styles instance, rebuilt whenever the theme changes.
var S Styles

// RebuildStyles reconstructs S from the ActiveTheme.
func RebuildStyles() {
	S = BuildStyles(ActiveTheme)
}

// BuildStyles creates a Styles from a Theme.
func BuildStyles(t Theme) Styles {
	return Styles{
		// Status dots.
		StatusRunning:  lipgloss.NewStyle().Foreground(t.Green),
		StatusStopped:  lipgloss.NewStyle().Foreground(t.Yellow),
		StatusFinished: lipgloss.NewStyle().Foreground(t.FgFaint),

		// Row styles.
		SelectedRow: lipgloss.NewStyle().Background(t.BgHighlight),
		NormalRow:   lipgloss.NewStyle(),
		ZebraRow:    lipgloss.NewStyle().Background(t.BgAlt),
		FinishedRow: lipgloss.NewStyle().Foreground(t.FgFaint),

		// Header.
		HeaderStyle: lipgloss.NewStyle().Bold(true).Foreground(t.FgDim),

		// Header bar.
		HeaderBarStyle: lipgloss.NewStyle().Background(t.BgSurface).Padding(0, 1),
		HeaderAccent:   lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		HeaderPipe:     lipgloss.NewStyle().Foreground(t.FgFaint),
		HeaderDim:      lipgloss.NewStyle().Foreground(t.FgDim),
		HeaderGreen:    lipgloss.NewStyle().Foreground(t.Green),
		HeaderYellow:   lipgloss.NewStyle().Foreground(t.Yellow),
		HeaderVersion:  lipgloss.NewStyle().Foreground(t.FgFaint),

		// Status bar.
		StatusBarStyle: lipgloss.NewStyle().
			Foreground(t.FgDim).
			Background(t.BgSurface).
			Padding(0, 1),

		// Help text.
		HelpKeyStyle:  lipgloss.NewStyle().Foreground(t.Accent),
		HelpDescStyle: lipgloss.NewStyle().Foreground(t.FgDim),
		HelpPipe:      lipgloss.NewStyle().Foreground(t.FgFaint),

		// Empty state.
		EmptyStyle: lipgloss.NewStyle().Foreground(t.FgDim),
		EmptyDim:   lipgloss.NewStyle().Foreground(t.FgFaint),
		EmptyPill: lipgloss.NewStyle().
			Background(t.BgHighlight).
			Foreground(t.FgDim).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.FgFaint).
			Padding(0, 1),

		// Error / info.
		ErrorStyle: lipgloss.NewStyle().Foreground(t.Red),
		InfoStyle:  lipgloss.NewStyle().Foreground(t.Accent),

		// Filter bar.
		FilterBarStyle: lipgloss.NewStyle().
			Foreground(t.Fg).
			Background(t.BgSurface).
			Padding(0, 1),
		FilterLabelStyle:      lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		FilterActiveIndicator: lipgloss.NewStyle().Foreground(t.FgDim).Italic(true),

		// Overlay.
		OverlayStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.FgFaint).
			Background(t.BgSurface).
			Padding(1, 3),
		OverlayTitle:     lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		OverlaySeparator: lipgloss.NewStyle().Foreground(t.FgFaint),
		OverlayCategory:  lipgloss.NewStyle().Foreground(t.FgDim).Bold(true),
		OverlayKey:       lipgloss.NewStyle().Foreground(t.Accent),
		OverlayDesc:      lipgloss.NewStyle().Foreground(t.FgMuted),
		OverlayFooter:    lipgloss.NewStyle().Foreground(t.FgFaint),

		// Wizard.
		WizardTitleStyle: lipgloss.NewStyle().Bold(true).Foreground(t.Accent).Padding(1, 2),
		WizardLabelStyle: lipgloss.NewStyle().Bold(true).Foreground(t.Fg),
		WizardChoiceStyle: lipgloss.NewStyle().Foreground(t.Accent).Bold(true),
		WizardSelectedPill: lipgloss.NewStyle().
			Background(t.Accent).
			Foreground(t.BgAlt).
			Bold(true).
			Padding(0, 1),
		WizardUnselectedPill: lipgloss.NewStyle().
			Background(t.BgHighlight).
			Foreground(t.FgMuted).
			Padding(0, 1),
		WizardDimStyle: lipgloss.NewStyle().Foreground(t.FgFaint).Italic(true),
		WizardStepActive: lipgloss.NewStyle().
			Background(t.Accent).
			Foreground(t.BgAlt).
			Bold(true),
		WizardStepActiveLabel:   lipgloss.NewStyle().Foreground(t.Accent),
		WizardStepInactive:      lipgloss.NewStyle().Foreground(t.FgFaint),
		WizardStepInactiveLabel: lipgloss.NewStyle().Foreground(t.FgFaint),
		WizardStepConnector:     lipgloss.NewStyle().Foreground(t.FgFaint),

		// Peek.
		PeekHeaderStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(t.Fg).
			Background(t.Accent).
			Padding(0, 1),

		// Contextual status label.
		StatusLabel: lipgloss.NewStyle().Italic(true),

		// Selected row accent border (rendered as left-border text).
		AccentBorder: lipgloss.NewStyle().Foreground(t.Accent),

		// Repo list.
		RepoListHeaderStyle: lipgloss.NewStyle().Bold(true).Foreground(t.FgDim).Padding(0, 1),
		RepoListCellStyle:   lipgloss.NewStyle().Padding(0, 1),
		RepoListDimStyle:    lipgloss.NewStyle().Foreground(t.FgDim).Italic(true).Padding(0, 1),
		RepoListTitleStyle:  lipgloss.NewStyle().Bold(true).Foreground(t.Accent).MarginBottom(1),
		RepoListBorderStyle: lipgloss.NewStyle().Foreground(t.FgFaint),
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

// init builds styles from the default theme on startup.
func init() {
	RebuildStyles()
}
```

- [ ] **Step 2: Verify compilation**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build ./internal/app/`
Expected: Compilation errors in other files referencing old style vars (`errorStyle`, `infoStyle`, `selectedRow`, etc.). This is expected — we'll fix them in subsequent tasks.

- [ ] **Step 3: Commit**

```bash
git add internal/app/styles.go internal/app/theme.go
git commit -m "feat(ui): add Styles struct with BuildStyles and theme-derived styles"
```

### Task 3: Add Theme field to config

**Files:**
- Modify: `internal/config/config.go`

- [ ] **Step 1: Add Theme field to DefaultsConfig**

In `internal/config/config.go`, add the `Theme` field:

```go
type DefaultsConfig struct {
	AITool       string `toml:"ai_tool"`
	WorktreeBase string `toml:"worktree_base"`
	Theme        string `toml:"theme,omitempty"`
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/config/`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/config/config.go
git commit -m "feat(config): add theme field to defaults"
```

### Task 4: Add theme keybinding

**Files:**
- Modify: `internal/app/keys.go`

- [ ] **Step 1: Add Theme keybinding to keyMap**

Add `Theme` field to the `keyMap` struct and initialize it in `keys`:

```go
type keyMap struct {
	Up     key.Binding
	Down   key.Binding
	Attach key.Binding
	Peek   key.Binding
	New    key.Binding
	Delete key.Binding
	Stop   key.Binding
	Resume key.Binding
	Prune  key.Binding
	Filter key.Binding
	Theme  key.Binding
	Help   key.Binding
	Quit   key.Binding
	Escape key.Binding
}
```

Add to `keys` var:

```go
Theme: key.NewBinding(
	key.WithKeys("t"),
	key.WithHelp("t", "theme"),
),
```

- [ ] **Step 2: Delete statusBarHelp() function**

The `statusBarHelp()` function at the bottom of `keys.go` (lines 78-94) references old style vars (`helpKeyStyle`, `helpDescStyle`) that no longer exist. It is fully replaced by `contextualStatusBar()` in Task 5. Delete the entire function.

- [ ] **Step 3: Verify compilation**

Run: `go build ./internal/app/`
Expected: Compile errors from other files referencing old style vars — expected.

- [ ] **Step 4: Commit**

```bash
git add internal/app/keys.go
git commit -m "feat(ui): add t keybinding, remove old statusBarHelp"
```

## Chunk 2: Migrate Existing Views to Theme System

### Task 5: Migrate app.go to use Styles struct

**Files:**
- Modify: `internal/app/app.go`

- [ ] **Step 1: Add viewThemePicker to view enum and import theme initialization**

Add `viewThemePicker` to the `view` const block after `viewPruneConfirm`:

```go
const (
	viewList view = iota
	viewPeek
	viewNew
	viewHelp
	viewPruneConfirm
	viewThemePicker
)
```

- [ ] **Step 2: Replace all old style references with S.xxx**

Throughout `app.go`, replace:
- `errorStyle` → `S.ErrorStyle`
- `infoStyle` → `S.InfoStyle`
- `overlayStyle` → `S.OverlayStyle`
- `statusBarStyle` → `S.StatusBarStyle`
- `filterBarStyle` → `S.FilterBarStyle`
- `filterLabelStyle` → `S.FilterLabelStyle`
- `filterActiveIndicator` → `S.FilterActiveIndicator`
- `helpKeyStyle` → `S.HelpKeyStyle`
- `helpDescStyle` → `S.HelpDescStyle`

- [ ] **Step 3: Rewrite View() with header bar and contextual status bar**

Replace the `View()` method's status bar logic with the contextual version. The header bar is rendered in `list.go` (Task 6), but the status bar is assembled in `app.go`:

```go
func (m AppModel) View() string {
	switch m.view {
	case viewNew:
		return m.create.View()
	case viewPeek:
		return m.peek.View()
	case viewHelp:
		return m.viewHelpOverlay()
	case viewPruneConfirm:
		return m.viewPruneConfirmOverlay()
	case viewThemePicker:
		return m.viewThemePickerOverlay()
	}

	body := m.list.View()

	var bar string
	if m.list.Filtering {
		bar = S.FilterBarStyle.Width(m.list.Width).Render(
			S.FilterLabelStyle.Render("filter: ") + m.list.FilterInputView())
	} else if m.list.FilterText != "" {
		bar = S.StatusBarStyle.Width(m.list.Width).Render(
			S.FilterActiveIndicator.Render("filter: "+m.list.FilterText) + "  " + m.contextualStatusBar())
	} else if m.flash != "" {
		bar = S.StatusBarStyle.Width(m.list.Width).Render(m.flash)
	} else {
		bar = S.StatusBarStyle.Width(m.list.Width).Render(m.contextualStatusBar())
	}

	return body + "\n" + bar
}
```

- [ ] **Step 4: Implement contextualStatusBar()**

```go
func (m AppModel) contextualStatusBar() string {
	sess := m.list.Selected()

	var left, middle string

	if sess != nil {
		// Left: status label.
		switch sess.Status {
		case "running":
			left = S.StatusLabel.Foreground(ActiveTheme.Green).Render("running")
		case "stopped":
			left = S.StatusLabel.Foreground(ActiveTheme.Yellow).Render("stopped")
		default:
			left = S.StatusLabel.Foreground(ActiveTheme.FgDim).Render(sess.Status)
		}

		pipe := S.HelpPipe.Render(" │ ")

		// Middle: session-specific keys.
		switch sess.Status {
		case "running":
			middle = pipe +
				S.HelpKeyStyle.Render("enter") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("attach") + "  " +
				S.HelpKeyStyle.Render("p") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("peek") + "  " +
				S.HelpKeyStyle.Render("s") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("stop")
		case "stopped":
			middle = pipe +
				S.HelpKeyStyle.Render("r") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("resume") + "  " +
				S.HelpKeyStyle.Render("d") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("delete")
		default:
			middle = pipe +
				S.HelpKeyStyle.Render("d") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("delete") + "  " +
				S.HelpKeyStyle.Render("x") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("prune")
		}
	}

	// Right: always-available keys.
	right := S.HelpPipe.Render(" │ ") +
		S.HelpKeyStyle.Render("n") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("new") + "  " +
		S.HelpKeyStyle.Render("?") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("help") + "  " +
		S.HelpKeyStyle.Render("q") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("quit")

	return left + middle + right
}
```

- [ ] **Step 5: Rewrite viewHelpOverlay() with grouped keybindings**

```go
func (m AppModel) viewHelpOverlay() string {
	var b strings.Builder

	b.WriteString(S.OverlayTitle.Render("Keybindings"))
	b.WriteByte('\n')
	b.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", 32)))
	b.WriteByte('\n')

	groups := []struct {
		name     string
		bindings []struct{ key, desc string }
	}{
		{"NAVIGATION", []struct{ key, desc string }{
			{"j/k", "down / up"},
			{"enter", "attach to session"},
			{"p", "peek (live preview)"},
		}},
		{"ACTIONS", []struct{ key, desc string }{
			{"n", "new session"},
			{"s", "stop session"},
			{"r", "resume session"},
			{"d", "delete session"},
			{"x", "prune worktree"},
		}},
		{"GENERAL", []struct{ key, desc string }{
			{"ctrl+f", "filter sessions"},
			{"t", "switch theme"},
			{"?", "toggle help"},
			{"q", "quit"},
		}},
	}

	for i, g := range groups {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(S.OverlayCategory.Render(g.name))
		b.WriteByte('\n')
		for _, bind := range g.bindings {
			b.WriteString(fmt.Sprintf("  %-12s %s\n",
				S.OverlayKey.Render(bind.key),
				S.OverlayDesc.Render(bind.desc)))
		}
	}

	b.WriteByte('\n')
	b.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", 32)))
	b.WriteByte('\n')
	b.WriteString(S.OverlayFooter.Render("    press ? or esc to close"))

	overlay := S.OverlayStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}
```

- [ ] **Step 6: Rewrite viewPruneConfirmOverlay() with themed styles**

```go
func (m AppModel) viewPruneConfirmOverlay() string {
	sess := m.prune.session
	wtPath := ""
	if sess.Worktree != nil {
		wtPath = *sess.Worktree
	}

	var msg strings.Builder
	msg.WriteString(S.OverlayTitle.Render("Prune Worktree"))
	msg.WriteByte('\n')
	msg.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", 40)))
	msg.WriteString("\n\n")
	msg.WriteString(S.OverlayDesc.Render("Session: " + sess.Name))
	msg.WriteByte('\n')
	msg.WriteString(S.OverlayDesc.Render("Path: " + wtPath))
	msg.WriteString("\n\n")

	if m.prune.dirty {
		msg.WriteString(S.ErrorStyle.Render("WARNING: worktree has uncommitted changes!"))
		msg.WriteString("\n\n")
	}

	if sess.Status == "stopped" || sess.Status == "finished" {
		msg.WriteString(S.OverlayFooter.Render("Session is " + sess.Status + " and will also be deleted."))
		msg.WriteString("\n\n")
	}

	msg.WriteString("Are you sure? ")
	msg.WriteString(S.HelpKeyStyle.Render("y") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("confirm"))
	msg.WriteString("  ")
	msg.WriteString(S.HelpKeyStyle.Render("n") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("cancel"))

	overlay := S.OverlayStyle.Render(msg.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}
```

- [ ] **Step 7: Create stub theme_picker.go so app.go compiles**

Create `internal/app/theme_picker.go` with just the types that `app.go` needs:

```go
package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/abhinav/grove/internal/config"
)

type themePickerDoneMsg struct{}
type themePickerCancelMsg struct{}

type ThemePickerModel struct {
	config *config.Config
}

func NewThemePickerModel(cfg *config.Config) ThemePickerModel {
	return ThemePickerModel{config: cfg}
}

func (m ThemePickerModel) Update(msg tea.Msg) (ThemePickerModel, tea.Cmd) {
	return m, nil
}

func (m ThemePickerModel) View() string {
	return "theme picker stub"
}
```

This stub will be replaced with the full implementation in Task 9.

- [ ] **Step 8: Update WindowSizeMsg handler for new header bar height**

In `app.go`, update the `WindowSizeMsg` handler to reserve 2 lines (header bar + status bar) instead of 1:

```go
case tea.WindowSizeMsg:
	m.width = msg.Width
	m.height = msg.Height
	m.list.Width = msg.Width
	m.list.Height = msg.Height - 2 // header bar + status bar
	return m, nil
```

Apply the same change in `updateCreate`, `updatePeek`, `updatePruneConfirm`, and `updateHelp` handlers.

- [ ] **Step 9: Add theme picker key handler and overlay routing**

In `handleKey()`, add before the Help case:

```go
case key.Matches(msg, keys.Theme):
	m.themePicker = NewThemePickerModel(m.config)
	m.view = viewThemePicker
	return m, nil
```

Add `themePicker ThemePickerModel` field to `AppModel`.

Add stub `viewThemePickerOverlay()`:

```go
func (m AppModel) viewThemePickerOverlay() string {
	overlay := S.OverlayStyle.Render(m.themePicker.View())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}
```

Add `updateThemePicker()` routing in `Update()`:

```go
case viewThemePicker:
	return m.updateThemePicker(msg)
```

And the handler method:

```go
func (m AppModel) updateThemePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case themePickerDoneMsg:
		m.view = viewList
		return m, nil
	case themePickerCancelMsg:
		m.view = viewList
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.reconcileAndLoad(), tickCmd())
	case sessionsMsg:
		m.list.Sessions = msg
		m.list.ClampCursor()
		return m, nil
	}

	var cmd tea.Cmd
	m.themePicker, cmd = m.themePicker.Update(msg)
	return m, cmd
}
```

- [ ] **Step 10: Verify compilation (expect errors from list.go, create.go etc)**

Run: `go build ./internal/app/`
Expected: Errors in `list.go`, `create.go`, `peek.go`, `repo_add.go`, `repo_list.go` referencing old style vars (`selectedRow`, `headerStyle`, `wizardTitleStyle`, etc). This is expected — they are migrated in Tasks 6-8.

- [ ] **Step 11: Commit**

```bash
git add internal/app/app.go internal/app/theme_picker.go
git commit -m "feat(ui): migrate app.go to theme system with contextual status bar and grouped help"
```

### Task 6: Migrate list.go with header bar, zebra striping, and empty state

**Files:**
- Modify: `internal/app/list.go`

- [ ] **Step 1: Rewrite View() with header bar, zebra rows, and accent border**

Replace the `View()` method. Key changes:
- Add header bar at top with `grove | N running / N stopped / N finished` and version
- Table header with `fgDim` separator line
- Selected row gets `bgHighlight` + left accent border character `▎`
- Odd unselected rows get `bgAlt` (zebra)
- Finished sessions dimmed to `fgFaint`

```go
func (m *ListModel) View() string {
	filtered := m.FilteredSessions()

	// Count statuses for header.
	var running, stopped, finished int
	for _, s := range m.Sessions {
		switch s.Status {
		case "running":
			running++
		case "stopped":
			stopped++
		default:
			finished++
		}
	}

	// Header bar.
	left := S.HeaderAccent.Render("grove") +
		S.HeaderPipe.Render(" | ") +
		S.HeaderGreen.Render(fmt.Sprintf("%d running", running)) +
		S.HeaderPipe.Render(" / ") +
		S.HeaderYellow.Render(fmt.Sprintf("%d stopped", stopped)) +
		S.HeaderPipe.Render(" / ") +
		S.HeaderDim.Render(fmt.Sprintf("%d finished", finished))
	headerBar := S.HeaderBarStyle.Width(m.Width).Render(left)

	if len(filtered) == 0 {
		var emptyContent string
		if m.FilterText != "" {
			emptyContent = S.EmptyStyle.Render("No matching sessions.")
		} else {
			diamond := S.EmptyDim.Render(" ◇")
			title := S.EmptyStyle.Render("No sessions yet")
			subtitle := S.EmptyDim.Render("Start your first AI coding session")
			pill := S.HelpKeyStyle.Render("n") + S.HelpDescStyle.Render(" new session")
			emptyContent = lipgloss.JoinVertical(lipgloss.Center,
				diamond, "", title, subtitle, "", pill)
		}
		body := lipgloss.Place(m.Width, m.Height-1, lipgloss.Center, lipgloss.Center, emptyContent)
		return headerBar + "\n" + body
	}

	colGap := "  "
	toolW := 10
	repoW := 14
	durW := 10
	nameW := 20
	fixedW := 2 + nameW + repoW + toolW + durW + len(colGap)*4
	dirW := m.Width - fixedW
	if dirW < 10 {
		dirW = 10
	}

	var b strings.Builder

	// Column header.
	hdr := fmt.Sprintf("  %-*s%s%-*s%s%-*s%s%-*s%s%*s",
		nameW, "NAME", colGap,
		repoW, "REPO", colGap,
		toolW, "TOOL", colGap,
		dirW, "DIRECTORY", colGap,
		durW, "AGE")
	b.WriteString(S.HeaderStyle.Render(hdr))
	b.WriteByte('\n')
	b.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", m.Width)))
	b.WriteByte('\n')

	maxRows := m.Height - 3 // header bar + column header + separator
	if maxRows < 1 {
		maxRows = len(filtered)
	}

	for i, sess := range filtered {
		if i >= maxRows {
			break
		}

		dot := statusDot(sess.Status)
		name := truncate(sess.Name, nameW-1)
		repo := truncate(repoDisplayName(sess), repoW)
		tool := truncate(sess.Tool, toolW)
		dir := truncateLeft(sess.Directory, dirW)
		dur := formatDuration(time.Since(sess.CreatedAt))

		row := fmt.Sprintf("%s %-*s%s%-*s%s%-*s%s%-*s%s%*s",
			dot,
			nameW-1, name, colGap,
			repoW, repo, colGap,
			toolW, tool, colGap,
			dirW, dir, colGap,
			durW, dur)

		if i == m.Cursor {
			// Selected: accent border + highlight bg.
			row = S.AccentBorder.Render("▎") + S.SelectedRow.Width(m.Width-1).Render(row)
		} else if sess.Status == "finished" {
			// Dimmed.
			row = " " + S.FinishedRow.Render(row)
		} else if i%2 == 1 {
			// Zebra.
			row = " " + S.ZebraRow.Width(m.Width-1).Render(row)
		} else {
			row = " " + S.NormalRow.Render(row)
		}

		b.WriteString(row)
		if i < len(filtered)-1 && i < maxRows-1 {
			b.WriteByte('\n')
		}
	}

	return headerBar + "\n" + b.String()
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/app/`
Expected: Errors from `create.go`, `peek.go`, `repo_add.go`, `repo_list.go` still.

- [ ] **Step 3: Commit**

```bash
git add internal/app/list.go
git commit -m "feat(ui): themed list view with header bar, zebra striping, empty state"
```

### Task 7: Migrate create.go with step progress bar

**Files:**
- Modify: `internal/app/create.go`

- [ ] **Step 1: Add renderStepProgress helper function**

Add at the bottom of `create.go`:

```go
// renderStepProgress renders the numbered horizontal step indicator.
func renderStepProgress(current int, steps []string) string {
	var parts []string
	for i, name := range steps {
		num := fmt.Sprintf("%d", i+1)
		if i <= current {
			// Active/completed: filled circle.
			circle := S.WizardStepActive.Render(" " + num + " ")
			label := S.WizardStepActiveLabel.Render(name)
			parts = append(parts, circle+" "+label)
		} else {
			// Future: dim.
			circle := S.WizardStepInactive.Render("(" + num + ")")
			label := S.WizardStepInactiveLabel.Render(name)
			parts = append(parts, circle+" "+label)
		}
		if i < len(steps)-1 {
			parts = append(parts, S.WizardStepConnector.Render(" ── "))
		}
	}
	return strings.Join(parts, "")
}
```

- [ ] **Step 2: Replace all old style references with S.xxx**

Throughout `create.go`, replace:
- `wizardTitleStyle` → `S.WizardTitleStyle`
- `wizardLabelStyle` → `S.WizardLabelStyle`
- `wizardChoiceStyle` → `S.WizardChoiceStyle`
- `wizardSelectedToolStyle` → `S.WizardSelectedPill`
- `wizardToolStyle` → `S.WizardUnselectedPill`
- `dimStyle` → `S.WizardDimStyle`
- `errorStyle` → `S.ErrorStyle`

- [ ] **Step 3: Rewrite View() to include step progress bar**

At the top of the `View()` method, after the title, add the progress bar. Determine step names based on `m.dirSource`:

```go
func (m CreateModel) View() string {
	var b strings.Builder

	// Header bar.
	headerBar := S.HeaderBarStyle.Width(m.width).Render(
		S.HeaderAccent.Render("New Session"))
	b.WriteString(headerBar)
	b.WriteString("\n\n")

	// Step progress.
	var stepNames []string
	var currentIdx int
	if m.dirSource == dirWorktree && m.step >= stepRepoSelect {
		stepNames = []string{"Source", "Repo", "Branch", "Tool", "Prompt", "Confirm"}
		switch m.step {
		case stepRepoSelect:
			currentIdx = 1
		case stepDirInput:
			currentIdx = 2
		case stepTool:
			currentIdx = 3
		case stepPrompt:
			currentIdx = 4
		case stepConfirm:
			currentIdx = 5
		default:
			currentIdx = 0
		}
	} else {
		stepNames = []string{"Source", "Path", "Tool", "Prompt", "Confirm"}
		switch m.step {
		case stepDirSource:
			currentIdx = 0
		case stepDirInput:
			currentIdx = 1
		case stepTool:
			currentIdx = 2
		case stepPrompt:
			currentIdx = 3
		case stepConfirm:
			currentIdx = 4
		default:
			currentIdx = 0
		}
	}
	b.WriteString("  " + renderStepProgress(currentIdx, stepNames))
	b.WriteString("\n\n")

	// Step content.
	switch m.step {
	case stepDirSource:
		b.WriteString(S.WizardLabelStyle.Render("Directory source"))
		b.WriteString("\n\n")
		b.WriteString("  " + S.WizardChoiceStyle.Render("[1]") + " Use existing directory\n")
		b.WriteString("  " + S.WizardChoiceStyle.Render("[2]") + " Create worktree\n")

	case stepRepoSelect:
		b.WriteString(S.WizardLabelStyle.Render("Select repository"))
		b.WriteString("\n\n  ")
		for i, name := range m.repoNames {
			if i == m.repoSelected {
				b.WriteString(S.WizardSelectedPill.Render(name))
			} else {
				b.WriteString(S.WizardUnselectedPill.Render(name))
			}
			if i < len(m.repoNames)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n\n  " + S.WizardDimStyle.Render("← → to select, enter to confirm"))

	case stepDirInput:
		if m.dirSource == dirExisting {
			b.WriteString(S.WizardLabelStyle.Render("Enter directory path"))
		} else {
			b.WriteString(S.WizardLabelStyle.Render("Enter branch name"))
		}
		b.WriteString("\n\n")
		b.WriteString("  " + m.dirInput.View())

	case stepTool:
		b.WriteString(S.WizardLabelStyle.Render("Select AI tool"))
		b.WriteString("\n\n  ")
		for i, name := range m.toolNames {
			if i == m.toolSelected {
				b.WriteString(S.WizardSelectedPill.Render(name))
			} else {
				b.WriteString(S.WizardUnselectedPill.Render(name))
			}
			if i < len(m.toolNames)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n\n  " + S.WizardDimStyle.Render("← → to select, enter to confirm"))

	case stepPrompt:
		b.WriteString(S.WizardLabelStyle.Render("Prompt or plan file (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.promptInput.View())
		b.WriteString("\n\n  " + S.WizardDimStyle.Render("enter to continue (leave empty for interactive)"))

	case stepConfirm:
		b.WriteString(S.WizardLabelStyle.Render("Confirm"))
		b.WriteString("\n\n")
		if m.dirSource == dirExisting {
			b.WriteString("  " + S.HeaderDim.Render("Directory:  ") + S.OverlayDesc.Render(m.resolvedDir) + "\n")
		} else {
			b.WriteString("  " + S.HeaderDim.Render("Repo:       ") + S.OverlayDesc.Render(m.selectedRepo.RepoRoot) + "\n")
			b.WriteString("  " + S.HeaderDim.Render("Worktree:   ") + S.OverlayDesc.Render(m.worktreeBranch) + "\n")
		}
		b.WriteString("  " + S.HeaderDim.Render("Tool:       ") + S.OverlayDesc.Render(m.toolNames[m.toolSelected]) + "\n")
		prompt := strings.TrimSpace(m.promptInput.Value())
		if prompt == "" {
			prompt = "(interactive)"
		}
		b.WriteString("  " + S.HeaderDim.Render("Prompt:     ") + S.OverlayDesc.Render(prompt) + "\n")
		b.WriteString("\n  " + S.WizardDimStyle.Render("enter to create, esc to cancel"))
	}

	if m.err != "" {
		b.WriteString("\n\n  " + S.ErrorStyle.Render(m.err))
	}

	// Status bar.
	bar := S.StatusBarStyle.Width(m.width).Render(
		S.HelpKeyStyle.Render("enter") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("next") + "  " +
			S.HelpKeyStyle.Render("esc") + S.HelpPipe.Render(":") + S.HelpDescStyle.Render("cancel"))

	return b.String() + "\n" + bar
}
```

- [ ] **Step 4: Remove old wizard style vars at the bottom of create.go**

Delete the `var (wizardTitleStyle = ...` block at the bottom since these are now in `Styles`.

- [ ] **Step 5: Verify compilation**

Run: `go build ./internal/app/`
Expected: Errors from `peek.go`, `repo_add.go`, `repo_list.go`.

- [ ] **Step 6: Commit**

```bash
git add internal/app/create.go
git commit -m "feat(ui): themed create wizard with step progress bar"
```

### Task 8: Migrate peek.go, repo_add.go, repo_list.go

**Files:**
- Modify: `internal/app/peek.go`
- Modify: `internal/app/repo_add.go`
- Modify: `internal/app/repo_list.go`

- [ ] **Step 1: Update peek.go**

Replace:
- `peekHeaderStyle` → `S.PeekHeaderStyle`
- `statusBarStyle` → `S.StatusBarStyle`
- `helpKeyStyle` → `S.HelpKeyStyle`
- `helpDescStyle` → `S.HelpDescStyle`

Delete the `peekHeaderStyle` var at the bottom.

- [ ] **Step 2: Update repo_add.go**

Replace style references:
- `wizardTitleStyle` → `S.WizardTitleStyle`
- `wizardLabelStyle` → `S.WizardLabelStyle`
- `wizardChoiceStyle` → `S.WizardChoiceStyle`
- `dimStyle` → `S.WizardDimStyle`
- `errorStyle` → `S.ErrorStyle`

Add step progress bar at the top of `View()`. Replace the title and add progress after it:

```go
func (m RepoAddModel) View() string {
	var b strings.Builder

	headerBar := S.HeaderBarStyle.Width(m.width).Render(
		S.HeaderAccent.Render("Add Repository"))
	b.WriteString(headerBar)
	b.WriteString("\n\n")

	// Step progress.
	stepNames := []string{"Root", "Worktree", "AI Tool", "Setup", "Save"}
	var currentIdx int
	switch m.step {
	case repoStepConfirmRoot:
		currentIdx = 0
	case repoStepWorktreeBase:
		currentIdx = 1
	case repoStepAITool:
		currentIdx = 2
	case repoStepSetupCommands:
		currentIdx = 3
	case repoStepFinalConfirm:
		currentIdx = 4
	}
	b.WriteString("  " + renderStepProgress(currentIdx, stepNames))
	b.WriteString("\n\n")

	// Then the existing switch on m.step with S.xxx style replacements...
```

Apply the same pattern as create.go's View() for each step case — use `S.WizardLabelStyle` for step titles, `S.WizardDimStyle` for hints, `S.ErrorStyle` for errors, and add a status bar at the bottom.

- [ ] **Step 3: Update repo_list.go**

Replace all local style vars with `S.xxx`:
- `repoListHeaderStyle` → `S.RepoListHeaderStyle`
- `repoListCellStyle` → `S.RepoListCellStyle`
- `repoListDimStyle` → `S.RepoListDimStyle`
- `repoListTitleStyle` → `S.RepoListTitleStyle`
- `repoListBorderStyle` → `S.RepoListBorderStyle`
- `emptyStyle` → `S.EmptyStyle`

Delete the local `var (repoListHeaderStyle = ...` block.

- [ ] **Step 4: Verify full compilation**

Run: `go build ./internal/app/`
Expected: PASS — all old style references replaced.

- [ ] **Step 5: Run all tests**

Run: `go test ./...`
Expected: All tests pass. Existing tests check model logic and view content (strings), not exact ANSI colors, so they should still work.

- [ ] **Step 6: Commit**

```bash
git add internal/app/peek.go internal/app/repo_add.go internal/app/repo_list.go
git commit -m "feat(ui): migrate peek, repo-add, repo-list to theme system"
```

## Chunk 3: Theme Picker and Integration

### Task 9: Create theme_picker.go

**Files:**
- Create: `internal/app/theme_picker.go`

- [ ] **Step 1: Create ThemePickerModel**

```go
package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abhinav/grove/internal/config"
)

// themePickerDoneMsg signals theme was selected and saved.
type themePickerDoneMsg struct{}

// themePickerCancelMsg signals the user cancelled.
type themePickerCancelMsg struct{}

// ThemePickerModel is the theme selection overlay.
type ThemePickerModel struct {
	cursor        int
	originalTheme Theme
	config        *config.Config
}

// NewThemePickerModel creates a theme picker starting on the current theme.
func NewThemePickerModel(cfg *config.Config) ThemePickerModel {
	cursor := 0
	for i, t := range Themes {
		if t.Name == ActiveTheme.Name {
			cursor = i
			break
		}
	}
	return ThemePickerModel{
		cursor:        cursor,
		originalTheme: ActiveTheme,
		config:        cfg,
	}
}

// Update handles input for the theme picker.
func (m ThemePickerModel) Update(msg tea.Msg) (ThemePickerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
				SetTheme(Themes[m.cursor])
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(Themes)-1 {
				m.cursor++
				SetTheme(Themes[m.cursor])
			}
		case key.Matches(msg, keys.Attach): // enter
			// Save to config.
			m.config.Defaults.Theme = Themes[m.cursor].Name
			if err := config.Save(m.config); err != nil {
				return m, func() tea.Msg {
					return errMsg{fmt.Errorf("saving theme: %w", err)}
				}
			}
			return m, func() tea.Msg { return themePickerDoneMsg{} }
		case key.Matches(msg, keys.Escape):
			// Revert to original theme.
			SetTheme(m.originalTheme)
			return m, func() tea.Msg { return themePickerCancelMsg{} }
		}
	}
	return m, nil
}

// View renders the theme picker list.
func (m ThemePickerModel) View() string {
	var b strings.Builder

	b.WriteString(S.OverlayTitle.Render("Theme"))
	b.WriteByte('\n')
	b.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", 28)))
	b.WriteByte('\n')
	b.WriteByte('\n')

	for i, t := range Themes {
		if i == m.cursor {
			b.WriteString(S.AccentBorder.Render("▎") + " ")
			b.WriteString(lipgloss.NewStyle().
				Background(ActiveTheme.BgHighlight).
				Foreground(ActiveTheme.Fg).
				Bold(true).
				Width(24).
				Render(t.Name))
		} else {
			b.WriteString("  ")
			b.WriteString(S.OverlayDesc.Render(t.Name))
		}
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	b.WriteString(S.OverlaySeparator.Render(strings.Repeat("─", 28)))
	b.WriteByte('\n')
	b.WriteString(S.OverlayFooter.Render("  j/k:navigate  enter:apply  esc:cancel"))

	return b.String()
}
```

- [ ] **Step 2: Verify compilation**

Run: `go build ./internal/app/`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/app/theme_picker.go
git commit -m "feat(ui): add theme picker overlay with live preview"
```

### Task 10: Wire theme loading from config on startup

**Files:**
- Modify: `internal/app/app.go`

- [ ] **Step 1: Initialize theme from config in New()**

In the `New()` function, after creating the `AppModel`, set the theme:

```go
func New(s *store.Store, cfg *config.Config, mgr *session.Manager) AppModel {
	// Set theme from config.
	if cfg.Defaults.Theme != "" {
		SetTheme(ThemeByName(cfg.Defaults.Theme))
	}

	var repoOrder []string
	for _, repo := range cfg.Repos {
		repoOrder = append(repoOrder, filepath.Base(repo.RepoRoot))
	}
	return AppModel{
		view:    viewList,
		store:   s,
		config:  cfg,
		manager: mgr,
		list:    ListModel{RepoOrder: repoOrder},
	}
}
```

- [ ] **Step 2: Build and run full test suite**

Run: `go build ./cmd/grove/ && go test ./...`
Expected: Build succeeds, all tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/app/app.go
git commit -m "feat(ui): load theme from config on startup"
```

### Task 11: Build binary, manual smoke test, and final commit

**Files:** None new — integration verification.

- [ ] **Step 1: Build release binary**

Run: `go build -o ./grove ./cmd/grove/`

- [ ] **Step 2: Run grove --help to verify CLI still works**

Run: `./grove --help`
Expected: Cobra help output unchanged.

- [ ] **Step 3: Run grove list to verify themed CLI output**

Run: `./grove list`
Expected: Session list with themed colors.

- [ ] **Step 4: Smoke test theme picker**

Run `./grove` (TUI mode). Verify:
1. Press `t` — theme picker overlay appears centered with "Theme" title
2. Cursor starts on the current theme (tokyo-night by default)
3. Press `j` — cursor moves to catppuccin-macchiato, UI colors change live
4. Press `k` — cursor moves back, colors revert to tokyo-night
5. Navigate to catppuccin-macchiato, press `enter` — overlay closes, theme persists
6. Verify config saved: `grep theme ~/.config/grove/config.toml` shows `theme = "catppuccin-macchiato"`
7. Restart `./grove` — should load with catppuccin-macchiato colors
8. Press `t`, navigate, press `esc` — should revert to catppuccin-macchiato (no change)

- [ ] **Step 5: Run full test suite one final time**

Run: `go test ./...`
Expected: All tests pass.

- [ ] **Step 6: Push branch and create PR**

```bash
git push -u origin feature/tui-visual-overhaul
```

Then create a PR to merge into main.
