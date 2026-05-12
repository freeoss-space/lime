package managers

import (
	"context"
	"strings"
)

// Emerge implements PackageManager for emerge (Gentoo Portage package manager).
type Emerge struct{ base }

// NewEmerge creates an Emerge manager using the given Commander.
func NewEmerge(cmd Commander) *Emerge {
	return &Emerge{base{name: "emerge", binary: "emerge", cmd: cmd}}
}

func (e *Emerge) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"emerge", pkg}
}

func (e *Emerge) Install(ctx context.Context, pkg string) error {
	bin, args := e.InstallArgs(pkg)
	return e.cmd.Run(ctx, bin, args...)
}

func (e *Emerge) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := e.searchLines(ctx, "--search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	var cur SearchResult
	for _, l := range lines {
		// emerge --search output:
		//   *  category/package
		//         Latest version available: 14.0.0
		if strings.HasPrefix(l, "*") {
			if cur.Package != "" {
				results = append(results, cur)
			}
			name := strings.TrimSpace(strings.TrimPrefix(l, "*"))
			cur = SearchResult{
				Manager: e.name,
				Package: name,
				Repo:    "gentoo",
			}
		} else if strings.Contains(l, "Latest version available:") {
			parts := strings.SplitN(l, ":", 2)
			if len(parts) == 2 {
				cur.Version = strings.TrimSpace(parts[1])
			}
		}
	}
	if cur.Package != "" {
		results = append(results, cur)
	}
	return results, nil
}
