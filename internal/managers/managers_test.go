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
		{managers.NewFlatpak(cmd), "flatpak"},
		{managers.NewSnap(cmd), "snap"},
		{managers.NewNix(cmd), "nix"},
		{managers.NewConda(cmd), "conda"},
		{managers.NewMamba(cmd), "mamba"},
		{managers.NewGem(cmd), "gem"},
		{managers.NewYarn(cmd), "yarn"},
		{managers.NewPnpm(cmd), "pnpm"},
		{managers.NewPipx(cmd), "pipx"},
		{managers.NewMise(cmd), "mise"},
		{managers.NewAsdf(cmd), "asdf"},
		{managers.NewPkgin(cmd), "pkgin"},
		{managers.NewEmerge(cmd), "emerge"},
		{managers.NewOpkg(cmd), "opkg"},
		{managers.NewHelm(cmd), "helm"},
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
		{managers.NewFlatpak(cmd), "flatpak", false},
		{managers.NewSnap(cmd), "snap", false},
		{managers.NewNix(cmd), "nix", false},
		{managers.NewConda(cmd), "conda", true},
		{managers.NewMamba(cmd), "mamba", true},
		{managers.NewGem(cmd), "gem", true},
		{managers.NewYarn(cmd), "yarn", true},
		{managers.NewPnpm(cmd), "pnpm", true},
		{managers.NewPipx(cmd), "pipx", true},
		{managers.NewMise(cmd), "mise", true},
		{managers.NewAsdf(cmd), "asdf", true},
		{managers.NewPkgin(cmd), "pkgin", false},
		{managers.NewEmerge(cmd), "emerge", false},
		{managers.NewOpkg(cmd), "opkg", false},
		{managers.NewHelm(cmd), "helm", true},
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
		{
			name: "conda", m: managers.NewConda(availableCmd("conda")),
			pkg: "numpy", version: "1.24.0",
			wantBin: "conda", wantArg: "numpy=1.24.0",
		},
		{
			name: "mamba", m: managers.NewMamba(availableCmd("mamba")),
			pkg: "numpy", version: "1.24.0",
			wantBin: "mamba", wantArg: "numpy=1.24.0",
		},
		{
			name: "gem", m: managers.NewGem(availableCmd("gem")),
			pkg: "rails", version: "7.1.0",
			wantBin: "gem", wantArg: "7.1.0",
		},
		{
			name: "yarn", m: managers.NewYarn(availableCmd("yarn")),
			pkg: "typescript", version: "5.0.0",
			wantBin: "yarn", wantArg: "typescript@5.0.0",
		},
		{
			name: "pnpm", m: managers.NewPnpm(availableCmd("pnpm")),
			pkg: "typescript", version: "5.0.0",
			wantBin: "pnpm", wantArg: "typescript@5.0.0",
		},
		{
			name: "pipx", m: managers.NewPipx(availableCmd("pipx")),
			pkg: "ruff", version: "0.4.0",
			wantBin: "pipx", wantArg: "ruff==0.4.0",
		},
		{
			name: "mise", m: managers.NewMise(availableCmd("mise")),
			pkg: "node", version: "20.0.0",
			wantBin: "mise", wantArg: "node@20.0.0",
		},
		{
			name: "asdf", m: managers.NewAsdf(availableCmd("asdf")),
			pkg: "nodejs", version: "20.0.0",
			wantBin: "asdf", wantArg: "20.0.0",
		},
		{
			name: "helm", m: managers.NewHelm(availableCmd("helm")),
			pkg: "bitnami/redis", version: "18.4.0",
			wantBin: "helm", wantArg: "18.4.0",
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

// --- Flatpak ---

func TestFlatpak_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewFlatpak(availableCmd("flatpak")).IsAvailable(context.Background()))
}

func TestFlatpak_IsAvailable_WhenAbsent(t *testing.T) {
	assert.False(t, managers.NewFlatpak(unavailableCmd()).IsAvailable(context.Background()))
}

func TestFlatpak_InstallArgs(t *testing.T) {
	m := managers.NewFlatpak(availableCmd("flatpak"))
	bin, args := m.InstallArgs("org.gnome.gedit")
	assert.Equal(t, "flatpak", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "org.gnome.gedit")
}

func TestFlatpak_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("flatpak")
	m := managers.NewFlatpak(cmd)
	require.NoError(t, m.Install(context.Background(), "org.gnome.gedit"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "flatpak", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "org.gnome.gedit")
}

