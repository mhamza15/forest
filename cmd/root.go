// Package cmd implements the forest CLI command tree.
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	configcmd "github.com/mhamza15/forest/cmd/config"
	projectcmd "github.com/mhamza15/forest/cmd/project"
	treecmd "github.com/mhamza15/forest/cmd/tree"
)

var verbose bool

// rootCmd is the top-level forest command.
var rootCmd = &cobra.Command{
	Use:   "forest",
	Short: "Manage git worktrees with tmux sessions",
	Long: `Forest is a CLI tool for managing git worktrees and tmux sessions.

It simplifies the workflow of creating and managing multiple working
copies of a repository, each in its own directory with a dedicated
tmux session.`,

	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")

	rootCmd.AddCommand(configcmd.Command())
	rootCmd.AddCommand(projectcmd.Command())
	rootCmd.AddCommand(treecmd.Command())
}

// Execute runs the root command. It is the single entry point called from main.
func Execute() {
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

	w := os.Stderr

	handler := tint.NewHandler(w, &tint.Options{
		Level:      level,
		TimeFormat: time.Kitchen,
		NoColor:    !isatty.IsTerminal(w.Fd()),
	})

	slog.SetDefault(slog.New(handler))
}
