package managers

import (
	"context"
	"strings"
)

// Uv implements PackageManager for uv (fast Python package manager).
type Uv struct{ base }

// NewUv creates a Uv manager using the given Commander.
func NewUv(cmd Commander) *Uv {
	return &Uv{base{name: "uv", binary: "uv", cmd: cmd}}
}

func (u *Uv) InstallArgs(pkg string) (string, []string) {
	return "uv", []string{"tool", "install", pkg}
}

func (u *Uv) Install(ctx context.Context, pkg string) error {
	bin, args := u.InstallArgs(pkg)
	return u.cmd.Run(ctx, bin, args...)
}

func (u *Uv) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := u.searchLines(ctx, "pip", "index", "versions", query)
	if err != nil {
		return nil, nil
	}
	var results []SearchResult
	for _, l := range lines {
		if strings.HasPrefix(l, "WARNING:") || strings.HasPrefix(l, "NOTICE:") ||
			strings.HasPrefix(l, "Available versions") {
			continue
		}
		idx := strings.Index(l, " (")
		if idx < 0 {
			continue
		}
		end := strings.Index(l[idx:], ")")
		if end < 0 {
			continue
		}
		name := l[:idx]
		version := l[idx+2 : idx+end]
		results = append(results, SearchResult{
			Manager: u.name,
			Package: name,
			Version: version,
			Repo:    "pypi",
		})
		break
	}
	return results, nil
}

func (u *Uv) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the uv command for a specific version.
// uv tool install uses PEP 508 specifiers: "pkg==version".
func (u *Uv) InstallVersionArgs(pkg, version string) (string, []string) {
	return "uv", []string{"tool", "install", pkg + "==" + version}
}

func (u *Uv) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := u.InstallVersionArgs(pkg, version)
	return u.cmd.Run(ctx, bin, args...)
}
