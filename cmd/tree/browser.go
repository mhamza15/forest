package tree

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mhamza15/forest/internal/tui"
)

// runTreeBrowser launches the inline TUI for browsing projects and
// their worktrees. It runs without the alternate screen so it renders
// inline below the shell prompt. When project is non-empty, the browser
// is scoped to that single project.
func runTreeBrowser(project string) error {
	m, err := tui.NewModel(project)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return err
	}
	// Execute any deferred action (e.g. switching to a tmux session)
	// after the TUI has finished and restored the terminal.
	if final, ok := result.(tui.Model); ok {
		if action := final.Action(); action != nil {
			return action()
		}
	}
	return nil
}
