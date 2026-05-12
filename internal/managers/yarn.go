package managers

import "context"

// Yarn implements PackageManager for yarn (alternative Node.js package manager).
type Yarn struct{ base }

// NewYarn creates a Yarn manager using the given Commander.
func NewYarn(cmd Commander) *Yarn {
	return &Yarn{base{name: "yarn", binary: "yarn", cmd: cmd}}
}

func (y *Yarn) InstallArgs(pkg string) (string, []string) {
	return "yarn", []string{"global", "add", pkg}
}

func (y *Yarn) Install(ctx context.Context, pkg string) error {
	bin, args := y.InstallArgs(pkg)
	return y.cmd.Run(ctx, bin, args...)
}

// Search returns empty results; yarn has no native search command.
func (y *Yarn) Search(_ context.Context, _ string) ([]SearchResult, error) {
	return nil, nil
}

func (y *Yarn) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the yarn command for a specific version.
// yarn uses "pkg@version" syntax.
func (y *Yarn) InstallVersionArgs(pkg, version string) (string, []string) {
	return "yarn", []string{"global", "add", pkg + "@" + version}
}

func (y *Yarn) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := y.InstallVersionArgs(pkg, version)
	return y.cmd.Run(ctx, bin, args...)
}
