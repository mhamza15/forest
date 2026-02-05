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

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))
)
