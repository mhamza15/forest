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

var forceFlag bool

func removeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "remove <project> <branch>",
		Short:             "Remove a worktree and its tmux session",
		Args:              cobra.ExactArgs(2),
		RunE:              runRemove,
		ValidArgsFunction: completion.ProjectThenBranch,
	}

	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "force removal of dirty worktrees")

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	err = forest.RemoveTree(rc, branch, forceFlag)
	if err != nil {
		if !errors.Is(err, git.ErrWorktreeDirty) {
			return err
		}

		// Worktree is dirty and --force was not set. Ask interactively.
		fmt.Printf("Worktree %s/%s has modified or untracked files.\n", project, branch)

		var confirm bool

		confirmErr := huh.NewConfirm().
			Title("Force remove?").
			Value(&confirm).
			Run()

		if confirmErr != nil {
			return confirmErr
		}

		if !confirm {
			return nil
		}

		if err := forest.RemoveTree(rc, branch, true); err != nil {
			return err
		}
	}

	fmt.Printf("Removed worktree %s/%s\n", project, branch)

	return nil
}
