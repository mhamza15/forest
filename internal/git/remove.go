package git

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

// RemoveFiles removes each configured file from the worktree. Paths in files
// are relative to the repo root. If a file is tracked, it is first marked with
// Git's skip-worktree bit so the deletion stays local to this worktree.
func RemoveFiles(worktreePath string, files []string) []string {
	var warnings []string

	for _, rawPath := range files {
		path, err := cleanWorktreeRelativePath(rawPath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("remove: %s: %s", rawPath, err))
			continue
		}

		if err := removeFile(worktreePath, path); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}

			warnings = append(warnings, fmt.Sprintf("remove: %s: %s", path, err))
		}
	}

	return warnings
}

func removeFile(worktreePath, path string) error {
	tracked, err := isTrackedPath(worktreePath, path)
	if err != nil {
		return err
	}

	if tracked {
		if err := skipWorktreePath(worktreePath, path); err != nil {
			return err
		}
	}

	dst := filepath.Join(worktreePath, path)

	if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return nil
}

func cleanWorktreeRelativePath(path string) (string, error) {
	clean := filepath.Clean(path)

	if clean == "." {
		return "", fmt.Errorf("path must name a file relative to the repo root")
	}

	if !filepath.IsLocal(clean) {
		return "", fmt.Errorf("path must stay within the repo root")
	}

	return clean, nil
}

func isTrackedPath(worktreePath, path string) (bool, error) {
	cmd := exec.Command("git", "-C", worktreePath, "ls-files", "--error-unmatch", "--", path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return false, nil
		}

		return false, fmt.Errorf("git ls-files %s: %s: %w", path, bytes.TrimSpace(output), err)
	}

	return true, nil
}

func skipWorktreePath(worktreePath, path string) error {
	cmd := exec.Command("git", "-C", worktreePath, "update-index", "--skip-worktree", "--", path)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git update-index --skip-worktree %s: %s: %w", path, bytes.TrimSpace(output), err)
	}

	return nil
}
