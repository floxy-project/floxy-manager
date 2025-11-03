package workflows

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/rom8726/floxy-manager/internal/domain"
)

// Helper function to parse JSONB
func parseJSONB(data sql.NullString) json.RawMessage {
	if !data.Valid {
		return nil
	}
	return json.RawMessage(data.String)
}

// Model types for database mapping

type workflowDefinitionModel struct {
	TenantID   int            `db:"tenant_id"`
	ProjectID  int            `db:"project_id"`
	ID         string         `db:"id"`
	Name       string         `db:"name"`
	Version    int            `db:"version"`
	Definition sql.NullString `db:"definition"`
	CreatedAt  time.Time      `db:"created_at"`
}

func (m *workflowDefinitionModel) toDomain() domain.WorkflowDefinition {
	return domain.WorkflowDefinition{
		TenantID:   domain.TenantID(m.TenantID),
		ProjectID:  domain.ProjectID(m.ProjectID),
		ID:         m.ID,
		Name:       m.Name,
		Version:    m.Version,
		Definition: parseJSONB(m.Definition),
		CreatedAt:  m.CreatedAt,
	}
}

type workflowInstanceModel struct {
	TenantID    int            `db:"tenant_id"`
	ProjectID   int            `db:"project_id"`
	ID          int            `db:"id"`
	WorkflowID  string         `db:"workflow_id"`
	Status      string         `db:"status"`
	Input       sql.NullString `db:"input"`
	Output      sql.NullString `db:"output"`
	Error       sql.NullString `db:"error"`
	StartedAt   sql.NullTime   `db:"started_at"`
	CompletedAt sql.NullTime   `db:"completed_at"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (m *workflowInstanceModel) toDomain() domain.WorkflowInstance {
	return domain.WorkflowInstance{
		TenantID:    domain.TenantID(m.TenantID),
		ProjectID:   domain.ProjectID(m.ProjectID),
		ID:          m.ID,
		WorkflowID:  m.WorkflowID,
		Status:      m.Status,
		Input:       parseJSONB(m.Input),
		Output:      parseJSONB(m.Output),
		Error:       m.Error,
		StartedAt:   m.StartedAt,
		CompletedAt: m.CompletedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type workflowStepModel struct {
	TenantID               int            `db:"tenant_id"`
	ProjectID              int            `db:"project_id"`
	ID                     int            `db:"id"`
	InstanceID             int            `db:"instance_id"`
	StepName               string         `db:"step_name"`
	StepType               string         `db:"step_type"`
	Status                 string         `db:"status"`
	Input                  sql.NullString `db:"input"`
	Output                 sql.NullString `db:"output"`
	Error                  sql.NullString `db:"error"`
	RetryCount             int            `db:"retry_count"`
	MaxRetries             int            `db:"max_retries"`
	CompensationRetryCount int            `db:"compensation_retry_count"`
	IdempotencyKey         string         `db:"idempotency_key"`
	StartedAt              sql.NullTime   `db:"started_at"`
	CompletedAt            sql.NullTime   `db:"completed_at"`
	CreatedAt              time.Time      `db:"created_at"`
}

func (m *workflowStepModel) toDomain() domain.WorkflowStep {
	return domain.WorkflowStep{
		TenantID:               domain.TenantID(m.TenantID),
		ProjectID:              domain.ProjectID(m.ProjectID),
		ID:                     m.ID,
		InstanceID:             m.InstanceID,
		StepName:               m.StepName,
		StepType:               m.StepType,
		Status:                 m.Status,
		Input:                  parseJSONB(m.Input),
		Output:                 parseJSONB(m.Output),
		Error:                  m.Error,
		RetryCount:             m.RetryCount,
		MaxRetries:             m.MaxRetries,
		CompensationRetryCount: m.CompensationRetryCount,
		IdempotencyKey:         m.IdempotencyKey,
		StartedAt:              m.StartedAt,
		CompletedAt:            m.CompletedAt,
		CreatedAt:              m.CreatedAt,
	}
}

type workflowEventModel struct {
	TenantID   int            `db:"tenant_id"`
	ProjectID  int            `db:"project_id"`
	ID         int            `db:"id"`
	InstanceID int            `db:"instance_id"`
	StepID     sql.NullInt64  `db:"step_id"`
	EventType  string         `db:"event_type"`
	Payload    sql.NullString `db:"payload"`
	CreatedAt  time.Time      `db:"created_at"`
}

func (m *workflowEventModel) toDomain() domain.WorkflowEvent {
	return domain.WorkflowEvent{
		TenantID:   domain.TenantID(m.TenantID),
		ProjectID:  domain.ProjectID(m.ProjectID),
		ID:         m.ID,
		InstanceID: m.InstanceID,
		StepID:     domain.NullInt64{NullInt64: m.StepID},
		EventType:  m.EventType,
		Payload:    parseJSONB(m.Payload),
		CreatedAt:  m.CreatedAt,
	}
}

type activeWorkflowModel struct {
	TenantID        int             `db:"tenant_id"`
	ProjectID       int             `db:"project_id"`
	ID              int             `db:"id"`
	WorkflowID      string          `db:"workflow_id"`
	WorkflowName    string          `db:"workflow_name"`
	Status          string          `db:"status"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	DurationSeconds sql.NullFloat64 `db:"duration_seconds"`
	TotalSteps      int             `db:"total_steps"`
	CompletedSteps  int             `db:"completed_steps"`
	FailedSteps     int             `db:"failed_steps"`
	RunningSteps    int             `db:"running_steps"`
}

