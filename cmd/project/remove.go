package project

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
)

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "remove <project>",
		Short:             "Unregister a project",
		Long:              "Remove a project's configuration file. This does not delete the repository or any worktrees.",
		Args:              cobra.ExactArgs(1),
		RunE:              runRemove,
		ValidArgsFunction: completion.Projects,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Verify the project exists before removing.
	if _, err := config.LoadProject(name); err != nil {
		return err
	}

	if err := config.RemoveProject(name); err != nil {
		return err
	}

	fmt.Printf("Removed project %q\n", name)

	return nil
}
