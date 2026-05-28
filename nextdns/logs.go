package nextdns

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	Protocol  DNSProtocol `json:"protocol"`
	ClientIP  string      `json:"clientIp"`
	Client    string      `json:"client,omitempty"`
	Device    *LogDevice  `json:"device,omitempty"`
	Status    LogStatus   `json:"status"`
	Reasons   []LogReason `json:"reasons,omitempty"`
}

// LogsQueryOptions contains parameters for querying logs.
type LogsQueryOptions struct {
	From   string    // Date filter (ISO 8601, Unix timestamp, or relative like "-7d")
	To     string    // Date filter
	Sort   SortOrder // "asc" or "desc" (default: "desc")
	Limit  int       // Results per page (10-1000, default 100)
	Cursor string    // Pagination cursor
	Device string    // Filter by device ID
	Status LogStatus // Filter: "default", "error", "blocked", "allowed"
	Search string    // Domain search (partial matching supported)
	Raw    bool      // When true, return all DNS queries (raw=1). Default returns navigational queries (A/AAAA/HTTPS) deduplicated.
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

// DownloadLogsRequest encapsulates a request to /logs/download.
type DownloadLogsRequest struct {
	ProfileID string
}

// DownloadLogsURLResponse holds the URL returned when downloading via redirect=0.
type DownloadLogsURLResponse struct {
	URL string `json:"url"`
}

// StreamLogsRequest encapsulates a request to /logs/stream.
type StreamLogsRequest struct {
	ProfileID string
	Device    string // optional device filter
	LastID    string // optional; resume from this SSE event id
}

// LogsService provides access to NextDNS query logs.
type LogsService interface {
	// Get queries DNS query logs with filtering and pagination.
	Get(ctx context.Context, request *GetLogsRequest) (*LogsResponse, error)

	// Clear deletes all logs for a profile.
	Clear(ctx context.Context, request *ClearLogsRequest) error

	// Download follows the 302 to fetch the CSV log archive. Caller owns
	// the returned ReadCloser and must Close() it.
	Download(ctx context.Context, request *DownloadLogsRequest) (io.ReadCloser, error)

	// DownloadURL returns the URL where the CSV log archive can be fetched,
	// without following the redirect. Useful for showing a loader while the
	// file is being generated.
	DownloadURL(ctx context.Context, request *DownloadLogsRequest) (*DownloadLogsURLResponse, error)

	// Stream consumes the /logs/stream SSE endpoint as an iterator.
	// Iteration ends when ctx is cancelled or the connection drops.
	// The last yielded pair carries err != nil to signal end of stream.
	//
	// Caller controls reconnect: track the last event ID and pass it
	// via StreamLogsRequest.LastID on the next call.
	Stream(ctx context.Context, request *StreamLogsRequest) iter.Seq2[*LogEntry, error]
}

type logsService struct {
	client *Client
}

// Compile-time check that logsService implements LogsService.
var _ LogsService = &logsService{}

// NewLogsService creates a new logs service.
func NewLogsService(client *Client) LogsService {
	return &logsService{
		client: client,
	}
}

// buildLogsQuery converts LogsQueryOptions to url.Values.
func buildLogsQuery(opts *LogsQueryOptions) url.Values {
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
	if opts.Sort != "" {
		query.Set("sort", string(opts.Sort))
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
	if opts.Status != "" {
		query.Set("status", string(opts.Status))
	}
	if opts.Search != "" {
		query.Set("search", opts.Search)
	}
	if opts.Raw {
		query.Set("raw", "1")
	}
	return query
}

func logsPath(profileID string) string {
	return fmt.Sprintf("%s/%s/%s", profilesAPIPath, profileID, logsAPIPath)
}

// Get queries DNS query logs with filtering and pagination.
func (s *logsService) Get(ctx context.Context, request *GetLogsRequest) (*LogsResponse, error) {
	path := logsPath(request.ProfileID)
	query := buildLogsQuery(request.Options)

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get logs: %w", err)
	}

	response := logsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get logs: %w", err)
	}

	return &LogsResponse{
		Data:       response.Data,
		Pagination: response.Meta.Pagination,
		Stream:     response.Meta.Stream,
	}, nil
}

