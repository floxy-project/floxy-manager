package workflows

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

// ListWorkflowDefinitions returns workflow definitions filtered by tenant_id and project_id
func (r *Repository) ListWorkflowDefinitions(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	page, pageSize int,
) ([]domain.WorkflowDefinition, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_workflow_definitions 
WHERE tenant_id = $1 AND project_id = $2`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workflow definitions: %w", err)
	}

	// Fetch items
	const query = `
SELECT * FROM workflows_manager.v_workflow_definitions 
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query workflow definitions: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowDefinitionModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect workflow definitions: %w", err)
	}

	definitions := make([]domain.WorkflowDefinition, 0, len(listModels))
	for i := range listModels {
		definitions = append(definitions, listModels[i].toDomain())
	}

	return definitions, total, nil
}

// GetWorkflowDefinition returns a workflow definition by ID
func (r *Repository) GetWorkflowDefinition(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	id string,
) (domain.WorkflowDefinition, error) {
	executor := r.getExecutor(ctx)

	const query = `
SELECT * FROM workflows_manager.v_workflow_definitions 
WHERE tenant_id = $1 AND project_id = $2 AND id = $3
LIMIT 1`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), id)
	if err != nil {
		return domain.WorkflowDefinition{}, fmt.Errorf("query workflow definition: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[workflowDefinitionModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkflowDefinition{}, domain.ErrEntityNotFound
		}
		return domain.WorkflowDefinition{}, fmt.Errorf("collect workflow definition: %w", err)
	}

	return model.toDomain(), nil
}

// ListWorkflowInstances returns workflow instances filtered by tenant_id and project_id
func (r *Repository) ListWorkflowInstances(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	workflowID string,
	page, pageSize int,
) ([]domain.WorkflowInstance, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	var countQuery string
	var query string
	var countArgs []interface{}
	var args []interface{}

	if workflowID != "" {
		countQuery = `
SELECT COUNT(*) FROM workflows_manager.v_workflow_instances 
WHERE tenant_id = $1 AND project_id = $2 AND workflow_id = $3`
		countArgs = []interface{}{tenantID.Int(), projectID.Int(), workflowID}

		query = `
SELECT * FROM workflows_manager.v_workflow_instances 
WHERE tenant_id = $1 AND project_id = $2 AND workflow_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5`
		args = []interface{}{tenantID.Int(), projectID.Int(), workflowID, pageSize, offset}
	} else {
		countQuery = `
SELECT COUNT(*) FROM workflows_manager.v_workflow_instances 
WHERE tenant_id = $1 AND project_id = $2`
		countArgs = []interface{}{tenantID.Int(), projectID.Int()}

		query = `
SELECT * FROM workflows_manager.v_workflow_instances 
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`
		args = []interface{}{tenantID.Int(), projectID.Int(), pageSize, offset}
	}

	// Count total
	var total int
	err := executor.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workflow instances: %w", err)
	}

	// Fetch items
	rows, err := executor.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query workflow instances: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowInstanceModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect workflow instances: %w", err)
	}

	instances := make([]domain.WorkflowInstance, 0, len(listModels))
	for i := range listModels {
		instances = append(instances, listModels[i].toDomain())
	}

	return instances, total, nil
}

// GetWorkflowInstance returns a workflow instance by ID
func (r *Repository) GetWorkflowInstance(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	id int,
) (domain.WorkflowInstance, error) {
	executor := r.getExecutor(ctx)

	const query = `
SELECT * FROM workflows_manager.v_workflow_instances 
WHERE tenant_id = $1 AND project_id = $2 AND id = $3
LIMIT 1`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), id)
	if err != nil {
		return domain.WorkflowInstance{}, fmt.Errorf("query workflow instance: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[workflowInstanceModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.WorkflowInstance{}, domain.ErrEntityNotFound
		}
		return domain.WorkflowInstance{}, fmt.Errorf("collect workflow instance: %w", err)
	}

	return model.toDomain(), nil
}

