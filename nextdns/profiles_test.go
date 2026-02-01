package nextdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestProfilesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles")
		c.Equal(r.URL.Query().Get("cursor"), "")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": [{"id": "abc123", "fingerprint": "fp123", "name": "Profile 1"}], "meta": {"pagination": {"cursor": "next_cursor"}}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	response, err := client.Profiles.List(ctx, &ListProfileRequest{})

	c.NoErr(err)
	c.Equal(len(response.Profiles), 1)
	c.Equal(response.Profiles[0].ID, "abc123")
	c.Equal(response.Profiles[0].Name, "Profile 1")
	c.Equal(response.Cursor, "next_cursor")
}

func TestProfilesListWithCursor(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles")
		c.Equal(r.URL.Query().Get("cursor"), "page2_cursor")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": [{"id": "def456", "fingerprint": "fp456", "name": "Profile 2"}], "meta": {"pagination": {"cursor": ""}}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	response, err := client.Profiles.List(ctx, &ListProfileRequest{
		Cursor: "page2_cursor",
	})

	c.NoErr(err)
	c.Equal(len(response.Profiles), 1)
	c.Equal(response.Profiles[0].ID, "def456")
	c.Equal(response.Cursor, "") // Empty cursor means no more pages
}

func TestProfilesListNilRequest(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": [], "meta": {"pagination": {"cursor": ""}}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	response, err := client.Profiles.List(ctx, nil)

	c.NoErr(err)
	c.Equal(len(response.Profiles), 0)
	c.Equal(response.Cursor, "")
}
