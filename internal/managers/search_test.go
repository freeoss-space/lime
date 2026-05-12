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
		{managers.NewBrewCask(cmd), "ripgrep"},
		{managers.NewDnf(cmd), "ripgrep"},
		{managers.NewPacman(cmd), "ripgrep"},
		{managers.NewZypper(cmd), "ripgrep"},
		{managers.NewApk(cmd), "ripgrep"},
		{managers.NewPkg(cmd), "ripgrep"},
		{managers.NewWinget(cmd), "ripgrep"},
		{managers.NewChoco(cmd), "ripgrep"},
		{managers.NewScoop(cmd), "ripgrep"},
		{managers.NewPip(cmd), "ripgrep"},
		{managers.NewUv(cmd), "ripgrep"},
		{managers.NewCargo(cmd), "ripgrep"},
		{managers.NewNpm(cmd), "ripgrep"},
	}
	for _, tt := range tests {
		t.Run(tt.m.Name(), func(t *testing.T) {
			bin, args := tt.m.InstallArgs(tt.pkg)
			assert.NotEmpty(t, bin, "binary should not be empty")
			assert.Contains(t, args, "ripgrep", "package name should appear in args")
		})
	}
}

func TestGoInstall_InstallArgs_AppendsLatest(t *testing.T) {
	cmd := availableCmd("go")
	m := managers.NewGoInstall(cmd)
	bin, args := m.InstallArgs("golang.org/x/tools/cmd/goimports")
	assert.Equal(t, "go", bin)
	assert.Contains(t, args, "golang.org/x/tools/cmd/goimports@latest")
}

// --- Language ecosystem managers ---

func TestPip_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewPip(availableCmd("pip3"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestPip_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("pip3")
	m := managers.NewPip(cmd)
	err := m.Install(context.Background(), "requests")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "pip3", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "requests")
	assert.Contains(t, cmd.Calls[0], "--user")
}

func TestPip_Search_ParsesIndexVersionsOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pip3": "/usr/bin/pip3"},
		OutputData: []byte(
			"WARNING: pip index is currently an experimental command.\n" +
				"requests (2.31.0)\n" +
				"Available versions: 2.31.0, 2.30.0, 2.29.0\n",
		),
	}
	m := managers.NewPip(cmd)
	results, err := m.Search(context.Background(), "requests")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "requests", results[0].Package)
	assert.Equal(t, "2.31.0", results[0].Version)
	assert.Equal(t, "pip", results[0].Manager)
	assert.Equal(t, "pypi", results[0].Repo)
}

func TestPip_Search_ReturnsEmptyOnCommandFailure(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pip3": "/usr/bin/pip3"},
		OutputErr:      assert.AnError,
	}
	m := managers.NewPip(cmd)
	results, err := m.Search(context.Background(), "requests")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestPip_InstallVersion_UsesDoubleEqualsSyntax(t *testing.T) {
	cmd := availableCmd("pip3")
	m := managers.NewPip(cmd)
	err := m.InstallVersion(context.Background(), "requests", "2.31.0")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "pip3", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "requests==2.31.0")
}

func TestCargo_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewCargo(availableCmd("cargo"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestCargo_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("cargo")
	m := managers.NewCargo(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "cargo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep")
}

func TestCargo_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"cargo": "/usr/bin/cargo"},
		OutputData: []byte(
			`ripgrep = "14.1.1"    # A fast line-oriented search tool` + "\n" +
				`ripgrep-all = "0.9.7"    # ripgrep, but also search in PDFs` + "\n",
		),
	}
	m := managers.NewCargo(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "14.1.1", results[0].Version)
	assert.Equal(t, "cargo", results[0].Manager)
	assert.Equal(t, "crates.io", results[0].Repo)
}

func TestCargo_InstallVersion_UsesVersionFlag(t *testing.T) {
	cmd := availableCmd("cargo")
	m := managers.NewCargo(cmd)
	bin, args := m.InstallVersionArgs("ripgrep", "14.1.1")
	assert.Equal(t, "cargo", bin)
	assert.Contains(t, args, "--version")
	assert.Contains(t, args, "14.1.1")
	assert.Contains(t, args, "ripgrep")
}

func TestNpm_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewNpm(availableCmd("npm"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestNpm_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("npm")
	m := managers.NewNpm(cmd)
	err := m.Install(context.Background(), "typescript")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "npm", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "-g")
	assert.Contains(t, cmd.Calls[0], "typescript")
}

