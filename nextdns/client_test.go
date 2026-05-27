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
