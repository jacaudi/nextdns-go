package nextdns

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorType defines the code of an error.
type ErrorType string

// ErrEmptyAPIToken is returned when an empty API token is provided during client initialization.
var ErrEmptyAPIToken = errors.New("api key must not be empty")

const (
	errInternalServiceError = "internal service error received"
	errResponseError        = "response error received"
	errMalformedError       = "malformed response body received"
	errMalformedErrorBody   = "malformed error response body received"
)

// ErrorType constants classify errors returned by the NextDNS Client.
const (
	ErrorTypeServiceError   ErrorType = "service_error"  // Internal service error.
	ErrorTypeRequest        ErrorType = "request"        // Regular request error.
	ErrorTypeMalformed      ErrorType = "malformed"      // Response body is malformed.
	ErrorTypeAuthentication ErrorType = "authentication" // Authentication error.
	ErrorTypeNotFound       ErrorType = "not_found"      // Resource not found.
)

// ErrorResponse represents the error response from the NextDNS API.
type ErrorResponse struct {
	Errors []struct {
		Code   string `json:"code"`
		Detail string `json:"detail,omitempty"`
		Source struct {
			Parameter string `json:"parameter,omitempty"`
		} `json:"source,omitempty"`
	} `json:"errors"`
}

// Error represents the error from the Client.
type Error struct {
	Type    ErrorType
	Message string
	Errors  *ErrorResponse
	Meta    map[string]string
}

// APIError represents a single error from the NextDNS API.
type APIError struct {
	Code      string
	Detail    string
	Parameter string
}

// Error returns the string representation of the API error.
func (e *APIError) Error() string {
	if e.Detail != "" {
		if e.Parameter != "" {
			return fmt.Sprintf("%s [%s] (parameter: %s)", e.Detail, e.Code, e.Parameter)
		}
		return fmt.Sprintf("%s [%s]", e.Detail, e.Code)
	}
	if e.Parameter != "" {
		return fmt.Sprintf("%s (parameter: %s)", e.Code, e.Parameter)
	}
	return e.Code
}

// Is reports whether the error matches the target by comparing error codes.
func (e *APIError) Is(target error) bool {
	t, ok := target.(*APIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Error returns the string representation of the error.
// TODO(jacaudi): Improve error handling for multiple errors. See https://github.com/jacaudi/nextdns-go/issues/7
func (e *Error) Error() string {
	var out strings.Builder
	out.WriteString(fmt.Sprintf("%s (%s)", e.Message, e.Type))

	if e.Errors != nil && len(e.Errors.Errors) > 0 {
		out.WriteString(": ")
		for i, er := range e.Errors.Errors {
			if i > 0 {
				out.WriteString("; ")
			}
			if er.Detail != "" {
				out.WriteString(fmt.Sprintf("%s [%s]", er.Detail, er.Code))
			} else {
				out.WriteString(er.Code)
			}
			if er.Source.Parameter != "" {
				out.WriteString(fmt.Sprintf(" (parameter: %s)", er.Source.Parameter))
			}
		}
	}

	return out.String()
}

// Unwrap returns the underlying API errors for use with errors.Is and errors.As.
// Returns nil if there are no underlying API errors.
func (e *Error) Unwrap() []error {
	if e.Errors == nil || len(e.Errors.Errors) == 0 {
		return nil
	}

	errs := make([]error, len(e.Errors.Errors))
	for i, apiErr := range e.Errors.Errors {
		errs[i] = &APIError{
			Code:      apiErr.Code,
			Detail:    apiErr.Detail,
			Parameter: apiErr.Source.Parameter,
		}
	}
	return errs
}
