package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type SettingsUseCase interface {
	GetLDAPConfig(ctx context.Context) (*domain.LDAPConfig, error)
	UpdateLDAPConfig(ctx context.Context, config *domain.LDAPConfig) error
	GetSetting(ctx context.Context, name string) (*domain.Setting, error)
	SetSetting(ctx context.Context, name string, value any, description string) error
	DeleteSetting(ctx context.Context, name string) error
	ListSettings(ctx context.Context) ([]*domain.Setting, error)
}

// SettingRepository defines the interface for settings operations.
type SettingRepository interface {
	GetByName(ctx context.Context, name string) (*domain.Setting, error)
	SetByName(ctx context.Context, name string, value any, description string) error
	DeleteByName(ctx context.Context, name string) error
	List(ctx context.Context) ([]*domain.Setting, error)
}
