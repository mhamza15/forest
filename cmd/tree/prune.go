package tree

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
)

var dryRunFlag bool

func pruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune [project]",
		Short: "Remove worktrees whose branches have been merged or deleted",
		Long: `Check each worktree's branch and remove it if it has been merged
into the project's base branch, or if the branch no longer exists on
the remote (common after squash-merge workflows). The base branch
worktree itself is never pruned.`,
		Args:              cobra.MaximumNArgs(1),
		RunE:              runPrune,
		ValidArgsFunction: completion.Projects,
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "show what would be pruned without removing")

	return cmd
}

func runPrune(cmd *cobra.Command, args []string) error {
	var names []string

	if len(args) == 1 {
		names = []string{args[0]}
	} else {
		var err error

		names, err = config.ListProjects()
		if err != nil {
			return err
		}
	}

	pruned := 0

	for _, name := range names {
		rc, err := config.Resolve(name)
		if err != nil {
			return err
		}

		trees, err := git.List(rc.Repo)
		if err != nil {
			return err
		}

		// Fetch remote branches once per project so we can detect
		// branches deleted after a squash-merge PR.
		remoteBranches, err := git.RemoteBranches(rc.Repo, "origin")
		if err != nil {
			slog.Debug("could not fetch remote branches", "project", name, "err", err)
		}

		for _, t := range trees {
			if t.Bare || t.Branch == "" || t.Branch == rc.Branch {
				continue
			}

			if !git.IsPrunable(rc.Repo, t.Branch, rc.Branch, remoteBranches) {
				continue
			}

			if dryRunFlag {
				fmt.Printf("would prune %s/%s\n", name, t.Branch)
				pruned++
				continue
			}

			if err := forest.RemoveTree(rc, t.Branch, true); err != nil {
				fmt.Printf("failed to prune %s/%s: %s\n", name, t.Branch, err)
				continue
			}

			fmt.Printf("pruned %s/%s\n", name, t.Branch)
			pruned++
		}
	}

	if pruned == 0 {
		fmt.Println("Nothing to prune.")
	}

	return nil
}
