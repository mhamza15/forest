// Package github parses GitHub issue and pull request URLs and
// retrieves PR metadata via the gh CLI.
package github

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

// LinkKind distinguishes between an issue and a pull request.
type LinkKind int

const (
	KindIssue LinkKind = iota
	KindPR
)

// Link holds the parsed components of a GitHub issue or PR URL.
type Link struct {
	// Kind is KindIssue or KindPR.
	Kind LinkKind

	// Owner is the repository owner (user or organization).
	Owner string

	// Repo is the repository name.
	Repo string

	// Number is the issue or PR number.
	Number int
}

// NWO returns the owner/repo string (name with owner).
func (l Link) NWO() string {
	return l.Owner + "/" + l.Repo
}

// ParseLink extracts owner, repo, number, and kind from a GitHub URL.
//
// Accepted formats:
//
//	https://github.com/owner/repo/issues/42
//	https://github.com/owner/repo/pull/99
func ParseLink(raw string) (Link, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return Link{}, fmt.Errorf("parsing URL: %w", err)
	}

	if u.Host != "github.com" {
		return Link{}, fmt.Errorf("unsupported host %q, expected github.com", u.Host)
	}

	// Path is /owner/repo/issues/42 or /owner/repo/pull/99, possibly
	// followed by extra segments like /files or /checks.
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")

	if len(parts) < 4 {
		return Link{}, fmt.Errorf("unexpected path %q: want /owner/repo/{issues,pull}/number", u.Path)
	}

	owner := parts[0]
	repo := parts[1]
	kindStr := parts[2]
	numStr := parts[3]

	var kind LinkKind

	switch kindStr {
	case "issues":
		kind = KindIssue
	case "pull":
		kind = KindPR
	default:
		return Link{}, fmt.Errorf("unexpected path segment %q: want \"issues\" or \"pull\"", kindStr)
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return Link{}, fmt.Errorf("parsing number %q: %w", numStr, err)
	}

	return Link{
		Kind:   kind,
		Owner:  owner,
		Repo:   repo,
		Number: num,
	}, nil
}

// IsGitHubURL returns true if the string looks like a GitHub URL.
func IsGitHubURL(s string) bool {
	return strings.HasPrefix(s, "https://github.com/")
}

// RepoInfo holds the owner and repository name parsed from a GitHub URL.
type RepoInfo struct {
	// Owner is the repository owner (user or organization).
	Owner string

	// Repo is the repository name.
	Repo string

	// CloneURL is the HTTPS clone URL for the repository.
	CloneURL string
}

// ParseRepoURL extracts the owner and repository name from a GitHub URL.
// It accepts plain repository URLs as well as issue or PR URLs,
// stripping extra path segments beyond owner/repo.
//
// Accepted formats:
//
//	https://github.com/owner/repo
//	https://github.com/owner/repo.git
//	https://github.com/owner/repo/issues/42
//	https://github.com/owner/repo/pull/99
func ParseRepoURL(raw string) (RepoInfo, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return RepoInfo{}, fmt.Errorf("parsing URL: %w", err)
	}

	if u.Host != "github.com" {
		return RepoInfo{}, fmt.Errorf("unsupported host %q, expected github.com", u.Host)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return RepoInfo{}, fmt.Errorf("cannot parse owner/repo from %q", raw)
	}

	owner := parts[0]
	repo := strings.TrimSuffix(parts[1], ".git")

	return RepoInfo{
		Owner:    owner,
		Repo:     repo,
		CloneURL: fmt.Sprintf("https://github.com/%s/%s.git", owner, repo),
	}, nil
}

// PRHead holds the head branch information for a pull request,
// including the fork's clone URL when the PR comes from a fork.
type PRHead struct {
	// Branch is the head branch name.
	Branch string

	// CloneURL is the HTTPS clone URL for the head repository. For
	// same-repo PRs this matches the base repo. For forks it points
	// to the contributor's fork.
	CloneURL string

	// IsFork is true when the head repo differs from the base repo.
	IsFork bool
}

// ghPRJSON is the subset of gh pr view --json output that we need.
type ghPRJSON struct {
	HeadRefName       string      `json:"headRefName"`
	HeadRepo          ghRepoJSON  `json:"headRepository"`
	HeadRepoOwner     ghOwnerJSON `json:"headRepositoryOwner"`
	IsCrossRepository bool        `json:"isCrossRepository"`
}

type ghRepoJSON struct {
	Name string `json:"name"`
}

type ghOwnerJSON struct {
	Login string `json:"login"`
}

// FetchPRHead retrieves the head branch and clone URL of a pull
// request using the gh CLI. The nwo argument is the "owner/repo"
// string for the base repository.
func FetchPRHead(nwo string, number int) (PRHead, error) {
	// gh pr view requires a repo flag when we are not inside the repo.
	num := strconv.Itoa(number)

	cmd := exec.Command(
		"gh", "pr", "view", num,
		"--repo", nwo,
		"--json", "headRefName,headRepository,headRepositoryOwner,isCrossRepository",
	)

	output, err := cmd.Output()
	if err != nil {
		return PRHead{}, fmt.Errorf("gh pr view: %w", err)
	}

	var pr ghPRJSON
	if err := json.Unmarshal(output, &pr); err != nil {
		return PRHead{}, fmt.Errorf("parsing gh output: %w", err)
	}

	cloneURL := fmt.Sprintf(
		"https://github.com/%s/%s.git",
		pr.HeadRepoOwner.Login,
		pr.HeadRepo.Name,
	)

	return PRHead{
		Branch:   pr.HeadRefName,
		CloneURL: cloneURL,
		IsFork:   pr.IsCrossRepository,
	}, nil
}

// IsPRMerged checks whether a merged pull request exists for the
// given branch in the repository identified by nwo ("owner/repo").
// It uses gh pr list to find merged PRs matching the head branch.
// Returns an error if the gh CLI is unavailable or the query fails.
func IsPRMerged(nwo, branch string) (bool, error) {
	cmd := exec.Command(
		"gh", "pr", "list",
		"--head", branch,
		"--state", "merged",
		"--repo", nwo,
		"--json", "number",
		"--limit", "1",
	)

	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("gh pr list: %w", err)
	}

	// gh pr list --json returns "[]\n" when no PRs match.
	return len(output) > 0 && strings.TrimSpace(string(output)) != "[]", nil
}
