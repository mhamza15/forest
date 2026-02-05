package project

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
)

var nameFlag string

func newCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new [path]",
		Short: "Register a new project",
		Long:  "Register a git repository as a forest project. If no path is given, an interactive prompt is shown.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runNew,
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "project name (defaults to repo directory name)")

	return cmd
}

func runNew(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runNewInteractive()
	}

	return registerProject(args[0], nameFlag)
}

// registerProject validates the repo path, derives a name, and writes
// the project config file.
func registerProject(repoPath string, name string) error {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	if err := validateGitRepo(absPath); err != nil {
		return err
	}

	if name == "" {
		name = filepath.Base(absPath)
	}

	cfg := config.ProjectConfig{
		Repo: absPath,
	}

	if err := config.SaveProject(name, cfg); err != nil {
		return err
	}

	slog.Info("project registered", "name", name, "repo", absPath)

	return nil
}

// runNewInteractive prompts the user for repo path and project name
// using a huh form with a file picker for directory selection.
func runNewInteractive() error {
	var repoPath string
	var name string

	home, _ := filepath.Abs(".")

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Repository path").
				Description("Select a git repository directory").
				DirAllowed(true).
				FileAllowed(false).
				ShowHidden(true).
				CurrentDirectory(home).
				Value(&repoPath),

			huh.NewInput().
				Title("Project name").
				Description("Leave blank to use the repo directory name").
				Value(&name),
		),
	).Run()
	if err != nil {
		return err
	}

	return registerProject(repoPath, name)
}

func validateGitRepo(path string) error {
	c := exec.Command("git", "-C", path, "rev-parse", "--git-dir")

	if err := c.Run(); err != nil {
		return fmt.Errorf("%s is not a git repository", path)
	}

	return nil
}
