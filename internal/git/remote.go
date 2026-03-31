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

// EnsureRemote adds a named remote if it does not already exist.
// If the remote name is already configured, the error is silently
// ignored so callers can safely call this unconditionally.
func EnsureRemote(repoPath, name, url string) error {
	cmd := exec.Command("git", "-C", repoPath, "remote", "add", name, url)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Ignore "already exists" errors.
		if bytes.Contains(output, []byte("already exists")) {
			return nil
		}

		return fmt.Errorf("git remote add: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// FetchBranch fetches a single branch from a named remote, updating
// the corresponding remote tracking ref (e.g. refs/remotes/<remote>/<branch>).
func FetchBranch(repoPath, remote, branch string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", remote, branch)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git fetch: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// CreateTrackingBranch creates a local branch from the given remote
// tracking ref and then configures its upstream explicitly.
//
// The local branch name and the remote tracking ref may both contain
// slashes. For fork PRs they intentionally differ:
//
//	local branch:  contributor/fix-bug
//	upstream ref:  refs/remotes/contributor/fix-bug
//
// Configuring the remote and merge keys explicitly avoids relying on
// Git to infer the intended upstream from an ambiguous owner/branch
// token.
func CreateTrackingBranch(repoPath, branch, remote, remoteBranch string) error {
	upstreamRef := fmt.Sprintf("refs/remotes/%s/%s", remote, remoteBranch)

	cmd := exec.Command("git", "-C", repoPath, "branch", "--no-track", branch, upstreamRef)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch: %s: %w", bytes.TrimSpace(output), err)
	}

	if err := SetBranchUpstream(repoPath, branch, remote, remoteBranch); err != nil {
		return fmt.Errorf("setting branch upstream: %w", err)
	}

	return nil
}

// SetBranchUpstream configures the upstream branch for the given local
// branch by writing the same remote and merge keys that Git uses for
// tracked branches.
func SetBranchUpstream(repoPath, branch, remote, remoteBranch string) error {
	if err := setBranchConfig(repoPath, branch, "remote", remote); err != nil {
		return err
	}

	if err := setBranchConfig(repoPath, branch, "merge", "refs/heads/"+remoteBranch); err != nil {
		return err
	}

	return nil
}

func setBranchConfig(repoPath, branch, key, value string) error {
	name := fmt.Sprintf("branch.%s.%s", branch, key)

	cmd := exec.Command("git", "-C", repoPath, "config", name, value)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git config %s: %s: %w", name, bytes.TrimSpace(output), err)
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
