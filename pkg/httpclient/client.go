// Package httpclient provides a rate-limited, retry-capable HTTP client
// for the Repology API and other HTTP targets.
package httpclient

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

const (
	defaultMaxRetries  = 3
	defaultBackoffBase = 500 * time.Millisecond
	defaultBackoffMax  = 30 * time.Second
	defaultTimeout     = 10 * time.Second
	defaultRatePerSec  = 1.0
)

// Options configures the Client.
type Options struct {
	// UserAgent is sent as the User-Agent header on every request.
	UserAgent string
	// Timeout is the per-request timeout (default: 10s).
	Timeout time.Duration
	// MaxRetries is the maximum number of retries after a transient failure (default: 3).
	MaxRetries int
	// RatePerSecond caps the number of requests per second (default: 1).
	RatePerSecond float64
	// BackoffBase is the initial retry backoff duration (default: 500ms).
	// Set lower in tests to keep them fast.
	BackoffBase time.Duration
	// Logger receives structured log events; nil disables logging.
	Logger *slog.Logger
}

// Client is an HTTP client with rate limiting and retry logic.
type Client struct {
	http        *http.Client
	limiter     *rate.Limiter
	opts        Options
	log         *slog.Logger
	backoffBase time.Duration
}

// New creates a Client with the supplied options. Zero-value fields receive
// sensible defaults.
func New(opts Options) *Client {
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = defaultMaxRetries
	}
	if opts.RatePerSecond <= 0 {
		opts.RatePerSecond = defaultRatePerSec
	}
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}
	backoff := opts.BackoffBase
	if backoff <= 0 {
		backoff = defaultBackoffBase
	}

	r := rate.NewLimiter(rate.Limit(opts.RatePerSecond), 1)
	return &Client{
		http:        &http.Client{Timeout: opts.Timeout},
		limiter:     r,
		opts:        opts,
		log:         opts.Logger,
		backoffBase: backoff,
	}
}

// Get performs a GET request to url, respecting rate limits and retrying on
// transient failures. Callers are responsible for closing the response body.
func (c *Client) Get(ctx context.Context, url string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	for attempt := 0; attempt <= c.opts.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.backoff(attempt)
			c.log.Debug("retrying request", "attempt", attempt, "backoff", backoff, "url", url)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err = c.limiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter: %w", err)
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, fmt.Errorf("building request: %w", reqErr)
		}
		if c.opts.UserAgent != "" {
			req.Header.Set("User-Agent", c.opts.UserAgent)
		}

		resp, err = c.http.Do(req)
		if err != nil {
			c.log.Warn("request failed", "attempt", attempt, "error", err)
			continue
		}

		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		// Drain and close so the connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		c.log.Warn("retryable status", "attempt", attempt, "status", resp.StatusCode)
	}

	if err != nil {
		return nil, fmt.Errorf("after %d retries: %w", c.opts.MaxRetries, err)
	}
	return nil, fmt.Errorf("request failed with retryable status after %d retries", c.opts.MaxRetries)
}

// isRetryable returns true for status codes that warrant a retry.
func isRetryable(code int) bool {
	return code == http.StatusTooManyRequests ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusBadGateway ||
		code == http.StatusGatewayTimeout ||
		code == http.StatusInternalServerError
}

// backoff returns the duration to wait before the nth retry using truncated
// exponential back-off anchored at c.backoffBase.
func (c *Client) backoff(attempt int) time.Duration {
	d := time.Duration(float64(c.backoffBase) * math.Pow(2, float64(attempt-1)))
	if d > defaultBackoffMax {
		d = defaultBackoffMax
	}
	return d
}
