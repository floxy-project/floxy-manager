package ldap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rom8726/floxy-manager/internal/domain"
)

// AuthService provides LDAP authentication functionality.
type AuthService struct {
	service *Service
	domain  string // Optional domain for username matching
}

// NewAuthService creates a new LDAP authentication service.
func NewAuthService(service *Service) *AuthService {
	var domainLDAP string

	if service == nil || service.client == nil {
		return &AuthService{
			service: service,
			domain:  domainLDAP,
		}
	}

	// Extract domain from the LDAP URL if available
	client, ok := service.client.(*Client)
	if !ok {
		return &AuthService{
			service: service,
			domain:  domainLDAP,
		}
	}

	url := client.config.URL
	if strings.HasPrefix(url, "ldap://") {
		domainLDAP = strings.TrimPrefix(url, "ldap://")
	} else if strings.HasPrefix(url, "ldaps://") {
		domainLDAP = strings.TrimPrefix(url, "ldaps://")
	}
	// Remove port if present
	if idx := strings.Index(domainLDAP, ":"); idx != -1 {
		domainLDAP = domainLDAP[:idx]
	}

	return &AuthService{
		service: service,
		domain:  domainLDAP,
	}
}

// Authenticate authenticates a user against the LDAP server.
func (s *AuthService) Authenticate(ctx context.Context, username, password string) (*domain.User, error) {
	if s.service == nil || !s.service.isEnabled() {
		return nil, errors.New("LDAP service is not configured or enabled")
	}

	// Remove domain from username if present
	username = strings.Split(username, "@")[0]

	authenticated, err := s.service.Authenticate(ctx, username, password)
	if err != nil {
		return nil, fmt.Errorf("LDAP authentication failed: %w", err)
	}

	if !authenticated {
		return nil, domain.ErrInvalidPassword
	}

	// Get or create the user in the local database
	// user, err := s.service.syncUser(ctx, username)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to get user from LDAP: %w", err)
	//}

	user, err := s.service.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from database: %w", err)
	}

	return &user, nil
}

// CanHandle returns true if the username matches the LDAP username pattern.
func (s *AuthService) CanHandle(username string) bool {
	// If LDAP is not configured, we can't handle any authentication
	if s.service == nil || !s.service.isEnabled() {
		return false
	}

	// If domain is configured, check if the username has the domain suffix
	if s.domain != "" {
		// Check if username has domain suffix (e.g., user@domain.com)
		if strings.Contains(username, "@") {
			parts := strings.Split(username, "@")
			if len(parts) == 2 && parts[1] == s.domain {
				return true
			}

			return false
		}
	}

	// If no domain is configured, we can handle any username
	return true
}
