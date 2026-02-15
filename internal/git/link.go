package git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// SymlinkFiles creates a symlink in worktreePath for each file listed in
// files. Each symlink points to the corresponding absolute path in
// repoPath. Paths in files are relative to the repo root. Parent
// directories in the worktree are created as needed. Files that do not
// exist in the source are skipped with a warning.
func SymlinkFiles(repoPath, worktreePath string, files []string) []string {
	var warnings []string

	for _, f := range files {
		src := filepath.Join(repoPath, f)
		dst := filepath.Join(worktreePath, f)

		if err := symlinkFile(src, dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				warnings = append(warnings, fmt.Sprintf("symlink: %s not found, skipping", f))
				continue
			}

			warnings = append(warnings, fmt.Sprintf("symlink: %s: %s", f, err))
		}
	}

	return warnings
}

// symlinkFile creates a symlink at dst pointing to the absolute path src.
// If a file or symlink already exists at dst, it is removed first.
// Parent directories are created as needed.
func symlinkFile(src, dst string) error {
	if _, err := os.Lstat(src); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	// Remove any existing file or symlink at the destination so
	// os.Symlink does not fail with "file exists."
	if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removing existing file: %w", err)
	}

	return os.Symlink(src, dst)
}
