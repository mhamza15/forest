package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleProject = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	styleTree = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	styleCursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))
)
