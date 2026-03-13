# Grove

A Bubbletea TUI for managing AI coding sessions with tmux isolation.

Grove wraps AI coding tools (Claude, OpenCode, Codex) in tmux sessions with a
visual interface for creating, monitoring, attaching, and managing multiple
concurrent sessions. Each tool's native session management is preserved — Claude
sessions resume with `--resume`, OpenCode with `--continue`.

## Features

- **Session list** — status dots, tool, directory, age at a glance
- **Live peek** — 500ms tmux capture-pane polling with ANSI passthrough
- **New session wizard** — existing directory or git worktree, tool selection, prompt
- **Native session resume** — uses each tool's own session management
- **Worktree management** — create, prune with dirty check
- **Reconciliation** — syncs DB with tmux reality every 5s

## Install

```bash
curl -sSL https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh | sh
```

Specific version or custom directory:

```bash
VERSION=v0.2.0 curl -sSL https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh | sh
INSTALL_DIR=~/.local/bin curl -sSL https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh | sh
```

Update to latest:

```bash
grove update
```

## Usage

```bash
grove                    # Launch TUI (default)
grove new --tool claude --dir . --prompt "fix the bug"
grove list               # Tab-separated session list
grove attach <id>        # Attach to session directly
grove repo add           # Add repo config (TUI wizard)
grove repo list          # List configured repos
grove update             # Self-update to latest release
grove --version          # Print version
grove <command> --help   # Help for any command
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `enter` | Attach to session |
| `p` | Peek (live tmux preview) |
| `n` | New session wizard |
| `d` | Delete session |
| `s` | Stop session |
| `r` | Resume session |
| `x` | Prune worktree |
| `?` | Help |
| `q` | Quit |

## Config

First run triggers a setup wizard. Config lives at `~/.config/grove/config.toml`:

```toml
[defaults]
ai_tool = "claude"
worktree_base = "~/Projects/Work"

[worktree]
setup_commands = ["npm install"]

[tools.claude]
command = "claude"

[tools.codex]
command = "codex"

[tools.opencode]
command = "opencode"
```

## Dependencies

- Go 1.24+
- tmux
- At least one AI coding tool (claude, opencode, codex)
