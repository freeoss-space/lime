package managers

import (
	"context"
	"strings"
)

// Cargo implements PackageManager for cargo (Rust package manager).
type Cargo struct{ base }

// NewCargo creates a Cargo manager using the given Commander.
func NewCargo(cmd Commander) *Cargo {
	return &Cargo{base{name: "cargo", binary: "cargo", cmd: cmd}}
}

func (c *Cargo) InstallArgs(pkg string) (string, []string) {
	return "cargo", []string{"install", pkg}
}

func (c *Cargo) Install(ctx context.Context, pkg string) error {
	bin, args := c.InstallArgs(pkg)
	return c.cmd.Run(ctx, bin, args...)
}

func (c *Cargo) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := c.searchLines(ctx, "search", "--limit", "10", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// cargo search output: `ripgrep = "14.1.1"    # description`
		if !strings.Contains(l, ` = "`) {
			continue
		}
		parts := strings.SplitN(l, ` = "`, 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		rest := parts[1]
		end := strings.Index(rest, `"`)
		version := ""
		if end >= 0 {
			version = rest[:end]
		}
		results = append(results, SearchResult{
			Manager: c.name,
			Package: name,
			Version: version,
			Repo:    "crates.io",
		})
	}
	return results, nil
}

func (c *Cargo) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the cargo command for a specific version.
// cargo uses "cargo install --version <version> <pkg>" syntax.
func (c *Cargo) InstallVersionArgs(pkg, version string) (string, []string) {
	return "cargo", []string{"install", "--version", version, pkg}
}

func (c *Cargo) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := c.InstallVersionArgs(pkg, version)
	return c.cmd.Run(ctx, bin, args...)
}
