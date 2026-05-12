package managers

import (
	"context"
	"strings"
)

// Mamba implements PackageManager for mamba (fast conda-compatible package manager).
type Mamba struct{ base }

// NewMamba creates a Mamba manager using the given Commander.
func NewMamba(cmd Commander) *Mamba {
	return &Mamba{base{name: "mamba", binary: "mamba", cmd: cmd}}
}

func (m *Mamba) InstallArgs(pkg string) (string, []string) {
	return "mamba", []string{"install", "-y", pkg}
}

func (m *Mamba) Install(ctx context.Context, pkg string) error {
	bin, args := m.InstallArgs(pkg)
	return m.cmd.Run(ctx, bin, args...)
}

func (m *Mamba) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := m.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		if strings.HasPrefix(l, "#") || strings.HasPrefix(l, "Loading") {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 2 {
			continue
		}
		results = append(results, SearchResult{
			Manager: m.name,
			Package: fields[0],
			Version: fields[1],
			Repo:    "conda-forge",
		})
	}
	return results, nil
}

func (m *Mamba) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the mamba command for a specific version.
// mamba uses "pkg=version" syntax (same as conda).
func (m *Mamba) InstallVersionArgs(pkg, version string) (string, []string) {
	return "mamba", []string{"install", "-y", pkg + "=" + version}
}

func (m *Mamba) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := m.InstallVersionArgs(pkg, version)
	return m.cmd.Run(ctx, bin, args...)
}
