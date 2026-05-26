package nextdns

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
