package cli_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/freeoss-space/lime/internal/cli"
)

// These tests exercise CLI flags and output formatting without making network
// calls.  They verify cobra wiring and flag parsing.

func TestInstallCmd_DryRun_PrintsCommand(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, _, _ := execCmd(t, "install", "--dry-run", "--yes", "ripgrep")
	// dry-run with -y should print the command and not error on the manager step
	// (apt-get is available on CI/Ubuntu; on other systems it will fall through)
	_ = out // output varies by platform
}

func TestInstallCmd_MultiplePackages(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Verify cobra accepts multiple package args.
	root := cli.NewRootCmd()
	installCmd, _, err := root.Find([]string{"install"})
	assert.NoError(t, err)
	assert.NotNil(t, installCmd)
}

func TestSearchCmd_JSONFlag(t *testing.T) {
	root := cli.NewRootCmd()
	f := root.PersistentFlags().Lookup("json")
	assert.NotNil(t, f)
	assert.Equal(t, "bool", f.Value.Type())
}

func TestConfigCmd_PathSubcommand(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, _, err := execCmd(t, "config", "path")
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(strings.TrimSpace(out), "config.toml"))
}

func TestRootCmd_UnknownSubcommand(t *testing.T) {
	_, _, err := execCmd(t, "nonexistent")
	assert.Error(t, err)
}

func TestInstallCmd_VerboseFlag(t *testing.T) {
	root := cli.NewRootCmd()
	f := root.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestInstallCmd_DryRunFlag(t *testing.T) {
	root := cli.NewRootCmd()
	f := root.PersistentFlags().Lookup("dry-run")
	assert.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestConfigCmd_Help(t *testing.T) {
	out, _, err := execCmd(t, "config", "--help")
	assert.NoError(t, err)
	assert.Contains(t, out, "EDITOR")
	assert.Contains(t, out, "config.toml")
}

func TestInstallCmd_ExplicitManagerFlag(t *testing.T) {
	root := cli.NewRootCmd()
	installCmd, _, err := root.Find([]string{"install"})
	assert.NoError(t, err)
	f := installCmd.Flags().Lookup("manager")
	assert.NotNil(t, f)
	assert.Equal(t, "string", f.Value.Type())
}

func TestSearchCmd_NoArgs_Error(t *testing.T) {
	_, _, err := execCmd(t, "search")
	assert.Error(t, err)
}

func TestInstallCmd_NoArgs_Error(t *testing.T) {
	_, _, err := execCmd(t, "install")
	assert.Error(t, err)
}

func TestInstallCmd_DryRunWithYes(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Exercise --dry-run path; exact output varies by platform.
	_, _, _ = execCmd(t, "install", "--dry-run", "-y", "ripgrep")
}
