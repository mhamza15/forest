package session

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active tmux sessions",
		RunE:  runList,
	}
}

// runList iterates over all registered projects and their worktrees,
// printing each tmux session that is currently running.
func runList(_ *cobra.Command, _ []string) error {
	projects, err := config.ListProjects()
	if err != nil {
		return fmt.Errorf("listing projects: %w", err)
	}

	var found int

	for _, name := range projects {
		proj, err := config.LoadProject(name)
		if err != nil {
			return fmt.Errorf("loading project %q: %w", name, err)
		}

		worktrees, err := git.List(proj.Repo)
		if err != nil {
			return fmt.Errorf("listing worktrees for %q: %w", name, err)
		}

		for _, wt := range worktrees {
			if wt.Bare || wt.Branch == "" {
				continue
			}

			session := tmux.SessionName(name, wt.Branch)

			if !tmux.SessionExists(session) {
				continue
			}

			fmt.Printf("%s\t%s\t%s\n", session, name, wt.Branch)
			found++
		}
	}

	if found == 0 {
		fmt.Println("No active sessions.")
	}

	return nil
}
