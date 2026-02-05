// Package tui implements the inline tree browser for forest.
package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines all keybindings for the tree browser.
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Open    key.Binding
	Delete  key.Binding
	New     key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Quit    key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k", "ctrl+p"),
			key.WithHelp("k/up", "up"),
		),

		Down: key.NewBinding(
			key.WithKeys("down", "j", "ctrl+n"),
			key.WithHelp("j/down", "down"),
		),

		Toggle: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "expand/collapse"),
		),

		Open: key.NewBinding(
			key.WithKeys("enter", "l"),
			key.WithHelp("enter", "open"),
		),

		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),

		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new tree"),
		),

		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yes"),
		),

		Cancel: key.NewBinding(
			key.WithKeys("N", "esc"),
			key.WithHelp("N/esc", "cancel"),
		),

		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
