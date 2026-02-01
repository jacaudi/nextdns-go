package nextdns

import (
	"context"
	"time"
)

// logsAPIPath is the HTTP path for the logs API.
const logsAPIPath = "logs"

// LogDevice represents device information in a log entry.
type LogDevice struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model,omitempty"`
}

// LogReason represents a block/allow reason.
type LogReason struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LogEntry represents a single DNS query log entry.
type LogEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Domain    string      `json:"domain"`
	Root      string      `json:"root"`
	Tracker   string      `json:"tracker,omitempty"`
	Encrypted bool        `json:"encrypted"`
	Protocol  string      `json:"protocol"`
	ClientIP  string      `json:"clientIp"`
	Client    string      `json:"client,omitempty"`
	Device    *LogDevice  `json:"device,omitempty"`
	Status    string      `json:"status"`
	Reasons   []LogReason `json:"reasons,omitempty"`
}

// LogsQueryOptions contains parameters for querying logs.
type LogsQueryOptions struct {
	From   string // Date filter (ISO 8601, Unix timestamp, or relative like "-7d")
	To     string // Date filter
	Sort   string // "asc" or "desc" (default: "desc")
	Limit  int    // Results per page (10-1000, default 100)
	Cursor string // Pagination cursor
	Device string // Filter by device ID
	Status string // Filter: "default", "error", "blocked", "allowed"
	Search string // Domain search (partial matching supported)
	Raw    bool   // Show all queries vs. cleaned navigational only
}

// LogsPagination contains cursor for pagination.
type LogsPagination struct {
	Cursor string `json:"cursor"`
}

// LogsStreamInfo contains stream ID for stitching with real-time streaming.
type LogsStreamInfo struct {
	ID string `json:"id"`
}

// logsResponse is the internal response wrapper.
type logsResponse struct {
	Data []*LogEntry `json:"data"`
	Meta struct {
		Pagination LogsPagination `json:"pagination"`
		Stream     LogsStreamInfo `json:"stream"`
	} `json:"meta"`
}

// LogsResponse contains log entries with pagination info.
type LogsResponse struct {
	Data       []*LogEntry
	Pagination LogsPagination
	Stream     LogsStreamInfo
}

// Request types for logs endpoints

// GetLogsRequest is used for querying logs.
type GetLogsRequest struct {
	ProfileID string
	Options   *LogsQueryOptions
}

// ClearLogsRequest is used for clearing logs.
type ClearLogsRequest struct {
	ProfileID string
}

// LogsService provides access to NextDNS query logs.
type LogsService interface {
	// Get queries DNS query logs with filtering and pagination.
	Get(ctx context.Context, request *GetLogsRequest) (*LogsResponse, error)

	// Clear deletes all logs for a profile.
	Clear(ctx context.Context, request *ClearLogsRequest) error
}

type logsService struct {
	client *Client
}

// Compile-time check that logsService implements LogsService.
var _ LogsService = &logsService{}

// NewLogsService creates a new logs service.
func NewLogsService(client *Client) *logsService {
	return &logsService{
		client: client,
	}
}

// Get queries DNS query logs with filtering and pagination.
func (s *logsService) Get(ctx context.Context, request *GetLogsRequest) (*LogsResponse, error) {
	// TODO: Implement in Task 4
	return nil, nil
}

// Clear deletes all logs for a profile.
func (s *logsService) Clear(ctx context.Context, request *ClearLogsRequest) error {
	// TODO: Implement in Task 5
	return nil
}
