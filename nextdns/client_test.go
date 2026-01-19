package nextdns

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestNewClient(t *testing.T) {
	c := is.New(t)

	client, err := New()
	c.NoErr(err)
	c.True(client != nil)
	c.Equal(client.baseURL.String(), "https://api.nextdns.io/")
}

func TestNewClientWithAPIKey(t *testing.T) {
	c := is.New(t)

	client, err := New(WithAPIKey("test-api-key"))
	c.NoErr(err)
	c.True(client != nil)
}

func TestNewClientWithEmptyAPIKey(t *testing.T) {
	c := is.New(t)

	_, err := New(WithAPIKey(""))
	c.True(err != nil)
	c.Equal(err, ErrEmptyAPIToken)
}

func TestNewClientWithBaseURL(t *testing.T) {
	c := is.New(t)

	client, err := New(WithBaseURL("https://custom.api.example.com/"))
	c.NoErr(err)
	c.Equal(client.baseURL.String(), "https://custom.api.example.com/")
}

func TestNewClientWithDebug(t *testing.T) {
	c := is.New(t)

	client, err := New(WithDebug())
	c.NoErr(err)
	c.True(client.Debug)
}

func TestNewClientWithHTTPClient(t *testing.T) {
	c := is.New(t)

	customClient := &http.Client{}
	client, err := New(WithHTTPClient(customClient))
	c.NoErr(err)
	c.Equal(client.client, customClient)
}

func TestNewClientWithNilHTTPClient(t *testing.T) {
	c := is.New(t)

	client, err := New(WithHTTPClient(nil))
	c.NoErr(err)
	c.True(client.client != nil)
}

func TestClientServicesInitialized(t *testing.T) {
	c := is.New(t)

	client, err := New()
	c.NoErr(err)

	c.True(client.Profiles != nil)
	c.True(client.Allowlist != nil)
	c.True(client.Denylist != nil)
	c.True(client.ParentalControl != nil)
	c.True(client.ParentalControlServices != nil)
	c.True(client.ParentalControlCategories != nil)
	c.True(client.Privacy != nil)
	c.True(client.PrivacyBlocklists != nil)
	c.True(client.PrivacyNatives != nil)
	c.True(client.Settings != nil)
	c.True(client.SettingsLogs != nil)
	c.True(client.SettingsBlockPage != nil)
	c.True(client.SettingsPerformance != nil)
	c.True(client.Security != nil)
	c.True(client.SecurityTlds != nil)
	c.True(client.Rewrites != nil)
	c.True(client.Setup != nil)
	c.True(client.SetupLinkedIP != nil)
}

func TestAPIKeyHeader(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		c.Equal(apiKey, "test-api-key-12345")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer ts.Close()

	client, err := New(
		WithBaseURL(ts.URL),
		WithAPIKey("test-api-key-12345"),
	)
	c.NoErr(err)

	ctx := context.Background()
	_, _ = client.Profiles.List(ctx, &ListProfileRequest{})
}

func TestRequestHeaders(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Equal(r.Header.Get("Accept"), "application/json")
		c.Equal(r.Header.Get("User-Agent"), "nextdns-go")

		if r.Method != http.MethodGet {
			c.Equal(r.Header.Get("Content-Type"), "application/json")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	_, _ = client.Profiles.List(ctx, &ListProfileRequest{})
}

func TestErrorResponseHandling(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   string
		expectedType   ErrorType
		expectedInBody string
	}{
		{
			name:           "forbidden error",
			statusCode:     http.StatusForbidden,
			responseBody:   `{"errors":[{"code":"forbidden","detail":"Access denied"}]}`,
			expectedType:   ErrorTypeAuthentication,
			expectedInBody: "Access denied",
		},
		{
			name:           "not found error",
			statusCode:     http.StatusNotFound,
			responseBody:   `{"errors":[{"code":"not_found","detail":"Profile not found"}]}`,
			expectedType:   ErrorTypeNotFound,
			expectedInBody: "Profile not found",
		},
		{
			name:           "bad request error",
			statusCode:     http.StatusBadRequest,
			responseBody:   `{"errors":[{"code":"invalid_request","detail":"Invalid profile ID"}]}`,
			expectedType:   ErrorTypeRequest,
			expectedInBody: "Invalid profile ID",
		},
		{
			name:         "internal server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: `{"errors":[{"code":"internal_error"}]}`,
			expectedType: ErrorTypeServiceError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := is.New(t)

			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer ts.Close()

			client, err := New(WithBaseURL(ts.URL))
			c.NoErr(err)

			ctx := context.Background()
			_, err = client.Profiles.Get(ctx, &GetProfileRequest{ProfileID: "test"})

			c.True(err != nil)
			var clientErr *Error
			ok := errors.As(err, &clientErr)
			c.True(ok)
			c.Equal(clientErr.Type, tt.expectedType)
		})
	}
}

func TestMalformedJSONResponse(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	_, err = client.Profiles.Get(ctx, &GetProfileRequest{ProfileID: "test"})

	c.True(err != nil)
	var clientErr *Error
	ok := errors.As(err, &clientErr)
	c.True(ok)
	c.Equal(clientErr.Type, ErrorTypeMalformed)
}

func TestDuplicateErrorInSuccessResponse(t *testing.T) {
	c := is.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"errors":[{"code":"duplicate","detail":"Entry already exists"}]}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	err = client.Allowlist.Create(ctx, &CreateAllowlistRequest{
		ProfileID: "test",
		Allowlist: []*Allowlist{{ID: "example.com", Active: true}},
	})

	c.True(err != nil)
}

func TestJSONRequestBodySerialization(t *testing.T) {
	c := is.New(t)

	var receivedBody map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		c.NoErr(err)

		err = json.Unmarshal(body, &receivedBody)
		c.NoErr(err)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{"id":"new-profile-id"}}`))
	}))
	defer ts.Close()

	client, err := New(WithBaseURL(ts.URL))
	c.NoErr(err)

	ctx := context.Background()
	_, err = client.Profiles.Create(ctx, &CreateProfileRequest{
		Name: "Test Profile",
		Security: &Security{
			ThreatIntelligenceFeeds: true,
			AiThreatDetection:       true,
		},
	})
	c.NoErr(err)

	c.Equal(receivedBody["name"], "Test Profile")

	security, ok := receivedBody["security"].(map[string]interface{})
	c.True(ok)
	c.Equal(security["threatIntelligenceFeeds"], true)
	c.Equal(security["aiThreatDetection"], true)
}
