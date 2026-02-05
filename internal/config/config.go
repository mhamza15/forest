package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// GlobalConfig holds the top-level forest configuration.
// It is stored at $XDG_CONFIG_HOME/forest/config.yaml.
type GlobalConfig struct {
	WorktreeDir string `yaml:"worktree_dir"`
	Branch      string `yaml:"branch"`
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

	content := []byte(`# Forest global configuration

# Default directory for storing worktrees. Worktrees are organized
# as: <worktree_dir>/<project>/<branch> (default: ~/.local/share/forest/worktrees/)
worktree_dir: ~/.local/share/forest/worktrees

# Default branch to base new worktrees on (default: main)
branch: main
`)

	if err := os.WriteFile(p, content, 0o644); err != nil {
		return fmt.Errorf("writing default global config: %w", err)
	}

	return nil
}
