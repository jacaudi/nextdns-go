package nextdns

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL     = "https://api.nextdns.io/"
	contentType = "application/json"
	userAgent   = "nextdns-go"
)

// defaultHTTPClient returns the SDK's default *http.Client.
// It carries an overall request timeout, TLS 1.3 floor, and tuned transport
// settings per go-standards §15.1.
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:       30 * time.Second,
		CheckRedirect: stripAuthOnCrossHost,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, // RESTORED: honors HTTPS_PROXY/HTTP_PROXY/NO_PROXY
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second, // RESTORED from cleanhttp default
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:       90 * time.Second,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			MaxConnsPerHost:       25,
			ForceAttemptHTTP2:     true,
		},
	}
}

// stripAuthOnCrossHost is the SDK's default CheckRedirect: it preserves the
// stdlib's default 10-redirect cap and removes X-Api-Key + Authorization from
// requests whose target host differs from the previous request's host.
//
// This prevents the customer's API key from leaking to CDN / object-storage
// hosts that NextDNS uses for binary download endpoints (logs/download).
// Go's stdlib only strips Authorization/Cookie/WWW-Authenticate on cross-host
// redirects, NOT custom X-* headers — without this, X-Api-Key would propagate.
func stripAuthOnCrossHost(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if len(via) == 0 {
		return nil
	}
	prev := via[len(via)-1].URL.Host
	if !strings.EqualFold(req.URL.Host, prev) {
		req.Header.Del("X-Api-Key")
		req.Header.Del("Authorization")
	}
	return nil
}

// Client represents a NextDNS client.
type Client struct {
	client  *http.Client
	baseURL *url.URL

	// Service for the Profile.
	Profiles ProfilesService

	// Services for the Allowlist and Denylist.
	Allowlist AllowlistService
	Denylist  DenylistService

	// Services for the ParentalControl.
	ParentalControl           ParentalControlService
	ParentalControlServices   ParentalControlServicesService
	ParentalControlCategories ParentalControlCategoriesService

	// Services for the Privacy.
	Privacy           PrivacyService
	PrivacyBlocklists PrivacyBlocklistsService
	PrivacyNatives    PrivacyNativesService

	// Services for the Settings.
	Settings            SettingsService
	SettingsLogs        SettingsLogsService
	SettingsBlockPage   SettingsBlockPageService
	SettingsPerformance SettingsPerformanceService

	// Services for the Security.
	Security     SecurityService
	SecurityTlds SecurityTldsService

	// Services for the Rewrites.
	Rewrites RewritesService

	// Services for the Setup.
	Setup         SetupService
	SetupLinkedIP SetupLinkedIPService

	// Services for Analytics.
	Analytics AnalyticsService

	// Services for Logs.
	Logs LogsService

	// Optional debug logger; when set, the SDK logs requests and responses
	// at slog.LevelDebug.
	debugLogger *slog.Logger

	// Optional rate-limit observer; lands in Task A9.
	rateLimitObserver func(RateLimit)
}

// ClientOption is a function that can be used to customize the client.
type ClientOption func(c *Client) error

// RateLimit holds the rate-limit headers returned by the NextDNS API on each
// response. Reset is the absolute time at which the current limit window
// resets (parsed from the X-RateLimit-Reset Unix timestamp).
type RateLimit struct {
	Limit     int
	Remaining int
	Reset     time.Time
}

// WithBaseURL sets the base URL of the NextDNS API.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		parsedURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}

		c.baseURL = parsedURL
		return nil
	}
}

