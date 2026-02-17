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

// SafeBranchDir converts a branch name into a flat directory name by
// replacing path separators with dashes. Without this, a branch like
// feature/login would create nested directories.
func SafeBranchDir(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

// Worktree represents a single git worktree entry as reported by
// git worktree list --porcelain.
type Worktree struct {
	// Path is the filesystem path to the worktree directory.
	Path string

	// Branch is the local branch name checked out in this worktree.
	Branch string

	// Bare is true if this entry represents a bare repository.
	Bare bool
}

// Add creates a new worktree at worktreePath for the given branch. If
// the branch already exists in the repo, it is checked out into the
// worktree. Otherwise a new branch is created off baseBranch.
func Add(repoPath, worktreePath, branch, baseBranch string) error {
	var args []string

	if BranchExists(repoPath, branch) {
		args = []string{"-C", repoPath, "worktree", "add", worktreePath, branch}
	} else {
		args = []string{"-C", repoPath, "worktree", "add", "-b", branch, worktreePath, baseBranch}
	}

	cmd := exec.Command("git", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree add: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// BranchExists returns true if a local branch with the given name
// exists in the repository.
func BranchExists(repoPath, branch string) bool {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--verify", "refs/heads/"+branch)
	return cmd.Run() == nil
}

// CurrentBranch returns the branch name checked out in the given
// directory, or an empty string if it cannot be determined (e.g.
// detached HEAD or not a git directory).
func CurrentBranch(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))

	if branch == "HEAD" {
		return ""
	}

	return branch
}

// WorktreeRoot returns the root directory of the worktree containing
// the given directory, or an empty string if it cannot be determined.
func WorktreeRoot(dir string) string {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// IsMerged returns true if the given branch has been merged into the
// target branch. It uses git merge-base --is-ancestor to check
// whether the branch's HEAD is an ancestor of the target.
func IsMerged(repoPath, branch, target string) bool {
	cmd := exec.Command("git", "-C", repoPath, "merge-base", "--is-ancestor", branch, target)
	return cmd.Run() == nil
}

// RemoteBranches returns the set of branch names that exist on the
// given remote. It calls git ls-remote --heads once and parses all
// results, making it efficient for checking many branches.
func RemoteBranches(repoPath, remote string) (map[string]bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "ls-remote", "--heads", remote)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-remote: %w", err)
	}

	branches := make(map[string]bool)
	prefix := "refs/heads/"

	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		// Each line is: <sha>\trefs/heads/<branch>
		parts := strings.SplitN(scanner.Text(), "\t", 2)
		if len(parts) != 2 {
			continue
		}

		if name, ok := strings.CutPrefix(parts[1], prefix); ok {
			branches[name] = true
		}
	}

	return branches, nil
}

// PruneReason describes why a branch is eligible for pruning,
// or that it is not.
type PruneReason int

const (
	// PruneNone means the branch should not be pruned.
	PruneNone PruneReason = iota

	// PruneMerged means the branch is an ancestor of the target branch
	// and has been fully merged.
	PruneMerged

	// PruneRemoteGone means the branch no longer exists on the remote.
	// This may indicate a squash-merge whose remote branch was deleted,
	// or a branch that was deleted without being merged. Callers should
	// verify merge status before acting on this reason.
	PruneRemoteGone
)

// PruneCheck determines whether a branch should be considered for
// pruning and returns the reason. A branch is unconditionally prunable
// if it has been merged into the target. When the branch is simply
// absent from the remote, PruneRemoteGone is returned so the caller
// can verify merge status before removing it.
func PruneCheck(repoPath, branch, target string, remoteBranches map[string]bool) PruneReason {
	if IsMerged(repoPath, branch, target) {
		return PruneMerged
	}
	if remoteBranches != nil && !remoteBranches[branch] {
		return PruneRemoteGone
	}

	return PruneNone
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
	cmd := exec.Command("git", "-C", repoPath, "worktree", "list", "--porcelain")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	return parsePorcelain(output), nil
}

// FindByBranch returns the worktree for the given branch, or nil if
// no worktree is checked out on that branch.
func FindByBranch(repoPath, branch string) *Worktree {
	trees, err := List(repoPath)
	if err != nil {
		return nil
	}

	for _, t := range trees {
		if t.Branch == branch {
			return &t
		}
	}

	return nil
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
