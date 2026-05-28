// Example: query, download, and stream DNS logs.
//
// Run query:    NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . query
// Run download: NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . download
// Run stream:   NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run . stream
//
// Set NEXTDNS_DEBUG=1 to enable structured request/response logging at
// slog.LevelDebug; otherwise the SDK is silent.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacaudi/nextdns-go/nextdns"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: logs <query|download|stream>")
	}
	apiKey := os.Getenv("NEXTDNS_API_KEY")
	profileID := os.Getenv("NEXTDNS_PROFILE_ID")
	if apiKey == "" || profileID == "" {
		log.Fatal("NEXTDNS_API_KEY and NEXTDNS_PROFILE_ID must be set")
	}

	opts := []nextdns.ClientOption{
		nextdns.WithAPIKey(nextdns.Secret(apiKey)),
		// Observe rate-limit headers on every response — even 4xx/5xx.
		// Surface a warning when the remaining budget gets thin.
		nextdns.WithRateLimitObserver(func(rl nextdns.RateLimit) {
			if rl.Limit > 0 && rl.Remaining < 50 {
				log.Printf("rate limit low: %d/%d remaining (resets %s)",
					rl.Remaining, rl.Limit, rl.Reset.Format("15:04:05"))
			}
		}),
	}
	// WithLogger gates SDK request/response tracing on a slog handler;
	// the SDK only logs at slog.LevelDebug, so callers control verbosity
	// by choosing the handler's minimum level.
	if os.Getenv("NEXTDNS_DEBUG") == "1" {
		opts = append(opts, nextdns.WithLogger(
			slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
		))
	}

	client, err := nextdns.New(opts...)
	if err != nil {
		log.Fatalf("client init: %v", err)
	}

	switch os.Args[1] {
	case "query":
		queryLogs(client, profileID)
	case "download":
		downloadLogs(client, profileID)
	case "stream":
		streamLogs(client, profileID)
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

func queryLogs(client *nextdns.Client, profileID string) {
	resp, err := client.Logs.Get(context.Background(), &nextdns.GetLogsRequest{
		ProfileID: profileID,
		Options:   &nextdns.LogsQueryOptions{Limit: 10, From: "-1h"},
	})
	if err != nil {
		log.Fatalf("get logs: %v", err)
	}
	for _, e := range resp.Data {
		fmt.Printf("%s %s %s %s\n", e.Timestamp.Format("15:04:05"), e.Status, e.Protocol, e.Domain)
	}
}

func downloadLogs(client *nextdns.Client, profileID string) {
	body, err := client.Logs.Download(context.Background(), &nextdns.DownloadLogsRequest{ProfileID: profileID})
	if err != nil {
		log.Fatalf("download logs: %v", err)
	}
	defer body.Close()

	f, err := os.Create("logs.csv")
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	defer f.Close()

	n, err := io.Copy(f, body)
	if err != nil {
		log.Fatalf("copy: %v", err)
	}
	fmt.Printf("Wrote %d bytes to logs.csv\n", n)
}

func streamLogs(client *nextdns.Client, profileID string) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	fmt.Println("Streaming logs — press Ctrl-C to stop.")
	var lastID string
	for entry, err := range client.Logs.Stream(ctx, &nextdns.StreamLogsRequest{
		ProfileID: profileID,
		LastID:    lastID,
	}) {
		if errors.Is(err, io.EOF) {
			log.Printf("stream ended cleanly; last seen id=%s", lastID)
			return
		}
		if err != nil {
			log.Printf("stream error: %v (resume from id=%s)", err, lastID)
			return
		}
		lastID = entry.ID
		fmt.Printf("%s %s %s %s\n", entry.Timestamp.Format("15:04:05"), entry.Status, entry.Protocol, entry.Domain)
	}
}
