package tools

import (
	"strings"
	"testing"

	"github.com/abhinav/grove/internal/config"
)

func TestClaudeAdapter_NewSessionCmd(t *testing.T) {
	a := &ClaudeAdapter{Command: "claude"}

	t.Run("with prompt", func(t *testing.T) {
		cmd, sessID := a.NewSessionCmd("/tmp/work", "fix the bug", "")
		if sessID == "" {
			t.Fatal("expected non-empty session ID")
		}
		if !strings.Contains(cmd, "--session-id "+sessID) {
			t.Errorf("expected --session-id flag, got: %s", cmd)
		}
		if !strings.Contains(cmd, `"fix the bug"`) {
			t.Errorf("expected prompt in command, got: %s", cmd)
		}
		if !strings.Contains(cmd, `-p`) {
			t.Errorf("expected -p flag for prompt, got: %s", cmd)
		}
	})

	t.Run("with plan file", func(t *testing.T) {
		cmd, sessID := a.NewSessionCmd("/tmp/work", "", "/tmp/plan.md")
		if sessID == "" {
			t.Fatal("expected non-empty session ID")
		}
		if !strings.Contains(cmd, "--plan-file /tmp/plan.md") {
			t.Errorf("expected --plan-file flag, got: %s", cmd)
		}
	})

	t.Run("interactive (no prompt)", func(t *testing.T) {
		cmd, sessID := a.NewSessionCmd("/tmp/work", "", "")
		if sessID == "" {
			t.Fatal("expected non-empty session ID")
		}
		if !strings.Contains(cmd, "--session-id") {
			t.Errorf("expected --session-id, got: %s", cmd)
		}
		if strings.Contains(cmd, "-p") {
			t.Errorf("should not have -p without prompt, got: %s", cmd)
		}
	})
}

func TestClaudeAdapter_ResumeSessionCmd(t *testing.T) {
	a := &ClaudeAdapter{Command: "claude"}
	cmd := a.ResumeSessionCmd("/tmp/work", "abc-123-uuid")
	if !strings.Contains(cmd, "--resume abc-123-uuid") {
		t.Errorf("expected --resume flag, got: %s", cmd)
	}
	if !strings.HasPrefix(cmd, `cd "/tmp/work"`) {
		t.Errorf("expected cd prefix, got: %s", cmd)
	}
}

func TestOpenCodeAdapter_NewSessionCmd(t *testing.T) {
	a := &OpenCodeAdapter{Command: "opencode"}
	cmd, sessID := a.NewSessionCmd("/tmp/work", "do stuff", "")
	if sessID != "" {
		t.Errorf("opencode should not return a session ID, got %q", sessID)
	}
	if !strings.Contains(cmd, `opencode "do stuff"`) {
		t.Errorf("expected prompt, got: %s", cmd)
	}
}

func TestOpenCodeAdapter_ResumeSessionCmd(t *testing.T) {
	a := &OpenCodeAdapter{Command: "opencode"}
	cmd := a.ResumeSessionCmd("/tmp/work", "some-id")
	if !strings.Contains(cmd, "--continue") {
		t.Errorf("expected --continue flag, got: %s", cmd)
	}
}

func TestCodexAdapter_NoResume(t *testing.T) {
	a := &CodexAdapter{Command: "codex"}
	if a.SupportsResume() {
		t.Error("codex should not support resume")
	}
}

func TestLoadAdapters_Defaults(t *testing.T) {
	adapters := LoadAdapters(nil)
	for _, name := range []string{"claude", "codex", "opencode"} {
		if _, ok := adapters[name]; !ok {
			t.Errorf("missing default adapter %q", name)
		}
	}
	if !adapters["claude"].SupportsResume() {
		t.Error("claude should support resume")
	}
	if !adapters["opencode"].SupportsResume() {
		t.Error("opencode should support resume")
	}
}

func TestLoadAdapters_ConfigOverride(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.ToolConfig{
			"claude": {Command: "my-claude", Args: []string{"--fast"}},
			"aider":  {Command: "aider", Args: []string{"--model", "gpt-4"}},
		},
	}
	adapters := LoadAdapters(cfg)

	// claude should still be a ClaudeAdapter (supports resume)
	if !adapters["claude"].SupportsResume() {
		t.Error("overridden claude should still support resume")
	}

	// aider should be GenericAdapter
	if adapters["aider"] == nil {
		t.Fatal("expected aider adapter from config")
	}
	if adapters["aider"].SupportsResume() {
		t.Error("generic adapter should not support resume")
	}

	// codex default should still exist
	if adapters["codex"] == nil {
		t.Error("expected codex default to survive")
	}
}
