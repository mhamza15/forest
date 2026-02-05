// Package git provides thin wrappers around git commands for worktree
// management.
package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrWorktreeDirty is returned when a worktree has modified or untracked
// files and cannot be removed without --force.
var ErrWorktreeDirty = errors.New("worktree contains modified or untracked files")

// Worktree represents a single git worktree entry as reported by
// git worktree list --porcelain.
type Worktree struct {
	Path   string
	Branch string
	Bare   bool
}

// Add creates a new worktree at worktreePath, branching off baseBranch
// with the new branch name newBranch. repoPath is the path to the
// main repository.
func Add(repoPath, worktreePath, newBranch, baseBranch string) error {
	//nolint:gosec // Arguments are derived from validated config, not user input.
	cmd := exec.Command(
		"git", "-C", repoPath,
		"worktree", "add",
		"-b", newBranch,
		worktreePath,
		baseBranch,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// Remove removes the worktree at the given path. If the worktree has
// modified or untracked files, it returns ErrWorktreeDirty. Use
// ForceRemove to remove it anyway.
func Remove(repoPath, worktreePath string) error {
	return removeWorktree(repoPath, worktreePath, false)
}

// ForceRemove removes the worktree at the given path even if it
// contains modified or untracked files.
func ForceRemove(repoPath, worktreePath string) error {
	return removeWorktree(repoPath, worktreePath, true)
}

func removeWorktree(repoPath, worktreePath string, force bool) error {
	args := []string{"-C", repoPath, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, worktreePath)

	//nolint:gosec // Arguments are derived from validated config, not user input.
	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := string(bytes.TrimSpace(output))

		if strings.Contains(msg, "modified or untracked") {
			return fmt.Errorf("%w: %s", ErrWorktreeDirty, worktreePath)
		}

		return fmt.Errorf("git worktree remove: %s: %w", msg, err)
	}

	return nil
}

// List returns all worktrees for the repository at repoPath by parsing
// the porcelain output of git worktree list.
func List(repoPath string) ([]Worktree, error) {
	//nolint:gosec // repoPath is derived from validated config.
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	return parsePorcelain(output), nil
}

// parsePorcelain parses the porcelain output of git worktree list.
// Each worktree block is separated by a blank line. Within a block:
//
//	worktree /path/to/worktree
//	HEAD <sha>
//	branch refs/heads/<branch>
//
// Bare worktrees have a "bare" line instead of a branch line.
func parsePorcelain(data []byte) []Worktree {
	var trees []Worktree
	var current Worktree

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")

		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")

		case line == "bare":
			current.Bare = true

		case line == "":
			if current.Path != "" {
				trees = append(trees, current)
			}
			current = Worktree{}
		}
	}

	// Final block may not end with a blank line.
	if current.Path != "" {
		trees = append(trees, current)
	}

	return trees
}