// WithAPIKey sets the API key to be used for requests.
//
// The injected X-Api-Key header is scoped to the host of c.baseURL at request
// time: authTransport only attaches the key when the outgoing request's host
// matches c.baseURL.Host. This prevents the key from leaking to CDN /
// object-storage hosts on cross-host redirects (e.g. Logs.Download → S3).
// stripAuthOnCrossHost (the default CheckRedirect) provides defense in depth.
//
// Option ordering: WithBaseURL may be applied before or after WithAPIKey; the
// host check is evaluated lazily at request time. WithHTTPClient, however,
// must be applied before WithAPIKey — it replaces c.client wholesale and
// would discard the auth wrapper otherwise.
func WithAPIKey(apiKey Secret) ClientOption {
	return func(c *Client) error {
		if apiKey.Expose() == "" {
			return ErrEmptyAPIToken
		}

		// http.Client.Do() falls back to http.DefaultTransport when Transport
		// is nil. Mirror that fallback explicitly so authTransport.rt is never
		// nil — otherwise t.rt.RoundTrip(req) panics inside the wrapper.
		if c.client.Transport == nil {
			c.client.Transport = http.DefaultTransport
		}

		transport := authTransport{
			rt:     c.client.Transport,
			apiKey: apiKey,
			// Lazy host resolution: read c.baseURL at request time so option
			// ordering (WithBaseURL before/after WithAPIKey) doesn't matter.
			trustedHostFn: func() string { return c.baseURL.Host },
		}

		c.client.Transport = &transport
		return nil
	}
}

// WithLogger sets a slog.Logger used by the SDK to emit request/response
// debug records at slog.LevelDebug. When unset, the SDK emits nothing.
//
// Per go-standards.md §7: library code does not log at Info or above. The
// logger is consulted at Debug only; consumers decide whether to enable.
func WithLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) error {
		c.debugLogger = logger
		return nil
	}
}

// WithRateLimitObserver sets a callback invoked after every HTTP response
// with the parsed X-RateLimit-* headers. When unset, no callback is invoked.
//
// The callback runs on the HTTP response goroutine. Keep it cheap — consumers
// that need heavy work should write to a channel or atomic and process
// asynchronously.
func WithRateLimitObserver(fn func(RateLimit)) ClientOption {
	return func(c *Client) error {
		c.rateLimitObserver = fn
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client that can be used for requests.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) error {
		if client == nil {
			client = defaultHTTPClient()
		}

		c.client = client
		return nil
	}
}

// New instantiates a new NextDNS client.
func New(opts ...ClientOption) (*Client, error) {
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:  defaultHTTPClient(),
		baseURL: baseURL,
	}

	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// Initialize the services for the Profile.
	c.Profiles = NewProfilesService(c)

	// Initialize the services for the Allowlist and Denylist.
	c.Allowlist = NewAllowlistService(c)
	c.Denylist = NewDenylistService(c)

	// Initialize the services for the ParentalControl.
	c.ParentalControl = NewParentalControlService(c)
	c.ParentalControlServices = NewParentalControlServicesService(c)
	c.ParentalControlCategories = NewParentalControlCategoriesService(c)

	// Initialize the services for the Privacy.
	c.Privacy = NewPrivacyService(c)
	c.PrivacyBlocklists = NewPrivacyBlocklistsService(c)
	c.PrivacyNatives = NewPrivacyNativesService(c)

	// Initialize the services for the Settings.
	c.Settings = NewSettingsService(c)
	c.SettingsLogs = NewSettingsLogsService(c)
	c.SettingsBlockPage = NewSettingsBlockPageService(c)
	c.SettingsPerformance = NewSettingsPerformanceService(c)

	// Initialize the services for the Security.
	c.Security = NewSecurityService(c)
	c.SecurityTlds = NewSecurityTldsService(c)

	// Initialize the services for the Rewrites.
	c.Rewrites = NewRewritesService(c)

	// Initialize the services for the Setup.
	c.Setup = NewSetupService(c)
	c.SetupLinkedIP = NewSetupLinkedIPService(c)

	// Initialize the services for Analytics.
	c.Analytics = NewAnalyticsService(c)

	// Initialize the services for Logs.
	c.Logs = NewLogsService(c)

	return c, nil
}

// do executes an HTTP request and decodes the response into v.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	return c.handleResponse(res, v)
}

