package auditlog

import (
	"context"
	"fmt"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
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
