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

	// SupportsVersioning reports whether this manager can install specific versions.
	// Managers that return false will fall back to installing the latest available version.
	SupportsVersioning() bool
	// InstallVersionArgs returns the command and arguments to install a specific version.
	// Only called when SupportsVersioning() returns true.
	InstallVersionArgs(pkg, version string) (string, []string)
	// InstallVersion installs a specific version of pkg.
	// Returns an error describing the limitation when SupportsVersioning() is false.
	InstallVersion(ctx context.Context, pkg, version string) error
}

// PackageSpec is a package with an optional explicit manager override.
type PackageSpec struct {
	// Name is the package name as understood by the manager.
	Name string
	// Manager is an optional explicit manager name. If empty, jil selects one.
	Manager string
	// Version is an optional version spec (e.g. "14.1.1", "20", "3.12").
	// Empty means "any version".
	Version string
}
