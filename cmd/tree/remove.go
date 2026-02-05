package tree

import (
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <project> <branch>",
		Short: "Remove a worktree and its tmux session",
		Args:  cobra.ExactArgs(2),
		RunE:  runRemove,
	}
}

func runRemove(cmd *cobra.Command, args []string) error {
	project := args[0]
	branch := args[1]

	rc, err := config.Resolve(project)
	if err != nil {
		return err
	}

	sessionName := tmux.SessionName(rc.Name, branch)
	wtPath := filepath.Join(rc.WorktreeDir, rc.Name, branch)

	if err := tmux.KillSession(sessionName); err != nil {
		slog.Warn("could not kill tmux session", "session", sessionName, "err", err)
	}

	if err := git.Remove(rc.Repo, wtPath); err != nil {
		return err
	}

	slog.Info("worktree removed", "project", project, "branch", branch)

	return nil
}
