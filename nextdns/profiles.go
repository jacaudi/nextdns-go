package nextdns

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// profilesService is the HTTP path for the profiles API.
const profilesAPIPath = "profiles"

// CreateProfileRequest encapsulates the request for creating a new profile.
type CreateProfileRequest struct {
	Name            string           `json:"name,omitempty"`
	Security        *Security        `json:"security,omitempty"`
	Privacy         *Privacy         `json:"privacy,omitempty"`
	ParentalControl *ParentalControl `json:"parentalControl,omitempty"`
	Denylist        []*Denylist      `json:"denylist,omitempty"`
	Allowlist       []*Allowlist     `json:"allowlist,omitempty"`
	Settings        *Settings        `json:"settings,omitempty"`
	Rewrites        []*Rewrites      `json:"rewrites,omitempty"`
}

// UpdateProfileRequest encapsulates the request for setting custom profile settings.
type UpdateProfileRequest struct {
	ProfileID string
	Profile   *Profile
}

// GetProfileRequest encapsulates the request for getting a profile.
type GetProfileRequest struct {
	ProfileID string
}

// ListProfileRequest encapsulates the request for listing all the profiles.
type ListProfileRequest struct {
	Cursor string
}

// DeleteProfileRequest encapsulates the request for deleting a profile.
type DeleteProfileRequest struct {
	ProfileID string
}

// ProfilesService is an interface for communicating with the NextDNS API.
type ProfilesService interface {
	Create(context.Context, *CreateProfileRequest) (string, error)
	Get(context.Context, *GetProfileRequest) (*Profile, error)
	Update(context.Context, *UpdateProfileRequest) error
	List(context.Context, *ListProfileRequest) (*ListProfilesResponse, error)
	Delete(context.Context, *DeleteProfileRequest) error
}

// Profile represents a NextDNS profile.
type Profile struct {
	Name            string           `json:"name,omitempty"`
	Security        *Security        `json:"security,omitempty"`
	Privacy         *Privacy         `json:"privacy,omitempty"`
	ParentalControl *ParentalControl `json:"parentalControl,omitempty"`
	Denylist        []*Denylist      `json:"denylist,omitempty"`
	Allowlist       []*Allowlist     `json:"allowlist,omitempty"`
	Settings        *Settings        `json:"settings,omitempty"`
	Rewrites        []*Rewrites      `json:"rewrites,omitempty"`
	Setup           *Setup           `json:"setup,omitempty"`
}

// newProfileRequest represents the response from a new profile request.
type newProfileResponse struct {
	Profile struct {
		ID string `json:"id"`
	} `json:"data"`
}

// Profiles represents a list of NextDNS profiles.
type Profiles struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
	Name        string `json:"name"`
}

// profileResponse represents the response for the profile from the NextDNS API.
type profileResponse struct {
	Profile *Profile `json:"data"`
}

// profilesResponse represents the response for listing the profiles from the NextDNS API.
type profilesResponse struct {
	Profiles []*Profiles `json:"data"`
	Metadata struct {
		Pagination struct {
			Cursor string `json:"cursor"`
		} `json:"pagination"`
	} `json:"meta,omitempty"`
	Errors ErrorResponse `json:"errors,omitempty"`
}

// ListProfilesResponse represents the response from listing profiles with pagination info.
type ListProfilesResponse struct {
	Profiles []*Profiles
	Cursor   string // Next page cursor, empty if no more pages
}

// profilesService represents the NextDNS profiles service.
type profilesService struct {
	client *Client
}

var _ ProfilesService = &profilesService{}

// NewProfilesService returns a new NextDNS profiles service.
// nolint: revive
func NewProfilesService(client *Client) *profilesService {
	return &profilesService{
		client: client,
	}
}

// List returns a list of profiles with pagination support.
func (s *profilesService) List(ctx context.Context, request *ListProfileRequest) (*ListProfilesResponse, error) {
	var req *http.Request
	var err error

	if request != nil && request.Cursor != "" {
		query := url.Values{}
		query.Set("cursor", request.Cursor)
		req, err = s.client.newRequestWithQuery(http.MethodGet, profilesAPIPath, query, nil)
	} else {
		req, err = s.client.newRequest(http.MethodGet, profilesAPIPath, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("error creating request to list the profiles: %w", err)
	}

	response := profilesResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making a request to list the profiles: %w", err)
	}

	return &ListProfilesResponse{
		Profiles: response.Profiles,
		Cursor:   response.Metadata.Pagination.Cursor,
	}, nil
}

// Create creates a profile and returns a profile ID.
func (s *profilesService) Create(ctx context.Context, request *CreateProfileRequest) (string, error) {
	req, err := s.client.newRequest(http.MethodPost, profilesAPIPath, request)
	if err != nil {
		return "", fmt.Errorf("error creating request to create a profile: %w", err)
	}

	response := &newProfileResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return "", fmt.Errorf("error making a request to create a profile: %w", err)
	}

	return response.Profile.ID, nil
}

// Update updates the settings of a profile.
func (s *profilesService) Update(ctx context.Context, request *UpdateProfileRequest) error {
	path := fmt.Sprintf("%s/%s", profilesAPIPath, request.ProfileID)
	req, err := s.client.newRequest(http.MethodPatch, path, request.Profile)
	if err != nil {
		return fmt.Errorf("error creating request to update the profile: %w", err)
	}

	response := profileResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return fmt.Errorf("error making a request to update the profile: %w", err)
	}

	return nil
}

// Get returns a profile.
func (s *profilesService) Get(ctx context.Context, request *GetProfileRequest) (*Profile, error) {
	path := fmt.Sprintf("%s/%s", profilesAPIPath, request.ProfileID)
	req, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request to get the profile: %w", err)
	}

	response := profileResponse{}
	err = s.client.do(ctx, req, &response)
	if err != nil {
		return nil, fmt.Errorf("error making a request to get the profile: %w", err)
	}

	return response.Profile, nil
}

// Delete deletes a profile.
func (s *profilesService) Delete(ctx context.Context, request *DeleteProfileRequest) error {
	path := fmt.Sprintf("%s/%s", profilesAPIPath, request.ProfileID)
	req, err := s.client.newRequest(http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("error creating request to delete the profile: %w", err)
	}

	err = s.client.do(ctx, req, nil)
	if err != nil {
		return fmt.Errorf("error making a request to delete the profile: %w", err)
	}

	return err
}

// profileAPIPath returns the profile API path.
func profileAPIPath(profile string) string {
	return fmt.Sprintf("%s/%s", profilesAPIPath, profile)
}
