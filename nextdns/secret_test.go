package nextdns

import (
	"encoding/json"
	"fmt"
	"strings"
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

func TestSecretFormatRedactsAllVerbs(t *testing.T) {
	c := is.New(t)
	s := Secret("supersecret")

	// Stringer path (already covered by TestSecretString) — keep.
	c.Equal(fmt.Sprintf("%v", s), "****")
	c.Equal(fmt.Sprintf("%s", s), "****")

	// New: Format must intercept Go-syntax and wrong-verb paths too.
	c.Equal(fmt.Sprintf("%#v", s), "****") // was "supersecret" pre-fix
	c.Equal(fmt.Sprintf("%q", s), "****")
	c.Equal(fmt.Sprintf("%d", s), "****") // was "%!d(nextdns.Secret=supersecret)"
	c.Equal(fmt.Sprintf("%x", s), "****")

	// Struct-embedded path — common via fmt.Errorf("...: %#v", cfg).
	type Container struct{ Key Secret }
	out := fmt.Sprintf("%#v", Container{Key: s})
	c.True(strings.Contains(out, "****"))
	c.True(!strings.Contains(out, "supersecret"))
}
