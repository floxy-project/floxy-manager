package tenants

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rom8726/floxy-manager/internal/domain"
	"github.com/rom8726/floxy-manager/internal/repository/auditlog"
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

func (r *Repository) GetByID(ctx context.Context, id domain.TenantID) (domain.Tenant, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM workflows_manager.tenants WHERE id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, id.Int())
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("query tenant by ID: %w", err)
	}
	defer rows.Close()

	tenant, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tenantModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Tenant{}, domain.ErrEntityNotFound
		}

		return domain.Tenant{}, fmt.Errorf("collect tenant: %w", err)
	}

	return tenant.toDomain(), nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Tenant, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM workflows_manager.tenants ORDER BY id`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tenants: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[tenantModel])
	if err != nil {
		return nil, fmt.Errorf("collect tenants: %w", err)
	}

	tenants := make([]domain.Tenant, 0, len(listModels))

	for i := range listModels {
		model := listModels[i]
		tenants = append(tenants, model.toDomain())
	}

	return tenants, nil
}

func (r *Repository) Create(ctx context.Context, name string) (domain.Tenant, error) {
	executor := r.getExecutor(ctx)

	const query = `INSERT INTO workflows_manager.tenants (name) VALUES ($1) RETURNING *`

	rows, err := executor.Query(ctx, query, name)
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("create tenant: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tenantModel])
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("collect tenant: %w", err)
	}

	if err := auditlog.WriteLog(ctx, executor, domain.EntityTenant, strconv.Itoa(model.ID), domain.ActionCreate); err != nil {
		return domain.Tenant{}, fmt.Errorf("write audit log: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) Update(ctx context.Context, id domain.TenantID, name string) (domain.Tenant, error) {
	executor := r.getExecutor(ctx)

	const query = `UPDATE workflows_manager.tenants SET name = $1 WHERE id = $2 RETURNING *`

	rows, err := executor.Query(ctx, query, name, id.Int())
	if err != nil {
		return domain.Tenant{}, fmt.Errorf("update tenant: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[tenantModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Tenant{}, domain.ErrEntityNotFound
		}
		return domain.Tenant{}, fmt.Errorf("collect tenant: %w", err)
	}

	if err := auditlog.WriteLog(ctx, executor, domain.EntityTenant, strconv.Itoa(id.Int()), domain.ActionUpdate); err != nil {
		return domain.Tenant{}, fmt.Errorf("write audit log: %w", err)
	}

	return model.toDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id domain.TenantID) error {
	executor := r.getExecutor(ctx)

	const query = `DELETE FROM workflows_manager.tenants WHERE id = $1`

	result, err := executor.Exec(ctx, query, id.Int())
	if err != nil {
		return fmt.Errorf("delete tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}

	if err := auditlog.WriteLog(ctx, executor, domain.EntityTenant, strconv.Itoa(id.Int()), domain.ActionDelete); err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}

	return nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
