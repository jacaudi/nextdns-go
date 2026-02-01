package nextdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestPrivacyNativesAdd(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "POST")
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/natives")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {"id": "apple"}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.PrivacyNatives.Add(ctx, &AddPrivacyNativesRequest{
		ProfileID: "abc123",
		ID:        "apple",
	})

	c.NoErr(err)
}

func TestPrivacyNativesUpdate(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "PATCH")
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/natives/apple")

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
	err = client.PrivacyNatives.Update(ctx, &UpdatePrivacyNativesRequest{
		ProfileID: "abc123",
		NativeID:  "apple",
		Active:    &active,
	})

	c.NoErr(err)
}
