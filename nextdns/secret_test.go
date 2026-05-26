package nextdns

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/matryer/is"
)

func TestSecretString(t *testing.T) {
	c := is.New(t)
	s := Secret("supersecret")
	c.Equal(s.String(), "****")
	c.Equal(fmt.Sprintf("%v", s), "****")
	c.Equal(fmt.Sprintf("%s", s), "****") //nolint:staticcheck // intentional: verifies fmt.Stringer routing via %s verb
}

func TestSecretMarshalJSON(t *testing.T) {
	c := is.New(t)
	s := Secret("supersecret")
	b, err := json.Marshal(s)
	c.NoErr(err)
	c.Equal(string(b), `"****"`)
}

func TestSecretMarshalText(t *testing.T) {
	c := is.New(t)
	s := Secret("supersecret")
	b, err := s.MarshalText()
	c.NoErr(err)
	c.Equal(string(b), "****")
}

func TestSecretExpose(t *testing.T) {
	c := is.New(t)
	s := Secret("supersecret")
	c.Equal(s.Expose(), "supersecret")
}
