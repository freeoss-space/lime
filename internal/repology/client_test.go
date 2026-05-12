package repology_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/internal/repology"
	"github.com/freeoss-space/lime/pkg/httpclient"
)

// newTestClient creates a repology.Client pointed at the test server.
func newTestClient(srv *httptest.Server) *repology.Client {
	hc := httpclient.New(httpclient.Options{
		RatePerSecond: 100,
		MaxRetries:    3,
		BackoffBase:   time.Millisecond, // fast backoff in tests
	})
	return repology.NewWithBaseURL(repology.HTTPGetterFromClient(hc), srv.URL)
}

// httpGetterFunc is a function that implements HTTPGetter for tests.
type httpGetterFunc func(ctx context.Context, url string) (*http.Response, error)

func (f httpGetterFunc) Get(ctx context.Context, url string) (*http.Response, error) {
	return f(ctx, url)
}

// fakeRipgrepPackages returns representative Repology-style response data.
func fakeRipgrepPackages() []repology.Package {
	return []repology.Package{
		{Repo: "debian_stable", Name: "ripgrep", Version: "14.0.3-1", Status: "newest"},
		{Repo: "homebrew", Name: "ripgrep", Version: "14.0.3", Status: "newest"},
		{Repo: "arch", Name: "ripgrep", Version: "14.0.3-1", Status: "newest"},
		{Repo: "fedora_40", Name: "ripgrep", Version: "14.0.3", Status: "newest"},
	}
}

func TestGetProject_ReturnsPackages(t *testing.T) {
	pkgs := fakeRipgrepPackages()
	data, _ := json.Marshal(pkgs)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "ripgrep")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	result, err := c.GetProject(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, result, 4)
	assert.Equal(t, "debian_stable", result[0].Repo)
}

func TestGetProject_ReturnsNilOnNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	result, err := c.GetProject(context.Background(), "nonexistent-package-xyz")
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestGetProject_ReturnsErrorOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetProject(context.Background(), "ripgrep")
	assert.Error(t, err)
}

func TestGetProject_ReturnsErrorOnInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.GetProject(context.Background(), "ripgrep")
	assert.Error(t, err)
}

func TestSearchProjects_ReturnsResults(t *testing.T) {
	result := repology.ProjectPackages{
		"ripgrep": fakeRipgrepPackages(),
		"ripgrep-all": {
			{Repo: "homebrew", Name: "ripgrep-all", Version: "0.9.6", Status: "newest"},
		},
	}
	data, _ := json.Marshal(result)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.String(), "search=ripgrep")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	projects, err := c.SearchProjects(context.Background(), "ripgrep")
	require.NoError(t, err)
	assert.Len(t, projects, 2)
	assert.Contains(t, projects, "ripgrep")
	assert.Contains(t, projects, "ripgrep-all")
}

func TestSearchProjects_ReturnsErrorOnBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.SearchProjects(context.Background(), "ripgrep")
	assert.Error(t, err)
}

func TestPackagesForManager_FiltersCorrectly(t *testing.T) {
	pkgs := fakeRipgrepPackages()

	tests := []struct {
		manager  string
		expected int
	}{
		{"apt", 1},
		{"brew", 1},
		{"pacman", 1},
		{"dnf", 1},
		{"zypper", 0},
	}

	for _, tt := range tests {
		t.Run(tt.manager, func(t *testing.T) {
			got := repology.PackagesForManager(pkgs, tt.manager)
			assert.Len(t, got, tt.expected, "manager=%s", tt.manager)
		})
	}
}

func TestManagerForRepo(t *testing.T) {
	tests := []struct {
		repo     string
		expected string
	}{
		{"debian_stable", "apt"},
		{"ubuntu_22_04", "apt"},
		{"homebrew", "brew"},
		{"arch", "pacman"},
		{"fedora_40", "dnf"},
		{"alpine_3_18", "apk"},
		{"freebsd", "pkg"},
		{"chocolatey", "choco"},
		{"winget", "winget"},
		{"opensuse_leap", "zypper"},
		{"unknown_repo", ""},
	}

	for _, tt := range tests {
		t.Run(tt.repo, func(t *testing.T) {
			got := repology.ManagerForRepo(tt.repo)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestPackageEffectiveName(t *testing.T) {
	tests := []struct {
		pkg         repology.Package
		projectName string
		expected    string
	}{
		{repology.Package{Name: "rg"}, "ripgrep", "rg"},
		{repology.Package{SrcName: "ripgrep-src"}, "ripgrep", "ripgrep-src"},
		{repology.Package{}, "ripgrep", "ripgrep"},
	}

	for _, tt := range tests {
		got := tt.pkg.EffectiveName(tt.projectName)
		assert.Equal(t, tt.expected, got)
	}
}
