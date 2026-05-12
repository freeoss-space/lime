package managers

import (
	"context"
	"strings"
)

// Winget implements PackageManager for winget (Windows Package Manager).
type Winget struct{ base }

// NewWinget creates a Winget manager.
func NewWinget(cmd Commander) *Winget {
	return &Winget{base{name: "winget", binary: "winget", cmd: cmd}}
}

func (w *Winget) InstallArgs(pkg string) (string, []string) {
	return "winget", []string{"install", "--id", pkg, "--silent", "--accept-package-agreements"}
}

func (w *Winget) Install(ctx context.Context, pkg string) error {
	bin, args := w.InstallArgs(pkg)
	return w.cmd.Run(ctx, bin, args...)
}

func (w *Winget) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := w.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// winget search output: "Name   Id   Version   Source"
		// Skip header lines.
		parts := strings.Fields(l)
		if len(parts) < 2 {
			continue
		}
		if parts[0] == "Name" || strings.HasPrefix(l, "-") {
			continue
		}
		results = append(results, SearchResult{
			Manager: w.name,
			Package: parts[0],
			Repo:    "winget",
		})
	}
	return results, nil
}
