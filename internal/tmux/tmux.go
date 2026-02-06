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

// NewWindow creates a new window in the named session with its working
// directory set to workdir. If command is non-empty, it is sent to the
// window as keystrokes followed by Enter.
func NewWindow(session, workdir, command string) error {
	cmd := exec.Command(
		"tmux", "new-window",
		"-t", session,
		"-c", workdir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux new-window: %s: %w", strings.TrimSpace(string(output)), err)
	}

	if command != "" {
		return SendKeys(session, command)
	}

	return nil
}

// SendKeys sends a command string to the current window of the named
// session, followed by Enter.
func SendKeys(session, command string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", session, command, "Enter")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux send-keys: %s: %w", strings.TrimSpace(string(output)), err)
	}

	return nil
}

// ApplyLayout creates tmux windows for the given layout in the named
// session. The first command is sent to the session's initial window;
// subsequent commands each create a new window. An empty string opens
// a plain shell. workdir is the working directory for all windows.
func ApplyLayout(session, workdir string, commands []string) error {
	for i, cmd := range commands {
		if i == 0 {
			if cmd != "" {
				if err := SendKeys(session, cmd); err != nil {
					return err
				}
			}
			continue
		}

		if err := NewWindow(session, workdir, cmd); err != nil {
			return err
		}
	}

	// Select the first window so the user lands there on attach.
	if len(commands) > 1 {
		_ = exec.Command("tmux", "select-window", "-t", session+":0").Run()
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
