// Package cmd implements the forest CLI command tree.
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"

	configcmd "github.com/mhamza15/forest/cmd/config"
	projectcmd "github.com/mhamza15/forest/cmd/project"
	sessioncmd "github.com/mhamza15/forest/cmd/session"
	treecmd "github.com/mhamza15/forest/cmd/tree"
	"github.com/mhamza15/forest/internal/completion"
	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags. When empty, the version
// is read from the Go module build info (populated by go install).
var version = ""

var verbose bool

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "forest",
		Short: "Manage git worktrees with tmux sessions",
		Long: `Forest is a CLI tool for managing git worktrees and tmux sessions.

It simplifies the workflow of creating and managing multiple working
copies of a repository, each in its own directory with a dedicated
tmux session.`,

		SilenceErrors: true,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Silence usage after the command has been resolved. This way
			// cobra still shows usage for unknown subcommands and arg
			// validation errors, but not for runtime errors.
			cmd.SilenceUsage = true
			initLogging()
		},
	}

	rootCmd.Version = resolveVersion()

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable debug logging")
	rootCmd.PersistentFlags().StringP("project", "p", "", "project name (inferred from working directory when omitted)")

	if err := rootCmd.RegisterFlagCompletionFunc("project", completion.Projects); err != nil {
		fmt.Fprintf(os.Stderr, "warning: registering --project completion: %s\n", err)
	}

	rootCmd.AddCommand(configcmd.Command())
	rootCmd.AddCommand(projectcmd.Command())
	rootCmd.AddCommand(sessioncmd.Command())
	rootCmd.AddCommand(treecmd.Command())

	return rootCmd
}

func resolveVersion() string {
	if version != "" {
		return version
	}

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}

	return "dev"
}

// Execute runs the root command. It is the single entry point called from main.
func Execute() {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func initLogging() {
	// Suppress all slog output by default. The --verbose flag drops
	// the level to Debug so that slog.Debug calls become visible.
	level := slog.LevelError + 1
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
}
