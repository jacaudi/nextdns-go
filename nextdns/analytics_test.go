package nextdns

import (
	"encoding/json"
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
