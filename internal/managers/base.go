package managers

import (
	"context"
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
