package forest

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mhamza15/forest/internal/config"
)

func TestAddTree_RemovesConfiguredFiles(t *testing.T) {
	repo := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(repo, ".env"), []byte("SECRET=abc"), 0o644))
	runGit(t, repo, "add", ".env")
	runGit(t, repo, "commit", "-m", "add env")

	worktreeRoot := t.TempDir()
	rc := config.ResolvedConfig{
		Name:        "demo",
		Repo:        repo,
		WorktreeDir: worktreeRoot,
		Branch:      "main",
		Remove:      []string{".env"},
	}

	result, err := AddTree(rc, "feature")
	require.NoError(t, err)

	assert.True(t, result.Created)
	assert.Empty(t, result.RemoveWarnings)

	_, err = os.Stat(filepath.Join(result.WorktreePath, ".env"))
	require.ErrorIs(t, err, fs.ErrNotExist)

	status := strings.TrimSpace(runGit(t, result.WorktreePath, "status", "--short"))
	assert.Empty(t, status)

	flags := strings.TrimSpace(runGit(t, result.WorktreePath, "ls-files", "-v", "--", ".env"))
	assert.Equal(t, "S .env", flags)
}

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

		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "command %v failed: %s", args, output)
	}

	return repo
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s: %s", strings.Join(args, " "), output)

	return string(output)
}
