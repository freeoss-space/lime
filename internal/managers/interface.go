// Package managers defines the PackageManager interface and supporting types
// for cross-platform package management.
package managers

import (
	"context"
)

// SearchResult is one result from a manager's local search.
type SearchResult struct {
	Manager string
	Package string
	Version string
	Repo    string
}

// PackageManager abstracts a system package manager.
type PackageManager interface {
	// Name returns the canonical name of this manager (e.g. "apt", "brew").
	Name() string
	// IsAvailable reports whether the manager binary exists on this system.
	IsAvailable(ctx context.Context) bool
	// Install installs pkg using this manager.
	Install(ctx context.Context, pkg string) error
	// Search searches for packages matching query.
	Search(ctx context.Context, query string) ([]SearchResult, error)
	// InstallArgs returns the command and arguments needed to install pkg.
	// This is exposed separately to support --dry-run and prompting.
	InstallArgs(pkg string) (string, []string)
}

// PackageSpec is a package with an optional explicit manager override.
type PackageSpec struct {
	// Name is the package name as understood by the manager.
	Name string
	// Manager is an optional explicit manager name. If empty, jil selects one.
	Manager string
}
