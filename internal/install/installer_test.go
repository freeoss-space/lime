package install_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/install"
	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
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

// --- tests ---

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
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep"}},
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
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep"}},
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
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep"}},
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
		"ripgrep": {{Repo: "debian_stable", Name: "ripgrep"}},
		"fd":      {{Repo: "debian_stable", Name: "fd-find"}},
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