// handleResponse handles the response from the NextDNS API and decodes the response into v if provided.
// The goal is to handle the common errors that can occur when making a request to the NextDNS API,
// and also provide custom error responses for the client.
func (c *Client) handleResponse(res *http.Response, v interface{}) error {
	if c.rateLimitObserver != nil {
		c.rateLimitObserver(parseRateLimit(res.Header))
	}

	out, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if c.debugLogger != nil {
		c.debugLogger.Debug("nextdns response",
			"status", res.StatusCode,
			"body", string(out))
	}

	// If there is no response body, then we don't need to do anything.
	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	// If the response is not a 200, then we need to handle the error.
	// TODO(jacaudi): Report the behavior to NextDNS, but there are errors that return HTTP 200 ("duplicate" case). See https://github.com/jacaudi/nextdns-go/issues/8
	if res.StatusCode >= http.StatusBadRequest || strings.Contains(string(out), "\"errors\"") {
		return c.errorFromBytes(res.StatusCode, out)
	}

	// Returns if there is no object to decode.
	if v == nil {
		return nil
	}

	// Sets some default additional informations that can be used by the client to debug the error.
	meta := map[string]string{
		"body":        string(out),
		"http_status": http.StatusText(res.StatusCode),
	}

	// Decodes the response body into the provided object.
	err = json.Unmarshal(out, &v)
	if err != nil {
		var jsonErr *json.SyntaxError
		if errors.As(err, &jsonErr) {
			meta["err"] = jsonErr.Error()
			return &Error{
				Type:    ErrorTypeMalformed,
				Message: errMalformedError,
				Errors:  nil,
				Meta:    meta,
			}
		}
		return err
	}

	return nil
}

// parseErrorResponse reads an *http.Response whose body is expected to be a
// NextDNS error payload, attempts to decode the JSON into *Error, and returns
// the typed error (falling back to a generic status-code error if the body is
// unreadable or malformed). The caller is responsible for closing res.Body.
//
// This is the entry point for callers that bypass handleResponse — currently
// Logs.Download and Logs.Stream, whose success-path bodies are not JSON and
// therefore can't be funneled through handleResponse. handleResponse itself
// goes through errorFromBytes directly because it has already read the body.
//
// Note: does NOT assume res.StatusCode >= 400. handleResponse also routes 200
// responses whose body contains "\"errors\"" through here (see the NextDNS
// "duplicate" quirk), and callers that detect a failure by content-type or
// other heuristic may legitimately invoke this on a 2xx status.
func (c *Client) parseErrorResponse(res *http.Response) error {
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return c.errorFromBytes(res.StatusCode, out)
}

// errorFromBytes maps a (status code, raw body) pair to the SDK's typed
// *Error. Shared between handleResponse (which has already read the body)
// and parseErrorResponse (which reads it for non-JSON-body callers).
//
// Behavior:
//   - 5xx: short-circuit to ErrorTypeServiceError without decoding the body.
//   - Otherwise: decode body as *ErrorResponse. JSON syntax errors return
//     ErrorTypeMalformed; other Unmarshal errors surface unwrapped.
//   - Status → ErrorType mapping: 403 → Authentication, 404 → NotFound,
//     everything else → Request.
//
// The meta map carries the raw body and status text (and, on syntax error,
// the json.SyntaxError text) to aid debugging.
func (c *Client) errorFromBytes(statusCode int, body []byte) error {
	meta := map[string]string{
		"body":        string(body),
		"http_status": http.StatusText(statusCode),
	}

	if statusCode >= http.StatusInternalServerError {
		return &Error{
			Type:    ErrorTypeServiceError,
			Message: errInternalServiceError,
			Errors:  nil,
			Meta:    meta,
		}
	}

	// Tries to handle the error response body from the NextDNS API,
	// encapsulated in a client error.
	errorRes := &ErrorResponse{}
	err := json.Unmarshal(body, errorRes)
	if err != nil {
		var jsonErr *json.SyntaxError
		if errors.As(err, &jsonErr) {
			meta["err"] = jsonErr.Error()
			return &Error{
				Type:    ErrorTypeMalformed,
				Message: errMalformedErrorBody,
				Errors:  nil,
				Meta:    meta,
			}
		}
		return err
	}

	// Sets custom error messages for the client based on the HTTP status code.
	var errType ErrorType

	switch statusCode {
	case http.StatusForbidden:
		errType = ErrorTypeAuthentication
	case http.StatusNotFound:
		errType = ErrorTypeNotFound
	default:
		errType = ErrorTypeRequest
	}

	// Returns the error response from the NextDNS API encapsulated in a client error.
	return &Error{
		Type:    errType,
		Message: errResponseError,
		Errors:  errorRes,
		Meta:    meta,
	}
}

