package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

const (
	DefaultRateLimitPerSecond = float64(1)
	DefaultHTTPTimeoutSeconds = 10
	appName                   = "lime"
)

// Config holds all lime configuration.
type Config struct {
	PreferredManagers  []string `toml:"preferred_managers"`
	AutoConfirm        bool     `toml:"auto_confirm"`
	RateLimitPerSecond float64  `toml:"rate_limit_per_second"`
	HTTPTimeoutSeconds int      `toml:"http_timeout_seconds"`
	// DefaultCooldown is the global cooldown duration applied to all installs when
	// no manager-specific cooldown is configured. Empty string means no cooldown.
	// Supported formats: "14d", "2w", "24h", "0d".
	DefaultCooldown string `toml:"default_cooldown"`
	// ManagerCooldowns maps manager names to cooldown durations, overriding
	// DefaultCooldown for specific managers. Example: {"brew": "7d", "apt": "0d"}.
	ManagerCooldowns map[string]string `toml:"manager_cooldowns"`
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		PreferredManagers: []string{
			"brew", "apt", "dnf", "pacman", "zypper",
			"apk", "pkg", "winget", "choco", "scoop",
		},
		AutoConfirm:        false,
		RateLimitPerSecond: DefaultRateLimitPerSecond,
		HTTPTimeoutSeconds: DefaultHTTPTimeoutSeconds,
		DefaultCooldown:    "",
		ManagerCooldowns:   nil,
	}
}

// Load reads the config file from the XDG config path, returning defaults
// if the file does not exist.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return Default(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := Default()
	if _, err := toml.Decode(string(data), cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Path returns the resolved path to the config file.
func Path() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, "config.toml"), nil
}

// EnsureExists creates the config file with defaults if it does not exist.
// Returns the path to the config file.
func EnsureExists() (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}

	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return "", fmt.Errorf("creating config directory: %w", err)
		}
		if err := Write(path, Default()); err != nil {
			return "", err
		}
	}
	return path, nil
}

// Write serialises cfg to path in TOML format.
func Write(path string, cfg *Config) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating config file: %w", err)
	}
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		_ = f.Close()
		return fmt.Errorf("encoding config: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("closing config file: %w", err)
	}
	return nil
}

// configDir returns the OS-appropriate config base directory following the
// XDG Base Directory specification on Linux/macOS and AppData on Windows.
func configDir() (string, error) {
	if runtime.GOOS == "windows" {
		return windowsConfigDir()
	}
	return xdgConfigDir()
}

func xdgConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config"), nil
}

func windowsConfigDir() (string, error) {
	if appData := os.Getenv("APPDATA"); appData != "" {
		return appData, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, "AppData", "Roaming"), nil
}
