package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type ProjectsHandler struct {
	projectsRepo contract.ProjectsRepository
}

func NewProjectsHandler(projectsRepo contract.ProjectsRepository) *ProjectsHandler {
	return &ProjectsHandler{
		projectsRepo: projectsRepo,
	}
}

func (h *ProjectsHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Get tenant_id from query parameter
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		slog.Warn("Missing tenant_id in request",
			"path", r.URL.Path,
		)
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	tenantID, err := strconv.Atoi(tenantIDStr)
	if err != nil {
		slog.Warn("Invalid tenant_id in request",
			"tenant_id", tenantIDStr,
			"error", err,
		)
		respondError(w, http.StatusBadRequest, "invalid tenant_id")
		return
	}

	projects, err := h.projectsRepo.ListByTenant(r.Context(), domain.TenantID(tenantID))
	if err != nil {
		slog.Error("Failed to list projects by tenant",
			"error", err,
			"tenant_id", tenantID,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, projects)
}
