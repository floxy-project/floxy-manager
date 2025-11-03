package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

// MembershipsUseCase provides operations for roles, permissions and project memberships.
// It is used by the REST API layer.
type MembershipsUseCase interface {
	// Roles and permissions
	ListRoles(ctx context.Context) ([]domain.Role, error)
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	GetRolePermissions(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error)
	ListRolePermissions(ctx context.Context) (map[domain.Role][]domain.Permission, error)

	// Memberships
	ListProjectMemberships(ctx context.Context, projectID domain.ProjectID) ([]domain.ProjectMembership, error)
	CreateProjectMembership(
		ctx context.Context,
		projectID domain.ProjectID,
		userID domain.UserID,
		roleID domain.RoleID,
	) (domain.ProjectMembership, error)
	GetProjectMembership(
		ctx context.Context,
		projectID domain.ProjectID,
		membershipID domain.MembershipID,
	) (domain.ProjectMembership, error)
	UpdateProjectMembership(
		ctx context.Context,
		projectID domain.ProjectID,
		membershipID domain.MembershipID,
		roleID domain.RoleID,
	) (domain.ProjectMembership, error)
	DeleteProjectMembership(ctx context.Context, projectID domain.ProjectID, membershipID domain.MembershipID) error
}
