package tree

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func switchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch <project> <branch>",
		Short: "Switch to an existing worktree's tmux session",
		Long: `Switch to the tmux session for an existing worktree. If the
worktree exists but the session does not, the session is created
with the configured layout. Does not create new worktrees.`,
		Args:              cobra.ExactArgs(2),
		RunE:              runSwitch,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

func runSwitch(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]
	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	existing := git.FindByBranch(rc.Repo, branch)
	if existing == nil {
		return fmt.Errorf("no worktree for branch %q in project %q", branch, project)
	}

	if err := forest.OpenSession(rc, branch, existing.Path); err != nil {
		return err
	}

	return tmux.SwitchTo(tmux.SessionName(rc.Name, branch))
}
