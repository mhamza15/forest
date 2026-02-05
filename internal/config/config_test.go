package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadGlobal_MissingFile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, err := LoadGlobal()
	require.NoError(t, err)

	assert.Equal(t, defaultBranch, cfg.Branch)
	assert.NotEmpty(t, cfg.WorktreeDir)
}

func TestLoadGlobal_CustomValues(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "forest")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	content := []byte("worktree_dir: /custom/trees\nbranch: develop\n")
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0o644))

	cfg, err := LoadGlobal()
	require.NoError(t, err)

	assert.Equal(t, "/custom/trees", cfg.WorktreeDir)
	assert.Equal(t, "develop", cfg.Branch)
}

func TestLoadGlobal_EmptyFieldsGetDefaults(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))

	configDir := filepath.Join(dir, "forest")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	// Empty YAML file: all fields should get defaults.
	content := []byte("{}\n")
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0o644))

	cfg, err := LoadGlobal()
	require.NoError(t, err)

	assert.Equal(t, defaultBranch, cfg.Branch)
	assert.Contains(t, cfg.WorktreeDir, "worktrees")
}

func TestLoadGlobal_TildeExpansion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "forest")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	content := []byte("worktree_dir: ~/my-trees\nbranch: main\n")
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0o644))

	cfg, err := LoadGlobal()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, "my-trees"), cfg.WorktreeDir)
}

func TestWriteDefaultGlobal_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	require.NoError(t, WriteDefaultGlobal())

	_, err := os.Stat(GlobalConfigPath())
	assert.NoError(t, err)
}

func TestWriteDefaultGlobal_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	configDir := filepath.Join(dir, "forest")
	require.NoError(t, os.MkdirAll(configDir, 0o755))

	existing := []byte("branch: custom\n")
	p := filepath.Join(configDir, "config.yaml")
	require.NoError(t, os.WriteFile(p, existing, 0o644))

	require.NoError(t, WriteDefaultGlobal())

	data, err := os.ReadFile(p)
	require.NoError(t, err)

	assert.Equal(t, existing, data)
}
