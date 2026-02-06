package tree

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "remove <project> <branch>",
		Short:             "Remove a worktree and its tmux session",
		Args:              cobra.ExactArgs(2),
		RunE:              runRemove,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	sessionName := tmux.SessionName(rc.Name, branch)

	// Look up the actual worktree path from git rather than constructing
	// it, so worktrees created before the SafeBranchDir convention or
	// outside of forest are found correctly.
	existing := git.FindByBranch(rc.Repo, branch)
	if existing == nil {
		return fmt.Errorf("no worktree found for branch %q in project %q", branch, project)
	}

	wtPath := existing.Path

	if err := tmux.KillSession(sessionName); err != nil {
		slog.Debug("could not kill tmux session", "session", sessionName, "err", err)
	}

	if err := git.Remove(rc.Repo, wtPath); err != nil {
		if !errors.Is(err, git.ErrWorktreeDirty) {
			return err
		}

		fmt.Printf("Worktree %s/%s has modified or untracked files.\n", project, branch)

		var force bool

		confirmErr := huh.NewConfirm().
			Title("Force remove?").
			Value(&force).
			Run()

		if confirmErr != nil {
			return confirmErr
		}

		if !force {
			return nil
		}

		if err := git.ForceRemove(rc.Repo, wtPath); err != nil {
			return err
		}
	}

	fmt.Printf("Removed worktree %s/%s\n", project, branch)

	return nil
}
