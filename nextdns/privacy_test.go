package nextdns

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestPrivacyGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"blocklists": [
					{"id": "nextdns-recommended", "name": "NextDNS Ads & Trackers Blocklist"},
					{"id": "easylist", "name": "EasyList"}
				],
				"natives": [
					{"id": "apple"},
					{"id": "windows"},
					{"id": "samsung"}
				],
				"disguisedTrackers": true,
				"allowAffiliate": false
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	privacy, err := client.Privacy.Get(ctx, &GetPrivacyRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(privacy.DisguisedTrackers, true)
	c.Equal(privacy.AllowAffiliate, false)

	c.Equal(len(privacy.Blocklists), 2)
	c.Equal(privacy.Blocklists[0].ID, "nextdns-recommended")
	c.Equal(privacy.Blocklists[0].Name, "NextDNS Ads & Trackers Blocklist")
	c.Equal(privacy.Blocklists[1].ID, "easylist")

	c.Equal(len(privacy.Natives), 3)
	c.Equal(privacy.Natives[0].ID, "apple")
	c.Equal(privacy.Natives[1].ID, "windows")
	c.Equal(privacy.Natives[2].ID, "samsung")
}

func TestPrivacyUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":{}}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &UpdatePrivacyRequest{
		ProfileID: "abc123",
		Privacy: &Privacy{
			DisguisedTrackers: true,
			AllowAffiliate:    true,
		},
	}

	err = client.Privacy.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["disguisedTrackers"], true)
	c.Equal(receivedBody["allowAffiliate"], true)
}

func TestPrivacyBlocklistsList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/blocklists")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{
					"id": "nextdns-recommended",
					"name": "NextDNS Ads & Trackers Blocklist",
					"website": "https://nextdns.io",
					"entries": 150000,
					"updatedOn": "2024-01-15T10:30:00Z"
				},
				{
					"id": "easylist",
					"name": "EasyList",
					"website": "https://easylist.to",
					"entries": 75000,
					"updatedOn": "2024-01-14T08:00:00Z"
				}
			]
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	blocklists, err := client.PrivacyBlocklists.List(ctx, &ListPrivacyBlocklistsRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(blocklists), 2)

	c.Equal(blocklists[0].ID, "nextdns-recommended")
	c.Equal(blocklists[0].Name, "NextDNS Ads & Trackers Blocklist")
	c.Equal(blocklists[0].Website, "https://nextdns.io")
	c.Equal(blocklists[0].Entries, 150000)
	c.True(blocklists[0].UpdatedOn != nil)

	expectedTime, _ := time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")
	c.Equal(blocklists[0].UpdatedOn.UTC(), expectedTime)
}

func TestPrivacyBlocklistsCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody []interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPut)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/blocklists")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreatePrivacyBlocklistsRequest{
		ProfileID: "abc123",
		PrivacyBlocklists: []*PrivacyBlocklists{
			{ID: "nextdns-recommended"},
			{ID: "easylist"},
			{ID: "adguard-dns"},
		},
	}

	err = client.PrivacyBlocklists.Create(ctx, request)
	c.NoErr(err)

	c.Equal(len(receivedBody), 3)

	first, ok := receivedBody[0].(map[string]interface{})
	c.True(ok)
	c.Equal(first["id"], "nextdns-recommended")

	second, ok := receivedBody[1].(map[string]interface{})
	c.True(ok)
	c.Equal(second["id"], "easylist")
}

func TestPrivacyNativesList(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/natives")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": [
				{"id": "apple"},
				{"id": "windows"},
				{"id": "huawei"},
				{"id": "samsung"},
				{"id": "xiaomi"},
				{"id": "amazon"},
				{"id": "roku"},
				{"id": "sonos"}
			]
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	natives, err := client.PrivacyNatives.List(ctx, &ListPrivacyNativesRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(len(natives), 8)
	c.Equal(natives[0].ID, "apple")
	c.Equal(natives[1].ID, "windows")
	c.Equal(natives[2].ID, "huawei")
}

func TestPrivacyNativesCreate(t *testing.T) {
	c := is.New(t)

	var receivedBody []interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPut)
		c.Equal(r.URL.Path, "/profiles/abc123/privacy/natives")

		body, err := io.ReadAll(r.Body)
		c.NoErr(err)
		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"data":[]}`))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	request := &CreatePrivacyNativesRequest{
		ProfileID: "abc123",
		PrivacyNatives: []*PrivacyNatives{
			{ID: "apple"},
			{ID: "windows"},
			{ID: "samsung"},
		},
	}

	err = client.PrivacyNatives.Create(ctx, request)
	c.NoErr(err)

	c.Equal(len(receivedBody), 3)

	first, ok := receivedBody[0].(map[string]interface{})
	c.True(ok)
	c.Equal(first["id"], "apple")
}

func TestPrivacyJSONFieldNames(t *testing.T) {
	c := is.New(t)

	privacy := &Privacy{
		DisguisedTrackers: true,
		AllowAffiliate:    false,
		Blocklists: []*PrivacyBlocklists{
			{ID: "test-blocklist"},
		},
		Natives: []*PrivacyNatives{
			{ID: "apple"},
		},
	}

	data, err := json.Marshal(privacy)
	c.NoErr(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	c.NoErr(err)

	_, ok := result["disguisedTrackers"]
	c.True(ok)
	_, ok = result["allowAffiliate"]
	c.True(ok)
	_, ok = result["blocklists"]
	c.True(ok)
	_, ok = result["natives"]
	c.True(ok)
}
