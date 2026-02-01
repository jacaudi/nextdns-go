package nextdns

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// privacyBlocklistsAPIPath is the HTTP path for the privacy blocklist API.
const privacyBlocklistsAPIPath = "privacy/blocklists"

// privacyBlocklistsIDAPIPath returns the HTTP path for a specific privacy blocklist.
func privacyBlocklistsIDAPIPath(id string) string {
	return fmt.Sprintf("%s/%s", privacyBlocklistsAPIPath, id)
}

// PrivacyBlocklists represents a privacy blocklist of a profile.
type PrivacyBlocklists struct {
	ID        string     `json:"id"`
	Name      string     `json:"name,omitempty"`
	Website   string     `json:"website,omitempty"`
	Entries   int        `json:"entries,omitempty"`
	UpdatedOn *time.Time `json:"updatedOn,omitempty"`
}

// CreatePrivacyBlocklistsRequest encapsulates the request for creating a privacy blocklist.
type CreatePrivacyBlocklistsRequest struct {
	ProfileID         string
	PrivacyBlocklists []*PrivacyBlocklists
}

// ListPrivacyBlocklistsRequest encapsulates the request for getting the privacy blocklist.
type ListPrivacyBlocklistsRequest struct {
	ProfileID string
}

// AddPrivacyBlocklistsRequest encapsulates the request for adding a single privacy blocklist.
type AddPrivacyBlocklistsRequest struct {
	ProfileID string
	ID        string `json:"id"`
}

// UpdatePrivacyBlocklistsRequest encapsulates the request for updating a privacy blocklist.
type UpdatePrivacyBlocklistsRequest struct {
	ProfileID   string
	BlocklistID string
	Active      *bool `json:"active,omitempty"`
}

// PrivacyBlocklistsService is an interface for communicating with the NextDNS privacy blocklist API endpoint.
type PrivacyBlocklistsService interface {
	Create(context.Context, *CreatePrivacyBlocklistsRequest) error
	List(context.Context, *ListPrivacyBlocklistsRequest) ([]*PrivacyBlocklists, error)
	Add(context.Context, *AddPrivacyBlocklistsRequest) error
	Update(context.Context, *UpdatePrivacyBlocklistsRequest) error
}

// privacyBlocklistsResponse represents the NextDNS privacy blocklist service.
type privacyBlocklistsResponse struct {
	PrivacyBlocklists []*PrivacyBlocklists `json:"data"`
}

// privacyBlocklistsService represents the NextDNS privacy blocklist service.
type privacyBlocklistsService struct {
	client *Client
}

var _ PrivacyBlocklistsService = &privacyBlocklistsService{}

// NewPrivacyBlocklistsService returns a new NextDNS privacy blocklist service.
// nolint: revive
func NewPrivacyBlocklistsService(client *Client) *privacyBlocklistsService {
	return &privacyBlocklistsService{
		client: client,
	}
}

// Create creates a privacy blocklist list for a profile.
func (s *privacyBlocklistsService) Create(ctx context.Context, request *CreatePrivacyBlocklistsRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyBlocklistsAPIPath)
	req, err := s.client.newRequest(http.MethodPut, path, request.PrivacyBlocklists)
	if err != nil {
		return fmt.Errorf("error creating request to create a privacy blocklist: %w", err)
	}

	response := privacyBlocklistsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return fmt.Errorf("error making a request to create a privacy blocklist: %w", err)
	}

	return nil
}

// List returns the privacy blocklist for a profile.
func (s *privacyBlocklistsService) List(ctx context.Context, request *ListPrivacyBlocklistsRequest) ([]*PrivacyBlocklists, error) {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyBlocklistsAPIPath)
	req, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to list the privacy blocklist: %w", err)
	}

	response := privacyBlocklistsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making a request to list the privacy blocklist: %w", err)
	}

	return response.PrivacyBlocklists, nil
}

// Add adds a single blocklist to the privacy settings.
func (s *privacyBlocklistsService) Add(ctx context.Context, request *AddPrivacyBlocklistsRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyBlocklistsAPIPath)
	body := struct {
		ID string `json:"id"`
	}{
		ID: request.ID,
	}
	req, err := s.client.newRequest(http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("error creating request to add privacy blocklist %s: %w", request.ID, err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making request to add privacy blocklist %s: %w", request.ID, err)
	}

	return nil
}

// Update modifies a single blocklist entry.
func (s *privacyBlocklistsService) Update(ctx context.Context, request *UpdatePrivacyBlocklistsRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyBlocklistsIDAPIPath(request.BlocklistID))
	body := struct {
		Active *bool `json:"active,omitempty"`
	}{
		Active: request.Active,
	}
	req, err := s.client.newRequest(http.MethodPatch, path, body)
	if err != nil {
		return fmt.Errorf("error creating request to update privacy blocklist %s: %w", request.BlocklistID, err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making request to update privacy blocklist %s: %w", request.BlocklistID, err)
	}

	return nil
}
