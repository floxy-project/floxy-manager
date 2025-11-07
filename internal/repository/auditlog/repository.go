package auditlog

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/pkg/db"
)

func WriteLog(ctx context.Context, executor db.Tx, entity, entityID, action string, projectID domain.ProjectID) error {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		tx = executor
	}

	username := appcontext.Username(ctx)
	if username == "" {
		return fmt.Errorf("username not found in context")
	}

	// Try to get project_id from context if not provided
	if projectID == 0 {
		projectID = appcontext.ProjectID(ctx)
	}

	// For project entity, project_id is the entity_id
	if projectID == 0 && entity == domain.EntityProject {
		if id, err := strconv.Atoi(entityID); err == nil {
			projectID = domain.ProjectID(id)
		}
	}

	// For membership entity, try to get project_id from memberships table
	if projectID == 0 && entity == domain.EntityMembership {
		if membershipID, err := strconv.Atoi(entityID); err == nil {
			var pid int
			query := `SELECT project_id FROM workflows_manager.memberships WHERE id = $1 LIMIT 1`
			if err := tx.QueryRow(ctx, query, membershipID).Scan(&pid); err == nil {
				projectID = domain.ProjectID(pid)
			}
		}
	}

	// For workflow entity, try to get project_id from project_workflows table
	if projectID == 0 && entity == domain.EntityWorkflow {
		query := `SELECT project_id FROM workflows_manager.project_workflows WHERE workflow_definition_id = $1 LIMIT 1`
		var pid int
		if err := tx.QueryRow(ctx, query, entityID).Scan(&pid); err == nil {
			projectID = domain.ProjectID(pid)
		}
	}

	const query = `
INSERT INTO workflows_manager.audit_log (entity, entity_id, username, action, project_id, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())`

	var projectIDVal *int
	if projectID > 0 {
		val := projectID.Int()
		projectIDVal = &val
	}

	_, err := tx.Exec(ctx, query, entity, entityID, username, action, projectIDVal)
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

	countQuery := `SELECT COUNT(*) FROM workflows_manager.audit_log WHERE project_id = $1`
	var total int
	err := executor.QueryRow(ctx, countQuery, projectID.Int()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count audit log: %w", err)
	}

	query := `
SELECT id, entity, entity_id, username, action, created_at
FROM workflows_manager.audit_log
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

	rows, err := executor.Query(ctx, query, projectID.Int(), pageSize, offset)
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
