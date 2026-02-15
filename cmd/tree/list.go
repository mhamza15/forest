package tree

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
)

var (
	projectStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89B4FA"))
	branchStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
	pathDimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List worktrees for one or all projects",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	var names []string

	if projectFlag != "" {
		names = []string{projectFlag}
	} else {
		var err error
		names, err = config.ListProjects()
		if err != nil {
			return err
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
		type row struct {
			branch string
			path   string
		}

		var rows []row

		for _, t := range trees {
			if t.Bare || t.Branch == "" {
				continue
			}

			rows = append(rows, row{branch: t.Branch, path: t.Path})
		}

		if len(rows) == 0 {
			continue
		}

		found = true
		_, _ = fmt.Fprintln(w, projectStyle.Render(name))

		for _, r := range rows {
			_, _ = fmt.Fprintf(w, "  %s\t%s\n",
				branchStyle.Render(r.branch),
				pathDimStyle.Render(r.path),
			)
		}
	}

	if !found {
		fmt.Println("No worktrees found.")
		return nil
	}

	return w.Flush()
}
