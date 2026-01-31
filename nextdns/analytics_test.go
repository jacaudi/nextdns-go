package nextdns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestAnalyticsResponseUnmarshal(t *testing.T) {
	c := is.New(t)

	jsonData := `{
		"data": [
			{"id": "default", "queries": 1000},
			{"id": "blocked", "name": "Blocked", "queries": 50}
		],
		"meta": {
			"pagination": {"cursor": "abc123"}
		}
	}`

	var resp analyticsResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	c.NoErr(err)

	c.Equal(len(resp.Data), 2)
	c.Equal(resp.Data[0].ID, "default")
	c.Equal(resp.Data[0].Queries, 1000)
	c.Equal(resp.Data[1].Name, "Blocked")
	c.Equal(resp.Meta.Pagination.Cursor, "abc123")
}

func TestAnalyticsTimeSeriesResponseUnmarshal(t *testing.T) {
	c := is.New(t)

	jsonData := `{
		"data": [
			{"id": "default", "queries": [100, 150, 200]}
		],
		"meta": {
			"pagination": {"cursor": ""},
			"series": {
				"times": ["2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "2024-01-01T02:00:00Z"],
				"interval": 3600
			}
		}
	}`

	var resp analyticsTimeSeriesResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	c.NoErr(err)

	c.Equal(len(resp.Data), 1)
	c.Equal(resp.Data[0].ID, "default")
	c.Equal(len(resp.Data[0].Queries), 3)
	c.Equal(resp.Data[0].Queries[0], 100)
	c.Equal(resp.Meta.Series.Interval, 3600)
	c.Equal(len(resp.Meta.Series.Times), 3)
}

func TestAnalyticsGetStatus(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/status")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{"id": "default", "queries": 1000},
				{"id": "blocked", "queries": 50},
				{"id": "allowed", "queries": 25}
			],
			"meta": {"pagination": {"cursor": ""}}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetStatus(ctx, &GetAnalyticsRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 3)
	c.Equal(resp.Data[0].ID, "default")
	c.Equal(resp.Data[0].Queries, 1000)
}

func TestAnalyticsGetStatusWithOptions(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/status")
		c.Equal(r.URL.Query().Get("from"), "-7d")
		c.Equal(r.URL.Query().Get("limit"), "100")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": [], "meta": {"pagination": {"cursor": ""}}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	_, err = client.Analytics.GetStatus(ctx, &GetAnalyticsRequest{
		ProfileID: "abc123",
		Options: &AnalyticsOptions{
			From:  "-7d",
			Limit: 100,
		},
	})

	c.NoErr(err)
}

func TestAnalyticsGetStatusSeries(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/status;series")
		c.Equal(r.URL.Query().Get("interval"), "1h")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{"id": "default", "queries": [100, 150, 200]},
				{"id": "blocked", "queries": [10, 15, 20]}
			],
			"meta": {
				"pagination": {"cursor": ""},
				"series": {
					"times": ["2024-01-01T00:00:00Z", "2024-01-01T01:00:00Z", "2024-01-01T02:00:00Z"],
					"interval": 3600
				}
			}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetStatusSeries(ctx, &GetAnalyticsTimeSeriesRequest{
		ProfileID: "abc123",
		Options: &AnalyticsTimeSeriesOptions{
			Interval: "1h",
		},
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 2)
	c.Equal(resp.Data[0].ID, "default")
	c.Equal(len(resp.Data[0].Queries), 3)
	c.Equal(resp.Series.Interval, 3600)
}

func TestAnalyticsGetDomains(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/domains")
		c.Equal(r.URL.Query().Get("status"), "blocked")
		c.Equal(r.URL.Query().Get("root"), "true")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{"id": "example.com", "queries": 500},
				{"id": "test.com", "queries": 300}
			],
			"meta": {"pagination": {"cursor": ""}}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDomains(ctx, &GetAnalyticsDomainsRequest{
		ProfileID: "abc123",
		Status:    "blocked",
		Root:      true,
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 2)
	c.Equal(resp.Data[0].ID, "example.com")
}

func TestAnalyticsGetDomainsSeries(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/domains;series")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [{"id": "example.com", "queries": [10, 20, 30]}],
			"meta": {
				"pagination": {"cursor": ""},
				"series": {"times": ["2024-01-01T00:00:00Z"], "interval": 3600}
			}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDomainsSeries(ctx, &GetAnalyticsDomainsTimeSeriesRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 1)
}

func TestAnalyticsGetDevices(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/devices")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{"id": "device-1", "name": "iPhone", "queries": 500},
				{"id": "device-2", "name": "MacBook", "queries": 300}
			],
			"meta": {"pagination": {"cursor": ""}}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDevices(ctx, &GetAnalyticsRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 2)
	c.Equal(resp.Data[0].Name, "iPhone")
}

func TestAnalyticsGetDevicesSeries(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/devices;series")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [{"id": "device-1", "name": "iPhone", "queries": [100, 200]}],
			"meta": {
				"pagination": {"cursor": ""},
				"series": {"times": ["2024-01-01T00:00:00Z"], "interval": 3600}
			}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDevicesSeries(ctx, &GetAnalyticsTimeSeriesRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 1)
}

func TestAnalyticsGetDestinations(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/destinations")
		c.Equal(r.URL.Query().Get("type"), "countries")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{"id": "US", "name": "United States", "queries": 5000},
				{"id": "DE", "name": "Germany", "queries": 1000}
			],
			"meta": {"pagination": {"cursor": ""}}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDestinations(ctx, &GetAnalyticsDestinationsRequest{
		ProfileID: "abc123",
		Type:      "countries",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 2)
	c.Equal(resp.Data[0].ID, "US")
	c.Equal(resp.Data[0].Name, "United States")
}

func TestAnalyticsGetDestinationsSeries(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/analytics/destinations;series")
		c.Equal(r.URL.Query().Get("type"), "gafam")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [{"id": "google", "name": "Google", "queries": [100, 200]}],
			"meta": {
				"pagination": {"cursor": ""},
				"series": {"times": ["2024-01-01T00:00:00Z"], "interval": 3600}
			}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Analytics.GetDestinationsSeries(ctx, &GetAnalyticsDestinationsTimeSeriesRequest{
		ProfileID: "abc123",
		Type:      "gafam",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 1)
	c.Equal(resp.Data[0].Name, "Google")
}
