# Grove TUI Visual Overhaul — Design Spec

**Date:** 2026-03-14
**Status:** Approved
**Scope:** Complete visual redesign of the Grove TUI using Tokyo Night palette, improved layout, and contextual UI patterns.

## Decisions

- **Color palette:** Tokyo Night (cool blues, muted pastels on deep navy)
- **Session list layout:** Flat table with zebra striping, status summary in header, contextual status bar
- **Create wizard:** Numbered horizontal step progress bar (①──②──③──④──⑤)
- **Help overlay:** Grouped keybindings (Navigation / Actions / General) in centered card
- **Empty state:** Minimal diamond icon, prominent `n` key CTA, reduced status bar
- **New dependencies:** None — all features use existing lipgloss v1.1.0 capabilities

## 1. Color System — Theme Architecture

Replace all raw ANSI color numbers with a theme system that supports multiple color themes. Ship with Tokyo Night as the default, but structure the code so new themes can be added by defining a new `Theme` struct.

### Theme Struct

Define a `Theme` struct in `internal/app/theme.go` with named color slots:

```go
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
```

Each color uses `lipgloss.CompleteColor` with TrueColor, ANSI256, and ANSI16 values for graceful degradation.

### Tokyo Night Theme (Default)

| Slot | Hex (TrueColor) | ANSI256 | ANSI16 | Role |
|------|-----------------|---------|--------|------|
| `Accent` | `#7aa2f7` | 111 | 12 (bright blue) | Primary accent — keys, selections, branding |
| `Green` | `#9ece6a` | 150 | 10 (bright green) | Running status, success checkmarks |
| `Yellow` | `#e0af68` | 180 | 11 (bright yellow) | Stopped/warning |
| `Red` | `#f7768e` | 210 | 9 (bright red) | Errors, destructive warnings |
| `Fg` | `#c0caf5` | 189 | 15 (bright white) | Primary text |
| `FgMuted` | `#a9b1d6` | 146 | 7 (white) | Secondary text |
| `FgDim` | `#565f89` | 60 | 8 (bright black) | Tertiary text, column headers, hints |
| `FgFaint` | `#414868` | 59 | 8 (bright black) | Borders, separators, disabled text |
| `BgHighlight` | `#292e42` | 236 | 0 (black) | Selected row, hover states |
| `BgSurface` | `#24283b` | 235 | 0 (black) | Header bar, status bar, overlay backgrounds |
| `BgAlt` | `#16161e` | 233 | 0 (black) | Zebra stripe alternate rows |

### Catppuccin Macchiato Theme

| Slot | Hex (TrueColor) | ANSI256 | ANSI16 | Role |
|------|-----------------|---------|--------|------|
| `Accent` | `#8aadf4` | 111 | 12 (bright blue) | Primary accent |
| `Green` | `#a6da95` | 150 | 10 (bright green) | Running status, success |
| `Yellow` | `#eed49f` | 186 | 11 (bright yellow) | Stopped/warning |
| `Red` | `#ed8796` | 211 | 9 (bright red) | Errors, destructive |
| `Fg` | `#cad3f5` | 189 | 15 (bright white) | Primary text |
| `FgMuted` | `#b8c0e0` | 146 | 7 (white) | Secondary text |
| `FgDim` | `#6e738d` | 60 | 8 (bright black) | Tertiary text, hints |
| `FgFaint` | `#494d64` | 59 | 8 (bright black) | Borders, separators |
| `BgHighlight` | `#363a4f` | 237 | 0 (black) | Selected row |
| `BgSurface` | `#1e2030` | 234 | 0 (black) | Header/status bars |
| `BgAlt` | `#181926` | 233 | 0 (black) | Zebra stripe |

### Theme Registry

A package-level slice of available themes (slice, not map, to preserve display order):

```go
var Themes = []Theme{
    TokyoNight,
    CatppuccinMacchiato,
}
```

Adding a new theme = defining another `Theme` var and appending it to this slice. No other code changes needed.

### Active Theme & Style Rebuilding

A package-level `ActiveTheme` pointer. All styles in `styles.go` are built via a `BuildStyles(theme Theme)` function that returns a `Styles` struct containing all lipgloss styles derived from the theme's color slots. When the theme changes:

1. `ActiveTheme` is updated
2. `BuildStyles` is called to rebuild all styles
3. The UI re-renders immediately with the new colors

This approach means styles are never stale — they're always derived from the current theme.

### Config Persistence

The selected theme name is stored in `config.toml` under `[defaults]`:

```toml
[defaults]
ai_tool = "claude"
theme = "tokyo-night"
```

