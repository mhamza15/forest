// Package tmux provides operations for managing tmux sessions.
package tmux

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrNotRunning is returned when a tmux operation requires an active
// tmux server but none is found.
var ErrNotRunning = errors.New("not inside a tmux session (is $TMUX set?)")

// IsRunning returns true if the current process is inside a tmux session.
func IsRunning() bool {
	return os.Getenv("TMUX") != ""
}

// RequireRunning returns ErrNotRunning if tmux is not active.
func RequireRunning() error {
	if !IsRunning() {
		return ErrNotRunning
	}
	return nil
}

// SessionExists checks whether a tmux session with the given name exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// NewSession creates a new detached tmux session with the given name
// and working directory. It does not switch to the session.
func NewSession(name, workdir string) error {
	cmd := exec.Command(
		"tmux", "new-session",
		"-d",
		"-s", name,
		"-c", workdir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-session: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// SwitchTo switches the current tmux client to the named session.
func SwitchTo(name string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux switch-client: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// CurrentSession returns the name of the active tmux session, or an
// empty string if it cannot be determined.
func CurrentSession() string {
	cmd := exec.Command("tmux", "display-message", "-p", "#S")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(output))
}

// SwitchToLast switches the current tmux client to the previous
// session. This is a no-op if there is no previous session.
func SwitchToLast() {
	_ = exec.Command("tmux", "switch-client", "-l").Run()
}

// KillSession kills the tmux session with the given name. If the
// session being killed is the one the user is currently in, it
// switches to the last session first. It is not an error if the
// session does not exist.
func KillSession(name string) error {
	if !SessionExists(name) {
		return nil
	}

	if CurrentSession() == name {
		SwitchToLast()
	}

	cmd := exec.Command("tmux", "kill-session", "-t", name)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux kill-session: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// SessionName builds the conventional forest session name from a
// project name and branch.
func SessionName(project, branch string) string {
	// Tmux does not allow dots in session names, so replace them.
	name := project + "-" + branch
	return strings.ReplaceAll(name, ".", "_")
}
