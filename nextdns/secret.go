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

// redacted is the single source of truth for the redaction marker emitted by
// String, MarshalJSON, MarshalText, and Format.
const redacted = "****"

// String implements fmt.Stringer. Always returns the redaction marker.
func (s Secret) String() string { return redacted }

// MarshalJSON implements json.Marshaler. Always emits the redaction marker.
func (s Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"` + redacted + `"`), nil
}

// MarshalText implements encoding.TextMarshaler. Always emits the redaction
// marker.
func (s Secret) MarshalText() ([]byte, error) {
	return []byte(redacted), nil
}

// Expose returns the underlying secret value. Name is intentional — it is
// grep-able in code review.
func (s Secret) Expose() string { return string(s) }

// Format implements fmt.Formatter. Verbs routed through fmt.Formatter redact
// to "****" — this closes two leak paths that the Stringer/Marshaler trio
// does not cover:
//
//   - %#v falls back to the underlying string (Secret is type Secret string),
//     printing the raw value.
//   - Any wrong verb (e.g. %d) embeds the raw value in fmt's bad-verb
//     diagnostic, e.g. "%!d(nextdns.Secret=raw-value)".
//
// Both patterns are extremely common in error wrapping (fmt.Errorf("...: %#v", cfg))
// and in developer-debug logging. Per go-standards.md §15.2.
//
// Carve-outs in the fmt package contract: %T and %p bypass fmt.Formatter.
//   - %T prints the Go type name ("nextdns.Secret") — safe, not a leak.
//   - %p on a value-typed Secret produces fmt's bad-verb diagnostic with the
//     raw value embedded. Callers must not format Secret with %p.
//
// Width and precision flags (e.g. %10s, %.3s) are ignored — the output is
// always exactly "****".
func (s Secret) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, redacted)
}
