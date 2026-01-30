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
