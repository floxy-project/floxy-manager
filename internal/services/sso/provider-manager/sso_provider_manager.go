package ssoprovidermanager

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

// SSOProviderManager manages multiple SSO providers.
type SSOProviderManager struct {
	providers map[string]contract.SSOProvider
	configs   map[string]domain.SSOProviderConfig
}

// New creates a new SSO provider manager.
func New() *SSOProviderManager {
	return &SSOProviderManager{
		providers: make(map[string]contract.SSOProvider),
		configs:   make(map[string]domain.SSOProviderConfig),
	}
}

// AddProvider adds a new SSO provider.
func (m *SSOProviderManager) AddProvider(name string, provider contract.SSOProvider, config domain.SSOProviderConfig) {
	m.providers[name] = provider
	m.configs[name] = config
}

// GetProvider returns a provider by name.
func (m *SSOProviderManager) GetProvider(name string) (contract.SSOProvider, bool) {
	provider, exists := m.providers[name]

	return provider, exists
}

// GetEnabledProviders returns all enabled providers.
func (m *SSOProviderManager) GetEnabledProviders() []contract.SSOProvider {
	var enabled []contract.SSOProvider

	for _, provider := range m.providers {
		if provider.IsEnabled() {
			enabled = append(enabled, provider)
		}
	}

	return enabled
}

// GetProviderConfig returns the configuration for a provider.
func (m *SSOProviderManager) GetProviderConfig(name string) (domain.SSOProviderConfig, bool) {
	providerConfig, exists := m.configs[name]

	return providerConfig, exists
}

func (m *SSOProviderManager) GetProviderMetadata(name string) ([]byte, error) {
	provider, exists := m.GetProvider(name)
	if !exists {
		return nil, errors.New("SSO provider not found")
	}

	return provider.GenerateSPMetadata()
}

// AuthenticateWithProvider authenticates using a specific provider.
func (m *SSOProviderManager) AuthenticateWithProvider(
	ctx context.Context,
	providerName string,
	req *http.Request,
	response string,
	state string,
) (*domain.User, error) {
	provider, exists := m.GetProvider(providerName)
	if !exists {
		return nil, fmt.Errorf("SSO provider '%s' not found", providerName)
	}

	if !provider.IsEnabled() {
		return nil, fmt.Errorf("SSO provider '%s' is not enabled", providerName)
	}

	return provider.Authenticate(ctx, req, response, state)
}
