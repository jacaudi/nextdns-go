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
	Limit  int    // Results per page (1-1000 per OpenAPI spec; SDK does not enforce)
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

	// QueryTypes returns counts by DNS query type (A, AAAA, etc.).
	GetQueryTypes(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error)
	GetQueryTypesSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// Reasons returns counts by block reason (e.g. blocklist name).
	GetReasons(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error)
	GetReasonsSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// IPs returns counts by client IP.
	GetIPs(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error)
	GetIPsSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error)

	// DNSSEC returns counts by DNSSEC validation status.
	GetDNSSEC(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsDNSSECResponse, error)
	GetDNSSECSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsDNSSECTimeSeriesResponse, error)

	// Encryption returns counts by query encryption status.
	GetEncryption(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsEncryptionResponse, error)
	GetEncryptionSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsEncryptionTimeSeriesResponse, error)
}

// AnalyticsDNSSECEntry is one row of the dnssec analytics response.
type AnalyticsDNSSECEntry struct {
	Validated bool `json:"validated"`
	Queries   int  `json:"queries"`
}

// AnalyticsDNSSECTimeSeriesEntry has queries as an array.
type AnalyticsDNSSECTimeSeriesEntry struct {
	Validated bool  `json:"validated"`
	Queries   []int `json:"queries"`
}

type analyticsDNSSECResponse struct {
	Data []*AnalyticsDNSSECEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
	} `json:"meta"`
}

type analyticsDNSSECTimeSeriesResponse struct {
	Data []*AnalyticsDNSSECTimeSeriesEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
		Series     AnalyticsSeriesInfo `json:"series"`
	} `json:"meta"`
}

// AnalyticsDNSSECResponse is the public response.
type AnalyticsDNSSECResponse struct {
	Data       []*AnalyticsDNSSECEntry
	Pagination AnalyticsPagination
}

// AnalyticsDNSSECTimeSeriesResponse is the public time-series response.
type AnalyticsDNSSECTimeSeriesResponse struct {
	Data       []*AnalyticsDNSSECTimeSeriesEntry
	Pagination AnalyticsPagination
	Series     AnalyticsSeriesInfo
}

// AnalyticsEncryptionEntry is one row of the encryption analytics response.
type AnalyticsEncryptionEntry struct {
	Encrypted bool `json:"encrypted"`
	Queries   int  `json:"queries"`
}

// AnalyticsEncryptionTimeSeriesEntry has queries as an array.
type AnalyticsEncryptionTimeSeriesEntry struct {
	Encrypted bool  `json:"encrypted"`
	Queries   []int `json:"queries"`
}

type analyticsEncryptionResponse struct {
	Data []*AnalyticsEncryptionEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
	} `json:"meta"`
}

type analyticsEncryptionTimeSeriesResponse struct {
	Data []*AnalyticsEncryptionTimeSeriesEntry `json:"data"`
	Meta struct {
		Pagination AnalyticsPagination `json:"pagination"`
		Series     AnalyticsSeriesInfo `json:"series"`
	} `json:"meta"`
}

// AnalyticsEncryptionResponse is the public response.
type AnalyticsEncryptionResponse struct {
	Data       []*AnalyticsEncryptionEntry
	Pagination AnalyticsPagination
}

// AnalyticsEncryptionTimeSeriesResponse is the public time-series response.
type AnalyticsEncryptionTimeSeriesResponse struct {
	Data       []*AnalyticsEncryptionTimeSeriesEntry
	Pagination AnalyticsPagination
	Series     AnalyticsSeriesInfo
}

type analyticsService struct {
	client *Client
}

// Compile-time check that analyticsService implements AnalyticsService.
var _ AnalyticsService = &analyticsService{}

// NewAnalyticsService creates a new analytics service.
func NewAnalyticsService(client *Client) AnalyticsService {
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
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

// GetQueryTypes returns counts by DNS query type.
func (s *analyticsService) GetQueryTypes(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "queryTypes")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics queryTypes: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics queryTypes: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetQueryTypesSeries returns counts by DNS query type as a time series.
func (s *analyticsService) GetQueryTypesSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "queryTypes;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics queryTypes series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics queryTypes series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetReasons returns counts by block reason.
func (s *analyticsService) GetReasons(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "reasons")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics reasons: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics reasons: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetReasonsSeries returns counts by block reason as a time series.
func (s *analyticsService) GetReasonsSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "reasons;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics reasons series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics reasons series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetIPs returns counts by client IP.
func (s *analyticsService) GetIPs(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsResponse, error) {
	path := analyticsPath(request.ProfileID, "ips")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics ips: %w", err)
	}

	response := analyticsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics ips: %w", err)
	}

	return &AnalyticsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetIPsSeries returns counts by client IP as a time series.
func (s *analyticsService) GetIPsSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "ips;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics ips series: %w", err)
	}

	response := analyticsTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics ips series: %w", err)
	}

	return &AnalyticsTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetDNSSEC returns DNSSEC validation analytics.
func (s *analyticsService) GetDNSSEC(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsDNSSECResponse, error) {
	path := analyticsPath(request.ProfileID, "dnssec")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics dnssec: %w", err)
	}

	response := analyticsDNSSECResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics dnssec: %w", err)
	}

	return &AnalyticsDNSSECResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetDNSSECSeries returns DNSSEC validation analytics as a time series.
func (s *analyticsService) GetDNSSECSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsDNSSECTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "dnssec;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics dnssec series: %w", err)
	}

	response := analyticsDNSSECTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics dnssec series: %w", err)
	}

	return &AnalyticsDNSSECTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}

// GetEncryption returns query encryption analytics.
func (s *analyticsService) GetEncryption(ctx context.Context, request *GetAnalyticsRequest) (*AnalyticsEncryptionResponse, error) {
	path := analyticsPath(request.ProfileID, "encryption")
	query := buildAnalyticsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics encryption: %w", err)
	}

	response := analyticsEncryptionResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics encryption: %w", err)
	}

	return &AnalyticsEncryptionResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
	}, nil
}

// GetEncryptionSeries returns query encryption analytics as a time series.
func (s *analyticsService) GetEncryptionSeries(ctx context.Context, request *GetAnalyticsTimeSeriesRequest) (*AnalyticsEncryptionTimeSeriesResponse, error) {
	path := analyticsPath(request.ProfileID, "encryption;series")
	query := buildTimeSeriesQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get analytics encryption series: %w", err)
	}

	response := analyticsEncryptionTimeSeriesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get analytics encryption series: %w", err)
	}

	return &AnalyticsEncryptionTimeSeriesResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Series:     response.Meta.Series,
	}, nil
}
