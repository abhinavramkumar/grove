package worktree

import (
	"fmt"
	"os/exec"
)

// RunSetupCommands runs each command sequentially in the given directory,
// stopping on first failure.
func RunSetupCommands(dir string, commands []string) error {
	for _, cmdStr := range commands {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("setup command %q failed: %s: %w", cmdStr, string(out), err)
		}
	}
	return nil
}
