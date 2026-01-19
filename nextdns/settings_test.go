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

func TestSettingsGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/settings")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"logs": {
					"enabled": true,
					"drop": {
						"ip": true,
						"domain": false
					},
					"retention": 90,
					"location": "eu"
				},
				"blockPage": {
					"enabled": true
				},
				"performance": {
					"ecs": true,
					"cacheBoost": true,
					"cnameFlattening": false
				},
				"web3": true
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	settings, err := client.Settings.Get(ctx, &GetSettingsRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(settings.Web3, true)

	c.True(settings.Logs != nil)
	c.Equal(settings.Logs.Enabled, true)
	c.Equal(settings.Logs.Retention, 90)
	c.Equal(settings.Logs.Location, "eu")
	c.True(settings.Logs.Drop != nil)
	c.Equal(settings.Logs.Drop.IP, true)
	c.Equal(settings.Logs.Drop.Domain, false)

	c.True(settings.BlockPage != nil)
	c.Equal(settings.BlockPage.Enabled, true)

	c.True(settings.Performance != nil)
	c.Equal(settings.Performance.Ecs, true)
	c.Equal(settings.Performance.CacheBoost, true)
	c.Equal(settings.Performance.CnameFlattening, false)
}

func TestSettingsUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/settings")

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
	request := &UpdateSettingsRequest{
		ProfileID: "abc123",
		Settings: &Settings{
			Web3: false,
			Logs: &SettingsLogs{
				Enabled:   true,
				Retention: 30,
				Location:  "us",
			},
		},
	}

	err = client.Settings.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["web3"], false)

	logs, ok := receivedBody["logs"].(map[string]interface{})
	c.True(ok)
	c.Equal(logs["enabled"], true)
	c.Equal(logs["retention"], float64(30))
	c.Equal(logs["location"], "us")
}

func TestSettingsLogsGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/logs")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"enabled": true,
				"drop": {
					"ip": true,
					"domain": true
				},
				"retention": 7,
				"location": "ch"
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	logs, err := client.SettingsLogs.Get(ctx, &GetSettingsLogsRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(logs.Enabled, true)
	c.Equal(logs.Retention, 7)
	c.Equal(logs.Location, "ch")
	c.True(logs.Drop != nil)
	c.Equal(logs.Drop.IP, true)
	c.Equal(logs.Drop.Domain, true)
}

func TestSettingsLogsUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/logs")

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
	request := &UpdateSettingsLogsRequest{
		ProfileID: "abc123",
		SettingsLogs: &SettingsLogs{
			Enabled:   false,
			Retention: 14,
			Location:  "us",
			Drop: &SettingsLogsDrop{
				IP:     true,
				Domain: false,
			},
		},
	}

	err = client.SettingsLogs.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["enabled"], false)
	c.Equal(receivedBody["retention"], float64(14))
	c.Equal(receivedBody["location"], "us")

	drop, ok := receivedBody["drop"].(map[string]interface{})
	c.True(ok)
	c.Equal(drop["ip"], true)
	c.Equal(drop["domain"], false)
}

func TestSettingsBlockPageGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/blockPage")

		w.WriteHeader(http.StatusOK)
		out := `{"data":{"enabled":true}}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	blockPage, err := client.SettingsBlockPage.Get(ctx, &GetSettingsBlockPageRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(blockPage.Enabled, true)
}

func TestSettingsBlockPageUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/blockPage")

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
	request := &UpdateSettingsBlockPageRequest{
		ProfileID: "abc123",
		SettingsBlockPage: &SettingsBlockPage{
			Enabled: false,
		},
	}

	err = client.SettingsBlockPage.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["enabled"], false)
}

func TestSettingsPerformanceGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodGet)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/performance")

		w.WriteHeader(http.StatusOK)
		out := `{
			"data": {
				"ecs": true,
				"cacheBoost": true,
				"cnameFlattening": true
			}
		}`
		_, err := w.Write([]byte(out))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	performance, err := client.SettingsPerformance.Get(ctx, &GetSettingsPerformanceRequest{ProfileID: "abc123"})
	c.NoErr(err)

	c.Equal(performance.Ecs, true)
	c.Equal(performance.CacheBoost, true)
	c.Equal(performance.CnameFlattening, true)
}

func TestSettingsPerformanceUpdate(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, http.MethodPatch)
		c.Equal(r.URL.Path, "/profiles/abc123/settings/performance")

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
	request := &UpdateSettingsPerformanceRequest{
		ProfileID: "abc123",
		SettingsPerformance: &SettingsPerformance{
			Ecs:             false,
			CacheBoost:      true,
			CnameFlattening: true,
		},
	}

	err = client.SettingsPerformance.Update(ctx, request)
	c.NoErr(err)

	c.Equal(receivedBody["ecs"], false)
	c.Equal(receivedBody["cacheBoost"], true)
	c.Equal(receivedBody["cnameFlattening"], true)
}

func TestSettingsJSONFieldNames(t *testing.T) {
	c := is.New(t)

	settings := &Settings{
		Web3: true,
		Logs: &SettingsLogs{
			Enabled:   true,
			Retention: 30,
			Location:  "us",
			Drop: &SettingsLogsDrop{
				IP:     true,
				Domain: false,
			},
		},
		BlockPage: &SettingsBlockPage{
			Enabled: true,
		},
		Performance: &SettingsPerformance{
			Ecs:             true,
			CacheBoost:      true,
			CnameFlattening: false,
		},
	}

	data, err := json.Marshal(settings)
	c.NoErr(err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	c.NoErr(err)

	_, ok := result["web3"]
	c.True(ok)
	_, ok = result["logs"]
	c.True(ok)
	_, ok = result["blockPage"]
	c.True(ok)
	_, ok = result["performance"]
	c.True(ok)

	logs, ok := result["logs"].(map[string]interface{})
	c.True(ok)
	_, ok = logs["enabled"]
	c.True(ok)
	_, ok = logs["retention"]
	c.True(ok)
	_, ok = logs["location"]
	c.True(ok)
	_, ok = logs["drop"]
	c.True(ok)

	performance, ok := result["performance"].(map[string]interface{})
	c.True(ok)
	_, ok = performance["ecs"]
	c.True(ok)
	_, ok = performance["cacheBoost"]
	c.True(ok)
	_, ok = performance["cnameFlattening"]
	c.True(ok)
}