func TestFlatpak_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"flatpak": "/usr/bin/flatpak"},
		// tab-separated: Name\tDescription\tApplication ID\tVersion\tBranch\tRemotes
		OutputData: []byte("GNOME Text Editor\tText editor\torg.gnome.gedit\t44.1\tstable\tflathub\n"),
	}
	m := managers.NewFlatpak(cmd)
	results, err := m.Search(context.Background(), "gedit")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "org.gnome.gedit", results[0].Package)
	assert.Equal(t, "44.1", results[0].Version)
	assert.Equal(t, "flathub", results[0].Repo)
}

// --- Snap ---

func TestSnap_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewSnap(availableCmd("snap")).IsAvailable(context.Background()))
}

func TestSnap_IsAvailable_WhenAbsent(t *testing.T) {
	assert.False(t, managers.NewSnap(unavailableCmd()).IsAvailable(context.Background()))
}

func TestSnap_InstallArgs(t *testing.T) {
	m := managers.NewSnap(availableCmd("snap"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "snap", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "ripgrep")
}

func TestSnap_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("snap")
	m := managers.NewSnap(cmd)
	require.NoError(t, m.Install(context.Background(), "ripgrep"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "snap", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep")
}

func TestSnap_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"snap": "/usr/bin/snap"},
		// space-separated: Name  Version  Publisher  Notes  Summary (header line first)
		OutputData: []byte("Name\tVersion\tPublisher\tNotes\tSummary\nripgrep\t14.0.0\tcanonical\tclassic\tFast search\n"),
	}
	m := managers.NewSnap(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "14.0.0", results[0].Version)
	assert.Equal(t, "snapcraft", results[0].Repo)
}

// --- Nix ---

func TestNix_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewNix(availableCmd("nix-env")).IsAvailable(context.Background()))
}

func TestNix_IsAvailable_WhenAbsent(t *testing.T) {
	assert.False(t, managers.NewNix(unavailableCmd()).IsAvailable(context.Background()))
}

func TestNix_InstallArgs(t *testing.T) {
	m := managers.NewNix(availableCmd("nix-env"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "nix-env", bin)
	assert.Contains(t, args, "-iA")
	assert.Contains(t, args, "nixpkgs.ripgrep")
}

func TestNix_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("nix-env")
	m := managers.NewNix(cmd)
	require.NoError(t, m.Install(context.Background(), "ripgrep"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "nix-env", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "nixpkgs.ripgrep")
}

func TestNix_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"nix-env": "/run/current-system/sw/bin/nix-env"},
		// nix-env -qaP output: "nixpkgs.ripgrep  ripgrep-14.0.0"
		OutputData: []byte("nixpkgs.ripgrep  ripgrep-14.0.0\nnixpkgs.ripgrep-all  ripgrep-all-0.9.0\n"),
	}
	m := managers.NewNix(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "14.0.0", results[0].Version)
	assert.Equal(t, "nixpkgs", results[0].Repo)
}

// --- Conda ---

func TestConda_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewConda(availableCmd("conda")).IsAvailable(context.Background()))
}

func TestConda_InstallArgs(t *testing.T) {
	m := managers.NewConda(availableCmd("conda"))
	bin, args := m.InstallArgs("numpy")
	assert.Equal(t, "conda", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "numpy")
}

func TestConda_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("conda")
	m := managers.NewConda(cmd)
	require.NoError(t, m.Install(context.Background(), "numpy"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "conda", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "numpy")
}

func TestConda_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"conda": "/usr/bin/conda"},
		OutputData: []byte(
			"# Name   Version   Build   Channel\n" +
				"numpy    1.24.0    py312   conda-forge\n" +
				"numpy    1.23.0    py311   conda-forge\n",
		),
	}
	m := managers.NewConda(cmd)
	results, err := m.Search(context.Background(), "numpy")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "numpy", results[0].Package)
	assert.Equal(t, "1.24.0", results[0].Version)
}

func TestConda_InstallVersion_UsesSingleEqualSyntax(t *testing.T) {
	cmd := availableCmd("conda")
	m := managers.NewConda(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "numpy", "1.24.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "numpy=1.24.0")
}

// --- Mamba ---

func TestMamba_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewMamba(availableCmd("mamba")).IsAvailable(context.Background()))
}

func TestMamba_InstallArgs(t *testing.T) {
	m := managers.NewMamba(availableCmd("mamba"))
	bin, args := m.InstallArgs("numpy")
	assert.Equal(t, "mamba", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "numpy")
}

func TestMamba_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("mamba")
	m := managers.NewMamba(cmd)
	require.NoError(t, m.Install(context.Background(), "numpy"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "mamba", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "numpy")
}

