package domain

import "time"

type PermissionID string

type Permission struct {
	ID   PermissionID
	Key  PermKey
	Name string
}

type MembershipID int

type ProjectMembership struct {
	ID        MembershipID
	UserID    UserID
	ProjectID ProjectID
	RoleID    RoleID
	RoleKey   string
	RoleName  string
	CreatedAt time.Time
}
