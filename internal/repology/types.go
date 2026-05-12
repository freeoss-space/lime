package repology

import "time"

// Package represents one package entry returned by the Repology API.
// Multiple entries may share the same project name across repositories.
type Package struct {
	// Repo is the Repology repository identifier (e.g. "debian_stable", "homebrew").
	Repo string `json:"repo"`
	// Subrepo is the sub-repository (e.g. "main", "contrib"), if any.
	Subrepo string `json:"subrepo,omitempty"`
	// Name is the package name in this repository.
	Name string `json:"name,omitempty"`
	// SrcName is the source package name.
	SrcName string `json:"srcname,omitempty"`
	// BinNames lists binary package names produced by the source.
	BinNames []string `json:"binnames,omitempty"`
	// Version is the packaged version string.
	Version string `json:"version"`
	// OrigVersion is the upstream version before packaging suffix.
	OrigVersion string `json:"origversion,omitempty"`
	// Status is the freshness status: "newest", "outdated", "devel", etc.
	Status string `json:"status,omitempty"`
	// Families lists the repository families this package belongs to.
	Families []string `json:"families,omitempty"`
	// UpdatedAt is the timestamp of the last version change in Repology, when
	// provided by the API. Used for cooldown filtering. Nil when unavailable.
	UpdatedAt *time.Time `json:"versionupdated,omitempty"`
}

// EffectiveName returns the most specific package name available.
func (p Package) EffectiveName(projectName string) string {
	if p.Name != "" {
		return p.Name
	}
	if p.SrcName != "" {
		return p.SrcName
	}
	return projectName
}

// ProjectPackages maps project names to their packages (used by /projects/ endpoint).
type ProjectPackages map[string][]Package

// repoFamilyMap maps Repology repository prefixes/identifiers to our manager names.
// This is used to map Repology repos to concrete package manager names.
var repoFamilyMap = map[string]string{
	"debian":     "apt",
	"ubuntu":     "apt",
	"homebrew":   "brew",
	"fedora":     "dnf",
	"centos":     "dnf",
	"rhel":       "dnf",
	"arch":       "pacman",
	"openSUSE":   "zypper",
	"opensuse":   "zypper",
	"alpine":     "apk",
	"freebsd":    "pkg",
	"winget":     "winget",
	"chocolatey": "choco",
	"scoop":      "scoop",
	"pypi":       "pip",
	"crates_io":  "cargo",
	"npmjs":      "npm",
}

// ManagerForRepo returns the local package manager name for a Repology repo
// identifier, or an empty string if unrecognised.
func ManagerForRepo(repo string) string {
	// Exact match first.
	if m, ok := repoFamilyMap[repo]; ok {
		return m
	}
	// Prefix match (e.g. "debian_stable" → "debian" → "apt").
	for prefix, manager := range repoFamilyMap {
		if len(repo) >= len(prefix) && repo[:len(prefix)] == prefix {
			return manager
		}
	}
	return ""
}
