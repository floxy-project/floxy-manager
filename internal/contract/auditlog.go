package contract

import (
	"context"
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type AuditLogEntry struct {
	ID        int64     `json:"id"`
	Entity    string    `json:"entity"`
	EntityID  string    `json:"entity_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLogRepository interface {
	List(ctx context.Context, projectID domain.ProjectID, page, pageSize int) ([]AuditLogEntry, int, error)
}
