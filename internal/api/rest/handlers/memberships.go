package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	appcontext "github.com/rom8726/floxy-manager/internal/context"
	"github.com/rom8726/floxy-manager/internal/contract"
	"github.com/rom8726/floxy-manager/internal/domain"
)

type MembershipsHandler struct {
	membershipsSrv contract.MembershipsUseCase
	usersSrv       contract.UsersUseCase
	permissionsSrv contract.PermissionsService
}

func NewMembershipsHandler(
	membershipsSrv contract.MembershipsUseCase,
	usersSrv contract.UsersUseCase,
	permissionsSrv contract.PermissionsService,
) *MembershipsHandler {
	return &MembershipsHandler{
		membershipsSrv: membershipsSrv,
		usersSrv:       usersSrv,
		permissionsSrv: permissionsSrv,
	}
}

// ListProjectMemberships returns all memberships for a project
func (h *MembershipsHandler) ListProjectMemberships(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Get project ID from URL parameter
	projectIDStr := appcontext.Param(r.Context(), "id")
	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	projID := domain.ProjectID(projectID)

	// Check if user has permission to view memberships (membership.manage or project.view)
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), projID, domain.PermMembershipManage)
		if err != nil || !hasManage {
			// Fallback to project.view permission
			hasView, err := h.permissionsSrv.HasProjectPermission(r.Context(), projID, domain.PermProjectView)
			if err != nil || !hasView {
				respondError(w, http.StatusForbidden, "Access denied to this project")
				return
			}
		}
	}

	memberships, err := h.membershipsSrv.ListProjectMemberships(r.Context(), projID)
	if err != nil {
		slog.Error("Failed to list project memberships",
			"error", err,
			"project_id", projectID,
		)
		respondError(w, http.StatusInternalServerError, "Failed to list memberships")
		return
	}

	// Enrich with user information
	type MembershipResponse struct {
		ID        string `json:"id"`
		ProjectID int    `json:"project_id"`
		UserID    int    `json:"user_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		RoleID    string `json:"role_id"`
		RoleKey   string `json:"role_key"`
		RoleName  string `json:"role_name"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]MembershipResponse, 0, len(memberships))
	for _, membership := range memberships {
		user, err := h.usersSrv.GetByID(r.Context(), membership.UserID)
		if err != nil {
			slog.Warn("Failed to get user for membership",
				"error", err,
				"user_id", membership.UserID,
				"membership_id", membership.ID,
			)
			continue
		}

		result = append(result, MembershipResponse{
			ID:        string(membership.ID),
			ProjectID: membership.ProjectID.Int(),
			UserID:    membership.UserID.Int(),
			Username:  user.Username,
			Email:     user.Email,
			RoleID:    string(membership.RoleID),
			RoleKey:   membership.RoleKey,
			RoleName:  membership.RoleName,
			CreatedAt: membership.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	respondJSON(w, http.StatusOK, result)
}

// CreateProjectMembership adds a user to a project with a specific role
func (h *MembershipsHandler) CreateProjectMembership(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Get project ID from URL parameter
	projectIDStr := appcontext.Param(r.Context(), "id")
	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	var req struct {
		UserID int    `json:"user_id"`
		RoleID string `json:"role_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == 0 {
		respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	if req.RoleID == "" {
		respondError(w, http.StatusBadRequest, "role_id is required")
		return
	}

	projID := domain.ProjectID(projectID)

	// Check if user has permission to manage memberships
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), projID, domain.PermMembershipManage)
		if err != nil || !hasManage {
			respondError(w, http.StatusForbidden, "Only users with membership.manage permission can add members")
			return
		}
	}

	// Verify user exists
	_, err = h.usersSrv.GetByID(r.Context(), domain.UserID(req.UserID))
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	membership, err := h.membershipsSrv.CreateProjectMembership(
		r.Context(),
		projID,
		domain.UserID(req.UserID),
		domain.RoleID(req.RoleID),
	)
	if err != nil {
		// Check for duplicate membership
		if errors.Is(err, domain.ErrEntityAlreadyExists) {
			respondError(w, http.StatusConflict, "User is already a member of this project")
			return
		}
		slog.Error("Failed to create project membership",
			"error", err,
			"project_id", projectID,
			"user_id", req.UserID,
			"role_id", req.RoleID,
		)
		respondError(w, http.StatusInternalServerError, "Failed to create membership")
		return
	}

	user, err := h.usersSrv.GetByID(r.Context(), membership.UserID)
	if err != nil {
		slog.Error("Failed to get user for membership response",
			"error", err,
			"user_id", membership.UserID,
		)
		respondError(w, http.StatusInternalServerError, "Failed to get user information")
		return
	}

	type MembershipResponse struct {
		ID        string `json:"id"`
		ProjectID int    `json:"project_id"`
		UserID    int    `json:"user_id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		RoleID    string `json:"role_id"`
		RoleKey   string `json:"role_key"`
		RoleName  string `json:"role_name"`
		CreatedAt string `json:"created_at"`
	}

	respondJSON(w, http.StatusCreated, MembershipResponse{
		ID:        string(membership.ID),
		ProjectID: membership.ProjectID.Int(),
		UserID:    membership.UserID.Int(),
		Username:  user.Username,
		Email:     user.Email,
		RoleID:    string(membership.RoleID),
		RoleKey:   membership.RoleKey,
		RoleName:  membership.RoleName,
		CreatedAt: membership.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteProjectMembership removes a user from a project
func (h *MembershipsHandler) DeleteProjectMembership(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	// Get project ID and membership ID from URL parameters
	projectIDStr := appcontext.Param(r.Context(), "id")
	membershipIDStr := appcontext.Param(r.Context(), "mid")

	if projectIDStr == "" {
		respondError(w, http.StatusBadRequest, "project_id is required")
		return
	}

	if membershipIDStr == "" {
		respondError(w, http.StatusBadRequest, "membership id is required")
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid project_id")
		return
	}

	membID, err := strconv.Atoi(membershipIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid membership_id")
		return
	}

	projID := domain.ProjectID(projectID)
	membershipID := domain.MembershipID(membID)

	// Check if user has permission to manage memberships
	isSuper := appcontext.IsSuper(r.Context())
	if !isSuper {
		hasManage, err := h.permissionsSrv.HasProjectPermission(r.Context(), projID, domain.PermMembershipManage)
		if err != nil || !hasManage {
			respondError(w, http.StatusForbidden, "Only users with membership.manage permission can remove members")
			return
		}
	}

	err = h.membershipsSrv.DeleteProjectMembership(r.Context(), projID, membershipID)
	if err != nil {
		if errors.Is(err, domain.ErrEntityNotFound) {
			respondError(w, http.StatusNotFound, "Membership not found")
			return
		}
		slog.Error("Failed to delete project membership",
			"error", err,
			"project_id", projectID,
			"membership_id", membershipIDStr,
		)
		respondError(w, http.StatusInternalServerError, "Failed to delete membership")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Membership deleted successfully"})
}

// ListRoles returns all available roles
func (h *MembershipsHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !checkAuthAndRespond(w, r) {
		return
	}

	roles, err := h.membershipsSrv.ListRoles(r.Context())
	if err != nil {
		slog.Error("Failed to list roles", "error", err)
		respondError(w, http.StatusInternalServerError, "Failed to list roles")
		return
	}

	type RoleResponse struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	result := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, RoleResponse{
			ID:          string(role.ID),
			Key:         role.Key,
			Name:        role.Name,
			Description: role.Description,
		})
	}

	respondJSON(w, http.StatusOK, result)
}
