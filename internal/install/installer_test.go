package install_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/install"
	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/internal/resolution"
)

// --- fakes ---

type fakeRepology struct {
	pkgs map[string][]repology.Package
	err  error
}

func (f *fakeRepology) GetProject(_ context.Context, name string) ([]repology.Package, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.pkgs[name], nil
}

func availableCmd(binary string) *managers.FakeCommander {
	return &managers.FakeCommander{
		LookPathResult: map[string]string{binary: "/usr/bin/" + binary},
	}
}

func makeRegistry(ms ...managers.PackageManager) *managers.Registry {
	return managers.NewRegistry(ms...)
}

func defaultOpts(stdout *bytes.Buffer) install.Options {
	return install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"apt", "brew"},
		Stdout:            stdout,
		Stdin:             strings.NewReader(""),
	}
}

var fixedNow = time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

// --- existing install behaviour ---

func TestInstaller_InstallsWithPreferredManager(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	brew := managers.NewBrew(&managers.FakeCommander{}) // unavailable

	reg := makeRegistry(apt, brew)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {
			{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3"},
		},
	}}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Equal(t, "apt", results[0].Manager)
	assert.Contains(t, aptCmd.Calls[0], "ripgrep")
}

func TestInstaller_FallsBackWhenManagerUnavailable(t *testing.T) {
	// apt unavailable, brew available
	brewCmd := availableCmd("brew")
	apt := managers.NewApt(&managers.FakeCommander{})
	brew := managers.NewBrew(brewCmd)

	reg := makeRegistry(apt, brew)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {
			{Repo: "homebrew", Name: "ripgrep", Version: "14.0.3"},
		},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"apt", "brew"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Equal(t, "brew", results[0].Manager)
}

func TestInstaller_DryRun_DoesNotInstall(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)

	reg := makeRegistry(apt)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		DryRun:            true,
		AutoConfirm:       true,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Len(t, aptCmd.Calls, 0, "should not call apt in dry-run")
	assert.Contains(t, out.String(), "dry-run")
}

func TestInstaller_ExplicitManager_UsesIt(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	apt := managers.NewApt(availableCmd("apt-get"))

	reg := makeRegistry(brew, apt)
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Manager: "brew"},
	})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Equal(t, "brew", results[0].Manager)
	assert.Len(t, brewCmd.Calls, 1)
}

func TestInstaller_ExplicitManager_UnknownReturnsError(t *testing.T) {
	reg := makeRegistry(managers.NewApt(availableCmd("apt-get")))
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Manager: "nonexistent"},
	})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err)
}

func TestInstaller_ExplicitManager_UnavailableReturnsError(t *testing.T) {
	// brew registered but binary not found
	brew := managers.NewBrew(&managers.FakeCommander{})
	reg := makeRegistry(brew)
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Manager: "brew"},
	})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err)
}

func TestInstaller_NoManagersAvailable(t *testing.T) {
	reg := makeRegistry() // empty
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err)
}

func TestInstaller_RepologyError_FallsThroughToManager(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)

	rep := &fakeRepology{err: errors.New("network unreachable")}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err, "should fall back to apt even without Repology")
}

func TestInstaller_Confirmation_Accepted(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       false,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
		Stdin:             strings.NewReader("y\n"),
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.False(t, results[0].Skipped)
	assert.Contains(t, out.String(), "Install package?")
}

func TestInstaller_Confirmation_Declined(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       false,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
		Stdin:             strings.NewReader("n\n"),
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.True(t, results[0].Skipped)
	assert.Len(t, aptCmd.Calls, 0, "should not install when declined")
}

func TestInstaller_InstallError_PropagatesInResult(t *testing.T) {
	cmd := &managers.FakeCommander{
		LookPathResult: map[string]string{"apt-get": "/usr/bin/apt-get"},
		RunErr:         errors.New("apt locked"),
	}
	apt := managers.NewApt(cmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err)
}

func TestInstaller_MultiplePackages(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3"}},
		"fd":      {{Repo: "debian_stable", Name: "fd-find", Version: "9.0.0"}},
	}}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep"},
		{Name: "fd"},
	})

	assert.Len(t, results, 2)
	assert.NoError(t, results[0].Err)
	assert.NoError(t, results[1].Err)
}

func TestResult_CommandContainsPackage(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{}

	var out bytes.Buffer
	ins := install.New(reg, rep, defaultOpts(&out))
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.Contains(t, results[0].Command, "ripgrep")
}

// --- versioned install ---

