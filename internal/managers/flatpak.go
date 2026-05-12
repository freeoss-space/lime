package managers

import (
	"context"
	"strings"
)

// Flatpak implements PackageManager for flatpak (sandboxed Linux apps).
type Flatpak struct{ base }

// NewFlatpak creates a Flatpak manager using the given Commander.
func NewFlatpak(cmd Commander) *Flatpak {
	return &Flatpak{base{name: "flatpak", binary: "flatpak", cmd: cmd}}
}

func (f *Flatpak) InstallArgs(pkg string) (string, []string) {
	return "flatpak", []string{"install", "-y", pkg}
}

func (f *Flatpak) Install(ctx context.Context, pkg string) error {
	bin, args := f.InstallArgs(pkg)
	return f.cmd.Run(ctx, bin, args...)
}

func (f *Flatpak) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := f.searchLines(ctx, "search", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// flatpak search output (tab-separated):
		// Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
		parts := strings.Split(l, "\t")
		if len(parts) < 4 {
			continue
		}
		appID := strings.TrimSpace(parts[2])
		version := strings.TrimSpace(parts[3])
		// Skip header row.
		if appID == "Application ID" || appID == "" {
			continue
		}
		results = append(results, SearchResult{
			Manager: f.name,
			Package: appID,
			Version: version,
			Repo:    "flathub",
		})
	}
	return results, nil
}
