package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type WorkflowsHandler struct {
	workflowsRepo  contract.WorkflowsRepository
	permissionsSrv contract.PermissionsService
}

func NewWorkflowsHandler(
	workflowsRepo contract.WorkflowsRepository,
	permissionsSrv contract.PermissionsService,
) *WorkflowsHandler {
	return &WorkflowsHandler{
		workflowsRepo:  workflowsRepo,
		permissionsSrv: permissionsSrv,
	}
}

// ListWorkflows handles GET /api/v1/workflows
func (h *WorkflowsHandler) ListWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	workflowDefs, total, err := h.workflowsRepo.ListWorkflowDefinitions(r.Context(), tenantID, projectID, page, pageSize)
	if err != nil {
		slog.Error("Failed to list workflow definitions",
			"error", err,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     workflowDefs,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// GetWorkflow handles GET /api/v1/workflows/:id
func (h *WorkflowsHandler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	id := appcontext.Param(r.Context(), "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"workflow_id", id,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	workflow, err := h.workflowsRepo.GetWorkflowDefinition(r.Context(), tenantID, projectID, id)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			slog.Warn("Workflow not found",
				"workflow_id", id,
				"tenant_id", tenantID,
				"project_id", projectID,
			)
			respondError(w, http.StatusNotFound, "Workflow not found")
			return
		}
		slog.Error("Failed to get workflow definition",
			"error", err,
			"workflow_id", id,
			"tenant_id", tenantID,
			"project_id", projectID,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, workflow)
}

// ListWorkflowInstances handles GET /api/v1/workflows/:id/instances
func (h *WorkflowsHandler) ListWorkflowInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	workflowID := appcontext.Param(r.Context(), "id")
	if workflowID == "" {
		respondError(w, http.StatusBadRequest, "Invalid workflow ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"workflow_id", workflowID,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	instances, total, err := h.workflowsRepo.ListWorkflowInstances(
		r.Context(),
		tenantID,
		projectID,
		workflowID,
		page,
		pageSize,
	)
	if err != nil {
		slog.Error("Failed to list workflow instances for workflow",
			"error", err,
			"workflow_id", workflowID,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     instances,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// ListInstances handles GET /api/v1/instances
func (h *WorkflowsHandler) ListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	instances, total, err := h.workflowsRepo.ListWorkflowInstances(
		r.Context(),
		tenantID,
		projectID,
		"",
		page,
		pageSize,
	)
	if err != nil {
		slog.Error("Failed to list workflow instances",
			"error", err,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     instances,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// ListActiveWorkflows handles GET /api/v1/active-workflows
func (h *WorkflowsHandler) ListActiveWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	activeWorkflows, total, err := h.workflowsRepo.ListActiveWorkflows(
		r.Context(),
		tenantID,
		projectID,
		page,
		pageSize,
	)
	if err != nil {
		slog.Error("Failed to list active workflows",
			"error", err,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     activeWorkflows,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// GetInstance handles GET /api/v1/instances/:id
func (h *WorkflowsHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("Invalid instance ID",
			"instance_id", idStr,
			"error", err,
		)
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"instance_id", id,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	instance, err := h.workflowsRepo.GetWorkflowInstance(r.Context(), tenantID, projectID, id)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			slog.Warn("Instance not found",
				"instance_id", id,
				"tenant_id", tenantID,
				"project_id", projectID,
			)
			respondError(w, http.StatusNotFound, "Instance not found")
			return
		}
		slog.Error("Failed to get workflow instance",
			"error", err,
			"instance_id", id,
			"tenant_id", tenantID,
			"project_id", projectID,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, instance)
}

// ListInstanceSteps handles GET /api/v1/instances/:id/steps
func (h *WorkflowsHandler) ListInstanceSteps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("Invalid instance ID",
			"instance_id", idStr,
			"error", err,
		)
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"instance_id", id,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	steps, total, err := h.workflowsRepo.ListWorkflowSteps(r.Context(), tenantID, projectID, id, page, pageSize)
	if err != nil {
		slog.Error("Failed to list workflow steps",
			"error", err,
			"instance_id", id,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     steps,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// ListInstanceEvents handles GET /api/v1/instances/:id/events
func (h *WorkflowsHandler) ListInstanceEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("Invalid instance ID",
			"instance_id", idStr,
			"error", err,
		)
		respondError(w, http.StatusBadRequest, "Invalid instance ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"instance_id", id,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	events, total, err := h.workflowsRepo.ListWorkflowEvents(r.Context(), tenantID, projectID, id, page, pageSize)
	if err != nil {
		slog.Error("Failed to list workflow events",
			"error", err,
			"instance_id", id,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     events,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// ListStats handles GET /api/v1/stats
func (h *WorkflowsHandler) ListStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	stats, total, err := h.workflowsRepo.ListWorkflowStats(r.Context(), tenantID, projectID, page, pageSize)
	if err != nil {
		slog.Error("Failed to list workflow stats",
			"error", err,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     stats,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// ListDLQ handles GET /api/v1/dlq
func (h *WorkflowsHandler) ListDLQ(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	page, pageSize := parsePagination(r)

	items, total, err := h.workflowsRepo.ListDLQItems(r.Context(), tenantID, projectID, page, pageSize)
	if err != nil {
		slog.Error("Failed to list DLQ items",
			"error", err,
			"tenant_id", tenantID,
			"project_id", projectID,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     items,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// GetDLQItem handles GET /api/v1/dlq/:id
func (h *WorkflowsHandler) GetDLQItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "Invalid DLQ item ID")
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Warn("Invalid DLQ item ID",
			"dlq_item_id", idStr,
			"error", err,
		)
		respondError(w, http.StatusBadRequest, "Invalid DLQ item ID")
		return
	}

	tenantID, projectID, err := parseTenantAndProject(r)
	if err != nil {
		slog.Warn("Invalid tenant_id or project_id in request",
			"error", err,
			"dlq_item_id", id,
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user has permission to view this project
	if err := h.permissionsSrv.CanViewProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	item, err := h.workflowsRepo.GetDLQItem(r.Context(), tenantID, projectID, id)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			slog.Warn("DLQ item not found",
				"dlq_item_id", id,
				"tenant_id", tenantID,
				"project_id", projectID,
			)
			respondError(w, http.StatusNotFound, "DLQ item not found")
			return
		}
		slog.Error("Failed to get DLQ item",
			"error", err,
			"dlq_item_id", id,
			"tenant_id", tenantID,
			"project_id", projectID,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, item)
}

// ListUnassignedWorkflows handles GET /api/v1/workflows/unassigned
func (h *WorkflowsHandler) ListUnassignedWorkflows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	// Check if user is superuser or has project management permission
	// Superusers can view all unassigned workflows
	// Project managers can also view unassigned workflows to assign them to their projects
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		// For non-superusers, we still allow viewing unassigned workflows
		// as they may need to assign them to their projects
		// The actual assignment will be checked by CanManageProject in AssignWorkflowsToProject
	}

	page, pageSize := parsePagination(r)

	workflowDefs, total, err := h.workflowsRepo.ListUnassignedWorkflowDefinitions(r.Context(), page, pageSize)
	if err != nil {
		slog.Error("Failed to list unassigned workflow definitions",
			"error", err,
			"page", page,
			"page_size", pageSize,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items":     workflowDefs,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

// AssignWorkflowsToProject handles POST /api/v1/projects/:id/workflows/assign
func (h *WorkflowsHandler) AssignWorkflowsToProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !requireAuthForWorkflows(w, r) {
		return
	}

	// Get project ID from URL params
	projectIDStr := appcontext.Param(r.Context(), "id")
	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project id is required")
		return
	}

	projectIDInt, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	projectID := domain.ProjectID(projectIDInt)

	// Check if user has permission to manage this project
	if err := h.permissionsSrv.CanManageProject(r.Context(), projectID); err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied to manage this project")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
		return
	}

	var req struct {
		WorkflowIDs []string `json:"workflow_ids"`
		// If workflow_ids is empty, assign all unassigned workflows
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	assignedCount, err := h.workflowsRepo.AssignWorkflowDefinitionsToProject(
		r.Context(),
		projectID,
		req.WorkflowIDs,
	)
	if err != nil {
		slog.Error("Failed to assign workflow definitions to project",
			"error", err,
			"project_id", projectID,
			"workflow_ids", req.WorkflowIDs,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Workflow definitions assigned successfully",
		"assigned_count": assignedCount,
	})
}

func requireAuthForWorkflows(w http.ResponseWriter, r *http.Request) bool {
	return checkAuthAndRespond(w, r)
}

func parseTenantAndProject(r *http.Request) (domain.TenantID, domain.ProjectID, error) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	projectIDStr := r.URL.Query().Get("project_id")

	if tenantIDStr == "" || projectIDStr == "" {
		return 0, 0, fmt.Errorf("tenant_id and project_id are required")
	}

	tenantID, err := strconv.Atoi(tenantIDStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid tenant_id")
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid project_id")
	}

	return domain.TenantID(tenantID), domain.ProjectID(projectID), nil
}

func parsePagination(r *http.Request) (page, pageSize int) {
	page = 1
	pageSize = 20
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	return page, pageSize
}
