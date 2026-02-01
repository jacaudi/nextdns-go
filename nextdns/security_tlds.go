package nextdns

import (
	"context"
	"fmt"
	"net/http"
)

// securityTldsAPIPath is the HTTP path for the security TLDs API.
const securityTldsAPIPath = "security/tlds"

// securityTldsIDAPIPath returns the HTTP path for a specific security TLD.
func securityTldsIDAPIPath(id string) string {
	return fmt.Sprintf("%s/%s", securityTldsAPIPath, id)
}

// SecurityTlds represents the security TLDs of a profile.
type SecurityTlds struct {
	ID string `json:"id"`
}

// CreateSecurityTldsRequest encapsulates the request for creating a security TLDs list.
type CreateSecurityTldsRequest struct {
	ProfileID    string
	SecurityTlds []*SecurityTlds
}

// ListSecurityTldsRequest encapsulates the request for getting a security TLDs list.
type ListSecurityTldsRequest struct {
	ProfileID string
}

// AddSecurityTldsRequest encapsulates the request for adding a single security TLD.
type AddSecurityTldsRequest struct {
	ProfileID string
	ID        string `json:"id"`
}

// SecurityTldsService is an interface for communicating with the NextDNS security TLDs API endpoint.
type SecurityTldsService interface {
	Create(context.Context, *CreateSecurityTldsRequest) error
	List(context.Context, *ListSecurityTldsRequest) ([]*SecurityTlds, error)
	Add(context.Context, *AddSecurityTldsRequest) error
}

// securityTldsResponse represents the security TLDs response.
type securityTldsResponse struct {
	SecurityTlds []*SecurityTlds `json:"data"`
}

// securityTldsService represents the NextDNS security TLDs service.
type securityTldsService struct {
	client *Client
}

var _ SecurityTldsService = &securityTldsService{}

// NewSecurityTldsService returns a new NextDNS security TLDs service.
// nolint: revive
func NewSecurityTldsService(client *Client) *securityTldsService {
	return &securityTldsService{
		client: client,
	}
}

// Create creates a security TLDs list.
func (s *securityTldsService) Create(ctx context.Context, request *CreateSecurityTldsRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), securityTldsAPIPath)
	req, err := s.client.newRequest(http.MethodPut, path, request.SecurityTlds)
	if err != nil {
		return fmt.Errorf("error creating request to create a security tlds list: %w", err)
	}

	response := securityTldsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return fmt.Errorf("error making a request to create a security tlds list: %w", err)
	}

	return nil
}

// List returns a security TLDs list.
func (s *securityTldsService) List(ctx context.Context, request *ListSecurityTldsRequest) ([]*SecurityTlds, error) {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), securityTldsAPIPath)
	req, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to list the security tlds list: %w", err)
	}

	response := securityTldsResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making a request to list the security tlds list: %w", err)
	}

	return response.SecurityTlds, nil
}

// Add adds a single TLD to the blocked list.
func (s *securityTldsService) Add(ctx context.Context, request *AddSecurityTldsRequest) error {
	path := fmt.Sprintf("%s/%s", profileAPIPath(request.ProfileID), securityTldsAPIPath)
	body := struct {
		ID string `json:"id"`
	}{
		ID: request.ID,
	}
	req, err := s.client.newRequest(http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("error creating request to add security TLD %s: %w", request.ID, err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making request to add security TLD %s: %w", request.ID, err)
	}

	return nil
}
