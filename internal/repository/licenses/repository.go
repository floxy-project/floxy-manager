package licenses

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rom8726/floxy-manager/internal/domain"
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

func (r *Repository) Create(ctx context.Context, license domain.License) (domain.License, error) {
	executor := r.getExecutor(ctx)

	model := fromDomain(license)

	const query = `
INSERT INTO  workflows_manager.license (id, license_text, issued_at, expires_at, client_id, type, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *`

	rows, err := executor.Query(ctx, query,
		model.ID,
		model.LicenseText,
		model.IssuedAt,
		model.ExpiresAt,
		model.ClientID,
		model.Type,
		model.CreatedAt,
	)
	if err != nil {
		return domain.License{}, fmt.Errorf("insert license: %w", err)
	}
	defer rows.Close()

	licenseResult, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[licenseModel])
	if err != nil {
		return domain.License{}, fmt.Errorf("collect license: %w", err)
	}

	return licenseResult.toDomain(), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.License, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM  workflows_manager.license WHERE id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, id)
	if err != nil {
		return domain.License{}, fmt.Errorf("query license by ID: %w", err)
	}
	defer rows.Close()

	license, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[licenseModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.License{}, domain.ErrEntityNotFound
		}

		return domain.License{}, fmt.Errorf("collect license: %w", err)
	}

	return license.toDomain(), nil
}

func (r *Repository) GetLastByExpiresAt(ctx context.Context) (domain.License, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM  workflows_manager.license ORDER BY expires_at DESC LIMIT 1`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return domain.License{}, fmt.Errorf("query license by ID: %w", err)
	}
	defer rows.Close()

	lic, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[licenseModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.License{}, domain.ErrEntityNotFound
		}

		return domain.License{}, fmt.Errorf("collect license: %w", err)
	}

	return lic.toDomain(), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	executor := r.getExecutor(ctx)

	const query = `DELETE FROM  workflows_manager.license WHERE id = $1`

	tag, err := executor.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete license: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
	}

	return nil
}

func (r *Repository) List(ctx context.Context) ([]domain.License, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT * FROM  workflows_manager.license ORDER BY created_at DESC`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query licenses: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[licenseModel])
	if err != nil {
		return nil, fmt.Errorf("collect licenses: %w", err)
	}

	licenses := make([]domain.License, 0, len(listModels))

	for i := range listModels {
		model := listModels[i]
		licenses = append(licenses, model.toDomain())
	}

	return licenses, nil
}

func (r *Repository) UpdateLicense(ctx context.Context, license domain.License) (domain.License, error) {
	executor := r.getExecutor(ctx)

	// Check if a license with this ID already exists
	_, err := r.GetByID(ctx, license.ID)
	if err != nil && !errors.Is(err, domain.ErrEntityNotFound) {
		return domain.License{}, fmt.Errorf("check existing license: %w", err)
	}

	if errors.Is(err, domain.ErrEntityNotFound) {
		// Create a new license
		return r.Create(ctx, license)
	}

	// Update existing license
	model := fromDomain(license)

	const query = `
UPDATE  workflows_manager.license 
SET license_text = $1, issued_at = $2, expires_at = $3, client_id = $4, type = $5
WHERE id = $6
RETURNING *`

	rows, err := executor.Query(ctx, query,
		model.LicenseText,
		model.IssuedAt,
		model.ExpiresAt,
		model.ClientID,
		model.Type,
		model.ID,
	)
	if err != nil {
		return domain.License{}, fmt.Errorf("update license: %w", err)
	}
	defer rows.Close()

	licenseResult, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[licenseModel])
	if err != nil {
		return domain.License{}, fmt.Errorf("collect updated license: %w", err)
	}

	return licenseResult.toDomain(), nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
