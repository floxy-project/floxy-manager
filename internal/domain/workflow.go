package domain

import (
	"database/sql"
	"encoding/json"
	"time"
)

// NullInt64 is a nullable int64 that marshals to JSON as null or a number
type NullInt64 struct {
	sql.NullInt64
}

// MarshalJSON implements json.Marshaler interface
func (n NullInt64) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.Int64)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (n *NullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &n.Int64)
	if err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// NullString is a nullable string that marshals to JSON as null or a string
type NullString struct {
	sql.NullString
}

// MarshalJSON implements json.Marshaler interface
func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (n *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		n.Valid = false
		return nil
	}
	err := json.Unmarshal(data, &n.String)
	if err != nil {
		return err
	}
	n.Valid = true
	return nil
}

// WorkflowDefinition represents a workflow definition
type WorkflowDefinition struct {
	TenantID   TenantID        `json:"tenant_id"`
	ProjectID  ProjectID       `json:"project_id"`
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Version    int             `json:"version"`
	Definition json.RawMessage `json:"definition"`
	CreatedAt  time.Time       `json:"created_at"`
}

// WorkflowInstance represents a workflow instance
type WorkflowInstance struct {
	TenantID    TenantID        `json:"tenant_id"`
	ProjectID   ProjectID       `json:"project_id"`
	ID          int             `json:"id"`
	WorkflowID  string          `json:"workflow_id"`
	Status      string          `json:"status"`
	Input       json.RawMessage `json:"input"`
	Output      json.RawMessage `json:"output"`
	Error       sql.NullString  `json:"error,omitempty"`
	StartedAt   sql.NullTime    `json:"started_at,omitempty"`
	CompletedAt sql.NullTime    `json:"completed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// WorkflowStep represents a workflow step
type WorkflowStep struct {
	TenantID               TenantID        `json:"tenant_id"`
	ProjectID              ProjectID       `json:"project_id"`
	ID                     int             `json:"id"`
	InstanceID             int             `json:"instance_id"`
	StepName               string          `json:"step_name"`
	StepType               string          `json:"step_type"`
	Status                 string          `json:"status"`
	Input                  json.RawMessage `json:"input"`
	Output                 json.RawMessage `json:"output"`
	Error                  sql.NullString  `json:"error,omitempty"`
	RetryCount             int             `json:"retry_count"`
	MaxRetries             int             `json:"max_retries"`
	CompensationRetryCount int             `json:"compensation_retry_count"`
	IdempotencyKey         string          `json:"idempotency_key"`
	StartedAt              sql.NullTime    `json:"started_at,omitempty"`
	CompletedAt            sql.NullTime    `json:"completed_at,omitempty"`
	CreatedAt              time.Time       `json:"created_at"`
}

// WorkflowEvent represents a workflow event
type WorkflowEvent struct {
	TenantID   TenantID        `json:"tenant_id"`
	ProjectID  ProjectID       `json:"project_id"`
	ID         int             `json:"id"`
	InstanceID int             `json:"instance_id"`
	StepID     NullInt64       `json:"step_id,omitempty"`
	EventType  string          `json:"event_type"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  time.Time       `json:"created_at"`
}

// ActiveWorkflow represents an active workflow
type ActiveWorkflow struct {
	TenantID        TenantID  `json:"tenant_id"`
	ProjectID       ProjectID `json:"project_id"`
	WorkflowID      string    `json:"workflow_id"`
	WorkflowName    string    `json:"workflow_name"`
	InstanceID      int       `json:"instance_id"`
	Status          string    `json:"status"`
	StartedAt       time.Time `json:"started_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CurrentStep     string    `json:"current_step"`
	TotalSteps      int       `json:"total_steps"`
	CompletedSteps  int       `json:"completed_steps"`
	RolledBackSteps int       `json:"rolled_back_steps"`
}

// WorkflowStat represents workflow statistics
type WorkflowStat struct {
	TenantID           TenantID  `json:"tenant_id"`
	ProjectID          ProjectID `json:"project_id"`
	Name               string    `json:"name"`
	Version            int       `json:"version"`
	TotalInstances     int       `json:"total_instances"`
	CompletedInstances int       `json:"completed_instances"`
	FailedInstances    int       `json:"failed_instances"`
	RunningInstances   int       `json:"running_instances"`
	AverageDuration    int64     `json:"average_duration"` // nanoseconds
}

// DLQItem represents a dead letter queue item
type DLQItem struct {
	TenantID   TenantID        `json:"tenant_id"`
	ProjectID  ProjectID       `json:"project_id"`
	ID         int             `json:"id"`
	InstanceID int             `json:"instance_id"`
	WorkflowID string          `json:"workflow_id"`
	StepID     int             `json:"step_id"`
	StepName   string          `json:"step_name"`
	StepType   string          `json:"step_type"`
	Input      json.RawMessage `json:"input"`
	Error      NullString      `json:"error,omitempty"`
	Reason     string          `json:"reason"`
	CreatedAt  time.Time       `json:"created_at"`
}
