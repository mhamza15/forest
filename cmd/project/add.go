package project

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
)

var nameFlag string

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [path]",
		Short: "Register a new project",
		Long:  "Register a git repository as a forest project. If no path is given, an interactive prompt is shown.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAdd,
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "project name (defaults to repo directory name)")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runAddInteractive()
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

	fmt.Printf("Registered project %q (%s)\n", name, absPath)

	return nil
}

// runAddInteractive prompts the user for repo path and project name
// using a huh form with a file picker for directory selection.
func runAddInteractive() error {
	var name string

	startDir, _ := filepath.Abs(".")

	global, err := config.LoadGlobal()
	if err == nil && global.ProjectsDir != "" {
		startDir = global.ProjectsDir
	}

	repoPath := startDir

	fmt.Println()

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit"))
	km.FilePicker.Up.SetEnabled(true)
	km.FilePicker.Down.SetEnabled(true)
	km.FilePicker.Close.SetEnabled(true)
	km.FilePicker.Open.SetEnabled(true)
	km.FilePicker.Back.SetEnabled(true)

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewFilePicker().
				Title("Repository path").
				Description("Select a git repository directory").
				DirAllowed(true).
				FileAllowed(false).
				ShowHidden(false).
				ShowPermissions(false).
				Height(5).
				CurrentDirectory(startDir).
				Picking(true).
				Value(&repoPath),

			huh.NewInput().
				Title("Project name").
				Description("Leave blank to use the repo directory name").
				Value(&name),
		),
	).WithKeyMap(km).Run()
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
