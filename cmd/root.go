// Package cmd implements the forest CLI command tree.
package cmd

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	configcmd "github.com/mhamza15/forest/cmd/config"
)

var verbose bool

// rootCmd is the top-level forest command.
var rootCmd = &cobra.Command{
	Use:   "forest",
	Short: "Manage git worktrees with tmux sessions",
	Long:  "Forest organizes git worktrees into per-project directories and pairs each with a tmux session.",

	SilenceUsage:  true,
	SilenceErrors: true,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogging()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")

	rootCmd.AddCommand(configcmd.Command())
}

// Execute runs the root command. It is the single entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}

func initLogging() {
	level := slog.LevelInfo
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
