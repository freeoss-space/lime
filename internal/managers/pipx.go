package managers

import "context"

// Pipx implements PackageManager for pipx (Python CLI tools in isolated venvs).
type Pipx struct{ base }

// NewPipx creates a Pipx manager using the given Commander.
func NewPipx(cmd Commander) *Pipx {
	return &Pipx{base{name: "pipx", binary: "pipx", cmd: cmd}}
}

func (p *Pipx) InstallArgs(pkg string) (string, []string) {
	return "pipx", []string{"install", pkg}
}

func (p *Pipx) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

// Search returns empty results; pipx has no native search command.
func (p *Pipx) Search(_ context.Context, _ string) ([]SearchResult, error) {
	return nil, nil
}

func (p *Pipx) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the pipx command for a specific version.
// pipx uses PEP 508 "pkg==version" specifier syntax.
func (p *Pipx) InstallVersionArgs(pkg, version string) (string, []string) {
	return "pipx", []string{"install", pkg + "==" + version}
}

func (p *Pipx) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := p.InstallVersionArgs(pkg, version)
	return p.cmd.Run(ctx, bin, args...)
}
