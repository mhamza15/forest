package git

import (
	"os"
	"os/exec"
	"path/filepath"
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
