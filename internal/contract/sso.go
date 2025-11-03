package contract

import (
	"context"
	"net/http"

	"github.com/rom8726/floxy-manager/internal/domain"
)

// SSOProvider represents an SSO provider interface.
type SSOProvider interface {
	GetType() string
	GetName() string
	GetDisplayName() string
	GetIconURL() string
	IsEnabled() bool
	GenerateAuthURL(state string) (string, error)
	GenerateSPMetadata() ([]byte, error)
	Authenticate(ctx context.Context, req *http.Request, response, state string) (*domain.User, error)
}

type SSOProviderManager interface {
	AddProvider(name string, provider SSOProvider, config domain.SSOProviderConfig)
	GetProvider(name string) (SSOProvider, bool)
	GetEnabledProviders() []SSOProvider
	GetProviderConfig(name string) (domain.SSOProviderConfig, bool)
	GetProviderMetadata(name string) ([]byte, error)
	AuthenticateWithProvider(
		ctx context.Context,
		providerName string,
		req *http.Request,
		response string,
		state string,
	) (*domain.User, error)
}
