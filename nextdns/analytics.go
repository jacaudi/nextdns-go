package nextdns

const analyticsAPIPath = "analytics"

// AnalyticsOptions contains common parameters for all analytics endpoints.
type AnalyticsOptions struct {
	From   string // Date filter (ISO 8601, Unix timestamp, or relative like "-7d")
	To     string // Date filter
	Limit  int    // Results per page (1-500, default 10)
	Cursor string // Pagination cursor
	Device string // Filter by device ID
}

// AnalyticsTimeSeriesOptions extends AnalyticsOptions with time series parameters.
type AnalyticsTimeSeriesOptions struct {
	AnalyticsOptions
	Interval  string // Window duration ("1h", "1d", or seconds)
	Alignment string // "start", "end", or "clock"
	Timezone  string // IANA timezone (e.g., "America/New_York")
	Partials  string // "none", "start", "end", "all"
}

// AnalyticsEntry represents a single item in analytics responses.
type AnalyticsEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Queries int    `json:"queries"`
}

// AnalyticsTimeSeriesEntry has queries as an array for each time window.
type AnalyticsTimeSeriesEntry struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Queries []int  `json:"queries"`
}

// AnalyticsPagination contains cursor for pagination.
type AnalyticsPagination struct {
	Cursor string `json:"cursor"`
}

// AnalyticsSeriesInfo contains time series metadata.
type AnalyticsSeriesInfo struct {
	Times    []string `json:"times"`
	Interval int      `json:"interval"`
}

// analyticsResponse is the internal response wrapper for standard analytics.
type analyticsResponse struct {
	Data []*AnalyticsEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
	} `json:"meta"`
}

// analyticsTimeSeriesResponse is the internal response wrapper for time series analytics.
type analyticsTimeSeriesResponse struct {
	Data []*AnalyticsTimeSeriesEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
		Series     AnalyticsSeriesInfo `json:"series"`
	} `json:"meta"`
}

// Public response types returned to users

// AnalyticsResponse contains analytics data with pagination info.
type AnalyticsResponse struct {
	Data       []*AnalyticsEntry
	Pagination AnalyticsPagination
}

// AnalyticsTimeSeriesResponse contains time series analytics data.
type AnalyticsTimeSeriesResponse struct {
	Data       []*AnalyticsTimeSeriesEntry
	Pagination AnalyticsPagination
	Series     AnalyticsSeriesInfo
}
