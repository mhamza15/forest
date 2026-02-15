package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/tmux"
)

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "kill <branch>",
		Short:             "Kill a tmux session without removing its worktree",
		Args:              cobra.ExactArgs(1),
		RunE:              runKill,
		ValidArgsFunction: completion.Branches,
	}
}

func runKill(cmd *cobra.Command, args []string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	branch := args[0]

	var project string

	if projectFlag != "" {
		project = projectFlag
	} else {
		var err error
		project, err = config.InferProject()
		if err != nil {
			return err
		}
	}

	sessionName := tmux.SessionName(project, branch)

	// KillSession treats a missing session as a no-op, but kill is
	// user-initiated so we surface an explicit error instead.
	// The redundant SessionExists check inside KillSession is harmless.
	if !tmux.SessionExists(sessionName) {
		return fmt.Errorf("session %q does not exist", sessionName)
	}

	if err := tmux.KillSession(sessionName); err != nil {
		return err
	}

	fmt.Printf("Killed session %s\n", sessionName)

	return nil
}
