package config

import (
	"fmt"
	"os"

	"github.com/mhamza15/forest/internal/git"
)

// InferProject determines the current project by matching the git
// remote URLs of the working directory against registered projects.
// It returns the project name, or an error if the working directory
// is not inside a recognized repository.
func InferProject() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	return InferProjectFromDir(cwd)
}

// InferProjectFromDir determines the project for the given directory
// by matching its git remote URLs against registered projects.
func InferProjectFromDir(dir string) (string, error) {
	remotes, err := git.Remotes(dir)
	if err != nil {
		return "", fmt.Errorf("not in a git repository, use --project")
	}

	for _, remote := range remotes {
		url, err := git.RemoteURL(dir, remote)
		if err != nil {
			continue
		}

		nwo := git.NormalizeRemoteURL(url)

		name, _, err := FindProjectByRemote(nwo)
		if err == nil {
			return name, nil
		}
	}

	return "", fmt.Errorf("no registered project matches current repository, use --project")
}
