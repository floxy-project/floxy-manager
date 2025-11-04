package contract

import (
	"context"
	"encoding/json"

	"github.com/rom8726/floxy-manager/internal/domain"
)

type WorkflowsRepository interface {
	ListWorkflowDefinitions(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		page, pageSize int,
	) ([]domain.WorkflowDefinition, int, error)
	GetWorkflowDefinition(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		id string,
	) (domain.WorkflowDefinition, error)
	ListWorkflowInstances(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		workflowID string,
		page, pageSize int,
	) ([]domain.WorkflowInstance, int, error)
	GetWorkflowInstance(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		id int,
	) (domain.WorkflowInstance, error)
	ListWorkflowSteps(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		instanceID int,
		page, pageSize int,
	) ([]domain.WorkflowStep, int, error)
	ListWorkflowEvents(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		instanceID int,
		page, pageSize int,
	) ([]domain.WorkflowEvent, int, error)
	ListActiveWorkflows(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		page, pageSize int,
	) ([]domain.ActiveWorkflow, int, error)
	ListWorkflowStats(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		page, pageSize int,
	) ([]domain.WorkflowStat, int, error)
	ListDLQItems(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		page, pageSize int,
	) ([]domain.DLQItem, int, error)
	GetDLQItem(
		ctx context.Context,
		tenantID domain.TenantID,
		projectID domain.ProjectID,
		id int,
	) (domain.DLQItem, error)
	ListUnassignedWorkflowDefinitions(
		ctx context.Context,
		page, pageSize int,
	) ([]domain.WorkflowDefinition, int, error)
	AssignWorkflowDefinitionsToProject(
		ctx context.Context,
		projectID domain.ProjectID,
		workflowIDs []string,
	) (int, error)
	CreateWorkflowDefinition(
		ctx context.Context,
		projectID domain.ProjectID,
		name string,
		version int,
		definition json.RawMessage,
	) (string, error)
}
