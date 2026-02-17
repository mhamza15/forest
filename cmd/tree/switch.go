package tree

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mhamza15/forest/internal/completion"
	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/forest"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/github"
	"github.com/mhamza15/forest/internal/tmux"
)

func switchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch {<branch> | <github-link>}",
		Short: "Switch to an existing worktree's tmux session",
		Long: `Switch to the tmux session for an existing worktree. If the
worktree exists but the session does not, the session is created
with the configured layout. Does not create new worktrees.

A GitHub issue or pull request URL may be passed instead of a branch
name. The project is resolved from the URL and the branch is derived
from the link (issue-<number> for issues, the PR head branch for PRs).`,
		Args:              cobra.ExactArgs(1),
		RunE:              runSwitch,
		ValidArgsFunction: completion.Branches,
	}
}

func runSwitch(cmd *cobra.Command, args []string) error {
	projectFlag, _ := cmd.Flags().GetString("project")

	var (
		project string
		branch  string
		rc      config.ResolvedConfig
	)

	if github.IsGitHubURL(args[0]) {
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

	existing := git.FindByBranch(rc.Repo, branch)
	if existing == nil {
		return fmt.Errorf("no worktree for branch %q in project %q", branch, project)
	}

	if err := forest.OpenSession(rc, branch, existing.Path); err != nil {
		return err
	}

	return tmux.SwitchTo(tmux.SessionName(rc.Name, branch))
}

// resolveProject determines the project name from the flag value,
// falling back to inference from the working directory.
func resolveProject(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	return config.InferProject()
}
