package nextdns

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestLogEntryUnmarshal(t *testing.T) {
	c := is.New(t)

	jsonData := `{
		"timestamp": "2024-01-15T10:30:00.000Z",
		"domain": "example.com",
		"root": "example.com",
		"tracker": "tracker-id",
		"encrypted": true,
		"protocol": "DNS-over-HTTPS",
		"clientIp": "192.168.1.100",
		"client": "client-name",
		"device": {
			"id": "device-1",
			"name": "iPhone",
			"model": "iPhone 15 Pro"
		},
		"status": "blocked",
		"reasons": [
			{"id": "reason-1", "name": "Tracker blocked"}
		]
	}`

	var entry LogEntry
	err := json.Unmarshal([]byte(jsonData), &entry)
	c.NoErr(err)

	c.Equal(entry.Domain, "example.com")
	c.Equal(entry.Root, "example.com")
	c.Equal(entry.Tracker, "tracker-id")
	c.Equal(entry.Encrypted, true)
	c.Equal(entry.Protocol, ProtocolDoH)
	c.Equal(entry.ClientIP, "192.168.1.100")
	c.Equal(entry.Client, "client-name")
	c.Equal(entry.Status, StatusBlocked)
	c.True(entry.Device != nil)
	c.Equal(entry.Device.ID, "device-1")
	c.Equal(entry.Device.Name, "iPhone")
	c.Equal(entry.Device.Model, "iPhone 15 Pro")
	c.Equal(len(entry.Reasons), 1)
	c.Equal(entry.Reasons[0].ID, "reason-1")
	c.Equal(entry.Reasons[0].Name, "Tracker blocked")
}

func TestLogsResponseUnmarshal(t *testing.T) {
	c := is.New(t)

	jsonData := `{
		"data": [
			{
				"timestamp": "2024-01-15T10:30:00.000Z",
				"domain": "example.com",
				"root": "example.com",
				"encrypted": false,
				"protocol": "UDP",
				"clientIp": "10.0.0.1",
				"status": "default"
			}
		],
		"meta": {
			"pagination": {"cursor": "abc123"},
			"stream": {"id": "stream-456"}
		}
	}`

	var resp logsResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	c.NoErr(err)

	c.Equal(len(resp.Data), 1)
	c.Equal(resp.Data[0].Domain, "example.com")
	c.Equal(resp.Meta.Pagination.Cursor, "abc123")
	c.Equal(resp.Meta.Stream.ID, "stream-456")
}

