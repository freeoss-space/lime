package managers

import (
	"context"
	"strings"
)

// Snap implements PackageManager for snap (Ubuntu/Canonical universal packages).
type Snap struct{ base }

// NewSnap creates a Snap manager using the given Commander.
func NewSnap(cmd Commander) *Snap {
	return &Snap{base{name: "snap", binary: "snap", cmd: cmd}}
}

func (s *Snap) InstallArgs(pkg string) (string, []string) {
	return "snap", []string{"install", pkg}
}

func (s *Snap) Install(ctx context.Context, pkg string) error {
	bin, args := s.InstallArgs(pkg)
	return s.cmd.Run(ctx, bin, args...)
}

func (s *Snap) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := s.searchLines(ctx, "find", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// snap find output: "Name\tVersion\tPublisher\tNotes\tSummary"
		// Skip header row.
		if strings.HasPrefix(l, "Name") {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		results = append(results, SearchResult{
			Manager: s.name,
			Package: fields[0],
			Version: fields[1],
			Repo:    "snapcraft",
		})
	}
	return results, nil
}
