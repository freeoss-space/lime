package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/config"
)

// setConfigDir redirects the OS-appropriate config base directory to a temp
// directory for the duration of the test.  On Windows, APPDATA is used; on
// all other platforms, XDG_CONFIG_HOME is used.
func setConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", dir)
	} else {
		t.Setenv("XDG_CONFIG_HOME", dir)
	}
	return dir
}

func TestDefault_HasSaneValues(t *testing.T) {
	cfg := config.Default()
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.PreferredManagers)
	assert.Equal(t, float64(1), cfg.RateLimitPerSecond)
	assert.Equal(t, 10, cfg.HTTPTimeoutSeconds)
	assert.False(t, cfg.AutoConfirm)
}

func TestDefault_ContainsCommonManagers(t *testing.T) {
	cfg := config.Default()
	managers := make(map[string]bool, len(cfg.PreferredManagers))
	for _, m := range cfg.PreferredManagers {
		managers[m] = true
	}
	for _, expected := range []string{"apt", "brew", "dnf", "pacman"} {
		assert.True(t, managers[expected], "expected manager %q in defaults", expected)
	}
}

func TestLoad_ReturnDefaultsWhenNoFile(t *testing.T) {
	setConfigDir(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.PreferredManagers)
	assert.Equal(t, config.DefaultHTTPTimeoutSeconds, cfg.HTTPTimeoutSeconds)
}

func TestLoad_ParsesValidConfig(t *testing.T) {
	dir := setConfigDir(t)

	cfgPath := filepath.Join(dir, "jil", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))

	content := `
preferred_managers = ["apt", "brew"]
auto_confirm = true
rate_limit_per_second = 2.0
http_timeout_seconds = 30
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, []string{"apt", "brew"}, cfg.PreferredManagers)
	assert.True(t, cfg.AutoConfirm)
	assert.Equal(t, 2.0, cfg.RateLimitPerSecond)
	assert.Equal(t, 30, cfg.HTTPTimeoutSeconds)
}

func TestLoad_ReturnsErrorOnInvalidTOML(t *testing.T) {
	dir := setConfigDir(t)

	cfgPath := filepath.Join(dir, "jil", "config.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(cfgPath), 0o755))
	require.NoError(t, os.WriteFile(cfgPath, []byte("not = valid toml [[["), 0o644))

	_, err := config.Load()
	assert.Error(t, err)
}

func TestPath_UsesXDGEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("XDG_CONFIG_HOME is not the config mechanism on Windows")
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	path, err := config.Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "jil", "config.toml"), path)
}

func TestPath_UsesAPPDATAOnWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("APPDATA is only the config mechanism on Windows")
	}
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)

	path, err := config.Path()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "jil", "config.toml"), path)
}

func TestEnsureExists_CreatesDefaultConfig(t *testing.T) {
	setConfigDir(t)

	path, err := config.EnsureExists()
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	_, err = os.Stat(path)
	require.NoError(t, err, "config file should exist")

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.NotEmpty(t, cfg.PreferredManagers)
}

func TestEnsureExists_IdempotentWhenFileExists(t *testing.T) {
	setConfigDir(t)

	path1, err := config.EnsureExists()
	require.NoError(t, err)

	path2, err := config.EnsureExists()
	require.NoError(t, err)

	assert.Equal(t, path1, path2)
}

func TestWrite_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	original := &config.Config{
		PreferredManagers:  []string{"apt", "dnf"},
		AutoConfirm:        true,
		RateLimitPerSecond: 5.0,
		HTTPTimeoutSeconds: 20,
	}

	require.NoError(t, config.Write(path, original))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "apt")
	assert.Contains(t, string(data), "dnf")
}
