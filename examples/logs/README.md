# Logs example

Demonstrates the three log access patterns:

- `query` — paginated query against `/logs`
- `download` — fetch the full CSV archive
- `stream` — real-time SSE stream

## Run

```bash
NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . query
NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . download
NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . stream
```
