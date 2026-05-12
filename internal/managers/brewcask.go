package managers

import (
	"context"
	"strings"
)

// BrewCask implements PackageManager for Homebrew casks (macOS GUI apps).
type BrewCask struct{ base }

// NewBrewCask creates a BrewCask manager using the given Commander.
func NewBrewCask(cmd Commander) *BrewCask {
	return &BrewCask{base{name: "brew-cask", binary: "brew", cmd: cmd}}
}

func (b *BrewCask) InstallArgs(pkg string) (string, []string) {
	return "brew", []string{"install", "--cask", pkg}
}

func (b *BrewCask) Install(ctx context.Context, pkg string) error {
	bin, args := b.InstallArgs(pkg)
	return b.cmd.Run(ctx, bin, args...)
}

func (b *BrewCask) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := b.searchLines(ctx, "search", "--casks", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		if strings.HasPrefix(l, "==>") {
			continue
		}
		results = append(results, SearchResult{
			Manager: b.name,
			Package: strings.TrimSpace(l),
			Repo:    "homebrew-cask",
		})
	}
	return results, nil
}
