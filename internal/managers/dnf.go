package managers

import (
	"context"
	"strings"
)

// Dnf implements PackageManager for dnf (Fedora/RHEL/CentOS).
type Dnf struct{ base }

// NewDnf creates a Dnf manager using the given Commander.
func NewDnf(cmd Commander) *Dnf {
	return &Dnf{base{name: "dnf", binary: "dnf", cmd: cmd}}
}

func (d *Dnf) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"dnf", "install", "-y", pkg}
}

func (d *Dnf) Install(ctx context.Context, pkg string) error {
	bin, args := d.InstallArgs(pkg)
	return d.cmd.Run(ctx, bin, args...)
}

func (d *Dnf) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := d.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// dnf search output: "<name>.<arch> : <description>"
		parts := strings.SplitN(l, " : ", 2)
		if len(parts) < 1 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		// Strip arch suffix (e.g. ".x86_64")
		if idx := strings.LastIndex(name, "."); idx > 0 {
			name = name[:idx]
		}
		if name == "" || strings.HasPrefix(name, "=") {
			continue
		}
		results = append(results, SearchResult{
			Manager: d.name,
			Package: name,
			Repo:    "dnf",
		})
	}
	return results, nil
}
