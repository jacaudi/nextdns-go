// Example: list, create, update, get, and delete a NextDNS profile.
//
// Run:
//
//	NEXTDNS_API_KEY=... go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jacaudi/nextdns-go/nextdns"
)

func main() {
	apiKey := os.Getenv("NEXTDNS_API_KEY")
	if apiKey == "" {
		log.Fatal("NEXTDNS_API_KEY must be set")
	}

	client, err := nextdns.New(
		nextdns.WithAPIKey(nextdns.Secret(apiKey)),
	)
	if err != nil {
		log.Fatalf("client init: %v", err)
	}

	ctx := context.Background()

	// List existing profiles.
	list, err := client.Profiles.List(ctx, &nextdns.ListProfileRequest{})
	if err != nil {
		log.Fatalf("list profiles: %v", err)
	}
	fmt.Printf("Existing profiles (%d):\n", len(list.Profiles))
	for _, p := range list.Profiles {
		fmt.Printf("  %s — %s\n", p.ID, p.Name)
	}

	// Create a new profile (CAUTION: this affects your real account).
	created, err := client.Profiles.Create(ctx, &nextdns.CreateProfileRequest{
		Name: "nextdns-go example",
	})
	if err != nil {
		log.Fatalf("create profile: %v", err)
	}
	fmt.Printf("\nCreated profile %s\n", created)

	// Get details on the new profile. The returned *Profile has ID populated.
	profile, err := client.Profiles.Get(ctx, &nextdns.GetProfileRequest{ProfileID: created})
	if err != nil {
		log.Fatalf("get profile: %v", err)
	}
	fmt.Printf("Profile name: %s (fingerprint %s)\n", profile.Name, profile.Fingerprint)

	// Get → mutate → Update. We reuse the *Profile returned from Get; the
	// SDK's UpdateProfileRequest.MarshalJSON strips Profile.ID from the
	// PATCH body so the API contract isn't widened (the id is already in
	// the URL path).
	profile.Name = "nextdns-go example (updated)"
	err = client.Profiles.Update(ctx, &nextdns.UpdateProfileRequest{
		ProfileID: created,
		Profile:   profile,
	})
	if err != nil {
		// Branch on the typed *nextdns.Error to react to specific failure modes.
		var apiErr *nextdns.Error
		if errors.As(err, &apiErr) {
			log.Fatalf("update profile (%s): %s", apiErr.Type, apiErr.Message)
		}
		log.Fatalf("update profile: %v", err)
	}
	fmt.Println("Profile updated")

	// Delete the profile we just created.
	err = client.Profiles.Delete(ctx, &nextdns.DeleteProfileRequest{ProfileID: created})
	if err != nil {
		log.Fatalf("delete profile: %v", err)
	}
	fmt.Println("Profile deleted")
}