func TestLogsGet(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles/abc123/logs")

		w.WriteHeader(http.StatusOK)
		resp := `{
			"data": [
				{
					"timestamp": "2024-01-15T10:30:00.000Z",
					"domain": "example.com",
					"root": "example.com",
					"encrypted": true,
					"protocol": "DNS-over-HTTPS",
					"clientIp": "192.168.1.100",
					"status": "default"
				}
			],
			"meta": {
				"pagination": {"cursor": "next123"},
				"stream": {"id": "stream456"}
			}
		}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	resp, err := client.Logs.Get(ctx, &GetLogsRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
	c.Equal(len(resp.Data), 1)
	c.Equal(resp.Data[0].Domain, "example.com")
	c.Equal(resp.Data[0].Status, StatusDefault)
	c.Equal(resp.Pagination.Cursor, "next123")
	c.Equal(resp.Stream.ID, "stream456")
}

func TestLogsGetWithOptions(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "GET")
		c.Equal(r.URL.Path, "/profiles/abc123/logs")
		c.Equal(r.URL.Query().Get("from"), "-24h")
		c.Equal(r.URL.Query().Get("status"), "blocked")
		c.Equal(r.URL.Query().Get("limit"), "50")
		c.Equal(r.URL.Query().Get("search"), "example")
		c.Equal(r.URL.Query().Get("raw"), "1")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": [], "meta": {"pagination": {"cursor": ""}, "stream": {"id": ""}}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	_, err = client.Logs.Get(ctx, &GetLogsRequest{
		ProfileID: "abc123",
		Options: &LogsQueryOptions{
			From:   "-24h",
			Status: "blocked",
			Limit:  50,
			Search: "example",
			Raw:    true,
		},
	})

	c.NoErr(err)
}

func TestLogsClear(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Method, "DELETE")
		c.Equal(r.URL.Path, "/profiles/abc123/logs")

		w.WriteHeader(http.StatusOK)
		resp := `{"data": {}}`
		_, err := w.Write([]byte(resp))
		c.NoErr(err)
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.Logs.Clear(ctx, &ClearLogsRequest{
		ProfileID: "abc123",
	})

	c.NoErr(err)
}

func TestLogsDownload(t *testing.T) {
	c := is.New(t)

	// File server simulating the redirect target.
	fileTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("timestamp,domain\n2024-01-01T00:00:00Z,example.com\n"))
	}))
	defer fileTS.Close()

	// API server returning a 302 to the file server.
	apiTS := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/logs/download")
		http.Redirect(w, r, fileTS.URL+"/file.csv", http.StatusFound)
	}))
	defer apiTS.Close()

	client, err := New(WithBaseURL(apiTS.URL))
	c.NoErr(err)

	body, err := client.Logs.Download(context.Background(), &DownloadLogsRequest{ProfileID: "abc123"})
	c.NoErr(err)
	defer func() { _ = body.Close() }()

	data, err := io.ReadAll(body)
	c.NoErr(err)
	c.True(strings.Contains(string(data), "example.com"))
}

func TestLogsDownloadURL(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/logs/download")
		c.Equal(r.URL.Query().Get("redirect"), "0")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"url": "https://files.nextdns.io/abc.csv"}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	resp, err := client.Logs.DownloadURL(context.Background(), &DownloadLogsRequest{ProfileID: "abc123"})
	c.NoErr(err)
	c.Equal(resp.URL, "https://files.nextdns.io/abc.csv")
}

func TestLogsStream(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.URL.Path, "/profiles/abc123/logs/stream")
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)

		// Two SSE events.
		_, _ = w.Write([]byte("id: 64v32d9r6rwkcctg6cu38e9g60\n"))
		_, _ = w.Write([]byte(`data: {"timestamp":"2024-01-01T00:00:00Z","domain":"example.com","root":"example.com","encrypted":true,"protocol":"DNS-over-HTTPS","clientIp":"203.0.113.1","status":"default"}`))
		_, _ = w.Write([]byte("\n\n"))
		flusher.Flush()

		_, _ = w.Write([]byte("id: 64v32d9r6rwkcctg6cu38e9g61\n"))
		_, _ = w.Write([]byte(`data: {"timestamp":"2024-01-01T00:00:01Z","domain":"test.com","root":"test.com","encrypted":true,"protocol":"DNS-over-HTTPS","clientIp":"203.0.113.1","status":"blocked"}`))
		_, _ = w.Write([]byte("\n\n"))
		flusher.Flush()
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got []*LogEntry
	for entry, err := range client.Logs.Stream(ctx, &StreamLogsRequest{ProfileID: "abc123"}) {
		if err != nil {
			break // EOF
		}
		got = append(got, entry)
		if len(got) == 2 {
			break
		}
	}

	c.Equal(len(got), 2)
	c.Equal(got[0].Domain, "example.com")
	c.Equal(got[1].Status, StatusBlocked)
}

func TestLogsDownloadInvokesRateLimitObserver(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "600")
		w.Header().Set("X-RateLimit-Remaining", "599")
		w.Header().Set("X-RateLimit-Reset", "1730000000")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("timestamp,domain\n"))
	}))
	defer ts.Close()

	var seen RateLimit
	client, err := New(
		WithBaseURL(ts.URL),
		WithRateLimitObserver(func(rl RateLimit) { seen = rl }),
	)
	c.NoErr(err)

	body, err := client.Logs.Download(context.Background(), &DownloadLogsRequest{ProfileID: "abc"})
	c.NoErr(err)
	defer func() { _ = body.Close() }()
	_, _ = io.ReadAll(body)

	c.Equal(seen.Limit, 600)
	c.Equal(seen.Remaining, 599)
}

func TestLogsDownloadReturnsTypedErrorOn429(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"code":"rate_limited"}]}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	_, err = client.Logs.Download(context.Background(), &DownloadLogsRequest{ProfileID: "abc"})
	c.True(err != nil)

	var apiErr *Error
	c.True(errors.As(err, &apiErr)) // typed; not a bare fmt.Errorf
}
