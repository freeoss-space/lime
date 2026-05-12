// Package install orchestrates package installation across managers.
package install

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
)

// RepologyClient is the subset of repology.Client used by the installer.
type RepologyClient interface {
	GetProject(ctx context.Context, name string) ([]repology.Package, error)
}

// Options controls Installer behaviour.
type Options struct {
	// DryRun prevents actual installation; only shows what would run.
	DryRun bool
	// AutoConfirm skips the interactive prompt.
	AutoConfirm bool
	// PreferredManagers is the ordered preference list from config.
	PreferredManagers []string
	// Stdout is where prompts and status messages are written.
	Stdout io.Writer
	// Stdin is read for interactive confirmation.
	Stdin io.Reader
}

// Result describes the outcome of installing one package.
type Result struct {
	Spec    managers.PackageSpec
	Manager string
	Command string
	Skipped bool
	Err     error
}

// Installer resolves and installs packages via the best available manager.
type Installer struct {
	registry *managers.Registry
	repology RepologyClient
	opts     Options
}

// New creates an Installer.
func New(reg *managers.Registry, rep RepologyClient, opts Options) *Installer {
	return &Installer{
		registry: reg,
		repology: rep,
		opts:     opts,
	}
}

// Install installs all specs, returning one Result per spec.
func (ins *Installer) Install(ctx context.Context, specs []managers.PackageSpec) []Result {
	results := make([]Result, 0, len(specs))
	for _, spec := range specs {
		results = append(results, ins.installOne(ctx, spec))
	}
	return results
}

func (ins *Installer) installOne(ctx context.Context, spec managers.PackageSpec) Result {
	res := Result{Spec: spec}

	// If an explicit manager was requested, use it directly.
	if spec.Manager != "" {
		m := ins.registry.ByName(spec.Manager)
		if m == nil {
			res.Err = fmt.Errorf("manager %q not recognised", spec.Manager)
			return res
		}
		if !m.IsAvailable(ctx) {
			res.Err = fmt.Errorf("manager %q is not available on this system", spec.Manager)
			return res
		}
		return ins.runInstall(ctx, m, spec.Name)
	}

	// Otherwise resolve via Repology, then fall back through available managers.
	candidates, err := ins.resolve(ctx, spec.Name)
	if err != nil {
		// Repology unavailable — fall through to best available manager.
		candidates = nil
	}

	ordered := ins.registry.Preferred(ctx, ins.opts.PreferredManagers)
	if len(ordered) == 0 {
		res.Err = fmt.Errorf("no package managers are available on this system")
		return res
	}

	for _, m := range ordered {
		pkgName := spec.Name
		if mapped, ok := candidates[m.Name()]; ok {
			pkgName = mapped
		}
		r := ins.runInstall(ctx, m, pkgName)
		if r.Err == nil {
			return r
		}
	}

	res.Err = fmt.Errorf("could not install %q with any available manager", spec.Name)
	return res
}

// resolve queries Repology to find the package name for each manager.
func (ins *Installer) resolve(ctx context.Context, name string) (map[string]string, error) {
	pkgs, err := ins.repology.GetProject(ctx, name)
	if err != nil {
		return nil, err
	}

	out := make(map[string]string)
	for _, p := range pkgs {
		mgr := repology.ManagerForRepo(p.Repo)
		if mgr == "" {
			continue
		}
		if _, already := out[mgr]; !already {
			out[mgr] = p.EffectiveName(name)
		}
	}
	return out, nil
}

// runInstall prompts (unless -y/dry-run), then executes the install.
func (ins *Installer) runInstall(ctx context.Context, m managers.PackageManager, pkg string) Result {
	bin, args := m.InstallArgs(pkg)
	cmdStr := bin + " " + strings.Join(args, " ")

	res := Result{
		Spec:    managers.PackageSpec{Name: pkg},
		Manager: m.Name(),
		Command: cmdStr,
	}

	if !ins.opts.AutoConfirm && !ins.opts.DryRun {
		if !ins.confirm(m.Name(), pkg, cmdStr) {
			res.Skipped = true
			return res
		}
	}

	if ins.opts.DryRun {
		fmt.Fprintf(ins.opts.Stdout, "[dry-run] would run: %s\n", cmdStr)
		return res
	}

	res.Err = m.Install(ctx, pkg)
	return res
}

// confirm shows the install prompt and reads user input.
func (ins *Installer) confirm(managerName, pkg, cmdStr string) bool {
	fmt.Fprintf(ins.opts.Stdout, "\nInstall package?\n\n")
	fmt.Fprintf(ins.opts.Stdout, "  Manager : %s\n", managerName)
	fmt.Fprintf(ins.opts.Stdout, "  Package : %s\n", pkg)
	fmt.Fprintf(ins.opts.Stdout, "  Command : %s\n\n", cmdStr)
	fmt.Fprintf(ins.opts.Stdout, "[y/N] ")

	var input string
	if ins.opts.Stdin != nil {
		buf := make([]byte, 64)
		n, _ := ins.opts.Stdin.Read(buf)
		input = strings.TrimSpace(string(buf[:n]))
	}
	return strings.ToLower(input) == "y"
}
