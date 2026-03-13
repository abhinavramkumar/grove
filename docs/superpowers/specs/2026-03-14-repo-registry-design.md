# Repo Registry with Repo-Scoped Views

**Date:** 2026-03-14
**Status:** Approved

## Problem

1. No error handling when grove is used in a non-git directory with worktree mode — cryptic `exit status 128` error.
2. No concept of which repository a session belongs to — flat global list, no grouping.
3. Worktree creation hardcodes `"."` as the repo source — no cross-repo support.
4. Single global `worktree_base` — no per-repo worktree directories.

## Approach

**Config-primary repo registry with DB denormalization (Approach C).**

- Config.toml owns repo definitions with per-repo overrides of all global settings.
- Sessions table gets a `repo_root` column as a denormalized grouping key.
- Config is authoritative for repo settings; DB enables fast session grouping.

## Config Schema

```toml
# Global defaults (existing, unchanged)
[defaults]
ai_tool = "claude"
worktree_base = "~/Projects/Work"

[worktree]
setup_commands = ["npm install"]

[tools.claude]
command = "claude"
args = ["-p"]

# New: repo registry
[[repos]]
repo_root = "/Users/abhinav/Projects/Work/fermat"
worktree_base = "/Users/abhinav/Projects/Work/fermat-worktrees"

[[repos]]
repo_root = "/Users/abhinav/Projects/Work/global-scripts/grove"
worktree_base = "/Users/abhinav/Projects/Work/grove-worktrees"
ai_tool = "claude"
setup_commands = ["go mod download"]
```

**Go structs:**

```go
type RepoConfig struct {
    RepoRoot      string   `toml:"repo_root"`
    WorktreeBase  string   `toml:"worktree_base"`
    AITool        string   `toml:"ai_tool,omitempty"`
    SetupCommands []string `toml:"setup_commands,omitempty"`
}
```

**Resolution order:** repo-level > global > built-in default.

Helper methods: `RepoFor(repoRoot)`, `EffectiveWorktreeBase(repo)`, `EffectiveAITool(repo)`, `EffectiveSetupCommands(repo)`.

## DB Schema Change

Migration v2 → v3:

```sql
ALTER TABLE sessions ADD COLUMN repo_root TEXT;
```

Nullable. Existing sessions get `repo_root = NULL`, shown under "Other" group.

Populated at creation time:
- Worktree mode: from selected repo's `repo_root`
- Existing directory mode: auto-detect via `git rev-parse --show-toplevel`, match against config

## `grove repo add` Command

**Invocation:** `grove repo add [--repo-root <path>] [--worktree-base <path>] [--ai-tool <tool>] [--setup-commands <cmds>]`

**Interactive wizard (when flags omitted):**
1. Detect repo root from cwd (via `GetMainRepoPath`), confirm
2. Prompt for worktree base (default: `<repo-parent>/<repo-name>-worktrees`)
3. Prompt for AI tool override (default: empty, inherit global)
4. Prompt for setup commands override (default: empty, inherit global)

**CLI mode:** Flag-provided values skip their corresponding prompts.

**Duplicate check:** Error if `repo_root` already registered.

**`grove repo list`:** TUI-styled table of registered repos showing repo root, worktree base, and any overrides.

Full Bubbletea TUI components — lipgloss-styled, consistent with existing wizard/list.

## Session List Changes

**Single table with REPO column:**

```
  NAME          REPO      TOOL      DIRECTORY                  AGE
  ● auth-fix     fermat    claude    ~/Projects/Work/fe~       2h 3m
  ● repo-scoping grove     claude    ~/Projects/Work/gr~       12m
  ○ scratch      —         claude    ~/tmp/experiment          3d 1h
```

- REPO column: basename of `repo_root`, `—` for NULL
- Sorted by repo (config order, NULL last), then `created_at` desc within repo

**Filtering (ctrl+f):**
- Filter bar replaces status bar, text input
- Case-insensitive substring match across name, repo, tool, directory
- `esc` clears filter and closes bar
- `enter` closes bar, keeps active filter (shown dimmed in status bar)

## Create Session Wizard Changes

**Worktree mode gains a repo selection step:**

1. Dir source (existing)
2. **Repo selection** (new, worktree only) — auto-detect from cwd if matches registered repo, otherwise show picker. If no repos registered: `"No repos configured yet. Run: grove repo add"`
3. Branch name (existing)
4. Tool — pre-selected from repo override if set
5. Prompt (existing)
6. Confirm — shows repo name and worktree base

**Existing directory mode:** auto-resolve `repo_root` silently via git, no picker.

## Error Handling

| Scenario | Behavior |
|----------|----------|
| No config (first run) | Existing wizard, no change |
| No repos + worktree mode in TUI | Show message in wizard, return to dir source |
| No repos + CLI worktree | Stderr message, exit 1 |
| `grove repo add` from non-git dir without `--repo-root` | Clear error with instructions |
| Worktree base doesn't exist at session creation | Auto-create (`mkdir -p`) |
| Session's repo_root removed from config | Session displays, repo column shows basename, can't create new worktrees |

## Changes by Package

- **`internal/config/`** — `RepoConfig` struct, `Repos` field, lookup/cascade helpers, `AddRepo`
- **`internal/store/`** — Migration v3, `repo_root` column, updated `CreateSession`/`Session`
- **`internal/app/create.go`** — `stepRepoSelect`, repo picker, repo-aware worktree creation
- **`internal/app/list.go`** — REPO column, sorting by repo, filter bar with `ctrl+f`
- **`internal/app/repo_add.go`** (new) — Bubbletea wizard for `grove repo add`
- **`cmd/grove/main.go`** — `repo add`/`repo list` subcommands with flag+interactive modes
- **`internal/worktree/`** — No changes needed
