package git

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveFiles_TrackedFile(t *testing.T) {
	repo := initTestRepo(t)

	require.NoError(t, os.WriteFile(filepath.Join(repo, ".env"), []byte("SECRET=abc"), 0o644))
	runGit(t, repo, "add", ".env")
	runGit(t, repo, "commit", "-m", "add env")

	wtPath := filepath.Join(t.TempDir(), "feature")
	require.NoError(t, Add(repo, wtPath, "feature", "main"))

	warnings := RemoveFiles(wtPath, []string{".env"})
	assert.Empty(t, warnings)

	_, err := os.Stat(filepath.Join(wtPath, ".env"))
	require.ErrorIs(t, err, fs.ErrNotExist)

	status := strings.TrimSpace(runGit(t, wtPath, "status", "--short"))
	assert.Empty(t, status)

	flags := strings.TrimSpace(runGit(t, wtPath, "ls-files", "-v", "--", ".env"))
	assert.Equal(t, "S .env", flags)
}

func TestRemoveFiles_UntrackedFile(t *testing.T) {
	repo := initTestRepo(t)
	wtPath := filepath.Join(t.TempDir(), "feature")

	require.NoError(t, Add(repo, wtPath, "feature", "main"))
	require.NoError(t, os.WriteFile(filepath.Join(wtPath, ".env"), []byte("SECRET=abc"), 0o644))

	warnings := RemoveFiles(wtPath, []string{".env"})
	assert.Empty(t, warnings)

	_, err := os.Stat(filepath.Join(wtPath, ".env"))
	require.ErrorIs(t, err, fs.ErrNotExist)

	status := strings.TrimSpace(runGit(t, wtPath, "status", "--short"))
	assert.Empty(t, status)
}

func TestRemoveFiles_InvalidPath(t *testing.T) {
	warnings := RemoveFiles(t.TempDir(), []string{"../outside"})

	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "path must stay within the repo root")
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s: %s", strings.Join(args, " "), output)

	return string(output)
}
