package managers

import (
	"context"
	"fmt"
	"strings"
)

// base provides common manager behaviour. Embed it in concrete managers.
type base struct {
	name   string
	binary string
	cmd    Commander
}

func (b *base) Name() string { return b.name }

func (b *base) IsAvailable(_ context.Context) bool {
	_, err := b.cmd.LookPath(b.binary)
	return err == nil
}

// searchLines runs a command and splits stdout into non-empty lines.
func (b *base) searchLines(ctx context.Context, args ...string) ([]string, error) {
	out, err := b.cmd.Output(ctx, b.binary, args...)
	if err != nil && len(out) == 0 {
		return nil, err
	}
	var lines []string
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			lines = append(lines, l)
		}
	}
	return lines, nil
}

// SupportsVersioning returns false by default. Managers that support version-specific
// installation override this method.
func (b *base) SupportsVersioning() bool { return false }

// InstallVersionArgs is a placeholder; concrete managers override this.
// For managers that don't support versioning it is never called by the installer.
func (b *base) InstallVersionArgs(_, _ string) (string, []string) {
	return b.binary, nil
}

// InstallVersion returns an error for managers that do not support versioned installs.
// Managers that support versioning override this method.
func (b *base) InstallVersion(_ context.Context, _, _ string) error {
	return fmt.Errorf("manager %q does not support version-specific installation", b.name)
}
