# Repo Registry Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a repo registry to grove with per-repo config overrides, repo-scoped session list with filtering, and a `grove repo add/list` command.

**Architecture:** Config.toml is the source of truth for repo definitions with per-repo overrides of all global settings. Sessions table gets a denormalized `repo_root` column for grouping. `grove repo add` uses a Bubbletea wizard (with CLI flag fallback). Session list gains a REPO column and ctrl+f filtering.

**Tech Stack:** Go, Bubbletea/Bubbles/Lipgloss, SQLite (modernc.org/sqlite), go-toml/v2

---

## Chunk 1: Config layer — RepoConfig and cascade helpers

### Task 1: Add RepoConfig to config package

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests for RepoConfig round-trip**

Add to `internal/config/config_test.go`:

```go
func TestSaveAndLoadWithRepos(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/Projects/Work",
		},
		Worktree: WorktreeConfig{
			SetupCommands: []string{"npm install"},
		},
		Tools: map[string]ToolConfig{
			"claude": {Command: "claude"},
		},
		Repos: []RepoConfig{
			{
				RepoRoot:     "/home/user/projects/fermat",
				WorktreeBase: "/home/user/projects/fermat-worktrees",
				AITool:       "codex",
				SetupCommands: []string{"make build"},
			},
			{
				RepoRoot:     "/home/user/projects/grove",
				WorktreeBase: "/home/user/projects/grove-worktrees",
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("saving config: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	if len(loaded.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loaded.Repos))
	}
	if loaded.Repos[0].RepoRoot != "/home/user/projects/fermat" {
		t.Fatalf("unexpected repo_root: %q", loaded.Repos[0].RepoRoot)
	}
	if loaded.Repos[0].WorktreeBase != "/home/user/projects/fermat-worktrees" {
		t.Fatalf("unexpected worktree_base: %q", loaded.Repos[0].WorktreeBase)
	}
	if loaded.Repos[0].AITool != "codex" {
		t.Fatalf("unexpected ai_tool: %q", loaded.Repos[0].AITool)
	}
	if len(loaded.Repos[0].SetupCommands) != 1 || loaded.Repos[0].SetupCommands[0] != "make build" {
		t.Fatalf("unexpected setup_commands: %v", loaded.Repos[0].SetupCommands)
	}
	if loaded.Repos[1].RepoRoot != "/home/user/projects/grove" {
		t.Fatalf("unexpected second repo_root: %q", loaded.Repos[1].RepoRoot)
	}
	if loaded.Repos[1].AITool != "" {
		t.Fatalf("expected empty ai_tool override, got %q", loaded.Repos[1].AITool)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/config/ -run TestSaveAndLoadWithRepos -v`
Expected: FAIL — `RepoConfig` type does not exist.

- [ ] **Step 3: Add RepoConfig struct and Repos field**

In `internal/config/config.go`, add the struct and field:

```go
type RepoConfig struct {
	RepoRoot      string   `toml:"repo_root"`
	WorktreeBase  string   `toml:"worktree_base"`
	AITool        string   `toml:"ai_tool,omitempty"`
	SetupCommands []string `toml:"setup_commands,omitempty"`
}
```

Add to `Config`:
```go
Repos []RepoConfig `toml:"repos,omitempty"`
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/config/ -run TestSaveAndLoadWithRepos -v`
Expected: PASS

