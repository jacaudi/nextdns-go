package nextdns

import "context"

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

// Request types for analytics endpoints

// GetAnalyticsRequest is used for status and devices endpoints.
type GetAnalyticsRequest struct {
	ProfileID string
	Options   *AnalyticsOptions
}

// GetAnalyticsTimeSeriesRequest is used for status and devices time series.
type GetAnalyticsTimeSeriesRequest struct {
	ProfileID string
	Options   *AnalyticsTimeSeriesOptions
}

// GetAnalyticsDomainsRequest includes domain-specific filters.
type GetAnalyticsDomainsRequest struct {
	ProfileID string
	Options   *AnalyticsOptions
	Status    string // Filter: "default", "blocked", "allowed"
	Root      bool   // Aggregate by root domain
}

// GetAnalyticsDomainsTimeSeriesRequest includes domain-specific filters for time series.
type GetAnalyticsDomainsTimeSeriesRequest struct {
	ProfileID string
	Options   *AnalyticsTimeSeriesOptions
	Status    string
	Root      bool
}

// GetAnalyticsDestinationsRequest requires a type parameter.
type GetAnalyticsDestinationsRequest struct {
	ProfileID string
	Options   *AnalyticsOptions
	Type      string // Required: "countries" or "gafam"
}

// GetAnalyticsDestinationsTimeSeriesRequest requires a type parameter.
type GetAnalyticsDestinationsTimeSeriesRequest struct {
	ProfileID string
	Options   *AnalyticsTimeSeriesOptions
	Type      string
}

// AnalyticsService provides access to NextDNS analytics data.
type AnalyticsService interface {
	// Status returns query counts by resolution status (default, blocked, allowed).
	GetStatus(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error)
	GetStatusSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// Domains returns top queried domains.
	GetDomains(ctx context.Context, request *GetAnalyticsDomainsRequest) (*AnalyticsResponse, error)
	GetDomainsSeries(ctx context.Context, request *GetAnalyticsDomainsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// Devices returns connected devices and query distribution.
	GetDevices(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error)
	GetDevicesSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// Destinations returns queries by country or GAFAM company.
	GetDestinations(ctx context.Context, request *GetAnalyticsDestinationsRequest) (*AnalyticsResponse, error)
	GetDestinationsSeries(ctx context.Context, request *GetAnalyticsDestinationsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)
}

type analyticsService struct {
	client *Client
}

// Compile-time check that analyticsService implements AnalyticsService.
var _ AnalyticsService = &analyticsService{}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(client *Client) *analyticsService {
	return &analyticsService{
		client: client,
	}
}

// GetStatus returns query counts by resolution status.
func (s *analyticsService) GetStatus(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	// TODO: Implement in Task 5
	return nil, nil
}

// GetStatusSeries returns query counts by resolution status as time series.
func (s *analyticsService) GetStatusSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	// TODO: Implement in Task 6
	return nil, nil
}

// GetDomains returns top queried domains.
func (s *analyticsService) GetDomains(ctx context.Context, request *GetAnalyticsDomainsRequest) (*AnalyticsResponse, error) {
	// TODO: Implement in Task 7
	return nil, nil
}

// GetDomainsSeries returns top queried domains as time series.
func (s *analyticsService) GetDomainsSeries(ctx context.Context, request *GetAnalyticsDomainsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	// TODO: Implement in Task 7
	return nil, nil
}

// GetDevices returns connected devices and query distribution.
func (s *analyticsService) GetDevices(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	// TODO: Implement in Task 8
	return nil, nil
}

// GetDevicesSeries returns connected devices and query distribution as time series.
func (s *analyticsService) GetDevicesSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	// TODO: Implement in Task 8
	return nil, nil
}

// GetDestinations returns queries by country or GAFAM company.
func (s *analyticsService) GetDestinations(ctx context.Context, request *GetAnalyticsDestinationsRequest) (*AnalyticsResponse, error) {
	// TODO: Implement in Task 9
	return nil, nil
}

// GetDestinationsSeries returns queries by country or GAFAM company as time series.
func (s *analyticsService) GetDestinationsSeries(ctx context.Context, request *GetAnalyticsDestinationsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	// TODO: Implement in Task 9
	return nil, nil
}
