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

// Branches returns worktree branch names for the project specified as
// the first argument. It is intended for use as the second positional
// argument completion on tree commands.
func Branches(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	project := args[0]

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

// ProjectThenBranch completes the first argument as a project name
// and the second argument as a branch name within that project.
func ProjectThenBranch(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return Projects(cmd, args, toComplete)
	case 1:
		return Branches(cmd, args, toComplete)
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
