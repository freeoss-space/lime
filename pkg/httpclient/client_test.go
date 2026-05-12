package httpclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freeoss-space/lime/pkg/httpclient"
)

func newClient(opts httpclient.Options) *httpclient.Client {
	if opts.RatePerSecond == 0 {
		opts.RatePerSecond = 100 // fast for tests
	}
	if opts.BackoffBase == 0 {
		opts.BackoffBase = time.Millisecond // fast for tests
	}
	return httpclient.New(opts)
}

func TestNew_AppliesDefaults(t *testing.T) {
	c := httpclient.New(httpclient.Options{})
	assert.NotNil(t, c)
}

func TestGet_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{})
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGet_SendsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{UserAgent: "jil/1.0 test"})
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, "jil/1.0 test", gotUA)
}

func TestGet_RetriesOnServerError(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{MaxRetries: 3})
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), calls.Load())
}

func TestGet_RetriesOnTooManyRequests(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{MaxRetries: 3})
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGet_ExhaustsRetriesAndErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{MaxRetries: 2})
	_, err := c.Get(context.Background(), srv.URL)
	assert.Error(t, err)
}

func TestGet_RespectsContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := newClient(httpclient.Options{Timeout: 5 * time.Second})
	_, err := c.Get(ctx, srv.URL)
	assert.Error(t, err)
}

func TestGet_DoesNotRetryNonRetryableStatus(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newClient(httpclient.Options{MaxRetries: 3})
	resp, err := c.Get(context.Background(), srv.URL)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, int32(1), calls.Load(), "should not retry 404")
}

func TestGet_EnforcesRateLimit(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// 2 requests/second — doing 3 requests should take at least 1 second.
	c := httpclient.New(httpclient.Options{RatePerSecond: 2, MaxRetries: 0})

	start := time.Now()
	for i := 0; i < 3; i++ {
		resp, err := c.Get(context.Background(), srv.URL)
		require.NoError(t, err)
		resp.Body.Close()
	}
	elapsed := time.Since(start)
	assert.GreaterOrEqual(t, elapsed, 500*time.Millisecond,
		"3 requests at 2 rps should take ≥500ms")
}
