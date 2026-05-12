package cli

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/freeoss-space/lime/internal/config"
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

Examples:
  jil search ripgrep
  jil search --json ripgrep`,
		Args:    cobra.ExactArgs(1),
		RunE:    runSearch,
		Example: "  jil search ripgrep",
	}
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	reg, rep := buildDependencies(cfg)

	s := search.New(reg, rep, search.Options{
		PreferredManagers: cfg.PreferredManagers,
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

	fmt.Fprintln(w, color.HiBlackString("MANAGER\tPACKAGE\tVERSION\tREPO\tAVAIL"))

	for _, r := range results {
		mgr := r.Manager
		if r.Preferred {
			mgr = color.CyanString(mgr)
		}
		avail := ""
		if r.Available {
			avail = color.GreenString("✓")
		}
		version := r.Version
		if version == "" {
			version = color.HiBlackString("—")
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			mgr, r.Package, version, r.Repo, avail)
	}
	return w.Flush()
}
