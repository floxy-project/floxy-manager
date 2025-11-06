package auditlog

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/db"
)

func WriteLog(ctx context.Context, executor db.Tx, entity, entityID, action string) error {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		tx = executor
	}

	username := appcontext.Username(ctx)
	if username == "" {
		return fmt.Errorf("username not found in context")
	}

	const query = `
INSERT INTO workflows_manager.audit_log (entity, entity_id, username, action, created_at)
VALUES ($1, $2, $3, $4, NOW())`

	_, err := tx.Exec(ctx, query, entity, entityID, username, action)
	if err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

type Repository struct {
	db db.Tx
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		db: pool,
	}
}

func (r *Repository) List(ctx context.Context, projectID domain.ProjectID, page, pageSize int) ([]contract.AuditLogEntry, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	countQuery := `SELECT COUNT(*) FROM workflows_manager.audit_log`
	var total int
	err := executor.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit log: %w", err)
	}

	query := `
SELECT id, entity, entity_id, username, action, created_at
FROM workflows_manager.audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`

	rows, err := executor.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit log: %w", err)
	}
	defer rows.Close()

	entries := make([]contract.AuditLogEntry, 0)
	for rows.Next() {
		var entry contract.AuditLogEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Entity,
			&entry.EntityID,
			&entry.Username,
			&entry.Action,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan audit log entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return entries, total, nil
}

func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db
}
