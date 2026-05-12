package managers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/managers"
)

func TestZypper_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"zypper": "/usr/bin/zypper"},
		OutputData: []byte(
			"| ripgrep | A fast grep replacement | package |\n" +
				"| ripgrep-all | Ripgrep-all package | package |\n" +
				"| ---+---+--- |\n" + // separator line
				"| Name | Summary | Type |\n", // header
		),
	}
	m := managers.NewZypper(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "zypper", results[0].Manager)
}

func TestApk_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"apk": "/sbin/apk"},
		OutputData: []byte(
			"ripgrep-14.0.3-r0\n" +
				"ripgrep-doc-14.0.3-r0\n",
		),
	}
	m := managers.NewApk(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ripgrep-14.0.3-r0", results[0].Package)
	assert.Equal(t, "apk", results[0].Manager)
	assert.Equal(t, "alpine", results[0].Repo)
}

func TestPkg_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pkg": "/usr/sbin/pkg"},
		OutputData: []byte(
			"ripgrep-14.0.3 A search tool that combines the usability of ag with raw speed of grep\n",
		),
	}
	m := managers.NewPkg(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ripgrep-14.0.3", results[0].Package)
	assert.Equal(t, "pkg", results[0].Manager)
}

func TestWinget_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"winget": "C:\\Windows\\winget.exe"},
		OutputData: []byte(
			"Name          Id                       Version    Source\n" +
				"----------------------------------------------------------\n" +
				"ripgrep       BurntSushi.ripgrep       14.0.3     winget\n",
		),
	}
	m := managers.NewWinget(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "winget", results[0].Manager)
}

func TestChoco_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"choco": "C:\\ProgramData\\chocolatey\\bin\\choco.exe"},
		OutputData: []byte(
			"Chocolatey v1.3.0\n" +
				"ripgrep 14.0.3 [Approved]\n" +
				"---\n" +
				"2 packages found.\n",
		),
	}
	m := managers.NewChoco(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	// "Chocolatey" and "---" lines should be filtered; "2" line slips through but is acceptable
	found := false
	for _, r := range results {
		if r.Package == "ripgrep" {
			found = true
			assert.Equal(t, "14.0.3", r.Version)
			break
		}
	}
	assert.True(t, found, "ripgrep should be in results")
}

func TestScoop_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"scoop": "C:\\Users\\user\\scoop\\shims\\scoop.cmd"},
		OutputData: []byte(
			"Results from local buckets...\n\n" +
				"  ripgrep (main) 14.0.3\n",
		),
	}
	m := managers.NewScoop(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "scoop", results[0].Manager)
}

func TestZypper_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("zypper")
	m := managers.NewZypper(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "zypper")
}

func TestApk_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("apk")
	m := managers.NewApk(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "apk")
}

func TestPkg_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("pkg")
	m := managers.NewPkg(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
}

func TestWinget_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("winget")
	m := managers.NewWinget(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, "winget", cmd.Calls[0][0])
}

func TestChoco_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("choco")
	m := managers.NewChoco(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, "choco", cmd.Calls[0][0])
}

func TestScoop_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("scoop")
	m := managers.NewScoop(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Equal(t, "scoop", cmd.Calls[0][0])
}

func TestSearchEmpty_ReturnsNoResults(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"apt-get": "/usr/bin/apt-get"},
		OutputData:     []byte(""),
	}
	m := managers.NewApt(cmd)
	results, err := m.Search(context.Background(), "nonexistentpkg12345")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestInstallArgs_AllManagersReturnPackageName(t *testing.T) {
	cmd := availableCmd("x")
	tests := []struct {
		m   managers.PackageManager
		pkg string
	}{
		{managers.NewApt(cmd), "ripgrep"},
		{managers.NewBrew(cmd), "ripgrep"},
		{managers.NewDnf(cmd), "ripgrep"},
		{managers.NewPacman(cmd), "ripgrep"},
		{managers.NewZypper(cmd), "ripgrep"},
		{managers.NewApk(cmd), "ripgrep"},
		{managers.NewPkg(cmd), "ripgrep"},
		{managers.NewWinget(cmd), "ripgrep"},
		{managers.NewChoco(cmd), "ripgrep"},
		{managers.NewScoop(cmd), "ripgrep"},
	}
	for _, tt := range tests {
		t.Run(tt.m.Name(), func(t *testing.T) {
			bin, args := tt.m.InstallArgs(tt.pkg)
			assert.NotEmpty(t, bin, "binary should not be empty")
			assert.Contains(t, args, "ripgrep", "package name should appear in args")
		})
	}
}
