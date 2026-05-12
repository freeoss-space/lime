package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/cooldown"
	"github.com/freeoss-space/lime/internal/managers"
	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/internal/search"
)

// --- fakes ---

type fakeRepology struct {
	searchResult repology.ProjectPackages
	searchErr    error
	getResult    []repology.Package
	getErr       error
}

func (f *fakeRepology) SearchProjects(_ context.Context, _ string) (repology.ProjectPackages, error) {
	return f.searchResult, f.searchErr
}

func (f *fakeRepology) GetProject(_ context.Context, _ string) ([]repology.Package, error) {
	return f.getResult, f.getErr
}

func availableCmd(binary string) *managers.FakeCommander {
	return &managers.FakeCommander{
		LookPathResult: map[string]string{binary: "/usr/bin/" + binary},
	}
}

func ripgrepProjects() repology.ProjectPackages {
	return repology.ProjectPackages{
		"ripgrep": {
			{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3-1", Status: "newest"},
			{Repo: "homebrew", Name: "ripgrep", Version: "14.0.3", Status: "newest"},
			{Repo: "arch", Name: "ripgrep", Version: "14.0.3-1", Status: "newest"},
		},
	}
}

// --- tests ---

func TestSearch_ReturnsResults(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(&managers.FakeCommander{}),
	)
	rep := &fakeRepology{searchResult: ripgrepProjects()}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{"apt"}})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestSearch_AnnotatesPreferred(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(availableCmd("brew")),
	)
	rep := &fakeRepology{searchResult: ripgrepProjects()}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{"apt"}})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	preferred := map[string]bool{}
	for _, r := range results {
		if r.Preferred {
			preferred[r.Manager] = true
		}
	}
	assert.True(t, preferred["apt"], "apt should be marked preferred")
	assert.False(t, preferred["brew"], "brew should not be preferred")
}

func TestSearch_AnnotatesAvailability(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(&managers.FakeCommander{}), // unavailable
	)
	rep := &fakeRepology{searchResult: ripgrepProjects()}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{}})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	for _, r := range results {
		switch r.Manager {
		case "apt":
			assert.True(t, r.Available, "apt should be available")
		case "brew":
			assert.False(t, r.Available, "brew should be unavailable")
		}
	}
}

func TestSearch_SortsPreferredFirst(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(availableCmd("brew")),
		managers.NewPacman(availableCmd("pacman")),
	)
	rep := &fakeRepology{searchResult: ripgrepProjects()}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{"brew"}})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.NotEmpty(t, results)
	// First result should be from brew (preferred).
	assert.True(t, results[0].Preferred, "first result should be preferred")
}

func TestSearch_ReturnsErrorOnRepologyFailure(t *testing.T) {
	reg := managers.NewRegistry(managers.NewApt(availableCmd("apt-get")))
	rep := &fakeRepology{searchErr: errors.New("network error")}

	s := search.New(reg, rep, search.Options{})
	_, err := s.Search(context.Background(), "ripgrep")
	assert.Error(t, err)
}

