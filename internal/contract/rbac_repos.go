package contract

import (
	"context"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type RolesRepository interface {
	GetByKey(ctx context.Context, key string) (domain.Role, error)
	GetByID(ctx context.Context, id domain.RoleID) (domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
}

type PermissionsRepository interface {
	RoleHasPermission(ctx context.Context, roleID string, key domain.PermKey) (bool, error)
	List(ctx context.Context) ([]domain.Permission, error)
	ListForRole(ctx context.Context, roleID domain.RoleID) ([]domain.Permission, error)
	ListForAllRoles(ctx context.Context) (map[domain.Role][]domain.Permission, error)
}

type MembershipsRepository interface {
	GetForUserProject(ctx context.Context, userID domain.UserID, projectID domain.ProjectID) (roleID string, err error)
	ListForProject(ctx context.Context, projectID domain.ProjectID) ([]domain.ProjectMembership, error)
	Create(
		ctx context.Context,
		projectID domain.ProjectID,
		userID domain.UserID,
		roleID domain.RoleID,
	) (domain.ProjectMembership, error)
	Get(
		ctx context.Context,
		projectID domain.ProjectID,
		membershipID domain.MembershipID,
	) (domain.ProjectMembership, error)
	Update(
		ctx context.Context,
		projectID domain.ProjectID,
		membershipID domain.MembershipID,
		roleID domain.RoleID,
	) (domain.ProjectMembership, error)
	Delete(ctx context.Context, projectID domain.ProjectID, membershipID domain.MembershipID) error
}
