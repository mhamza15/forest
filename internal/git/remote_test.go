package git

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "https with .git",
			raw:  "https://github.com/acme/widgets.git",
			want: "acme/widgets",
		},
		{
			name: "https without .git",
			raw:  "https://github.com/acme/widgets",
			want: "acme/widgets",
		},
		{
			name: "https trailing slash",
			raw:  "https://github.com/acme/widgets/",
			want: "acme/widgets",
		},
		{
			name: "ssh with .git",
			raw:  "git@github.com:acme/widgets.git",
			want: "acme/widgets",
		},
		{
			name: "ssh without .git",
			raw:  "git@github.com:acme/widgets",
			want: "acme/widgets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeRemoteURL(tt.raw))
		})
	}
}

func TestCreateTrackingBranch_PrefixedLocalBranch(t *testing.T) {
	local, remote := initTestRepoWithRemote(t, "docker-compose-examples")

	const (
		localBranch  = "mhamza15/docker-compose-examples"
		remoteName   = "mhamza15"
		remoteBranch = "docker-compose-examples"
	)

	cmd := exec.Command("git", "-C", local, "remote", "add", remoteName, remote)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git remote add: %s", output)

	require.NoError(t, FetchBranch(local, remoteName, remoteBranch))
	require.NoError(t, CreateTrackingBranch(local, localBranch, remoteName, remoteBranch))

	assert.Equal(t, remoteName, branchConfig(t, local, localBranch, "remote"))
	assert.Equal(t, "refs/heads/"+remoteBranch, branchConfig(t, local, localBranch, "merge"))
}

func TestSetBranchUpstream_PrefixedLocalBranch(t *testing.T) {
	local, remote := initTestRepoWithRemote(t, "docker-compose-examples")

	const (
		localBranch  = "mhamza15/docker-compose-examples"
		remoteName   = "mhamza15"
		remoteBranch = "docker-compose-examples"
	)

	cmd := exec.Command("git", "-C", local, "remote", "add", remoteName, remote)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git remote add: %s", output)

	cmd = exec.Command("git", "-C", local, "branch", localBranch, "origin/"+remoteBranch)

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git branch: %s", output)

	cmd = exec.Command("git", "-C", local, "config", "branch."+localBranch+".merge", "refs/heads/"+localBranch)

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git config merge: %s", output)

	require.NoError(t, SetBranchUpstream(local, localBranch, remoteName, remoteBranch))

	assert.Equal(t, remoteName, branchConfig(t, local, localBranch, "remote"))
	assert.Equal(t, "refs/heads/"+remoteBranch, branchConfig(t, local, localBranch, "merge"))
}

func TestConfigureWorktreePush_PrefixedLocalBranch(t *testing.T) {
	local, remote := initTestRepoWithRemote(t, "docker-compose-examples")

	const (
		localBranch  = "mhamza15/docker-compose-examples"
		remoteName   = "mhamza15"
		remoteBranch = "docker-compose-examples"
	)

	cmd := exec.Command("git", "-C", local, "remote", "add", remoteName, remote)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git remote add: %s", output)

	require.NoError(t, FetchBranch(local, remoteName, remoteBranch))
	require.NoError(t, CreateTrackingBranch(local, localBranch, remoteName, remoteBranch))

	worktreePath := filepath.Join(t.TempDir(), "worktree")
	require.NoError(t, Add(local, worktreePath, localBranch, "main"))

	cmd = exec.Command("git", "-C", worktreePath, "config", "user.email", "test@test.com")

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git config user.email: %s", output)

	cmd = exec.Command("git", "-C", worktreePath, "config", "user.name", "test")

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git config user.name: %s", output)

	cmd = exec.Command("git", "-C", worktreePath, "commit", "--allow-empty", "-m", "test")

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git commit: %s", output)

	require.NoError(t, ConfigureWorktreePush(local, worktreePath, localBranch))

	pushDefaultOrigin, pushDefault := worktreeConfigValue(t, worktreePath, "push.default")
	assert.Equal(t, "upstream", pushDefault)
	assert.Contains(t, pushDefaultOrigin, "config.worktree")

	cmd = exec.Command("git", "-C", worktreePath, "push", "--dry-run")

	output, err = cmd.CombinedOutput()
	require.NoError(t, err, "git push --dry-run: %s", output)

	assert.Contains(t, string(output), localBranch+" -> "+remoteBranch)
}

func TestConfigureWorktreePush_SameBranchName(t *testing.T) {
	local, _ := initTestRepoWithRemote(t, "feature")
	worktreePath := filepath.Join(t.TempDir(), "worktree")

	require.NoError(t, Add(local, worktreePath, "feature", "main"))
	require.NoError(t, ConfigureWorktreePush(local, worktreePath, "feature"))

	enabled, err := repoConfigEnabled(local, "extensions.worktreeConfig")
	require.NoError(t, err)

	assert.False(t, enabled)
}

func branchConfig(t *testing.T, repoPath, branch, key string) string {
	t.Helper()

	name := "branch." + branch + "." + key
	cmd := exec.Command("git", "-C", repoPath, "config", name)

	output, err := cmd.Output()
	require.NoError(t, err, "git config %s", name)

	return strings.TrimSpace(string(output))
}

func worktreeConfigValue(t *testing.T, worktreePath, key string) (string, string) {
	t.Helper()

	origin, value, ok := worktreeConfigLookup(t, worktreePath, key)
	require.True(t, ok, "missing worktree config %s", key)

	return origin, value
}

func worktreeConfigLookup(t *testing.T, worktreePath, key string) (string, string, bool) {
	t.Helper()

	cmd := exec.Command("git", "-C", worktreePath, "config", "--show-origin", "--worktree", "--get", key)

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", "", false
		}

		require.NoError(t, err, "git config --worktree --get %s: %s", key, output)
	}

	parts := strings.SplitN(strings.TrimSpace(string(output)), "\t", 2)
	require.Len(t, parts, 2, "unexpected git config output: %s", output)

	return parts[0], parts[1], true
}
