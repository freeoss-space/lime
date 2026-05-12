package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/freeoss-space/lime/internal/config"
	"github.com/freeoss-space/lime/internal/install"
	"github.com/freeoss-space/lime/internal/managers"
)

func newInstallCmd() *cobra.Command {
	var (
		yes        bool
		managerArg string
	)

	cmd := &cobra.Command{
		Use:   "install <package> [package...]",
		Short: "Install one or more packages",
		Long: `Install one or more packages using the best available package manager.

jil queries Repology to resolve the canonical package name for each manager,
then installs using your preferred manager. Falls back to other available
managers if the preferred one is not installed on the system.

Examples:
  jil install ripgrep
  jil install ripgrep fd bat
  jil install -y ripgrep
  jil install --manager brew ripgrep`,
		Args:    cobra.MinimumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return runInstall(cmd, args, yes, managerArg) },
		Example: "  jil install ripgrep\n  jil install -y ripgrep fd bat",
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "auto-confirm installation without prompt")
	cmd.Flags().StringVar(&managerArg, "manager", "", "force a specific package manager")

	return cmd
}

func runInstall(cmd *cobra.Command, pkgs []string, yes bool, managerArg string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.AutoConfirm {
		yes = true
	}

	reg, rep := buildDependencies(cfg)

	specs := make([]managers.PackageSpec, len(pkgs))
	for i, p := range pkgs {
		specs[i] = managers.PackageSpec{Name: p, Manager: managerArg}
	}

	opts := install.Options{
		DryRun:            global.dryRun,
		AutoConfirm:       yes,
		PreferredManagers: cfg.PreferredManagers,
		Stdout:            cmd.OutOrStdout(),
		Stdin:             os.Stdin,
	}

	installer := install.New(reg, rep, opts)
	ctx := cmd.Context()
	results := installer.Install(ctx, specs)

	if global.jsonOut {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
	}

	return printInstallResults(cmd, results)
}

func printInstallResults(cmd *cobra.Command, results []install.Result) error {
	hasErr := false
	for _, r := range results {
		switch {
		case r.Skipped:
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s (skipped)\n",
				color.YellowString("~"), r.Spec.Name)
		case r.Err != nil:
			fmt.Fprintf(cmd.ErrOrStderr(), "%s %s: %s\n",
				color.RedString("✗"), r.Spec.Name, r.Err)
			hasErr = true
		default:
			if !global.dryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "%s installed %s via %s\n",
					color.GreenString("✓"), r.Spec.Name, r.Manager)
			}
		}
	}
	if hasErr {
		return fmt.Errorf("one or more packages failed to install")
	}
	return nil
}
