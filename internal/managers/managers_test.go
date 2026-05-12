package managers_test

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/managers"
)

// availableCmd returns a FakeCommander that reports name as available.
func availableCmd(name string) *managers.FakeCommander {
	return &managers.FakeCommander{
		LookPathResult: map[string]string{name: "/usr/bin/" + name},
	}
}

// unavailableCmd returns a FakeCommander that reports no binary as available.
func unavailableCmd() *managers.FakeCommander {
	return &managers.FakeCommander{
		LookPathErr: &exec.Error{Name: "x", Err: exec.ErrNotFound},
	}
}

// --- Availability ---

func TestApt_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewApt(availableCmd("apt-get"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestApt_IsAvailable_WhenAbsent(t *testing.T) {
	m := managers.NewApt(unavailableCmd())
	assert.False(t, m.IsAvailable(context.Background()))
}

func TestBrew_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewBrew(availableCmd("brew"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestPacman_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewPacman(availableCmd("pacman"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestDnf_IsAvailable_WhenPresent(t *testing.T) {
	m := managers.NewDnf(availableCmd("dnf"))
	assert.True(t, m.IsAvailable(context.Background()))
}

// --- InstallArgs ---

func TestApt_InstallArgs(t *testing.T) {
	m := managers.NewApt(availableCmd("apt-get"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "sudo", bin)
	assert.Contains(t, args, "ripgrep")
	assert.Contains(t, args, "apt-get")
}

func TestBrew_InstallArgs(t *testing.T) {
	m := managers.NewBrew(availableCmd("brew"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "brew", bin)
	assert.Contains(t, args, "ripgrep")
}

func TestPacman_InstallArgs(t *testing.T) {
	m := managers.NewPacman(availableCmd("pacman"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "sudo", bin)
	assert.Contains(t, args, "-S")
	assert.Contains(t, args, "ripgrep")
}

func TestDnf_InstallArgs(t *testing.T) {
	m := managers.NewDnf(availableCmd("dnf"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "sudo", bin)
	assert.Contains(t, args, "ripgrep")
}

// --- Install calls commander ---

func TestApt_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("apt-get")
	m := managers.NewApt(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep")
}

func TestBrew_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("brew")
	m := managers.NewBrew(cmd)
	err := m.Install(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "brew", cmd.Calls[0][0])
}

func TestApt_Install_PropagatesError(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"apt-get": "/usr/bin/apt-get"},
		RunErr:         errors.New("permission denied"),
	}
	m := managers.NewApt(cmd)
	err := m.Install(context.Background(), "ripgrep")
	assert.Error(t, err)
}

// --- Search ---

func TestApt_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"apt-get": "/usr/bin/apt-get"},
		OutputData: []byte(
			"ripgrep - recursively searches directories for a regex pattern\n" +
				"ripgrep-all - Runs ripgrep over PDFs, ebooks and more\n",
		),
	}
	m := managers.NewApt(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "apt", results[0].Manager)
}

func TestBrew_Search_FiltersHeaders(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"brew": "/usr/local/bin/brew"},
		OutputData: []byte(
			"==> Formulae\nripgrep\nripgrep-all\n==> Casks\n",
		),
	}
	m := managers.NewBrew(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
}

func TestPacman_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pacman": "/usr/bin/pacman"},
		OutputData: []byte(
			"extra/ripgrep 14.0.3-1\n    ripgrep combines the usability of The Silver Searcher\n",
		),
	}
	m := managers.NewPacman(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "14.0.3-1", results[0].Version)
	assert.Equal(t, "extra", results[0].Repo)
}

func TestDnf_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"dnf": "/usr/bin/dnf"},
		OutputData: []byte(
			"ripgrep.x86_64 : recursively searches directories\n" +
				"ripgrep-debuginfo.x86_64 : Debug information for package ripgrep\n",
		),
	}
	m := managers.NewDnf(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
}

// --- Registry ---

func TestRegistry_Available_FiltersUnavailable(t *testing.T) {
	avail := managers.NewApt(availableCmd("apt-get"))
	unavail := managers.NewBrew(unavailableCmd())

	reg := managers.NewRegistry(avail, unavail)
	got := reg.Available(context.Background())
	assert.Len(t, got, 1)
	assert.Equal(t, "apt", got[0].Name())
}

func TestRegistry_ByName_FindsExisting(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(unavailableCmd()),
	)
	m := reg.ByName("apt")
	require.NotNil(t, m)
	assert.Equal(t, "apt", m.Name())
}

func TestRegistry_ByName_ReturnsNilForUnknown(t *testing.T) {
	reg := managers.NewRegistry(managers.NewApt(availableCmd("apt-get")))
	assert.Nil(t, reg.ByName("nonexistent"))
}

func TestRegistry_Preferred_OrdersCorrectly(t *testing.T) {
	apt := managers.NewApt(availableCmd("apt-get"))
	brew := managers.NewBrew(availableCmd("brew"))
	dnf := managers.NewDnf(availableCmd("dnf"))

	reg := managers.NewRegistry(apt, brew, dnf)
	prefs := []string{"dnf", "brew", "apt"}

	got := reg.Preferred(context.Background(), prefs)
	require.Len(t, got, 3)
	assert.Equal(t, "dnf", got[0].Name())
	assert.Equal(t, "brew", got[1].Name())
	assert.Equal(t, "apt", got[2].Name())
}

