package domain

// SSOProviderType represents the type of SSO provider.
type SSOProviderType string

const (
	SSOProviderSAML SSOProviderType = "saml"
)

// SSOProviderConfig holds configuration for SSO providers.
type SSOProviderConfig struct {
	Type        SSOProviderType `json:"type"`
	Enabled     bool            `json:"enabled"`
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	IconURL     string          `json:"icon_url"`
	SAMLConfig  *SAMLConfig     `json:"saml_config,omitempty"`
}

type SAMLConfig struct {
	Enabled          bool              `json:"enabled"`
	CreateCerts      bool              `json:"create_certs"`
	EntityID         string            `json:"entity_id"`
	CertificatePath  string            `json:"certificate_path"`
	PrivateKeyPath   string            `json:"private_key_path"`
	IDPMetadataURL   string            `json:"idp_metadata_url"`
	AttributeMapping map[string]string `json:"attribute_mapping"`
	CallbackURL      string            `json:"callback_url"`
	PublicRootURL    string            `json:"public_root_url"`
	SkipTLSVerify    bool              `yaml:"skip_tls_verify"`
}

const (
	SSOProviderNameADSaml = "ad_saml"
)
