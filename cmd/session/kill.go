package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/tmux"
)

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "kill <project> <branch>",
		Short:             "Kill a tmux session without removing its worktree",
		Args:              cobra.ExactArgs(2),
		RunE:              runKill,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

func runKill(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	sessionName := tmux.SessionName(project, branch)

	if !tmux.SessionExists(sessionName) {
		return fmt.Errorf("session %q does not exist", sessionName)
	}

	if err := tmux.KillSession(sessionName); err != nil {
		return err
	}

	fmt.Printf("Killed session %s\n", sessionName)

	return nil
}
