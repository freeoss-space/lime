package managers

import (
	"context"
)

// Registry holds all known package managers and provides filtering helpers.
type Registry struct {
	all []PackageManager
}

// NewRegistry creates a Registry containing the given managers.
func NewRegistry(managers ...PackageManager) *Registry {
	return &Registry{all: managers}
}

// All returns every registered manager.
func (r *Registry) All() []PackageManager {
	return r.all
}

// Available returns managers whose binary exists on the current system.
func (r *Registry) Available(ctx context.Context) []PackageManager {
	var out []PackageManager
	for _, m := range r.all {
		if m.IsAvailable(ctx) {
			out = append(out, m)
		}
	}
	return out
}

// ByName returns the manager with the given name, or nil.
func (r *Registry) ByName(name string) PackageManager {
	for _, m := range r.all {
		if m.Name() == name {
			return m
		}
	}
	return nil
}

// Preferred returns available managers ordered by the preference list.
// Managers not in prefs appear after preferred ones, in registration order.
func (r *Registry) Preferred(ctx context.Context, prefs []string) []PackageManager {
	avail := r.Available(ctx)
	idx := make(map[string]int, len(prefs))
	for i, p := range prefs {
		idx[p] = i
	}

	preferred := make([]PackageManager, 0, len(avail))
	rest := make([]PackageManager, 0, len(avail))

	for _, m := range avail {
		if _, ok := idx[m.Name()]; ok {
			preferred = append(preferred, m)
		} else {
			rest = append(rest, m)
		}
	}

	sortByPreference(preferred, idx)

	return append(preferred, rest...)
}

// DefaultRegistry creates the full set of managers backed by RealCommander.
func DefaultRegistry() *Registry {
	cmd := RealCommander{}
	return NewRegistry(
		NewBrew(cmd),
		NewBrewCask(cmd),
		NewApt(cmd),
		NewDnf(cmd),
		NewPacman(cmd),
		NewZypper(cmd),
		NewApk(cmd),
		NewPkg(cmd),
		NewWinget(cmd),
		NewChoco(cmd),
		NewScoop(cmd),
		NewPip(cmd),
		NewUv(cmd),
		NewCargo(cmd),
		NewNpm(cmd),
		NewGoInstall(cmd),
	)
}

func sortByPreference(ms []PackageManager, idx map[string]int) {
	// Insertion sort — small slice, stable order matters.
	for i := 1; i < len(ms); i++ {
		for j := i; j > 0 && idx[ms[j].Name()] < idx[ms[j-1].Name()]; j-- {
			ms[j], ms[j-1] = ms[j-1], ms[j]
		}
	}
}
