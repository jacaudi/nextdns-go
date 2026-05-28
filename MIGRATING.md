# Migrating from v0.x to v1.0.0

This guide walks through each breaking change and the migration steps.

## 1. `WithAPIKey` takes `Secret`

```go
// v0.x
client, _ := nextdns.New(nextdns.WithAPIKey("my-key"))

// v1.0.0
client, _ := nextdns.New(nextdns.WithAPIKey(nextdns.Secret("my-key")))
```

`Secret` is a string wrapper that hides its value in fmt and JSON output. Call `.Expose()` to retrieve the underlying string when needed.

## 2. `Debug` removed; use `WithLogger`

```go
// v0.x
client, _ := nextdns.New(nextdns.WithDebug())

// v1.0.0
import "log/slog"
import "os"

logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
client, _ := nextdns.New(nextdns.WithLogger(logger))
```

Debug output now goes through `slog` at `LevelDebug`. The SDK never writes to stdout.

## 3. `Profiles` (list-item) renamed to `ProfileSummary`

```go
// v0.x
var p nextdns.Profiles

// v1.0.0
var p nextdns.ProfileSummary
```

`Profile` (the detail type) is unchanged. The plural-vs-singular ambiguity is gone.

## 4. Service constructors return interfaces

```go
// v0.x — these returned *profilesService (unexported)
srv := nextdns.NewProfilesService(client)

// v1.0.0 — these return ProfilesService (interface)
srv := nextdns.NewProfilesService(client)
```

The change is invisible in nearly all consumer code. Type assertions on the returned value need updating.

## 5. Strong-typed enums

```go
// v0.x
opts := &nextdns.AnalyticsOptions{...}
client.Analytics.GetDomains(ctx, &nextdns.GetAnalyticsDomainsRequest{
    Status: "blocked",
})

// v1.0.0
client.Analytics.GetDomains(ctx, &nextdns.GetAnalyticsDomainsRequest{
    Status: nextdns.StatusBlocked,
})
```

Affected types: `LogStatus`, `SortOrder`, `DestinationType`, `DNSProtocol`, `LogRetention`.

## 6. `Add` requests use named structs

```go
// v0.x — denylist.Add accepted an AddDenylistRequest but security_tlds.Add
// used an anonymous body struct internally; in tests you may have constructed
// these inconsistently.

// v1.0.0 — every Add takes a named request:
client.Denylist.Add(ctx, &nextdns.AddDenylistRequest{
    ProfileID: "abc",
    ID:        "example.com",
})
client.SecurityTlds.Add(ctx, &nextdns.AddSecurityTldsRequest{
    ProfileID: "abc",
    ID:        "xyz",
})
```

## 7. HTTP client default changed

The default `*http.Client` now has:

- Overall `Timeout: 30s`
- TLS 1.3 minimum (`MinVersion: tls.VersionTLS13`)
- Tuned dial / handshake / response-header timeouts (5s, 5s, 10s)
- HTTP/2 preferred (`ForceAttemptHTTP2: true`)

If you depended on `cleanhttp.DefaultClient`'s specific behavior (no overall timeout, TLS 1.2 floor), pass your own client via `WithHTTPClient`:

```go
client, _ := nextdns.New(
    nextdns.WithHTTPClient(&http.Client{Timeout: 0, /* ... */}),
)
```

For long-lived streaming like `Logs.Stream`, the SDK internally clones the client with `Timeout: 0` to bypass the overall timeout. No consumer action required.

## 8. New features (no migration required)

These are additive — no v0 code breaks:

- `Profile.ID` field is now populated on Get.
- `Logs.Download`, `Logs.DownloadURL`, `Logs.Stream` (with `LogEntry.ID` for SSE event tracking and `io.EOF` signalling a clean end of stream).
- 7 new analytics endpoint families.
- `Delete` on parental-control services and categories.
- `WithRateLimitObserver`.

See [CHANGELOG.md](./CHANGELOG.md) for the complete list.
