// Package search orchestrates package search across Repology and local managers.
package search

import (
	"context"
	"sort"

	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
)

// Result is one search hit, combining Repology metadata with local availability.
type Result struct {
	// ProjectName is the canonical Repology project name.
	ProjectName string
	// Manager is the package manager that provides this package.
	Manager string
	// Package is the package name as known to the manager.
	Package string
	// Version is the packaged version, if known.
	Version string
	// Repo is the Repology repository identifier.
	Repo string
	// Preferred indicates this manager is in the user's preference list.
	Preferred bool
	// Available indicates the manager binary exists on this system.
	Available bool
}

// RepologySearcher is the subset of repology.Client used by the searcher.
type RepologySearcher interface {
	SearchProjects(ctx context.Context, name string) (repology.ProjectPackages, error)
	GetProject(ctx context.Context, name string) ([]repology.Package, error)
}

// Options controls Searcher behaviour.
type Options struct {
	// PreferredManagers is the ordered preference list from config.
	PreferredManagers []string
}

// Searcher queries Repology and annotates results with local availability.
type Searcher struct {
	registry *managers.Registry
	repology RepologySearcher
	opts     Options
}

// New creates a Searcher.
func New(reg *managers.Registry, rep RepologySearcher, opts Options) *Searcher {
	return &Searcher{registry: reg, repology: rep, opts: opts}
}

// Search queries Repology for query and returns annotated results.
func (s *Searcher) Search(ctx context.Context, query string) ([]Result, error) {
	projects, err := s.repology.SearchProjects(ctx, query)
	if err != nil {
		return nil, err
	}

	available := s.availableSet(ctx)
	preferred := s.preferredSet()

	var results []Result
	for projectName, pkgs := range projects {
		for _, p := range pkgs {
			mgr := repology.ManagerForRepo(p.Repo)
			if mgr == "" {
				continue
			}
			results = append(results, Result{
				ProjectName: projectName,
				Manager:     mgr,
				Package:     p.EffectiveName(projectName),
				Version:     p.Version,
				Repo:        p.Repo,
				Preferred:   preferred[mgr],
				Available:   available[mgr],
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		// Preferred first, then available, then alphabetical.
		ri, rj := results[i], results[j]
		if ri.Preferred != rj.Preferred {
			return ri.Preferred
		}
		if ri.Available != rj.Available {
			return ri.Available
		}
		if ri.ProjectName != rj.ProjectName {
			return ri.ProjectName < rj.ProjectName
		}
		return ri.Manager < rj.Manager
	})

	return results, nil
}

// GetProject returns search results for an exact project name.
func (s *Searcher) GetProject(ctx context.Context, name string) ([]Result, error) {
	pkgs, err := s.repology.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	available := s.availableSet(ctx)
	preferred := s.preferredSet()

	var results []Result
	for _, p := range pkgs {
		mgr := repology.ManagerForRepo(p.Repo)
		if mgr == "" {
			continue
		}
		results = append(results, Result{
			ProjectName: name,
			Manager:     mgr,
			Package:     p.EffectiveName(name),
			Version:     p.Version,
			Repo:        p.Repo,
			Preferred:   preferred[mgr],
			Available:   available[mgr],
		})
	}

	sort.Slice(results, func(i, j int) bool {
		ri, rj := results[i], results[j]
		if ri.Preferred != rj.Preferred {
			return ri.Preferred
		}
		if ri.Available != rj.Available {
			return ri.Available
		}
		return ri.Manager < rj.Manager
	})

	return results, nil
}

func (s *Searcher) availableSet(ctx context.Context) map[string]bool {
	avail := s.registry.Available(ctx)
	m := make(map[string]bool, len(avail))
	for _, mgr := range avail {
		m[mgr.Name()] = true
	}
	return m
}

func (s *Searcher) preferredSet() map[string]bool {
	m := make(map[string]bool, len(s.opts.PreferredManagers))
	for _, p := range s.opts.PreferredManagers {
		m[p] = true
	}
	return m
}
