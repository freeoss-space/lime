package managers

import (
	"context"
	"strings"
)

// Pkgin implements PackageManager for pkgin (NetBSD binary package manager).
type Pkgin struct{ base }

// NewPkgin creates a Pkgin manager using the given Commander.
func NewPkgin(cmd Commander) *Pkgin {
	return &Pkgin{base{name: "pkgin", binary: "pkgin", cmd: cmd}}
}

func (p *Pkgin) InstallArgs(pkg string) (string, []string) {
	return "pkgin", []string{"-y", "install", pkg}
}

func (p *Pkgin) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

func (p *Pkgin) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := p.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// pkgin search output: "name-version; Description"
		idx := strings.Index(l, ";")
		nameVer := l
		if idx >= 0 {
			nameVer = strings.TrimSpace(l[:idx])
		}
		if nameVer == "" {
			continue
		}
		// Split name-version at the last dash.
		dashIdx := strings.LastIndex(nameVer, "-")
		name := nameVer
		version := ""
		if dashIdx > 0 {
			name = nameVer[:dashIdx]
			version = nameVer[dashIdx+1:]
		}
		results = append(results, SearchResult{
			Manager: p.name,
			Package: name,
			Version: version,
			Repo:    "pkgsrc",
		})
	}
	return results, nil
}
