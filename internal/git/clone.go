package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Clone clones a git repository from url into dest.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone: %s: %w", bytes.TrimSpace(output), err)
	}

	return nil
}

// DefaultBranch returns the name of the current branch (HEAD) in the
// repository, which for a fresh clone is the remote's default branch.
func DefaultBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "symbolic-ref", "--short", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("detecting default branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
