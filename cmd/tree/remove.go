package tree

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
)

var forceFlag bool

func removeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove [branch]",
		Short: "Remove a worktree and its tmux session",
		Long: `Remove a worktree and its tmux session. With no arguments, detects
the current worktree from the working directory and prompts for
confirmation.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              runRemove,
		ValidArgsFunction: completion.Branches,
	}

	cmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "force removal of dirty worktrees")

	return cmd
}

func runRemove(cmd *cobra.Command, args []string) error {
	var project, branch string

	projectFlag, _ := cmd.Flags().GetString("project")

	switch len(args) {
	case 1:
		branch = args[0]

		var err error

		project, err = resolveProject(projectFlag)
		if err != nil {
			return err
		}

	case 0:
		var err error

		project, branch, err = detectCurrentWorktree()
		if err != nil {
			return err
		}

		if !confirm(fmt.Sprintf("Remove worktree %s/%s? [y/N] ", project, branch)) {
			return nil
		}
	}

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	err = forest.RemoveTree(rc, branch, forceFlag)
	if err != nil {
		if !errors.Is(err, git.ErrWorktreeDirty) {
			return err
		}

		fmt.Printf("Worktree %s/%s has modified or untracked files.\n", project, branch)

		if !confirm("Force remove? [y/N] ") {
			return nil
		}

		if err := forest.RemoveTree(rc, branch, true); err != nil {
			return err
		}
	}

	fmt.Printf("Removed worktree %s/%s\n", project, branch)

	return nil
}

// confirm prints a prompt and returns true if the user types y or yes.
func confirm(prompt string) bool {
	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	return answer == "y" || answer == "yes"
}

// detectCurrentWorktree figures out which project and branch the
// current working directory belongs to by matching the worktree root
// against registered projects.
func detectCurrentWorktree() (project string, branch string, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("getting working directory: %w (was this worktree already removed?)", err)
	}

	if _, statErr := os.Stat(cwd); statErr != nil {
		return "", "", fmt.Errorf("current directory no longer exists (was this worktree already removed?)")
	}

	currentBranch := git.CurrentBranch(cwd)
	if currentBranch == "" {
		return "", "", fmt.Errorf("not in a git worktree")
	}

	wtRoot := git.WorktreeRoot(cwd)
	if wtRoot == "" {
		return "", "", fmt.Errorf("could not determine worktree root")
	}

	names, err := config.ListProjects()
	if err != nil {
		return "", "", err
	}

	for _, name := range names {
		proj, err := config.LoadProject(name)
		if err != nil {
			continue
		}

		trees, err := git.List(proj.Repo)
		if err != nil {
			continue
		}

		for _, t := range trees {
			if filepath.Clean(t.Path) == filepath.Clean(wtRoot) {
				return name, t.Branch, nil
			}
		}
	}

	return "", "", fmt.Errorf("current directory is not a registered forest worktree")
}
