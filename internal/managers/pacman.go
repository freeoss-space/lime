package managers

import (
	"context"
	"strings"
)

// Pacman implements PackageManager for pacman (Arch Linux).
type Pacman struct{ base }

// NewPacman creates a Pacman manager using the given Commander.
func NewPacman(cmd Commander) *Pacman {
	return &Pacman{base{name: "pacman", binary: "pacman", cmd: cmd}}
}

func (p *Pacman) InstallArgs(pkg string) (string, []string) {
	return "sudo", []string{"pacman", "-S", "--noconfirm", pkg}
}

func (p *Pacman) Install(ctx context.Context, pkg string) error {
	bin, args := p.InstallArgs(pkg)
	return p.cmd.Run(ctx, bin, args...)
}

func (p *Pacman) Search(ctx context.Context, query string) ([]SearchResult, error) {
	lines, err := p.searchLines(ctx, "-Ss", query)
	if err != nil {
		return nil, err
	}
	var results []SearchResult
	// pacman -Ss output alternates: "repo/name version [flags]" then "    description"
	for _, l := range lines {
		if strings.HasPrefix(l, " ") || strings.HasPrefix(l, "\t") {
			continue
		}
		parts := strings.Fields(l)
		if len(parts) < 2 {
			continue
		}
		// parts[0] = "repo/name", parts[1] = version
		repoPkg := strings.SplitN(parts[0], "/", 2)
		if len(repoPkg) != 2 {
			continue
		}
		results = append(results, SearchResult{
			Manager: p.name,
			Package: repoPkg[1],
			Version: parts[1],
			Repo:    repoPkg[0],
		})
	}
	return results, nil
}
