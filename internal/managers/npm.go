package managers

import (
	"context"
	"strings"
)

// Npm implements PackageManager for npm (Node.js package manager).
type Npm struct{ base }

// NewNpm creates an Npm manager using the given Commander.
func NewNpm(cmd Commander) *Npm {
	return &Npm{base{name: "npm", binary: "npm", cmd: cmd}}
}

func (n *Npm) InstallArgs(pkg string) (string, []string) {
	return "npm", []string{"install", "-g", pkg}
}

func (n *Npm) Install(ctx context.Context, pkg string) error {
	bin, args := n.InstallArgs(pkg)
	return n.cmd.Run(ctx, bin, args...)
}

func (n *Npm) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := n.searchLines(ctx, "search", "--parseable", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// npm search --parseable: NAME\tDESCRIPTION\tAUTHOR\tDATE\tVERSION\tKEYWORDS
		parts := strings.Split(l, "\t")
		if len(parts) < 5 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[4])
		// Skip the header row.
		if name == "" || name == "NAME" {
			continue
		}
		results = append(results, SearchResult{
			Manager: n.name,
			Package: name,
			Version: version,
			Repo:    "npm",
		})
	}
	return results, nil
}

func (n *Npm) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the npm command for a specific version.
// npm uses the "pkg@version" syntax for global installs.
func (n *Npm) InstallVersionArgs(pkg, version string) (string, []string) {
	return "npm", []string{"install", "-g", pkg + "@" + version}
}

func (n *Npm) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := n.InstallVersionArgs(pkg, version)
	return n.cmd.Run(ctx, bin, args...)
}
