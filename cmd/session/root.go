// Package session implements the "forest session" command group.
package session

import "github.com/spf13/cobra"

// Command returns the session parent command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage tmux sessions",
		Long:  "Manage tmux sessions without affecting worktrees.",
	}

	cmd.AddCommand(killCmd())

	return cmd
}
