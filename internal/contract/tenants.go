package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type TenantsRepository interface {
	GetByID(ctx context.Context, id domain.TenantID) (domain.Tenant, error)
	List(ctx context.Context) ([]domain.Tenant, error)
	Create(ctx context.Context, name string) (domain.Tenant, error)
	Delete(ctx context.Context, id domain.TenantID) error
}
