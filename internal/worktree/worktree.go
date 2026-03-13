package worktree

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// Create creates a new git worktree at basePath/branch.
// It handles three cases:
//   - Remote-only branch: creates worktree tracking the remote branch
//   - Existing local branch: creates worktree from that branch
//   - New branch: creates worktree with a new branch from startPoint
func Create(repoDir, basePath, branch, startPoint string) (string, error) {
	wtPath := filepath.Join(basePath, SanitizeBranchName(branch))

	remoteOnly, err := IsRemoteOnly(repoDir, branch, "origin")
	if err != nil {
		return "", fmt.Errorf("checking remote: %w", err)
	}

	localExists, err := BranchExists(repoDir, branch)
	if err != nil {
		return "", fmt.Errorf("checking local branch: %w", err)
	}

	var cmd *exec.Cmd
	switch {
	case remoteOnly:
		// Create worktree tracking remote branch
		cmd = exec.Command("git", "worktree", "add", "-b", branch, wtPath, "origin/"+branch)
	case localExists:
		// Use existing local branch
		cmd = exec.Command("git", "worktree", "add", wtPath, branch)
	default:
		// New branch from startPoint
		if startPoint == "" {
			startPoint = "HEAD"
		}
		cmd = exec.Command("git", "worktree", "add", "-b", branch, wtPath, startPoint)
	}

	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("creating worktree: %s: %w", string(out), err)
	}

	return wtPath, nil
}

// Remove removes a worktree at the given path.
func Remove(repoDir, wtPath string, force bool) error {
	args := []string{"worktree", "remove", wtPath}
	if force {
		args = append(args, "--force")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("removing worktree: %s: %w", string(out), err)
	}
	return nil
}

// List returns worktree info for all worktrees under basePath by filtering
// the full worktree list from the repo.
func List(repoDir, basePath string) ([]WorktreeInfo, error) {
	all, err := ListWorktrees(repoDir)
	if err != nil {
		return nil, err
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	var filtered []WorktreeInfo
	for _, wt := range all {
		absPath, err := filepath.Abs(wt.Path)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(absBase, absPath)
		if err != nil {
			continue
		}
		// Only include paths that are under basePath (no ".." prefix)
		if len(rel) > 0 && rel[0] != '.' {
			filtered = append(filtered, wt)
		}
	}
	return filtered, nil
}
