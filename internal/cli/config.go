package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/freeoss-space/lime/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Open the lime configuration file in $EDITOR",
		Long: `Open the lime configuration file in your preferred editor.

If the config file does not exist, a default one is created first.

The editor is determined by the EDITOR environment variable (fallback: vi).

Example config (~/.config/lime/config.toml):
  preferred_managers = ["brew", "apt", "dnf", "pacman", "apk"]
  auto_confirm = false
  rate_limit_per_second = 1
  http_timeout_seconds = 10`,
		RunE: runConfig,
	}
	cmd.AddCommand(newConfigPathCmd())
	return cmd
}

func runConfig(cmd *cobra.Command, _ []string) error {
	path, err := config.EnsureExists()
	if err != nil {
		return fmt.Errorf("ensuring config exists: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Opening %s in %s\n", path, editor)

	c := exec.Command(editor, path) //nolint:gosec // editor from env, intentional
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}
	return nil
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the path to the config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}
