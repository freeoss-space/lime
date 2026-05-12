// Package resolution provides version selection logic that combines version spec
// matching and cooldown filtering over Repology package metadata.
//
// It is intentionally independent from the CLI and installation layers, making
// version selection logic fully testable in isolation.
package resolution

import (
	"fmt"
	"time"

	"github.com/freeoss-space/lime/internal/cooldown"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/internal/versioning"
)

// Candidate is a resolved package version for a specific manager.
type Candidate struct {
	// Package is the name as understood by the package manager.
	Package string
	// Version is the package version string.
	Version string
	// Manager is the package manager name (e.g. "brew", "apt").
	Manager string
	// Repo is the Repology repository identifier.
	Repo string
	// UpdatedAt is the release/update timestamp when known.
	UpdatedAt *time.Time
}

// BlockedVersion records a version that was skipped by the cooldown filter.
type BlockedVersion struct {
	// Version is the version string that was blocked.
	Version string
	// Age is how long ago this version was released.
	Age time.Duration
}

// CooldownError is returned when matching versions exist for a manager but all
// are younger than the cooldown window.
type CooldownError struct {
	// Blocked lists versions that were rejected by the cooldown.
	Blocked []BlockedVersion
	// Cooldown is the active cooldown duration.
	Cooldown cooldown.Duration
	// Manager is the manager for which no version was acceptable.
	Manager string
}

func (e *CooldownError) Error() string {
	if len(e.Blocked) == 0 {
		return fmt.Sprintf("no acceptable version found for %s (cooldown: %s)", e.Manager, e.Cooldown)
	}
	b := e.Blocked[0]
	days := int(b.Age.Hours() / 24)
	return fmt.Sprintf(
		"no acceptable version found for %s\n\nNewest version:\n  %s (released %d days ago)\n\nCooldown policy:\n  %s\n\nUse --cooldown 0d to override.",
		e.Manager, b.Version, days, e.Cooldown,
	)
}

// Resolver selects package versions by applying version spec and cooldown constraints.
type Resolver struct {
	now func() time.Time
}

// New creates a Resolver using the system clock.
func New() *Resolver { return &Resolver{now: time.Now} }

// NewWithClock creates a Resolver with an injectable clock, enabling deterministic tests.
func NewWithClock(now func() time.Time) *Resolver { return &Resolver{now: now} }

// ForManager selects the best available version of a package for manager, applying
// spec and cooldown constraints.
//
// Return semantics:
//   - (candidate, nil): acceptable version found.
//   - (nil, *CooldownError): packages exist but are all cooldown-blocked.
//   - (nil, error): another error occurred.
//   - (nil, nil): no packages exist for this manager (not an error).
func (r *Resolver) ForManager(
	pkgs []repology.Package,
	manager string,
	spec versioning.Spec,
	cd cooldown.Duration,
	projectName string,
) (*Candidate, error) {
	now := r.now()
	var matching []Candidate
	var blocked []BlockedVersion

	for _, p := range pkgs {
		if repology.ManagerForRepo(p.Repo) != manager {
			continue
		}
		if !spec.Matches(p.Version) {
			continue
		}
		if p.UpdatedAt != nil && cd.Blocks(*p.UpdatedAt, now) {
			age := now.Sub(*p.UpdatedAt)
			blocked = append(blocked, BlockedVersion{Version: p.Version, Age: age})
			continue
		}
		matching = append(matching, Candidate{
			Package:   p.EffectiveName(projectName),
			Version:   p.Version,
			Manager:   manager,
			Repo:      p.Repo,
			UpdatedAt: p.UpdatedAt,
		})
	}

	if len(matching) > 0 {
		// When multiple candidates exist, prefer the one marked "newest" in Repology.
		// As a fallback, take the first candidate encountered.
		return &matching[0], nil
	}
	if len(blocked) > 0 {
		return nil, &CooldownError{Blocked: blocked, Cooldown: cd, Manager: manager}
	}
	// No packages for this manager — not an error, just not available.
	return nil, nil
}
