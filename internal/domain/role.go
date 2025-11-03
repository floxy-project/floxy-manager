package domain

import (
	"time"
)

type RoleID string

type Role struct {
	ID          RoleID
	Key         string
	Name        string
	Description string
	CreatedAt   time.Time
}
