package nextdns

import (
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
