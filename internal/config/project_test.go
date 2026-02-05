package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadProject(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	cfg := ProjectConfig{
		Repo:   "/home/user/repos/myapp",
		Branch: "develop",
	}

	require.NoError(t, SaveProject("myapp", cfg))

	loaded, err := LoadProject("myapp")
	require.NoError(t, err)

	assert.Equal(t, cfg.Repo, loaded.Repo)
	assert.Equal(t, cfg.Branch, loaded.Branch)
	assert.Empty(t, loaded.WorktreeDir)
}

func TestLoadProject_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := LoadProject("nonexistent")
	assert.Error(t, err)
}

func TestResolve_GlobalDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))

	cfg := ProjectConfig{
		Repo: "/home/user/repos/myapp",
	}
	require.NoError(t, SaveProject("myapp", cfg))

	rc, err := Resolve("myapp")
	require.NoError(t, err)

	assert.Equal(t, "myapp", rc.Name)
	assert.Equal(t, "/home/user/repos/myapp", rc.Repo)
	assert.Equal(t, defaultBranch, rc.Branch)
	assert.Contains(t, rc.WorktreeDir, "worktrees")
}

func TestResolve_ProjectOverrides(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))

	cfg := ProjectConfig{
		Repo:        "/home/user/repos/myapp",
		WorktreeDir: "/custom/trees",
		Branch:      "develop",
	}
	require.NoError(t, SaveProject("myapp", cfg))

	rc, err := Resolve("myapp")
	require.NoError(t, err)

	assert.Equal(t, "/custom/trees", rc.WorktreeDir)
	assert.Equal(t, "develop", rc.Branch)
}

func TestListProjects_Empty(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	names, err := ListProjects()
	require.NoError(t, err)

	assert.Empty(t, names)
}

func TestListProjects_MultipleProjects(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		require.NoError(t, SaveProject(name, ProjectConfig{
			Repo: "/repos/" + name,
		}))
	}

	names, err := ListProjects()
	require.NoError(t, err)

	assert.Len(t, names, 3)
	assert.Contains(t, names, "alpha")
	assert.Contains(t, names, "beta")
	assert.Contains(t, names, "gamma")
}
