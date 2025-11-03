package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type ProjectsHandler struct {
	projectsRepo    contract.ProjectsRepository
	permissionsSrv  contract.PermissionsService
	rolesRepo       contract.RolesRepository
	membershipsRepo contract.MembershipsRepository
}

func NewProjectsHandler(
	projectsRepo contract.ProjectsRepository,
	permissionsSrv contract.PermissionsService,
	rolesRepo contract.RolesRepository,
	membershipsRepo contract.MembershipsRepository,
) *ProjectsHandler {
	return &ProjectsHandler{
		projectsRepo:    projectsRepo,
		permissionsSrv:  permissionsSrv,
		rolesRepo:       rolesRepo,
		membershipsRepo: membershipsRepo,
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

func (h *ProjectsHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser or has project.manage permission (meaning they are project_owner somewhere)
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		if userID == 0 {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if user has project.manage permission in any project (meaning they are project_owner)
		allProjects, err := h.projectsRepo.List(r.Context())
		if err != nil {
			slog.Error("Failed to list projects for permission check", "error", err)
			respondError(w, http.StatusInternalServerError, "Failed to check permissions")
			return
		}

		hasManagePermission := false
		for _, project := range allProjects {
			hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), project.ID, domain.PermProjectManage)
			if err == nil && hasManage {
				hasManagePermission = true
				break
			}
		}

		if !hasManagePermission {
			respondError(w, http.StatusForbidden, "Only superusers or project owners can create projects")
			return
		}
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TenantID    int    `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	if req.TenantID == 0 {
		respondError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	projectDTO := &domain.ProjectDTO{
		Name:        req.Name,
		Description: req.Description,
	}

	projectID, err := h.projectsRepo.Create(r.Context(), projectDTO, domain.TenantID(req.TenantID))
	if err != nil {
		slog.Error("Failed to create project",
			"error", err,
			"name", req.Name,
			"tenant_id", req.TenantID,
		)
		respondError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	project, err := h.projectsRepo.GetByID(r.Context(), projectID)
	if err != nil {
		slog.Error("Failed to get created project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve created project")
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

func (h *ProjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Get project ID from URL params
	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "project id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	projectID := domain.ProjectID(id)

	// Check if user is superuser or project_owner of this project
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		if userID == 0 {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if user is project_owner of this project
		roleID, err := h.membershipsRepo.GetForUserProject(r.Context(), userID, projectID)
		if err != nil || roleID == "" {
			respondError(w, http.StatusForbidden, "Only superusers or project owners can delete projects")
			return
		}

		// Get role to check if it's project_owner
		role, err := h.rolesRepo.GetByID(r.Context(), domain.RoleID(roleID))
		if err != nil {
			slog.Error("Failed to get role", "error", err, "role_id", roleID)
			respondError(w, http.StatusInternalServerError, "Failed to verify permissions")
			return
		}

		if role.Key != "project_owner" {
			respondError(w, http.StatusForbidden, "Only superusers or project owners can delete projects")
			return
		}
	}

	err = h.projectsRepo.Delete(r.Context(), projectID)
	if err != nil {
		if err == domain.ErrEntityNotFound {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
		slog.Error("Failed to delete project",
			"error", err,
			"project_id", id,
		)
		respondError(w, http.StatusInternalServerError, "Failed to delete project")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "project deleted successfully"})
}
