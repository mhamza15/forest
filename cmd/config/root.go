// Package config implements the "forest config" command.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	iconfig "github.com/mhamza15/forest/internal/config"
)

// Command returns the config cobra command, ready to be added as a
// subcommand of root.
func Command() *cobra.Command {
	return &cobra.Command{
		Use:               "config [project]",
		Short:             "Open configuration in your editor",
		Long:              "Opens the global config in $EDITOR. If a project name is given, opens that project's config instead.",
		Args:              cobra.MaximumNArgs(1),
		RunE:              run,
		ValidArgsFunction: completion.Projects,
	}
}

func run(cmd *cobra.Command, args []string) error {
	var path string

	if len(args) == 0 {
		if err := iconfig.WriteDefaultGlobal(); err != nil {
			return err
		}
		path = iconfig.GlobalConfigPath()
	} else {
		path = iconfig.ProjectConfigPath(args[0])
	}

	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config file not found: %s", path)
	}

	slog.Debug("opening config", "path", path)

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
