package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/freeoss-space/lime/internal/config"
	"github.com/freeoss-space/lime/internal/cooldown"
	"github.com/freeoss-space/lime/internal/search"
)

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages across package managers",
		Long: `Search for packages matching query via the Repology API.

Results are annotated with:
  - which managers have the package
  - whether those managers are installed on this system
  - your preferred managers (highlighted)
  - version age and cooldown eligibility when timestamps are available

Examples:
  lime search ripgrep
  lime search --json ripgrep`,
		Args:    cobra.ExactArgs(1),
		RunE:    runSearch,
		Example: "  lime search ripgrep",
	}
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reg, rep := buildDependencies(cfg)

	cd, err := cooldown.Parse(cfg.DefaultCooldown)
	if err != nil {
		return fmt.Errorf("invalid default_cooldown in config: %w", err)
	}

	s := search.New(reg, rep, search.Options{
		PreferredManagers: cfg.PreferredManagers,
		Cooldown:          cd,
	})

	ctx := cmd.Context()
	results, err := s.Search(ctx, query)
	if err != nil {
		return fmt.Errorf("searching for %q: %w", query, err)
	}

	if len(results) == 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "No results found for %q\n", query)
		return nil
	}

	if global.jsonOut {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(results)
	}

	return printSearchResults(cmd, results)
}

func printSearchResults(cmd *cobra.Command, results []search.Result) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, color.HiBlackString("MANAGER\tPACKAGE\tVERSION\tAGE\tSTATUS"))

	for _, r := range results {
		mgr := r.Manager
		if r.Preferred {
			mgr = color.CyanString(mgr)
		}
		version := r.Version
		if version == "" {
			version = color.HiBlackString("—")
		}
		age := color.HiBlackString("—")
		if r.DaysOld != nil {
			age = fmt.Sprintf("%dd", *r.DaysOld)
		}
		status := availStatus(r)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			mgr, r.Package, version, age, status)
	}
	return w.Flush()
}

// availStatus returns a short status string for a search result.
func availStatus(r search.Result) string {
	if !r.Available {
		return color.HiBlackString("unavailable")
	}
	if !r.CooldownOK {
		return color.YellowString("cooldown-blocked")
	}
	return color.GreenString("eligible")
}
