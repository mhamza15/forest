package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Remotes returns the names of all configured remotes for the
// repository (e.g. "origin", "upstream").
func Remotes(repoPath string) ([]string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git remote: %w", err)
	}

	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil, nil
	}

	return strings.Split(raw, "\n"), nil
}

// RemoteURL returns the fetch URL for the given remote (typically
// "origin"). The result is the raw URL as configured in the repo,
// which may be HTTPS or SSH.
func RemoteURL(repoPath, remote string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", remote)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git remote get-url %s: %w", remote, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// NormalizeRemoteURL converts a git remote URL into a canonical
// "owner/repo" form so that HTTPS and SSH URLs for the same repo
// compare equal.
//
// Supported formats:
//
//	https://github.com/owner/repo.git
//	https://github.com/owner/repo
//	git@github.com:owner/repo.git
//	git@github.com:owner/repo
func NormalizeRemoteURL(raw string) string {
	s := raw

	// SSH format: git@github.com:owner/repo.git
	if after, ok := strings.CutPrefix(s, "git@github.com:"); ok {
		s = after
	}

	// HTTPS format: https://github.com/owner/repo.git
	if after, ok := strings.CutPrefix(s, "https://github.com/"); ok {
		s = after
	}

	s = strings.TrimSuffix(s, ".git")
	s = strings.TrimSuffix(s, "/")

	return s
}

// Fetch fetches a branch from a remote URL into the local repo. If
// localBranch differs from remoteBranch, it creates a local tracking
// branch. This is used to pull PR branches from forks.
func Fetch(repoPath, remoteURL, remoteBranch, localBranch string) error {
	refspec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", remoteBranch, localBranch)

	cmd := exec.Command("git", "-C", repoPath, "fetch", remoteURL, refspec)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// FetchRemoteBranch attempts to fetch a single branch from the
// repository's configured remotes. It tries each remote in order and
// returns the name of the first remote that provides the branch.
// The remote tracking ref (e.g., refs/remotes/origin/<branch>) is
// updated as a side effect, enabling subsequent git operations to
// create a local branch that tracks the remote.
func FetchRemoteBranch(repoPath, branch string) (string, error) {
	remotes, err := Remotes(repoPath)
	if err != nil || len(remotes) == 0 {
		return "", fmt.Errorf("no remotes configured")
	}

	for _, remote := range remotes {
		cmd := exec.Command("git", "-C", repoPath, "fetch", remote, branch)

		if _, err := cmd.CombinedOutput(); err == nil {
			return remote, nil
		}
	}

	return "", fmt.Errorf("branch %q not found on any remote", branch)
}
