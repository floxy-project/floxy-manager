package domain

import (
	"time"
)

type TenantID int

type Tenant struct {
	ID        TenantID
	Name      string
	CreatedAt time.Time
}

func (id TenantID) Int() int {
	return int(id)
}
