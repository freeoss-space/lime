package managers

import (
	"context"
	"strings"
)

// Conda implements PackageManager for conda (Python/data-science environment manager).
type Conda struct{ base }

// NewConda creates a Conda manager using the given Commander.
func NewConda(cmd Commander) *Conda {
	return &Conda{base{name: "conda", binary: "conda", cmd: cmd}}
}

func (c *Conda) InstallArgs(pkg string) (string, []string) {
	return "conda", []string{"install", "-y", pkg}
}

func (c *Conda) Install(ctx context.Context, pkg string) error {
	bin, args := c.InstallArgs(pkg)
	return c.cmd.Run(ctx, bin, args...)
}

func (c *Conda) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := c.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// Skip comment/header lines ("# Name  Version  Build  Channel", "Loading channels:").
		if strings.HasPrefix(l, "#") || strings.HasPrefix(l, "Loading") {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		results = append(results, SearchResult{
			Manager: c.name,
			Package: fields[0],
			Version: fields[1],
			Repo:    "conda-forge",
		})
	}
	return results, nil
}

func (c *Conda) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the conda command for a specific version.
// conda uses "pkg=version" syntax.
func (c *Conda) InstallVersionArgs(pkg, version string) (string, []string) {
	return "conda", []string{"install", "-y", pkg + "=" + version}
}

func (c *Conda) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := c.InstallVersionArgs(pkg, version)
	return c.cmd.Run(ctx, bin, args...)
}
