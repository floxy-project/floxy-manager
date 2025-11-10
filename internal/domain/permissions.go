package domain

// PermKey is a canonical permission identifier used across layers.
type PermKey string

const (
	// Project-level.
	PermProjectView   PermKey = "project.view"
	PermProjectManage PermKey = "project.manage"
	PermProjectCreate PermKey = "project.create"

	// Workflow-level.
	PermWorkflowCreate PermKey = "workflow.create"

	// Audit & Membership.
	PermAuditView        PermKey = "audit.view"
	PermMembershipManage PermKey = "membership.manage"
)
