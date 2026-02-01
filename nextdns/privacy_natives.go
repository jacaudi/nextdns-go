package nextdns

import (
	"context"
	"fmt"
	"net/http"
)

// privacyNativesAPIPath is the HTTP path for the privacy native tracking protection API.
const privacyNativesAPIPath = "privacy/natives"

// privacyNativesIDAPIPath returns the HTTP path for a specific privacy native.
func privacyNativesIDAPIPath(id string) string {
	return fmt.Sprintf("%s/%s", privacyNativesAPIPath, id)
}

// PrivacyNatives represents a privacy native tracking protection of a profile.
type PrivacyNatives struct {
	ID string `json:"id"`
}

// CreatePrivacyNativesRequest encapsulates the request for creating a privacy native tracking protection list.
type CreatePrivacyNativesRequest struct {
	ProfileID      string
	PrivacyNatives []*PrivacyNatives
}

// ListPrivacyNativesRequest encapsulates the request for getting the privacy native tracking protection list.
type ListPrivacyNativesRequest struct {
	ProfileID string
}

// AddPrivacyNativesRequest encapsulates the request for adding a single privacy native.
type AddPrivacyNativesRequest struct {
	ProfileID string
	ID        string `json:"id"`
}

// PrivacyNativesService is an interface for communicating with the NextDNS privacy native tracking protection API endpoint.
type PrivacyNativesService interface {
	Create(context.Context, *CreatePrivacyNativesRequest) error
	List(context.Context, *ListPrivacyNativesRequest) ([]*PrivacyNatives, error)
	Add(context.Context, *AddPrivacyNativesRequest) error
}

// privacyNativesResponse represents the NextDNS privacy native tracking protection service.
type privacyNativesResponse struct {
	PrivacyNatives []*PrivacyNatives `json:"data"`
}

// privacyNativesService represents the NextDNS privacy native tracking protection service.
type privacyNativesService struct {
	client *Client
}

var _ PrivacyNativesService = &privacyNativesService{}

// NewPrivacyNativesService returns a new NextDNS privacy native tracking protection service.
// nolint: revive
func NewPrivacyNativesService(client *Client) *privacyNativesService {
	return &privacyNativesService{
		client: client,
	}
}

// Create creates a privacy native tracking protection list.
func (s *privacyNativesService) Create(ctx context.Context, request *CreatePrivacyNativesRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyNativesAPIPath)
	req, err := s.client.newRequest(http.MethodPut, path, request.PrivacyNatives)
	if err != nil {
		return fmt.Errorf("error creating request to create a privacy native list: %w", err)
	}

	response := privacyNativesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return fmt.Errorf("error making a request to create a privacy native list: %w", err)
	}

	return nil
}

// List returns the privacy native tracking protection list.
func (s *privacyNativesService) List(ctx context.Context, request *ListPrivacyNativesRequest) ([]*PrivacyNatives, error) {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyNativesAPIPath)
	req, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to list the privacy native list: %w", err)
	}

	response := privacyNativesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making a request to list the privacy native list: %w", err)
	}

	return response.PrivacyNatives, nil
}

// Add adds a single native tracking protection.
func (s *privacyNativesService) Add(ctx context.Context, request *AddPrivacyNativesRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), privacyNativesAPIPath)
	body := struct {
		ID string `json:"id"`
	}{
		ID: request.ID,
	}
	req, err := s.client.newRequest(http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("error creating request to add privacy native %s: %w", request.ID, err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making request to add privacy native %s: %w", request.ID, err)
	}

	return nil
}
