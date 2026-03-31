package tree

import (
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "add {<branch> | <github-link>}",
		Short:      "Deprecated alias for `switch`",
		Hidden:     true,
		Deprecated: "use `forest tree switch` instead",
		Args:       cobra.ExactArgs(1),
		RunE:       runSwitch,

		ValidArgsFunction: completion.Branches,
	}

	cmd.Flags().StringVarP(&baseBranchFlag, "branch", "b", "", "base branch for a new worktree (overrides project config)")

	return cmd
}
