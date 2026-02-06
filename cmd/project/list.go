package project

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
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

	for _, name := range names {
		proj, err := config.LoadProject(name)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\n", name, proj.Repo)
	}

	return w.Flush()
}