func TestSearch_FiltersUnknownRepos(t *testing.T) {
	reg := managers.NewRegistry(managers.NewApt(availableCmd("apt-get")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {
			{Repo: "some_unknown_repo_xyz", Name: "ripgrep", Version: "14.0.3"},
		},
	}}

	s := search.New(reg, rep, search.Options{})
	results, err := s.Search(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Empty(t, results, "unknown repos should be filtered out")
}

func TestGetProject_ReturnsResults(t *testing.T) {
	reg := managers.NewRegistry(managers.NewApt(availableCmd("apt-get")))
	rep := &fakeRepology{getResult: []repology.Package{
		{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3-1"},
		{Repo: "homebrew", Name: "ripgrep", Version: "14.0.3"},
	}}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{"apt"}})
	results, err := s.GetProject(context.Background(), "ripgrep")

	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestGetProject_ReturnsErrorOnRepologyFailure(t *testing.T) {
	reg := managers.NewRegistry()
	rep := &fakeRepology{getErr: errors.New("timeout")}

	s := search.New(reg, rep, search.Options{})
	_, err := s.GetProject(context.Background(), "ripgrep")
	assert.Error(t, err)
}

func TestGetProject_SortsPreferredFirst(t *testing.T) {
	reg := managers.NewRegistry(
		managers.NewApt(availableCmd("apt-get")),
		managers.NewBrew(availableCmd("brew")),
	)
	rep := &fakeRepology{getResult: []repology.Package{
		{Repo: "homebrew", Name: "ripgrep", Version: "14.0.3"},
		{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3-1"},
	}}

	s := search.New(reg, rep, search.Options{PreferredManagers: []string{"apt"}})
	results, err := s.GetProject(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.NotEmpty(t, results)
	assert.Equal(t, "apt", results[0].Manager, "apt should be first (preferred)")
}

// --- age and cooldown eligibility ---

var fixedNow = time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)

func TestSearch_DaysOld_PopulatedWhenTimestampAvailable(t *testing.T) {
	releasedTenDaysAgo := fixedNow.Add(-10 * 24 * time.Hour)
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1", UpdatedAt: &releasedTenDaysAgo}},
	}}

	s := search.New(reg, rep, search.Options{Now: fixedNow})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].DaysOld)
	assert.Equal(t, 10, *results[0].DaysOld)
}

func TestSearch_DaysOld_NilWhenNoTimestamp(t *testing.T) {
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1"}},
	}}

	s := search.New(reg, rep, search.Options{Now: fixedNow})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Nil(t, results[0].DaysOld)
}

func TestSearch_CooldownOK_TrueWhenNoTimestamp(t *testing.T) {
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1"}},
	}}
	cd, _ := cooldown.Parse("14d")

	s := search.New(reg, rep, search.Options{Now: fixedNow, Cooldown: cd})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].CooldownOK, "no timestamp = can't determine age = treat as OK")
}

func TestSearch_CooldownOK_FalseWhenTooNew(t *testing.T) {
	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}
	cd, _ := cooldown.Parse("14d")

	s := search.New(reg, rep, search.Options{Now: fixedNow, Cooldown: cd})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].CooldownOK, "1-day-old release should not pass 14d cooldown")
}

func TestSearch_CooldownOK_TrueWhenOldEnough(t *testing.T) {
	releasedLongAgo := fixedNow.Add(-30 * 24 * time.Hour)
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "14.1.1", UpdatedAt: &releasedLongAgo}},
	}}
	cd, _ := cooldown.Parse("14d")

	s := search.New(reg, rep, search.Options{Now: fixedNow, Cooldown: cd})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].CooldownOK)
}

func TestSearch_CooldownOK_TrueWhenZeroCooldown(t *testing.T) {
	releasedYesterday := fixedNow.Add(-24 * time.Hour)
	reg := managers.NewRegistry(managers.NewBrew(availableCmd("brew")))
	rep := &fakeRepology{searchResult: repology.ProjectPackages{
		"ripgrep": {{Repo: "homebrew", Name: "ripgrep", Version: "15.0.0", UpdatedAt: &releasedYesterday}},
	}}

	s := search.New(reg, rep, search.Options{Now: fixedNow, Cooldown: cooldown.Zero})
	results, err := s.Search(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].CooldownOK, "zero cooldown never blocks")
}

func TestGetProject_DaysOld_Populated(t *testing.T) {
	releasedFiveDaysAgo := fixedNow.Add(-5 * 24 * time.Hour)
	reg := managers.NewRegistry(managers.NewApt(availableCmd("apt-get")))
	rep := &fakeRepology{getResult: []repology.Package{
		{Repo: "debian_stable", Name: "ripgrep", Version: "13.0.0", UpdatedAt: &releasedFiveDaysAgo},
	}}

	s := search.New(reg, rep, search.Options{Now: fixedNow})
	results, err := s.GetProject(context.Background(), "ripgrep")

	require.NoError(t, err)
	require.Len(t, results, 1)
	require.NotNil(t, results[0].DaysOld)
	assert.Equal(t, 5, *results[0].DaysOld)
}
