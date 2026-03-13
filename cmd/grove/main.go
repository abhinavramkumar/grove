package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/tabwriter"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/app"
	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
	"github.com/abhinav/grove/internal/worktree"
)

var version = "dev"

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "--version" {
		fmt.Println("grove " + version)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		launchTUI()
		return
	}

	switch os.Args[1] {
	case "new":
		cmdNew(os.Args[2:])
	case "list":
		cmdList(os.Args[2:])
	case "attach":
		cmdAttach(os.Args[2:])
	case "repo":
		cmdRepo(os.Args[2:])
	case "update":
		cmdUpdate()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "usage: grove [new|list|attach|repo|update]")
		os.Exit(1)
	}
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

func cmdNew(args []string) {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	tool := fs.String("tool", "", "AI tool to use (e.g. claude, codex)")
	dir := fs.String("dir", "", "working directory for the session")
	prompt := fs.String("prompt", "", "prompt text")
	plan := fs.String("plan", "", "path to plan file")
	name := fs.String("name", "", "session name")
	fs.Parse(args)

	if *tool == "" {
		fmt.Fprintln(os.Stderr, "error: --tool is required")
		os.Exit(1)
	}

	if *dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
			os.Exit(1)
		}
		*dir = wd
	}

	if *name == "" {
		*name = *tool
	}

	mgr, s := openManager()
	defer s.Close()

	sess, err := mgr.Create(*name, *tool, *dir, nil, *prompt, *plan, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Created session %s (%s)\n", sess.ID, sess.TmuxSession)
}

func cmdList(args []string) {
	_ = args // no flags

	_, s := openManager()
	defer s.Close()

	sessions, err := s.ListSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error listing sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tTOOL\tDIRECTORY\tSTATUS\tCREATED")
	for _, sess := range sessions {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			sess.ID, sess.Name, sess.Tool, sess.Directory,
			sess.Status, sess.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	w.Flush()
}

func cmdAttach(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: grove attach <session-id>")
		os.Exit(1)
	}
	sessionID := args[0]

	mgr, s := openManager()
	defer s.Close()

	cmd, err := mgr.Attach(sessionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Replace the current process with tmux attach via syscall.Exec.
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: tmux not found in PATH: %v\n", err)
		os.Exit(1)
	}

	if err := syscall.Exec(tmuxPath, cmd.Args, os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "error attaching: %v\n", err)
		os.Exit(1)
	}
}

const installURL = "https://raw.githubusercontent.com/abhinavramkumar/grove/main/install.sh"

func cmdUpdate() {
	fmt.Printf("grove %s — checking for updates...\n", version)

	// Download and run install.sh, which handles version check + upgrade.
	shell := "sh"
	cmd := exec.Command(shell, "-c", fmt.Sprintf("curl -fsSL %s | GROVE_INSTALL_COLOR=1 sh", installURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "update failed: %v\n", err)
		os.Exit(1)
	}
}

func cmdRepo(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: grove repo [add|list]")
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		cmdRepoAdd(args[1:])
	case "list":
		cmdRepoList()
	default:
		fmt.Fprintf(os.Stderr, "unknown repo subcommand: %s\n", args[0])
		fmt.Fprintln(os.Stderr, "usage: grove repo [add|list]")
		os.Exit(1)
	}
}

func cmdRepoAdd(args []string) {
	fs := flag.NewFlagSet("repo add", flag.ExitOnError)
	repoRoot := fs.String("repo-root", "", "repository root path (default: detected from cwd)")
	worktreeBase := fs.String("worktree-base", "", "worktree base directory")
	aiTool := fs.String("ai-tool", "", "AI tool override")
	setupCommands := fs.String("setup-commands", "", "comma-separated setup commands")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Detect repo root from cwd if not provided.
	root := *repoRoot
	if root == "" {
		detected, err := worktree.GetMainRepoPath(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error detecting repo root (use --repo-root): %v\n", err)
			os.Exit(1)
		}
		root = detected
	}

	// If worktree-base is not provided, launch TUI wizard.
	if *worktreeBase == "" && *aiTool == "" && *setupCommands == "" {
		repo, err := app.RunRepoAddTUI(cfg, root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if repo == nil {
			os.Exit(0)
		}
		if err := cfg.AddRepo(*repo); err != nil {
			fmt.Fprintf(os.Stderr, "error saving repo: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added repo %s\n", repo.RepoRoot)
		return
	}

	// Non-interactive mode: build RepoConfig from flags.
	repo := config.RepoConfig{
		RepoRoot:     root,
		WorktreeBase: *worktreeBase,
		AITool:       *aiTool,
	}
	if *setupCommands != "" {
		for _, cmd := range strings.Split(*setupCommands, ",") {
			trimmed := strings.TrimSpace(cmd)
			if trimmed != "" {
				repo.SetupCommands = append(repo.SetupCommands, trimmed)
			}
		}
	}

	if err := cfg.AddRepo(repo); err != nil {
		fmt.Fprintf(os.Stderr, "error saving repo: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Added repo %s\n", repo.RepoRoot)
}

func cmdRepoList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}
	app.PrintRepoList(cfg)
}

func launchTUI() {
	// First-run: if no config exists, run the setup wizard.
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
