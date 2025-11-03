package projects

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *Repository) GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT id, name, description, created_at, updated_at, archived_at FROM workflows_manager.projects WHERE id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, id.Int())
	if err != nil {
		return domain.Project{}, fmt.Errorf("query project by ID: %w", err)
	}
	defer rows.Close()

	project, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[projectModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Project{}, domain.ErrEntityNotFound
		}

		return domain.Project{}, fmt.Errorf("collect project: %w", err)
	}

	return project.toDomain(), nil
}

func (r *Repository) Create(ctx context.Context, project *domain.ProjectDTO, tenantID domain.TenantID) (domain.ProjectID, error) {
	executor := r.getExecutor(ctx)

	const query = `
INSERT INTO  workflows_manager.projects (name, description, tenant_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $4)
RETURNING id`

	var id int

	err := executor.QueryRow(ctx, query,
		project.Name,
		project.Description,
		tenantID.Int(),
		time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert project: %w", err)
	}

	return domain.ProjectID(id), nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Project, error) {
	executor := r.getExecutor(ctx)

	const query = `
SELECT id, name, description, created_at, updated_at, archived_at 
FROM workflows_manager.projects p
WHERE p.archived_at IS NULL
ORDER BY p.id
`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[projectModel])
	if err != nil {
		return nil, fmt.Errorf("collect projects: %w", err)
	}

	projects := make([]domain.Project, 0, len(listModels))

	for i := range listModels {
		model := listModels[i]
		projects = append(projects, model.toDomain())
	}

	return projects, nil
}

func (r *Repository) ListByTenant(ctx context.Context, tenantID domain.TenantID) ([]domain.Project, error) {
	executor := r.getExecutor(ctx)

	const query = `
SELECT id, name, description, created_at, updated_at, archived_at 
FROM workflows_manager.projects p
WHERE p.tenant_id = $1 AND p.archived_at IS NULL
ORDER BY p.id
`

	rows, err := executor.Query(ctx, query, tenantID.Int())
	if err != nil {
		return nil, fmt.Errorf("query projects by tenant: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[projectModel])
	if err != nil {
		return nil, fmt.Errorf("collect projects: %w", err)
	}

	projects := make([]domain.Project, 0, len(listModels))

	for i := range listModels {
		model := listModels[i]
		projects = append(projects, model.toDomain())
	}

	return projects, nil
}

func (r *Repository) Update(ctx context.Context, id domain.ProjectID, name, description string) error {
	executor := r.getExecutor(ctx)

	const query = `
UPDATE  workflows_manager.projects
	SET name = $1, description = $2, updated_at = NOW()
WHERE id = $3`

	_, err := executor.Exec(ctx, query, name, description, id.Int())
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

func (r *Repository) Archive(ctx context.Context, id domain.ProjectID) error {
	executor := r.getExecutor(ctx)

	const query = `
UPDATE  workflows_manager.projects
	SET archived_at = NOW()
WHERE id = $1 AND archived_at IS NULL`

	result, err := executor.Exec(ctx, query, id.Int())
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		// Check if the project exists
		exists, err := r.projectExists(ctx, id)
		if err != nil {
			return fmt.Errorf("check if project exists: %w", err)
		}

		if !exists {
			return domain.ErrEntityNotFound
		}
		// Project exists but was already archived
		return nil
	}

	return nil
}

func (r *Repository) projectExists(ctx context.Context, id domain.ProjectID) (bool, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT 1 FROM  workflows_manager.projects WHERE id = $1 LIMIT 1`

	rows, err := executor.Query(ctx, query, id.Int())
	if err != nil {
		return false, fmt.Errorf("query project existence: %w", err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (r *Repository) GetProjectIDs(ctx context.Context) ([]domain.ProjectID, error) {
	executor := r.getExecutor(ctx)

	const query = `SELECT id FROM  workflows_manager.projects WHERE archived_at IS NULL`

	rows, err := executor.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query project IDs: %w", err)
	}
	defer rows.Close()

	var projectIDs []domain.ProjectID

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan project ID: %w", err)
		}

		projectIDs = append(projectIDs, domain.ProjectID(id))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project IDs: %w", err)
	}

	return projectIDs, nil
}

func (r *Repository) Count(ctx context.Context) (uint, error) {
	executor := r.getExecutor(ctx)

	const query = "SELECT COUNT(*) FROM  workflows_manager.projects"

	var count64 int64

	err := executor.QueryRow(ctx, query).Scan(&count64)
	if err != nil {
		return 0, fmt.Errorf("query projects count: %w", err)
	}

	return uint(count64), nil
}

func (r *Repository) Delete(ctx context.Context, id domain.ProjectID) error {
	executor := r.getExecutor(ctx)

	const query = `DELETE FROM workflows_manager.projects WHERE id = $1`

	result, err := executor.Exec(ctx, query, id.Int())
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrEntityNotFound
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
