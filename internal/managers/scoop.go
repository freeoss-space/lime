package managers

import (
	"context"
	"strings"
)

// Scoop implements PackageManager for Scoop (Windows).
type Scoop struct{ base }

// NewScoop creates a Scoop manager.
func NewScoop(cmd Commander) *Scoop {
	return &Scoop{base{name: "scoop", binary: "scoop", cmd: cmd}}
}

func (s *Scoop) InstallArgs(pkg string) (string, []string) {
	return "scoop", []string{"install", pkg}
}

func (s *Scoop) Install(ctx context.Context, pkg string) error {
	bin, args := s.InstallArgs(pkg)
	return s.cmd.Run(ctx, bin, args...)
}

func (s *Scoop) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := s.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// scoop search: "  <name> (<bucket>) <version>"
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "Results") {
			continue
		}
		parts := strings.Fields(l)
		if len(parts) < 1 {
			continue
		}
		results = append(results, SearchResult{
			Manager: s.name,
			Package: parts[0],
			Repo:    "scoop",
		})
	}
	return results, nil
}
