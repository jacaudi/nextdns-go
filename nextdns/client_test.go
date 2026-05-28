package nextdns

import (
	"bytes"
	"context"
	"crypto/tls"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/matryer/is"
)

func TestNewRequestWithQuery(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)

	query := url.Values{}
	query.Set("from", "-7d")
	query.Set("limit", "100")

	req, err := client.newRequest("GET", "profiles/abc123/analytics/status", query, nil)
	c.NoErr(err)

	c.Equal(req.URL.String(), "https://api.nextdns.io/profiles/abc123/analytics/status?from=-7d&limit=100")
	c.Equal(req.Method, "GET")
}

func TestNewRequestWithQueryEmpty(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)

	req, err := client.newRequest("GET", "profiles/abc123/analytics/status", nil, nil)
	c.NoErr(err)

	c.Equal(req.URL.String(), "https://api.nextdns.io/profiles/abc123/analytics/status")
}

func TestDefaultClientHasTimeout(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)
	c.True(client.client.Timeout > 0)
}

func TestDefaultClientTLSMinVersion(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)

	tr, ok := client.client.Transport.(*http.Transport)
	c.True(ok)
	c.True(tr.TLSClientConfig != nil)
	c.Equal(tr.TLSClientConfig.MinVersion, uint16(tls.VersionTLS13))
}

func TestNewRequestUnifiedNoQuery(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)

	req, err := client.newRequest(http.MethodGet, "profiles", nil, nil)
	c.NoErr(err)
	c.Equal(req.URL.String(), "https://api.nextdns.io/profiles")
}

func TestNewRequestUnifiedWithQuery(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://api.nextdns.io/"))
	c.NoErr(err)

	q := url.Values{}
	q.Set("from", "-7d")

	req, err := client.newRequest(http.MethodGet, "profiles/abc/analytics/status", q, nil)
	c.NoErr(err)
	c.Equal(req.URL.String(), "https://api.nextdns.io/profiles/abc/analytics/status?from=-7d")
}

func TestWithLoggerInjectsHandler(t *testing.T) {
	c := is.New(t)

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"abc"}}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL), WithLogger(logger))
	c.NoErr(err)

	_, err = client.Profiles.Get(context.Background(), &GetProfileRequest{ProfileID: "abc"})
	c.NoErr(err)

	out := buf.String()
	c.True(strings.Contains(out, "nextdns request"))
}

func TestRateLimitObserverInvoked(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "597")
		w.Header().Set("X-RateLimit-Reset", "1730000000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"abc"}}`))
	}))
	defer ts.Close()

	var seen RateLimit
	observer := func(rl RateLimit) { seen = rl }

	client, err := New(WithBaseURL(ts.URL), WithRateLimitObserver(observer))
	c.NoErr(err)

	_, err = client.Profiles.Get(context.Background(), &GetProfileRequest{ProfileID: "abc"})
	c.NoErr(err)

	c.Equal(seen.Limit, 600)
	c.Equal(seen.Remaining, 597)
	c.Equal(seen.Reset.Unix(), int64(1730000000))
}

func TestRateLimitObserverUnsetNoOp(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "600")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"abc"}}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	_, err = client.Profiles.Get(context.Background(), &GetProfileRequest{ProfileID: "abc"})
	c.NoErr(err)
	// No observer set; nothing to assert except that the call succeeds.
}

func TestWithHTTPClientNoTransport(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer ts.Close()

	// Empty *http.Client (Transport == nil). Without the nil-Transport guard
	// in WithAPIKey, the authTransport wrapper captures nil into t.rt, and
	// the first RoundTrip panics with a nil-pointer dereference.
	client, err := New(
		WithBaseURL(ts.URL),
		WithHTTPClient(&http.Client{}),
		WithAPIKey(Secret("test-key")),
	)
	c.NoErr(err)

	req, err := client.newRequest(http.MethodGet, "anything", nil, nil)
	c.NoErr(err)

	// This RoundTrip is the line that panics pre-fix. It must complete
	// without panicking and produce a non-nil response.
	res, err := client.client.Transport.RoundTrip(req)
	c.NoErr(err)
	c.True(res != nil)
	_ = res.Body.Close()
}

func TestAuthTransportDoesNotMutateRequest(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"abc"}}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL), WithAPIKey(Secret("k")))
	c.NoErr(err)

	// Build a request, round-trip it twice through the auth transport.
	// Header should have exactly one X-Api-Key on each attempt (no duplication).
	req, err := client.newRequest(http.MethodGet, "profiles/abc", nil, nil)
	c.NoErr(err)
	c.Equal(len(req.Header.Values("X-Api-Key")), 0) // not added pre-RoundTrip

	res1, err := client.client.Transport.RoundTrip(req)
	c.NoErr(err)
	_ = res1.Body.Close()
	c.Equal(len(req.Header.Values("X-Api-Key")), 0) // not mutated by RoundTrip

	res2, err := client.client.Transport.RoundTrip(req)
	c.NoErr(err)
	_ = res2.Body.Close()
	c.Equal(len(req.Header.Values("X-Api-Key")), 0) // still not mutated
}

func TestRedirectStripsAPIKeyAcrossHosts(t *testing.T) {
	c := is.New(t)

	// Target server (different host) — must NOT receive X-Api-Key.
	var targetSawKey atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "" {
			targetSawKey.Store(true)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("payload"))
	}))
	defer target.Close()

	// API server: 302 → target. Different host because httptest assigns
	// different ports (host == "127.0.0.1:N" differs across servers).
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Header.Get("X-Api-Key"), "k") // API host sees the key
		http.Redirect(w, r, target.URL+"/file", http.StatusFound)
	}))
	defer apiServer.Close()

	client, err := New(WithBaseURL(apiServer.URL), WithAPIKey(Secret("k")))
	c.NoErr(err)

	// Issue any request that goes through the default client; redirect
	// follows automatically. (Don't need Logs.Download for the test —
	// any GET works because the test isolates the redirect behavior.)
	req, err := client.newRequest(http.MethodGet, "anything", nil, nil)
	c.NoErr(err)
	res, err := client.client.Do(req)
	c.NoErr(err)
	_ = res.Body.Close()

	c.True(!targetSawKey.Load()) // target host must not have seen the key
}
