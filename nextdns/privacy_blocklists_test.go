package nextdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestPrivacyBlocklistsAdd(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "POST")
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/blocklists")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {"id": "nextdns-recommended"}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.PrivacyBlocklists.Add(ctx, &AddPrivacyBlocklistsRequest{
		ProfileID: "abc123",
		ID:        "nextdns-recommended",
	})

	c.NoErr(err)
}

func TestPrivacyBlocklistsUpdate(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "PATCH")
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/blocklists/nextdns-recommended")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	active := false
	err = client.PrivacyBlocklists.Update(ctx, &UpdatePrivacyBlocklistsRequest{
		ProfileID:   "abc123",
		BlocklistID: "nextdns-recommended",
		Active:      &active,
	})

	c.NoErr(err)
}