func TestRegistry_Preferred_PutsUnpreferencedLast(t *testing.T) {
	apt := managers.NewApt(availableCmd("apt-get"))
	brew := managers.NewBrew(availableCmd("brew"))
	pacman := managers.NewPacman(availableCmd("pacman"))

	reg := managers.NewRegistry(apt, brew, pacman)
	prefs := []string{"brew"}

	got := reg.Preferred(context.Background(), prefs)
	require.Len(t, got, 3)
	assert.Equal(t, "brew", got[0].Name(), "brew should be first (preferred)")
}

// --- Name() ---

func TestAllManagerNames(t *testing.T) {
	cmd := unavailableCmd()
	tests := []struct {
		m    managers.PackageManager
		name string
	}{
		{managers.NewApt(cmd), "apt"},
		{managers.NewBrew(cmd), "brew"},
		{managers.NewBrewCask(cmd), "brew-cask"},
		{managers.NewDnf(cmd), "dnf"},
		{managers.NewPacman(cmd), "pacman"},
		{managers.NewZypper(cmd), "zypper"},
		{managers.NewApk(cmd), "apk"},
		{managers.NewPkg(cmd), "pkg"},
		{managers.NewWinget(cmd), "winget"},
		{managers.NewChoco(cmd), "choco"},
		{managers.NewScoop(cmd), "scoop"},
		{managers.NewPip(cmd), "pip"},
		{managers.NewUv(cmd), "uv"},
		{managers.NewCargo(cmd), "cargo"},
		{managers.NewNpm(cmd), "npm"},
		{managers.NewGoInstall(cmd), "go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.name, tt.m.Name())
		})
	}
}

// --- SupportsVersioning ---

func TestSupportsVersioning(t *testing.T) {
	cmd := unavailableCmd()
	tests := []struct {
		m       managers.PackageManager
		name    string
		support bool
	}{
		{managers.NewApt(cmd), "apt", true},
		{managers.NewBrew(cmd), "brew", true},
		{managers.NewDnf(cmd), "dnf", true},
		{managers.NewZypper(cmd), "zypper", true},
		{managers.NewApk(cmd), "apk", true},
		{managers.NewPkg(cmd), "pkg", true},
		{managers.NewWinget(cmd), "winget", true},
		{managers.NewChoco(cmd), "choco", true},
		{managers.NewPip(cmd), "pip", true},
		{managers.NewUv(cmd), "uv", true},
		{managers.NewCargo(cmd), "cargo", true},
		{managers.NewNpm(cmd), "npm", true},
		{managers.NewGoInstall(cmd), "go", true},
		{managers.NewPacman(cmd), "pacman", false},
		{managers.NewScoop(cmd), "scoop", false},
		{managers.NewBrewCask(cmd), "brew-cask", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.support, tt.m.SupportsVersioning())
		})
	}
}

// --- InstallVersionArgs ---

func TestInstallVersionArgs(t *testing.T) {
	tests := []struct {
		name    string
		m       managers.PackageManager
		pkg     string
		version string
		wantBin string
		wantArg string // expected substring in args
	}{
		{
			name: "apt", m: managers.NewApt(availableCmd("apt-get")),
			pkg: "ripgrep", version: "13.0.0",
			wantBin: "sudo", wantArg: "ripgrep=13.0.0",
		},
		{
			name: "brew", m: managers.NewBrew(availableCmd("brew")),
			pkg: "ripgrep", version: "14.1.1",
			wantBin: "brew", wantArg: "ripgrep@14.1.1",
		},
		{
			name: "dnf", m: managers.NewDnf(availableCmd("dnf")),
			pkg: "ripgrep", version: "14.0.0",
			wantBin: "sudo", wantArg: "ripgrep-14.0.0",
		},
		{
			name: "zypper", m: managers.NewZypper(availableCmd("zypper")),
			pkg: "ripgrep", version: "14.0.0",
			wantBin: "sudo", wantArg: "ripgrep=14.0.0",
		},
		{
			name: "apk", m: managers.NewApk(availableCmd("apk")),
			pkg: "ripgrep", version: "14.0.0",
			wantBin: "sudo", wantArg: "ripgrep=14.0.0",
		},
		{
			name: "pkg", m: managers.NewPkg(availableCmd("pkg")),
			pkg: "ripgrep", version: "14.0.0",
			wantBin: "sudo", wantArg: "ripgrep-14.0.0",
		},
		{
			name: "winget", m: managers.NewWinget(availableCmd("winget")),
			pkg: "ripgrep", version: "14.1.1",
			wantBin: "winget", wantArg: "14.1.1",
		},
		{
			name: "choco", m: managers.NewChoco(availableCmd("choco")),
			pkg: "ripgrep", version: "14.1.1",
			wantBin: "choco", wantArg: "14.1.1",
		},
		{
			name: "pip", m: managers.NewPip(availableCmd("pip3")),
			pkg: "requests", version: "2.31.0",
			wantBin: "pip3", wantArg: "requests==2.31.0",
		},
		{
			name: "cargo", m: managers.NewCargo(availableCmd("cargo")),
			pkg: "ripgrep", version: "14.1.1",
			wantBin: "cargo", wantArg: "14.1.1",
		},
		{
			name: "npm", m: managers.NewNpm(availableCmd("npm")),
			pkg: "typescript", version: "5.0.0",
			wantBin: "npm", wantArg: "typescript@5.0.0",
		},
		{
			name: "go", m: managers.NewGoInstall(availableCmd("go")),
			pkg: "golang.org/x/tools/cmd/goimports", version: "v0.1.0",
			wantBin: "go", wantArg: "golang.org/x/tools/cmd/goimports@v0.1.0",
		},
		{
			name: "uv", m: managers.NewUv(availableCmd("uv")),
			pkg: "ruff", version: "0.4.0",
			wantBin: "uv", wantArg: "ruff==0.4.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin, args := tt.m.InstallVersionArgs(tt.pkg, tt.version)
			assert.Equal(t, tt.wantBin, bin)
			assert.Contains(t, args, tt.wantArg)
		})
	}
}

