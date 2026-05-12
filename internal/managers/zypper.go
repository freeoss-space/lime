package managers

import (
	"context"
	"strings"
)

// Zypper implements PackageManager for zypper (openSUSE).
type Zypper struct{ base }

// NewZypper creates a Zypper manager.
func NewZypper(cmd Commander) *Zypper {
	return &Zypper{base{name: "zypper", binary: "zypper", cmd: cmd}}
}

func (z *Zypper) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"zypper", "install", "-y", pkg}
}

func (z *Zypper) Install(ctx context.Context, pkg string) error {
	bin, args := z.InstallArgs(pkg)
	return z.cmd.Run(ctx, bin, args...)
}

func (z *Zypper) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := z.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// zypper search output uses | delimiters: "| name | summary | type |"
		if !strings.Contains(l, "|") || strings.Contains(l, "---") {
			continue
		}
		parts := strings.Split(l, "|")
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[1])
		if name == "Name" || name == "" {
			continue
		}
		results = append(results, SearchResult{
			Manager: z.name,
			Package: name,
			Repo:    "zypper",
		})
	}
	return results, nil
}

func (z *Zypper) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the zypper command for a specific version.
// zypper uses the "pkg=version" syntax.
func (z *Zypper) InstallVersionArgs(pkg, version string) (string, []string) {
	return "sudo", []string{"zypper", "install", "-y", pkg + "=" + version}
}

func (z *Zypper) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := z.InstallVersionArgs(pkg, version)
	return z.cmd.Run(ctx, bin, args...)
}
