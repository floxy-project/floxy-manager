package settings

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/crypt"
)

// Service provides settings management functionality.
type Service struct {
	settingsRepo contract.SettingRepository
	secret       []byte
}

// New creates a new settings use case.
func New(settingsRepo contract.SettingRepository, secret string) *Service {
	return &Service{
		settingsRepo: settingsRepo,
		secret:       []byte(secret),
	}
}

// GetLDAPConfig retrieves LDAP configuration from settings.
func (s *Service) GetLDAPConfig(ctx context.Context) (*domain.LDAPConfig, error) {
	setting, err := s.settingsRepo.GetByName(ctx, "ldap_config")
	if err != nil {
		return nil, fmt.Errorf("get LDAP config: %w", err)
	}

	var config domain.LDAPConfig

	err = json.Unmarshal(setting.Value, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal LDAP config: %w", err)
	}

	if config.BindPassword != "" {
		bindPasswordBytes, err := base64.StdEncoding.DecodeString(config.BindPassword)
		if err != nil {
			return nil, fmt.Errorf("decode LDAP bind password: %w", err)
		}

		bindPasswordBytes, err = crypt.DecryptAESGCM(bindPasswordBytes, s.secret)
		if err != nil {
			return nil, fmt.Errorf("decrypt LDAP bind password: %w", err)
		}

		config.BindPassword = string(bindPasswordBytes)
	}

	return &config, nil
}

// UpdateLDAPConfig updates LDAP configuration in settings.
func (s *Service) UpdateLDAPConfig(ctx context.Context, config *domain.LDAPConfig) error {
	bindPasswordEncrypted, err := crypt.EncryptAESGCM([]byte(config.BindPassword), s.secret)
	if err != nil {
		return fmt.Errorf("encrypt LDAP bind password: %w", err)
	}

	config.BindPassword = base64.StdEncoding.EncodeToString(bindPasswordEncrypted)

	err = s.settingsRepo.SetByName(
		ctx,
		"ldap_config",
		config,
		"LDAP server configuration for user and group synchronization",
	)
	if err != nil {
		return fmt.Errorf("failed to update LDAP config: %w", err)
	}

	return nil
}

// GetSetting retrieves a setting by name.
func (s *Service) GetSetting(ctx context.Context, name string) (*domain.Setting, error) {
	return s.settingsRepo.GetByName(ctx, name)
}

// SetSetting creates or updates a setting.
func (s *Service) SetSetting(ctx context.Context, name string, value interface{}, description string) error {
	return s.settingsRepo.SetByName(ctx, name, value, description)
}

// DeleteSetting deletes a setting by name.
func (s *Service) DeleteSetting(ctx context.Context, name string) error {
	return s.settingsRepo.DeleteByName(ctx, name)
}

// ListSettings retrieves all settings.
func (s *Service) ListSettings(ctx context.Context) ([]*domain.Setting, error) {
	return s.settingsRepo.List(ctx)
}
