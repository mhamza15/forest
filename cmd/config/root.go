// Package config implements the "forest config" command.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	iconfig "github.com/mhamza15/forest/internal/config"
)

// Command returns the config cobra command, ready to be added as a
// subcommand of root.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Open configuration in your editor",
		Long:  "Opens the global config in $EDITOR. Use --project to open a specific project's config instead.",
		Args:  cobra.NoArgs,
		RunE:  run,
	}
}

func run(cmd *cobra.Command, _ []string) error {
	project, _ := cmd.Flags().GetString("project")

	var path string

	if project == "" {
		if err := iconfig.WriteDefaultGlobal(); err != nil {
			return err
		}
		path = iconfig.GlobalConfigPath()
	} else {
		path = iconfig.ProjectConfigPath(project)
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config file not found: %s", path)
	}

	slog.Debug("opening config", slog.String("path", path))
	return openEditor(path)
}

// openEditor launches the user's preferred editor on the given file path.
func openEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	c := exec.Command(editor, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
