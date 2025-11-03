package tenants

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type tenantModel struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func (m *tenantModel) toDomain() domain.Tenant {
	return domain.Tenant{
		ID:        domain.TenantID(m.ID),
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
	}
}
