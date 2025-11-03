package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type ProjectsUseCase interface {
	CreateProject(ctx context.Context, name, description string, tenantID domain.TenantID) (domain.Project, error)
	GetProject(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	List(ctx context.Context) ([]domain.Project, error)
	UpdateInfo(ctx context.Context, id domain.ProjectID, name, description string) (domain.Project, error)
	ArchiveProject(ctx context.Context, id domain.ProjectID) error
}

type ProjectsRepository interface {
	GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error)
	Create(ctx context.Context, project *domain.ProjectDTO, tenantID domain.TenantID) (domain.ProjectID, error)
	List(ctx context.Context) ([]domain.Project, error)
	ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]domain.Project, error)
	Update(ctx context.Context, id domain.ProjectID, name, description string) error
	Archive(ctx context.Context, id domain.ProjectID) error
	Delete(ctx context.Context, id domain.ProjectID) error
}
