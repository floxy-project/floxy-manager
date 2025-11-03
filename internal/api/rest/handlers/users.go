package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type UsersHandler struct {
	usersService       contract.UsersUseCase
	projectsRepo       contract.ProjectsRepository
	permissionsService contract.PermissionsService
}

func NewUsersHandler(
	usersService contract.UsersUseCase,
	projectsRepo contract.ProjectsRepository,
	permissionsService contract.PermissionsService,
) *UsersHandler {
	return &UsersHandler{
		usersService:       usersService,
		projectsRepo:       projectsRepo,
		permissionsService: permissionsService,
	}
}

// GetCurrentUser returns information about the current authenticated user
func (h *UsersHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.usersService.GetByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":                  user.ID,
		"username":            user.Username,
		"email":               user.Email,
		"is_superuser":        user.IsSuperuser,
		"is_active":           user.IsActive,
		"is_external":         user.IsExternal,
		"two_fa_enabled":      user.TwoFAEnabled,
		"two_fa_confirmed_at": user.TwoFAConfirmedAt,
		"created_at":          user.CreatedAt,
		"updated_at":          user.UpdatedAt,
		"last_login":          user.LastLogin,
		"license_accepted":    user.LicenseAccepted,
	})
}

// GetMyProjects returns projects and permissions for the current user
func (h *UsersHandler) GetMyProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	userID := appcontext.UserID(r.Context())
	isSuper := appcontext.IsSuper(r.Context())

	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var projects []domain.Project
	var projectRoles map[domain.ProjectID]domain.Role
	var projectPermissions map[domain.ProjectID][]domain.PermKey

	// Get all projects
	allProjects, err := h.projectsRepo.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get projects")
		return
	}

	if isSuper {
		// Superuser has access to all projects
		projects = allProjects

		// Superuser has all permissions
		projectRoles = make(map[domain.ProjectID]domain.Role)
		projectPermissions = make(map[domain.ProjectID][]domain.PermKey)

		// For superuser, mark all projects with superuser role
		allPermKeys := []domain.PermKey{
			domain.PermProjectView,
			domain.PermProjectManage,
			domain.PermProjectCreate,
			domain.PermAuditView,
			domain.PermMembershipManage,
		}

		for _, proj := range projects {
			projectRoles[proj.ID] = domain.Role{
				ID:          domain.RoleID("superuser"),
				Key:         "superuser",
				Name:        "Superuser",
				Description: "Full access to all projects and permissions",
				CreatedAt:   proj.CreatedAt, // Use project creation date as fallback
			}
			projectPermissions[proj.ID] = allPermKeys
		}
	} else {
		// Get user's projects via permissions service
		accessibleProjects, err := h.permissionsService.GetAccessibleProjects(r.Context(), allProjects)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get accessible projects")
			return
		}
		projects = accessibleProjects

		projectRoles, err = h.permissionsService.GetMyProjectRoles(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get project roles")
			return
		}

		projectPermissions, err = h.permissionsService.GetMyProjectPermissions(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get project permissions")
			return
		}
	}

	// Build response
	type ProjectResponse struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		CreatedAt   string  `json:"created_at"`
		UpdatedAt   string  `json:"updated_at"`
		ArchivedAt  *string `json:"archived_at,omitempty"`
	}

	type RoleResponse struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		CreatedAt   string `json:"created_at"`
	}

	type ProjectInfo struct {
		Project     ProjectResponse `json:"project"`
		Role        *RoleResponse   `json:"role,omitempty"`
		Permissions []string        `json:"permissions"`
	}

	result := make([]ProjectInfo, 0, len(projects))
	for _, project := range projects {
		role, hasRole := projectRoles[project.ID]
		perms := projectPermissions[project.ID]
		if perms == nil {
			perms = []domain.PermKey{}
		}

		// Convert permissions to strings
		permStrings := make([]string, len(perms))
		for i, p := range perms {
			permStrings[i] = string(p)
		}

		// Format dates
		createdAt := project.CreatedAt.Format(time.RFC3339)
		updatedAt := project.UpdatedAt.Format(time.RFC3339)
		var archivedAt *string
		if project.ArchivedAt != nil {
			formatted := project.ArchivedAt.Format(time.RFC3339)
			archivedAt = &formatted
		}

		projInfo := ProjectInfo{
			Project: ProjectResponse{
				ID:          project.ID.Int(),
				Name:        project.Name,
				Description: project.Description,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				ArchivedAt:  archivedAt,
			},
			Permissions: permStrings,
		}

		if hasRole {
			roleCreatedAt := role.CreatedAt.Format(time.RFC3339)
			projInfo.Role = &RoleResponse{
				ID:          string(role.ID),
				Key:         role.Key,
				Name:        role.Name,
				Description: role.Description,
				CreatedAt:   roleCreatedAt,
			}
		}

		result = append(result, projInfo)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"is_superuser": isSuper,
		"projects":     result,
	})
}

// UpdatePassword allows authenticated user to change their password
func (h *UsersHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	userID := appcontext.UserID(r.Context())
	if userID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "old_password and new_password are required")
		return
	}

	if len(req.NewPassword) < 6 {
		respondError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	err := h.usersService.UpdatePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPassword) {
			respondError(w, http.StatusBadRequest, "Invalid current password")
			return
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "You are not allowed to change password")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password updated successfully",
	})
}
