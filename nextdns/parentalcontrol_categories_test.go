package nextdns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestParentalControlCategoriesDelete(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "DELETE")
		c.Equal(r.URL.Path, "/profiles/abc123/parentalControl/categories/gambling")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	err = client.ParentalControlCategories.Delete(context.Background(), &DeleteParentalControlCategoriesRequest{
		ProfileID: "abc123",
		ID:        "gambling",
	})
	c.NoErr(err)
}
