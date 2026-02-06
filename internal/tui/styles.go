package tui

import "github.com/charmbracelet/lipgloss"

var (
	styleProject = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#89B4FA"))

	styleTree = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E3A1"))

	styleCursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F9E2AF")).
			Bold(true)

	styleHeader = lipgloss.NewStyle().
			Bold(true)

	styleDim = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6C7086"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))
)
