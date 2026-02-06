package tree

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <project> <branch>",
		Short: "Create a new worktree and tmux session",
		Long: `Create a new git worktree for the project and open it in a tmux session.

If the worktree already exists, the command switches to the existing
tmux session instead. The new branch is based on the project's configured
base branch, falling back to the global default.`,
		Args:              cobra.ExactArgs(2),
		RunE:              runAdd,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

func runAdd(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	if err := tmux.RequireRunning(); err != nil {
		return err
	}

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	sessionName := tmux.SessionName(rc.Name, branch)

	// Check if a worktree for this branch already exists (at any path,
	// including paths created before the SafeBranchDir convention).
	var wtPath string

	if existing := git.FindByBranch(rc.Repo, branch); existing != nil {
		wtPath = existing.Path
		fmt.Printf("Worktree %s/%s already exists\n", project, branch)
	} else {
		wtPath = filepath.Join(rc.WorktreeDir, rc.Name, git.SafeBranchDir(branch))

		slog.Debug("creating worktree", "path", wtPath, "base", rc.Branch)

		if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
			return fmt.Errorf("creating worktree parent dir: %w", err)
		}

		if err := git.Add(rc.Repo, wtPath, branch, rc.Branch); err != nil {
			return err
		}

		fmt.Printf("Created worktree %s/%s\n", project, branch)

		if len(rc.Copy) > 0 {
			warnings := git.CopyFiles(rc.Repo, wtPath, rc.Copy)
			for _, w := range warnings {
				fmt.Println(w)
			}
		}
	}

	if !tmux.SessionExists(sessionName) {
		if err := tmux.NewSession(sessionName, wtPath); err != nil {
			return err
		}

		if len(rc.Layout) > 0 {
			windows := make([]tmux.LayoutWindow, len(rc.Layout))
			for i, w := range rc.Layout {
				windows[i] = tmux.LayoutWindow{Name: w.Name, Command: w.Command}
			}

			if err := tmux.ApplyLayout(sessionName, wtPath, windows); err != nil {
				return err
			}
		}
	}

	slog.Debug("switching to tmux session", "session", sessionName)

	return tmux.SwitchTo(sessionName)
}
