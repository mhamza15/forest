package tree

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/tmux"
)

var baseBranchFlag string

func switchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "switch {<branch> | <github-link>}",
		Short: "Switch to a worktree, creating it if needed",
		Long: "Switch to the tmux session for a worktree.\n" +
			"\n" +
			"If the worktree does not already exist, it is created first. If the\n" +
			"tmux session does not already exist, the session is created with the\n" +
			"configured layout.\n" +
			"\n" +
			"The project is resolved from the --project flag when set, or inferred\n" +
			"from the current working directory by matching its git remote against\n" +
			"registered projects.\n" +
			"\n" +
			"If the branch does not exist locally but exists on a remote, it is\n" +
			"fetched and the local branch is created with upstream tracking\n" +
			"configured. Otherwise, a new branch is created based on the project's\n" +
			"configured base branch, falling back to the global default. Use\n" +
			"--branch to override the base branch for new worktrees.\n" +
			"\n" +
			"A GitHub issue or pull request URL may be passed instead of a branch:\n" +
			"\n" +
			"  forest tree switch https://github.com/owner/repo/issues/42\n" +
			"  forest tree switch https://github.com/owner/repo/pull/99\n" +
			"\n" +
			"For issues, a branch named \"issue-<number>\" is used (e.g. \"issue-42\").\n" +
			"For pull requests, the PR's head branch is used. If the PR comes from\n" +
			"a fork, the branch is fetched from the fork's remote.",
		Args:              cobra.ExactArgs(1),
		RunE:              runSwitch,
		ValidArgsFunction: completion.Branches,
	}

	cmd.Flags().StringVarP(&baseBranchFlag, "branch", "b", "", "base branch for a new worktree (overrides project config)")

	return cmd
}

func runSwitch(cmd *cobra.Command, args []string) error {
	project, branch, rc, err := resolveTreeTarget(cmd, args[0])
	if err != nil {
		return err
	}

	result, err := forest.AddTree(rc, branch)
	if err != nil {
		return err
	}

	if result.Created {
		if result.Fetched {
			fmt.Printf("Fetched branch %q from %s\n", branch, result.Remote)
		}

		fmt.Printf("Created worktree %s/%s\n", project, branch)
	}

	for _, w := range result.CopyWarnings {
		fmt.Println(w)
	}

	for _, w := range result.SymlinkWarnings {
		fmt.Println(w)
	}

	if err := forest.OpenSession(rc, branch, result.WorktreePath); err != nil {
		return err
	}

	slog.Debug("switching to tmux session", slog.String("session", result.SessionName))

	return tmux.SwitchTo(result.SessionName)
}
