package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/mhamza15/forest/internal/git"
)

// ProjectConfig holds per-project overrides. Empty fields fall through
// to the global config during resolution.
type ProjectConfig struct {
	// Repo is the absolute path to the git repository.
	Repo string `yaml:"repo"`

	// WorktreeDir overrides the global worktree directory for this project.
	WorktreeDir string `yaml:"worktree_dir,omitempty"`

	// Branch overrides the global base branch for this project.
	Branch string `yaml:"branch,omitempty"`

	// Copy lists files relative to the repo root to copy into each new worktree.
	Copy []string `yaml:"copy,omitempty"`

	// Symlink lists files relative to the repo root to symlink into each new worktree.
	// Unlike copy, symlinked files reference the original in the repo root directly.
	Symlink []string `yaml:"symlink,omitempty"`

	// Layout overrides the global tmux window layout for this project.
	Layout []Window `yaml:"layout,omitempty"`
}

// ResolvedConfig is the final configuration for a project after merging
// project-level overrides onto the global defaults.
type ResolvedConfig struct {
	// Name is the project name used in config paths and session names.
	Name string

	// Repo is the absolute path to the git repository.
	Repo string

	// WorktreeDir is the resolved base directory for worktrees.
	WorktreeDir string

	// Branch is the resolved base branch for new worktrees.
	Branch string

	// Copy lists files to copy from the repo root into each new worktree.
	Copy []string

	// Symlink lists files to symlink from the repo root into each new worktree.
	Symlink []string

	// Layout defines the tmux windows to create for each new session.
	Layout []Window
}

// LoadProject reads a project config file by name.
func LoadProject(name string) (ProjectConfig, error) {
	var cfg ProjectConfig

	data, err := os.ReadFile(ProjectConfigPath(name))
	if err != nil {
		return cfg, fmt.Errorf("reading project config %q: %w", name, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing project config %q: %w", name, err)
	}

	return cfg, nil
}

// SaveProject writes a project config to disk, creating parent directories
// as needed.
func SaveProject(name string, cfg ProjectConfig) error {
	p := ProjectConfigPath(name)

	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("creating projects directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling project config: %w", err)
	}

	content := ProjectSchemaModeline() + "\n" + string(data)

	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing project config %q: %w", name, err)
	}

	return nil
}

// RemoveProject deletes the project config file. It is not an error
// if the file does not exist.
func RemoveProject(name string) error {
	p := ProjectConfigPath(name)

	if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing project config %q: %w", name, err)
	}

	return nil
}

// Resolve loads the global and project configs, then merges them.
// Project-level fields take precedence when non-empty.
func Resolve(name string) (ResolvedConfig, error) {
	global, err := LoadGlobal()
	if err != nil {
		return ResolvedConfig{}, err
	}

	proj, err := LoadProject(name)
	if err != nil {
		return ResolvedConfig{}, err
	}

	layout := global.Layout
	if len(proj.Layout) > 0 {
		layout = proj.Layout
	}

	rc := ResolvedConfig{
		Name:        name,
		Repo:        proj.Repo,
		WorktreeDir: global.WorktreeDir,
		Branch:      global.Branch,
		Copy:        proj.Copy,
		Symlink:     proj.Symlink,
		Layout:      layout,
	}

	if proj.WorktreeDir != "" {
		rc.WorktreeDir = ExpandPath(proj.WorktreeDir)
	}

	if proj.Branch != "" {
		rc.Branch = proj.Branch
	}

	return rc, nil
}

// FindProjectByRemote finds a registered project that has any remote
// matching the given "owner/repo" string.
func FindProjectByRemote(nwo string) (string, ResolvedConfig, error) {
	names, err := ListProjects()
	if err != nil {
		return "", ResolvedConfig{}, err
	}

	for _, name := range names {
		rc, err := Resolve(name)
		if err != nil {
			continue
		}

		remotes, err := git.Remotes(rc.Repo)
		if err != nil {
			continue
		}

		for _, remote := range remotes {
			raw, err := git.RemoteURL(rc.Repo, remote)
			if err != nil {
				continue
			}

			if git.NormalizeRemoteURL(raw) == nwo {
				return name, rc, nil
			}
		}
	}

	return "", ResolvedConfig{}, fmt.Errorf("no project found with remote matching %q", nwo)
}

// ListProjects returns the names of all registered projects by scanning
// the projects directory for .yaml files.
func ListProjects() ([]string, error) {
	dir := ProjectsDir()

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading projects directory: %w", err)
	}

	var names []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		ext := filepath.Ext(name)

		if ext == ".yaml" || ext == ".yml" {
			names = append(names, name[:len(name)-len(ext)])
		}
	}

	return names, nil
}
