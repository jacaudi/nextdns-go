# nextdns-go

Go client library for the [NextDNS](https://nextdns.io/) API.

## Install

```bash
go get github.com/jacaudi/nextdns-go/nextdns
```

## Requirements

- Go 1.23+
- A NextDNS API key from https://my.nextdns.io/account

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jacaudi/nextdns-go/nextdns"
)

func main() {
	client, err := nextdns.New(
		nextdns.WithAPIKey(nextdns.Secret(os.Getenv("NEXTDNS_API_KEY"))),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	profiles, err := client.Profiles.List(ctx, &nextdns.ListProfileRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, p := range profiles.Profiles {
		fmt.Printf("%s — %s\n", p.ID, p.Name)
	}
}
```

## API Coverage

| Area | Status |
|---|---|
| Profiles | ✅ list, get, create, update, delete |
| Settings | ✅ (incl. logs, block page, performance) |
| Security | ✅ (incl. TLDs) |
| Privacy | ✅ (incl. blocklists, natives) |
| Parental Control | ✅ (incl. services, categories) |
| Allowlist / Denylist | ✅ |
| Rewrites | ✅ |
| Setup / Setup Linked IP | ✅ |
| Analytics | ✅ all 11 endpoint families (status, domains, queryTypes, reasons, ips, dnssec, encryption, ipVersions, protocols, destinations, devices) plus all `;series` variants |
| Logs | ✅ get, clear, download, downloadURL, stream (SSE) |

## Examples

Each feature has a runnable example in [`examples/`](./examples/):

- [`examples/profiles/`](./examples/profiles/) — list, create, update, delete a profile
- [`examples/denylist/`](./examples/denylist/) — manage a profile's denylist
- [`examples/analytics/`](./examples/analytics/) — query analytics and time series
- [`examples/logs/`](./examples/logs/) — query, download, and stream DNS logs

## Migration from v0.x

See [`MIGRATING.md`](./MIGRATING.md) for the v0.x → v1.0.0 migration guide.

## Workflow & Release Process

See [`WORKFLOW.md`](./WORKFLOW.md).

## License

MIT — see [`LICENSE.md`](./LICENSE.md).
