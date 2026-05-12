package managers

import (
	"context"
	"strings"
)

// Pkg implements PackageManager for FreeBSD pkg.
type Pkg struct{ base }

// NewPkg creates a Pkg manager.
func NewPkg(cmd Commander) *Pkg {
	return &Pkg{base{name: "pkg", binary: "pkg", cmd: cmd}}
}

func (p *Pkg) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"pkg", "install", "-y", pkg}
}

func (p *Pkg) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

func (p *Pkg) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := p.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// pkg search: "<name>-<version> <description>"
		parts := strings.Fields(l)
		if len(parts) < 1 {
			continue
		}
		results = append(results, SearchResult{
			Manager: p.name,
			Package: parts[0],
			Repo:    "freebsd",
		})
	}
	return results, nil
}
