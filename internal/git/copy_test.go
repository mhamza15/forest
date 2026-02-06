package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFiles(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, ".env"), []byte("SECRET=abc"), 0o644))

	require.NoError(t, os.MkdirAll(filepath.Join(src, "config"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "config", "local.yml"), []byte("db: local"), 0o644))

	warnings := CopyFiles(src, dst, []string{".env", "config/local.yml"})
	assert.Empty(t, warnings)

	data, err := os.ReadFile(filepath.Join(dst, ".env"))
	require.NoError(t, err)
	assert.Equal(t, "SECRET=abc", string(data))

	data, err = os.ReadFile(filepath.Join(dst, "config", "local.yml"))
	require.NoError(t, err)
	assert.Equal(t, "db: local", string(data))
}

func TestCopyFiles_MissingSource(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	warnings := CopyFiles(src, dst, []string{"nonexistent.txt"})

	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "not found, skipping")
}

func TestCopyFiles_PreservesPermissions(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(src, "script.sh"), []byte("#!/bin/sh"), 0o755))

	warnings := CopyFiles(src, dst, []string{"script.sh"})
	assert.Empty(t, warnings)

	info, err := os.Stat(filepath.Join(dst, "script.sh"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}
