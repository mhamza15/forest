// Package forest provides high-level workflow operations that
// orchestrate config, git, and tmux together.
package forest

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

// AddTreeResult holds the outcome of an AddTree operation.
type AddTreeResult struct {
	// Created is true if a new worktree was created, false if it
	// already existed.
	Created bool

	// WorktreePath is the filesystem path to the worktree.
	WorktreePath string

	// SessionName is the tmux session name for this worktree.
	SessionName string

	// CopyWarnings contains any warnings generated while copying
	// files into the new worktree.
	CopyWarnings []string
}

// AddTree creates a worktree for the given project and branch. If the
// worktree already exists, it reuses it. The caller is responsible for
// creating a tmux session and switching to it.
func AddTree(rc config.ResolvedConfig, branch string) (AddTreeResult, error) {
	sessionName := tmux.SessionName(rc.Name, branch)

	result := AddTreeResult{
		SessionName: sessionName,
	}

	// Check if a worktree for this branch already exists (at any
	// path, including paths created before the SafeBranchDir
	// convention).
	if existing := git.FindByBranch(rc.Repo, branch); existing != nil {
		result.WorktreePath = existing.Path
		return result, nil
	}

	wtPath := filepath.Join(rc.WorktreeDir, rc.Name, git.SafeBranchDir(branch))

	slog.Debug("creating worktree", "path", wtPath, "base", rc.Branch)

	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		return result, fmt.Errorf("creating worktree parent dir: %w", err)
	}

	if err := git.Add(rc.Repo, wtPath, branch, rc.Branch); err != nil {
		return result, err
	}

	result.Created = true
	result.WorktreePath = wtPath

	if len(rc.Copy) > 0 {
		result.CopyWarnings = git.CopyFiles(rc.Repo, wtPath, rc.Copy)
	}

	return result, nil
}

// OpenSession creates a tmux session for an existing worktree if one
// does not already exist, and applies the configured layout. It does
// not switch to the session.
func OpenSession(rc config.ResolvedConfig, branch string, wtPath string) error {
	sessionName := tmux.SessionName(rc.Name, branch)

	if tmux.SessionExists(sessionName) {
		return nil
	}

	if err := tmux.NewSession(sessionName, wtPath); err != nil {
		return err
	}

	if len(rc.Layout) == 0 {
		return nil
	}

	windows := make([]tmux.LayoutWindow, len(rc.Layout))
	for i, w := range rc.Layout {
		windows[i] = tmux.LayoutWindow{Name: w.Name, Command: w.Command}
	}

	return tmux.ApplyLayout(sessionName, wtPath, windows)
}

// RemoveTree removes a worktree and its tmux session. If force is
// true, dirty worktrees are removed anyway. Returns
// git.ErrWorktreeDirty (wrapped) if the worktree has modifications
// and force is false.
func RemoveTree(rc config.ResolvedConfig, branch string, force bool) error {
	existing := git.FindByBranch(rc.Repo, branch)
	if existing == nil {
		return fmt.Errorf("no worktree found for branch %q in project %q", branch, rc.Name)
	}

	wtPath := existing.Path
	sessionName := tmux.SessionName(rc.Name, branch)

	if err := tmux.KillSession(sessionName); err != nil {
		slog.Debug("could not kill tmux session", "session", sessionName, "err", err)
	}

	if force {
		return git.ForceRemove(rc.Repo, wtPath)
	}

	return git.Remove(rc.Repo, wtPath)
}
