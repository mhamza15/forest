// Package config handles loading, saving, and resolving forest configuration
// from XDG-compliant paths.
package config

import (
	"os"
	"path/filepath"
	"strings"
)

const appName = "forest"

// ConfigDir returns the directory where forest stores its configuration.
// It respects $XDG_CONFIG_HOME, falling back to ~/.config/forest.
func ConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")

	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".config", appName)
		}
		base = filepath.Join(home, ".config")
	}

	return filepath.Join(base, appName)
}

// DataDir returns the directory where forest stores its data.
// It respects $XDG_DATA_HOME, falling back to ~/.local/share/forest.
func DataDir() string {
	base := os.Getenv("XDG_DATA_HOME")

	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".local", "share", appName)
		}
		base = filepath.Join(home, ".local", "share")
	}

	return filepath.Join(base, appName)
}

// GlobalConfigPath returns the path to the global config.yaml file.
func GlobalConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

// ProjectsDir returns the directory containing per-project config files.
func ProjectsDir() string {
	return filepath.Join(ConfigDir(), "projects")
}

// ProjectConfigPath returns the path to a project's config file.
func ProjectConfigPath(name string) string {
	return filepath.Join(ProjectsDir(), name+".yaml")
}

// DefaultWorktreeDir returns the default base directory for worktrees.
func DefaultWorktreeDir() string {
	return filepath.Join(DataDir(), "worktrees")
}

// ExpandPath replaces a leading ~ with the user's home directory.
// If the path does not start with ~, it is returned unchanged.
func ExpandPath(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}

	return filepath.Join(home, p[1:])
}