func (m *activeWorkflowModel) toDomain() domain.ActiveWorkflow {
	return domain.ActiveWorkflow{
		TenantID:        domain.TenantID(m.TenantID),
		ProjectID:       domain.ProjectID(m.ProjectID),
		WorkflowID:      m.WorkflowID,
		WorkflowName:    m.WorkflowName,
		InstanceID:      m.ID,
		Status:          m.Status,
		StartedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
		CurrentStep:     "",
		TotalSteps:      m.TotalSteps,
		CompletedSteps:  m.CompletedSteps,
		RolledBackSteps: 0,
	}
}

type workflowStatModel struct {
	TenantID           int             `db:"tenant_id"`
	ProjectID          int             `db:"project_id"`
	Name               string          `db:"name"`
	Version            int             `db:"version"`
	TotalInstances     int             `db:"total_instances"`
	CompletedInstances int             `db:"completed"`
	FailedInstances    int             `db:"failed"`
	RunningInstances   int             `db:"running"`
	AverageDuration    sql.NullFloat64 `db:"avg_duration_seconds"`
}

func (m *workflowStatModel) toDomain() domain.WorkflowStat {
	var avgDurationNanos int64
	if m.AverageDuration.Valid && m.AverageDuration.Float64 > 0 {
		avgDurationNanos = int64(m.AverageDuration.Float64 * 1e9)
	}
	return domain.WorkflowStat{
		TenantID:           domain.TenantID(m.TenantID),
		ProjectID:          domain.ProjectID(m.ProjectID),
		Name:               m.Name,
		Version:            m.Version,
		TotalInstances:     m.TotalInstances,
		CompletedInstances: m.CompletedInstances,
		FailedInstances:    m.FailedInstances,
		RunningInstances:   m.RunningInstances,
		AverageDuration:    avgDurationNanos,
	}
}

type dlqItemModel struct {
	TenantID   int            `db:"tenant_id"`
	ProjectID  int            `db:"project_id"`
	ID         int            `db:"id"`
	InstanceID int            `db:"instance_id"`
	WorkflowID string         `db:"workflow_id"`
	StepID     int            `db:"step_id"`
	StepName   string         `db:"step_name"`
	StepType   string         `db:"step_type"`
	Input      sql.NullString `db:"input"`
	Error      sql.NullString `db:"error"`
	Reason     string         `db:"reason"`
	CreatedAt  time.Time      `db:"created_at"`
}

func (m *dlqItemModel) toDomain() domain.DLQItem {
	return domain.DLQItem{
		TenantID:   domain.TenantID(m.TenantID),
		ProjectID:  domain.ProjectID(m.ProjectID),
		ID:         m.ID,
		InstanceID: m.InstanceID,
		WorkflowID: m.WorkflowID,
		StepID:     m.StepID,
		StepName:   m.StepName,
		StepType:   m.StepType,
		Input:      parseJSONB(m.Input),
		Error:      domain.NullString{NullString: m.Error},
		Reason:     m.Reason,
		CreatedAt:  m.CreatedAt,
	}
}
