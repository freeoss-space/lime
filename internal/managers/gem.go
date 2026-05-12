package managers

import (
	"context"
	"strings"
)

// Gem implements PackageManager for gem (Ruby package manager).
type Gem struct{ base }

// NewGem creates a Gem manager using the given Commander.
func NewGem(cmd Commander) *Gem {
	return &Gem{base{name: "gem", binary: "gem", cmd: cmd}}
}

func (g *Gem) InstallArgs(pkg string) (string, []string) {
	return "gem", []string{"install", pkg}
}

func (g *Gem) Install(ctx context.Context, pkg string) error {
	bin, args := g.InstallArgs(pkg)
	return g.cmd.Run(ctx, bin, args...)
}

func (g *Gem) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := g.searchLines(ctx, "search", "--remote", "-q", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// gem search output: "name (version, ...)"
		// Skip header banner "*** REMOTE GEMS ***".
		if strings.HasPrefix(l, "***") {
			continue
		}
		idx := strings.Index(l, " (")
		if idx < 0 {
			continue
		}
		name := strings.TrimSpace(l[:idx])
		if name == "" {
			continue
		}
		rest := l[idx+2:]
		end := strings.Index(rest, ")")
		version := ""
		if end >= 0 {
			// May list multiple versions; use the first one.
			versions := rest[:end]
			if comma := strings.Index(versions, ","); comma >= 0 {
				version = strings.TrimSpace(versions[:comma])
			} else {
				version = strings.TrimSpace(versions)
			}
		}
		results = append(results, SearchResult{
			Manager: g.name,
			Package: name,
			Version: version,
			Repo:    "rubygems",
		})
	}
	return results, nil
}

func (g *Gem) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the gem command for a specific version.
func (g *Gem) InstallVersionArgs(pkg, version string) (string, []string) {
	return "gem", []string{"install", pkg, "--version", version}
}

func (g *Gem) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := g.InstallVersionArgs(pkg, version)
	return g.cmd.Run(ctx, bin, args...)
}
