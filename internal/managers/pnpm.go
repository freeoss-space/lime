package managers

import (
	"context"
	"strings"
)

// Pnpm implements PackageManager for pnpm (alternative Node.js package manager).
type Pnpm struct{ base }

// NewPnpm creates a Pnpm manager using the given Commander.
func NewPnpm(cmd Commander) *Pnpm {
	return &Pnpm{base{name: "pnpm", binary: "pnpm", cmd: cmd}}
}

func (p *Pnpm) InstallArgs(pkg string) (string, []string) {
	return "pnpm", []string{"add", "-g", pkg}
}

func (p *Pnpm) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

func (p *Pnpm) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := p.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// pnpm search output: "NAME\tVERSION\tDESCRIPTION"
		// Skip header row.
		fields := strings.Fields(l)
		if len(fields) < 2 || strings.EqualFold(fields[0], "NAME") {
			continue
		}
		results = append(results, SearchResult{
			Manager: p.name,
			Package: fields[0],
			Version: fields[1],
			Repo:    "npm",
		})
	}
	return results, nil
}

func (p *Pnpm) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the pnpm command for a specific version.
// pnpm uses "pkg@version" syntax.
func (p *Pnpm) InstallVersionArgs(pkg, version string) (string, []string) {
	return "pnpm", []string{"add", "-g", pkg + "@" + version}
}

func (p *Pnpm) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := p.InstallVersionArgs(pkg, version)
	return p.cmd.Run(ctx, bin, args...)
}
