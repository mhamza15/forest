package session

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

var (
	sessionStyle   = lipgloss.NewStyle().Bold(true)
	sessionProject = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
	sessionBranch  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active tmux sessions",
		Args:  cobra.NoArgs,
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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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

			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
				sessionStyle.Render(session),
				sessionProject.Render(name),
				sessionBranch.Render(wt.Branch),
			)
			found++
		}
	}

	if found == 0 {
		fmt.Println("No active sessions.")
		return nil
	}

	return w.Flush()
}