func TestMamba_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"mamba": "/usr/bin/mamba"},
		OutputData: []byte(
			"# Name   Version   Build   Channel\n" +
				"numpy    1.24.0    py312   conda-forge\n",
		),
	}
	m := managers.NewMamba(cmd)
	results, err := m.Search(context.Background(), "numpy")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "numpy", results[0].Package)
	assert.Equal(t, "1.24.0", results[0].Version)
}

func TestMamba_InstallVersion_UsesSingleEqualSyntax(t *testing.T) {
	cmd := availableCmd("mamba")
	m := managers.NewMamba(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "numpy", "1.24.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "numpy=1.24.0")
}

// --- Gem ---

func TestGem_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewGem(availableCmd("gem")).IsAvailable(context.Background()))
}

func TestGem_InstallArgs(t *testing.T) {
	m := managers.NewGem(availableCmd("gem"))
	bin, args := m.InstallArgs("rails")
	assert.Equal(t, "gem", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "rails")
}

func TestGem_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("gem")
	m := managers.NewGem(cmd)
	require.NoError(t, m.Install(context.Background(), "rails"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "gem", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "rails")
}

func TestGem_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"gem": "/usr/bin/gem"},
		OutputData: []byte(
			"*** REMOTE GEMS ***\n\n" +
				"rails (7.1.0, 7.0.8)\n" +
				"railties (7.1.0)\n",
		),
	}
	m := managers.NewGem(cmd)
	results, err := m.Search(context.Background(), "rails")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "rails", results[0].Package)
	assert.Equal(t, "7.1.0", results[0].Version)
	assert.Equal(t, "rubygems", results[0].Repo)
}

func TestGem_InstallVersion_UsesVersionFlag(t *testing.T) {
	cmd := availableCmd("gem")
	m := managers.NewGem(cmd)
	bin, args := m.InstallVersionArgs("rails", "7.1.0")
	assert.Equal(t, "gem", bin)
	assert.Contains(t, args, "--version")
	assert.Contains(t, args, "7.1.0")
}

// --- Yarn ---

func TestYarn_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewYarn(availableCmd("yarn")).IsAvailable(context.Background()))
}

func TestYarn_InstallArgs(t *testing.T) {
	m := managers.NewYarn(availableCmd("yarn"))
	bin, args := m.InstallArgs("typescript")
	assert.Equal(t, "yarn", bin)
	assert.Equal(t, []string{"global", "add", "typescript"}, args)
}

func TestYarn_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("yarn")
	m := managers.NewYarn(cmd)
	require.NoError(t, m.Install(context.Background(), "typescript"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "yarn", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "global")
	assert.Contains(t, cmd.Calls[0], "typescript")
}

func TestYarn_Search_ReturnsEmpty(t *testing.T) {
	m := managers.NewYarn(availableCmd("yarn"))
	results, err := m.Search(context.Background(), "typescript")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestYarn_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("yarn")
	m := managers.NewYarn(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "typescript", "5.0.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "typescript@5.0.0")
}

// --- Pnpm ---

func TestPnpm_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewPnpm(availableCmd("pnpm")).IsAvailable(context.Background()))
}

func TestPnpm_InstallArgs(t *testing.T) {
	m := managers.NewPnpm(availableCmd("pnpm"))
	bin, args := m.InstallArgs("typescript")
	assert.Equal(t, "pnpm", bin)
	assert.Equal(t, []string{"add", "-g", "typescript"}, args)
}

func TestPnpm_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("pnpm")
	m := managers.NewPnpm(cmd)
	require.NoError(t, m.Install(context.Background(), "typescript"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "pnpm", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "-g")
	assert.Contains(t, cmd.Calls[0], "typescript")
}

func TestPnpm_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pnpm": "/usr/bin/pnpm"},
		OutputData:     []byte("NAME\t\tVERSION\tDESCRIPTION\ntypescript\t5.4.5\tTypeScript\n"),
	}
	m := managers.NewPnpm(cmd)
	results, err := m.Search(context.Background(), "typescript")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "typescript", results[0].Package)
	assert.Equal(t, "npm", results[0].Repo)
}

func TestPnpm_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("pnpm")
	m := managers.NewPnpm(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "typescript", "5.0.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "typescript@5.0.0")
}

// --- Pipx ---

func TestPipx_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewPipx(availableCmd("pipx")).IsAvailable(context.Background()))
}

func TestPipx_InstallArgs(t *testing.T) {
	m := managers.NewPipx(availableCmd("pipx"))
	bin, args := m.InstallArgs("ruff")
	assert.Equal(t, "pipx", bin)
	assert.Equal(t, []string{"install", "ruff"}, args)
}

