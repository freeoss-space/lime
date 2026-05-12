// Package cli implements the jil command-line interface using Cobra.
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/freeoss-space/lime/internal/config"
	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/pkg/httpclient"
)

// globalFlags holds flag values shared across subcommands.
type globalFlags struct {
	verbose bool
	jsonOut bool
	dryRun  bool
}

var global globalFlags

// NewRootCmd builds and returns the root cobra.Command.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "jil",
		Short: "A cross-platform package installer and search tool",
		Long: `jil — Just Install it!

A package manager abstraction layer that searches and installs software
using the best available package manager on your system.

Uses the Repology API (https://repology.org) to resolve packages across
package managers and repositories.`,
		Version:          fmt.Sprintf("%s (commit %s, built %s)", version, commit, date),
		SilenceUsage:     true,
		SilenceErrors:    true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) { applyNoColor() },
	}

	root.PersistentFlags().BoolVarP(&global.verbose, "verbose", "v", false, "verbose output")
	root.PersistentFlags().BoolVar(&global.jsonOut, "json", false, "output as JSON")
	root.PersistentFlags().BoolVar(&global.dryRun, "dry-run", false, "print commands without executing")

	root.AddCommand(
		newInstallCmd(),
		newSearchCmd(),
		newConfigCmd(),
	)

	return root
}

// Execute is the main entry point called from main().
func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", color.RedString("error:"), err)
		os.Exit(1)
	}
}

// applyNoColor disables colour output when stdout is not a TTY.
func applyNoColor() {
	if os.Getenv("NO_COLOR") != "" {
		color.NoColor = true
	}
}

// buildDependencies wires up the shared infrastructure from config.
func buildDependencies(cfg *config.Config) (*managers.Registry, *repology.Client) {
	ua := fmt.Sprintf("jil/%s (https://github.com/freeoss-space/lime)", version)
	hc := httpclient.New(httpclient.Options{
		UserAgent:     ua,
		RatePerSecond: cfg.RateLimitPerSecond,
		Timeout:       time.Duration(cfg.HTTPTimeoutSeconds) * time.Second,
	})
	reg := managers.DefaultRegistry()
	rep := repology.NewDefault(hc)
	return reg, rep
}
