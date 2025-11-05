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

	allProjects, err := h.projectsRepo.ListByTenant(r.Context(), domain.TenantID(tenantID))
	if err != nil {
		slog.Error("Failed to list projects by tenant",
			"error", err,
			"tenant_id", tenantID,
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Filter projects by user permissions
	isSuper := appcontext.IsSuper(r.Context())
	var projects []domain.Project
	if isSuper {
		// Superuser can see all projects
		projects = allProjects
	} else {
		// Filter projects by permissions
		accessibleProjects, err := h.permissionsSrv.GetAccessibleProjects(r.Context(), allProjects)
		if err != nil {
			slog.Error("Failed to filter projects by permissions",
				"error", err,
				"tenant_id", tenantID,
			)
			respondError(w, http.StatusInternalServerError, "Failed to filter projects")
			return
		}
		projects = accessibleProjects
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

	// Check if user is superuser or is project_owner in any project
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		if userID == 0 {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if user is project_owner in any project
		allProjects, err := h.projectsRepo.List(r.Context())
		if err != nil {
			slog.Error("Failed to list projects for permission check", "error", err)
			respondError(w, http.StatusInternalServerError, "Failed to check permissions")
			return
		}

		isProjectOwner := false
		for _, project := range allProjects {
			roleID, err := h.membershipsRepo.GetForUserProject(r.Context(), userID, project.ID)
			if err != nil || roleID == "" {
				continue
			}

			role, err := h.rolesRepo.GetByID(r.Context(), domain.RoleID(roleID))
			if err != nil {
				continue
			}

			if role.Key == "project_owner" {
				isProjectOwner = true
				break
			}
		}

		if !isProjectOwner {
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

	// Check if user is superuser or has project.manage permission for this project
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		// Check if user has project.manage permission for this project
		hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), projectID, domain.PermProjectManage)
		if err != nil {
			slog.Error("Failed to check project.manage permission for delete",
				"error", err,
				"user_id", userID,
				"project_id", projectID,
			)
			respondError(w, http.StatusForbidden, "Only superusers or users with project.manage permission can delete projects")
			return
		}
		if !hasManage {
			slog.Warn("User denied delete project - no project.manage permission",
				"user_id", userID,
				"project_id", projectID,
			)
			respondError(w, http.StatusForbidden, "Only superusers or users with project.manage permission can delete projects")
			return
		}
		slog.Debug("User has project.manage permission for delete",
			"user_id", userID,
			"project_id", projectID,
		)
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

func (h *ProjectsHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
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

	// Check if user is superuser or has project.manage permission for this project
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		// Check if user has project.manage permission for this project
		hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), projectID, domain.PermProjectManage)
		if err != nil {
			slog.Error("Failed to check project.manage permission for update",
				"error", err,
				"user_id", userID,
				"project_id", projectID,
			)
			respondError(w, http.StatusForbidden, "Only superusers or users with project.manage permission can update projects")
			return
		}
		if !hasManage {
			slog.Warn("User denied update project - no project.manage permission",
				"user_id", userID,
				"project_id", projectID,
			)
			respondError(w, http.StatusForbidden, "Only superusers or users with project.manage permission can update projects")
			return
		}
		slog.Debug("User has project.manage permission for update",
			"user_id", userID,
			"project_id", projectID,
		)
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	err = h.projectsRepo.Update(r.Context(), projectID, req.Name, req.Description)
	if err != nil {
		if err == domain.ErrEntityNotFound {
			respondError(w, http.StatusNotFound, "project not found")
			return
		}
		slog.Error("Failed to update project",
			"error", err,
			"project_id", id,
			"name", req.Name,
		)
		respondError(w, http.StatusInternalServerError, "Failed to update project")
		return
	}

	// Get updated project to return
	project, err := h.projectsRepo.GetByID(r.Context(), projectID)
	if err != nil {
		slog.Error("Failed to get updated project", "error", err, "project_id", projectID)
		respondError(w, http.StatusInternalServerError, "Failed to retrieve updated project")
		return
	}

	respondJSON(w, http.StatusOK, project)
}