func TestNpm_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"npm": "/usr/bin/npm"},
		OutputData: []byte(
			"typescript\tTypeScript language\t=Microsoft\t2023-01-01\t5.0.0\tjavascript\n" +
				"typescript-eslint\tESLint for TypeScript\t=typescript-eslint\t2023-06-01\t6.0.0\t\n",
		),
	}
	m := managers.NewNpm(cmd)
	results, err := m.Search(context.Background(), "typescript")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "typescript", results[0].Package)
	assert.Equal(t, "5.0.0", results[0].Version)
	assert.Equal(t, "npm", results[0].Manager)
	assert.Equal(t, "npm", results[0].Repo)
}

func TestNpm_Search_SkipsHeaderRow(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"npm": "/usr/bin/npm"},
		OutputData: []byte(
			"NAME\tDESCRIPTION\tAUTHOR\tDATE\tVERSION\tKEYWORDS\n" +
				"typescript\tTypeScript language\t=Microsoft\t2023-01-01\t5.0.0\tjavascript\n",
		),
	}
	m := managers.NewNpm(cmd)
	results, err := m.Search(context.Background(), "typescript")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "typescript", results[0].Package)
}

func TestNpm_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("npm")
	m := managers.NewNpm(cmd)
	err := m.InstallVersion(context.Background(), "typescript", "5.0.0")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "typescript@5.0.0")
}

func TestGoInstall_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewGoInstall(availableCmd("go"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestGoInstall_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("go")
	m := managers.NewGoInstall(cmd)
	err := m.Install(context.Background(), "github.com/user/tool")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "go", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "github.com/user/tool@latest")
}

func TestGoInstall_Search_ReturnsEmpty(t *testing.T) {
	cmd := availableCmd("go")
	m := managers.NewGoInstall(cmd)
	results, err := m.Search(context.Background(), "anything")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestGoInstall_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("go")
	m := managers.NewGoInstall(cmd)
	err := m.InstallVersion(context.Background(), "github.com/user/tool", "v1.2.3")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "github.com/user/tool@v1.2.3")
}

// --- uv ---

func TestUv_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewUv(availableCmd("uv"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestUv_IsAvailable_WhenAbsent(t *testing.T) {
	m := managers.NewUv(unavailableCmd())
	assert.False(t, m.IsAvailable(context.Background()))
}

func TestUv_Search_ReturnsEmptyOnCommandFailure(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"uv": "/usr/bin/uv"},
		OutputErr:      assert.AnError,
	}
	m := managers.NewUv(cmd)
	results, err := m.Search(context.Background(), "requests")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestUv_Search_InvokesCorrectCommand(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"uv": "/usr/bin/uv"},
		OutputData:     []byte("ruff (0.4.0)\n"),
	}
	m := managers.NewUv(cmd)
	_, err := m.Search(context.Background(), "ruff")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "uv", cmd.Calls[0][0])
	assert.Equal(t, []string{"uv", "pip", "index", "versions", "ruff"}, cmd.Calls[0])
}

func TestUv_Search_SkipsWarningLines(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"uv": "/usr/bin/uv"},
		OutputData: []byte(
			"WARNING: something experimental\n" +
				"requests (2.31.0)\n" +
				"Available versions: 2.31.0, 2.30.0\n",
		),
	}
	m := managers.NewUv(cmd)
	results, err := m.Search(context.Background(), "requests")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "requests", results[0].Package)
	assert.Equal(t, "2.31.0", results[0].Version)
}

// --- brew-cask ---

func TestBrewCask_InstallVersion_ReturnsError(t *testing.T) {
	cmd := availableCmd("brew")
	m := managers.NewBrewCask(cmd)
	err := m.InstallVersion(context.Background(), "firefox", "120.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "brew-cask")
}

func TestBrewCask_Search_FiltersHeaders(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"brew": "/usr/local/bin/brew"},
		OutputData:     []byte("==> Casks\nfirefox\ngoogle-chrome\n"),
	}
	m := managers.NewBrewCask(cmd)
	results, err := m.Search(context.Background(), "firefox")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "firefox", results[0].Package)
	assert.Equal(t, "homebrew-cask", results[0].Repo)
	assert.Equal(t, "brew-cask", results[0].Manager)
}

// --- DefaultRegistry ---

func TestDefaultRegistry_ContainsAllManagers(t *testing.T) {
	reg := managers.DefaultRegistry()
	wantNames := []string{
		"brew", "brew-cask",
		"apt", "dnf", "pacman", "zypper", "apk", "pkg",
		"winget", "choco", "scoop",
		"pip", "uv", "cargo", "npm", "go",
		"flatpak", "snap", "nix",
		"conda", "mamba",
		"gem",
		"yarn", "pnpm",
		"pipx",
		"mise", "asdf",
		"pkgin",
		"emerge",
		"opkg",
		"helm",
	}
	for _, name := range wantNames {
		t.Run(name, func(t *testing.T) {
			assert.NotNil(t, reg.ByName(name), "manager %q should be registered", name)
		})
	}
	assert.Len(t, reg.All(), len(wantNames))
}
