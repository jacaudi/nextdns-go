package nextdns

import (
	"encoding/json"
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
	c.Equal(entry.Protocol, "DNS-over-HTTPS")
	c.Equal(entry.ClientIP, "192.168.1.100")
	c.Equal(entry.Client, "client-name")
	c.Equal(entry.Status, "blocked")
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
