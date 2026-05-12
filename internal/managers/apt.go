package managers

import (
	"context"
	"strings"
)

// Apt implements PackageManager for apt (Debian/Ubuntu).
type Apt struct{ base }

// NewApt creates an Apt manager using the given Commander.
func NewApt(cmd Commander) *Apt {
	return &Apt{base{name: "apt", binary: "apt-get", cmd: cmd}}
}

func (a *Apt) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"apt-get", "install", "-y", pkg}
}

func (a *Apt) Install(ctx context.Context, pkg string) error {
	bin, args := a.InstallArgs(pkg)
	return a.cmd.Run(ctx, bin, args...)
}

func (a *Apt) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := a.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// apt-cache search output: "<name> - <description>"
		parts := strings.SplitN(l, " - ", 2)
		if len(parts) < 1 {
			continue
		}
		results = append(results, SearchResult{
			Manager: a.name,
			Package: strings.TrimSpace(parts[0]),
			Repo:    "apt",
		})
	}
	return results, nil
}