func TestInstallVersion_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("apt-get")
	m := managers.NewApt(cmd)
	err := m.InstallVersion(context.Background(), "ripgrep", "13.0.0")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep=13.0.0")
}

func TestInstallVersion_UnsupportedManagerReturnsError(t *testing.T) {
	cmd := availableCmd("pacman")
	m := managers.NewPacman(cmd)
	err := m.InstallVersion(context.Background(), "ripgrep", "14.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pacman")
}

func TestBrew_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("brew")
	m := managers.NewBrew(cmd)
	err := m.InstallVersion(context.Background(), "ripgrep", "14.1.1")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "brew", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep@14.1.1")
}

func TestWinget_InstallVersion_UsesVersionFlag(t *testing.T) {
	cmd := availableCmd("winget")
	m := managers.NewWinget(cmd)
	bin, args := m.InstallVersionArgs("ripgrep", "14.1.1")
	assert.Equal(t, "winget", bin)
	assert.Contains(t, args, "--version")
	assert.Contains(t, args, "14.1.1")
}

func TestChoco_InstallVersion_UsesVersionFlag(t *testing.T) {
	cmd := availableCmd("choco")
	m := managers.NewChoco(cmd)
	bin, args := m.InstallVersionArgs("ripgrep", "14.1.1")
	assert.Equal(t, "choco", bin)
	assert.Contains(t, args, "--version")
	assert.Contains(t, args, "14.1.1")
}

func TestUv_InstallArgs(t *testing.T) {
	m := managers.NewUv(availableCmd("uv"))
	bin, args := m.InstallArgs("ruff")
	assert.Equal(t, "uv", bin)
	assert.Equal(t, []string{"tool", "install", "ruff"}, args)
}

func TestUv_InstallVersion_UsesToolInstallWithEqEqSyntax(t *testing.T) {
	cmd := availableCmd("uv")
	m := managers.NewUv(cmd)
	err := m.InstallVersion(context.Background(), "ruff", "0.4.0")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "uv", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ruff==0.4.0")
	assert.Contains(t, cmd.Calls[0], "tool")
}

func TestUv_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"uv": "/usr/bin/uv"},
		OutputData: []byte(
			"requests (2.31.0)\n",
		),
	}
	m := managers.NewUv(cmd)
	results, err := m.Search(context.Background(), "requests")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "requests", results[0].Package)
	assert.Equal(t, "2.31.0", results[0].Version)
	assert.Equal(t, "pypi", results[0].Repo)
}

func TestBrewCask_InstallArgs(t *testing.T) {
	m := managers.NewBrewCask(availableCmd("brew"))
	bin, args := m.InstallArgs("firefox")
	assert.Equal(t, "brew", bin)
	assert.Equal(t, []string{"install", "--cask", "firefox"}, args)
}

func TestBrewCask_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("brew")
	m := managers.NewBrewCask(cmd)
	err := m.Install(context.Background(), "firefox")
	require.NoError(t, err)
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "brew", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "--cask")
	assert.Contains(t, cmd.Calls[0], "firefox")
}

func TestBrewCask_Search_FiltersCaskResults(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"brew": "/usr/local/bin/brew"},
		OutputData:     []byte("firefox\ngoogle-chrome\n"),
	}
	m := managers.NewBrewCask(cmd)
	results, err := m.Search(context.Background(), "firefox")
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "firefox", results[0].Package)
	assert.Equal(t, "homebrew-cask", results[0].Repo)
}

func TestBrewCask_IsAvailable_UsesBrewBinary(t *testing.T) {
	m := managers.NewBrewCask(availableCmd("brew"))
	assert.True(t, m.IsAvailable(context.Background()))
}

func TestBrewCask_IsAvailable_WhenBrewAbsent(t *testing.T) {
	m := managers.NewBrewCask(unavailableCmd())
	assert.False(t, m.IsAvailable(context.Background()))
}
