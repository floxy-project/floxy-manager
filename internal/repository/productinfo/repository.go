package productinfo

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rom8726/floxy-manager/pkg/db"
)

type Repository struct {
	db db.Tx
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{
		db: pool,
	}
}

func (r *Repository) GetClientID(ctx context.Context) (string, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT value FROM workflows_manager.product_info WHERE key = 'client_id' LIMIT 1`

	row := executor.QueryRow(ctx, query)

	var clientID string

	err := row.Scan(&clientID)
	if err != nil {
		return "", err
	}

	return clientID, nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