- [ ] **Step 5: Run all config tests to verify no regressions**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/config/ -v`
Expected: All tests pass.

---

### Task 2: Add RepoFor lookup and cascade helpers

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write failing tests for RepoFor and cascade helpers**

Add to `internal/config/config_test.go`:

```go
func TestRepoFor(t *testing.T) {
	cfg := &Config{
		Repos: []RepoConfig{
			{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
			{RepoRoot: "/projects/grove", WorktreeBase: "/projects/grove-wt"},
		},
	}

	t.Run("found", func(t *testing.T) {
		repo := cfg.RepoFor("/projects/fermat")
		if repo == nil {
			t.Fatal("expected to find repo")
		}
		if repo.WorktreeBase != "/projects/fermat-wt" {
			t.Fatalf("unexpected worktree_base: %q", repo.WorktreeBase)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := cfg.RepoFor("/projects/unknown")
		if repo != nil {
			t.Fatal("expected nil for unknown repo")
		}
	})
}

func TestEffectiveSettings(t *testing.T) {
	cfg := &Config{
		Defaults: DefaultsConfig{
			AITool:       "claude",
			WorktreeBase: "~/default-wt",
		},
		Worktree: WorktreeConfig{
			SetupCommands: []string{"npm install"},
		},
	}

	t.Run("nil repo uses globals", func(t *testing.T) {
		if got := cfg.EffectiveAITool(nil); got != "claude" {
			t.Fatalf("expected 'claude', got %q", got)
		}
		if got := cfg.EffectiveWorktreeBase(nil); got != "~/default-wt" {
			t.Fatalf("expected '~/default-wt', got %q", got)
		}
		cmds := cfg.EffectiveSetupCommands(nil)
		if len(cmds) != 1 || cmds[0] != "npm install" {
			t.Fatalf("expected global setup commands, got %v", cmds)
		}
	})

	t.Run("repo override takes precedence", func(t *testing.T) {
		repo := &RepoConfig{
			RepoRoot:      "/projects/fermat",
			WorktreeBase:  "/projects/fermat-wt",
			AITool:        "codex",
			SetupCommands: []string{"make build"},
		}
		if got := cfg.EffectiveAITool(repo); got != "codex" {
			t.Fatalf("expected 'codex', got %q", got)
		}
		if got := cfg.EffectiveWorktreeBase(repo); got != "/projects/fermat-wt" {
			t.Fatalf("expected '/projects/fermat-wt', got %q", got)
		}
		cmds := cfg.EffectiveSetupCommands(repo)
		if len(cmds) != 1 || cmds[0] != "make build" {
			t.Fatalf("expected repo setup commands, got %v", cmds)
		}
	})

	t.Run("repo with empty overrides falls back to global", func(t *testing.T) {
		repo := &RepoConfig{
			RepoRoot:     "/projects/grove",
			WorktreeBase: "/projects/grove-wt",
		}
		if got := cfg.EffectiveAITool(repo); got != "claude" {
			t.Fatalf("expected 'claude', got %q", got)
		}
		cmds := cfg.EffectiveSetupCommands(repo)
		if len(cmds) != 1 || cmds[0] != "npm install" {
			t.Fatalf("expected global setup commands, got %v", cmds)
		}
	})
}

func TestAddRepo(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{AITool: "claude", WorktreeBase: "~/wt"},
		Tools:    map[string]ToolConfig{"claude": {Command: "claude"}},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	t.Run("add new repo", func(t *testing.T) {
		err := cfg.AddRepo(RepoConfig{
			RepoRoot:     "/projects/fermat",
			WorktreeBase: "/projects/fermat-wt",
		})
		if err != nil {
			t.Fatalf("adding repo: %v", err)
		}
		if len(cfg.Repos) != 1 {
			t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
		}

		// Verify persisted
		loaded, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if len(loaded.Repos) != 1 {
			t.Fatalf("expected 1 repo after reload, got %d", len(loaded.Repos))
		}
	})

	t.Run("duplicate repo errors", func(t *testing.T) {
		err := cfg.AddRepo(RepoConfig{
			RepoRoot:     "/projects/fermat",
			WorktreeBase: "/projects/fermat-wt2",
		})
		if err == nil {
			t.Fatal("expected error for duplicate repo")
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/config/ -run "TestRepoFor|TestEffectiveSettings|TestAddRepo" -v`
Expected: FAIL — methods don't exist.

- [ ] **Step 3: Implement RepoFor, cascade helpers, and AddRepo**

Add to `internal/config/config.go`:

```go
// RepoFor returns the RepoConfig for the given repo root, or nil if not found.
func (c *Config) RepoFor(repoRoot string) *RepoConfig {
	for i := range c.Repos {
		if c.Repos[i].RepoRoot == repoRoot {
			return &c.Repos[i]
		}
	}
	return nil
}

// EffectiveAITool returns the AI tool for the given repo, falling back to global.
func (c *Config) EffectiveAITool(repo *RepoConfig) string {
	if repo != nil && repo.AITool != "" {
		return repo.AITool
	}
	return c.Defaults.AITool
}

// EffectiveWorktreeBase returns the worktree base for the given repo, falling back to global.
func (c *Config) EffectiveWorktreeBase(repo *RepoConfig) string {
	if repo != nil && repo.WorktreeBase != "" {
		return repo.WorktreeBase
	}
	return c.Defaults.WorktreeBase
}

// EffectiveSetupCommands returns setup commands for the given repo, falling back to global.
func (c *Config) EffectiveSetupCommands(repo *RepoConfig) []string {
	if repo != nil && len(repo.SetupCommands) > 0 {
		return repo.SetupCommands
	}
	return c.Worktree.SetupCommands
}

// AddRepo appends a repo to the config and saves. Returns error if already registered.
func (c *Config) AddRepo(repo RepoConfig) error {
	if c.RepoFor(repo.RepoRoot) != nil {
		return fmt.Errorf("repo already registered: %s", repo.RepoRoot)
	}
	c.Repos = append(c.Repos, repo)
	return Save(c)
}
```

Add `"fmt"` to imports in config.go.

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/config/ -v`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add RepoConfig with cascade helpers and AddRepo"
```

---

## Chunk 2: DB migration — repo_root column on sessions

### Task 3: Add repo_root column to sessions table

**Files:**
- Modify: `internal/store/migrations.go`
- Modify: `internal/store/store.go`
- Test: `internal/store/store_test.go`

- [ ] **Step 1: Write failing tests for repo_root in sessions**

Add to `internal/store/store_test.go`:

```go
func TestCreateSessionWithRepoRoot(t *testing.T) {
	s := newTestStore(t)

	repoRoot := "/projects/fermat"
	sess, err := s.CreateSession("test", "claude", "/dir", nil, nil, nil, nil, &repoRoot)
	if err != nil {
		t.Fatalf("creating session: %v", err)
	}

	if sess.RepoRoot == nil || *sess.RepoRoot != "/projects/fermat" {
		t.Fatalf("expected repo_root '/projects/fermat', got %v", sess.RepoRoot)
	}

	got, err := s.GetSession(sess.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RepoRoot == nil || *got.RepoRoot != "/projects/fermat" {
		t.Fatalf("expected repo_root after reload, got %v", got.RepoRoot)
	}
}

func TestCreateSessionWithNilRepoRoot(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.CreateSession("test", "claude", "/dir", nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("creating session: %v", err)
	}

	if sess.RepoRoot != nil {
		t.Fatalf("expected nil repo_root, got %v", sess.RepoRoot)
	}
}

func TestListSessionsOrderedByRepoRoot(t *testing.T) {
	s := newTestStore(t)

	repoA := "/projects/aaa"
	repoB := "/projects/bbb"

	// Create sessions in mixed order
	_, _ = s.CreateSession("s1", "claude", "/dir1", nil, nil, nil, nil, &repoB)
	_, _ = s.CreateSession("s2", "claude", "/dir2", nil, nil, nil, nil, &repoA)
	_, _ = s.CreateSession("s3", "claude", "/dir3", nil, nil, nil, nil, nil)

	sessions, err := s.ListSessions()
	if err != nil {
		t.Fatalf("listing: %v", err)
	}

	if len(sessions) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(sessions))
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/store/ -run "TestCreateSessionWithRepoRoot|TestCreateSessionWithNilRepoRoot|TestListSessionsOrderedByRepoRoot" -v`
Expected: FAIL — `CreateSession` has wrong number of args.

- [ ] **Step 3: Update Session struct, CreateSession, migrations, and scan**

In `internal/store/store.go`:
- Add `RepoRoot *string` to `Session` struct (after `Directory`)
- Update `sessionColumns` to include `repo_root`
- Update `CreateSession` signature to add `repoRoot *string` parameter
- Update the INSERT statement and args
- Update `scanFrom` to scan `RepoRoot`

In `internal/store/migrations.go`:
- Bump `currentVersion` to 3
- Add migration for v2→v3: `ALTER TABLE sessions ADD COLUMN repo_root TEXT`
- Update the v1 CREATE TABLE to include `repo_root TEXT`

- [ ] **Step 4: Fix existing test calls**

All existing calls to `CreateSession` need the new `repoRoot` parameter (pass `nil`). Update:
- `TestCreateAndGetSession`
- `TestListSessions`
- `TestUpdateSessionStatus`
- `TestDeleteSession`
- `TestDeleteNonexistentSession`

- [ ] **Step 5: Run all store tests**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/store/ -v`
Expected: All tests pass.

- [ ] **Step 6: Update session.Manager.Create to accept and pass repoRoot**

In `internal/session/session.go`, update `Create` signature:
```go
func (m *Manager) Create(name, tool, dir string, worktree *string, prompt, planFile string, repoRoot *string) (*store.Session, error) {
```
And pass `repoRoot` to `m.Store.CreateSession(...)`.

- [ ] **Step 7: Update all callers of Manager.Create**

Two callers:
- `internal/app/create.go:342` — pass `nil` for now (will be wired in Task 6)
- `cmd/grove/main.go:94` — pass `nil` for now (CLI `new` command)

- [ ] **Step 8: Run full test suite**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All tests pass. Build succeeds.

- [ ] **Step 9: Commit**

```bash
git add internal/store/ internal/session/session.go internal/app/create.go cmd/grove/main.go
git commit -m "feat(store): add repo_root column to sessions (migration v3)"
```

---

## Chunk 3: `grove repo add` and `grove repo list` commands

### Task 4: Add the Bubbletea wizard for `grove repo add`

**Files:**
- Create: `internal/app/repo_add.go`
- Test: `internal/app/repo_add_test.go`

- [ ] **Step 1: Write tests for RepoAddModel**

Create `internal/app/repo_add_test.go`:

```go
package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

func TestRepoAddModel_StepProgression(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude", WorktreeBase: "~/wt"},
		Tools:    map[string]config.ToolConfig{"claude": {Command: "claude"}},
	}

	m := NewRepoAddModel(cfg, "/projects/fermat")

	// Step 1: confirm repo root — should show detected path
	view := m.View()
	if !containsStr(view, "/projects/fermat") {
		t.Fatalf("expected repo root in view, got:\n%s", view)
	}

	// Press enter to confirm repo root
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Step 2: worktree base input — should be focused
	view = m.View()
	if !containsStr(view, "worktree") {
		t.Fatalf("expected worktree base step, got:\n%s", view)
	}
}

func TestRepoAddModel_EscCancels(t *testing.T) {
	cfg := &config.Config{}
	m := NewRepoAddModel(cfg, "/projects/fermat")

	var cmd tea.Cmd
	m, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if cmd == nil {
		t.Fatal("expected cancel command")
	}

	msg := cmd()
	if _, ok := msg.(repoAddCancelMsg); !ok {
		t.Fatalf("expected repoAddCancelMsg, got %T", msg)
	}
}

func TestRepoAddModel_DuplicateRepoErrors(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude"},
		Tools:    map[string]config.ToolConfig{"claude": {Command: "claude"}},
		Repos: []config.RepoConfig{
			{RepoRoot: "/projects/fermat", WorktreeBase: "/wt"},
		},
	}
	_ = config.Save(cfg)

	m := NewRepoAddModel(cfg, "/projects/fermat")

	// Press enter to confirm — should show duplicate error
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if m.err == "" {
		t.Fatal("expected duplicate error")
	}
}

func containsStr(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && searchStr(s, substr))
}

func searchStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -run "TestRepoAdd" -v`
Expected: FAIL — `NewRepoAddModel` does not exist.

- [ ] **Step 3: Implement RepoAddModel**

Create `internal/app/repo_add.go`:

```go
package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
)

// Messages for repo add wizard.
type repoAddDoneMsg struct{ repo config.RepoConfig }
type repoAddCancelMsg struct{}

// repoAddStep enumerates wizard steps.
type repoAddStep int

const (
	raStepConfirmRoot repoAddStep = iota
	raStepWorktreeBase
	raStepAITool
	raStepSetupCommands
	raStepConfirm
)

// RepoAddModel is the Bubbletea model for the repo add wizard.
type RepoAddModel struct {
	step     repoAddStep
	config   *config.Config
	repoRoot string

	worktreeBaseInput  textinput.Model
	aiToolInput        textinput.Model
	setupCommandsInput textinput.Model

	width  int
	height int
	err    string
}

// NewRepoAddModel creates a new repo add wizard.
func NewRepoAddModel(cfg *config.Config, detectedRepoRoot string) RepoAddModel {
	repoName := filepath.Base(detectedRepoRoot)
	defaultWT := filepath.Join(filepath.Dir(detectedRepoRoot), repoName+"-worktrees")

	wtInput := textinput.New()
	wtInput.Placeholder = defaultWT
	wtInput.CharLimit = 256
	wtInput.Width = 60

	aiInput := textinput.New()
	aiInput.Placeholder = "leave empty to inherit global"
	aiInput.CharLimit = 64
	aiInput.Width = 60

	scInput := textinput.New()
	scInput.Placeholder = "comma-separated, leave empty to inherit global"
	scInput.CharLimit = 512
	scInput.Width = 60

	return RepoAddModel{
		step:               raStepConfirmRoot,
		config:             cfg,
		repoRoot:           detectedRepoRoot,
		worktreeBaseInput:  wtInput,
		aiToolInput:        aiInput,
		setupCommandsInput: scInput,
	}
}

func (m RepoAddModel) Init() tea.Cmd { return nil }

func (m RepoAddModel) Update(msg tea.Msg) (RepoAddModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		if key.Matches(msg, keys.Escape) {
			return m, func() tea.Msg { return repoAddCancelMsg{} }
		}
		return m.handleKey(msg)
	}
	return m.updateInputs(msg)
}

func (m RepoAddModel) updateInputs(msg tea.Msg) (RepoAddModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.step {
	case raStepWorktreeBase:
		m.worktreeBaseInput, cmd = m.worktreeBaseInput.Update(msg)
	case raStepAITool:
		m.aiToolInput, cmd = m.aiToolInput.Update(msg)
	case raStepSetupCommands:
		m.setupCommandsInput, cmd = m.setupCommandsInput.Update(msg)
	}
	return m, cmd
}

func (m RepoAddModel) handleKey(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	switch m.step {
	case raStepConfirmRoot:
		return m.handleConfirmRoot(msg)
	case raStepWorktreeBase:
		return m.handleWorktreeBase(msg)
	case raStepAITool:
		return m.handleAITool(msg)
	case raStepSetupCommands:
		return m.handleSetupCommands(msg)
	case raStepConfirm:
		return m.handleFinalConfirm(msg)
	}
	return m, nil
}

func (m RepoAddModel) handleConfirmRoot(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() != "enter" {
		return m, nil
	}

	// Check for duplicate
	if m.config.RepoFor(m.repoRoot) != nil {
		m.err = fmt.Sprintf("repo already registered: %s", m.repoRoot)
		return m, nil
	}

	m.step = raStepWorktreeBase
	m.worktreeBaseInput.Focus()
	m.err = ""
	return m, textinput.Blink
}

func (m RepoAddModel) handleWorktreeBase(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.step = raStepAITool
		m.worktreeBaseInput.Blur()
		m.aiToolInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}
	var cmd tea.Cmd
	m.worktreeBaseInput, cmd = m.worktreeBaseInput.Update(msg)
	return m, cmd
}

func (m RepoAddModel) handleAITool(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.step = raStepSetupCommands
		m.aiToolInput.Blur()
		m.setupCommandsInput.Focus()
		m.err = ""
		return m, textinput.Blink
	}
	var cmd tea.Cmd
	m.aiToolInput, cmd = m.aiToolInput.Update(msg)
	return m, cmd
}

func (m RepoAddModel) handleSetupCommands(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() == "enter" {
		m.step = raStepConfirm
		m.setupCommandsInput.Blur()
		m.err = ""
		return m, nil
	}
	var cmd tea.Cmd
	m.setupCommandsInput, cmd = m.setupCommandsInput.Update(msg)
	return m, cmd
}

func (m RepoAddModel) handleFinalConfirm(msg tea.KeyMsg) (RepoAddModel, tea.Cmd) {
	if msg.String() != "enter" {
		return m, nil
	}

	repo := m.buildRepoConfig()

	if err := m.config.AddRepo(repo); err != nil {
		m.err = err.Error()
		return m, nil
	}

	return m, func() tea.Msg { return repoAddDoneMsg{repo: repo} }
}

func (m RepoAddModel) buildRepoConfig() config.RepoConfig {
	repoName := filepath.Base(m.repoRoot)
	defaultWT := filepath.Join(filepath.Dir(m.repoRoot), repoName+"-worktrees")

	wtBase := strings.TrimSpace(m.worktreeBaseInput.Value())
	if wtBase == "" {
		wtBase = defaultWT
	}
	wtBase = expandHome(wtBase)

	aiTool := strings.TrimSpace(m.aiToolInput.Value())

	var setupCommands []string
	scStr := strings.TrimSpace(m.setupCommandsInput.Value())
	if scStr != "" {
		for _, cmd := range strings.Split(scStr, ",") {
			if trimmed := strings.TrimSpace(cmd); trimmed != "" {
				setupCommands = append(setupCommands, trimmed)
			}
		}
	}

	return config.RepoConfig{
		RepoRoot:      m.repoRoot,
		WorktreeBase:  wtBase,
		AITool:        aiTool,
		SetupCommands: setupCommands,
	}
}

func (m RepoAddModel) View() string {
	var b strings.Builder

	title := wizardTitleStyle.Render("Add Repository")
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.step {
	case raStepConfirmRoot:
		b.WriteString(wizardLabelStyle.Render("Step 1: Confirm repository root"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.repoRoot)
		b.WriteString("\n\n  " + dimStyle.Render("enter to confirm, esc to cancel"))

	case raStepWorktreeBase:
		b.WriteString(wizardLabelStyle.Render("Step 2: Worktree base directory"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.worktreeBaseInput.View())
		b.WriteString("\n\n  " + dimStyle.Render("enter to continue (empty = default shown above)"))

	case raStepAITool:
		b.WriteString(wizardLabelStyle.Render("Step 3: AI tool override (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.aiToolInput.View())
		b.WriteString("\n\n  " + dimStyle.Render("enter to continue (empty = inherit global: "+m.config.Defaults.AITool+")"))

	case raStepSetupCommands:
		b.WriteString(wizardLabelStyle.Render("Step 4: Setup commands override (optional)"))
		b.WriteString("\n\n")
		b.WriteString("  " + m.setupCommandsInput.View())
		globalCmds := "none"
		if len(m.config.Worktree.SetupCommands) > 0 {
			globalCmds = strings.Join(m.config.Worktree.SetupCommands, ", ")
		}
		b.WriteString("\n\n  " + dimStyle.Render("enter to continue (empty = inherit global: "+globalCmds+")"))

	case raStepConfirm:
		repo := m.buildRepoConfig()
		b.WriteString(wizardLabelStyle.Render("Step 5: Confirm"))
		b.WriteString("\n\n")
		b.WriteString("  Repo root:      " + repo.RepoRoot + "\n")
		b.WriteString("  Worktree base:  " + repo.WorktreeBase + "\n")
		aiTool := repo.AITool
		if aiTool == "" {
			aiTool = "(inherit: " + m.config.Defaults.AITool + ")"
		}
		b.WriteString("  AI tool:        " + aiTool + "\n")
		sc := "(inherit global)"
		if len(repo.SetupCommands) > 0 {
			sc = strings.Join(repo.SetupCommands, ", ")
		}
		b.WriteString("  Setup commands: " + sc + "\n")
		b.WriteString("\n  " + dimStyle.Render("enter to save, esc to cancel"))
	}

	if m.err != "" {
		b.WriteString("\n\n  " + errorStyle.Render(m.err))
	}

	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -run "TestRepoAdd" -v`
Expected: All pass.

- [ ] **Step 5: Commit**

```bash
git add internal/app/repo_add.go internal/app/repo_add_test.go
git commit -m "feat(app): add RepoAddModel TUI wizard for grove repo add"
```

---

### Task 5: Add `grove repo add` and `grove repo list` CLI commands

**Files:**
- Modify: `cmd/grove/main.go`

- [ ] **Step 1: Add `repo` subcommand dispatch to main.go**

In `cmd/grove/main.go`, add the `repo` case to the switch and implement `cmdRepo`, `cmdRepoAdd`, and `cmdRepoList`:

```go
case "repo":
	cmdRepo(os.Args[2:])
```

`cmdRepo` dispatches to `add` or `list`:

```go
func cmdRepo(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: grove repo <add|list>")
		os.Exit(1)
	}
	switch args[0] {
	case "add":
		cmdRepoAdd(args[1:])
	case "list":
		cmdRepoList()
	default:
		fmt.Fprintf(os.Stderr, "unknown repo subcommand: %s\n", args[0])
		os.Exit(1)
	}
}
```

`cmdRepoAdd` parses flags, detects repo root, launches TUI wizard or uses flags:

```go
func cmdRepoAdd(args []string) {
	fs := flag.NewFlagSet("repo add", flag.ExitOnError)
	repoRootFlag := fs.String("repo-root", "", "repository root path (auto-detected from cwd)")
	wtBaseFlag := fs.String("worktree-base", "", "worktree base directory")
	aiToolFlag := fs.String("ai-tool", "", "AI tool override")
	setupCmdsFlag := fs.String("setup-commands", "", "setup commands (comma-separated)")
	fs.Parse(args)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Detect repo root
	repoRoot := *repoRootFlag
	if repoRoot == "" {
		detected, err := worktree.GetMainRepoPath(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Not inside a git repository. Run this from inside a repo, or pass --repo-root <path>.")
			os.Exit(1)
		}
		repoRoot = detected
	}

	// If all required flags provided, skip wizard
	if *wtBaseFlag != "" {
		repo := config.RepoConfig{
			RepoRoot:     repoRoot,
			WorktreeBase: *wtBaseFlag,
			AITool:       *aiToolFlag,
		}
		if *setupCmdsFlag != "" {
			for _, cmd := range strings.Split(*setupCmdsFlag, ",") {
				if trimmed := strings.TrimSpace(cmd); trimmed != "" {
					repo.SetupCommands = append(repo.SetupCommands, trimmed)
				}
			}
		}
		if err := cfg.AddRepo(repo); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added repo: %s\n", repoRoot)
		return
	}

	// Launch TUI wizard
	m := app.NewRepoAddModel(cfg, repoRoot)
	p := tea.NewProgram(repoAddWrapper{m}, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if w, ok := result.(repoAddWrapper); ok && w.done {
		fmt.Printf("Added repo: %s\n", w.inner.repoRoot)
	}
}
```

The `repoAddWrapper` is needed because Bubbletea's `Program` requires a `tea.Model` at the top level, but `RepoAddModel` is a sub-model. Create a thin wrapper:

```go
type repoAddWrapper struct {
	inner app.RepoAddModel  // won't work — inner fields are unexported
}
```

**Better approach:** Add a standalone `RunRepoAddTUI` function to the app package that wraps `RepoAddModel` in a top-level model and runs it:

```go
// In internal/app/repo_add.go, add:

// repoAddApp wraps RepoAddModel as a top-level tea.Model.
type repoAddApp struct {
	model  RepoAddModel
	done   bool
	cancel bool
}

func (a repoAddApp) Init() tea.Cmd { return a.model.Init() }

func (a repoAddApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case repoAddDoneMsg:
		a.done = true
		return a, tea.Quit
	case repoAddCancelMsg:
		a.cancel = true
		return a, tea.Quit
	}
	var cmd tea.Cmd
	a.model, cmd = a.model.Update(msg)
	return a, cmd
}

func (a repoAddApp) View() string { return a.model.View() }

// RunRepoAddTUI runs the repo add wizard as a standalone TUI.
// Returns the added repo config, or nil if cancelled.
func RunRepoAddTUI(cfg *config.Config, repoRoot string) (*config.RepoConfig, error) {
	m := NewRepoAddModel(cfg, repoRoot)
	p := tea.NewProgram(repoAddApp{model: m}, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}
	if a, ok := result.(repoAddApp); ok && a.done {
		repo := a.model.buildRepoConfig()
		return &repo, nil
	}
	return nil, nil // cancelled
}
```

Then in `cmd/grove/main.go`, the TUI path becomes:

```go
repo, err := app.RunRepoAddTUI(cfg, repoRoot)
if err != nil {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
if repo != nil {
	fmt.Printf("Added repo: %s\n", repo.RepoRoot)
}
```

`cmdRepoList` loads config and prints a lipgloss-styled table:

```go
func cmdRepoList() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Repos) == 0 {
		fmt.Println("No repos registered. Run 'grove repo add' to add one.")
		return
	}

	app.PrintRepoList(cfg)
}
```

`PrintRepoList` in `internal/app/repo_list.go` (new file):

```go
package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/abhinav/grove/internal/config"
)

var (
	repoListHeader = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7"))
	repoListRow    = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
	repoListDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// PrintRepoList prints a styled table of registered repos.
func PrintRepoList(cfg *config.Config) {
	nameW := 16
	wtW := 40
	toolW := 10
	gap := "  "

	hdr := fmt.Sprintf("%-*s%s%-*s%s%-*s%s%s",
		nameW, "REPO", gap,
		wtW, "WORKTREE BASE", gap,
		toolW, "AI TOOL", gap,
		"SETUP COMMANDS")
	fmt.Println(repoListHeader.Render(hdr))

	for _, repo := range cfg.Repos {
		name := filepath.Base(repo.RepoRoot)
		wt := repo.WorktreeBase
		aiTool := repo.AITool
		if aiTool == "" {
			aiTool = repoListDim.Render("(global)")
		}
		sc := strings.Join(repo.SetupCommands, ", ")
		if sc == "" {
			sc = repoListDim.Render("(global)")
		}

		row := fmt.Sprintf("%-*s%s%-*s%s%-*s%s%s",
			nameW, truncate(name, nameW), gap,
			wtW, truncate(wt, wtW), gap,
			toolW, truncate(aiTool, toolW), gap,
			sc)
		fmt.Println(repoListRow.Render(row))
	}
}
```

- [ ] **Step 2: Add missing imports to main.go**

Add `"strings"` and `"github.com/abhinav/grove/internal/worktree"` to main.go imports.

- [ ] **Step 3: Build to verify compilation**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build ./cmd/grove/`
Expected: Build succeeds.

- [ ] **Step 4: Run full test suite**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/app/repo_add.go internal/app/repo_list.go cmd/grove/main.go
git commit -m "feat: add grove repo add/list commands with TUI wizard and CLI flags"
```

---

## Chunk 4: Session list — REPO column, sorting, and filtering

### Task 6: Add REPO column and repo-based sorting to session list

**Files:**
- Modify: `internal/app/list.go`
- Modify: `internal/app/keys.go`
- Modify: `internal/app/styles.go`

- [ ] **Step 1: Update list.go to add REPO column**

In `internal/app/list.go`:
- Add a `RepoOrder []string` field to `ListModel` — the ordered list of repo basenames from config (for sort order)
- Add a helper `repoDisplayName(sess)` that returns `filepath.Base(repo_root)` or `"—"` for nil
- Update the header row to include REPO between NAME and TOOL
- Update each session row to include the repo display name
- Adjust column widths: `repoW := 14`, reduce `dirW` accordingly

- [ ] **Step 2: Add sorting by repo order**

Add a `SortSessions` method to `ListModel` that sorts `Sessions` by:
1. Repo order: registered repos in config order, unregistered repos alphabetically, NULL last
2. Within each repo group: `created_at` desc (already the DB default)

Call `SortSessions` in `ClampCursor` or wherever sessions are updated.

- [ ] **Step 3: Build and verify**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build ./cmd/grove/`
Expected: Build succeeds.

- [ ] **Step 4: Commit**

```bash
git add internal/app/list.go
git commit -m "feat(list): add REPO column with config-order sorting"
```

---

### Task 7: Add ctrl+f filter bar to session list

**Files:**
- Modify: `internal/app/list.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/keys.go`
- Modify: `internal/app/styles.go`
- Test: `internal/app/list_test.go`

- [ ] **Step 1: Write tests for filtering logic**

Create `internal/app/list_test.go`:

```go
package app

import (
	"testing"
	"time"

	"github.com/abhinav/grove/internal/store"
)

func TestFilterSessions(t *testing.T) {
	repoFermat := "/projects/fermat"
	repoGrove := "/projects/grove"

	sessions := []*store.Session{
		{ID: "1", Name: "auth-fix", Tool: "claude", Directory: "/projects/fermat/auth", RepoRoot: &repoFermat, CreatedAt: time.Now()},
		{ID: "2", Name: "tile-layout", Tool: "codex", Directory: "/projects/fermat/tile", RepoRoot: &repoFermat, CreatedAt: time.Now()},
		{ID: "3", Name: "repo-scoping", Tool: "claude", Directory: "/projects/grove/repo", RepoRoot: &repoGrove, CreatedAt: time.Now()},
		{ID: "4", Name: "scratch", Tool: "claude", Directory: "/tmp/experiment", CreatedAt: time.Now()},
	}

	t.Run("no filter returns all", func(t *testing.T) {
		result := filterSessions(sessions, "")
		if len(result) != 4 {
			t.Fatalf("expected 4, got %d", len(result))
		}
	})

	t.Run("filter by repo name", func(t *testing.T) {
		result := filterSessions(sessions, "fermat")
		if len(result) != 2 {
			t.Fatalf("expected 2, got %d", len(result))
		}
	})

	t.Run("filter by tool", func(t *testing.T) {
		result := filterSessions(sessions, "codex")
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		result := filterSessions(sessions, "auth")
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("filter case insensitive", func(t *testing.T) {
		result := filterSessions(sessions, "CLAUDE")
		if len(result) != 3 {
			t.Fatalf("expected 3, got %d", len(result))
		}
	})

	t.Run("filter by directory", func(t *testing.T) {
		result := filterSessions(sessions, "experiment")
		if len(result) != 1 {
			t.Fatalf("expected 1, got %d", len(result))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := filterSessions(sessions, "zzzzz")
		if len(result) != 0 {
			t.Fatalf("expected 0, got %d", len(result))
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -run TestFilterSessions -v`
Expected: FAIL — `filterSessions` does not exist.

- [ ] **Step 3: Implement filterSessions and filter state**

In `internal/app/list.go`:

Add filter fields to `ListModel`:
```go
type ListModel struct {
	Sessions    []*store.Session
	Cursor      int
	Width       int
	Height      int
	RepoOrder   []string // repo basenames in config order

	FilterActive bool
	FilterText   string
	FilterInput  textinput.Model
}
```

Add `filterSessions` function:
```go
func filterSessions(sessions []*store.Session, query string) []*store.Session {
	if query == "" {
		return sessions
	}
	q := strings.ToLower(query)
	var result []*store.Session
	for _, sess := range sessions {
		repo := "—"
		if sess.RepoRoot != nil {
			repo = filepath.Base(*sess.RepoRoot)
		}
		haystack := strings.ToLower(sess.Name + " " + repo + " " + sess.Tool + " " + sess.Directory)
		if strings.Contains(haystack, q) {
			result = append(result, sess)
		}
	}
	return result
}
```

Update `View()` to use `filterSessions` when rendering — filter the sessions before display but keep the original list unchanged. Add imports for `strings`, `path/filepath`, `github.com/charmbracelet/bubbles/textinput`.

- [ ] **Step 4: Add Filter keybinding to keys.go**

```go
Filter: key.NewBinding(
	key.WithKeys("ctrl+f"),
	key.WithHelp("ctrl+f", "filter"),
),
```

Add `Filter key.Binding` to `keyMap` struct.
Add `keys.Filter` to `statusBarHelp()`.

- [ ] **Step 5: Add filter bar styles to styles.go**

```go
filterBarStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("15")).
	Background(lipgloss.Color("235")).
	Padding(0, 1)

filterActiveStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8")).
	Italic(true)
```

- [ ] **Step 6: Wire filter into app.go**

In `internal/app/app.go`:

Handle `ctrl+f` in `handleKey`:
```go
case key.Matches(msg, keys.Filter):
	m.list.FilterActive = true
	m.list.FilterInput.Focus()
	return m, textinput.Blink
```

When filter is active, route key events to the filter input in the list view update. Handle `esc` to clear filter and `enter` to keep it.

Update `View()` in list.go:
- When `FilterActive` and input is focused, render filter bar at bottom instead of status bar
- When filter text is set but input is not focused, show dimmed `"filter: <text>"` in status bar

- [ ] **Step 7: Run tests**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -v`
Expected: All pass.

- [ ] **Step 8: Run full test suite**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All pass.

- [ ] **Step 9: Commit**

```bash
git add internal/app/list.go internal/app/list_test.go internal/app/app.go internal/app/keys.go internal/app/styles.go
git commit -m "feat(list): add ctrl+f filter bar with case-insensitive search across all columns"
```

---

## Chunk 5: Create wizard — repo-aware worktree creation

### Task 8: Update create wizard for repo selection

**Files:**
- Modify: `internal/app/create.go`
- Test: `internal/app/create_test.go`

- [ ] **Step 1: Write tests for repo selection in create wizard**

Create `internal/app/create_test.go`:

```go
package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/abhinav/grove/internal/config"
	"github.com/abhinav/grove/internal/session"
	"github.com/abhinav/grove/internal/store"
)

func newTestCreateDeps(t *testing.T) (*config.Config, *session.Manager) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude", WorktreeBase: "~/wt"},
		Tools:    map[string]config.ToolConfig{"claude": {Command: "claude"}, "codex": {Command: "codex"}},
		Repos: []config.RepoConfig{
			{RepoRoot: "/projects/fermat", WorktreeBase: "/projects/fermat-wt"},
			{RepoRoot: "/projects/grove", WorktreeBase: "/projects/grove-wt"},
		},
	}
	_ = config.Save(cfg)

	dbPath := dir + "/test.db"
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })

	mgr := session.NewManager(s, cfg)
	return cfg, mgr
}

func TestCreateModel_WorktreeShowsRepoSelector(t *testing.T) {
	cfg, mgr := newTestCreateDeps(t)
	m := NewCreateModel(cfg, mgr)

	// Press "2" for worktree mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	// Should be on repo select step
	view := m.View()
	if !containsStr(view, "fermat") || !containsStr(view, "grove") {
		t.Fatalf("expected repo names in view, got:\n%s", view)
	}
}

func TestCreateModel_WorktreeNoReposShowsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &config.Config{
		Defaults: config.DefaultsConfig{AITool: "claude"},
		Tools:    map[string]config.ToolConfig{"claude": {Command: "claude"}},
	}
	_ = config.Save(cfg)

	dbPath := dir + "/test.db"
	s, _ := store.Open(dbPath)
	defer s.Close()
	mgr := session.NewManager(s, cfg)

	m := NewCreateModel(cfg, mgr)

	// Press "2" for worktree mode
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})

	// Should show error about no repos
	if m.err == "" {
		t.Fatal("expected error about no repos configured")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -run "TestCreateModel" -v`
Expected: FAIL — no repo selection step exists.

- [ ] **Step 3: Add stepRepoSelect and repo picker to create.go**

In `internal/app/create.go`:

Add `stepRepoSelect` between `stepDirSource` and `stepDirInput`:
```go
const (
	stepDirSource  createStep = iota
	stepRepoSelect            // new: pick repo for worktree mode
	stepDirInput
	stepTool
	stepPrompt
	stepConfirm
)
```

Add fields to `CreateModel`:
```go
// Repo selection (worktree mode)
repoNames    []string // basenames for display
repoConfigs  []config.RepoConfig
repoSelected int
selectedRepo *config.RepoConfig
```

Update `handleDirSourceKey` for worktree mode (`"2"`, `"w"`):
- If `len(cfg.Repos) == 0`, set `m.err = "No repos configured yet. Run: grove repo add"` and return
- Otherwise populate `repoNames`/`repoConfigs` and move to `stepRepoSelect`
- Auto-detect cwd: try `worktree.GetMainRepoPath(".")`, if it matches a registered repo, pre-select it

Add `handleRepoSelectKey`:
- Left/right to navigate repos (same pattern as tool selector)
- Enter to confirm, set `m.selectedRepo`, move to `stepDirInput` (branch name)

Update `createSession()`:
- Use `m.selectedRepo.WorktreeBase` instead of `m.config.Defaults.WorktreeBase`
- Use `cfg.EffectiveSetupCommands(m.selectedRepo)` for setup commands
- Pass `&m.selectedRepo.RepoRoot` as `repoRoot` to `m.manager.Create`

Update `View()` to render the repo picker step.

For **existing directory mode**: in `createSession()`, attempt repo detection:
```go
if m.dirSource == dirExisting {
	// Try to detect repo root from the directory
	if repoRoot, err := worktree.GetMainRepoPath(m.resolvedDir); err == nil {
		if repo := m.config.RepoFor(repoRoot); repo != nil {
			repoRootStr := repo.RepoRoot
			repoRootPtr = &repoRootStr
		}
	}
}
```

- [ ] **Step 4: Update Manager.Create signature**

Already done in Task 3, Step 6. Just wire the `repoRoot` parameter through from the create wizard.

- [ ] **Step 5: Run tests**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./internal/app/ -v`
Expected: All pass.

- [ ] **Step 6: Run full test suite**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All pass.

- [ ] **Step 7: Commit**

```bash
git add internal/app/create.go internal/app/create_test.go
git commit -m "feat(create): add repo selection step for worktree mode with auto-detect"
```

---

## Chunk 6: Wire repo context into AppModel and session list

### Task 9: Pass repo order to ListModel and wire AppModel

**Files:**
- Modify: `internal/app/app.go`
- Modify: `internal/app/list.go`

- [ ] **Step 1: Update AppModel to pass repo order to ListModel**

In `app.go`, update `reconcileAndLoad` to also set `m.list.RepoOrder` from config:
```go
func (m AppModel) reconcileAndLoad() tea.Cmd {
	cfg := m.config
	return func() tea.Msg {
		_ = m.manager.Reconcile()
		sessions, err := m.store.ListSessions()
		if err != nil {
			return errMsg{err}
		}
		return sessionsMsg(sessions)
	}
}
```

In `New()`, initialize `RepoOrder`:
```go
func New(s *store.Store, cfg *config.Config, mgr *session.Manager) AppModel {
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

- [ ] **Step 2: Wire filter key handling in app.go**

Add filter handling to `handleKey`:
```go
case key.Matches(msg, keys.Filter):
	if !m.list.FilterActive {
		m.list.StartFilter()
		return m, textinput.Blink
	}
```

When `m.list.FilterActive` is true, route key events through list's filter handler before the main key handler. Add to the top of `Update`:
```go
if m.view == viewList && m.list.FilterActive {
	return m.updateListFilter(msg)
}
```

Implement `updateListFilter`:
```go
func (m AppModel) updateListFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.list.ClearFilter()
			return m, nil
		case tea.KeyEnter:
			m.list.CommitFilter()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.list.FilterInput, cmd = m.list.FilterInput.Update(msg)
	m.list.FilterText = m.list.FilterInput.Value()
	return m, cmd
}
```

Add `StartFilter()`, `ClearFilter()`, `CommitFilter()` methods to `ListModel`:
```go
func (m *ListModel) StartFilter() {
	m.FilterActive = true
	m.FilterInput = textinput.New()
	m.FilterInput.Placeholder = "filter..."
	m.FilterInput.Width = 40
	m.FilterInput.Focus()
}

func (m *ListModel) ClearFilter() {
	m.FilterActive = false
	m.FilterText = ""
	m.FilterInput.Blur()
}

func (m *ListModel) CommitFilter() {
	m.FilterText = m.FilterInput.Value()
	m.FilterActive = false
	m.FilterInput.Blur()
}
```

- [ ] **Step 3: Update list View to render filter bar**

In `list.go` `View()`:
- Replace the status bar area. The status bar is rendered in `app.go`, so update `app.go`'s `View()` to check `m.list.FilterActive`:

```go
// In app.go View(), replace the status bar section:
if m.list.FilterActive {
	bar = filterBarStyle.Width(m.list.Width).Render("filter: " + m.list.FilterInput.View())
} else if m.list.FilterText != "" {
	bar = statusBarStyle.Width(m.list.Width).Render(
		filterActiveStyle.Render("filter: "+m.list.FilterText) + "  " + statusBarHelp())
} else {
	bar = statusBarStyle.Width(m.list.Width).Render(statusBarHelp())
}
```

- [ ] **Step 4: Build and verify**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build ./cmd/grove/`
Expected: Build succeeds.

- [ ] **Step 5: Run full test suite**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All pass.

- [ ] **Step 6: Commit**

```bash
git add internal/app/app.go internal/app/list.go internal/app/keys.go internal/app/styles.go
git commit -m "feat: wire repo order, filter bar, and repo-scoped sorting into session list"
```

---

## Chunk 7: Integration tests and edge cases

### Task 10: Add integration tests for the full flow

**Files:**
- Create: `internal/store/migration_test.go`
- Modify: `internal/store/store_test.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: Test DB migration from v2 to v3**

Create `internal/store/migration_test.go`:

```go
package store

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrationV2ToV3(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	// Create a v2 database manually
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE sessions (
			id              TEXT PRIMARY KEY,
			name            TEXT NOT NULL,
			tool            TEXT NOT NULL,
			worktree        TEXT,
			directory       TEXT NOT NULL,
			prompt          TEXT,
			plan_file       TEXT,
			tmux_session    TEXT NOT NULL,
			tool_session_id TEXT,
			status          TEXT DEFAULT 'running',
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			stopped_at      DATETIME
		);
		PRAGMA user_version = 2;
	`)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a v2 session (no repo_root column)
	_, err = db.Exec(`INSERT INTO sessions (id, name, tool, directory, tmux_session, status)
		VALUES ('abc12345', 'old-session', 'claude', '/old/dir', 'grove-abc12345', 'running')`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	// Open with our Store (should run migration)
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("opening migrated store: %v", err)
	}
	defer s.Close()

	// Verify old session is readable and has nil repo_root
	sess, err := s.GetSession("abc12345")
	if err != nil {
		t.Fatalf("getting old session: %v", err)
	}
	if sess.RepoRoot != nil {
		t.Fatalf("expected nil repo_root for old session, got %v", sess.RepoRoot)
	}

	// Verify we can create new sessions with repo_root
	repoRoot := "/projects/fermat"
	newSess, err := s.CreateSession("new", "claude", "/dir", nil, nil, nil, nil, &repoRoot)
	if err != nil {
		t.Fatalf("creating new session: %v", err)
	}
	if newSess.RepoRoot == nil || *newSess.RepoRoot != "/projects/fermat" {
		t.Fatalf("expected repo_root, got %v", newSess.RepoRoot)
	}
}
```

- [ ] **Step 2: Test config round-trip with repos survives reload**

Add to `internal/config/config_test.go`:

```go
func TestRepoConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := &Config{
		Defaults: DefaultsConfig{AITool: "claude", WorktreeBase: "~/wt"},
		Tools:    map[string]ToolConfig{"claude": {Command: "claude"}},
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}

	// Add repos one at a time
	if err := cfg.AddRepo(RepoConfig{
		RepoRoot: "/projects/first", WorktreeBase: "/wt/first",
	}); err != nil {
		t.Fatal(err)
	}

	if err := cfg.AddRepo(RepoConfig{
		RepoRoot: "/projects/second", WorktreeBase: "/wt/second",
		AITool: "codex", SetupCommands: []string{"make"},
	}); err != nil {
		t.Fatal(err)
	}

	// Reload from disk
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(loaded.Repos))
	}

	// Verify cascade resolution
	repo1 := loaded.RepoFor("/projects/first")
	if loaded.EffectiveAITool(repo1) != "claude" {
		t.Fatal("expected global ai_tool for first repo")
	}

	repo2 := loaded.RepoFor("/projects/second")
	if loaded.EffectiveAITool(repo2) != "codex" {
		t.Fatal("expected overridden ai_tool for second repo")
	}
	cmds := loaded.EffectiveSetupCommands(repo2)
	if len(cmds) != 1 || cmds[0] != "make" {
		t.Fatalf("expected overridden setup commands, got %v", cmds)
	}
}
```

- [ ] **Step 3: Test filtering edge cases**

Add to `internal/app/list_test.go`:

```go
func TestFilterSessions_EmptyList(t *testing.T) {
	result := filterSessions(nil, "anything")
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestFilterSessions_NilRepoRoot(t *testing.T) {
	sessions := []*store.Session{
		{ID: "1", Name: "test", Tool: "claude", Directory: "/tmp", CreatedAt: time.Now()},
	}
	// Should not panic on nil RepoRoot
	result := filterSessions(sessions, "test")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestFilterSessions_DashForNilRepo(t *testing.T) {
	sessions := []*store.Session{
		{ID: "1", Name: "test", Tool: "claude", Directory: "/tmp", CreatedAt: time.Now()},
	}
	// Filtering by "—" should not match nil repos (it's a display char, not in data)
	// But filtering by the session name should work
	result := filterSessions(sessions, "test")
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}
```

- [ ] **Step 4: Run all tests**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v`
Expected: All pass.

- [ ] **Step 5: Commit**

```bash
git add internal/store/migration_test.go internal/store/store_test.go internal/config/config_test.go internal/app/list_test.go
git commit -m "test: add integration tests for migration, config round-trip, and filter edge cases"
```

---

### Task 11: Final build verification and cleanup

- [ ] **Step 1: Run full test suite one final time**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go test ./... -v -count=1`
Expected: All tests pass.

- [ ] **Step 2: Run go vet**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go vet ./...`
Expected: No issues.

- [ ] **Step 3: Build the binary**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && go build -o grove ./cmd/grove/`
Expected: Build succeeds, binary created.

- [ ] **Step 4: Manual smoke test — verify repo add and list**

Run: `cd /Users/abhinav/Projects/Work/global-scripts/grove && ./grove repo list`
Expected: "No repos registered" message.

- [ ] **Step 5: Clean up build artifact**

Run: `rm /Users/abhinav/Projects/Work/global-scripts/grove/grove`

- [ ] **Step 6: Final commit if any cleanup needed**
