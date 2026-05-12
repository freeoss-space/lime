package managers

import (
	"context"
	"strings"
)

// Helm implements PackageManager for helm (Kubernetes chart manager).
type Helm struct{ base }

// NewHelm creates a Helm manager using the given Commander.
func NewHelm(cmd Commander) *Helm {
	return &Helm{base{name: "helm", binary: "helm", cmd: cmd}}
}

// releaseName derives a Helm release name from a chart reference.
// For "bitnami/redis" it returns "redis"; for "redis" it returns "redis".
func releaseName(chart string) string {
	if idx := strings.LastIndex(chart, "/"); idx >= 0 {
		return chart[idx+1:]
	}
	return chart
}

func (h *Helm) InstallArgs(pkg string) (string, []string) {
	return "helm", []string{"install", releaseName(pkg), pkg}
}

func (h *Helm) Install(ctx context.Context, pkg string) error {
	bin, args := h.InstallArgs(pkg)
	return h.cmd.Run(ctx, bin, args...)
}

func (h *Helm) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := h.searchLines(ctx, "search", "hub", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// helm search hub output (tab-separated):
		// URL\tCHART VERSION\tAPP VERSION\tDESCRIPTION
		if strings.HasPrefix(l, "URL") {
			continue // skip header
		}
		parts := strings.Split(l, "\t")
		if len(parts) < 2 {
			continue
		}
		url := strings.TrimSpace(parts[0])
		chartVersion := strings.TrimSpace(parts[1])

		// Extract chart name from Artifact Hub URL (last path component).
		name := url
		if idx := strings.LastIndex(url, "/"); idx >= 0 {
			name = url[idx+1:]
		}

		results = append(results, SearchResult{
			Manager: h.name,
			Package: name,
			Version: chartVersion,
			Repo:    "artifacthub",
		})
	}
	return results, nil
}

func (h *Helm) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the helm command for a specific chart version.
func (h *Helm) InstallVersionArgs(pkg, version string) (string, []string) {
	return "helm", []string{"install", releaseName(pkg), pkg, "--version", version}
}

func (h *Helm) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := h.InstallVersionArgs(pkg, version)
	return h.cmd.Run(ctx, bin, args...)
}