// Clear deletes all logs for a profile.
func (s *logsService) Clear(ctx context.Context, request *ClearLogsRequest) error {
	path := logsPath(request.ProfileID)

	req, err := s.client.newRequest(http.MethodDelete, path, nil, nil)
	if err != nil {
		return fmt.Errorf("error creating request to clear logs: %w", err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making request to clear logs: %w", err)
	}

	return nil
}

// Download fetches the CSV log archive, following the 302 redirect.
// Caller MUST Close() the returned ReadCloser.
//
// The HTTP client is cloned per-call to zero the overall Timeout — CSV
// archives for high-volume profiles can be tens of MB, easily exceeding
// the SDK default 30s. The Transport (and X-Api-Key injection / proxy
// settings / CheckRedirect) is shared with the SDK default.
func (s *logsService) Download(ctx context.Context, request *DownloadLogsRequest) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/download", logsPath(request.ProfileID))

	req, err := s.client.newRequest(http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to download logs: %w", err)
	}
	req = req.WithContext(ctx)

	downloadClient := *s.client.client
	downloadClient.Timeout = 0

	res, err := downloadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to download logs: %w", err)
	}

	if s.client.rateLimitObserver != nil {
		s.client.rateLimitObserver(parseRateLimit(res.Header))
	}

	if res.StatusCode >= 400 {
		defer func() { _ = res.Body.Close() }()
		return nil, s.client.parseErrorResponse(res)
	}

	return res.Body, nil
}

// DownloadURL returns the public URL where the CSV log archive can be fetched.
func (s *logsService) DownloadURL(ctx context.Context, request *DownloadLogsRequest) (*DownloadLogsURLResponse, error) {
	path := fmt.Sprintf("%s/download", logsPath(request.ProfileID))
	query := url.Values{}
	query.Set("redirect", "0")

	req, err := s.client.newRequest(http.MethodGet, path, query, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get logs download URL: %w", err)
	}

	response := DownloadLogsURLResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making request to get logs download URL: %w", err)
	}

	return &response, nil
}

// Stream consumes the SSE event stream from /logs/stream.
func (s *logsService) Stream(ctx context.Context, request *StreamLogsRequest) iter.Seq2[*LogEntry, error] {
	return func(yield func(*LogEntry, error) bool) {
		path := fmt.Sprintf("%s/stream", logsPath(request.ProfileID))
		query := url.Values{}
		if request.Device != "" {
			query.Set("device", request.Device)
		}
		if request.LastID != "" {
			query.Set("id", request.LastID)
		}

		req, err := s.client.newRequest(http.MethodGet, path, query, nil)
		if err != nil {
			yield(nil, fmt.Errorf("error creating request to stream logs: %w", err))
			return
		}
		req = req.WithContext(ctx)
		req.Header.Set("Accept", "text/event-stream")

		// SSE connections are long-lived; do NOT use the default Client.Timeout
		// (30s, set in defaultHTTPClient). We need a copy of the *http.Client
		// without the overall timeout.
		streamClient := *s.client.client
		streamClient.Timeout = 0

		res, err := streamClient.Do(req)
		if err != nil {
			yield(nil, fmt.Errorf("error opening logs stream: %w", err))
			return
		}
		defer func() { _ = res.Body.Close() }()

		if s.client.rateLimitObserver != nil {
			s.client.rateLimitObserver(parseRateLimit(res.Header))
		}
		if res.StatusCode != http.StatusOK {
			yield(nil, s.client.parseErrorResponse(res))
			return
		}

		scanner := bufio.NewScanner(res.Body)
		// SSE event bodies can be large; raise the buffer ceiling.
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		var dataPayload strings.Builder

		const dataPrefix = "data:"

		for scanner.Scan() {
			if ctx.Err() != nil {
				yield(nil, ctx.Err())
				return
			}
			line := scanner.Text()
			switch {
			case strings.HasPrefix(line, dataPrefix):
				payload := strings.TrimPrefix(line, dataPrefix)
				payload = strings.TrimPrefix(payload, " ") // optional space per SSE spec
				if dataPayload.Len() > 0 {
					dataPayload.WriteString("\n")
				}
				dataPayload.WriteString(payload)
			case line == "":
				// Event boundary — emit if we have a payload.
				if dataPayload.Len() == 0 {
					continue
				}
				entry := &LogEntry{}
				if err := json.Unmarshal([]byte(dataPayload.String()), entry); err != nil {
					if !yield(nil, fmt.Errorf("error decoding stream event: %w", err)) {
						return
					}
				} else if !yield(entry, nil) {
					return
				}
				dataPayload.Reset()
			default:
				// id: lines or comments — currently ignored. (Future: track LastID.)
			}
		}

		if err := scanner.Err(); err != nil {
			yield(nil, fmt.Errorf("logs stream read error: %w", err))
		}
	}
}
