package nextdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

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

// buildAnalyticsQuery converts AnalyticsOptions to url.Values.
func buildAnalyticsQuery(opts *AnalyticsOptions) url.Values {
	query := url.Values{}
	if opts == nil {
		return query
	}
	if opts.From != "" {
		query.Set("from", opts.From)
	}
	if opts.To != "" {
		query.Set("to", opts.To)
	}
	if opts.Limit > 0 {
		query.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		query.Set("cursor", opts.Cursor)
	}
	if opts.Device != "" {
		query.Set("device", opts.Device)
	}
	return query
}

// buildTimeSeriesQuery adds time series parameters to the query.
func buildTimeSeriesQuery(opts *AnalyticsTimeSeriesOptions) url.Values {
	if opts == nil {
		return url.Values{}
	}
	query := buildAnalyticsQuery(&opts.AnalyticsOptions)
	if opts.Interval != "" {
		query.Set("interval", opts.Interval)
	}
	if opts.Alignment != "" {
		query.Set("alignment", opts.Alignment)
	}
	if opts.Timezone != "" {
		query.Set("timezone", opts.Timezone)
	}
	if opts.Partials != "" {
		query.Set("partials", opts.Partials)
	}
	return query
}

func analyticsPath(profileID, endpoint string) string {
	return fmt.Sprintf("%s/%s/%s/%s", profilesAPIPath, profileID, analyticsAPIPath, endpoint)
}

// GetStatus returns query counts by resolution status.
func (s *analyticsService) GetStatus(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "status")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics status: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics status: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetStatusSeries returns query counts by resolution status as time series.
func (s *analyticsService) GetStatusSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "status;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics status series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics status series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetDomains returns top queried domains.
func (s *analyticsService) GetDomains(ctx context.Context, request *GetAnalyticsDomainsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "domains")
	query := buildAnalyticsQuery(request.Options)
	if request.Status != "" {
		query.Set("status", request.Status)
	}
	if request.Root {
		query.Set("root", "true")
	}

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics domains: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics domains: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetDomainsSeries returns top queried domains as time series.
func (s *analyticsService) GetDomainsSeries(ctx context.Context, request *GetAnalyticsDomainsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "domains;series")
	query := buildTimeSeriesQuery(request.Options)
	if request.Status != "" {
		query.Set("status", request.Status)
	}
	if request.Root {
		query.Set("root", "true")
	}

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics domains series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics domains series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetDevices returns connected devices and query distribution.
func (s *analyticsService) GetDevices(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "devices")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics devices: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics devices: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetDevicesSeries returns connected devices and query distribution as time series.
func (s *analyticsService) GetDevicesSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "devices;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics devices series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics devices series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetDestinations returns queries by country or GAFAM company.
func (s *analyticsService) GetDestinations(ctx context.Context, request *GetAnalyticsDestinationsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "destinations")
	query := buildAnalyticsQuery(request.Options)
	if request.Type != "" {
		query.Set("type", request.Type)
	}

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics destinations: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics destinations: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetDestinationsSeries returns queries by country or GAFAM company as time series.
func (s *analyticsService) GetDestinationsSeries(ctx context.Context, request *GetAnalyticsDestinationsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "destinations;series")
	query := buildTimeSeriesQuery(request.Options)
	if request.Type != "" {
		query.Set("type", request.Type)
	}

	req, err := s.client.newRequestWithQuery(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics destinations series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics destinations series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}