// ListWorkflowSteps returns workflow steps for an instance
func (r *Repository) ListWorkflowSteps(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	instanceID int,
	page, pageSize int,
) ([]domain.WorkflowStep, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_workflow_steps 
WHERE tenant_id = $1 AND project_id = $2 AND instance_id = $3`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int(), instanceID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workflow steps: %w", err)
	}

	// Fetch items
	const query = `
SELECT * FROM workflows_manager.v_workflow_steps 
WHERE tenant_id = $1 AND project_id = $2 AND instance_id = $3
ORDER BY created_at ASC
LIMIT $4 OFFSET $5`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), instanceID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query workflow steps: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowStepModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect workflow steps: %w", err)
	}

	steps := make([]domain.WorkflowStep, 0, len(listModels))
	for i := range listModels {
		steps = append(steps, listModels[i].toDomain())
	}

	return steps, total, nil
}

// ListWorkflowEvents returns workflow events for an instance
func (r *Repository) ListWorkflowEvents(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	instanceID int,
	page, pageSize int,
) ([]domain.WorkflowEvent, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_workflow_events 
WHERE tenant_id = $1 AND project_id = $2 AND instance_id = $3`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int(), instanceID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workflow events: %w", err)
	}

	// Fetch items
	const query = `
SELECT * FROM workflows_manager.v_workflow_events 
WHERE tenant_id = $1 AND project_id = $2 AND instance_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), instanceID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query workflow events: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowEventModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect workflow events: %w", err)
	}

	events := make([]domain.WorkflowEvent, 0, len(listModels))
	for i := range listModels {
		events = append(events, listModels[i].toDomain())
	}

	return events, total, nil
}

// ListActiveWorkflows returns active workflows
func (r *Repository) ListActiveWorkflows(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	page, pageSize int,
) ([]domain.ActiveWorkflow, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_active_workflows 
WHERE tenant_id = $1 AND project_id = $2`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count active workflows: %w", err)
	}

	// Fetch items
	const query = `
SELECT 
    vaw.tenant_id,
    vaw.project_id,
    vaw.id,
    vaw.workflow_id,
    wd.name as workflow_name,
    vaw.status,
    vaw.created_at,
    vaw.updated_at,
    vaw.duration_seconds,
    vaw.total_steps,
    vaw.completed_steps,
    vaw.failed_steps,
    vaw.running_steps
FROM workflows_manager.v_active_workflows vaw
JOIN workflows.workflow_definitions wd ON wd.id = vaw.workflow_id
WHERE vaw.tenant_id = $1 AND vaw.project_id = $2
ORDER BY vaw.created_at DESC
LIMIT $3 OFFSET $4`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query active workflows: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[activeWorkflowModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect active workflows: %w", err)
	}

	workflows := make([]domain.ActiveWorkflow, 0, len(listModels))
	for i := range listModels {
		workflows = append(workflows, listModels[i].toDomain())
	}

	return workflows, total, nil
}

// ListWorkflowStats returns workflow statistics
func (r *Repository) ListWorkflowStats(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	page, pageSize int,
) ([]domain.WorkflowStat, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_workflow_stats 
WHERE tenant_id = $1 AND project_id = $2`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count workflow stats: %w", err)
	}

	// Fetch items
	const query = `
SELECT tenant_id, project_id, name, version, total_instances, completed, failed, running, avg_duration_seconds
FROM workflows_manager.v_workflow_stats 
WHERE tenant_id = $1 AND project_id = $2
ORDER BY name, version
LIMIT $3 OFFSET $4`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query workflow stats: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowStatModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect workflow stats: %w", err)
	}

	stats := make([]domain.WorkflowStat, 0, len(listModels))
	for i := range listModels {
		stats = append(stats, listModels[i].toDomain())
	}

	return stats, total, nil
}

// ListDLQItems returns DLQ items with pagination
func (r *Repository) ListDLQItems(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	page, pageSize int,
) ([]domain.DLQItem, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total
	countQuery := `
SELECT COUNT(*) FROM workflows_manager.v_workflow_dlq 
WHERE tenant_id = $1 AND project_id = $2`

	var total int
	err := executor.QueryRow(ctx, countQuery, tenantID.Int(), projectID.Int()).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count DLQ items: %w", err)
	}

	// Fetch items
	const query = `
SELECT * FROM workflows_manager.v_workflow_dlq 
WHERE tenant_id = $1 AND project_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query DLQ items: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[dlqItemModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect DLQ items: %w", err)
	}

	items := make([]domain.DLQItem, 0, len(listModels))
	for i := range listModels {
		items = append(items, listModels[i].toDomain())
	}

	return items, total, nil
}

// GetDLQItem returns a DLQ item by ID
func (r *Repository) GetDLQItem(
	ctx context.Context,
	tenantID domain.TenantID,
	projectID domain.ProjectID,
	id int,
) (domain.DLQItem, error) {
	executor := r.getExecutor(ctx)

	const query = `
SELECT * FROM workflows_manager.v_workflow_dlq 
WHERE tenant_id = $1 AND project_id = $2 AND id = $3
LIMIT 1`

	rows, err := executor.Query(ctx, query, tenantID.Int(), projectID.Int(), id)
	if err != nil {
		return domain.DLQItem{}, fmt.Errorf("query DLQ item: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[dlqItemModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.DLQItem{}, domain.ErrEntityNotFound
		}
		return domain.DLQItem{}, fmt.Errorf("collect DLQ item: %w", err)
	}

	return model.toDomain(), nil
}

