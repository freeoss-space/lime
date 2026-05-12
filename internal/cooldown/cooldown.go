// Package cooldown provides duration parsing for version cooldown windows.
// A cooldown window prevents installation of package versions newer than a
// configured age, reducing risk from freshly-released (potentially broken) software.
package cooldown

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Duration represents a cooldown window. The zero value means no cooldown.
type Duration struct {
	d time.Duration
}

// Zero is the zero-cooldown sentinel meaning "no cooldown — install any version".
var Zero = Duration{}

var durationRe = regexp.MustCompile(`^(\d+)(h|d|w)$`)

// Parse parses a cooldown string. Supported formats:
//
//	"0", "0d", "0h", "0w"  → no cooldown (zero)
//	"24h"                  → 24 hours
//	"14d"                  → 14 days
//	"2w"                   → 2 weeks (14 days)
//
// An empty string is treated as no cooldown.
func Parse(s string) (Duration, error) {
	if s == "" || s == "0" {
		return Zero, nil
	}
	m := durationRe.FindStringSubmatch(s)
	if m == nil {
		return Zero, fmt.Errorf("invalid cooldown %q: use formats like 14d, 2w, 24h", s)
	}
	n, _ := strconv.Atoi(m[1])
	if n == 0 {
		return Zero, nil
	}
	var d time.Duration
	switch m[2] {
	case "h":
		d = time.Duration(n) * time.Hour
	case "d":
		d = time.Duration(n) * 24 * time.Hour
	case "w":
		d = time.Duration(n) * 7 * 24 * time.Hour
	}
	return Duration{d: d}, nil
}

// IsZero reports whether this represents no cooldown.
func (c Duration) IsZero() bool { return c.d == 0 }

// D returns the underlying time.Duration.
func (c Duration) D() time.Duration { return c.d }

// Blocks reports whether a release at t should be blocked given the current time now.
// Always returns false when the cooldown is zero.
func (c Duration) Blocks(t, now time.Time) bool {
	if c.IsZero() {
		return false
	}
	return now.Sub(t) < c.d
}

// String returns a human-readable representation such as "14d", "2w", "24h", or "0d".
func (c Duration) String() string {
	if c.IsZero() {
		return "0d"
	}
	totalHours := int(c.d.Hours())
	if totalHours%168 == 0 {
		return fmt.Sprintf("%dw", totalHours/168)
	}
	if totalHours%24 == 0 {
		return fmt.Sprintf("%dd", totalHours/24)
	}
	return fmt.Sprintf("%dh", totalHours)
}
