package managers

import (
	"context"
	"strings"
)

// Choco implements PackageManager for Chocolatey (Windows).
type Choco struct{ base }

// NewChoco creates a Choco manager.
func NewChoco(cmd Commander) *Choco {
	return &Choco{base{name: "choco", binary: "choco", cmd: cmd}}
}

func (c *Choco) InstallArgs(pkg string) (string, []string) {
	return "choco", []string{"install", pkg, "-y"}
}

func (c *Choco) Install(ctx context.Context, pkg string) error {
	bin, args := c.InstallArgs(pkg)
	return c.cmd.Run(ctx, bin, args...)
}

func (c *Choco) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := c.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// choco search: "<name> <version> [tags]"
		parts := strings.Fields(l)
		if len(parts) < 1 {
			continue
		}
		name := parts[0]
		if name == "Chocolatey" || strings.HasPrefix(name, "---") {
			continue
		}
		r := SearchResult{
			Manager: c.name,
			Package: name,
			Repo:    "chocolatey",
		}
		if len(parts) >= 2 {
			r.Version = parts[1]
		}
		results = append(results, r)
	}
	return results, nil
}
