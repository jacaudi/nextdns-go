// Example: query status analytics and a time series for a profile.
//
// Run:
//
//	NEXTDNS_API_KEY=... NEXTDNS_PROFILE_ID=abc123 go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jacaudi/nextdns-go/nextdns"
)

func main() {
	apiKey := os.Getenv("NEXTDNS_API_KEY")
	profileID := os.Getenv("NEXTDNS_PROFILE_ID")
	if apiKey == "" || profileID == "" {
		log.Fatal("NEXTDNS_API_KEY and NEXTDNS_PROFILE_ID must be set")
	}

	client, err := nextdns.New(nextdns.WithAPIKey(nextdns.Secret(apiKey)))
	if err != nil {
		log.Fatalf("client init: %v", err)
	}

	ctx := context.Background()

	// Past 24h status counts.
	status, err := client.Analytics.GetStatus(ctx, &nextdns.GetAnalyticsRequest{
		ProfileID: profileID,
		Options:   &nextdns.AnalyticsOptions{From: "-24h"},
	})
	if err != nil {
		log.Fatalf("get status: %v", err)
	}
	fmt.Println("Status counts (last 24h):")
	for _, e := range status.Data {
		fmt.Printf("  %s — %d\n", e.ID, e.Queries)
	}

	// Hourly status time series.
	series, err := client.Analytics.GetStatusSeries(ctx, &nextdns.GetAnalyticsTimeSeriesRequest{
		ProfileID: profileID,
		Options: &nextdns.AnalyticsTimeSeriesOptions{
			AnalyticsOptions: nextdns.AnalyticsOptions{From: "-24h"},
			Interval:         "1h",
		},
	})
	if err != nil {
		log.Fatalf("get status series: %v", err)
	}
	fmt.Printf("\nStatus series (%d buckets, %ds interval):\n", len(series.Series.Times), series.Series.Interval)
	for _, e := range series.Data {
		fmt.Printf("  %s: %v\n", e.ID, e.Queries)
	}
}