func TestInstaller_VersionedInstall_UsesInstallVersion(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Version: "14.1.1"},
	})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Equal(t, "brew", results[0].Manager)
	require.Len(t, brewCmd.Calls, 1)
	assert.Contains(t, brewCmd.Calls[0], "ripgrep@14.1.1")
}

func TestInstaller_VersionedInstall_AptUsesEqualsSyntax(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "13.0.0"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Version: "13.0.0"},
	})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	require.Len(t, aptCmd.Calls, 1)
	assert.Contains(t, aptCmd.Calls[0], "ripgrep=13.0.0")
}

func TestInstaller_VersionConfirmationPrompt_ShowsVersion(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       false,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		Stdin:             strings.NewReader("y\n"),
	}
	ins := install.New(reg, rep, opts)
	ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Version: "14.1.1"},
	})

	prompt := out.String()
	assert.Contains(t, prompt, "Install package?")
	assert.Contains(t, prompt, "14.1.1")
}

// --- cooldown ---

func TestInstaller_CooldownBlocksNewVersion(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		Cooldown:          "14d",
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err, "should fail when cooldown blocks the only available version")
	assert.Len(t, brewCmd.Calls, 0, "should not install when blocked")
	var ce *resolution.CooldownError
	assert.ErrorAs(t, results[0].Err, &ce, "error should be CooldownError")
}

func TestInstaller_CooldownPassesOldVersion(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedLongAgo := fixedNow.Add(-30 * 24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1", UpdatedAt: &releasedLongAgo}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		Cooldown:          "14d",
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Len(t, brewCmd.Calls, 1)
}

func TestInstaller_CooldownFallsBackToOtherManager(t *testing.T) {
	brewCmd := availableCmd("brew")
	aptCmd := availableCmd("apt-get")
	brew := managers.NewBrew(brewCmd)
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(brew, apt)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	releasedLongAgo := fixedNow.Add(-90 * 24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {
			{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday},
			{Repo: "debian_stable", Name: "ripgrep", Version: "13.0.0", UpdatedAt: &releasedLongAgo},
		},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew", "apt"},
		Stdout:            &out,
		Cooldown:          "14d",
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err)
	assert.Equal(t, "apt", results[0].Manager, "should fall back to apt when brew is cooldown-blocked")
	assert.Len(t, brewCmd.Calls, 0, "brew should not be called")
	assert.Len(t, aptCmd.Calls, 1)
}

func TestInstaller_DefaultCooldownFromConfig(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		DefaultCooldown:   "14d", // from config, not CLI
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.Error(t, results[0].Err)
}

func TestInstaller_ManagerSpecificCooldownOverridesDefault(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		DefaultCooldown:   "30d",                           // global: would block
		ManagerCooldowns:  map[string]string{"brew": "0d"}, // override: no cooldown for brew
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err, "manager-specific cooldown of 0d overrides global 30d")
}

func TestInstaller_CLICooldownOverridesConfig(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		Cooldown:          "0d",  // CLI: no cooldown (override)
		DefaultCooldown:   "14d", // config: would block
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err, "CLI --cooldown 0d should override config 14d")
}

func TestInstaller_NoCooldownByDefault(t *testing.T) {
	brewCmd := availableCmd("brew")
	brew := managers.NewBrew(brewCmd)
	reg := makeRegistry(brew)

	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"brew"},
		Stdout:            &out,
		// No cooldown options set
	}
	ins := install.New(reg, rep, opts)
	results := ins.Install(context.Background(), []managers.PackageSpec{{Name: "ripgrep"}})

	require.Len(t, results, 1)
	assert.NoError(t, results[0].Err, "no cooldown by default — any version is acceptable")
}

func TestInstaller_VersionNotFound_FallsBackGracefully(t *testing.T) {
	aptCmd := availableCmd("apt-get")
	apt := managers.NewApt(aptCmd)
	reg := makeRegistry(apt)
	// apt only has version 13.0.0, not 99.0.0
	rep := &fakeRepology{pkgs: map[string][]repology.Package{
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep", Version: "13.0.0"}},
	}}

	var out bytes.Buffer
	opts := install.Options{
		AutoConfirm:       true,
		PreferredManagers: []string{"apt"},
		Stdout:            &out,
	}
	ins := install.New(reg, rep, opts)
	// Request version that doesn't exist
	results := ins.Install(context.Background(), []managers.PackageSpec{
		{Name: "ripgrep", Version: "99.0.0"},
	})

	require.Len(t, results, 1)
	// Version 99.0.0 not in Repology data → no manager has it
	assert.Error(t, results[0].Err)
	assert.Len(t, aptCmd.Calls, 0, "should not install when version not found")
}