func TestPipx_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("pipx")
	m := managers.NewPipx(cmd)
	require.NoError(t, m.Install(context.Background(), "ruff"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "pipx", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ruff")
}

func TestPipx_Search_ReturnsEmpty(t *testing.T) {
	m := managers.NewPipx(availableCmd("pipx"))
	results, err := m.Search(context.Background(), "ruff")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestPipx_InstallVersion_UsesEqEqSyntax(t *testing.T) {
	cmd := availableCmd("pipx")
	m := managers.NewPipx(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "ruff", "0.4.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "ruff==0.4.0")
}

// --- Mise ---

func TestMise_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewMise(availableCmd("mise")).IsAvailable(context.Background()))
}

func TestMise_InstallArgs(t *testing.T) {
	m := managers.NewMise(availableCmd("mise"))
	bin, args := m.InstallArgs("node")
	assert.Equal(t, "mise", bin)
	assert.Equal(t, []string{"use", "-g", "node"}, args)
}

func TestMise_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("mise")
	m := managers.NewMise(cmd)
	require.NoError(t, m.Install(context.Background(), "node"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "mise", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "node")
}

func TestMise_Search_FiltersRegistry(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"mise": "/usr/bin/mise"},
		OutputData:     []byte("node\tcore\tNode.js\npython\tcore\tPython\nruby\tcore\tRuby\n"),
	}
	m := managers.NewMise(cmd)
	results, err := m.Search(context.Background(), "node")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "node", results[0].Package)
	assert.Equal(t, "mise-registry", results[0].Repo)
}

func TestMise_InstallVersion_UsesAtSyntax(t *testing.T) {
	cmd := availableCmd("mise")
	m := managers.NewMise(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "node", "20.0.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "node@20.0.0")
}

// --- Asdf ---

func TestAsdf_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewAsdf(availableCmd("asdf")).IsAvailable(context.Background()))
}

func TestAsdf_InstallArgs(t *testing.T) {
	m := managers.NewAsdf(availableCmd("asdf"))
	bin, args := m.InstallArgs("nodejs")
	assert.Equal(t, "asdf", bin)
	assert.Equal(t, []string{"install", "nodejs", "latest"}, args)
}

func TestAsdf_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("asdf")
	m := managers.NewAsdf(cmd)
	require.NoError(t, m.Install(context.Background(), "nodejs"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "asdf", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "nodejs")
	assert.Contains(t, cmd.Calls[0], "latest")
}

func TestAsdf_Search_ReturnsEmpty(t *testing.T) {
	m := managers.NewAsdf(availableCmd("asdf"))
	results, err := m.Search(context.Background(), "nodejs")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAsdf_InstallVersion_PassesVersionAsArg(t *testing.T) {
	cmd := availableCmd("asdf")
	m := managers.NewAsdf(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "nodejs", "20.0.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, []string{"asdf", "install", "nodejs", "20.0.0"}, cmd.Calls[0])
}

// --- Pkgin ---

func TestPkgin_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewPkgin(availableCmd("pkgin")).IsAvailable(context.Background()))
}

func TestPkgin_InstallArgs(t *testing.T) {
	m := managers.NewPkgin(availableCmd("pkgin"))
	bin, args := m.InstallArgs("ripgrep")
	assert.Equal(t, "pkgin", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "ripgrep")
}

func TestPkgin_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("pkgin")
	m := managers.NewPkgin(cmd)
	require.NoError(t, m.Install(context.Background(), "ripgrep"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "pkgin", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "ripgrep")
}

func TestPkgin_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"pkgin": "/usr/pkg/bin/pkgin"},
		// pkgin search output: "name-version; Description"
		OutputData: []byte("ripgrep-14.0.0;        A search tool combining usability of ag\nbat-0.24.0;            A cat(1) clone with syntax highlighting\n"),
	}
	m := managers.NewPkgin(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "ripgrep", results[0].Package)
	assert.Equal(t, "14.0.0", results[0].Version)
	assert.Equal(t, "pkgsrc", results[0].Repo)
}

// --- Emerge ---

func TestEmerge_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewEmerge(availableCmd("emerge")).IsAvailable(context.Background()))
}

func TestEmerge_InstallArgs(t *testing.T) {
	m := managers.NewEmerge(availableCmd("emerge"))
	bin, args := m.InstallArgs("sys-apps/ripgrep")
	assert.Equal(t, "sudo", bin)
	assert.Contains(t, args, "emerge")
	assert.Contains(t, args, "sys-apps/ripgrep")
}

func TestEmerge_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("emerge")
	m := managers.NewEmerge(cmd)
	require.NoError(t, m.Install(context.Background(), "sys-apps/ripgrep"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "sudo", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "sys-apps/ripgrep")
}

