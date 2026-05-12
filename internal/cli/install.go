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
	"github.com/freeoss-space/lime/internal/versioning"
)

func newInstallCmd() *cobra.Command {
	var (
		yes        bool
		managerArg string
		cooldown   string
	)

	cmd := &cobra.Command{
		Use:   "install <package[@version]> [package...]",
		Short: "Install one or more packages",
		Long: `Install one or more packages using the best available package manager.

lime queries Repology to resolve the canonical package name for each manager,
then installs using your preferred manager. Falls back to other available
managers if the preferred one is not installed on the system.

You can request a specific version using the "@" syntax:

  lime install ripgrep@14.1.1
  lime install nodejs@20
  lime install python@3.12

Use --cooldown to skip versions newer than a given age:

  lime install ripgrep --cooldown 14d
  lime install ripgrep --cooldown 2w

Supported cooldown units: h (hours), d (days), w (weeks).
Configure a global default in config: default_cooldown = "14d"`,
		Args:    cobra.MinimumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return runInstall(cmd, args, yes, managerArg, cooldown) },
		Example: "  lime install ripgrep\n  lime install ripgrep@14.1.1\n  lime install -y ripgrep --cooldown 14d",
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "auto-confirm installation without prompt")
	cmd.Flags().StringVar(&managerArg, "manager", "", "force a specific package manager")
	cmd.Flags().StringVar(&cooldown, "cooldown", "", "skip versions newer than this age (e.g. 14d, 2w, 24h)")

	return cmd
}

func runInstall(cmd *cobra.Command, pkgs []string, yes bool, managerArg, cooldownArg string) error {
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
		name, vspec := versioning.ParsePackageArg(p)
		specs[i] = managers.PackageSpec{
			Name:    name,
			Manager: managerArg,
			Version: vspec.Raw,
		}
	}

	opts := install.Options{
		DryRun:            global.dryRun,
		AutoConfirm:       yes,
		PreferredManagers: cfg.PreferredManagers,
		Stdout:            cmd.OutOrStdout(),
		Stdin:             os.Stdin,
		Cooldown:          cooldownArg,
		DefaultCooldown:   cfg.DefaultCooldown,
		ManagerCooldowns:  cfg.ManagerCooldowns,
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
				if r.Version != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "%s installed %s@%s via %s\n",
						color.GreenString("✓"), r.Spec.Name, r.Version, r.Manager)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s installed %s via %s\n",
						color.GreenString("✓"), r.Spec.Name, r.Manager)
				}
			}
		}
	}
	if hasErr {
		return fmt.Errorf("one or more packages failed to install")
	}
	return nil
}
