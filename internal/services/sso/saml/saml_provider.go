package samlprovider

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

const (
	metadataPath = "/api/v1/saml/metadata"
	acsPath      = "/api/v1/saml/acs"
)

// SAMLProvider implements SSOProvider for SAML.
type SAMLProvider struct {
	name        string
	displayName string
	iconURL     string
	config      *domain.SAMLConfig
	usersRepo   contract.UsersRepository
	httpClient  *http.Client
	certificate *x509.Certificate
	privateKey  crypto.Signer

	sp         *saml.ServiceProvider
	requestIDs sync.Map
}

type SAMLParams struct {
	Name        string
	DisplayName string
	IconURL     string
	Config      *domain.SAMLConfig
}

// New creates a new SAML provider.
func New(
	params *SAMLParams,
	manager contract.SSOProviderManager,
	usersRepo contract.UsersRepository,
) (*SAMLProvider, error) {
	provider := &SAMLProvider{
		name:        params.Name,
		displayName: params.DisplayName,
		iconURL:     params.IconURL,
		config:      params.Config,
		usersRepo:   usersRepo,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}

	if !params.Config.Enabled {
		return provider, nil
	}

	if params.Config.SkipTLSVerify {
		provider.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // by demand
			},
		}
	}

	if params.Config.CreateCerts {
		if _, err := os.Stat(params.Config.CertificatePath); err != nil {
			err := provider.generateSAMLKeys(params.Config.CertificatePath, params.Config.PrivateKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to generate SAML keys: %w", err)
			}
		}
	}

	// Load certificate if provided
	if params.Config.CertificatePath != "" {
		cert, err := provider.loadCertificate(params.Config.CertificatePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate: %w", err)
		}

		provider.certificate = cert
	}

	// Load private key if provided
	if params.Config.PrivateKeyPath != "" {
		key, err := provider.loadPrivateKey(params.Config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}

		provider.privateKey = key
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error

	provider.sp, err = provider.makeSP(ctx)
	if err != nil {
		slog.Error("failed to create SAML service provider, SAML disabled", "error", err)
		provider.config.Enabled = false

		return provider, nil
	}

	manager.AddProvider(params.Name, provider, domain.SSOProviderConfig{
		Type:        domain.SSOProviderSAML,
		Enabled:     params.Config.Enabled,
		Name:        params.Name,
		DisplayName: params.DisplayName,
		IconURL:     params.IconURL,
		SAMLConfig:  params.Config,
	})

	return provider, nil
}

// GetType returns the type of SSO provider.
func (p *SAMLProvider) GetType() string {
	return string(domain.SSOProviderSAML)
}

// GetName returns the name of the provider.
func (p *SAMLProvider) GetName() string {
	return p.name
}

// GetDisplayName returns the display name for UI.
func (p *SAMLProvider) GetDisplayName() string {
	return p.displayName
}

// GetIconURL returns the icon URL for UI.
func (p *SAMLProvider) GetIconURL() string {
	return p.iconURL
}

// IsEnabled returns true if the provider is enabled and the SSO feature is available in the current license.
func (p *SAMLProvider) IsEnabled() bool {
	if p.config == nil || !p.config.Enabled {
		return false
	}

	return true
}

func (p *SAMLProvider) GenerateSPMetadata() ([]byte, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("SAML provider '%s' is not enabled", p.name)
	}

	metadata := p.sp.Metadata()

	var buf bytes.Buffer

	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")

	if err := enc.Encode(metadata); err != nil {
		return nil, err
	}

	if err := enc.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateAuthURL generates the authorization URL with SAML AuthnRequest.
func (p *SAMLProvider) GenerateAuthURL(state string) (string, error) {
	if !p.IsEnabled() {
		return "", fmt.Errorf("SAML provider '%s' is not enabled", p.name)
	}

	authReq, err := p.sp.MakeAuthenticationRequest(
		p.sp.GetSSOBindingLocation(saml.HTTPRedirectBinding),
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return "", fmt.Errorf("create authentication request: %w", err)
	}

	p.requestIDs.Store(state, authReq.ID)

	redirectURL, err := authReq.Redirect(state, p.sp)
	if err != nil {
		return "", fmt.Errorf("create redirect to IDP: %w", err)
	}

	return redirectURL.String(), nil
}

func (p *SAMLProvider) Authenticate(
	ctx context.Context,
	req *http.Request,
	_, state string,
) (*domain.User, error) {
	if !p.IsEnabled() {
		return nil, fmt.Errorf("SAML provider '%s' is not enabled", p.name)
	}

	id, ok := p.requestIDs.LoadAndDelete(state)
	if !ok {
		return nil, fmt.Errorf("invalid state: %s", state)
	}

	idStr, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf("invalid id type: %T", id)
	}
	assertion, err := p.sp.ParseResponse(req, []string{idStr})
	if err != nil {
		return nil, fmt.Errorf("invalid SAML response: %w", err)
	}

	username, email := p.extractUserInfoFromAssertion(assertion)

	user, err := p.findOrCreateUser(ctx, username, email)
	if err != nil {
		return nil, fmt.Errorf("find or create user: %w", err)
	}

	return user, nil
}

// extractUserInfoFromAssertion extracts user information from SAML assertion.
func (p *SAMLProvider) extractUserInfoFromAssertion(assertion *saml.Assertion) (username, email string) {
	collected := p.collectByMapping(assertion)

	username = collected["username"]
	email = collected["email"]

	return username, email
}

func (p *SAMLProvider) collectByMapping(assertion *saml.Assertion) map[string]string {
	collected := make(map[string]string, len(p.config.AttributeMapping))

	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			name, ok := p.config.AttributeMapping[attr.Name]
			if !ok {
				continue
			}

			if len(attr.Values) > 0 {
				collected[name] = attr.Values[0].Value
			}
		}
	}

	return collected
}