func TestEmerge_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"emerge": "/usr/bin/emerge"},
		OutputData: []byte(
			"*  sys-apps/ripgrep\n" +
				"      Latest version available: 14.0.0\n" +
				"      Description:   A search tool\n" +
				"*  sys-apps/ripgrep-all\n" +
				"      Latest version available: 0.9.0\n",
		),
	}
	m := managers.NewEmerge(cmd)
	results, err := m.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	require.Len(t, results, 2)
	assert.Equal(t, "sys-apps/ripgrep", results[0].Package)
	assert.Equal(t, "14.0.0", results[0].Version)
	assert.Equal(t, "gentoo", results[0].Repo)
}

// --- Opkg ---

func TestOpkg_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewOpkg(availableCmd("opkg")).IsAvailable(context.Background()))
}

func TestOpkg_InstallArgs(t *testing.T) {
	m := managers.NewOpkg(availableCmd("opkg"))
	bin, args := m.InstallArgs("curl")
	assert.Equal(t, "opkg", bin)
	assert.Contains(t, args, "install")
	assert.Contains(t, args, "curl")
}

func TestOpkg_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("opkg")
	m := managers.NewOpkg(cmd)
	require.NoError(t, m.Install(context.Background(), "curl"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "opkg", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "curl")
}

func TestOpkg_Search_ParsesOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"opkg": "/bin/opkg"},
		// opkg find output: "name - version - description"
		OutputData: []byte("curl - 8.5.0-r1 - A client-side URL transfer utility\n"),
	}
	m := managers.NewOpkg(cmd)
	results, err := m.Search(context.Background(), "curl")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "curl", results[0].Package)
	assert.Equal(t, "8.5.0-r1", results[0].Version)
	assert.Equal(t, "openwrt", results[0].Repo)
}

// --- Helm ---

func TestHelm_IsAvailable_WhenPresent(t *testing.T) {
	assert.True(t, managers.NewHelm(availableCmd("helm")).IsAvailable(context.Background()))
}

func TestHelm_InstallArgs_SimpleChart(t *testing.T) {
	m := managers.NewHelm(availableCmd("helm"))
	bin, args := m.InstallArgs("redis")
	assert.Equal(t, "helm", bin)
	assert.Equal(t, []string{"install", "redis", "redis"}, args)
}

func TestHelm_InstallArgs_RepoChart(t *testing.T) {
	m := managers.NewHelm(availableCmd("helm"))
	bin, args := m.InstallArgs("bitnami/redis")
	assert.Equal(t, "helm", bin)
	// release name derived from chart name after "/"
	assert.Equal(t, []string{"install", "redis", "bitnami/redis"}, args)
}

func TestHelm_Install_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("helm")
	m := managers.NewHelm(cmd)
	require.NoError(t, m.Install(context.Background(), "bitnami/redis"))
	require.Len(t, cmd.Calls, 1)
	assert.Equal(t, "helm", cmd.Calls[0][0])
	assert.Contains(t, cmd.Calls[0], "bitnami/redis")
}

func TestHelm_Search_ParsesHubOutput(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"helm": "/usr/bin/helm"},
		// helm search hub output (tab-separated): URL\tCHART VERSION\tAPP VERSION\tDESCRIPTION
		OutputData: []byte(
			"URL\tCHART VERSION\tAPP VERSION\tDESCRIPTION\n" +
				"https://artifacthub.io/packages/helm/bitnami/redis\t18.4.0\t7.2.3\tRedis chart\n",
		),
	}
	m := managers.NewHelm(cmd)
	results, err := m.Search(context.Background(), "redis")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "redis", results[0].Package)
	assert.Equal(t, "18.4.0", results[0].Version)
	assert.Equal(t, "artifacthub", results[0].Repo)
}

func TestHelm_InstallVersion_UsesVersionFlag(t *testing.T) {
	cmd := availableCmd("helm")
	m := managers.NewHelm(cmd)
	bin, args := m.InstallVersionArgs("bitnami/redis", "18.4.0")
	assert.Equal(t, "helm", bin)
	assert.Contains(t, args, "--version")
	assert.Contains(t, args, "18.4.0")
}

func TestHelm_InstallVersion_CallsCorrectCommand(t *testing.T) {
	cmd := availableCmd("helm")
	m := managers.NewHelm(cmd)
	require.NoError(t, m.InstallVersion(context.Background(), "bitnami/redis", "18.4.0"))
	require.Len(t, cmd.Calls, 1)
	assert.Contains(t, cmd.Calls[0], "--version")
	assert.Contains(t, cmd.Calls[0], "18.4.0")
}
