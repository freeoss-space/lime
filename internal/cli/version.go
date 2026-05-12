package cli

// Build-time variables injected via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// SetVersion injects build-time version information (called from main).
func SetVersion(v, c, d string) {
	if v != "" {
		version = v
	}
	if c != "" {
		commit = c
	}
	if d != "" {
		date = d
	}
}
