# nextdns-go — Agent Notes

This is the public Go SDK for the NextDNS API.

## Layout

- `nextdns/` — the single domain package. SDK code, types, services.
- `examples/` — one runnable example per major feature, each with its own `go.mod`.
- `docs/` — published documentation (`migrating.md`, `workflow.md`).
- `docs/plans/` — design and implementation plans (gitignored).
- `docs/review/` — review notes (gitignored).
- `docs/prompts/` — session handoff artifacts (gitignored).

## Build / test

```bash
go test ./...           # unit tests
golangci-lint run ./... # lint
go vet ./...
```

The Makefile wires the standard targets: `make test`, `make lint`, `make vet`, `make clean`, `make coverage`, `make tparse`.

## Gotchas

- **`Profile` (detail) vs `ProfileSummary` (list-item)** are two distinct types. The list endpoint returns `[]*ProfileSummary`; `Get` returns `*Profile`.
- **`Secret` type wraps the API key.** Use `.Expose()` only at the HTTP boundary.
- **`Logs.Stream` uses `iter.Seq2`** (Go 1.23+ range-over-func). It clones the HTTP client to set `Timeout: 0` because SSE connections are long-lived.
- **Service constructors return interfaces** (`ProfilesService`, etc.), not the unexported struct.
- **Strong-typed enums:** `LogStatus`, `SortOrder`, `DestinationType`, `DNSProtocol`, `LogRetention`. Don't use string literals where these apply — the compiler will catch typos.
- **Analytics families:** 11 total (status, domains, queryTypes, reasons, ips, dnssec, encryption, ipVersions, protocols, destinations, devices), each with a `;series` variant. The `dnssec`/`encryption`/`ipVersions`/`protocols` families have typed response structs because their data shapes differ from the generic `id`/`name`/`queries`.

## Reference

- Go standards: `~/.config/reference/go-standards.md`
- NextDNS public docs: https://nextdns.github.io/api/
- Design: `docs/plans/2026-05-25-standards-alignment-and-coverage-design.md`
- Implementation plan: `docs/plans/2026-05-25-standards-alignment-and-coverage-implementation.md`
