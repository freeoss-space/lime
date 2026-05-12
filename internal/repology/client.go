// Package repology provides a client for the Repology package metadata API.
// See https://repology.org/api for API documentation and usage guidelines.
package repology

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/freeoss-space/lime/pkg/httpclient"
)

const (
	baseURL = "https://repology.org/api/v0"
)

// HTTPGetter abstracts the HTTP GET operation for testability.
type HTTPGetter interface {
	Get(ctx context.Context, url string) (*http.Response, error)
}

// Client queries the Repology API.
type Client struct {
	http    HTTPGetter
	baseURL string
}

// New creates a Client backed by the given HTTPGetter.
// Use httpclient.New(...) for production; swap with a fake in tests.
func New(hc HTTPGetter) *Client {
	return &Client{
		http:    hc,
		baseURL: baseURL,
	}
}

// NewWithBaseURL creates a Client targeting a custom base URL — used in tests.
func NewWithBaseURL(hc HTTPGetter, base string) *Client {
	return &Client{http: hc, baseURL: base}
}

// GetProject returns all known packages for project name across all repositories.
// Endpoint: GET /api/v0/project/{name}
func (c *Client) GetProject(ctx context.Context, name string) ([]Package, error) {
	u := fmt.Sprintf("%s/project/%s", c.baseURL, url.PathEscape(name))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("repology get project %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repology API returned status %d for project %q", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading repology response: %w", err)
	}

	var pkgs []Package
	if err := json.Unmarshal(body, &pkgs); err != nil {
		return nil, fmt.Errorf("parsing repology response: %w", err)
	}
	return pkgs, nil
}

// SearchProjects searches for projects matching the given name prefix.
// Endpoint: GET /api/v0/projects/?search={name}&count=20
func (c *Client) SearchProjects(ctx context.Context, name string) (ProjectPackages, error) {
	u := fmt.Sprintf("%s/projects/?search=%s&count=20", c.baseURL, url.QueryEscape(name))

	resp, err := c.http.Get(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("repology search %q: %w", name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("repology API returned status %d for search %q", resp.StatusCode, name)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading repology response: %w", err)
	}

	var result ProjectPackages
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing repology response: %w", err)
	}
	return result, nil
}

// PackagesForManager filters packages to those available through the named manager.
func PackagesForManager(pkgs []Package, manager string) []Package {
	var out []Package
	for _, p := range pkgs {
		if ManagerForRepo(p.Repo) == manager {
			out = append(out, p)
		}
	}
	return out
}

// defaultHTTPGetter wraps httpclient.Client to satisfy HTTPGetter.
type defaultHTTPGetter struct {
	c *httpclient.Client
}

func (d *defaultHTTPGetter) Get(ctx context.Context, url string) (*http.Response, error) {
	return d.c.Get(ctx, url)
}

// HTTPGetterFromClient wraps a *httpclient.Client as an HTTPGetter. Useful in
// tests that need to override the base URL while using the real HTTP client.
func HTTPGetterFromClient(hc *httpclient.Client) HTTPGetter {
	return &defaultHTTPGetter{c: hc}
}

// NewDefault creates a production-ready Client with the supplied httpclient.
func NewDefault(hc *httpclient.Client) *Client {
	return New(&defaultHTTPGetter{c: hc})
}
