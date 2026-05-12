// Package versioning provides version specification parsing and matching
// for the lime package installer.
package versioning

import "strings"

// Spec represents a version requirement parsed from user input such as "pkg@14.1.1".
// The zero value (empty Spec) means "any version".
type Spec struct {
	// Raw is the original version string provided by the user (e.g. "14.1.1", "20", "3.12").
	Raw string
}

// ParsePackageArg parses the "pkg@version" syntax used on the CLI.
// Returns (packageName, Spec). When no "@" separator is present, Spec is the
// zero value meaning "any version".
//
// Examples:
//
//	"ripgrep"        → ("ripgrep", Spec{})
//	"ripgrep@14.1.1" → ("ripgrep", Spec{Raw: "14.1.1"})
//	"nodejs@20"      → ("nodejs",  Spec{Raw: "20"})
func ParsePackageArg(arg string) (string, Spec) {
	idx := strings.LastIndex(arg, "@")
	if idx <= 0 {
		return arg, Spec{}
	}
	ver := arg[idx+1:]
	if ver == "" {
		return arg, Spec{}
	}
	return arg[:idx], Spec{Raw: ver}
}

// IsAny reports whether this spec accepts any version (zero value).
func (s Spec) IsAny() bool { return s.Raw == "" }

// Matches reports whether version v satisfies this spec.
//
// Matching rules:
//   - Empty spec: matches everything.
//   - Exact match: always checked first.
//   - Prefix match: applied when the spec has fewer than 2 dots (i.e. it is a
//     major or major.minor version). Specs with 2+ dots (e.g. "14.1.1") are
//     treated as exact to avoid ambiguity with 4-part version numbers.
//
// Examples:
//
//	Spec{Raw: "20"}     matches "20", "20.1.0", "20.11.3"
//	Spec{Raw: "3.12"}   matches "3.12", "3.12.1", "3.12.10"
//	Spec{Raw: "14.1.1"} matches only "14.1.1" (exact)
func (s Spec) Matches(v string) bool {
	if s.IsAny() {
		return true
	}
	if v == s.Raw {
		return true
	}
	// Only apply prefix matching for major (e.g. "20") or major.minor (e.g. "3.12") specs.
	// Patch-level specs ("14.1.1") require an exact match.
	if strings.Count(s.Raw, ".") >= 2 {
		return false
	}
	return strings.HasPrefix(v, s.Raw+".")
}
