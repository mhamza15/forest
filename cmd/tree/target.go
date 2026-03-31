package tree

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/github"
)

func resolveTreeTarget(cmd *cobra.Command, arg string) (string, string, config.ResolvedConfig, error) {
	projectFlag, _ := cmd.Flags().GetString("project")

	var (
		project string
		branch  string
		rc      config.ResolvedConfig
	)

	if github.IsGitHubURL(arg) {
		link, err := github.ParseLink(arg)
		if err != nil {
			return "", "", rc, err
		}

		name, resolved, err := config.FindProjectByRemote(link.NWO())
		if err != nil {
			return "", "", rc, err
		}

		project = name
		rc = resolved

		branch, err = resolveLinkBranch(link, rc.Repo)
		if err != nil {
			return "", "", rc, err
		}
	} else {
		branch = arg

		name, err := resolveProject(projectFlag)
		if err != nil {
			return "", "", rc, err
		}

		project = name

		resolved, err := config.Resolve(project)
		if err != nil {
			return "", "", rc, err
		}

		rc = resolved
	}

	if baseBranchFlag != "" {
		rc.Branch = baseBranchFlag
	}

	return project, branch, rc, nil
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

		// For fork PRs, prefix the local branch with the fork owner
		// so it does not collide with identically named branches in
		// the base repository.
		localBranch := head.Branch
		if head.IsFork {
			localBranch = head.ForkOwner + "/" + head.Branch
		}

		if head.IsFork {
			if err := git.EnsureRemote(repoPath, head.ForkOwner, head.CloneURL); err != nil {
				return "", fmt.Errorf("adding fork remote: %w", err)
			}
		}

		// The branch may not exist locally. Fetch it from the head
		// repository. Skip the fetch when the local branch already
		// exists. For same-repo PRs, the fetch writes directly to the
		// local branch name. That fails when the branch is already
		// checked out in another worktree.
		//
		// For fork PRs, the local branch name is prefixed with the fork
		// owner to avoid collisions. The upstream should still point to
		// the fork's actual branch name, not the prefixed local branch.
		if !git.BranchExists(repoPath, localBranch) {
			if head.IsFork {
				fmt.Printf("Fetching branch %q from %s\n", head.Branch, head.ForkOwner)

				if err := git.FetchBranch(repoPath, head.ForkOwner, head.Branch); err != nil {
					return "", fmt.Errorf("fetching branch: %w", err)
				}

				if err := git.CreateTrackingBranch(repoPath, localBranch, head.ForkOwner, head.Branch); err != nil {
					return "", fmt.Errorf("creating tracking branch: %w", err)
				}
			} else {
				fmt.Printf("Fetching branch %q from %s\n", localBranch, head.CloneURL)

				if err := git.Fetch(repoPath, head.CloneURL, head.Branch, localBranch); err != nil {
					return "", fmt.Errorf("fetching branch: %w", err)
				}
			}
		} else if head.IsFork {
			// Older worktrees may have been created before fork PR
			// upstreams were configured explicitly. Repair them when the
			// user reopens the PR by URL.
			if err := git.SetBranchUpstream(repoPath, localBranch, head.ForkOwner, head.Branch); err != nil {
				return "", fmt.Errorf("setting branch upstream: %w", err)
			}
		}

		return localBranch, nil

	default:
		return "", fmt.Errorf("unexpected link kind: %d", link.Kind)
	}
}

// resolveProject determines the project name from the flag value,
// falling back to inference from the working directory.
func resolveProject(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	return config.InferProject()
}
