package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/cli"
)

// execCmd runs the root command with args, capturing stdout/stderr.
func execCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	root := cli.NewRootCmd()
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs(args)
	err = root.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestRootCmd_Help(t *testing.T) {
	out, _, err := execCmd(t, "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "lime")
	assert.Contains(t, out, "install")
	assert.Contains(t, out, "search")
	assert.Contains(t, out, "config")
}

func TestRootCmd_Version(t *testing.T) {
	out, _, err := execCmd(t, "--version")
	require.NoError(t, err)
	assert.Contains(t, out, "lime")
}

func TestInstallCmd_Help(t *testing.T) {
	out, _, err := execCmd(t, "install", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "install")
	assert.Contains(t, out, "--yes")
	assert.Contains(t, out, "--dry-run")
}

func TestInstallCmd_RequiresAtLeastOneArg(t *testing.T) {
	_, _, err := execCmd(t, "install")
	assert.Error(t, err)
}

func TestSearchCmd_Help(t *testing.T) {
	out, _, err := execCmd(t, "search", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "search")
	assert.Contains(t, strings.ToLower(out), "repology")
}

func TestSearchCmd_RequiresExactlyOneArg(t *testing.T) {
	_, _, err := execCmd(t, "search")
	assert.Error(t, err)
}

func TestConfigPathCmd(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	out, _, err := execCmd(t, "config", "path")
	require.NoError(t, err)
	assert.Contains(t, out, "lime")
	assert.Contains(t, out, "config.toml")
}

func TestGlobalFlags_DryRunParsed(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	// Ensure --dry-run is a valid global flag (flag parsing must not error).
	// RunE may fail due to network; we only assert no panic.
	root := cli.NewRootCmd()
	root.SetArgs([]string{"install", "--dry-run", "--yes", "ripgrep"})
	_ = root.Execute()
}

func TestGlobalFlags_JSONFlagExists(t *testing.T) {
	root := cli.NewRootCmd()
	f := root.PersistentFlags().Lookup("json")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

func TestGlobalFlags_VerboseFlagExists(t *testing.T) {
	root := cli.NewRootCmd()
	f := root.PersistentFlags().Lookup("verbose")
	require.NotNil(t, f)
}

func TestInstallCmd_YesFlagShorthand(t *testing.T) {
	// -y shorthand should be registered
	root := cli.NewRootCmd()
	installCmd, _, err := root.Find([]string{"install"})
	require.NoError(t, err)
	f := installCmd.Flags().ShorthandLookup("y")
	require.NotNil(t, f, "-y flag should exist on install command")
}

func TestInstallCmd_ManagerFlag(t *testing.T) {
	root := cli.NewRootCmd()
	installCmd, _, err := root.Find([]string{"install"})
	require.NoError(t, err)
	f := installCmd.Flags().Lookup("manager")
	require.NotNil(t, f)
}

func TestInstallCmd_CooldownFlag(t *testing.T) {
	root := cli.NewRootCmd()
	installCmd, _, err := root.Find([]string{"install"})
	require.NoError(t, err)
	f := installCmd.Flags().Lookup("cooldown")
	require.NotNil(t, f, "--cooldown flag should exist on install command")
	assert.Equal(t, "", f.DefValue, "default cooldown should be empty (no cooldown)")
}

func TestInstallCmd_HelpMentionsVersionSyntax(t *testing.T) {
	out, _, err := execCmd(t, "install", "--help")
	require.NoError(t, err)
	assert.Contains(t, out, "@", "help should mention @version syntax")
	assert.Contains(t, strings.ToLower(out), "cooldown")
}

func TestInstallCmd_HelpMentionsCooldownUnits(t *testing.T) {
	out, _, err := execCmd(t, "install", "--help")
	require.NoError(t, err)
	// Help should mention supported cooldown units
	assert.Contains(t, out, "14d")
}

func TestSearchCmd_HelpMentionsAge(t *testing.T) {
	out, _, err := execCmd(t, "search", "--help")
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(out), "age")
}
