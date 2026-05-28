package nextdns

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
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
		// Don't use is.Equal here — it calls t.FailNow, which from a non-test
		// goroutine just goexits the handler goroutine without failing the test.
		if got := r.Header.Get("X-Api-Key"); got != "k" {
			t.Errorf("API host: X-Api-Key = %q, want %q", got, "k")
		}
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

func TestStripAuthOnCrossHost(t *testing.T) {
	c := is.New(t)

	mkReq := func(host string) *http.Request {
		req, err := http.NewRequest(http.MethodGet, "https://"+host+"/x", nil)
		c.NoErr(err)
		req.Header.Set("X-Api-Key", "secret")
		req.Header.Set("Authorization", "Bearer x")
		return req
	}

	// First hop: via is empty, headers preserved.
	first := mkReq("api.nextdns.io")
	err := stripAuthOnCrossHost(first, nil)
	c.NoErr(err)
	c.Equal(first.Header.Get("X-Api-Key"), "secret")
	c.Equal(first.Header.Get("Authorization"), "Bearer x")

	// Same-host redirect: headers preserved.
	same := mkReq("api.nextdns.io")
	err = stripAuthOnCrossHost(same, []*http.Request{mkReq("api.nextdns.io")})
	c.NoErr(err)
	c.Equal(same.Header.Get("X-Api-Key"), "secret")

	// Case-insensitive same-host: still preserved (RFC 4343).
	caseDiff := mkReq("API.NextDNS.io")
	err = stripAuthOnCrossHost(caseDiff, []*http.Request{mkReq("api.nextdns.io")})
	c.NoErr(err)
	c.Equal(caseDiff.Header.Get("X-Api-Key"), "secret")

	// Cross-host redirect: both headers stripped.
	cross := mkReq("cdn.example.com")
	err = stripAuthOnCrossHost(cross, []*http.Request{mkReq("api.nextdns.io")})
	c.NoErr(err)
	c.Equal(cross.Header.Get("X-Api-Key"), "")
	c.Equal(cross.Header.Get("Authorization"), "")

	// 10-redirect cap honored.
	via := make([]*http.Request, 10)
	for i := range via {
		via[i] = mkReq("api.nextdns.io")
	}
	err = stripAuthOnCrossHost(mkReq("api.nextdns.io"), via)
	c.True(err != nil)
}

// TestParseErrorResponse_ContractParity exercises (*Client).parseErrorResponse
// directly to lock in the extracted method's contract. The same status →
// ErrorType mapping is already covered indirectly via the per-service error
// tests (settings/security/profiles/etc.); this test pins the contract for
// the direct callers added in Tasks 5 and 6 (Logs.Download, Logs.Stream).
func TestParseErrorResponse_ContractParity(t *testing.T) {
	// Helper: synthesize an *http.Response from a status code and body string,
	// the way Download/Stream will hand one to parseErrorResponse.
	mkRes := func(status int, body string) *http.Response {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{},
		}
	}

	client := &Client{}

	t.Run("5xx short-circuits without decoding the body", func(t *testing.T) {
		c := is.New(t)

		res := mkRes(http.StatusInternalServerError, "not even JSON")
		err := client.parseErrorResponse(res)

		var apiErr *Error
		c.True(errors.As(err, &apiErr))
		c.Equal(apiErr.Type, ErrorTypeServiceError)
		c.Equal(apiErr.Errors, (*ErrorResponse)(nil))
		c.Equal(apiErr.Meta["body"], "not even JSON")
		c.Equal(apiErr.Meta["http_status"], http.StatusText(http.StatusInternalServerError))
	})

	t.Run("403 maps to ErrorTypeAuthentication", func(t *testing.T) {
		c := is.New(t)

		res := mkRes(http.StatusForbidden, `{"errors":[{"code":"forbidden"}]}`)
		err := client.parseErrorResponse(res)

		var apiErr *Error
		c.True(errors.As(err, &apiErr))
		c.Equal(apiErr.Type, ErrorTypeAuthentication)
		c.True(apiErr.Errors != nil)
	})

	t.Run("404 maps to ErrorTypeNotFound", func(t *testing.T) {
		c := is.New(t)

		res := mkRes(http.StatusNotFound, `{"errors":[{"code":"notFound"}]}`)
		err := client.parseErrorResponse(res)

		var apiErr *Error
		c.True(errors.As(err, &apiErr))
		c.Equal(apiErr.Type, ErrorTypeNotFound)
	})

	t.Run("other 4xx maps to ErrorTypeRequest", func(t *testing.T) {
		c := is.New(t)

		res := mkRes(http.StatusBadRequest, `{"errors":[{"code":"invalidDomain"}]}`)
		err := client.parseErrorResponse(res)

		var apiErr *Error
		c.True(errors.As(err, &apiErr))
		c.Equal(apiErr.Type, ErrorTypeRequest)
	})

	t.Run("malformed JSON body yields ErrorTypeMalformed", func(t *testing.T) {
		c := is.New(t)

		res := mkRes(http.StatusBadRequest, `{"errors": [`)
		err := client.parseErrorResponse(res)

		var apiErr *Error
		c.True(errors.As(err, &apiErr))
		c.Equal(apiErr.Type, ErrorTypeMalformed)
		c.True(apiErr.Meta["err"] != "")
	})
}

// TestDefaultHTTPClientHonorsProxyEnv asserts that defaultHTTPClient's
// Transport consults the HTTPS_PROXY/HTTP_PROXY/NO_PROXY environment via
// http.ProxyFromEnvironment.
//
// Implementation note: an earlier form of this test ran a real httptest proxy
// and set HTTPS_PROXY via t.Setenv. That works in isolation but is flaky
// under the full suite because http.ProxyFromEnvironment memoizes the
// resolved proxy config via sync.Once on first invocation — once another
// test triggers it with an empty env, subsequent t.Setenv calls have no
// effect on proxy resolution. The transport-introspection form below
// asserts the wiring directly (Transport.Proxy is set to the stdlib
// env-aware function), which is a more reliable signal than the
// end-to-end check.
func TestDefaultHTTPClientHonorsProxyEnv(t *testing.T) {
	c := is.New(t)

	client := defaultHTTPClient()
	tr, ok := client.Transport.(*http.Transport)
	c.True(ok)
	c.True(tr.Proxy != nil)

	// Verify it's specifically http.ProxyFromEnvironment (not some other
	// proxy func), comparing function pointers via reflect.
	want := reflect.ValueOf(http.ProxyFromEnvironment).Pointer()
	got := reflect.ValueOf(tr.Proxy).Pointer()
	c.Equal(got, want)
}

