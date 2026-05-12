package managers

import (
	"context"
	"strings"
)

// Mise implements PackageManager for mise (multi-language runtime version manager).
type Mise struct{ base }

// NewMise creates a Mise manager using the given Commander.
func NewMise(cmd Commander) *Mise {
	return &Mise{base{name: "mise", binary: "mise", cmd: cmd}}
}

func (m *Mise) InstallArgs(pkg string) (string, []string) {
	return "mise", []string{"use", "-g", pkg}
}

func (m *Mise) Install(ctx context.Context, pkg string) error {
	bin, args := m.InstallArgs(pkg)
	return m.cmd.Run(ctx, bin, args...)
}

func (m *Mise) Search(ctx context.Context, query string) ([]SearchResult, error) {
	// mise registry lists all available tools; filter locally by query.
	lines, err := m.searchLines(ctx, "registry")
	if err != nil {
		return nil, err
	}
	query = strings.ToLower(query)
	var results []SearchResult
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) < 1 {
			continue
		}
		name := fields[0]
		if !strings.Contains(strings.ToLower(name), query) {
			continue
		}
		results = append(results, SearchResult{
			Manager: m.name,
			Package: name,
			Repo:    "mise-registry",
		})
	}
	return results, nil
}

func (m *Mise) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the mise command for a specific version.
// mise uses "plugin@version" syntax.
func (m *Mise) InstallVersionArgs(pkg, version string) (string, []string) {
	return "mise", []string{"use", "-g", pkg + "@" + version}
}

func (m *Mise) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := m.InstallVersionArgs(pkg, version)
	return m.cmd.Run(ctx, bin, args...)
}
