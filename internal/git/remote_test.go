package git

import (
	"os/exec"
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

func branchConfig(t *testing.T, repoPath, branch, key string) string {
	t.Helper()

	name := "branch." + branch + "." + key
	cmd := exec.Command("git", "-C", repoPath, "config", name)

	output, err := cmd.Output()
	require.NoError(t, err, "git config %s", name)

	return strings.TrimSpace(string(output))
}
