package managers

import (
	"context"
	"strings"
)

// Opkg implements PackageManager for opkg (OpenWrt/embedded Linux package manager).
type Opkg struct{ base }

// NewOpkg creates an Opkg manager using the given Commander.
func NewOpkg(cmd Commander) *Opkg {
	return &Opkg{base{name: "opkg", binary: "opkg", cmd: cmd}}
}

func (o *Opkg) InstallArgs(pkg string) (string, []string) {
	return "opkg", []string{"install", pkg}
}

func (o *Opkg) Install(ctx context.Context, pkg string) error {
	bin, args := o.InstallArgs(pkg)
	return o.cmd.Run(ctx, bin, args...)
}

func (o *Opkg) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := o.searchLines(ctx, "find", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	for _, l := range lines {
		// opkg find output: "name - version - description"
		parts := strings.SplitN(l, " - ", 3)
		if len(parts) < 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		if name == "" {
			continue
		}
		results = append(results, SearchResult{
			Manager: o.name,
			Package: name,
			Version: version,
			Repo:    "openwrt",
		})
	}
	return results, nil
}
