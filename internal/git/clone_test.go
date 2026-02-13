package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	src := initTestRepo(t)

	dest := filepath.Join(t.TempDir(), "cloned")

	err := Clone(src, dest)
	require.NoError(t, err)

	// Verify the clone is a valid git repo.
	cmd := exec.Command("git", "-C", dest, "rev-parse", "--git-dir")
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "rev-parse failed: %s", out)
}

func TestClone_InvalidURL(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "cloned")

	err := Clone("/nonexistent/path/repo", dest)
	assert.Error(t, err)
}

func TestDefaultBranch(t *testing.T) {
	repo := initTestRepo(t)

	branch, err := DefaultBranch(repo)
	require.NoError(t, err)
	assert.Equal(t, "main", branch)
}

func TestDefaultBranch_CustomBranch(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")

	require.NoError(t, os.MkdirAll(repo, 0o755))

	commands := [][]string{
		{"git", "init", "--initial-branch=develop"},
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

	branch, err := DefaultBranch(repo)
	require.NoError(t, err)
	assert.Equal(t, "develop", branch)
}
