package tree

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:               "list [project]",
		Short:             "List worktrees for one or all projects",
		Args:              cobra.MaximumNArgs(1),
		RunE:              runList,
		ValidArgsFunction: completion.Projects,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	var names []string

	if len(args) == 1 {
		names = []string{args[0]}
	} else {
		var err error

		names, err = config.ListProjects()
		if err != nil {
			return err
		}
	}

	found := false

	for _, name := range names {
		proj, err := config.LoadProject(name)
		if err != nil {
			return err
		}

		trees, err := git.List(proj.Repo)
		if err != nil {
			return err
		}

		// Buffer worktree lines so we only print the project header
		// when there is at least one non-skipped worktree.
		var lines []string

		for _, t := range trees {
			if t.Bare || t.Branch == "" {
				continue
			}

			lines = append(lines, fmt.Sprintf("  %s\t%s\n", t.Branch, t.Path))
		}

		if len(lines) == 0 {
			continue
		}

		found = true
		fmt.Println(name)

		for _, l := range lines {
			fmt.Print(l)
		}
	}

	if !found {
		fmt.Println("No worktrees found.")
	}

	return nil
}
