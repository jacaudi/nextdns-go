package nextdns

import (
	"fmt"
	"io"
)

// Secret is a string wrapper that hides its content in default formatting,
// JSON marshaling, and text marshaling. Use Expose to retrieve the raw value
// when passing it to an API call or HTTP header.
//
// The intent is to make accidental secret exposure (fmt.Println, log lines,
// JSON dumps) grep-able and harder to produce. Per go-standards.md §15.2.
type Secret string

// String implements fmt.Stringer. Always returns "****".
func (s Secret) String() string { return "****" }

// MarshalJSON implements json.Marshaler. Always emits "****".
func (s Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"****"`), nil
}

// MarshalText implements encoding.TextMarshaler. Always emits "****".
func (s Secret) MarshalText() ([]byte, error) {
	return []byte("****"), nil
}

// Expose returns the underlying secret value. Name is intentional — it is
// grep-able in code review.
func (s Secret) Expose() string { return string(s) }

// Format implements fmt.Formatter. Every verb redacts to "****" — this closes
// two leak paths that the Stringer/Marshaler trio does not cover:
//
//   - %#v falls back to the underlying string (Secret is type Secret string),
//     printing the raw value.
//   - %d (or any wrong verb) embeds the raw value in fmt's bad-verb
//     diagnostic, e.g. "%!d(nextdns.Secret=raw-value)".
//
// Both patterns are extremely common in error wrapping (`fmt.Errorf("...: %#v", cfg)`)
// and in developer-debug logging. Per go-standards.md §15.2.
func (s Secret) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, "****")
}
