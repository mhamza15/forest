// Package tree implements the "forest tree" command group.
package tree

import "github.com/spf13/cobra"

// Command returns the tree parent command. When invoked without a
// subcommand it launches the inline tree browser TUI.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Manage and browse worktrees",
		Long:  "Create, remove, and browse worktrees. Run without a subcommand to open the interactive tree browser.",
		RunE:  runBrowser,
	}

	cmd.AddCommand(newCmd())
	cmd.AddCommand(removeCmd())

	return cmd
}

// runBrowser launches the inline TUI for browsing projects and trees.
// Wired up in a later commit.
func runBrowser(cmd *cobra.Command, args []string) error {
	return runTreeBrowser()
}
