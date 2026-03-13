package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/app"
	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/worktree"
)

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "grove",
	Short:   "AI-powered tmux session manager",
	Version: version,
	// When invoked with no subcommand, launch the TUI.
	Run: func(cmd *cobra.Command, args []string) {
		launchTUI()
	},
	SilenceUsage: true,
}

// ── new ─────────────────────────────────────────────────────────────────

var newFlags struct {
	tool   string
	dir    string
	prompt string
	plan   string
	name   string
}

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new AI coding session",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := newFlags.dir
		if dir == "" {
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error getting working directory: %w", err)
			}
			dir = wd
		}

		name := newFlags.name
		if name == "" {
			name = newFlags.tool
		}

		mgr, s := openManager()
		defer s.Close()

		sess, err := mgr.Create(name, newFlags.tool, dir, nil, newFlags.prompt, newFlags.plan, nil)
		if err != nil {
			return fmt.Errorf("error creating session: %w", err)
		}

		fmt.Printf("Created session %s (%s)\n", sess.ID, sess.TmuxSession)
		return nil
	},
}

// ── list ────────────────────────────────────────────────────────────────

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, s := openManager()
		defer s.Close()

		sessions, err := s.ListSessions()
		if err != nil {
			return fmt.Errorf("error listing sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTOOL\tDIRECTORY\tSTATUS\tCREATED")
		for _, sess := range sessions {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				sess.ID, sess.Name, sess.Tool, sess.Directory,
				sess.Status, sess.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		w.Flush()
		return nil
	},
}

// ── attach ──────────────────────────────────────────────────────────────

var attachCmd = &cobra.Command{
	Use:   "attach <session-id>",
	Short: "Attach to an existing session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]

		mgr, s := openManager()
		defer s.Close()

		tmuxCmd, err := mgr.Attach(sessionID)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}

		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return fmt.Errorf("error: tmux not found in PATH: %w", err)
		}

		if err := syscall.Exec(tmuxPath, tmuxCmd.Args, os.Environ()); err != nil {
			return fmt.Errorf("error attaching: %w", err)
		}
		return nil
	},
}

// ── repo ────────────────────────────────────────────────────────────────

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository configurations",
}

var repoAddFlags struct {
	repoRoot      string
	worktreeBase  string
	aiTool        string
	setupCommands string
}

var repoAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}

		root := repoAddFlags.repoRoot
		if root == "" {
			detected, err := worktree.GetMainRepoPath(".")
			if err != nil {
				return fmt.Errorf("error detecting repo root (use --repo-root): %w", err)
			}
			root = detected
		}

		// If no flags provided, launch TUI wizard.
		if repoAddFlags.worktreeBase == "" && repoAddFlags.aiTool == "" && repoAddFlags.setupCommands == "" {
			repo, err := app.RunRepoAddTUI(cfg, root)
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}
			if repo == nil {
				return nil
			}
			if err := cfg.AddRepo(*repo); err != nil {
				return fmt.Errorf("error saving repo: %w", err)
			}
			fmt.Printf("Added repo %s\n", repo.RepoRoot)
			return nil
		}

		// Non-interactive mode: build RepoConfig from flags.
		repo := config.RepoConfig{
			RepoRoot:     root,
			WorktreeBase: repoAddFlags.worktreeBase,
			AITool:       repoAddFlags.aiTool,
		}
		if repoAddFlags.setupCommands != "" {
			for _, cmd := range strings.Split(repoAddFlags.setupCommands, ",") {
				trimmed := strings.TrimSpace(cmd)
				if trimmed != "" {
					repo.SetupCommands = append(repo.SetupCommands, trimmed)
				}
			}
		}

		if err := cfg.AddRepo(repo); err != nil {
			return fmt.Errorf("error saving repo: %w", err)
		}
		fmt.Printf("Added repo %s\n", repo.RepoRoot)
		return nil
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("error loading config: %w", err)
		}
		app.PrintRepoList(cfg)
		return nil
	},
}

// ── update ──────────────────────────────────────────────────────────────

const installURL = "https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update grove to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("grove %s — checking for updates...\n", version)

		shell := "sh"
		c := exec.Command(shell, "-c", fmt.Sprintf("curl -fsSL %s | GROVE_INSTALL_COLOR=1 sh", installURL))
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin
		if err := c.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	},
}

// ── init & helpers ──────────────────────────────────────────────────────

func init() {
	// Disable Cobra's default completion command.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	newCmd.Flags().StringVar(&newFlags.tool, "tool", "", "AI tool to use (e.g. claude, codex)")
	newCmd.Flags().StringVar(&newFlags.dir, "dir", "", "working directory for the session")
	newCmd.Flags().StringVar(&newFlags.prompt, "prompt", "", "prompt text")
	newCmd.Flags().StringVar(&newFlags.plan, "plan", "", "path to plan file")
	newCmd.Flags().StringVar(&newFlags.name, "name", "", "session name")
	_ = newCmd.MarkFlagRequired("tool")

	repoAddCmd.Flags().StringVar(&repoAddFlags.repoRoot, "repo-root", "", "repository root path (default: detected from cwd)")
	repoAddCmd.Flags().StringVar(&repoAddFlags.worktreeBase, "worktree-base", "", "worktree base directory")
	repoAddCmd.Flags().StringVar(&repoAddFlags.aiTool, "ai-tool", "", "AI tool override")
	repoAddCmd.Flags().StringVar(&repoAddFlags.setupCommands, "setup-commands", "", "comma-separated setup commands")

	repoCmd.AddCommand(repoAddCmd, repoListCmd)
	rootCmd.AddCommand(newCmd, listCmd, attachCmd, repoCmd, updateCmd)
}

func openManager() (*session.Manager, *store.Store) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening store: %v\n", err)
		os.Exit(1)
	}

	return session.NewManager(s, cfg), s
}

func launchTUI() {
	if !config.ConfigExists() {
		fmt.Println("Welcome to Grove! Let's set up your configuration.")
		fmt.Println()
		_, err := config.RunWizard(os.Stdin, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error running setup wizard: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()
		fmt.Println("Configuration saved. Launching Grove...")
		fmt.Println()
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	s, err := store.Open(store.DefaultDBPath())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	mgr := session.NewManager(s, cfg)

	p := tea.NewProgram(app.New(s, cfg, mgr), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
