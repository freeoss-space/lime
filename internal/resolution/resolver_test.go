package resolution_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/cooldown"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/internal/resolution"
	"github.com/freeoss-space/lime/internal/versioning"
)

var fixedNow = time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

func newResolver() *resolution.Resolver {
	return resolution.NewWithClock(func() time.Time { return fixedNow })
}

func ptr(t time.Time) *time.Time { return &t }

// pkg14old is ripgrep 14.1.1 for homebrew, released 30 days ago.
var pkg14old = repology.Package{
	Repo:      "homebrew",
	Name:      "ripgrep",
	Version:   "14.1.1",
	UpdatedAt: ptr(fixedNow.Add(-30 * 24 * time.Hour)),
}

// pkg15new is a newer version released only 2 days ago.
var pkg15new = repology.Package{
	Repo:      "homebrew",
	Name:      "ripgrep",
	Version:   "15.0.0",
	UpdatedAt: ptr(fixedNow.Add(-2 * 24 * time.Hour)),
}

// pkgNoTimestamp has no release date.
var pkgNoTimestamp = repology.Package{
	Repo:    "homebrew",
	Name:    "ripgrep",
	Version: "14.0.0",
}

// aptPkg is an apt package.
var aptPkg = repology.Package{
	Repo:      "debian_stable",
	Name:      "ripgrep",
	Version:   "13.0.0",
	UpdatedAt: ptr(fixedNow.Add(-180 * 24 * time.Hour)),
}

// --- Basic selection ---

func TestForManager_ReturnsCandidateWhenAvailable(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old}

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "ripgrep", c.Package)
	assert.Equal(t, "14.1.1", c.Version)
	assert.Equal(t, "brew", c.Manager)
}

func TestForManager_ReturnsNilWhenNoPackagesForManager(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{aptPkg} // only apt

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	assert.Nil(t, c, "no error and no candidate when manager has no packages")
}

func TestForManager_ReturnsNilWhenEmptyPackageList(t *testing.T) {
	r := newResolver()
	c, err := r.ForManager(nil, "brew", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	assert.Nil(t, c)
}

// --- Version spec filtering ---

func TestForManager_ExactVersionMatch(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old}
	spec := versioning.Spec{Raw: "14.1.1"}

	c, err := r.ForManager(pkgs, "brew", spec, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "14.1.1", c.Version)
}

func TestForManager_ExactVersionNoMatch(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old}
	spec := versioning.Spec{Raw: "99.0.0"} // doesn't exist

	c, err := r.ForManager(pkgs, "brew", spec, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	assert.Nil(t, c, "no candidate when version doesn't match")
}

func TestForManager_PrefixVersionMatch(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old}
	spec := versioning.Spec{Raw: "14"} // prefix: matches 14.1.1

	c, err := r.ForManager(pkgs, "brew", spec, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "14.1.1", c.Version)
}

func TestForManager_PrefixVersionNoMatch(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old}
	spec := versioning.Spec{Raw: "15"} // no 15.x in packages

	c, err := r.ForManager(pkgs, "brew", spec, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	assert.Nil(t, c)
}

// --- Cooldown filtering ---

func TestForManager_CooldownBlocksNewRelease(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg15new} // 2 days old
	cd, _ := cooldown.Parse("10d")       // 10d is not a whole week, stays "10d"

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	assert.Nil(t, c)
	require.Error(t, err)

	var cooldownErr *resolution.CooldownError
	require.ErrorAs(t, err, &cooldownErr)
	assert.Equal(t, "brew", cooldownErr.Manager)
	assert.Equal(t, "10d", cooldownErr.Cooldown.String())
	require.Len(t, cooldownErr.Blocked, 1)
	assert.Equal(t, "15.0.0", cooldownErr.Blocked[0].Version)
}

func TestForManager_CooldownPassesOldRelease(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old} // 30 days old
	cd, _ := cooldown.Parse("14d")

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "14.1.1", c.Version)
}

func TestForManager_NoTimestampPassesCooldown(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkgNoTimestamp}
	cd, _ := cooldown.Parse("14d")

	// No timestamp → can't determine age → treat as acceptable
	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c, "package without timestamp should not be blocked")
	assert.Equal(t, "14.0.0", c.Version)
}

func TestForManager_ZeroCooldownIgnoresTimestamp(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg15new} // 2 days old

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c, "zero cooldown never blocks")
}

func TestForManager_CooldownSelectsOlderWhenNewerBlocked(t *testing.T) {
	r := newResolver()
	// Two homebrew packages: new one blocked, old one passes
	oldPkg := repology.Package{
		Repo:      "homebrew",
		Name:      "ripgrep",
		Version:   "14.1.1",
		UpdatedAt: ptr(fixedNow.Add(-30 * 24 * time.Hour)),
	}
	newPkg := repology.Package{
		Repo:      "homebrew",
		Name:      "ripgrep",
		Version:   "15.0.0",
		UpdatedAt: ptr(fixedNow.Add(-2 * 24 * time.Hour)),
	}
	pkgs := []repology.Package{oldPkg, newPkg}
	cd, _ := cooldown.Parse("14d")

	c, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	// Old package should be returned since new one is blocked
	assert.Equal(t, "14.1.1", c.Version)
}

// --- Multi-manager scenarios ---

func TestForManager_OnlyFiltersTargetManager(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg14old, aptPkg}

	// Asking for apt — should return the apt package, not brew
	c, err := r.ForManager(pkgs, "apt", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "apt", c.Manager)
	assert.Equal(t, "13.0.0", c.Version)
}

// --- CooldownError details ---

func TestCooldownError_ContainsVersion(t *testing.T) {
	r := newResolver()
	pkgs := []repology.Package{pkg15new}
	cd, _ := cooldown.Parse("10d") // 10d stays as "10d" (not a whole week)

	_, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "15.0.0")
	assert.Contains(t, err.Error(), "brew")
	assert.Contains(t, err.Error(), "10d")
	assert.Contains(t, err.Error(), "--cooldown 0d")
}

func TestCooldownError_AgeInDays(t *testing.T) {
	r := newResolver()
	releasedThreeDaysAgo := fixedNow.Add(-3 * 24 * time.Hour)
	pkgs := []repology.Package{{
		Repo:      "homebrew",
		Name:      "ripgrep",
		Version:   "15.0.0",
		UpdatedAt: &releasedThreeDaysAgo,
	}}
	cd, _ := cooldown.Parse("14d")

	_, err := r.ForManager(pkgs, "brew", versioning.Spec{}, cd, "ripgrep")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "3 days ago")
}

// --- EffectiveName fallback ---

func TestForManager_UsesProjectNameWhenPackageNameEmpty(t *testing.T) {
	r := newResolver()
	pkg := repology.Package{
		Repo:    "homebrew",
		Version: "14.1.1",
		// Name is empty — EffectiveName should fall back to projectName
	}

	c, err := r.ForManager([]repology.Package{pkg}, "brew", versioning.Spec{}, cooldown.Zero, "ripgrep")
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.Equal(t, "ripgrep", c.Package, "should fall back to projectName")
}
