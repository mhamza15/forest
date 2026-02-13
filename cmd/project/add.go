package project

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/github"
	"github.com/mhamza15/forest/internal/tmux"
)

var nameFlag string

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [path | github-url]",
		Short: "Register a new project",
		Long: `Register a git repository as a forest project.

If no path is given, an interactive prompt is shown.

A GitHub repository URL may be passed instead of a local path:

  forest project add https://github.com/owner/repo

The repository is cloned into the current directory (or projects_dir
if configured), registered as a project, and the default branch is
opened in a tmux session.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runAdd,
	}

	cmd.Flags().StringVar(&nameFlag, "name", "", "project name (defaults to repo directory name)")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runAddInteractive()
	}

	if github.IsGitHubURL(args[0]) {
		return runAddFromGitHub(args[0], nameFlag)
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

// runAddFromGitHub clones a GitHub repository, registers it as a
// project, and opens the default branch in a tmux session.
func runAddFromGitHub(rawURL, name string) error {
	info, err := github.ParseRepoURL(rawURL)
	if err != nil {
		return err
	}

	if name == "" {
		name = info.Repo
	}

	// Clone into projects_dir if configured, otherwise the current directory.
	baseDir, _ := filepath.Abs(".")

	if global, err := config.LoadGlobal(); err == nil && global.ProjectsDir != "" {
		baseDir = global.ProjectsDir
	}

	dest := filepath.Join(baseDir, name)

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return fmt.Errorf("creating projects directory: %w", err)
	}

	fmt.Printf("Cloning %s/%s into %s\n", info.Owner, info.Repo, dest)

	if err := git.Clone(info.CloneURL, dest); err != nil {
		return err
	}

	absPath, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	cfg := config.ProjectConfig{
		Repo: absPath,
	}

	if err := config.SaveProject(name, cfg); err != nil {
		return err
	}

	fmt.Printf("Registered project %q (%s)\n", name, absPath)

	// Detect the default branch of the freshly cloned repo.
	branch, err := git.DefaultBranch(absPath)
	if err != nil {
		return err
	}

	rc, err := config.Resolve(name)
	if err != nil {
		return err
	}

	// Open a tmux session if tmux is running.
	if err := tmux.RequireRunning(); err != nil {
		slog.Debug("tmux not running, skipping session", "err", err)
		return nil
	}

	result, err := forest.AddTree(rc, branch)
	if err != nil {
		return err
	}

	if err := forest.OpenSession(rc, branch, result.WorktreePath); err != nil {
		return err
	}

	slog.Debug("switching to tmux session", "session", result.SessionName)

	return tmux.SwitchTo(result.SessionName)
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
