package nextdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestSecurityTldsAdd(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "POST")
		c.Equal(r.URL.Path, "/profiles/abc123/security/tlds")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {"id": "xyz"}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.SecurityTlds.Add(ctx, &AddSecurityTldsRequest{
		ProfileID: "abc123",
		ID:        "xyz",
	})

	c.NoErr(err)
}

func TestSecurityTldsUpdate(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "PATCH")
		c.Equal(r.URL.Path, "/profiles/abc123/security/tlds/xyz")

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
	err = client.SecurityTlds.Update(ctx, &UpdateSecurityTldsRequest{
		ProfileID: "abc123",
		TldID:     "xyz",
		Active:    &active,
	})

	c.NoErr(err)
}

func TestSecurityTldsDelete(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "DELETE")
		c.Equal(r.URL.Path, "/profiles/abc123/security/tlds/xyz")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.SecurityTlds.Delete(ctx, &DeleteSecurityTldsRequest{
		ProfileID: "abc123",
		TldID:     "xyz",
	})

	c.NoErr(err)
}
