package managers

import (
	"context"
)

// GoInstall implements PackageManager for "go install" (Go toolchain).
type GoInstall struct{ base }

// NewGoInstall creates a GoInstall manager using the given Commander.
func NewGoInstall(cmd Commander) *GoInstall {
	return &GoInstall{base{name: "go", binary: "go", cmd: cmd}}
}

func (g *GoInstall) InstallArgs(pkg string) (string, []string) {
	return "go", []string{"install", pkg + "@latest"}
}

func (g *GoInstall) Install(ctx context.Context, pkg string) error {
	bin, args := g.InstallArgs(pkg)
	return g.cmd.Run(ctx, bin, args...)
}

// Search returns no results: the Go toolchain has no native package search command.
func (g *GoInstall) Search(_ context.Context, _ string) ([]SearchResult, error) {
	return nil, nil
}

func (g *GoInstall) SupportsVersioning() bool { return true }

// InstallVersionArgs returns the go install command for a specific version.
// Go uses the "pkg@version" module path syntax (e.g. "pkg@v1.2.3" or "pkg@latest").
func (g *GoInstall) InstallVersionArgs(pkg, version string) (string, []string) {
	return "go", []string{"install", pkg + "@" + version}
}

func (g *GoInstall) InstallVersion(ctx context.Context, pkg, version string) error {
	bin, args := g.InstallVersionArgs(pkg, version)
	return g.cmd.Run(ctx, bin, args...)
}