// newRequest creates a new HTTP request with optional query parameters and body.
// Pass nil for query and/or body when not needed.
func (c *Client) newRequest(method string, path string, query url.Values, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		u.RawQuery = query.Encode()
	}

	var (
		bodyReader io.Reader
		bodyBuf    *bytes.Buffer // captured for debug; nil if no body
	)
	if body != nil && method != http.MethodGet {
		bodyBuf = new(bytes.Buffer)
		if err := json.NewEncoder(bodyBuf).Encode(body); err != nil {
			return nil, err
		}
		bodyReader = bodyBuf
	}

	if c.debugLogger != nil {
		if bodyBuf == nil {
			c.debugLogger.Debug("nextdns request",
				"method", method,
				"url", u.String())
		} else {
			c.debugLogger.Debug("nextdns request",
				"method", method,
				"url", u.String(),
				"body", strings.TrimSuffix(bodyBuf.String(), "\n"))
		}
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	if method != http.MethodGet {
		// Every non-GET advertises JSON content even when the body is empty.
		// Old code did this implicitly by passing an empty *bytes.Buffer to
		// http.NewRequest; the unified helper now passes nil, so we set the
		// header explicitly. Some strict gateways validate Content-Type even
		// on DELETE.
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// authTransport adds the NextDNS authentication header to outgoing requests
// whose host matches the configured trusted host.
type authTransport struct {
	rt     http.RoundTripper
	apiKey Secret
	// trustedHostFn returns the host that is allowed to receive X-Api-Key.
	// It is invoked at request time so option ordering between WithBaseURL
	// and WithAPIKey at construction does not matter.
	trustedHostFn func() string
}

// RoundTrip adds the authorization header to a CLONE of the inbound request,
// per the http.RoundTripper contract ("RoundTrip should not modify the
// request"). Uses Set rather than Add so retries through the same transport
// do not accumulate duplicate X-Api-Key headers.
//
// The header is only attached when the request's host matches the trusted
// host (case-insensitive per RFC 4343). On a cross-host redirect — e.g. the
// NextDNS logs/download endpoint redirecting to a CDN / S3 host — the key is
// withheld so it cannot leak to third-party access logs or proxies. An empty
// trusted host is treated as no-trust (the key is withheld) so the check
// fails secure even though net/http rejects empty-Host requests in practice.
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	if t.trustedHostFn != nil {
		if trusted := t.trustedHostFn(); trusted != "" && strings.EqualFold(clone.URL.Host, trusted) {
			clone.Header.Set("X-Api-Key", t.apiKey.Expose())
		}
	}
	return t.rt.RoundTrip(clone)
}

// parseRateLimit pulls X-RateLimit-* headers out of an HTTP response.
// Missing or malformed headers leave the corresponding field zero-valued.
func parseRateLimit(h http.Header) RateLimit {
	rl := RateLimit{}
	if v := h.Get("X-RateLimit-Limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.Limit = n
		}
	}
	if v := h.Get("X-RateLimit-Remaining"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rl.Remaining = n
		}
	}
	if v := h.Get("X-RateLimit-Reset"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			rl.Reset = time.Unix(n, 0).UTC()
		}
	}
	return rl
}
