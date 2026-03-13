package worktree

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// BranchExists checks if a branch exists locally.
func BranchExists(repoDir, branch string) (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+branch)
	cmd.Dir = repoDir
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 128 {
			return false, nil
		}
		// Also treat exit code 1 as "not found"
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
			return false, nil
		}
		return false, fmt.Errorf("checking branch %s: %w", branch, err)
	}
	return true, nil
}

// IsRemoteOnly checks if a branch exists only on the remote (not locally).
func IsRemoteOnly(repoDir, branch, remote string) (bool, error) {
	local, err := BranchExists(repoDir, branch)
	if err != nil {
		return false, err
	}
	if local {
		return false, nil
	}

	cmd := exec.Command("git", "rev-parse", "--verify", "refs/remotes/"+remote+"/"+branch)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return false, nil // not on remote either
	}
	return true, nil
}

// FetchRemote runs git fetch for the given remote.
func FetchRemote(repoDir, remote string) error {
	cmd := exec.Command("git", "fetch", remote)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch %s: %s: %w", remote, string(out), err)
	}
	return nil
}

// WorktreeInfo holds parsed info from `git worktree list --porcelain`.
type WorktreeInfo struct {
	Path   string
	HEAD   string
	Branch string
	Bare   bool
}

// ListWorktrees parses `git worktree list --porcelain` output.
func ListWorktrees(repoDir string) ([]WorktreeInfo, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("listing worktrees: %w", err)
	}
	return parseWorktreeList(string(out)), nil
}

func parseWorktreeList(output string) []WorktreeInfo {
	var result []WorktreeInfo
	var current *WorktreeInfo

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			if current != nil {
				result = append(result, *current)
			}
			current = &WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "HEAD "):
			if current != nil {
				current.HEAD = strings.TrimPrefix(line, "HEAD ")
			}
		case strings.HasPrefix(line, "branch "):
			if current != nil {
				current.Branch = strings.TrimPrefix(line, "branch ")
			}
		case line == "bare":
			if current != nil {
				current.Bare = true
			}
		case line == "":
			if current != nil {
				result = append(result, *current)
				current = nil
			}
		}
	}
	if current != nil {
		result = append(result, *current)
	}
	return result
}

// IsWorktreeClean returns true if the working directory has no uncommitted changes.
func IsWorktreeClean(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(string(out)) == "", nil
}

var invalidBranchChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// ValidateBranchName checks if a string is a valid git branch name.
func ValidateBranchName(name string) bool {
	if name == "" {
		return false
	}
	cmd := exec.Command("git", "check-ref-format", "--branch", name)
	return cmd.Run() == nil
}

// SanitizeBranchName converts a string into a valid branch name.
// e.g., "feature/foo" → "feature-foo"
func SanitizeBranchName(name string) string {
	name = strings.ReplaceAll(name, "/", "-")
	name = invalidBranchChars.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-.")
	if name == "" {
		return "branch"
	}
	return name
}

// GetMainRepoPath returns the main repository path for a worktree directory.
func GetMainRepoPath(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-common-dir")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting common dir: %w", err)
	}
	commonDir := strings.TrimSpace(string(out))
	// --git-common-dir returns a path relative to the worktree or absolute
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(dir, commonDir)
	}
	// The common dir is the .git directory; the repo root is its parent
	return filepath.Dir(filepath.Clean(commonDir)), nil
}
