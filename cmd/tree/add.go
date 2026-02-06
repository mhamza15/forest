package tree

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/tmux"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <project> <branch>",
		Short: "Create a new worktree and tmux session",
		Long: `Create a new git worktree for the project and open it in a tmux session.

If the worktree already exists, the command switches to the existing
tmux session instead. The new branch is based on the project's configured
base branch, falling back to the global default.`,
		Args:              cobra.ExactArgs(2),
		RunE:              runAdd,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	if err := tmux.RequireRunning(); err != nil {
		return err
	}

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	result, err := forest.AddTree(rc, branch)
	if err != nil {
		return err
	}

	if result.Created {
		fmt.Printf("Created worktree %s/%s\n", project, branch)
	} else {
		fmt.Printf("Worktree %s/%s already exists\n", project, branch)
	}

	for _, w := range result.CopyWarnings {
		fmt.Println(w)
	}

	if err := forest.OpenSession(rc, branch, result.WorktreePath); err != nil {
		return err
	}

	slog.Debug("switching to tmux session", "session", result.SessionName)

	return tmux.SwitchTo(result.SessionName)
}
