// Package nextdns is a Go client for the NextDNS API.
//
// # Quick start
//
//	import "github.com/jacaudi/nextdns-go/nextdns"
//
//	client, err := nextdns.New(
//	    nextdns.WithAPIKey(nextdns.Secret(os.Getenv("NEXTDNS_API_KEY"))),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	profile, err := client.Profiles.Get(ctx, &nextdns.GetProfileRequest{ProfileID: "abc123"})
//
// # Authentication
//
// All requests require an API key (https://my.nextdns.io/account) supplied via
// the WithAPIKey client option. The Secret wrapper hides the value in fmt and
// JSON output to reduce accidental leakage. See go-standards.md §15.2.
//
// # HTTP defaults
//
// The default *http.Client has a 30 second overall timeout, a TLS 1.3 floor,
// and a tuned transport (5s dial, 5s TLS handshake, 10s response header).
// Override with WithHTTPClient when you need different behavior — for example,
// to disable the overall timeout for long-lived streaming endpoints.
//
// # Services
//
// The client exposes one Service per logical area of the NextDNS API:
//   - Profiles, Settings (and sub-services: Logs, BlockPage, Performance)
//   - Security (and SecurityTlds), Privacy (and Blocklists, Natives)
//   - ParentalControl (and Services, Categories)
//   - Allowlist, Denylist, Rewrites
//   - Analytics, Logs, Setup, SetupLinkedIP
//
// # Errors
//
// API errors are wrapped in *Error. Use the helpers IsNotFound, IsAuthError,
// IsDuplicateError, or HasErrorCode to branch on common cases. Underlying
// errors are accessible via errors.As (*APIError).
//
// # Rate limiting
//
// NextDNS returns X-RateLimit-* headers on every response. Subscribe via
// WithRateLimitObserver to receive a callback with the parsed values.
//
// # Streaming
//
// Logs.Stream consumes /logs/stream as a Go 1.23 range-over-func iterator:
//
//	for entry, err := range client.Logs.Stream(ctx, req) {
//	    if err != nil { break }
//	    process(entry)
//	}
//
// Known limitations:
//   - Manual reconnect via StreamLogsRequest.LastID is documented but not
//     yet wired: the SDK currently ignores the SSE id: lines, so callers
//     cannot learn the last event ID from a stream. Track external state
//     (e.g., the timestamp of the last entry) if you need to resume.
//   - The streaming HTTP client clones the SDK default and zeroes its
//     Timeout, but shares the Transport. Each concurrent stream consumes
//     one slot from MaxIdleConnsPerHost (default 10). Applications running
//     many concurrent streams should pass a custom *http.Client via
//     WithHTTPClient with a larger MaxConnsPerHost.
//
// See examples/ for runnable demos.
package nextdns
