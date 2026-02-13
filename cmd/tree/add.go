package tree

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/github"
	"github.com/mhamza15/forest/internal/tmux"
)

func addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add {<project> <branch> | <github-link>}",
		Short: "Create a new worktree and tmux session",
		Long: `Create a new git worktree for the project and open it in a tmux session.

If the worktree already exists, the command switches to the existing
tmux session instead. The new branch is based on the project's configured
base branch, falling back to the global default.

A GitHub issue or pull request URL may be passed instead of the project
and branch arguments:

  forest tree add https://github.com/owner/repo/issues/42
  forest tree add https://github.com/owner/repo/pull/99

For issues, a branch named "issue-<number>" is created (e.g. "issue-42").
For pull requests, the PR's head branch is fetched. If the PR comes
from a fork, the branch is fetched from the fork's remote.

The project is determined automatically by matching the repository in
the URL against the origin remotes of registered projects.`,
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runAdd,
		ValidArgsFunction: completion.ProjectThenBranch,
	}
}

// isGitHubLink returns true if the argument looks like a GitHub URL
// rather than a plain project name.
func isGitHubLink(s string) bool {
	return strings.HasPrefix(s, "https://github.com/")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if err := tmux.RequireRunning(); err != nil {
		return err
	}

	var (
		project string
		branch  string
		rc      config.ResolvedConfig
	)

	switch {

	// forest tree add <github-link>
	case len(args) == 1 && isGitHubLink(args[0]):
		link, err := github.ParseLink(args[0])
		if err != nil {
			return err
		}

		name, resolved, err := config.FindProjectByRemote(link.NWO())
		if err != nil {
			return err
		}

		project = name
		rc = resolved

		branch, err = resolveLinkBranch(link, rc.Repo)
		if err != nil {
			return err
		}

	// forest tree add <project> <branch>
	case len(args) == 2:
		project = args[0]
		branch = args[1]

		resolved, err := config.Resolve(project)
		if err != nil {
			return err
		}

		rc = resolved

	default:
		return fmt.Errorf("expected <project> <branch> or a GitHub issue/PR link")
	}

	result, err := forest.AddTree(rc, branch)
	if err != nil {
		return err
	}

	if result.Created {
		fmt.Printf("Created worktree %s/%s\n", project, branch)
	} else {
		fmt.Printf("Worktree %s/%s already exists\n", project, branch)
	}

	for _, w := range result.CopyWarnings {
		fmt.Println(w)
	}

	if err := forest.OpenSession(rc, branch, result.WorktreePath); err != nil {
		return err
	}

	slog.Debug("switching to tmux session", "session", result.SessionName)

	return tmux.SwitchTo(result.SessionName)
}

// resolveLinkBranch determines the branch name for a GitHub link. For
// issues it returns the issue number as the branch name. For PRs it
// fetches the head branch, handling forks by fetching from the fork's
// remote when necessary.
func resolveLinkBranch(link github.Link, repoPath string) (string, error) {
	switch link.Kind {

	case github.KindIssue:
		return fmt.Sprintf("issue-%d", link.Number), nil

	case github.KindPR:
		head, err := github.FetchPRHead(link.NWO(), link.Number)
		if err != nil {
			return "", fmt.Errorf("fetching PR metadata: %w", err)
		}

		slog.Debug("resolved PR head",
			"branch", head.Branch,
			"fork", head.IsFork,
			"clone_url", head.CloneURL,
		)

		// For forks, the branch may not exist locally. Fetch it from
		// the contributor's remote so that git worktree add can check
		// it out. Skip the fetch if the branch already exists — it
		// may already be checked out in a worktree, and git refuses
		// to fetch into a checked-out branch.
		if head.IsFork && !git.BranchExists(repoPath, head.Branch) {
			fmt.Printf("Fetching branch %q from fork %s\n", head.Branch, head.CloneURL)
			if err := git.Fetch(repoPath, head.CloneURL, head.Branch, head.Branch); err != nil {
				return "", fmt.Errorf("fetching fork branch: %w", err)
			}
		}

		return head.Branch, nil

	default:
		return "", fmt.Errorf("unexpected link kind: %d", link.Kind)
	}
}
