package tree

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
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

	if err := forest.RemoveTree(rc, branch, false); err != nil {
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

		if err := forest.RemoveTree(rc, branch, true); err != nil {
			return err
		}
	}

	fmt.Printf("Removed worktree %s/%s\n", project, branch)

	return nil
}
