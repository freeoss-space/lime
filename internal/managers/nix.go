package managers

import (
	"context"
	"strings"
)

// Nix implements PackageManager for nix-env (Nix/NixOS package manager).
type Nix struct{ base }

// NewNix creates a Nix manager using the given Commander.
func NewNix(cmd Commander) *Nix {
	return &Nix{base{name: "nix", binary: "nix-env", cmd: cmd}}
}

func (n *Nix) InstallArgs(pkg string) (string, []string) {
	return "nix-env", []string{"-iA", "nixpkgs." + pkg}
}

func (n *Nix) Install(ctx context.Context, pkg string) error {
	bin, args := n.InstallArgs(pkg)
	return n.cmd.Run(ctx, bin, args...)
}

func (n *Nix) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := n.searchLines(ctx, "-qaP", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// nix-env -qaP output: "nixpkgs.ripgrep  ripgrep-14.0.0"
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		attr := fields[0]    // e.g., nixpkgs.ripgrep
		nameVer := fields[1] // e.g., ripgrep-14.0.0

		// Extract short name from attribute path (after last dot).
		name := attr
		if idx := strings.LastIndex(attr, "."); idx >= 0 {
			name = attr[idx+1:]
		}

		// Extract version by stripping "name-" prefix.
		version := ""
		if strings.HasPrefix(nameVer, name+"-") {
			version = nameVer[len(name)+1:]
		}

		results = append(results, SearchResult{
			Manager: n.name,
			Package: name,
			Version: version,
			Repo:    "nixpkgs",
		})
	}
	return results, nil
}
