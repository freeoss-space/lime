package managers

import (
	"context"
	"strings"
)

// Pip implements PackageManager for pip (Python package manager).
type Pip struct{ base }

// NewPip creates a Pip manager using the given Commander.
func NewPip(cmd Commander) *Pip {
	return &Pip{base{name: "pip", binary: "pip3", cmd: cmd}}
}

func (p *Pip) InstallArgs(pkg string) (string, []string) {
	return "pip3", []string{"install", "--user", pkg}
}

func (p *Pip) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

func (p *Pip) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := p.searchLines(ctx, "index", "versions", query)
	if err != nil {
		// pip index is experimental and may not be available; silently return empty.
		return nil, nil
	}
	var results []SearchResult
	for _, l := range lines {
		// Skip pip warning/notice lines and the "Available versions" line.
		if strings.HasPrefix(l, "WARNING:") || strings.HasPrefix(l, "NOTICE:") ||
			strings.HasPrefix(l, "Available versions") {
			continue
		}
		// Expected format: "<name> (<version>)"
		idx := strings.Index(l, " (")
		if idx < 0 {
			continue
		}
		end := strings.Index(l[idx:], ")")
		if end < 0 {
			continue
		}
		name := l[:idx]
		version := l[idx+2 : idx+end]
		results = append(results, SearchResult{
			Manager: p.name,
			Package: name,
			Version: version,
			Repo:    "pypi",
		})
		break // pip index versions shows one package; stop after first match
	}
	return results, nil
}

func (p *Pip) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the pip command for a specific version.
// pip uses the "pkg==version" syntax.
func (p *Pip) InstallVersionArgs(pkg, version string) (string, []string) {
	return "pip3", []string{"install", "--user", pkg + "==" + version}
}

func (p *Pip) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := p.InstallVersionArgs(pkg, version)
	return p.cmd.Run(ctx, bin, args...)
}
