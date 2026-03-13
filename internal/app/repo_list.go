package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/abhinav/grove/internal/config"
)

// PrintRepoList prints a lipgloss-styled table of registered repos.
func PrintRepoList(cfg *config.Config) {
	if len(cfg.Repos) == 0 {
		fmt.Println(S.Empty.Render("  No repositories registered. Use `grove repo add` to add one."))
		return
	}

	fmt.Println(S.RepoListTitle.Render("Registered Repositories"))
	fmt.Println()

	// Calculate column widths.
	headers := []string{"REPO", "WORKTREE BASE", "AI TOOL", "SETUP COMMANDS"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	type row struct {
		repo, wtBase, aiTool, setupCmds string
		wtDim, aiDim, setupDim          bool
	}
	rows := make([]row, 0, len(cfg.Repos))

	for _, repo := range cfg.Repos {
		r := row{
			repo: filepath.Base(repo.RepoRoot),
		}

		// Worktree base.
		if repo.WorktreeBase != "" {
			r.wtBase = repo.WorktreeBase
		} else {
			r.wtBase = cfg.EffectiveWorktreeBase(&repo)
			if r.wtBase == "" {
				r.wtBase = "(none)"
			} else {
				r.wtBase = r.wtBase + " (global)"
			}
			r.wtDim = true
		}

		// AI tool.
		if repo.AITool != "" {
			r.aiTool = repo.AITool
		} else {
			r.aiTool = cfg.EffectiveAITool(&repo)
			if r.aiTool == "" {
				r.aiTool = "(none)"
			} else {
				r.aiTool = r.aiTool + " (global)"
			}
			r.aiDim = true
		}

		// Setup commands.
		if len(repo.SetupCommands) > 0 {
			r.setupCmds = strings.Join(repo.SetupCommands, ", ")
		} else {
			cmds := cfg.EffectiveSetupCommands(&repo)
			if len(cmds) == 0 {
				r.setupCmds = "(none)"
			} else {
				r.setupCmds = strings.Join(cmds, ", ") + " (global)"
			}
			r.setupDim = true
		}

		// Update widths.
		vals := []string{r.repo, stripAnsi(r.wtBase), stripAnsi(r.aiTool), stripAnsi(r.setupCmds)}
		for i, v := range vals {
			if len(v) > widths[i] {
				widths[i] = len(v)
			}
		}

		rows = append(rows, r)
	}

	// Print header row.
	var headerLine strings.Builder
	for i, h := range headers {
		headerLine.WriteString(S.RepoListHeader.Width(widths[i] + 2).Render(h))
	}
	fmt.Println(headerLine.String())

	// Print separator.
	var sep strings.Builder
	for i, w := range widths {
		sep.WriteString(S.RepoListBorder.Render(strings.Repeat("─", w+2)))
		if i < len(widths)-1 {
			sep.WriteString(S.RepoListBorder.Render("─"))
		}
	}
	fmt.Println(sep.String())

	// Print rows.
	for _, r := range rows {
		var line strings.Builder
		cells := []struct {
			val string
			dim bool
		}{
			{r.repo, false},
			{r.wtBase, r.wtDim},
			{r.aiTool, r.aiDim},
			{r.setupCmds, r.setupDim},
		}
		for i, c := range cells {
			if c.dim {
				line.WriteString(S.RepoListDim.Width(widths[i] + 2).Render(c.val))
			} else {
				line.WriteString(S.RepoListCell.Width(widths[i] + 2).Render(c.val))
			}
		}
		fmt.Println(line.String())
	}
}

// stripAnsi removes the " (global)" suffix for width calculation purposes.
// This is a simple approach since our dim indicators are plain text.
func stripAnsi(s string) string {
	return s
}
