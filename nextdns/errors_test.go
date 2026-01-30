package nextdns

import (
	"errors"
	"testing"

	"github.com/matryer/is"
)

func TestAPIError_Error_CodeOnly(t *testing.T) {
	c := is.New(t)

	err := &APIError{Code: "invalidDomain"}

	c.Equal(err.Error(), "invalidDomain")
}

func TestAPIError_Error_WithDetail(t *testing.T) {
	c := is.New(t)

	err := &APIError{
		Code:   "invalidDomain",
		Detail: "The domain format is invalid",
	}

	c.Equal(err.Error(), "The domain format is invalid [invalidDomain]")
}

func TestAPIError_Error_WithParameter(t *testing.T) {
	c := is.New(t)

	err := &APIError{
		Code:      "required",
		Parameter: "name",
	}

	c.Equal(err.Error(), "required (parameter: name)")
}

func TestAPIError_Error_WithDetailAndParameter(t *testing.T) {
	c := is.New(t)

	err := &APIError{
		Code:      "invalid",
		Detail:    "Invalid value provided",
		Parameter: "domain",
	}

	c.Equal(err.Error(), "Invalid value provided [invalid] (parameter: domain)")
}

func TestAPIError_Is(t *testing.T) {
	c := is.New(t)

	err := &APIError{Code: "duplicate", Detail: "Entry already exists"}
	target := &APIError{Code: "duplicate"}

	c.True(errors.Is(err, target))
}

func TestAPIError_Is_NoMatch(t *testing.T) {
	c := is.New(t)

	err := &APIError{Code: "duplicate"}
	target := &APIError{Code: "invalidDomain"}

	c.True(!errors.Is(err, target))
}

func TestError_Error_NoAPIErrors(t *testing.T) {
	c := is.New(t)

	err := &Error{
		Type:    ErrorTypeRequest,
		Message: "request failed",
	}

	c.Equal(err.Error(), "request failed (request)")
}

func TestError_Error_SingleAPIError(t *testing.T) {
	c := is.New(t)

	err := &Error{
		Type:    ErrorTypeRequest,
		Message: "response error received",
		Errors: &ErrorResponse{
			Errors: []struct {
				Code   string `json:"code"`
				Detail string `json:"detail,omitempty"`
				Source struct {
					Parameter string `json:"parameter,omitempty"`
				} `json:"source,omitempty"`
			}{
				{Code: "invalidDomain", Detail: "Invalid domain format"},
			},
		},
	}

	c.Equal(err.Error(), "response error received (request): Invalid domain format [invalidDomain]")
}

func TestError_Error_MultipleAPIErrors(t *testing.T) {
	c := is.New(t)

	err := &Error{
		Type:    ErrorTypeRequest,
		Message: "response error received",
		Errors: &ErrorResponse{
			Errors: []struct {
				Code   string `json:"code"`
				Detail string `json:"detail,omitempty"`
				Source struct {
					Parameter string `json:"parameter,omitempty"`
				} `json:"source,omitempty"`
			}{
				{Code: "invalidDomain", Detail: "Invalid domain format"},
				{Code: "duplicate"},
			},
		},
	}

	c.Equal(err.Error(), "response error received (request): Invalid domain format [invalidDomain]; duplicate")
}

func TestError_Error_WithParameter(t *testing.T) {
	c := is.New(t)

	err := &Error{
		Type:    ErrorTypeRequest,
		Message: "response error received",
		Errors: &ErrorResponse{
			Errors: []struct {
				Code   string `json:"code"`
				Detail string `json:"detail,omitempty"`
				Source struct {
					Parameter string `json:"parameter,omitempty"`
				} `json:"source,omitempty"`
			}{
				{
					Code:   "required",
					Detail: "Field is required",
					Source: struct {
						Parameter string `json:"parameter,omitempty"`
					}{Parameter: "name"},
				},
			},
		},
	}

	c.Equal(err.Error(), "response error received (request): Field is required [required] (parameter: name)")
}
