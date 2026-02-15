package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initTestRepo creates a git repo with a remote URL so that
// InferProjectFromDir can match it against registered projects.
func initTestRepo(t *testing.T, remoteURL string) string {
	t.Helper()

	dir := t.TempDir()
	repo := filepath.Join(dir, "repo")

	require.NoError(t, os.MkdirAll(repo, 0o755))

	commands := [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "test"},
		{"git", "remote", "add", "origin", remoteURL},
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

// registerProject writes a minimal project config file so that
// InferProjectFromDir can discover it.
func registerProject(t *testing.T, name, repoPath string) {
	t.Helper()

	dir := ProjectsDir()

	require.NoError(t, os.MkdirAll(dir, 0o755))

	content := "repo: " + repoPath + "\n"

	require.NoError(t, os.WriteFile(
		filepath.Join(dir, name+".yaml"),
		[]byte(content),
		0o644,
	))
}

func TestInferProjectFromDir(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	repo := initTestRepo(t, "https://github.com/acme/widgets.git")
	registerProject(t, "widgets", repo)

	name, err := InferProjectFromDir(repo)
	require.NoError(t, err)
	assert.Equal(t, "widgets", name)
}

func TestInferProjectFromDir_Subdirectory(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	repo := initTestRepo(t, "git@github.com:acme/widgets.git")
	registerProject(t, "widgets", repo)

	sub := filepath.Join(repo, "src", "pkg")
	require.NoError(t, os.MkdirAll(sub, 0o755))

	name, err := InferProjectFromDir(sub)
	require.NoError(t, err)
	assert.Equal(t, "widgets", name)
}

func TestInferProjectFromDir_NoMatch(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)

	repo := initTestRepo(t, "https://github.com/acme/unknown.git")

	_, err := InferProjectFromDir(repo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--project")
}

func TestInferProjectFromDir_NotGitRepo(t *testing.T) {
	dir := t.TempDir()

	_, err := InferProjectFromDir(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--project")
}
