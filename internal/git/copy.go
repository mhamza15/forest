package git

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyFiles copies each file from repoPath to worktreePath, preserving
// relative directory structure. Paths in files are relative to the repo
// root. Files that do not exist in the source are skipped with a
// warning returned in the warnings slice.
func CopyFiles(repoPath, worktreePath string, files []string) []string {
	var warnings []string

	for _, f := range files {
		src := filepath.Join(repoPath, f)
		dst := filepath.Join(worktreePath, f)

		if err := copyFile(src, dst); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				warnings = append(warnings, fmt.Sprintf("copy: %s not found, skipping", f))
				continue
			}

			warnings = append(warnings, fmt.Sprintf("copy: %s: %s", f, err))
		}
	}

	return warnings
}

// copyFile copies a single file from src to dst, creating parent
// directories as needed. File permissions are preserved.
func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent directory: %w", err)
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)

	// Close explicitly to flush writes and catch errors.
	if err := out.Close(); err != nil {
		return err
	}

	return copyErr
}