// findOrCreateUser finds an existing user or creates a new one.
//
//nolint:nestif // todo: refactor
func (p *SAMLProvider) findOrCreateUser(ctx context.Context, username, email string) (*domain.User, error) {
	// Try to find the user by username first
	user, err := p.usersRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			// Try to find by email
			if email != "" {
				user, err = p.usersRepo.GetByEmail(ctx, email)
				if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
					return nil, fmt.Errorf("failed to check user by email: %w", err)
				}
			}
		} else {
			return nil, fmt.Errorf("failed to get user by username: %w", err)
		}
	}

	// If user not found, create a new user
	if errors.Is(err, domain.ErrEntityNotFound) {
		userDTO := domain.UserDTO{
			Username:      username,
			Email:         email,
			IsSuperuser:   false,
			PasswordHash:  "", // No password for SAML users
			IsTmpPassword: false,
			IsExternal:    true, // Mark as an external user
		}

		user, err = p.usersRepo.Create(ctx, userDTO)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Check if the user is active
	if !user.IsActive {
		return nil, domain.ErrInactiveUser
	}

	return &user, nil
}

// loadCertificate loads a certificate from a file.
//
//nolint:gosec // it's ok
func (p *SAMLProvider) loadCertificate(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// loadPrivateKey loads a private key from a file.
//
//nolint:gosec // it's ok
func (p *SAMLProvider) loadPrivateKey(path string) (crypto.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return key, nil
	}
	// Try PKCS8
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := parsedKey.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}

	return nil, errors.New("unsupported private key format")
}

func (p *SAMLProvider) makeSP(ctx context.Context) (*saml.ServiceProvider, error) {
	rootURL, err := url.Parse(p.config.PublicRootURL)
	if err != nil {
		return nil, fmt.Errorf("invalid public root URL: %w", err)
	}

	serviceProvider := &saml.ServiceProvider{
		EntityID:              path.Join(p.config.PublicRootURL, metadataPath),
		MetadataURL:           *rootURL.ResolveReference(&url.URL{Path: metadataPath}),
		AcsURL:                *rootURL.ResolveReference(&url.URL{Path: acsPath}),
		MetadataValidDuration: 24 * time.Hour,
	}

	switch {
	case p.privateKey != nil && p.certificate != nil:
		if rsaKey, ok := p.privateKey.(*rsa.PrivateKey); ok {
			serviceProvider.Key = rsaKey
		} else {
			return nil, errors.New("private key is not *rsa.PrivateKey")
		}

		serviceProvider.Certificate = p.certificate
	default:
		// fallback – generate self-signed certificates
		slog.Warn("generate self-signed certificates for SAML provider")

		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}

		serviceProvider.Key = key

		now := time.Now()
		template := x509.Certificate{
			SerialNumber: big.NewInt(now.UnixNano()),
			NotBefore:    now.Add(-time.Hour),
			NotAfter:     now.Add(365 * 24 * time.Hour),

			Subject: pkix.Name{
				CommonName:   "floxy-manager",
				Organization: []string{"FloxyManager"},
				Country:      []string{"US"},
			},
			Issuer: pkix.Name{
				CommonName:   "floxy-manager",
				Organization: []string{"FloxyManager"},
				Country:      []string{"US"},
			},

			KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
		if err != nil {
			return nil, err
		}

		serviceProvider.Certificate, err = x509.ParseCertificate(der)
		if err != nil {
			return nil, err
		}

		err = dumpOrStoreKeyPair(serviceProvider.Certificate, key, p.config.CertificatePath, p.config.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to store key pair: %w", err)
		}
	}

	idpURL, err := url.Parse(p.config.IDPMetadataURL)
	if err != nil {
		return nil, fmt.Errorf("invalid IDP metadata URL: %w", err)
	}

	metadata, err := samlsp.FetchMetadata(
		ctx,
		p.httpClient,
		*idpURL,
	)
	if err != nil {
		return nil, fmt.Errorf("fetch IDP metadata: %w", err)
	}

	serviceProvider.IDPMetadata = metadata

	return serviceProvider, nil
}

func (p *SAMLProvider) generateSAMLKeys(certPath, keyPath string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("rsa key generation failed: %w", err)
	}

	now := time.Now()
	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(now.UnixNano()),
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(365 * 24 * time.Hour),

		Subject: pkix.Name{
			CommonName:   "floxy-manager",
			Organization: []string{"FloxyManager"},
			Country:      []string{"RU"},
		},
		Issuer: pkix.Name{
			CommonName:   "floxy-manager",
			Organization: []string{"FloxyManager"},
			Country:      []string{"RU"},
		},

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("certificate creation failed: %w", err)
	}

	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open key file: %w", err)
	}
	err = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err != nil {
		_ = keyFile.Close()

		return fmt.Errorf("write key: %w", err)
	}
	_ = keyFile.Close()

	certFile, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("open cert file: %w", err)
	}
	if err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		_ = certFile.Close()

		return fmt.Errorf("write cert: %w", err)
	}
	_ = certFile.Close()

	return nil
}

func dumpOrStoreKeyPair(cert *x509.Certificate, key *rsa.PrivateKey, certFile, keyFile string) error {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(certFile), 0o750); err != nil {
			return err
		}

		if err := os.WriteFile(certFile, certPEM, 0o600); err != nil {
			return err
		}
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(keyFile), 0o750); err != nil {
			return err
		}

		if err := os.WriteFile(keyFile, keyPEM, 0o600); err != nil {
			return err
		}
	}

	return nil
}
