// Example: list, add, update, and remove denylist entries.
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

	// Add a single entry.
	err = client.Denylist.Add(ctx, &nextdns.AddDenylistRequest{
		ProfileID: profileID,
		ID:        "example.com",
	})
	if err != nil {
		log.Fatalf("add denylist entry: %v", err)
	}
	fmt.Println("Added example.com to denylist")

	// List the denylist.
	entries, err := client.Denylist.List(ctx, &nextdns.ListDenylistRequest{ProfileID: profileID})
	if err != nil {
		log.Fatalf("list denylist: %v", err)
	}
	fmt.Printf("Denylist (%d entries):\n", len(entries))
	for _, e := range entries {
		fmt.Printf("  %s active=%v\n", e.ID, e.Active)
	}

	// Disable the entry we just added.
	inactive := false
	err = client.Denylist.Update(ctx, &nextdns.UpdateDenylistRequest{
		ProfileID: profileID,
		ID:        "example.com",
		Denylist:  &nextdns.Denylist{Active: inactive},
	})
	if err != nil {
		log.Fatalf("update denylist entry: %v", err)
	}
	fmt.Println("Disabled example.com")

	// Remove it.
	err = client.Denylist.Delete(ctx, &nextdns.DeleteDenylistRequest{
		ProfileID: profileID,
		ID:        "example.com",
	})
	if err != nil {
		log.Fatalf("delete denylist entry: %v", err)
	}
	fmt.Println("Removed example.com")
}
