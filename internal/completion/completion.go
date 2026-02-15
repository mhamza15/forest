// Package completion provides dynamic shell completions for forest
// commands.
package completion

import (
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
)

// Projects returns registered project names for shell completion.
func Projects(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	names, err := config.ListProjects()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	completions := make([]cobra.Completion, len(names))
	for i, name := range names {
		completions[i] = cobra.Completion(name)
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// Branches returns worktree branch names for shell completion. The
// project is determined from the --project flag or inferred from the
// working directory.
func Branches(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	// Only complete the first argument (the branch name).
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	project, _ := cmd.Flags().GetString("project")
	if project == "" {
		var err error

		project, err = config.InferProject()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
	proj, err := config.LoadProject(project)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	trees, err := git.List(proj.Repo)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []cobra.Completion
	for _, t := range trees {
		if t.Branch != "" && !t.Bare {
			completions = append(completions, cobra.Completion(t.Branch))
		}
	}
	return completions, cobra.ShellCompDirectiveNoFileComp
}
