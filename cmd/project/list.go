package project

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
)

var (
	projectNameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#89B4FA"))
	pathStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086"))
	headerStyle      = lipgloss.NewStyle().Bold(true)
)

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered projects",
		Long:  "Print all registered projects and their repository paths.",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}
}

func runList(cmd *cobra.Command, args []string) error {
	names, err := config.ListProjects()
	if err != nil {
		return err
	}

	if len(names) == 0 {
		fmt.Println("No projects registered.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "%s\t%s\n", headerStyle.Render("Project"), headerStyle.Render("Repo"))

	for _, name := range names {
		proj, err := config.LoadProject(name)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\n",
			projectNameStyle.Render(name),
			pathStyle.Render(proj.Repo),
		)
	}

	return w.Flush()
}
