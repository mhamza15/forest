// Package project implements the "forest project" command group.
package project

import "github.com/spf13/cobra"

// Command returns the project parent command.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
		Long:  "Register, list, and manage forest projects.",
	}

	cmd.AddCommand(addCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(removeCmd())

	return cmd
}
