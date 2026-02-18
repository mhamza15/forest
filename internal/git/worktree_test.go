package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initTestRepo creates a bare-minimum git repo with one commit so
// that worktree operations have something to branch from.
func initTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")

	require.NoError(t, os.MkdirAll(repo, 0o755))

	commands := [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = repo
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, out)
	}

	return repo
}

func TestAddAndList(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "feature")

	require.NoError(t, Add(repo, wtPath, "feature", "main"))

	trees, err := List(repo)
	require.NoError(t, err)

	// There should be the main worktree plus the one we just added.
	assert.GreaterOrEqual(t, len(trees), 2)

	// Resolve symlinks because macOS /var -> /private/var.
	wtPathResolved, err := filepath.EvalSymlinks(wtPath)
	require.NoError(t, err)

	var found bool
	for _, wt := range trees {
		if wt.Branch == "feature" {
			found = true
			assert.Equal(t, wtPathResolved, wt.Path)
		}
	}

	assert.True(t, found, "expected to find worktree with branch 'feature'")
}

func TestRemove(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "to-remove")

	require.NoError(t, Add(repo, wtPath, "to-remove", "main"))
	require.NoError(t, Remove(repo, wtPath))

	trees, err := List(repo)
	require.NoError(t, err)

	for _, wt := range trees {
		assert.NotEqual(t, "to-remove", wt.Branch)
	}
}

func TestAdd_InvalidBase(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "bad")

	err := Add(repo, wtPath, "bad", "nonexistent-branch")
	assert.Error(t, err)
}

func TestParsePorcelain(t *testing.T) {
	input := `worktree /home/user/repo
HEAD abc123
branch refs/heads/main

worktree /home/user/trees/feature
HEAD def456
branch refs/heads/feature

worktree /home/user/bare-repo
HEAD 000000
bare

`

	trees := parsePorcelain([]byte(input))

	require.Len(t, trees, 3)

	assert.Equal(t, "/home/user/repo", trees[0].Path)
	assert.Equal(t, "main", trees[0].Branch)
	assert.False(t, trees[0].Bare)

	assert.Equal(t, "/home/user/trees/feature", trees[1].Path)
	assert.Equal(t, "feature", trees[1].Branch)

	assert.Equal(t, "/home/user/bare-repo", trees[2].Path)
	assert.True(t, trees[2].Bare)
}

func TestParsePorcelain_NoTrailingNewline(t *testing.T) {
	input := `worktree /path
HEAD abc
branch refs/heads/main`

	trees := parsePorcelain([]byte(input))

	require.Len(t, trees, 1)
	assert.Equal(t, "main", trees[0].Branch)
}

func TestSafeBranchDir(t *testing.T) {
	tests := []struct {
		branch string
		want   string
	}{
		{branch: "main", want: "main"},
		{branch: "feature/login", want: "feature-login"},
		{branch: "dependabot/go_modules/all-ac6d7e69db", want: "dependabot-go_modules-all-ac6d7e69db"},
		{branch: "no-slashes", want: "no-slashes"},
	}

	for _, tt := range tests {
		t.Run(tt.branch, func(t *testing.T) {
			assert.Equal(t, tt.want, SafeBranchDir(tt.branch))
		})
	}
}

