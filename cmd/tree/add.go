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

var (
	baseBranchFlag string
	noSessionFlag  bool
)

func addCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add {<branch> | <github-link>}",
		Short: "Create a new worktree and tmux session",
		Long: "Create a new git worktree for the project and open it in a tmux session.\n" +
			"\n" +
			"The project is resolved from the --project flag when set, or inferred\n" +
			"from the current working directory by matching its git remote against\n" +
			"registered projects.\n" +
			"\n" +
			"If the worktree already exists, the command switches to the existing\n" +
			"tmux session instead. The new branch is based on the project's configured\n" +
			"base branch, falling back to the global default. Use --branch to override.\n" +
			"\n" +
			"A GitHub issue or pull request URL may be passed instead of a branch:\n" +
			"\n" +
			"  forest tree add https://github.com/owner/repo/issues/42\n" +
			"  forest tree add https://github.com/owner/repo/pull/99\n" +
			"\n" +
			"For issues, a branch named \"issue-<number>\" is created (e.g. \"issue-42\").\n" +
			"For pull requests, the PR's head branch is fetched. If the PR comes\n" +
			"from a fork, the branch is fetched from the fork's remote.\n" +
			"\n" +
			"When using a GitHub link, the project is determined automatically by\n" +
			"matching the repository in the URL against the origin remotes of\n" +
			"registered projects.",
		Args:              cobra.ExactArgs(1),
		RunE:              runAdd,
		ValidArgsFunction: completion.Branches,
	}

	cmd.Flags().StringVarP(&baseBranchFlag, "branch", "b", "", "base branch for the new worktree (overrides project config)")
	cmd.Flags().BoolVar(&noSessionFlag, "no-session", false, "create worktree without opening a tmux session")

	return cmd
}

// isGitHubLink returns true if the argument looks like a GitHub URL
// rather than a plain branch name.
func isGitHubLink(s string) bool {
	return strings.HasPrefix(s, "https://github.com/")
}

func runAdd(cmd *cobra.Command, args []string) error {
	var (
		project string
		branch  string
		rc      config.ResolvedConfig
	)

	projectFlag, _ := cmd.Flags().GetString("project")

	if isGitHubLink(args[0]) {
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
	} else {
		branch = args[0]

		var err error

		project, err = resolveProject(projectFlag)
		if err != nil {
			return err
		}

		resolved, err := config.Resolve(project)
		if err != nil {
			return err
		}

		rc = resolved
	}

	if baseBranchFlag != "" {
		rc.Branch = baseBranchFlag
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

	for _, w := range result.SymlinkWarnings {
		fmt.Println(w)
	}

	if noSessionFlag {
		return nil
	}
	if err := forest.OpenSession(rc, branch, result.WorktreePath); err != nil {
		return err
	}

	slog.Debug("switching to tmux session", slog.String("session", result.SessionName))
	return tmux.SwitchTo(result.SessionName)
}

// resolveLinkBranch determines the branch name for a GitHub link. For
// issues it returns the issue number as the branch name. For PRs it
// fetches the head branch from the appropriate remote when the branch
// does not exist locally.
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
			slog.String("branch", head.Branch),
			slog.Bool("fork", head.IsFork),
			slog.String("clone_url", head.CloneURL),
		)

		// The branch may not exist locally. Fetch it from the head
		// repository's clone URL, which is the fork URL for cross-repo
		// PRs and the base repo URL for same-repo PRs. We always use
		// the clone URL rather than "origin" because the user's origin
		// remote may point to their own fork, not the PR's repository.
		// Skip the fetch if the branch already exists -- it may already
		// be checked out in a worktree, and git refuses to fetch into a
		// checked-out branch.
		if !git.BranchExists(repoPath, head.Branch) {
			fmt.Printf("Fetching branch %q from %s\n", head.Branch, head.CloneURL)

			if err := git.Fetch(repoPath, head.CloneURL, head.Branch, head.Branch); err != nil {
				return "", fmt.Errorf("fetching branch: %w", err)
			}
		}

		return head.Branch, nil

	default:
		return "", fmt.Errorf("unexpected link kind: %d", link.Kind)
	}
}
