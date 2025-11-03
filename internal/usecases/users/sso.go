package users

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

func (s *UsersService) GetSSOMetadata(_ context.Context, providerName string) ([]byte, error) {
	if s.ssoManager == nil {
		return nil, errors.New("SSO is not enabled")
	}

	provider, exists := s.ssoManager.GetProvider(providerName)
	if !exists {
		return nil, fmt.Errorf("SSO provider '%s' not found", providerName)
	}

	if !provider.IsEnabled() {
		return nil, fmt.Errorf("SSO provider '%s' is not enabled", providerName)
	}

	return s.ssoManager.GetProviderMetadata(providerName)
}

// SSOInitiate initiates the SSO login flow by generating a redirect URL to the specified provider.
func (s *UsersService) SSOInitiate(_ context.Context, providerName string) (redirectURL string, err error) {
	if s.ssoManager == nil {
		return "", errors.New("SSO is not enabled")
	}

	provider, exists := s.ssoManager.GetProvider(providerName)
	if !exists {
		return "", fmt.Errorf("SSO provider '%s' not found", providerName)
	}

	if !provider.IsEnabled() {
		return "", fmt.Errorf("SSO provider '%s' is not enabled", providerName)
	}

	// Generate state for CSRF protection
	state, err := s.generateState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	redirectURL, err = provider.GenerateAuthURL(state)
	if err != nil {
		return "", fmt.Errorf("failed to generate auth URL: %w", err)
	}

	return redirectURL, nil
}

// GetSSOProviders returns all enabled SSO providers.
func (s *UsersService) GetSSOProviders(context.Context) ([]contract.SSOProvider, error) {
	if s.ssoManager == nil {
		return nil, errors.New("SSO is not enabled")
	}

	return s.ssoManager.GetEnabledProviders(), nil
}

// SSOCallback handles the SSO callback from Keycloak.
func (s *UsersService) SSOCallback(
	ctx context.Context,
	providerName string,
	req *http.Request,
	response string,
	state string,
) (accessToken, refreshToken string, expiresIn int, err error) {
	if s.ssoManager == nil {
		return "", "", 0, errors.New("SSO is not enabled")
	}

	// Authenticate using the specified provider
	user, err := s.ssoManager.AuthenticateWithProvider(ctx, providerName, req, response, state)
	if err != nil {
		return "", "", 0, fmt.Errorf("SSO authentication failed: %w", err)
	}

	// Check if the user is active
	if !user.IsActive {
		return "", "", 0, domain.ErrInactiveUser
	}

	// Generate tokens
	accessToken, err = s.tokenizer.AccessToken(user)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.tokenizer.RefreshToken(user)
	if err != nil {
		return "", "", 0, fmt.Errorf("generate refresh token: %w", err)
	}

	// Update last login
	if err := s.usersRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", "", 0, fmt.Errorf("update last login: %w", err)
	}

	// Calculate expiration time
	expiresIn = int(s.tokenizer.AccessTokenTTL().Seconds())

	return accessToken, refreshToken, expiresIn, nil
}

// generateState generates a random state parameter for CSRF protection.
func (s *UsersService) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
