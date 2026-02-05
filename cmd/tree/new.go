package tree

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <project> <branch>",
		Short: "Create a new worktree and tmux session",
		Long: `Create a new git worktree for the project and open it in a tmux session.

If the worktree already exists, the command switches to the existing
tmux session instead. The new branch is based on the project's configured
base branch, falling back to the global default.`,
		Args: cobra.ExactArgs(2),
		RunE: runNew,
	}
}

func runNew(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	if err := tmux.RequireRunning(); err != nil {
		return err
	}

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	wtPath := filepath.Join(rc.WorktreeDir, rc.Name, branch)
	sessionName := tmux.SessionName(rc.Name, branch)

	// If the worktree directory already exists, skip creation and just
	// ensure the tmux session is running.
	if _, err := os.Stat(wtPath); err != nil {
		slog.Debug("creating worktree", "path", wtPath, "base", rc.Branch)

		if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
			return fmt.Errorf("creating worktree parent dir: %w", err)
		}

		if err := git.Add(rc.Repo, wtPath, branch, rc.Branch); err != nil {
			return err
		}

		fmt.Printf("Created worktree %s/%s\n", project, branch)
	} else {
		fmt.Printf("Worktree %s/%s already exists\n", project, branch)
	}

	if !tmux.SessionExists(sessionName) {
		if err := tmux.NewSession(sessionName, wtPath); err != nil {
			return err
		}
	}

	slog.Debug("switching to tmux session", "session", sessionName)

	return tmux.SwitchTo(sessionName)
}
