package managers

import "context"

// Asdf implements PackageManager for asdf (multi-language runtime version manager).
type Asdf struct{ base }

// NewAsdf creates an Asdf manager using the given Commander.
func NewAsdf(cmd Commander) *Asdf {
	return &Asdf{base{name: "asdf", binary: "asdf", cmd: cmd}}
}

func (a *Asdf) InstallArgs(pkg string) (string, []string) {
	return "asdf", []string{"install", pkg, "latest"}
}

func (a *Asdf) Install(ctx context.Context, pkg string) error {
	bin, args := a.InstallArgs(pkg)
	return a.cmd.Run(ctx, bin, args...)
}

// Search returns empty results; asdf has no native search command.
func (a *Asdf) Search(_ context.Context, _ string) ([]SearchResult, error) {
	return nil, nil
}

func (a *Asdf) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the asdf command for a specific version.
// asdf takes the version as a positional argument: "asdf install <plugin> <version>".
func (a *Asdf) InstallVersionArgs(pkg, version string) (string, []string) {
	return "asdf", []string{"install", pkg, version}
}

func (a *Asdf) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := a.InstallVersionArgs(pkg, version)
	return a.cmd.Run(ctx, bin, args...)
}
