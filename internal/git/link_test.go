package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, ".env"), []byte("SECRET=abc"), 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(src, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "config", "local.yml"), []byte("db: local"), 0o644))

	warnings := SymlinkFiles(src, dst, []string{".env", "config/local.yml"})
	assert.Empty(t, warnings)

	// Verify symlink targets.
	target, err := os.Readlink(filepath.Join(dst, ".env"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(src, ".env"), target)

	target, err = os.Readlink(filepath.Join(dst, "config", "local.yml"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(src, "config", "local.yml"), target)

	// Verify content is accessible through the symlink.
	data, err := os.ReadFile(filepath.Join(dst, ".env"))
	require.NoError(t, err)
	assert.Equal(t, "SECRET=abc", string(data))
}

func TestSymlinkFiles_MissingSource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	warnings := SymlinkFiles(src, dst, []string{"nonexistent.txt"})

	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "not found, skipping")
}

func TestSymlinkFiles_OverwritesExisting(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "file.txt"), []byte("new"), 0o644))

	// Place a regular file at the destination first.
	require.NoError(t, os.WriteFile(filepath.Join(dst, "file.txt"), []byte("old"), 0o644))

	warnings := SymlinkFiles(src, dst, []string{"file.txt"})
	assert.Empty(t, warnings)

	// Destination should now be a symlink, not the old regular file.
	target, err := os.Readlink(filepath.Join(dst, "file.txt"))
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(src, "file.txt"), target)
}

func TestSymlinkFiles_OverwritesExistingSymlink(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	fileA := filepath.Join(src, "a.txt")
	fileB := filepath.Join(src, "b.txt")

	require.NoError(t, os.WriteFile(fileA, []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(fileB, []byte("b"), 0o644))

	// Create a symlink pointing to fileA first.
	dstLink := filepath.Join(dst, "a.txt")
	require.NoError(t, os.Symlink(fileB, dstLink))

	warnings := SymlinkFiles(src, dst, []string{"a.txt"})
	assert.Empty(t, warnings)

	// Symlink should now point to fileA, not fileB.
	target, err := os.Readlink(dstLink)
	require.NoError(t, err)
	assert.Equal(t, fileA, target)
}