func TestPruneCheck(t *testing.T) {
	repo := initTestRepo(t)

	// Create a branch that is merged (ancestor of main).
	mergedBranch := "already-merged"
	cmd := exec.Command("git", "-C", repo, "branch", mergedBranch, "main")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "creating branch: %s", out)

	// Create a branch with an extra commit that is not merged.
	unmergedBranch := "not-merged"
	cmd = exec.Command("git", "-C", repo, "branch", unmergedBranch, "main")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "creating branch: %s", out)

	cmd = exec.Command("git", "-C", repo, "checkout", unmergedBranch)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "checkout: %s", out)

	cmd = exec.Command("git", "-C", repo, "commit", "--allow-empty", "-m", "diverge")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "commit: %s", out)

	cmd = exec.Command("git", "-C", repo, "checkout", "main")
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "checkout main: %s", out)

	tests := []struct {
		name           string
		branch         string
		remoteBranches map[string]bool
		want           PruneReason
	}{
		{
			name:           "merged into target",
			branch:         mergedBranch,
			remoteBranches: map[string]bool{mergedBranch: true},
			want:           PruneMerged,
		},
		{
			name:           "not merged and present on remote",
			branch:         unmergedBranch,
			remoteBranches: map[string]bool{unmergedBranch: true},
			want:           PruneNone,
		},
		{
			name:           "not merged and gone from remote",
			branch:         unmergedBranch,
			remoteBranches: map[string]bool{},
			want:           PruneRemoteGone,
		},
		{
			name:           "merged takes precedence over remote gone",
			branch:         mergedBranch,
			remoteBranches: map[string]bool{},
			want:           PruneMerged,
		},
		{
			name:           "nil remote branches skips remote check",
			branch:         unmergedBranch,
			remoteBranches: nil,
			want:           PruneNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PruneCheck(repo, tt.branch, "main", tt.remoteBranches)
			assert.Equal(t, tt.want, got)
		})
	}
}

// initTestRepoWithRemote creates a "remote" repo containing the given
// branch, and clones it into a "local" repo. This produces a local
// repo with an "origin" remote so that fetch and tracking behavior
// can be tested. Returns (local, remote) repo paths.
func initTestRepoWithRemote(t *testing.T, remoteBranch string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	remote := filepath.Join(dir, "remote")
	local := filepath.Join(dir, "local")

	require.NoError(t, os.MkdirAll(remote, 0o755))

	// Set up the remote repo with an initial commit and branch.
	setup := [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "test"},
		{"git", "commit", "--allow-empty", "-m", "init"},
		{"git", "branch", remoteBranch},
	}

	for _, args := range setup {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = remote
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, out)
	}

	// Clone so that "origin" points back to the remote.
	cmd := exec.Command("git", "clone", remote, local)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "clone failed: %s", out)

	return local, remote
}

func TestFetchRemoteBranch(t *testing.T) {
	local, _ := initTestRepoWithRemote(t, "feature")

	// The branch should not exist locally after a fresh clone that
	// checked out only the default branch.
	assert.False(t, BranchExists(local, "feature"))

	remote, err := FetchRemoteBranch(local, "feature")
	require.NoError(t, err)
	assert.Equal(t, "origin", remote)

	// After fetching, a remote tracking ref should exist.
	ref := remoteTrackingRef(local, "feature")
	assert.Equal(t, "origin/feature", ref)
}

func TestFetchRemoteBranch_NotFound(t *testing.T) {
	local, _ := initTestRepoWithRemote(t, "feature")

	_, err := FetchRemoteBranch(local, "nonexistent")
	assert.Error(t, err)
}

func TestFetchRemoteBranch_NoRemotes(t *testing.T) {
	repo := initTestRepo(t)

	_, err := FetchRemoteBranch(repo, "feature")
	assert.Error(t, err)
}

func TestAdd_RemoteBranch(t *testing.T) {
	local, _ := initTestRepoWithRemote(t, "feature")
	wtPath := filepath.Join(t.TempDir(), "feature")

	// Fetch the branch so the remote tracking ref is available.
	_, err := FetchRemoteBranch(local, "feature")
	require.NoError(t, err)

	// The local branch should not exist yet.
	assert.False(t, BranchExists(local, "feature"))

	// Add should detect the tracking ref and create a tracking
	// worktree rather than branching off the base.
	require.NoError(t, Add(local, wtPath, "feature", "main"))
	assert.True(t, BranchExists(local, "feature"))

	// Verify upstream is set to origin/feature.
	cmd := exec.Command("git", "-C", wtPath, "config", "branch.feature.remote")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "origin", strings.TrimSpace(string(out)))

	cmd = exec.Command("git", "-C", wtPath, "config", "branch.feature.merge")
	out, err = cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "refs/heads/feature", strings.TrimSpace(string(out)))
}

func TestRemoteTrackingRef_None(t *testing.T) {
	repo := initTestRepo(t)

	// No remotes configured, so no tracking ref should exist.
	assert.Empty(t, remoteTrackingRef(repo, "feature"))
}
