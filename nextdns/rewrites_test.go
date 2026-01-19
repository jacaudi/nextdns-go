package nextdns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestRewritesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/rewrites")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{"id": "rewrite-1", "name": "local.example.com", "type": "A", "content": "192.168.1.1"},
				{"id": "rewrite-2", "name": "mail.example.com", "type": "CNAME", "content": "mail.provider.com"},
				{"id": "rewrite-3", "name": "ipv6.example.com", "type": "AAAA", "content": "2001:db8::1"}
			]
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	rewrites, err := client.Rewrites.List(ctx, &ListRewritesRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(rewrites), 3)

	c.Equal(rewrites[0].ID, "rewrite-1")
	c.Equal(rewrites[0].Name, "local.example.com")
	c.Equal(rewrites[0].Type, "A")
	c.Equal(rewrites[0].Content, "192.168.1.1")

	c.Equal(rewrites[1].ID, "rewrite-2")
	c.Equal(rewrites[1].Name, "mail.example.com")
	c.Equal(rewrites[1].Type, "CNAME")
	c.Equal(rewrites[1].Content, "mail.provider.com")

	c.Equal(rewrites[2].ID, "rewrite-3")
	c.Equal(rewrites[2].Type, "AAAA")
	c.Equal(rewrites[2].Content, "2001:db8::1")
}

func TestRewritesCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPost)
		c.Equal(r.URL.Path, "/profiles/abc123/rewrites")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{"id":"new-rewrite-id","name":"new.example.com","type":"A","content":"10.0.0.1"}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateRewritesRequest{
		ProfileID: "abc123",
		Rewrites: &Rewrites{
			Name:    "new.example.com",
			Type:    "A",
			Content: "10.0.0.1",
		},
	}

	id, err := client.Rewrites.Create(ctx, request)
	c.NoErr(err)
	c.Equal(id, "new-rewrite-id")

	c.Equal(receivedBody["name"], "new.example.com")
	c.Equal(receivedBody["type"], "A")
	c.Equal(receivedBody["content"], "10.0.0.1")
}

func TestRewritesCreateCNAME(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{"id":"cname-rewrite","name":"alias.example.com","type":"CNAME","content":"target.example.com"}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateRewritesRequest{
		ProfileID: "abc123",
		Rewrites: &Rewrites{
			Name:    "alias.example.com",
			Type:    "CNAME",
			Content: "target.example.com",
		},
	}

	id, err := client.Rewrites.Create(ctx, request)
	c.NoErr(err)
	c.Equal(id, "cname-rewrite")

	c.Equal(receivedBody["name"], "alias.example.com")
	c.Equal(receivedBody["type"], "CNAME")
	c.Equal(receivedBody["content"], "target.example.com")
}

func TestRewritesDelete(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodDelete)
		c.Equal(r.URL.Path, "/profiles/abc123/rewrites/rewrite-to-delete")

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.Rewrites.Delete(ctx, &DeleteRewritesRequest{
		ProfileID: "abc123",
		ID:        "rewrite-to-delete",
	})
	c.NoErr(err)
}

func TestRewritesJSONFieldNames(t *testing.T) {
	c := is.New(t)

	rewrite := &Rewrites{
		ID:      "test-id",
		Name:    "test.example.com",
		Type:    "A",
		Content: "1.2.3.4",
	}

	data, err := json.Marshal(rewrite)
	c.NoErr(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	c.NoErr(err)

	_, ok := result["id"]
	c.True(ok)
	_, ok = result["name"]
	c.True(ok)
	_, ok = result["type"]
	c.True(ok)
	_, ok = result["content"]
	c.True(ok)
}

func TestRewritesEmptyList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	rewrites, err := client.Rewrites.List(ctx, &ListRewritesRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(rewrites), 0)
}

func TestRewritesWithWildcard(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{"id":"wildcard-rewrite","name":"*.internal.example.com","type":"A","content":"10.0.0.100"}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreateRewritesRequest{
		ProfileID: "abc123",
		Rewrites: &Rewrites{
			Name:    "*.internal.example.com",
			Type:    "A",
			Content: "10.0.0.100",
		},
	}

	id, err := client.Rewrites.Create(ctx, request)
	c.NoErr(err)
	c.Equal(id, "wildcard-rewrite")

	c.Equal(receivedBody["name"], "*.internal.example.com")
}
