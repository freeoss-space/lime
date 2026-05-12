// Package install orchestrates package installation across managers.
package install

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/freeoss-space/lime/internal/cooldown"
	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/internal/resolution"
	"github.com/freeoss-space/lime/internal/versioning"
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
	// Cooldown is a CLI-provided cooldown string (e.g. "14d") that overrides config.
	// Empty string means "use config values".
	Cooldown string
	// DefaultCooldown is the global cooldown from config (used when Cooldown is empty
	// and no manager-specific cooldown applies).
	DefaultCooldown string
	// ManagerCooldowns maps manager names to cooldown strings from config,
	// taking precedence over DefaultCooldown for specific managers.
	ManagerCooldowns map[string]string
}

// Result describes the outcome of installing one package.
type Result struct {
	Spec    managers.PackageSpec
	Manager string
	Command string
	// Version is the actual version installed (empty if unversioned install).
	Version string
	Skipped bool
	Err     error
}

// Installer resolves and installs packages via the best available manager.
type Installer struct {
	registry *managers.Registry
	repology RepologyClient
	resolver *resolution.Resolver
	opts     Options
}

// New creates an Installer.
func New(reg *managers.Registry, rep RepologyClient, opts Options) *Installer {
	return &Installer{
		registry: reg,
		repology: rep,
		resolver: resolution.New(),
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

	vspec := versioning.Spec{Raw: spec.Version}

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
		cd, err := ins.effectiveCooldown(spec.Manager)
		if err != nil {
			res.Err = err
			return res
		}
		// Try to use Repology to resolve the package name and check cooldown.
		pkgs, _ := ins.repology.GetProject(ctx, spec.Name)
		if candidate, err := ins.resolver.ForManager(pkgs, spec.Manager, vspec, cd, spec.Name); err != nil {
			res.Err = err
			return res
		} else if candidate != nil {
			return ins.runInstall(ctx, m, candidate.Package, candidate.Version, vspec)
		}
		// No Repology data — install with original name (no version checking possible).
		return ins.runInstall(ctx, m, spec.Name, spec.Version, vspec)
	}

	// Resolve via Repology, then fall back through available managers.
	pkgs, repErr := ins.repology.GetProject(ctx, spec.Name)
	if repErr != nil {
		// Repology unavailable — fall through to best available manager without cooldown.
		pkgs = nil
	}

	ordered := ins.registry.Preferred(ctx, ins.opts.PreferredManagers)
	if len(ordered) == 0 {
		res.Err = fmt.Errorf("no package managers are available on this system")
		return res
	}

	var lastCooldownErr error
	for _, m := range ordered {
		cd, err := ins.effectiveCooldown(m.Name())
		if err != nil {
			continue
		}
		if pkgs != nil {
			candidate, candErr := ins.resolver.ForManager(pkgs, m.Name(), vspec, cd, spec.Name)
			if candErr != nil {
				// Cooldown-blocked or version not found for this manager — try next.
				var ce *resolution.CooldownError
				if isErr(candErr, &ce) {
					lastCooldownErr = candErr
				}
				continue
			}
			if candidate != nil {
				r := ins.runInstall(ctx, m, candidate.Package, candidate.Version, vspec)
				if r.Err == nil {
					return r
				}
				// Install failed — try next manager.
				continue
			}
			// No package for this manager in Repology — still try with original name.
		}
		// No Repology data or manager not in Repology results — try with original name.
		if !vspec.IsAny() {
			// Version requested but no Repology data to validate — skip if we have
			// Repology data (means manager genuinely doesn't have the package).
			if pkgs != nil {
				continue
			}
		}
		r := ins.runInstall(ctx, m, spec.Name, spec.Version, vspec)
		if r.Err == nil {
			return r
		}
	}

	if lastCooldownErr != nil {
		res.Err = lastCooldownErr
		return res
	}
	res.Err = fmt.Errorf("could not install %q with any available manager", spec.Name)
	return res
}

// effectiveCooldown resolves the cooldown duration for a specific manager
// using the priority: CLI arg > manager-specific config > global config > none.
func (ins *Installer) effectiveCooldown(manager string) (cooldown.Duration, error) {
	if ins.opts.Cooldown != "" {
		return cooldown.Parse(ins.opts.Cooldown)
	}
	if ins.opts.ManagerCooldowns != nil {
		if s, ok := ins.opts.ManagerCooldowns[manager]; ok {
			return cooldown.Parse(s)
		}
	}
	if ins.opts.DefaultCooldown != "" {
		return cooldown.Parse(ins.opts.DefaultCooldown)
	}
	return cooldown.Zero, nil
}

// runInstall prompts (unless -y/dry-run), then executes the install.
func (ins *Installer) runInstall(ctx context.Context, m managers.PackageManager, pkg, version string, vspec versioning.Spec) Result {
	useVersion := !vspec.IsAny() && version != "" && m.SupportsVersioning()
	var bin string
	var args []string
	if useVersion {
		bin, args = m.InstallVersionArgs(pkg, version)
	} else {
		bin, args = m.InstallArgs(pkg)
	}
	cmdStr := bin + " " + strings.Join(args, " ")

	res := Result{
		Spec:    managers.PackageSpec{Name: pkg, Version: version},
		Manager: m.Name(),
		Command: cmdStr,
		Version: version,
	}

	if !ins.opts.AutoConfirm && !ins.opts.DryRun {
		if !ins.confirm(m.Name(), pkg, version, cmdStr) {
			res.Skipped = true
			return res
		}
	}

	if ins.opts.DryRun {
		fmt.Fprintf(ins.opts.Stdout, "[dry-run] would run: %s\n", cmdStr)
		return res
	}

	if useVersion {
		res.Err = m.InstallVersion(ctx, pkg, version)
	} else {
		res.Err = m.Install(ctx, pkg)
	}
	return res
}

// confirm shows the install prompt and reads user input.
func (ins *Installer) confirm(managerName, pkg, version, cmdStr string) bool {
	fmt.Fprintf(ins.opts.Stdout, "\nInstall package?\n\n")
	fmt.Fprintf(ins.opts.Stdout, "  Manager : %s\n", managerName)
	fmt.Fprintf(ins.opts.Stdout, "  Package : %s\n", pkg)
	if version != "" {
		fmt.Fprintf(ins.opts.Stdout, "  Version : %s\n", version)
	}
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

// isErr is a helper to check and unwrap typed errors without reflect.
func isErr(err error, target **resolution.CooldownError) bool {
	var ce *resolution.CooldownError
	if ok := asCooldownError(err, &ce); ok {
		*target = ce
		return true
	}
	return false
}

func asCooldownError(err error, target **resolution.CooldownError) bool {
	if ce, ok := err.(*resolution.CooldownError); ok {
		*target = ce
		return true
	}
	return false
}