// ListUnassignedWorkflowDefinitions returns workflow definitions that are not assigned to any project
func (r *Repository) ListUnassignedWorkflowDefinitions(
	ctx context.Context,
	page, pageSize int,
) ([]domain.WorkflowDefinition, int, error) {
	executor := r.getExecutor(ctx)

	offset := (page - 1) * pageSize

	// Count total - workflows that are not in project_workflows
	countQuery := `
SELECT COUNT(*) 
FROM workflows.workflow_definitions wd
WHERE NOT EXISTS (
    SELECT 1 
    FROM workflows_manager.project_workflows pw 
    WHERE pw.workflow_definition_id = wd.id
)`

	var total int
	err := executor.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count unassigned workflow definitions: %w", err)
	}

	// Fetch items
	const query = `
SELECT 
    0 as tenant_id,
    0 as project_id,
    wd.id,
    wd.name,
    wd.version,
    wd.definition::text as definition,
    wd.created_at
FROM workflows.workflow_definitions wd
WHERE NOT EXISTS (
    SELECT 1 
    FROM workflows_manager.project_workflows pw 
    WHERE pw.workflow_definition_id = wd.id
)
ORDER BY wd.created_at DESC
LIMIT $1 OFFSET $2`

	rows, err := executor.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query unassigned workflow definitions: %w", err)
	}
	defer rows.Close()

	listModels, err := pgx.CollectRows(rows, pgx.RowToStructByName[workflowDefinitionModel])
	if err != nil {
		return nil, 0, fmt.Errorf("collect unassigned workflow definitions: %w", err)
	}

	definitions := make([]domain.WorkflowDefinition, 0, len(listModels))
	for i := range listModels {
		definitions = append(definitions, listModels[i].toDomain())
	}

	return definitions, total, nil
}

// AssignWorkflowDefinitionsToProject assigns workflow definitions to a project
// If workflowIDs is empty, assigns all unassigned workflows
func (r *Repository) AssignWorkflowDefinitionsToProject(
	ctx context.Context,
	projectID domain.ProjectID,
	workflowIDs []string,
) (int, error) {
	executor := r.getExecutor(ctx)

	var query string
	var args []interface{}

	if len(workflowIDs) == 0 {
		// Assign all unassigned workflows
		query = `
INSERT INTO workflows_manager.project_workflows (project_id, workflow_definition_id)
SELECT $1, wd.id
FROM workflows.workflow_definitions wd
WHERE NOT EXISTS (
    SELECT 1 
    FROM workflows_manager.project_workflows pw 
    WHERE pw.workflow_definition_id = wd.id
)
ON CONFLICT (project_id, workflow_definition_id) DO NOTHING`
		args = []interface{}{projectID.Int()}
	} else {
		// Assign specific workflows - use VALUES clause for better performance
		query = `
INSERT INTO workflows_manager.project_workflows (project_id, workflow_definition_id)
SELECT $1, wd.id
FROM workflows.workflow_definitions wd
WHERE wd.id = ANY($2::text[])
ON CONFLICT (project_id, workflow_definition_id) DO NOTHING`
		args = []interface{}{projectID.Int(), workflowIDs}
	}

	result, err := executor.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("assign workflow definitions to project: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return int(rowsAffected), nil
}

// CreateWorkflowDefinition creates a new workflow definition in the database
// and automatically assigns it to the specified project
func (r *Repository) CreateWorkflowDefinition(
	ctx context.Context,
	projectID domain.ProjectID,
	name string,
	version int,
	definition json.RawMessage,
) (string, error) {
	executor := r.getExecutor(ctx)

	id := fmt.Sprintf("%s-v%d", name, version)

	// First, create or update the workflow definition
	createQuery := `
INSERT INTO workflows.workflow_definitions (id, name, version, definition)
VALUES ($1, $2, $3, $4::jsonb)
ON CONFLICT (name, version) DO UPDATE 
SET definition = EXCLUDED.definition
RETURNING id`

	var workflowID string
	err := executor.QueryRow(ctx, createQuery, id, name, version, definition).Scan(&workflowID)
	if err != nil {
		return "", fmt.Errorf("create workflow definition: %w", err)
	}

	// Then assign to project (this will be atomic within the same transaction)
	assignQuery := `
INSERT INTO workflows_manager.project_workflows (project_id, workflow_definition_id)
VALUES ($1, $2)
ON CONFLICT (project_id, workflow_definition_id) DO NOTHING`

	_, err = executor.Exec(ctx, assignQuery, projectID.Int(), workflowID)
	if err != nil {
		return "", fmt.Errorf("assign workflow to project: %w", err)
	}

	if err := auditlog.WriteLog(ctx, executor, domain.EntityWorkflow, workflowID, domain.ActionCreate); err != nil {
		return "", fmt.Errorf("write audit log: %w", err)
	}

	return workflowID, nil
}

//nolint:ireturn // it's ok here
func (r *Repository) getExecutor(ctx context.Context) db.Tx {
	if tx := db.TxFromContext(ctx); tx != nil {
		return tx
	}

	return r.db
}
