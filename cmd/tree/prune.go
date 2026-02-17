package tree

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/github"
)

var dryRunFlag bool

func pruneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove worktrees whose branches have been merged or deleted",
		Long: `Check each worktree's branch and remove it if it has been merged
into the project's base branch. When a branch no longer exists on
the remote, forest checks via gh whether the PR was merged (common
after squash-merge workflows). If gh is unavailable or the PR was
not merged, an interactive confirmation is shown instead.`,
		Args: cobra.NoArgs,
		RunE: runPrune,
	}

	cmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "show what would be pruned without removing")

	return cmd
}

func runPrune(cmd *cobra.Command, args []string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	var names []string
	if projectFlag != "" {
		names = []string{projectFlag}
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
			slog.Debug("could not fetch remote branches", slog.String("project", name), slog.Any("err", err))
		}

		// Resolve NWO once per project for gh PR lookups. A failure
		// here is non-fatal; we fall back to interactive confirmation.
		nwo := resolveNWO(rc.Repo)

		repoPath := filepath.Clean(rc.Repo)

		for _, t := range trees {
			if t.Bare || t.Branch == "" || t.Branch == rc.Branch {
				continue
			}

			// Skip the main working tree. It cannot be removed by
			// git worktree remove and should never be pruned.
			if filepath.Clean(t.Path) == repoPath {
				continue
			}

			reason := git.PruneCheck(rc.Repo, t.Branch, rc.Branch, remoteBranches)
			if reason == git.PruneNone {
				continue
			}

			// When the branch is gone from the remote but not merged
			// locally, verify via gh that the PR was actually merged.
			// Fall back to an interactive prompt when gh is
			// unavailable or the PR was not merged.
			if reason == git.PruneRemoteGone {
				if !shouldPruneRemoteGone(nwo, name, t.Branch) {
					continue
				}
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

// shouldPruneRemoteGone determines whether a branch whose remote
// tracking branch has been deleted should be pruned. It first tries
// the gh CLI to check for a merged PR. If gh confirms the PR was
// merged, pruning proceeds. Otherwise, the user is prompted.
func shouldPruneRemoteGone(nwo, project, branch string) bool {
	if nwo != "" {
		merged, err := github.IsPRMerged(nwo, branch)
		if err != nil {
			slog.Debug("gh PR check failed, falling back to prompt",
				slog.String("branch", branch),
				slog.Any("err", err),
			)
		}

		if merged {
			return true
		}
	}

	return confirm(fmt.Sprintf(
		"Branch %s/%s is gone from the remote but may not be merged. Remove? [y/N] ",
		project, branch,
	))
}

// resolveNWO returns the "owner/repo" string for the origin remote,
// or an empty string if it cannot be determined. An empty result
// causes the caller to skip gh lookups and fall back to prompting.
func resolveNWO(repoPath string) string {
	raw, err := git.RemoteURL(repoPath, "origin")
	if err != nil {
		return ""
	}

	return git.NormalizeRemoteURL(raw)
}
