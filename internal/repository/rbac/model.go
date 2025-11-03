package rbac

import (
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type roleModel struct {
	ID          string    `db:"id"`
	Key         string    `db:"key"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
}

func (role *roleModel) toDomain() domain.Role {
	return domain.Role{
		ID:          domain.RoleID(role.ID),
		Key:         role.Key,
		Name:        role.Name,
		Description: role.Description,
		CreatedAt:   role.CreatedAt,
	}
}

type permissionModel struct {
	ID   string `db:"id"`
	Key  string `db:"key"`
	Name string `db:"name"`
}

func (p *permissionModel) toDomain() domain.Permission {
	return domain.Permission{ID: domain.PermissionID(p.ID), Key: domain.PermKey(p.Key), Name: p.Name}
}

type membershipModel struct {
	ID        string    `db:"id"`
	ProjectID int       `db:"project_id"`
	UserID    int       `db:"user_id"`
	RoleID    string    `db:"role_id"`
	RoleKey   string    `db:"role_key"`
	RoleName  string    `db:"role_name"`
	CreatedAt time.Time `db:"created_at"`
}

func (m *membershipModel) toDomain() domain.ProjectMembership {
	return domain.ProjectMembership{
		ID:        domain.MembershipID(m.ID),
		UserID:    domain.UserID(m.UserID),
		ProjectID: domain.ProjectID(m.ProjectID),
		RoleID:    domain.RoleID(m.RoleID),
		RoleKey:   m.RoleKey,
		RoleName:  m.RoleName,
		CreatedAt: m.CreatedAt,
	}
}
