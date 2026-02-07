package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Window describes a tmux window to create as part of a session layout.
type Window struct {
	// Name is the tmux window title. If empty, tmux uses its default.
	Name string `yaml:"name,omitempty"`

	// Command is the shell command to run in this window.
	// An empty string opens a plain shell.
	Command string `yaml:"command"`
}

// GlobalConfig holds the top-level forest configuration.
// It is stored at $XDG_CONFIG_HOME/forest/config.yaml.
type GlobalConfig struct {
	// WorktreeDir is the base directory for storing worktrees.
	// When omitted, defaults to $XDG_DATA_HOME/forest/worktrees,
	// falling back to ~/.local/share/forest/worktrees.
	WorktreeDir string `yaml:"worktree_dir,omitempty"`

	// Branch is the default base branch for new worktrees.
	Branch string `yaml:"branch"`

	// ProjectsDir is the starting directory for the project add file
	// picker. When empty, the current working directory is used.
	ProjectsDir string `yaml:"projects_dir,omitempty"`

	// Layout defines the tmux windows to create for each new session.
	Layout []Window `yaml:"layout,omitempty"`
}

const (
	defaultBranch = "main"
)

// LoadGlobal reads the global config file and returns it with defaults
// applied for any unset fields. If the file does not exist, the defaults
// are returned without error.
func LoadGlobal() (GlobalConfig, error) {
	cfg := GlobalConfig{
		WorktreeDir: DefaultWorktreeDir(),
		Branch:      defaultBranch,
	}

	data, err := os.ReadFile(GlobalConfigPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading global config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing global config: %w", err)
	}

	// Apply defaults for any fields left empty after parsing.
	if cfg.WorktreeDir == "" {
		cfg.WorktreeDir = DefaultWorktreeDir()
	}

	if cfg.Branch == "" {
		cfg.Branch = defaultBranch
	}

	cfg.WorktreeDir = ExpandPath(cfg.WorktreeDir)

	if cfg.ProjectsDir != "" {
		cfg.ProjectsDir = ExpandPath(cfg.ProjectsDir)
	}

	return cfg, nil
}

// WriteDefaultGlobal writes a default global config file if one does not
// already exist. It creates parent directories as needed.
func WriteDefaultGlobal() error {
	p := GlobalConfigPath()

	if _, err := os.Stat(p); err == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	content := ConfigSchemaModeline() + "\n" + `
# Forest global configuration

# Default directory for storing worktrees. Worktrees are organized
# as: <worktree_dir>/<project>/<branch>.
# Defaults to $XDG_DATA_HOME/forest/worktrees, or ~/.local/share/forest/worktrees.
# worktree_dir:

# Default branch to base new worktrees on (default: main)
branch: main

# Default directory that your projects live in. Used as starting
# point in ` + "`" + `project add` + "`" + ` directory picker. Example: ~/dev.
# projects_dir:
`

	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing default global config: %w", err)
	}

	return nil
}
