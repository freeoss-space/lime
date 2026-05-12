package managers

import (
	"context"
	"strings"
)

// Apk implements PackageManager for apk (Alpine Linux).
type Apk struct{ base }

// NewApk creates an Apk manager.
func NewApk(cmd Commander) *Apk {
	return &Apk{base{name: "apk", binary: "apk", cmd: cmd}}
}

func (a *Apk) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"apk", "add", pkg}
}

func (a *Apk) Install(ctx context.Context, pkg string) error {
	bin, args := a.InstallArgs(pkg)
	return a.cmd.Run(ctx, bin, args...)
}

func (a *Apk) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := a.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// apk search output: "<name>-<version>"
		name := strings.TrimSpace(l)
		if name == "" {
			continue
		}
		results = append(results, SearchResult{
			Manager: a.name,
			Package: name,
			Repo:    "alpine",
		})
	}
	return results, nil
}
