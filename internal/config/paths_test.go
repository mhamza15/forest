package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir_XDGSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	assert.Equal(t, "/tmp/xdg-config/forest", ConfigDir())
}

func TestConfigDir_XDGUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, ".config", "forest"), ConfigDir())
}

func TestDataDir_XDGSet(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")

	assert.Equal(t, "/tmp/xdg-data/forest", DataDir())
}

func TestDataDir_XDGUnset(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	assert.Equal(t, filepath.Join(home, ".local", "share", "forest"), DataDir())
}

func TestGlobalConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	assert.Equal(t, "/tmp/xdg-config/forest/config.yaml", GlobalConfigPath())
}

func TestProjectConfigPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	assert.Equal(t, "/tmp/xdg-config/forest/projects/myapp.yaml", ProjectConfigPath("myapp"))
}

func TestProjectsDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg-config")

	assert.Equal(t, "/tmp/xdg-config/forest/projects", ProjectsDir())
}

func TestDefaultWorktreeDir(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/xdg-data")

	assert.Equal(t, "/tmp/xdg-data/forest/worktrees", DefaultWorktreeDir())
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "tilde prefix",
			path: "~/projects",
			want: filepath.Join(home, "projects"),
		},
		{
			name: "tilde alone",
			path: "~",
			want: home,
		},
		{
			name: "absolute path unchanged",
			path: "/usr/local/bin",
			want: "/usr/local/bin",
		},
		{
			name: "relative path unchanged",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "empty string unchanged",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExpandPath(tt.path))
		})
	}
}
