// Package nextdns provides a client library for interacting with the NextDNS API.
package nextdns

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS13,
			},
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
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

	// Debug mode for the HTTP requests.
	Debug bool
}

// ClientOption is a function that can be used to customize the client.
type ClientOption func(c *Client) error

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
func WithAPIKey(apiKey Secret) ClientOption {
	return func(c *Client) error {
		if apiKey.Expose() == "" {
			return ErrEmptyAPIToken
		}

		transport := authTransport{
			rt:     c.client.Transport,
			apiKey: apiKey,
		}

		c.client.Transport = &transport
		return nil
	}
}

// WithDebug enables debug mode.
func WithDebug() ClientOption {
	return func(c *Client) error {
		c.Debug = true
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
	out, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if c.Debug {
		if string(out) == "" {
			fmt.Printf("[DEBUG] RESPONSE: StatusCode:%d\n", res.StatusCode)
		} else {
			fmt.Printf("[DEBUG] RESPONSE: StatusCode:%d, Body:%v\n", res.StatusCode, string(out))
		}
	}

	// If there is no response body, then we don't need to do anything.
	if res.StatusCode == http.StatusNoContent {
		return nil
	}

	// Sets some default additional informations that can be used by the client to debug the error.
	meta := map[string]string{
		"body":        string(out),
		"http_status": http.StatusText(res.StatusCode),
	}

	// If the response is not a 200, then we need to handle the error.
	// TODO(jacaudi): Report the behavior to NextDNS, but there are errors that return HTTP 200 ("duplicate" case). See https://github.com/jacaudi/nextdns-go/issues/8
	if res.StatusCode >= http.StatusBadRequest || strings.Contains(string(out), "\"errors\"") {
		if res.StatusCode >= http.StatusInternalServerError {
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
		err = json.Unmarshal(out, errorRes)
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

		switch res.StatusCode {
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

	// Returns if there is no object to decode.
	if v == nil {
		return nil
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

	var bodyReader io.Reader
	if body != nil && method != http.MethodGet {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
		bodyReader = buf
	}

	if c.Debug {
		if bodyReader == nil {
			fmt.Printf("[DEBUG] REQUEST: Method:%s, URL:%s\n", method, u.String())
		} else {
			buf := bodyReader.(*bytes.Buffer)
			fmt.Printf("[DEBUG] REQUEST: Method:%s, URL:%s, Body:%s\n", method, u.String(), strings.TrimSuffix(buf.String(), "\n"))
		}
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", userAgent)
	return req, nil
}

// authTransport adds the NextDNS authentication header to outgoing requests.
type authTransport struct {
	rt     http.RoundTripper
	apiKey Secret
}

// RoundTrip adds the authorization header to requests.
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Api-Key", t.apiKey.Expose())
	return t.rt.RoundTrip(req)
}
