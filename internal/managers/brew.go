package managers

import (
	"context"
	"strings"
)

// Brew implements PackageManager for Homebrew (macOS/Linux).
type Brew struct{ base }

// NewBrew creates a Brew manager using the given Commander.
func NewBrew(cmd Commander) *Brew {
	return &Brew{base{name: "brew", binary: "brew", cmd: cmd}}
}

func (b *Brew) InstallArgs(pkg string) (string, []string) {
	return "brew", []string{"install", pkg}
}

func (b *Brew) Install(ctx context.Context, pkg string) error {
	bin, args := b.InstallArgs(pkg)
	return b.cmd.Run(ctx, bin, args...)
}

func (b *Brew) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := b.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// brew search output may include section headers like "==> Formulae"
		if strings.HasPrefix(l, "==>") {
			continue
		}
		results = append(results, SearchResult{
			Manager: b.name,
			Package: strings.TrimSpace(l),
			Repo:    "homebrew",
		})
	}
	return results, nil
}
