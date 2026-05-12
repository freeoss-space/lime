package versioning_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/freeoss-space/lime/internal/versioning"
)

// --- ParsePackageArg ---

func TestParsePackageArg_NoVersion(t *testing.T) {
	name, spec := versioning.ParsePackageArg("ripgrep")
	assert.Equal(t, "ripgrep", name)
	assert.True(t, spec.IsAny())
}

func TestParsePackageArg_ExactVersion(t *testing.T) {
	name, spec := versioning.ParsePackageArg("ripgrep@14.1.1")
	assert.Equal(t, "ripgrep", name)
	assert.Equal(t, "14.1.1", spec.Raw)
}

func TestParsePackageArg_MajorOnlyVersion(t *testing.T) {
	name, spec := versioning.ParsePackageArg("nodejs@20")
	assert.Equal(t, "nodejs", name)
	assert.Equal(t, "20", spec.Raw)
}

func TestParsePackageArg_MinorVersion(t *testing.T) {
	name, spec := versioning.ParsePackageArg("python@3.12")
	assert.Equal(t, "python", name)
	assert.Equal(t, "3.12", spec.Raw)
}

func TestParsePackageArg_EmptyVersionSuffix(t *testing.T) {
	// "pkg@" — trailing @ with no version → treat as no version
	name, spec := versioning.ParsePackageArg("ripgrep@")
	assert.Equal(t, "ripgrep@", name)
	assert.True(t, spec.IsAny())
}

func TestParsePackageArg_LeadingAtSign(t *testing.T) {
	// "@pkg" — leading @ is unusual; treat whole string as package name
	name, spec := versioning.ParsePackageArg("@scope/pkg")
	// idx of last @ is at position 0, so idx <= 0 → no version
	assert.Equal(t, "@scope/pkg", name)
	assert.True(t, spec.IsAny())
}

func TestParsePackageArg_LastAtUsed(t *testing.T) {
	// Scoped package with version: "@scope/pkg@1.0.0"
	// LastIndex finds the last @, which is before "1.0.0"
	name, spec := versioning.ParsePackageArg("@scope/pkg@1.0.0")
	assert.Equal(t, "@scope/pkg", name)
	assert.Equal(t, "1.0.0", spec.Raw)
}

func TestParsePackageArg_NoAt(t *testing.T) {
	name, spec := versioning.ParsePackageArg("bat")
	assert.Equal(t, "bat", name)
	assert.True(t, spec.IsAny())
}

// --- IsAny ---

func TestSpec_IsAny_ZeroValue(t *testing.T) {
	var s versioning.Spec
	assert.True(t, s.IsAny())
}

func TestSpec_IsAny_WithVersion(t *testing.T) {
	s := versioning.Spec{Raw: "1.0"}
	assert.False(t, s.IsAny())
}

// --- Matches ---

func TestSpec_Matches_AnyMatchesEverything(t *testing.T) {
	s := versioning.Spec{} // any
	assert.True(t, s.Matches("14.1.1"))
	assert.True(t, s.Matches(""))
	assert.True(t, s.Matches("anything"))
}

func TestSpec_Matches_ExactVersion(t *testing.T) {
	s := versioning.Spec{Raw: "14.1.1"}
	assert.True(t, s.Matches("14.1.1"))
	assert.False(t, s.Matches("14.1.2"))
	assert.False(t, s.Matches("14.1"))
}

func TestSpec_Matches_MajorPrefix(t *testing.T) {
	s := versioning.Spec{Raw: "20"}
	assert.True(t, s.Matches("20.1.0"), "prefix match")
	assert.True(t, s.Matches("20.11.3"), "prefix match with double digit minor")
	assert.True(t, s.Matches("20"), "exact match on prefix")
	assert.False(t, s.Matches("200.0"), "must not match 200")
	assert.False(t, s.Matches("19.9.9"), "earlier major")
	assert.False(t, s.Matches("21.0.0"), "later major")
}

func TestSpec_Matches_MinorPrefix(t *testing.T) {
	s := versioning.Spec{Raw: "3.12"}
	assert.True(t, s.Matches("3.12"))
	assert.True(t, s.Matches("3.12.1"))
	assert.True(t, s.Matches("3.12.10"))
	assert.False(t, s.Matches("3.120.0"), "must not match 3.120")
	assert.False(t, s.Matches("3.11.9"))
	assert.False(t, s.Matches("3.13.0"))
}

func TestSpec_Matches_PatchMatch(t *testing.T) {
	s := versioning.Spec{Raw: "1.2.3"}
	assert.True(t, s.Matches("1.2.3"))
	assert.False(t, s.Matches("1.2.30"))
	assert.False(t, s.Matches("1.2.3.4"))
}