- On startup, `config.Load()` reads the `theme` field and sets `ActiveTheme` accordingly
- If the field is missing or the theme name is unrecognized, default to Tokyo Night
- When the user switches themes via the picker, the config is saved immediately

This requires adding a `Theme string` field to the config `Defaults` struct and a `SaveTheme(name string)` method (or reuse existing config save logic).

### Implementation

All theme types, built-in themes, the registry, and `BuildStyles` defined in `internal/app/theme.go`. Styles in `styles.go` become a `Styles` struct populated by `BuildStyles`.

## 2. Session List

### Header Bar

- Full-width bar with `bgSurface` background
- Left side: `grove` in bold `accent` + pipe separator (`fgFaint`) + status summary
- Status summary: colored counts — e.g. `2 running` in `green`, `1 stopped` in `yellow`, `2 finished` in `fgDim`, separated by `/` in `fgFaint`
- Right side: version string in `fgFaint`

### Table

Hand-rendered using lipgloss layout functions (not `lipgloss/table` — we need per-row interactivity and conditional styling that the static table renderer doesn't support).

- **Column headers:** `fgDim` bold with slight letter-spacing feel, followed by a `fgFaint` horizontal rule separator
- **Columns:** status dot (2w), NAME, REPO, TOOL, DIRECTORY, AGE
- **Selected row:** `bgHighlight` background + 2px left border in `accent` + name in bold `fg`
- **Unselected rows:** alternate between default background and `bgAlt` (zebra striping)
- **Finished/dead sessions:** entire row rendered in `fgFaint` (dimmed)

### Filter Bar

When filtering is active (`ctrl+f`):
- Full-width `bgSurface` bar replaces the status bar
- "filter: " label in bold `accent`, followed by the text input
- Active filter (after committing): shown as italic `fgDim` text with the filter query, plus condensed status bar help to the right

### Flash Messages

Transient messages displayed in the status bar area:
- Error messages: `red` foreground
- Info messages: `accent` foreground

### Contextual Status Bar

Full-width `bgSurface` bar whose content changes based on the selected session's state:

- **Left section:** selected session status as italic colored label (e.g. `running` in green)
- **Middle section:** pipe separator + session-relevant keys only:
  - Running: `enter:attach  p:peek  s:stop`
  - Stopped: `r:resume  d:delete`
  - Finished: `d:delete  x:prune`
  - No selection: empty
- **Right section:** always-available keys separated by pipe: `n:new  ?:help  q:quit`
- Key names in `accent`, descriptions in `fgDim`, pipes in `fgFaint`

### Empty State

When no sessions exist:

- Header bar present with "0 sessions"
- Centered vertically in the content area:
  - Small diamond glyph (`╱╲ / ╲╱`) in `fgFaint`
  - "No sessions yet" in `fgDim`
  - "Start your first AI coding session" in `fgFaint`
  - Prominent CTA: `n` key in a `bgHighlight` pill with `fgFaint` border, "new session" label in `fgDim`
- Status bar only shows: `n:new  ?:help  q:quit`

## 3. Create Wizard

### Header Bar

Same `bgSurface` bar style as list view. Content: "New Session" in bold `accent`.

### Numbered Step Progress

Horizontal indicator rendered below the header bar:

- **Completed steps:** filled circle (number in dark text on `accent` background) + step name in `accent`
- **Current step:** same filled styling as completed
- **Future steps:** outlined circle (number in `fgFaint` with `fgFaint` border) + name in `fgFaint`
- Steps connected by `────` dash segments in `fgFaint`
- Step names for existing-directory flow: Source → Path → Tool → Prompt → Confirm
- Step names for worktree flow: Source → Repo → Branch → Tool → Prompt → Confirm

### Step Content

- Step title in bold `fg`
- Text inputs: `bgSurface` background with `fgFaint` border, typed text in `fg`, cursor in `accent`
- Tool/repo selector pills: selected = `accent` background with dark text (bold), unselected = `bgHighlight` background with `fgMuted` text
- Hint text below inputs in `fgFaint` italic (e.g. "← → to select, enter to confirm")

### Confirm Step

Summary of all choices as aligned key-value pairs:
- Labels in `fgDim` bold (e.g. "Directory:", "Tool:", "Prompt:")
- Values in `fgMuted`
- Hint: "enter to create, esc to cancel" in `fgFaint` italic

### Status Bar

Only relevant keys: `enter:next  esc:cancel`

## 4. Help Overlay

Centered card rendered with `lipgloss.Place`:

- **Background:** `bgSurface`
- **Border:** `fgFaint` color, `lipgloss.RoundedBorder()`
- **Padding:** 1 vertical, 3 horizontal
- **Title:** "Keybindings" in bold `accent`
- **Separator:** horizontal rule in `fgFaint` below title

### Grouped Keybindings

Three groups, each with an uppercase category label in `fgDim`:

**Navigation**
- `j/k` — down / up
- `enter` — attach to session
- `p` — peek (live preview)

**Actions**
- `n` — new session
- `s` — stop session
- `r` — resume session
- `d` — delete session
- `x` — prune worktree

**General**
- `ctrl+f` — filter sessions
- `t` — switch theme
- `?` — toggle help
- `q` — quit

Each row: key in `accent` left-aligned, description in `fgMuted` right-aligned.

### Footer

Centered text: "press ? or esc to close" in `fgFaint`, separated from content by a `fgFaint` rule.

## 5. Theme Picker Overlay

A new view (`viewThemePicker`) triggered by a keybinding.

### Keybinding

`t` — open theme picker. Added to `keyMap` and shown in the help overlay under **General**.

### Overlay UI

Centered floating card (same style as help overlay — `bgSurface` bg, `fgFaint` rounded border):

- **Title:** "Theme" in bold `accent`
- **Theme list:** vertical list of theme names, one per line
  - Selected (cursor) row: `bgHighlight` background + `accent` left border + name in bold `fg`
  - Other rows: name in `fgMuted`
- **Live preview:** as the cursor moves between themes, `ActiveTheme` and all styles are immediately rebuilt. The overlay itself and any visible background UI re-render with the previewed theme's colors in real-time.
- **Navigation:** `j/k` or `↑/↓` to move cursor
- **Confirm:** `enter` — saves the theme to `config.toml` and closes the overlay
- **Cancel:** `esc` — reverts to the previously active theme and closes the overlay

### Behavior

1. User presses `t` → overlay opens, cursor on currently active theme
2. User navigates → each cursor move triggers `BuildStyles` with the hovered theme, UI re-renders live
3. User presses `enter` → theme is saved to config, overlay closes
4. User presses `esc` → original theme is restored via `BuildStyles`, overlay closes

### Status Bar

While theme picker is open: `j/k:navigate  enter:apply  esc:cancel`

## 6. Peek View

- **Header bar:** full-width with `accent` background, session name + tmux session ID in white bold
- **Viewport:** unchanged — raw tmux `capture-pane` content with ANSI passthrough
- **Status bar:** `bgSurface` bar with `esc:back  enter:attach  ↑↓:scroll`

## 7. Prune Confirm Overlay

Same centered card style as help overlay:

- Session name and worktree path in `fgMuted`
- If dirty: "WARNING: worktree has uncommitted changes!" in bold `red`
- If session is stopped/finished: note that session will also be deleted in `fgDim`
- Prompt: "Are you sure? (y/n)" in `fg`
- Status bar: `y:confirm  n:cancel` keys only

## 8. Repo Add Wizard

Follows the same visual pattern as the create wizard:

- Header bar with "Add Repository" in bold `accent`
- Numbered step progress: Confirm Root → Worktree Base → AI Tool → Setup Commands → Save
- Text inputs with `bgSurface` background and `fgFaint` borders
- Confirm step with key-value summary
- Contextual status bar

## 9. File Changes

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/app/theme.go` | **New** | Theme struct, Tokyo Night + Catppuccin Macchiato definitions, registry, BuildStyles, ActiveTheme |
| `internal/app/styles.go` | **Rewrite** | Replace raw ANSI numbers with `Styles` struct populated by `BuildStyles` |
| `internal/app/list.go` | **Modify** | Header bar, zebra striping, left accent border, status summary, empty state |
| `internal/app/app.go` | **Modify** | Contextual status bar, help overlay with grouped keybindings, theme picker view routing |
| `internal/app/theme_picker.go` | **New** | Theme picker overlay model — list, live preview, save on confirm |
| `internal/app/create.go` | **Modify** | Numbered step progress bar, styled input fields, confirm summary |
| `internal/app/peek.go` | **Minor** | Update header/status bar colors to theme |
| `internal/app/repo_add.go` | **Minor** | Match create wizard visual pattern |
| `internal/app/repo_list.go` | **Minor** | Migrate hardcoded styles to theme references |
| `internal/app/keys.go` | **Modify** | Add `t` keybinding for theme picker, update help text |
| `internal/config/config.go` | **Modify** | Add `Theme` field to `Defaults` struct, save theme on switch |

## 10. Non-Goals

- No new Go dependencies (glamour, huh, harmonica) — all achieved with existing lipgloss v1.1.0
- No animation/transitions — keep it snappy and deterministic
- No custom user-defined themes — only built-in themes (Tokyo Night, Catppuccin Macchiato) selectable via picker
- No changes to business logic, keybindings, or data flow
- No changes to CLI commands, config, or store layer
