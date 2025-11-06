package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

// CreateUser creates a new internal user. Only superusers can create users.
func (h *UsersHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can create users")
		return
	}

	currentUserID := appcontext.UserID(r.Context())
	if currentUserID == 0 {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	currentUser, err := h.usersService.GetByID(r.Context(), currentUserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get current user")
		return
	}

	var req struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		IsSuperuser bool   `json:"is_superuser"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		respondError(w, http.StatusBadRequest, "username is required")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	if req.Password == "" {
		respondError(w, http.StatusBadRequest, "password is required")
		return
	}

	if len(req.Password) < 6 {
		respondError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	user, err := h.usersService.Create(r.Context(), currentUser, req.Username, req.Email, req.Password, req.IsSuperuser)
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Only superusers can create users")
			return
		}
		if errors.Is(err, domain.ErrUsernameAlreadyInUse) {
			respondError(w, http.StatusConflict, "Username already in use")
			return
		}
		if errors.Is(err, domain.ErrEmailAlreadyInUse) {
			respondError(w, http.StatusConflict, "Email already in use")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":              user.ID,
		"username":        user.Username,
		"email":           user.Email,
		"is_superuser":    user.IsSuperuser,
		"is_active":       user.IsActive,
		"is_external":     user.IsExternal,
		"is_tmp_password": user.IsTmpPassword,
		"created_at":      user.CreatedAt.Format(time.RFC3339),
	})
}

// ListUsers returns all users. Superusers and users with membership.manage permission can list users.
func (h *UsersHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		userID := appcontext.UserID(r.Context())
		if userID == 0 {
			respondError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		// Check if user has membership.manage permission in any of their projects
		projectPermissions, err := h.permissionsService.GetMyProjectPermissions(r.Context())
		if err != nil {
			slog.Error("Failed to get project permissions", "error", err, "user_id", userID)
			respondError(w, http.StatusInternalServerError, "Failed to check permissions")
			return
		}

		slog.Debug("Checking user permissions for list users",
			"user_id", userID,
			"projects_count", len(projectPermissions),
		)

		hasMembershipManage := false
		for projectID, perms := range projectPermissions {
			slog.Debug("Checking project permissions",
				"user_id", userID,
				"project_id", projectID,
				"permissions", perms,
			)
			for _, perm := range perms {
				if perm == domain.PermMembershipManage {
					slog.Debug("User has membership.manage permission",
						"user_id", userID, "project_id", projectID, "permission", perm)
					hasMembershipManage = true
					break
				}
			}
			if hasMembershipManage {
				break
			}
		}

		if !hasMembershipManage {
			slog.Warn("User denied access to list users",
				"user_id", userID,
				"project_permissions_count", len(projectPermissions),
				"project_permissions", projectPermissions,
			)
			respondError(w, http.StatusForbidden, "Only superusers or users with membership.manage permission can list users")
			return
		}
	}

	users, err := h.usersService.List(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Access denied")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		var lastLogin *string
		if user.LastLogin != nil {
			formatted := user.LastLogin.Format(time.RFC3339)
			lastLogin = &formatted
		}
		var twoFAConfirmedAt *string
		if user.TwoFAConfirmedAt != nil {
			formatted := user.TwoFAConfirmedAt.Format(time.RFC3339)
			twoFAConfirmedAt = &formatted
		}

		result = append(result, map[string]interface{}{
			"id":                  user.ID,
			"username":            user.Username,
			"email":               user.Email,
			"is_superuser":        user.IsSuperuser,
			"is_active":           user.IsActive,
			"is_external":         user.IsExternal,
			"two_fa_enabled":      user.TwoFAEnabled,
			"two_fa_confirmed_at": twoFAConfirmedAt,
			"created_at":          user.CreatedAt.Format(time.RFC3339),
			"updated_at":          user.UpdatedAt.Format(time.RFC3339),
			"last_login":          lastLogin,
			"license_accepted":    user.LicenseAccepted,
			"is_tmp_password":     user.IsTmpPassword,
		})
	}

	respondJSON(w, http.StatusOK, result)
}

// UpdateUserStatus updates user active status. Only superusers can update user status.
func (h *UsersHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can update user status")
		return
	}

	// Get user ID from URL params
	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "user id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req struct {
		IsActive    *bool `json:"is_active,omitempty"`
		IsSuperuser *bool `json:"is_superuser,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := domain.UserID(id)
	var updatedUser domain.User

	// Get current user first
	updatedUser, err = h.usersService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get user")
		return
	}

	if req.IsActive != nil {
		updatedUser, err = h.usersService.SetActiveStatus(r.Context(), userID, *req.IsActive)
		if err != nil {
			if errors.Is(err, domain.ErrEntityNotFound) {
				respondError(w, http.StatusNotFound, "user not found")
				return
			}
			if errors.Is(err, domain.ErrPermissionDenied) {
				respondError(w, http.StatusForbidden, "Only superusers can update user status")
				return
			}
			respondError(w, http.StatusInternalServerError, "Failed to update user status")
			return
		}
	}

	if req.IsSuperuser != nil {
		updatedUser, err = h.usersService.SetSuperuserStatus(r.Context(), userID, *req.IsSuperuser)
		if err != nil {
			if errors.Is(err, domain.ErrEntityNotFound) {
				respondError(w, http.StatusNotFound, "user not found")
				return
			}
			if errors.Is(err, domain.ErrPermissionDenied) {
				respondError(w, http.StatusForbidden, "Only superusers can update user status")
				return
			}
			respondError(w, http.StatusInternalServerError, "Failed to update user status")
			return
		}
	}

	var lastLogin *string
	if updatedUser.LastLogin != nil {
		formatted := updatedUser.LastLogin.Format(time.RFC3339)
		lastLogin = &formatted
	}
	var twoFAConfirmedAt *string
	if updatedUser.TwoFAConfirmedAt != nil {
		formatted := updatedUser.TwoFAConfirmedAt.Format(time.RFC3339)
		twoFAConfirmedAt = &formatted
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id":                  updatedUser.ID,
		"username":            updatedUser.Username,
		"email":               updatedUser.Email,
		"is_superuser":        updatedUser.IsSuperuser,
		"is_active":           updatedUser.IsActive,
		"is_external":         updatedUser.IsExternal,
		"two_fa_enabled":      updatedUser.TwoFAEnabled,
		"two_fa_confirmed_at": twoFAConfirmedAt,
		"created_at":          updatedUser.CreatedAt.Format(time.RFC3339),
		"updated_at":          updatedUser.UpdatedAt.Format(time.RFC3339),
		"last_login":          lastLogin,
		"license_accepted":    updatedUser.LicenseAccepted,
		"is_tmp_password":     updatedUser.IsTmpPassword,
	})
}

// DeleteUser deletes a user. Only superusers can delete users.
func (h *UsersHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Check if user is superuser
	if !appcontext.IsSuper(r.Context()) {
		respondError(w, http.StatusForbidden, "Only superusers can delete users")
		return
	}

	// Get user ID from URL params
	idStr := appcontext.Param(r.Context(), "id")
	if idStr == "" {
		respondError(w, http.StatusBadRequest, "user id is required")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	err = h.usersService.Delete(r.Context(), domain.UserID(id))
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, domain.ErrPermissionDenied) {
			respondError(w, http.StatusForbidden, "Only superusers can delete users")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "user deleted successfully"})
}
