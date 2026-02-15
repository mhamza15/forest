package config

import (
	"os/exec"
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

func TestRemoveProject(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	require.NoError(t, SaveProject("doomed", ProjectConfig{
		Repo: "/repos/doomed",
	}))

	_, err := LoadProject("doomed")
	require.NoError(t, err)

	require.NoError(t, RemoveProject("doomed"))

	_, err = LoadProject("doomed")
	assert.Error(t, err)
}

func TestRemoveProject_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	// Removing a nonexistent project is not an error.
	assert.NoError(t, RemoveProject("ghost"))
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

// addRemote adds a named remote to an existing test repo.
func addRemote(t *testing.T, repoPath, name, url string) {
	t.Helper()

	cmd := exec.Command("git", "-C", repoPath, "remote", "add", name, url)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git remote add %s: %s", name, out)
}

func TestFindProjectByRemote_PrefersOrigin(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(configDir, "data"))
	const sharedNWO = "https://github.com/org/myproject.git"

	// myproject-internal has org/myproject only as an upstream remote.
	// Its origin points to a different repo.
	internalRepo := initTestRepo(t, "https://github.com/corp/myproject-internal.git")
	addRemote(t, internalRepo, "upstream", sharedNWO)
	registerProject(t, "myproject-internal", internalRepo)

	// myproject has org/myproject as its origin.
	primaryRepo := initTestRepo(t, sharedNWO)
	registerProject(t, "myproject", primaryRepo)

	// FindProjectByRemote must choose myproject (origin match) over
	// myproject-internal (upstream match), regardless of directory order.
	name, rc, err := FindProjectByRemote("org/myproject")
	require.NoError(t, err)
	assert.Equal(t, "myproject", name)
	assert.Equal(t, primaryRepo, rc.Repo)
}

func TestFindProjectByRemote_PrefersForkOverUpstream(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(configDir, "data"))
	// myproject-internal: origin is corp/myproject-internal, upstream is
	// org/myproject. This repo is unrelated to org/myproject.
	internalRepo := initTestRepo(t, "git@github.com:corp/myproject-internal.git")
	addRemote(t, internalRepo, "upstream", "git@github.com:org/myproject.git")
	registerProject(t, "myproject-internal", internalRepo)

	// myproject: origin is user/myproject (a personal fork), upstream is
	// org/myproject. The origin repo name "myproject" matches the target.
	forkRepo := initTestRepo(t, "git@github.com:user/myproject.git")
	addRemote(t, forkRepo, "upstream", "git@github.com:org/myproject.git")
	registerProject(t, "myproject", forkRepo)

	// Both projects have org/myproject as upstream but neither has it
	// as origin. The fork (user/myproject) shares the repo name with
	// the target, so it must win.
	name, rc, err := FindProjectByRemote("org/myproject")
	require.NoError(t, err)
	assert.Equal(t, "myproject", name)
	assert.Equal(t, forkRepo, rc.Repo)
}

func TestFindProjectByRemote_FallsBackToNonOrigin(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(configDir, "data"))

	// Only one project, and the matching NWO is on upstream, not origin.
	repo := initTestRepo(t, "https://github.com/corp/fork.git")
	addRemote(t, repo, "upstream", "https://github.com/acme/widgets.git")
	registerProject(t, "fork", repo)

	name, _, err := FindProjectByRemote("acme/widgets")
	require.NoError(t, err)

	assert.Equal(t, "fork", name)
}

func TestFindProjectByRemote_NoMatch(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(configDir, "data"))

	repo := initTestRepo(t, "https://github.com/acme/widgets.git")
	registerProject(t, "widgets", repo)

	_, _, err := FindProjectByRemote("acme/unknown")
	assert.Error(t, err)
}
